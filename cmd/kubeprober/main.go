package main

import (
	"github.com/lukeelten/kubeprober/internal/config"
	"github.com/lukeelten/kubeprober/internal/runtime"
	"log"
	"net/http"
	"time"
)

func main() {
	kubeprober, err := config.NewKubeprober("example-config.yaml")

	if err != nil {
		log.Fatal(err)
	}

	err = runtime.Run(kubeprober)
	if err != http.ErrServerClosed {
		kubeprober.Shutdown() // Send signal to running goroutines
		log.Fatalf("Received error: %v", err)
	}

	// Wait for all go routines to terminate
	time.Sleep(50  * time.Millisecond)
}