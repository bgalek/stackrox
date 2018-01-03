package main

import (
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"bitbucket.org/stack-rox/apollo/apollo/db"
	"bitbucket.org/stack-rox/apollo/apollo/db/boltdb"
	"bitbucket.org/stack-rox/apollo/apollo/db/inmem"
	"bitbucket.org/stack-rox/apollo/apollo/detection/image_processor"
	"bitbucket.org/stack-rox/apollo/apollo/notifications"
	"bitbucket.org/stack-rox/apollo/apollo/service"
	"bitbucket.org/stack-rox/apollo/pkg/grpc"
	"bitbucket.org/stack-rox/apollo/pkg/logging"
	_ "bitbucket.org/stack-rox/apollo/pkg/notifications/notifiers/all"
	_ "bitbucket.org/stack-rox/apollo/pkg/registries/all"
	_ "bitbucket.org/stack-rox/apollo/pkg/scanners/all"
)

var (
	log = logging.New("main")
)

func main() {
	apollo := newApollo()

	var err error
	persistence, err := boltdb.NewWithDefaults("/var/lib/")
	if err != nil {
		panic(err)
	}
	apollo.database = inmem.New(persistence)
	if err = apollo.database.Load(); err != nil {
		log.Fatal(err)
	}

	apollo.imageProcessor, err = imageprocessor.New(apollo.database)
	if err != nil {
		panic(err)
	}

	apollo.notificationProcessor, err = notifications.NewNotificationProcessor(apollo.database)
	if err != nil {
		panic(err)
	}
	go apollo.notificationProcessor.Start()

	go apollo.startGRPCServer()

	apollo.processForever()
}

type apollo struct {
	signalsC              chan (os.Signal)
	imageProcessor        *imageprocessor.ImageProcessor
	notificationProcessor *notifications.Processor
	database              db.Storage
	server                grpc.API
}

func newApollo() *apollo {
	apollo := &apollo{}

	apollo.signalsC = make(chan os.Signal, 1)
	signal.Notify(apollo.signalsC, os.Interrupt)
	signal.Notify(apollo.signalsC, syscall.SIGINT, syscall.SIGTERM)

	return apollo
}

func (a *apollo) startGRPCServer() {
	a.server = grpc.NewAPIWithUI()
	a.server.Register(service.NewAgentEventService(a.imageProcessor, a.notificationProcessor, a.database))
	a.server.Register(service.NewAlertService(a.database))
	a.server.Register(service.NewBenchmarkService(a.database))
	a.server.Register(service.NewBenchmarkScheduleService(a.database))
	a.server.Register(service.NewBenchmarkResultsService(a.database))
	a.server.Register(service.NewBenchmarkTriggerService(a.database))
	a.server.Register(service.NewClusterService(a.database))
	a.server.Register(service.NewDeploymentService(a.database))
	a.server.Register(service.NewImageService(a.database))
	a.server.Register(service.NewNotifierService(a.database, a.notificationProcessor))
	a.server.Register(service.NewPingService())
	a.server.Register(service.NewPolicyService(a.database, a.imageProcessor))
	a.server.Register(service.NewRegistryService(a.database, a.imageProcessor))
	a.server.Register(service.NewScannerService(a.database, a.imageProcessor))
	a.server.Start()
}

func (a *apollo) processForever() {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Caught panic in process loop; restarting. Stack: %s", string(debug.Stack()))
			a.processForever()
		}
	}()

	for {
		select {
		case sig := <-a.signalsC:
			log.Infof("Caught %s signal", sig)
			log.Infof("Apollo terminated")
			return
		}
	}
}
