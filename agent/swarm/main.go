package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "bitbucket.org/stack-rox/apollo/pkg/registries/all"
	_ "bitbucket.org/stack-rox/apollo/pkg/scanners/all"

	"bitbucket.org/stack-rox/apollo/agent/swarm/listener"
	"bitbucket.org/stack-rox/apollo/agent/swarm/orchestrator"
	"bitbucket.org/stack-rox/apollo/pkg/agent"
	"bitbucket.org/stack-rox/apollo/pkg/api/generated/api/v1"
	"bitbucket.org/stack-rox/apollo/pkg/benchmarks"
	"bitbucket.org/stack-rox/apollo/pkg/registries"
	"bitbucket.org/stack-rox/apollo/pkg/scanners"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	a := initializeAgent()

	a.Start()

	for {
		select {
		case sig := <-sigs:
			a.Logger.Infof("Caught %s signal", sig)
			a.Stop()
			a.Logger.Info("Swarm Agent terminated")
			return
		}
	}
}

func initializeAgent() *agent.Agent {
	a := agent.New()
	var err error

	a.Listener, err = listener.New()
	if err != nil {
		panic(err)
	}

	a.Orchestrator, err = orchestrator.New()
	if err != nil {
		panic(err)
	}

	a.BenchScheduler, err = benchmarks.NewSchedulerClient(a.Orchestrator, a.ApolloEndpoint, a.AdvertisedEndpoint, a.Image, a.ClusterID)
	if err != nil {
		panic(err)
	}
	a.ScannerPoller = scanners.NewScannersClient(a.ApolloEndpoint, a.ClusterID)
	a.RegistryPoller = registries.NewRegistriesClient(a.ApolloEndpoint, a.ClusterID)

	a.ServiceRegistrationFunc = registerAPIServices

	a.Logger.Info("Swarm Agent Initialized")
	return a
}

func registerAPIServices(a *agent.Agent) {
	a.Server.Register(benchmarks.NewBenchmarkRelayService(benchmarks.NewLRURelayer(v1.NewBenchmarkResultsServiceClient(a.Conn), a.ClusterID)))
	a.Logger.Info("API services registered")
}
