package scheduler

import (
	"context"
	"sync"
	"time"
)

// Task 定时任务接口
type Task interface {
	Run(ctx context.Context) error
	Name() string
}

// Scheduler 任务调度器
type Scheduler struct {
	mu       sync.RWMutex
	tasks    []scheduledTask
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
	timezone *time.Location
}

type scheduledTask struct {
	task     Task
	interval time.Duration
	lastRun  time.Time
	runAt    *time.Time // 可选：指定每天的运行时间
}

// New 创建新的调度器
func New(timezone string) (*Scheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return &Scheduler{
		tasks:    make([]scheduledTask, 0),
		stopChan: make(chan struct{}),
		timezone: loc,
	}, nil
}

// AddTask 添加定时任务
func (s *Scheduler) AddTask(task Task, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, scheduledTask{
		task:     task,
		interval: interval,
	})
}

// AddDailyTask 添加每日定时任务
func (s *Scheduler) AddDailyTask(task Task, hour, minute int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().In(s.timezone)
	runTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, s.timezone)
	if runTime.Before(now) {
		runTime = runTime.Add(24 * time.Hour)
	}

	s.tasks = append(s.tasks, scheduledTask{
		task:     task,
		interval: 24 * time.Hour,
		runAt:    &runTime,
	})
}

// Start 启动调度器
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run(ctx)
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
}

func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndRunTasks(ctx)
		}
	}
}

func (s *Scheduler) checkAndRunTasks(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().In(s.timezone)

	for i := range s.tasks {
		task := &s.tasks[i]

		shouldRun := false
		if task.runAt != nil {
			// 每日定时任务
			if now.After(*task.runAt) || now.Equal(*task.runAt) {
				shouldRun = true
				nextRun := task.runAt.Add(24 * time.Hour)
				task.runAt = &nextRun
			}
		} else {
			// 间隔任务
			if task.lastRun.IsZero() || now.Sub(task.lastRun) >= task.interval {
				shouldRun = true
			}
		}

		if shouldRun {
			task.lastRun = now
			go func(t Task) {
				_ = t.Run(ctx)
			}(task.task)
		}
	}
}

// RunNow 立即执行所有任务
func (s *Scheduler) RunNow(ctx context.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, task := range s.tasks {
		_ = task.task.Run(ctx)
	}
}
