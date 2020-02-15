package runtime

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lukeelten/kubeprober/internal/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(state *config.KubeproberState) error {
	go shutdownOnSyscall(state)
	go listenErrorChannel(state)

	err := SetupTasks(state)
	if err != nil {
		return err
	}

	state.Engine.GET("/health", func(context *gin.Context) {
		context.Status(http.StatusOK)
	})

	liveness := state.Engine.Group("/liveness")
	liveness.GET("/", func(context *gin.Context) {
		if state.IsAlive() {
			context.Status(http.StatusOK)
		} else {
			context.Status(http.StatusInternalServerError)
		}
	})

	liveness.GET("/:container", func(context *gin.Context) {
		containerName := context.Param("container")
		if len(containerName) == 0 {
			_ = context.AbortWithError(http.StatusBadRequest, errors.New("invalid container name"))
		}

		if container, ok := state.Container[containerName]; ok {
			if container.IsAlive() {
				context.Status(http.StatusOK)
			} else {
				context.Status(http.StatusInternalServerError)
			}
		} else {
			_ = context.AbortWithError(http.StatusBadRequest, errors.New("container not found"))
		}
	})

	readiness := state.Engine.Group("/readiness")
	readiness.GET("/", func(context *gin.Context) {
		if state.IsReady() {
			context.Status(http.StatusOK)
		} else {
			context.Status(http.StatusInternalServerError)
		}
	})

	readiness.GET("/:container", func(context *gin.Context) {
		containerName := context.Param("container")
		if len(containerName) == 0 {
			_ = context.AbortWithError(http.StatusBadRequest, errors.New("invalid container name"))
		}

		if container, ok := state.Container[containerName]; ok {
			if container.IsReady() {
				context.Status(http.StatusOK)
			} else {
				context.Status(http.StatusInternalServerError)
			}
		} else {
			_ = context.AbortWithError(http.StatusBadRequest, errors.New("container not found"))
		}
	})


	return createServer(state).ListenAndServe()
}


func shutdownOnSyscall(state *config.KubeproberState) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-signalChannel
	log.Printf("Received signal: %v", sig)
	state.Shutdown()
}

func listenErrorChannel(state *config.KubeproberState) {
	termChannel := state.CreateTerminationChannel()

	for {
		select {
			case err := <-state.ErrorChannel:
				log.Printf("Found error: %v", err)
			case <-termChannel:
				return
		}
	}
}

func createServer(state *config.KubeproberState) *http.Server {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", state.Config.Port),
		Handler: state.Engine,
	}

	termChannel := state.CreateTerminationChannel()
	go func() {
		<- termChannel
		_ = server.Close()
	}()

	return server
}