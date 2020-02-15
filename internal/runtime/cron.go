package runtime

import (
	"github.com/lukeelten/kubeprober/internal/config"
	"github.com/lukeelten/kubeprober/internal/probes"
	"sync"
	"time"
)

func SetupTasks(state *config.KubeproberState) error {

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


	// setup liveness

	// setup readiness


	return nil
}
