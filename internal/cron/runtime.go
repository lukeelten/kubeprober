package cron

import "time"

func (cron *ThreadedCron) Start(terminationChannel chan bool) {
	go cron.runtime(terminationChannel)
}

func (cron *ThreadedCron) runtime(closeChannel chan bool) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
			case <- closeChannel:
				return
			case <-ticker.C:
				cron.checkTasks()
		}
	}
}

func (cron *ThreadedCron) checkTasks() {
	now := time.Now()

	for _, task := range cron.Tasks {
		elapsed := now.Sub(task.lastRun)
		if elapsed >= task.Period {
			go task.Task(task.lastRun)
			task.lastRun = now
		}
	}
}

func (cron *ThreadedCron) AddTask(period time.Duration, task CronFunc) {
	cron.Tasks = append(cron.Tasks, Task{Period:period, Task:task})
}