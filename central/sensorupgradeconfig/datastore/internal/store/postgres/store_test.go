// Code generated by pg-bindings generator. DO NOT EDIT.

//go:build sql_integration

package postgres

import (
	"context"
	"testing"

	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/postgres/pgtest"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/testutils"
	"github.com/stretchr/testify/suite"
)

type SensorUpgradeConfigsStoreSuite struct {
	suite.Suite
	store  Store
	testDB *pgtest.TestPostgres
}

func TestSensorUpgradeConfigsStore(t *testing.T) {
	suite.Run(t, new(SensorUpgradeConfigsStoreSuite))
}

func (s *SensorUpgradeConfigsStoreSuite) SetupTest() {
	s.T().Setenv(env.PostgresDatastoreEnabled.EnvVar(), "true")

	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	s.testDB = pgtest.ForT(s.T())
	s.store = New(s.testDB.Pool)
}

func (s *SensorUpgradeConfigsStoreSuite) TearDownTest() {
	s.testDB.Teardown(s.T())
}

func (s *SensorUpgradeConfigsStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	sensorUpgradeConfig := &storage.SensorUpgradeConfig{}
	s.NoError(testutils.FullInit(sensorUpgradeConfig, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundSensorUpgradeConfig, exists, err := store.Get(ctx)
	s.NoError(err)
	s.False(exists)
	s.Nil(foundSensorUpgradeConfig)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, sensorUpgradeConfig))
	foundSensorUpgradeConfig, exists, err = store.Get(ctx)
	s.NoError(err)
	s.True(exists)
	s.Equal(sensorUpgradeConfig, foundSensorUpgradeConfig)

	foundSensorUpgradeConfig, exists, err = store.Get(ctx)
	s.NoError(err)
	s.True(exists)
	s.Equal(sensorUpgradeConfig, foundSensorUpgradeConfig)

	s.NoError(store.Delete(ctx))
	foundSensorUpgradeConfig, exists, err = store.Get(ctx)
	s.NoError(err)
	s.False(exists)
	s.Nil(foundSensorUpgradeConfig)

	s.ErrorIs(store.Delete(withNoAccessCtx), sac.ErrResourceAccessDenied)

	sensorUpgradeConfig = &storage.SensorUpgradeConfig{}
	s.NoError(testutils.FullInit(sensorUpgradeConfig, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))
	s.NoError(store.Upsert(ctx, sensorUpgradeConfig))

	foundSensorUpgradeConfig, exists, err = store.Get(ctx)
	s.NoError(err)
	s.True(exists)
	s.Equal(sensorUpgradeConfig, foundSensorUpgradeConfig)

	sensorUpgradeConfig = &storage.SensorUpgradeConfig{}
	s.NoError(testutils.FullInit(sensorUpgradeConfig, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))
	s.NoError(store.Upsert(ctx, sensorUpgradeConfig))

	foundSensorUpgradeConfig, exists, err = store.Get(ctx)
	s.NoError(err)
	s.True(exists)
	s.Equal(sensorUpgradeConfig, foundSensorUpgradeConfig)
}
