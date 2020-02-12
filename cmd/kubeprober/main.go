package main

import (
	"fmt"
	config2 "github.com/lukeelten/kubeprober/internal/config"
)

func main() {
	config, err := config2.LoadConfiguration("example-config.yaml")
	if err != nil {
		panic(err)
	}

	err = config.Validate()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v", *config.LivenessProbe[0].Timespan)
}