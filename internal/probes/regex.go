package probes

import (
	"github.com/lukeelten/kubeprober/internal/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
	"time"
)

type RegexProbe struct {
	Probe

	state *config.KubeproberState
	regex []*regexp.Regexp
}

func createRegexProbe(state *config.KubeproberState, test *config.TestConfig) (*RegexProbe, error) {
	testInstance := new(RegexProbe)
	testInstance.state = state
	testInstance.regex = make([]*regexp.Regexp, len(test.Regex))

	for _, regexTest := range test.Regex {
		regex, err := regexp.Compile(regexTest)
		if err != nil {
			return nil, err
		}

		testInstance.regex = append(testInstance.regex, regex)
	}

	return testInstance, nil
}

func (test *RegexProbe) Test(container *config.ContainerStatus) bool {
	since := time.Now()
	// @todo

	options := &v1.PodLogOptions{
		Container: container.Name,
		SinceTime: &metav1.Time{since},
	}
	req := test.state.Kubernetes.CoreV1().Pods(test.state.Pod.Namespace).GetLogs(test.state.Pod.Name, options)

	result := req.Do()
	rawResponse, err := result.Raw()
	if err != nil {
		return false
	}

	lines := make([]string, 0)
	var builder strings.Builder
	for _, char := range rawResponse {
		if char == '\n' {
			lines = append(lines, builder.String())
			builder.Reset()
			continue
		}

		if char == '\r' {
			continue
		}

		builder.WriteByte(char)
	}

	if builder.Len() > 0 {
		lines = append(lines, builder.String())
	}

	for _, line := range lines {
		for _, regex := range test.regex {
			if regex.MatchString(line) {
				return false
			}
		}
	}

	return true
}


