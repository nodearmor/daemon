package nodearmord

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nodearmor/daemon/internal/controller"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type signalCh chan os.Signal

var ctrl controller.Controller

func Run() {
	// Bind os signals to stop channel
	var stop = make(signalCh)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)

	// goroutine waitgroup
	var wg sync.WaitGroup

	// Setup logs
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	controllerLog := log.With().Str("component", "controller").Logger()
	controller.SetLogger(&controllerLog)

	LoadConfig()

	err := ctrl.Connect(config.GetString("ControllerURL"))
	if err != nil {
		log.Fatal().Err(err).Msg("Controller connection failed")
	}

	// Begin processing controller messages
	ctrl.Start(wg, stop)

	StartRPCServer(&wg, stop)

	// Wait for all goroutines
	wg.Wait()
}
