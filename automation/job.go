package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

const DefaultLockStaleAfter = 15 * time.Minute

type Job struct {
	Trajectory string    `json:"trajectory"`
	Status     string    `json:"status"`
	Result     *Result   `json:"result,omitempty"`
	Error      string    `json:"error,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Lock struct {
	PID       int       `json:"pid"`
	Hostname  string    `json:"hostname"`
	StartedAt time.Time `json:"started_at"`
}

type QueueHealth struct {
	Queued             int `json:"queued"`
	Processing         int `json:"processing"`
	Complete           int `json:"complete"`
	Failed             int `json:"failed"`
	StaleLocks         int `json:"stale_locks"`
	OrphanedProcessing int `json:"orphaned_processing"`
}

type Recovery struct {
	LocksRemoved int `json:"locks_removed"`
	JobsRequeued int `json:"jobs_requeued"`
}

func Enqueue(p *project.Project, trajectoryPath string) (string, error) {
	jobPath := filepath.Join(p.QueueDir(), filepath.Base(trajectoryPath))
	if existing, err := loadJob(jobPath); err == nil && existing.Status != "failed" {
		return jobPath, nil
	}
	job := Job{Trajectory: trajectoryPath, Status: "queued", UpdatedAt: time.Now().UTC()}
	return jobPath, saveJob(jobPath, job)
}

func ProcessJob(ctx context.Context, p *project.Project, jobPath string) (Job, error) {
	lockPath := jobPath + ".lock"
	if err := acquireLock(lockPath, time.Now().UTC(), DefaultLockStaleAfter); err != nil {
		return Job{}, err
	}
	defer os.Remove(lockPath)

	job, err := loadJob(jobPath)
	if err != nil {
		return Job{}, err
	}
	if job.Status == "complete" {
		return job, nil
	}
	job.Status, job.UpdatedAt = "processing", time.Now().UTC()
	if err := saveJob(jobPath, job); err != nil {
		return Job{}, err
	}
	result, processErr := Process(ctx, p, job.Trajectory)
	job.UpdatedAt = time.Now().UTC()
	if processErr != nil {
		job.Status, job.Error = "failed", processErr.Error()
	} else {
		job.Status, job.Result, job.Error = "complete", &result, ""
	}
	if err := saveJob(jobPath, job); err != nil {
		return Job{}, err
	}
	return job, processErr
}

func InspectQueue(p *project.Project, now time.Time, staleAfter time.Duration) (QueueHealth, error) {
	entries, err := os.ReadDir(p.QueueDir())
	if errors.Is(err, os.ErrNotExist) {
		return QueueHealth{}, nil
	}
	if err != nil {
		return QueueHealth{}, err
	}
	health := QueueHealth{}
	locks := map[string]bool{}
	for _, entry := range entries {
		path := filepath.Join(p.QueueDir(), entry.Name())
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".lock") {
			jobName := strings.TrimSuffix(entry.Name(), ".lock")
			locks[jobName] = true
			stale, staleErr := lockIsStale(path, now, staleAfter)
			if staleErr != nil {
				return QueueHealth{}, staleErr
			}
			if stale {
				health.StaleLocks++
			}
		}
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		job, loadErr := loadJob(filepath.Join(p.QueueDir(), entry.Name()))
		if loadErr != nil {
			return QueueHealth{}, loadErr
		}
		switch job.Status {
		case "queued":
			health.Queued++
		case "processing":
			health.Processing++
			if !locks[entry.Name()] {
				health.OrphanedProcessing++
			}
		case "complete":
			health.Complete++
		case "failed":
			health.Failed++
		}
	}
	return health, nil
}

// RecoverQueue removes only expired locks and requeues processing jobs that no
// longer have a lock. Fresh locks and failed jobs are left untouched.
func RecoverQueue(p *project.Project, now time.Time, staleAfter time.Duration) (Recovery, error) {
	entries, err := os.ReadDir(p.QueueDir())
	if errors.Is(err, os.ErrNotExist) {
		return Recovery{}, nil
	}
	if err != nil {
		return Recovery{}, err
	}
	recovery := Recovery{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lock") {
			continue
		}
		path := filepath.Join(p.QueueDir(), entry.Name())
		stale, staleErr := lockIsStale(path, now, staleAfter)
		if staleErr != nil {
			return recovery, staleErr
		}
		if stale {
			if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				return recovery, fmt.Errorf("remove stale job lock: %w", removeErr)
			}
			recovery.LocksRemoved++
		}
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(p.QueueDir(), entry.Name())
		job, loadErr := loadJob(path)
		if loadErr != nil {
			return recovery, loadErr
		}
		if job.Status != "processing" {
			continue
		}
		if _, statErr := os.Stat(path + ".lock"); statErr == nil {
			continue
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return recovery, statErr
		}
		job.Status, job.Error, job.UpdatedAt = "queued", "", now.UTC()
		if saveErr := saveJob(path, job); saveErr != nil {
			return recovery, saveErr
		}
		recovery.JobsRequeued++
	}
	return recovery, nil
}

func acquireLock(path string, now time.Time, staleAfter time.Duration) error {
	lock, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if errors.Is(err, os.ErrExist) {
		stale, staleErr := lockIsStale(path, now, staleAfter)
		if staleErr != nil {
			return staleErr
		}
		if !stale {
			return fmt.Errorf("job is already processing")
		}
		if removeErr := os.Remove(path); removeErr != nil {
			return fmt.Errorf("remove stale job lock: %w", removeErr)
		}
		lock, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	}
	if err != nil {
		return err
	}
	hostname, _ := os.Hostname()
	encodeErr := json.NewEncoder(lock).Encode(Lock{PID: os.Getpid(), Hostname: hostname, StartedAt: now.UTC()})
	closeErr := lock.Close()
	if encodeErr != nil {
		_ = os.Remove(path)
		return encodeErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return closeErr
	}
	return nil
}

func lockIsStale(path string, now time.Time, staleAfter time.Duration) (bool, error) {
	if staleAfter <= 0 {
		staleAfter = DefaultLockStaleAfter
	}
	started := time.Time{}
	if data, err := os.ReadFile(path); err == nil {
		var lock Lock
		if json.Unmarshal(data, &lock) == nil {
			started = lock.StartedAt
		}
	} else {
		return false, err
	}
	if started.IsZero() {
		info, err := os.Stat(path)
		if err != nil {
			return false, err
		}
		started = info.ModTime()
	}
	return !started.After(now) && now.Sub(started) >= staleAfter, nil
}

func loadJob(path string) (Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Job{}, err
	}
	var value Job
	return value, json.Unmarshal(data, &value)
}

func saveJob(path string, value Job) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return project.AtomicWrite(path, append(data, '\n'), 0o644)
}
