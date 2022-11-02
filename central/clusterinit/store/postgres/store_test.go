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

type ClusterInitBundlesStoreSuite struct {
	suite.Suite
	store  Store
	testDB *pgtest.TestPostgres
}

func TestClusterInitBundlesStore(t *testing.T) {
	suite.Run(t, new(ClusterInitBundlesStoreSuite))
}

func (s *ClusterInitBundlesStoreSuite) SetupSuite() {
	s.T().Setenv(env.PostgresDatastoreEnabled.EnvVar(), "true")

	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	s.testDB = pgtest.ForT(s.T())
	s.store = New(s.testDB.Pool)
}

func (s *ClusterInitBundlesStoreSuite) SetupTest() {
	ctx := sac.WithAllAccess(context.Background())
	tag, err := s.testDB.Exec(ctx, "TRUNCATE cluster_init_bundles CASCADE")
	s.T().Log("cluster_init_bundles", tag)
	s.NoError(err)
}

func (s *ClusterInitBundlesStoreSuite) TearDownSuite() {
	s.testDB.Teardown(s.T())
}

func (s *ClusterInitBundlesStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	initBundleMeta := &storage.InitBundleMeta{}
	s.NoError(testutils.FullInit(initBundleMeta, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundInitBundleMeta, exists, err := store.Get(ctx, initBundleMeta.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundInitBundleMeta)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, initBundleMeta))
	foundInitBundleMeta, exists, err = store.Get(ctx, initBundleMeta.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(initBundleMeta, foundInitBundleMeta)

	initBundleMetaCount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, initBundleMetaCount)
	initBundleMetaCount, err = store.Count(withNoAccessCtx)
	s.NoError(err)
	s.Zero(initBundleMetaCount)

	initBundleMetaExists, err := store.Exists(ctx, initBundleMeta.GetId())
	s.NoError(err)
	s.True(initBundleMetaExists)
	s.NoError(store.Upsert(ctx, initBundleMeta))
	s.ErrorIs(store.Upsert(withNoAccessCtx, initBundleMeta), sac.ErrResourceAccessDenied)

	foundInitBundleMeta, exists, err = store.Get(ctx, initBundleMeta.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(initBundleMeta, foundInitBundleMeta)

	s.NoError(store.Delete(ctx, initBundleMeta.GetId()))
	foundInitBundleMeta, exists, err = store.Get(ctx, initBundleMeta.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundInitBundleMeta)
	s.ErrorIs(store.Delete(withNoAccessCtx, initBundleMeta.GetId()), sac.ErrResourceAccessDenied)

	var initBundleMetas []*storage.InitBundleMeta
	var initBundleMetaIDs []string
	for i := 0; i < 200; i++ {
		initBundleMeta := &storage.InitBundleMeta{}
		s.NoError(testutils.FullInit(initBundleMeta, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		initBundleMetas = append(initBundleMetas, initBundleMeta)
		initBundleMetaIDs = append(initBundleMetaIDs, initBundleMeta.GetId())
	}

	s.NoError(store.UpsertMany(ctx, initBundleMetas))

	initBundleMetaCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, initBundleMetaCount)

	s.NoError(store.DeleteMany(ctx, initBundleMetaIDs))

	initBundleMetaCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(0, initBundleMetaCount)
}
