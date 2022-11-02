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

type NodeCvesStoreSuite struct {
	suite.Suite
	store  Store
	testDB *pgtest.TestPostgres
}

func TestNodeCvesStore(t *testing.T) {
	suite.Run(t, new(NodeCvesStoreSuite))
}

func (s *NodeCvesStoreSuite) SetupSuite() {
	s.T().Setenv(env.PostgresDatastoreEnabled.EnvVar(), "true")

	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	s.testDB = pgtest.ForT(s.T())
	s.store = New(s.testDB.Pool)
}

func (s *NodeCvesStoreSuite) SetupTest() {
	ctx := sac.WithAllAccess(context.Background())
	tag, err := s.testDB.Exec(ctx, "TRUNCATE node_cves CASCADE")
	s.T().Log("node_cves", tag)
	s.NoError(err)
}

func (s *NodeCvesStoreSuite) TearDownSuite() {
	s.testDB.Teardown(s.T())
}

func (s *NodeCvesStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	nodeCVE := &storage.NodeCVE{}
	s.NoError(testutils.FullInit(nodeCVE, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundNodeCVE, exists, err := store.Get(ctx, nodeCVE.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundNodeCVE)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, nodeCVE))
	foundNodeCVE, exists, err = store.Get(ctx, nodeCVE.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(nodeCVE, foundNodeCVE)

	nodeCVECount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, nodeCVECount)
	nodeCVECount, err = store.Count(withNoAccessCtx)
	s.NoError(err)
	s.Zero(nodeCVECount)

	nodeCVEExists, err := store.Exists(ctx, nodeCVE.GetId())
	s.NoError(err)
	s.True(nodeCVEExists)
	s.NoError(store.Upsert(ctx, nodeCVE))
	s.ErrorIs(store.Upsert(withNoAccessCtx, nodeCVE), sac.ErrResourceAccessDenied)

	foundNodeCVE, exists, err = store.Get(ctx, nodeCVE.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(nodeCVE, foundNodeCVE)

	s.NoError(store.Delete(ctx, nodeCVE.GetId()))
	foundNodeCVE, exists, err = store.Get(ctx, nodeCVE.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundNodeCVE)
	s.NoError(store.Delete(withNoAccessCtx, nodeCVE.GetId()))

	var nodeCVEs []*storage.NodeCVE
	var nodeCVEIDs []string
	for i := 0; i < 200; i++ {
		nodeCVE := &storage.NodeCVE{}
		s.NoError(testutils.FullInit(nodeCVE, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		nodeCVEs = append(nodeCVEs, nodeCVE)
		nodeCVEIDs = append(nodeCVEIDs, nodeCVE.GetId())
	}

	s.NoError(store.UpsertMany(ctx, nodeCVEs))

	nodeCVECount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, nodeCVECount)

	s.NoError(store.DeleteMany(ctx, nodeCVEIDs))

	nodeCVECount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(0, nodeCVECount)
}
