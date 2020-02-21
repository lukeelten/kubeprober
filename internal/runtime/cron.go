package runtime

import (
	"errors"
	"github.com/lukeelten/kubeprober/internal/config"
	"github.com/lukeelten/kubeprober/internal/cron"
	"github.com/lukeelten/kubeprober/internal/probes"
	"log"
	"sync"
	"time"
)

func SetupTasks(state *config.KubeproberState) error {
	timer := new(cron.ThreadedCron)

	containers := make(map[string]*config.ContainerStatus, len(state.Pod.Spec.Containers))
	for _, container := range state.Pod.Spec.Containers {
		containerStatus := new(config.ContainerStatus)
		containerStatus.Name = container.Name
		containerStatus.Readiness = config.CheckStatus{
			Lock:      sync.RWMutex{},
			Status:    true,
			LastCheck: time.Now(),
		}
		containerStatus.Liveness = config.CheckStatus{
			Lock:      sync.RWMutex{},
			Status:    true,
			LastCheck: time.Now(),
		}

		containers[container.Name] = containerStatus
	}

	tests := make(map[string]probes.Probe)
	for _, test := range state.Config.Tests {
		var err error
		tests[test.Name], err = probes.CreateProbe(state, &test)
		if err != nil {
			return err
		}
	}

	for _, probe := range state.Config.LivenessProbe {
		if probe.Container == nil {
			return errors.New("invalid container")
		}

		container, ok := containers[*probe.Container]
		if !ok {
			return errors.New("cannot find container")
		}

		for _, test := range probe.PerformTests {
			realTest, ok := tests[test]
			if !ok {
				return errors.New("cannot find test")
			}

			timer.AddTask(time.Duration(probe.PeriodSeconds) * time.Second, makeLivenessFunc(container, realTest))
		}
	}

	for _, probe := range state.Config.ReadinessProbe {
		if probe.Container == nil {
			return errors.New("invalid container")
		}

		container, ok := containers[*probe.Container]
		if !ok {
			return errors.New("cannot find container")
		}

		for _, test := range probe.PerformTests {
			realTest, ok := tests[test]
			if !ok {
				return errors.New("cannot find test")
			}

			timer.AddTask(time.Duration(probe.PeriodSeconds) * time.Second, makeReadinessFunc(container, realTest))
		}
	}

	timer.Start(state.CreateTerminationChannel())

	return nil
}

func makeLivenessFunc(container *config.ContainerStatus, test probes.Probe) cron.CronFunc {
	return func(lastCall time.Time) {
		result := test.Test(container, lastCall)
		container.SetAlive(result)
		log.Printf("Container %s is alive %v", container.Name, result)
	}
}

func makeReadinessFunc(container *config.ContainerStatus, test probes.Probe) cron.CronFunc {
	return func(lastCall time.Time) {
		result := test.Test(container, lastCall)
		container.SetReady(result)
		log.Printf("Container %s is ready %v", container.Name, result)
	}
}