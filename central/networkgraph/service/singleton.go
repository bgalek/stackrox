package service

import (
	clusterDataStore "github.com/stackrox/rox/central/cluster/datastore"
	deploymentDataStore "github.com/stackrox/rox/central/deployment/datastore"
	graphConfigDataStore "github.com/stackrox/rox/central/networkgraph/config/datastore"
	networkEntityDatastore "github.com/stackrox/rox/central/networkgraph/entity/datastore"
	nfDS "github.com/stackrox/rox/central/networkgraph/flow/datastore"
	"github.com/stackrox/rox/pkg/sync"
)

var (
	once sync.Once

	as Service
)

func initialize() {
	as = New(nfDS.Singleton(),
		networkEntityDatastore.Singleton(),
		deploymentDataStore.Singleton(),
		clusterDataStore.Singleton(),
		graphConfigDataStore.Singleton())
}

// Singleton provides the instance of the Service interface to register.
func Singleton() Service {
	once.Do(initialize)
	return as
}
