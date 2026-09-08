package automation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rudrakshkarpe/agentsmd-cli/project"
)

func TestRecoverQueueRemovesExpiredLockAndRequeuesJob(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 9, 3, 0, 0, 0, time.UTC)
	jobPath := filepath.Join(p.QueueDir(), "claude-session.json")
	if err := saveJob(jobPath, Job{Trajectory: "run.json", Status: "processing", UpdatedAt: now.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := acquireLock(jobPath+".lock", now.Add(-time.Hour), DefaultLockStaleAfter); err != nil {
		t.Fatal(err)
	}
	health, err := InspectQueue(p, now, DefaultLockStaleAfter)
	if err != nil || health.StaleLocks != 1 || health.Processing != 1 {
		t.Fatalf("health=%+v err=%v", health, err)
	}
	recovery, err := RecoverQueue(p, now, DefaultLockStaleAfter)
	if err != nil || recovery.LocksRemoved != 1 || recovery.JobsRequeued != 1 {
		t.Fatalf("recovery=%+v err=%v", recovery, err)
	}
	job, err := loadJob(jobPath)
	if err != nil || job.Status != "queued" || !job.UpdatedAt.Equal(now) {
		t.Fatalf("job=%+v err=%v", job, err)
	}
}

func TestRecoverQueuePreservesFreshWorkerLock(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 9, 3, 0, 0, 0, time.UTC)
	jobPath := filepath.Join(p.QueueDir(), "cursor-session.json")
	if err := saveJob(jobPath, Job{Trajectory: "run.json", Status: "processing", UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := acquireLock(jobPath+".lock", now, DefaultLockStaleAfter); err != nil {
		t.Fatal(err)
	}
	recovery, err := RecoverQueue(p, now.Add(time.Minute), DefaultLockStaleAfter)
	if err != nil || recovery != (Recovery{}) {
		t.Fatalf("recovery=%+v err=%v", recovery, err)
	}
	data, err := os.ReadFile(jobPath + ".lock")
	if err != nil {
		t.Fatal(err)
	}
	var lock Lock
	if err := json.Unmarshal(data, &lock); err != nil || lock.PID == 0 || lock.StartedAt.IsZero() {
		t.Fatalf("lock=%+v err=%v", lock, err)
	}
}

func TestRecoverQueueSupportsLegacyTimestampLocks(t *testing.T) {
	p, _ := project.Open(t.TempDir())
	if err := p.Scaffold(); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 9, 3, 0, 0, 0, time.UTC)
	lockPath := filepath.Join(p.QueueDir(), "orphan.json.lock")
	if err := os.WriteFile(lockPath, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	old := now.Add(-time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatal(err)
	}
	recovery, err := RecoverQueue(p, now, DefaultLockStaleAfter)
	if err != nil || recovery.LocksRemoved != 1 {
		t.Fatalf("recovery=%+v err=%v", recovery, err)
	}
}
