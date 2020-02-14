package main

import (
	"github.com/lukeelten/kubeprober/internal/config"
	"log"
	"net/http"
	"time"
)

func main() {
	kubeprober, err := config.NewKubeprober("example-config.yaml")

	if err != nil {
		log.Fatal(err)
	}

	err = kubeprober.Run()
	if err != http.ErrServerClosed {
		kubeprober.Shutdown() // Send signal to running goroutines
		log.Fatalf("Received error: %v", err)
	}

	// Wait for all go routines to terminate
	time.Sleep(50  * time.Millisecond)
}