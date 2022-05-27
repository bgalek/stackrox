package datastore

import (
	"context"

	store "github.com/stackrox/rox/central/complianceoperator/scans/store"
	"github.com/stackrox/rox/central/complianceoperator/scans/store/postgres"
	"github.com/stackrox/rox/central/complianceoperator/scans/store/rocksdb"
	"github.com/stackrox/rox/central/globaldb"
	"github.com/stackrox/rox/pkg/features"
	"github.com/stackrox/rox/pkg/sync"
	"github.com/stackrox/rox/pkg/utils"
)

var (
	once sync.Once
	ds   DataStore
)

// Singleton returns the singleton datastore
func Singleton() DataStore {
	once.Do(func() {
		var storage store.Store
		if features.PostgresDatastore.Enabled() {
			storage = postgres.New(context.TODO(), globaldb.GetPostgres())
		} else {
			var err error
			storage, err = rocksdb.New(globaldb.GetRocksDB())
			utils.CrashOnError(err)
		}

		ds = NewDatastore(storage)
	})
	return ds
}
