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

type ImageCvesStoreSuite struct {
	suite.Suite
	store  Store
	testDB *pgtest.TestPostgres
}

func TestImageCvesStore(t *testing.T) {
	suite.Run(t, new(ImageCvesStoreSuite))
}

func (s *ImageCvesStoreSuite) SetupSuite() {
	s.T().Setenv(env.PostgresDatastoreEnabled.EnvVar(), "true")

	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		s.T().Skip("Skip postgres store tests")
		s.T().SkipNow()
	}

	s.testDB = pgtest.ForT(s.T())
	s.store = New(s.testDB.Pool)
}

func (s *ImageCvesStoreSuite) SetupTest() {
	ctx := sac.WithAllAccess(context.Background())
	tag, err := s.testDB.Exec(ctx, "TRUNCATE image_cves CASCADE")
	s.T().Log("image_cves", tag)
	s.NoError(err)
}

func (s *ImageCvesStoreSuite) TearDownSuite() {
	s.testDB.Teardown(s.T())
}

func (s *ImageCvesStoreSuite) TestStore() {
	ctx := sac.WithAllAccess(context.Background())

	store := s.store

	imageCVE := &storage.ImageCVE{}
	s.NoError(testutils.FullInit(imageCVE, testutils.SimpleInitializer(), testutils.JSONFieldsFilter))

	foundImageCVE, exists, err := store.Get(ctx, imageCVE.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundImageCVE)

	withNoAccessCtx := sac.WithNoAccess(ctx)

	s.NoError(store.Upsert(ctx, imageCVE))
	foundImageCVE, exists, err = store.Get(ctx, imageCVE.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(imageCVE, foundImageCVE)

	imageCVECount, err := store.Count(ctx)
	s.NoError(err)
	s.Equal(1, imageCVECount)
	imageCVECount, err = store.Count(withNoAccessCtx)
	s.NoError(err)
	s.Zero(imageCVECount)

	imageCVEExists, err := store.Exists(ctx, imageCVE.GetId())
	s.NoError(err)
	s.True(imageCVEExists)
	s.NoError(store.Upsert(ctx, imageCVE))
	s.ErrorIs(store.Upsert(withNoAccessCtx, imageCVE), sac.ErrResourceAccessDenied)

	foundImageCVE, exists, err = store.Get(ctx, imageCVE.GetId())
	s.NoError(err)
	s.True(exists)
	s.Equal(imageCVE, foundImageCVE)

	s.NoError(store.Delete(ctx, imageCVE.GetId()))
	foundImageCVE, exists, err = store.Get(ctx, imageCVE.GetId())
	s.NoError(err)
	s.False(exists)
	s.Nil(foundImageCVE)
	s.NoError(store.Delete(withNoAccessCtx, imageCVE.GetId()))

	var imageCVEs []*storage.ImageCVE
	var imageCVEIDs []string
	for i := 0; i < 200; i++ {
		imageCVE := &storage.ImageCVE{}
		s.NoError(testutils.FullInit(imageCVE, testutils.UniqueInitializer(), testutils.JSONFieldsFilter))
		imageCVEs = append(imageCVEs, imageCVE)
		imageCVEIDs = append(imageCVEIDs, imageCVE.GetId())
	}

	s.NoError(store.UpsertMany(ctx, imageCVEs))

	imageCVECount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(200, imageCVECount)

	s.NoError(store.DeleteMany(ctx, imageCVEIDs))

	imageCVECount, err = store.Count(ctx)
	s.NoError(err)
	s.Equal(0, imageCVECount)
}
