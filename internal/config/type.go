package config

import (
	"time"
)

type KubeproberConfig struct {
	Tests []TestConfig `json:"tests" yaml:"tests"`
	ReadinessProbe []ProbeConfig `json:"readinessProbe" yaml:"readinessProbe"`
	LivenessProbe []ProbeConfig `json:"livenessProbe" yaml:"livenessProbe"`
}

type ProbeConfig struct {
	Name *string
	Description *string
	Container string
	PerformTests []string
	SuccessThreshold *int
	FailureThreshold *int
	InitialDelay *time.Duration
	Timespan *time.Duration
}

type TestConfig struct {
	Name string
	Description *string
	Regex []string
	HttpGet *HttpTest
}

type HttpTest struct {
	HttpHeaders []HttpHeader
	Path string
	Port int
	Scheme string
	Sync bool
}

type HttpHeader struct {
	Name string
	Value string
}
