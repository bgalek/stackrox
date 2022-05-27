// Code generated by pg-bindings generator. DO NOT EDIT.

//go:build sql_integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/features"
	"github.com/stackrox/rox/pkg/postgres/pgtest"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/testutils"
	"github.com/stackrox/rox/pkg/testutils/envisolator"
	"github.com/stretchr/testify/suite"
)

type ComplianceOperatorScanSettingBindingsStoreSuite struct {
	suite.Suite
	envIsolator *envisolator.EnvIsolator
	store       Store
	pool        *pgxpool.Pool
}

func TestComplianceOperatorScanSettingBindingsStore(t *testing.T) {
	suite.Run(t, new(ComplianceOperatorScanSettingBindingsStoreSuite))
}

func (s *ComplianceOperatorScanSettingBindingsStoreSuite) SetupTest() {
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

func (s *ComplianceOperatorScanSettingBindingsStoreSuite) TearDownTest() {
	if s.pool != nil {
		s.pool.Close()
	}
	s.envIsolator.RestoreAll()
}

func (s *ComplianceOperatorScanSettingBindingsStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	complianceOperatorScanSettingBinding := &storage.ComplianceOperatorScanSettingBinding{}
	s.NoError(testutils.FullInit(complianceOperatorScanSettingBinding, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundComplianceOperatorScanSettingBinding, exists, err := store.Get(ctx, complianceOperatorScanSettingBinding.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundComplianceOperatorScanSettingBinding)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, complianceOperatorScanSettingBinding))
	foundComplianceOperatorScanSettingBinding, exists, err = store.Get(ctx, complianceOperatorScanSettingBinding.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(complianceOperatorScanSettingBinding, foundComplianceOperatorScanSettingBinding)

	complianceOperatorScanSettingBindingCount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, complianceOperatorScanSettingBindingCount)
	complianceOperatorScanSettingBindingCount, err = store.Count(withNoAccessCtx)
	s.NoError(err)
	s.Zero(complianceOperatorScanSettingBindingCount)

	complianceOperatorScanSettingBindingExists, err := store.Exists(ctx, complianceOperatorScanSettingBinding.GetId())
	s.NoError(err)
	s.True(complianceOperatorScanSettingBindingExists)
	s.NoError(store.Upsert(ctx, complianceOperatorScanSettingBinding))
	s.ErrorIs(store.Upsert(withNoAccessCtx, complianceOperatorScanSettingBinding), sac.ErrResourceAccessDenied)

	foundComplianceOperatorScanSettingBinding, exists, err = store.Get(ctx, complianceOperatorScanSettingBinding.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(complianceOperatorScanSettingBinding, foundComplianceOperatorScanSettingBinding)

	s.NoError(store.Delete(ctx, complianceOperatorScanSettingBinding.GetId()))
	foundComplianceOperatorScanSettingBinding, exists, err = store.Get(ctx, complianceOperatorScanSettingBinding.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundComplianceOperatorScanSettingBinding)
	s.ErrorIs(store.Delete(withNoAccessCtx, complianceOperatorScanSettingBinding.GetId()), sac.ErrResourceAccessDenied)

	var complianceOperatorScanSettingBindings []*storage.ComplianceOperatorScanSettingBinding
	for i := 0; i < 200; i++ {
		complianceOperatorScanSettingBinding := &storage.ComplianceOperatorScanSettingBinding{}
		s.NoError(testutils.FullInit(complianceOperatorScanSettingBinding, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		complianceOperatorScanSettingBindings = append(complianceOperatorScanSettingBindings, complianceOperatorScanSettingBinding)
	}

	s.NoError(store.UpsertMany(ctx, complianceOperatorScanSettingBindings))

	complianceOperatorScanSettingBindingCount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, complianceOperatorScanSettingBindingCount)
}
