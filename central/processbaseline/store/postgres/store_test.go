// Code generated by pg-bindings generator. DO NOT EDIT.

//go:build sql_integration

package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/features"
	"github.com/stackrox/rox/pkg/postgres/pgtest"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/testutils"
	"github.com/stackrox/rox/pkg/testutils/envisolator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProcessbaselinesStoreSuite struct {
	suite.Suite
	envIsolator *envisolator.EnvIsolator
	store       Store
	pool        *pgxpool.Pool
}

func TestProcessbaselinesStore(t *testing.T) {
	suite.Run(t, new(ProcessbaselinesStoreSuite))
}

func (s *ProcessbaselinesStoreSuite) SetupTest() {
	s.envIsolator = envisolator.NewEnvIsolator(s.T())
	s.envIsolator.Setenv(features.PostgresDatastore.EnvVar(), "true")

	if !features.PostgresDatastore.Enabled() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	ctx := sac.WithAllAccess(context.Background())

	source := pgtest.GetConnectionString(s.T())
	config, err := pgxpool.ParseConfig(source)
	s.Require().NoError(err)
	pool, err := pgxpool.ConnectConfig(ctx, config)
	s.Require().NoError(err)

	Destroy(ctx, pool)

	s.pool = pool
	s.store = New(ctx, pool)
}

func (s *ProcessbaselinesStoreSuite) TearDownTest() {
	if s.pool != nil {
		s.pool.Close()
	}
	s.envIsolator.RestoreAll()
}

func (s *ProcessbaselinesStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	processBaseline := &storage.ProcessBaseline{}
	s.NoError(testutils.FullInit(processBaseline, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundProcessBaseline, exists, err := store.Get(ctx, processBaseline.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundProcessBaseline)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, processBaseline))
	foundProcessBaseline, exists, err = store.Get(ctx, processBaseline.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(processBaseline, foundProcessBaseline)

	processBaselineCount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, processBaselineCount)

	processBaselineExists, err := store.Exists(ctx, processBaseline.GetId())
	s.NoError(err)
	s.True(processBaselineExists)
	s.NoError(store.Upsert(ctx, processBaseline))
	s.ErrorIs(store.Upsert(withNoAccessCtx, processBaseline), sac.ErrResourceAccessDenied)

	foundProcessBaseline, exists, err = store.Get(ctx, processBaseline.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(processBaseline, foundProcessBaseline)

	s.NoError(store.Delete(ctx, processBaseline.GetId()))
	foundProcessBaseline, exists, err = store.Get(ctx, processBaseline.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundProcessBaseline)

	var processBaselines []*storage.ProcessBaseline
	for i := 0; i < 200; i++ {
		processBaseline := &storage.ProcessBaseline{}
		s.NoError(testutils.FullInit(processBaseline, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		processBaselines = append(processBaselines, processBaseline)
	}

	s.NoError(store.UpsertMany(ctx, processBaselines))

	processBaselineCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, processBaselineCount)
}

func (s *ProcessbaselinesStoreSuite) TestSACUpsert() {
	obj := &storage.ProcessBaseline{}
	s.NoError(testutils.FullInit(obj, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	ctxs := getSACContexts(obj, storage.Access_READ_WRITE_ACCESS)
	for name, expectedErr := range map[string]error{
		withAllAccess:           nil,
		withNoAccess:            sac.ErrResourceAccessDenied,
		withNoAccessToCluster:   sac.ErrResourceAccessDenied,
		withAccessToDifferentNs: sac.ErrResourceAccessDenied,
		withAccess:              nil,
		withAccessToCluster:     nil,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			assert.ErrorIs(t, s.store.Upsert(ctxs[name], obj), expectedErr)
		})
	}
}

func (s *ProcessbaselinesStoreSuite) TestSACUpsertMany() {
	obj := &storage.ProcessBaseline{}
	s.NoError(testutils.FullInit(obj, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	ctxs := getSACContexts(obj, storage.Access_READ_WRITE_ACCESS)
	for name, expectedErr := range map[string]error{
		withAllAccess:           nil,
		withNoAccess:            sac.ErrResourceAccessDenied,
		withNoAccessToCluster:   sac.ErrResourceAccessDenied,
		withAccessToDifferentNs: sac.ErrResourceAccessDenied,
		withAccess:              nil,
		withAccessToCluster:     nil,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			assert.ErrorIs(t, s.store.UpsertMany(ctxs[name], []*storage.ProcessBaseline{obj}), expectedErr)
		})
	}
}

func (s *ProcessbaselinesStoreSuite) TestSACCount() {
	objA := &storage.ProcessBaseline{}
	s.NoError(testutils.FullInit(objA, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	objB := &storage.ProcessBaseline{}
	s.NoError(testutils.FullInit(objB, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))

	withAllAccessCtx := sac.WithAllAccess(context.Background())
	s.store.Upsert(withAllAccessCtx, objA)
	s.store.Upsert(withAllAccessCtx, objB)

	ctxs := getSACContexts(objA, storage.Access_READ_ACCESS)
	for name, expectedCount := range map[string]int{
		withAllAccess:           2,
		withNoAccess:            0,
		withNoAccessToCluster:   0,
		withAccessToDifferentNs: 0,
		withAccess:              1,
		withAccessToCluster:     1,
	} {
		s.T().Run(fmt.Sprintf("with %s", name), func(t *testing.T) {
			count, err := s.store.Count(ctxs[name])
			assert.NoError(t, err)
			assert.Equal(t, expectedCount, count)
		})
	}
}

const (
	withAllAccess           = "AllAccess"
	withNoAccess            = "NoAccess"
	withAccessToDifferentNs = "AccessToDifferentNs"
	withAccess              = "Access"
	withAccessToCluster     = "AccessToCluster"
	withNoAccessToCluster   = "NoAccessToCluster"
)

func getSACContexts(obj *storage.ProcessBaseline, access storage.Access) map[string]context.Context {
	return map[string]context.Context{
		withAllAccess: sac.WithAllAccess(context.Background()),
		withNoAccess:  sac.WithNoAccess(context.Background()),
		withAccessToDifferentNs: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys(obj.GetKey().GetClusterId()),
				sac.NamespaceScopeKeys("unknown ns"),
			)),
		withAccess: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys(obj.GetKey().GetClusterId()),
				sac.NamespaceScopeKeys(obj.GetKey().GetNamespace()),
			)),
		withAccessToCluster: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys(obj.GetKey().GetClusterId()),
			)),
		withNoAccessToCluster: sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(access),
				sac.ResourceScopeKeys(targetResource),
				sac.ClusterScopeKeys("unknown cluster"),
			)),
	}
}
