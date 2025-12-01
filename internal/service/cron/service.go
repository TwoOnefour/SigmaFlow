package cron

import (
	"github.com/robfig/cron/v3"
	"log"
)

type Service struct {
	Task []cron.EntryID
	c    *cron.Cron
}

func NewService() *Service {
	c := cron.New()
	c.Start()
	return &Service{
		c:    c,
		Task: make([]cron.EntryID, 0),
	}
}

func (s *Service) AddCron(crontime string, cmd func()) error {
	id, err := s.c.AddFunc(crontime, cmd)
	log.Println("[cron.Service] 添加任务")
	if err != nil {
		return err
	}
	s.Task = append(s.Task, id)
	return nil
}
