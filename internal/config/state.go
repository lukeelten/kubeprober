package config

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"sync"
	"time"
)

type TerminationChannel chan bool
type CheckFunc func() bool
type Checkable interface {
	IsAlive() bool
	IsReady() bool
}

type KubeproberState struct {
	Checkable

	Config *KubeproberConfig

	Kubernetes *kubernetes.Clientset
	KubernetesConfig *rest.Config
	Engine *gin.Engine

	Pod *v1.Pod

	TerminationChannels []TerminationChannel
	ErrorChannel chan error

	Container map[string]Checkable
}

type ContainerStatus struct {
	Checkable

	Name string
	Liveness CheckStatus
	Readiness CheckStatus
}

type CheckStatus struct {
	Lock sync.RWMutex
	Status bool
	LastCheck time.Time
}

func NewKubeprober(configFile string) (*KubeproberState, error) {
	config, err := LoadConfiguration(configFile)
	if err != nil {
		return nil, err
	}

	if err = config.Validate(); err != nil {
		return nil, err
	}

	kubernetesConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	kubernetesClient, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, err
	}


	engine := gin.Default()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	namespace := getNamespace()
	if len(namespace) == 0 {
		return nil, errors.New("cannot determine namespace")
	}

	podName := getPodname()
	if len(podName) == 0 {
		return nil, errors.New("cannot determine pod name")
	}

	pod, err := kubernetesClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}



	return &KubeproberState{
		Config:           config,
		Kubernetes:       kubernetesClient,
		KubernetesConfig: kubernetesConfig,
		Engine:           engine,
		TerminationChannels: make([]TerminationChannel, 0),
		ErrorChannel: make(chan error, 10),
		Pod: pod,
		Container: make(map[string]Checkable, 0),
	}, nil
}

func (state *KubeproberState) Shutdown() {
	for _, terminationChannel := range state.TerminationChannels {
		terminationChannel <- true
	}
}

func (state *KubeproberState) CreateTerminationChannel() TerminationChannel {
	termChannel := make(TerminationChannel, 1)
	state.TerminationChannels = append(state.TerminationChannels, termChannel)
	return termChannel
}

func (state *KubeproberState) IsAlive() bool {
	result := true
	for _, container := range state.Container {
		if container == nil {
			continue
		}

		result = result && container.IsAlive()
	}

	return result
}

func (state *KubeproberState) IsReady() bool {
	result := true
	for _, container := range state.Container {
		if container == nil {
			continue
		}

		result = result && container.IsReady()
	}

	return result
}

func (container *ContainerStatus) IsAlive() bool {
	container.Liveness.Lock.RLock()
	defer container.Liveness.Lock.RUnlock()
	return container.Liveness.Status
}

func (container *ContainerStatus) IsReady() bool {
	container.Readiness.Lock.RLock()
	defer container.Readiness.Lock.RUnlock()
	return container.Readiness.Status
}

func (container *ContainerStatus) SetAlive(alive bool) {
	container.Liveness.Lock.Lock()
	defer container.Liveness.Lock.Unlock()
	container.Liveness.Status = alive
}

func (container *ContainerStatus) SetReady(ready bool) {
	container.Readiness.Lock.Lock()
	defer container.Readiness.Lock.Unlock()
	container.Readiness.Status = ready
}

func getNamespace() string {
	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if _, err := os.Stat(namespaceFile); err == nil {
		namespace, err := ioutil.ReadFile(namespaceFile)
		if err == nil {
			return string(namespace)
		}
	}

	namespace, exists := os.LookupEnv("NAMESPACE")
	if exists {
		return namespace
	}

	return ""
}

func getPodname() string {
	podName, exists := os.LookupEnv("POD_NAME")
	if exists {
		return podName
	}

	podName, err := os.Hostname()
	if err == nil {
		return podName
	}

	return ""
}
