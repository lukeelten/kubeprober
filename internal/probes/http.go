package probes

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/lukeelten/kubeprober/internal/config"
	"net"
	"net/http"
	"strings"
	"time"
)

type HttpProbe struct {
	Probe

	state *config.KubeproberState
	client *http.Client
	request *http.Request
}

func createHttpProbe(state *config.KubeproberState, test *config.TestConfig) (*HttpProbe, error) {
	testInstance := new(HttpProbe)
	testInstance.state = state
	testInstance.client = createHttpClient()

	if len(test.HttpGet.Scheme) == 0 {
		test.HttpGet.Scheme = "http"
	} else {
		test.HttpGet.Scheme = strings.ToLower(test.HttpGet.Scheme)
	}

	if !strings.HasPrefix(test.HttpGet.Path, "/") {
		test.HttpGet.Path = "/" + test.HttpGet.Path
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s://127.0.0.1:%d%s", test.HttpGet.Scheme, test.HttpGet.Port, test.HttpGet.Path), nil)
	if err != nil {
		return nil, fmt.Errorf("invalid test '%s': %v", test.Name, err)
	}

	for _, header := range test.HttpGet.HttpHeaders {
		request.Header.Set(header.Name, header.Value)
	}
	testInstance.request = request

	return testInstance, nil
}

func (h *HttpProbe) Test(container *config.ContainerStatus) bool {
	ctx := context.Background()
	request := h.request.Clone(ctx)
	response, err := h. client.Do(request)

	if err != nil {
		return false
	}

	return response.StatusCode < 400
}

func createHttpClient() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{InsecureSkipVerify:true},
	}

	return &http.Client{Transport:transport}
}