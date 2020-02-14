package config

import (
	"time"
)

type KubeproberConfig struct {
	Port int `json:"port" yaml:"port"`
	Tests []TestConfig `json:"tests" yaml:"tests"`
	ReadinessProbe []ProbeConfig `json:"readinessProbe" yaml:"readinessProbe"`
	LivenessProbe []ProbeConfig `json:"livenessProbe" yaml:"livenessProbe"`
}

type ProbeConfig struct {
	Name *string `json:"name" yaml:"name"`
	Description *string `json:"description" yaml:"description"`
	Container *string `json:"container" yaml:"container"`
	PerformTests []string `json:"performTests" yaml:"performTests"`
	SuccessThreshold *int `json:"successThreshold" yaml:"successThreshold"`
	FailureThreshold *int `json:"failureThreshold" yaml:"failureThreshold"`
	InitialDelaySeconds *int `json:"initialDelay" yaml:"initialDelay"`
	Timespan *time.Duration `json:"timespan" yaml:"timespan"`
	PeriodSeconds int `json:"periodSeconds" yaml:"periodSeconds"`
}

type TestConfig struct {
	Name string `json:"name" yaml:"name"`
	Description *string `json:"description" yaml:"description"`
	Regex []string `json:"regex" yaml:"regex"`
	HttpGet *HttpTest `json:"httpGet" yaml:"httpGet"`
}

type HttpTest struct {
	HttpHeaders []HttpHeader `json:"httpHeaders" yaml:"httpHeaders"`
	Path string `json:"path" yaml:"path"`
	Port int `json:"port" yaml:"port"`
	Scheme string `json:"scheme" yaml:"scheme"`
	Sync bool `json:"sync" yaml:"sync"`
}

type HttpHeader struct {
	Name string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}
