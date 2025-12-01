package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type mockTask struct {
	name     string
	runCount int32
}

func (m *mockTask) Name() string {
	return m.name
}

func (m *mockTask) Run(ctx context.Context) error {
	atomic.AddInt32(&m.runCount, 1)
	return nil
}

func TestScheduler_AddTask(t *testing.T) {
	s, err := New("UTC")
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	task := &mockTask{name: "test-task"}
	s.AddTask(task, 1*time.Hour)

	if len(s.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(s.tasks))
	}
}

func TestScheduler_RunNow(t *testing.T) {
	s, err := New("UTC")
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	task := &mockTask{name: "test-task"}
	s.AddTask(task, 1*time.Hour)

	ctx := context.Background()
	s.RunNow(ctx)

	if task.runCount != 1 {
		t.Errorf("Task should have run once, ran %d times", task.runCount)
	}
}

func TestScheduler_StartStop(t *testing.T) {
	s, err := New("UTC")
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)

	// 验证调度器正在运行
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		t.Error("Scheduler should be running")
	}

	s.Stop()

	s.mu.RLock()
	running = s.running
	s.mu.RUnlock()

	if running {
		t.Error("Scheduler should be stopped")
	}
}
