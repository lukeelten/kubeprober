package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

func LoadConfiguration(path string) (*KubeproberConfig, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = nil
	var config KubeproberConfig
	if strings.HasSuffix(path, ".json") {
		err = json.Unmarshal(configBytes, &config)
	} else {
		err = yaml.Unmarshal(configBytes, &config)
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *KubeproberConfig) Validate() error {
	if len(config.ReadinessProbe) == 0 && len(config.LivenessProbe) == 0 {
		return errors.New("invalid configuration: no probe defined")
	}

	if config.Port <= 0 {
		return errors.New("invalid configuration: invalid port number")
	}

	if len(config.Tests) == 0 {
		return errors.New("invalid configuration: no tests defined")
	}

	availableTests := make([]string, len(config.Tests))
	for _, test := range config.Tests {
		err := test.Validate()
		if err != nil {
			return err
		}

		if i := sort.SearchStrings(availableTests, test.Name); i == len(availableTests) {
			availableTests = append(availableTests, test.Name)
			sort.Strings(availableTests)
		} else {
			return fmt.Errorf("invalid test '%s': test already exists", test.Name)
		}
	}

	for _, probe := range append(config.LivenessProbe, config.ReadinessProbe...) {
		err := probe.Validate()
		if err != nil {
			return err
		}

		numTests := len(availableTests)
		for _, test := range probe.PerformTests {
			if i := sort.SearchStrings(availableTests, test); i == numTests {
				return fmt.Errorf("invalid probe '%s': test '%s' not found", probe.GetName(), test)
			}
		}
	}

	return nil
}

func (test *TestConfig) Validate() error {
	if len(test.Name) == 0 {
		return errors.New("invalid test: test does not have a name")
	}

	if test.HttpGet != nil && len(test.Regex) > 0 {
		return fmt.Errorf("invalid test '%s': cannot perform http get and regex at the same time", test.Name)
	}

	if test.HttpGet == nil && len(test.Regex) == 0 {
		return fmt.Errorf("invalid test '%s': no check defined", test.Name)
	}

	if test.HttpGet != nil {
		if test.HttpGet.Port <= 0 {
			return fmt.Errorf("invalid test '%s': invalid port number", test.Name)
		}

		if test.HttpGet.Scheme != "HTTP" && test.HttpGet.Scheme != "HTTPS" {
			return fmt.Errorf("invalid test '%s': invalid scheme", test.Name)
		}
	}

	if len(test.Regex) > 0 {
		for _, regex := range test.Regex {
			_, err := regexp.Compile(regex)
			if err != nil {
				return fmt.Errorf("invalid test '%s': cannot compile regex %v", test.Name, regex)
			}
		}
	}

	return nil
}

func (probe *ProbeConfig) GetName() string {
	if probe.Name != nil && len(*probe.Name) > 0 {
		return *probe.Name
	}

	return "Probe #"
}

func (probe *ProbeConfig) Validate() error {
	if len(probe.PerformTests) == 0 {
		return fmt.Errorf("invalid probe '%s': no tests defined", probe.GetName())
	}

	if probe.PeriodSeconds <= 0 {
		return fmt.Errorf("invalid probe '%s': invalid periodSeconds", probe.GetName())
	}

	if probe.SuccessThreshold != nil && *probe.SuccessThreshold <= 0 {
		return fmt.Errorf("invalid probe '%s': invalid success threshold", probe.GetName())
	}

	if probe.SuccessThreshold != nil && *probe.FailureThreshold <= 0 {
		return fmt.Errorf("invalid probe '%s': invalid failure threshold", probe.GetName())
	}

	return nil
}