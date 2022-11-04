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

type ServiceIdentitiesStoreSuite struct {
	suite.Suite
	store  Store
	testDB *pgtest.TestPostgres
}

func TestServiceIdentitiesStore(t *testing.T) {
	suite.Run(t, new(ServiceIdentitiesStoreSuite))
}

func (s *ServiceIdentitiesStoreSuite) SetupSuite() {
	s.T().Setenv(env.PostgresDatastoreEnabled.EnvVar(), "true")

	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	s.testDB = pgtest.ForT(s.T())
	s.store = New(s.testDB.Pool)
}

func (s *ServiceIdentitiesStoreSuite) SetupTest() {
	ctx := sac.WithAllAccess(context.Background())
	tag, err := s.testDB.Exec(ctx, "TRUNCATE service_identities CASCADE")
	s.T().Log("service_identities", tag)
	s.NoError(err)
}

func (s *ServiceIdentitiesStoreSuite) TearDownSuite() {
	s.testDB.Teardown(s.T())
}

func (s *ServiceIdentitiesStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	serviceIdentity := &storage.ServiceIdentity{}
	s.NoError(testutils.FullInit(serviceIdentity, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundServiceIdentity, exists, err := store.Get(ctx, serviceIdentity.GetSerialStr())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundServiceIdentity)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, serviceIdentity))
	foundServiceIdentity, exists, err = store.Get(ctx, serviceIdentity.GetSerialStr())
	s.NoError(err)
	s.True(exists)
	s.Equal(serviceIdentity, foundServiceIdentity)

	serviceIdentityCount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, serviceIdentityCount)
	serviceIdentityCount, err = store.Count(withNoAccessCtx)
	s.NoError(err)
	s.Zero(serviceIdentityCount)

	serviceIdentityExists, err := store.Exists(ctx, serviceIdentity.GetSerialStr())
	s.NoError(err)
	s.True(serviceIdentityExists)
	s.NoError(store.Upsert(ctx, serviceIdentity))
	s.ErrorIs(store.Upsert(withNoAccessCtx, serviceIdentity), sac.ErrResourceAccessDenied)

	foundServiceIdentity, exists, err = store.Get(ctx, serviceIdentity.GetSerialStr())
	s.NoError(err)
	s.True(exists)
	s.Equal(serviceIdentity, foundServiceIdentity)

	s.NoError(store.Delete(ctx, serviceIdentity.GetSerialStr()))
	foundServiceIdentity, exists, err = store.Get(ctx, serviceIdentity.GetSerialStr())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundServiceIdentity)
	s.ErrorIs(store.Delete(withNoAccessCtx, serviceIdentity.GetSerialStr()), sac.ErrResourceAccessDenied)

	var serviceIdentitys []*storage.ServiceIdentity
	var serviceIdentityIDs []string
	for i := 0; i < 200; i++ {
		serviceIdentity := &storage.ServiceIdentity{}
		s.NoError(testutils.FullInit(serviceIdentity, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		serviceIdentitys = append(serviceIdentitys, serviceIdentity)
		serviceIdentityIDs = append(serviceIdentityIDs, serviceIdentity.GetSerialStr())
	}

	s.NoError(store.UpsertMany(ctx, serviceIdentitys))
	allServiceIdentity, err := store.GetAll(ctx)
	s.NoError(err)
	s.ElementsMatch(serviceIdentitys, allServiceIdentity)

	serviceIdentityCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, serviceIdentityCount)

	s.NoError(store.DeleteMany(ctx, serviceIdentityIDs))

	serviceIdentityCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(0, serviceIdentityCount)
}
