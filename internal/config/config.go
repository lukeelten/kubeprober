package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"gopkg.in/yaml.v2"
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

	if len(config.Tests) == 0 {
		return errors.New("invalid configuration: no tests defined")
	}

	for _, test := range config.Tests {
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
	}

	return nil
}