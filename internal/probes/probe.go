package probes

import (
	"errors"
	"github.com/lukeelten/kubeprober/internal/config"
)

type Probe interface {
	Test(container *config.ContainerStatus) bool
}

func CreateProbe(state *config.KubeproberState, test *config.TestConfig) (Probe, error) {
	if len(test.Regex) > 0 {
		return createRegexProbe(state, test)
	}

	if test.HttpGet != nil {
		return createHttpProbe(state, test)
	}

	return nil, errors.New("cannot create probe: invalid test config")
}
