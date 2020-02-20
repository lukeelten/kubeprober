package cron

import "time"

type ThreadedCron struct {
	Tasks []Task
}

type Task struct {
	Period time.Duration
	lastRun time.Time

	Task CronFunc
}

type CronFunc func(lastCall time.Time)