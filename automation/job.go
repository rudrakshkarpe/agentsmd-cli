package automation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

type Job struct {
	Trajectory string    `json:"trajectory"`
	Status     string    `json:"status"`
	Result     *Result   `json:"result,omitempty"`
	Error      string    `json:"error,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
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
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if errors.Is(err, os.ErrExist) {
		info, statErr := os.Stat(lockPath)
		if statErr != nil || time.Since(info.ModTime()) < 15*time.Minute {
			return Job{}, fmt.Errorf("job is already processing")
		}
		if removeErr := os.Remove(lockPath); removeErr != nil {
			return Job{}, fmt.Errorf("remove stale job lock: %w", removeErr)
		}
		lock, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	}
	if err != nil {
		return Job{}, err
	}
	lock.Close()
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
