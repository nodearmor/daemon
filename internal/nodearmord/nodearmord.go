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

var ctrlTransport controller.WebsocketTransport
var ctrl = controller.NewJsonAPI(&ctrlTransport)

func RunController(wg *sync.WaitGroup, stop chan os.Signal) {
	wg.Add(1)

	go func() {
		<-stop
		ctrlTransport.Disconnect()
	}()

	go func() {
		defer wg.Done()
		ctrl.Run()
	}()
}

func Run() {
	// Bind os signals to stop channel
	var stop = make(signalCh)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)

	// goroutine waitgroup
	var wg sync.WaitGroup

	// Setup logs
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	LoadConfig()

	err := ctrlTransport.Connect(config.GetString("ControllerURL"))
	if err != nil {
		log.Fatal().Err(err).Msg("Controller connection failed")
	}

	// Begin processing controller
	RunController(&wg, stop)
	// Start RPC server
	StartRPCServer(&wg, stop)

	ctrl.Init()

	// Wait for all goroutines
	wg.Wait()
}
