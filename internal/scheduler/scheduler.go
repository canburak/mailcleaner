package scheduler

import (
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/mailcleaner/mailcleaner/internal/config"
	"github.com/mailcleaner/mailcleaner/internal/rules"
)

// Scheduler manages periodic rule execution
type Scheduler struct {
	scheduler *gocron.Scheduler
	engine    *rules.Engine
	config    *config.Config
}

// New creates a new scheduler
func New(cfg *config.Config, dryRun bool) *Scheduler {
	return &Scheduler{
		scheduler: gocron.NewScheduler(time.Local),
		engine:    rules.NewEngine(cfg, dryRun),
		config:    cfg,
	}
}

// Start begins the scheduled rule execution
func (s *Scheduler) Start() error {
	// Execute immediately on start
	log.Println("Running initial rule execution...")
	if err := s.engine.ExecuteAll(); err != nil {
		log.Printf("Initial execution error: %v", err)
	}

	// Schedule periodic execution
	if s.config.Schedule.Cron != "" {
		log.Printf("Scheduling with cron expression: %s", s.config.Schedule.Cron)
		_, err := s.scheduler.Cron(s.config.Schedule.Cron).Do(s.runRules)
		if err != nil {
			return err
		}
	} else {
		interval := s.config.Schedule.GetInterval()
		log.Printf("Scheduling every %v", interval)
		_, err := s.scheduler.Every(interval).Do(s.runRules)
		if err != nil {
			return err
		}
	}

	s.scheduler.StartBlocking()
	return nil
}

// StartAsync starts the scheduler in a non-blocking way
func (s *Scheduler) StartAsync() {
	// Execute immediately on start
	log.Println("Running initial rule execution...")
	if err := s.engine.ExecuteAll(); err != nil {
		log.Printf("Initial execution error: %v", err)
	}

	// Schedule periodic execution
	if s.config.Schedule.Cron != "" {
		log.Printf("Scheduling with cron expression: %s", s.config.Schedule.Cron)
		s.scheduler.Cron(s.config.Schedule.Cron).Do(s.runRules)
	} else {
		interval := s.config.Schedule.GetInterval()
		log.Printf("Scheduling every %v", interval)
		s.scheduler.Every(interval).Do(s.runRules)
	}

	s.scheduler.StartAsync()
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	s.scheduler.Stop()
}

// RunOnce executes all rules once without scheduling
func (s *Scheduler) RunOnce() error {
	return s.engine.ExecuteAll()
}

func (s *Scheduler) runRules() {
	log.Println("Scheduled rule execution starting...")
	if err := s.engine.ExecuteAll(); err != nil {
		log.Printf("Scheduled execution error: %v", err)
	}
}
