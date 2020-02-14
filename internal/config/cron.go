package config

import (
	"regexp"
	"sync"
)

func (state *KubeproberState) SetupTasks() error {

	containers := make(map[string]*ContainerStatus, len(state.Pod.Spec.Containers))
	for _, container := range state.Pod.Spec.Containers {
		containerStatus := new(ContainerStatus)
		containerStatus.ReadinessMutex = new(sync.RWMutex)
		containerStatus.LivenessMutex = new(sync.RWMutex)
		containerStatus.Liveness = true
		containerStatus.Readiness = true
		containerStatus.Name = container.Name
		containers[container.Name] = containerStatus
	}

	tests := make(map[string]TestInstance)
	for _, test := range state.Config.Tests {
		if len(test.Regex) > 0 {
			testInstance := new(RegexTestInstance)
			testInstance.Regex = make([]*regexp.Regexp, len(test.Regex))

			for _, regexTest := range test.Regex {
				regex, err := regexp.Compile(regexTest)
				if err != nil {
					return err
				}

				testInstance.Regex = append(testInstance.Regex, regex)
			}

			tests[test.Name] = testInstance
		}

		// http get test


	}

	// setup liveness

	// setup readiness


	return nil
}

type TestInstance interface {
	Test(container *ContainerStatus)
}

type RegexTestInstance struct {
	TestInstance

	Regex []*regexp.Regexp
	State *KubeproberState
}

type HttpTestInstance struct {
	TestInstance

	State *KubeproberState
}