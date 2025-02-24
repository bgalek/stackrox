package datastoretest

import (
	"context"
	"testing"

	"github.com/blevesearch/bleve"
	activeComponentDackbox "github.com/stackrox/rox/central/activecomponent/dackbox"
	activeComponentIndex "github.com/stackrox/rox/central/activecomponent/datastore/index"
	clusterCVEEdgeDackbox "github.com/stackrox/rox/central/clustercveedge/dackbox"
	clusterCVEEdgeIndex "github.com/stackrox/rox/central/clustercveedge/index"
	componentCVEEdgeDackbox "github.com/stackrox/rox/central/componentcveedge/dackbox"
	componentCVEEdgeIndex "github.com/stackrox/rox/central/componentcveedge/index"
	cveDackbox "github.com/stackrox/rox/central/cve/dackbox"
	cveIndex "github.com/stackrox/rox/central/cve/index"
	deploymentDackbox "github.com/stackrox/rox/central/deployment/dackbox"
	deploymentDataStore "github.com/stackrox/rox/central/deployment/datastore"
	deploymentIndex "github.com/stackrox/rox/central/deployment/index"
	"github.com/stackrox/rox/central/globalindex"
	imageDackbox "github.com/stackrox/rox/central/image/dackbox"
	"github.com/stackrox/rox/central/image/datastore"
	imageIndex "github.com/stackrox/rox/central/image/index"
	"github.com/stackrox/rox/central/image/mappings"
	imageComponentDackbox "github.com/stackrox/rox/central/imagecomponent/dackbox"
	imageComponentIndex "github.com/stackrox/rox/central/imagecomponent/index"
	imageComponentEdgeDackbox "github.com/stackrox/rox/central/imagecomponentedge/dackbox"
	imageComponentEdgeIndex "github.com/stackrox/rox/central/imagecomponentedge/index"
	imageCVEEdgeDackbox "github.com/stackrox/rox/central/imagecveedge/dackbox"
	imageCVEEdgeDataStore "github.com/stackrox/rox/central/imagecveedge/datastore"
	imageCVEEdgeIndex "github.com/stackrox/rox/central/imagecveedge/index"
	namespaceDataStore "github.com/stackrox/rox/central/namespace/datastore"
	nodeDackbox "github.com/stackrox/rox/central/node/dackbox"
	nodeIndex "github.com/stackrox/rox/central/node/index"
	nodeComponentEdgeDackbox "github.com/stackrox/rox/central/nodecomponentedge/dackbox"
	nodeComponentEdgeIndex "github.com/stackrox/rox/central/nodecomponentedge/index"
	"github.com/stackrox/rox/central/role/resources"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/concurrency"
	"github.com/stackrox/rox/pkg/dackbox"
	dackboxConcurrency "github.com/stackrox/rox/pkg/dackbox/concurrency"
	"github.com/stackrox/rox/pkg/dackbox/edges"
	"github.com/stackrox/rox/pkg/dackbox/indexer"
	"github.com/stackrox/rox/pkg/dackbox/utils/queue"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/fixtures"
	"github.com/stackrox/rox/pkg/images/types"
	"github.com/stackrox/rox/pkg/postgres/pgtest"
	"github.com/stackrox/rox/pkg/postgres/schema"
	"github.com/stackrox/rox/pkg/rocksdb"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/sac/testconsts"
	"github.com/stackrox/rox/pkg/sac/testutils"
	searchPkg "github.com/stackrox/rox/pkg/search"
	"github.com/stackrox/rox/pkg/search/postgres"
	"github.com/stackrox/rox/pkg/uuid"
	"github.com/stretchr/testify/suite"
)

func TestImageDataStoreSAC(t *testing.T) {
	suite.Run(t, new(imageDatastoreSACSuite))
}

type imageDatastoreSACSuite struct {
	suite.Suite

	// Elements for bleve+rocksdb mode
	engine   *rocksdb.RocksDB
	index    bleve.Index
	dacky    *dackbox.DackBox
	keyFence dackboxConcurrency.KeyFence
	indexQ   queue.WaitableQueue

	// Elements for postgres mode
	pgtestbase *pgtest.TestPostgres

	datastore datastore.DataStore

	imageVulnDatastore  imageCVEEdgeDataStore.DataStore
	deploymentDatastore deploymentDataStore.DataStore
	namespaceDatastore  namespaceDataStore.DataStore

	optionsMap searchPkg.OptionsMap

	testContexts map[string]context.Context
	testImageIDs []string

	extraImage *storage.Image
}

func (s *imageDatastoreSACSuite) SetupSuite() {
	var err error
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		s.pgtestbase = pgtest.ForT(s.T())
		s.Require().NotNil(s.pgtestbase)
		s.datastore, err = datastore.GetTestPostgresDataStore(s.T(), s.pgtestbase.Pool)
		s.Require().NoError(err)
		s.imageVulnDatastore = imageCVEEdgeDataStore.GetTestPostgresDataStore(s.T(), s.pgtestbase.Pool)
		s.deploymentDatastore, err = deploymentDataStore.GetTestPostgresDataStore(s.T(), s.pgtestbase.Pool)
		s.Require().NoError(err)
		s.namespaceDatastore, err = namespaceDataStore.GetTestPostgresDataStore(s.T(), s.pgtestbase.Pool)
		s.Require().NoError(err)
		s.optionsMap = schema.ImagesSchema.OptionsMap
	} else {
		s.engine, err = rocksdb.NewTemp("imageSACTest")
		s.Require().NoError(err)
		s.index, err = globalindex.MemOnlyIndex()
		s.Require().NoError(err)
		s.keyFence = dackboxConcurrency.NewKeyFence()
		s.indexQ = queue.NewWaitableQueue()
		s.dacky, err = dackbox.NewRocksDBDackBox(s.engine, s.indexQ, []byte("graph"), []byte("dirty"), []byte("valid"))
		s.Require().NoError(err)
		reg := indexer.NewWrapperRegistry()
		indexer.NewLazy(s.indexQ, reg, s.index, s.dacky.AckIndexed).Start()
		reg.RegisterWrapper(activeComponentDackbox.Bucket, activeComponentIndex.Wrapper{})
		reg.RegisterWrapper(clusterCVEEdgeDackbox.Bucket, clusterCVEEdgeIndex.Wrapper{})
		reg.RegisterWrapper(componentCVEEdgeDackbox.Bucket, componentCVEEdgeIndex.Wrapper{})
		reg.RegisterWrapper(cveDackbox.Bucket, cveIndex.Wrapper{})
		reg.RegisterWrapper(deploymentDackbox.Bucket, deploymentIndex.Wrapper{})
		reg.RegisterWrapper(imageDackbox.Bucket, imageIndex.Wrapper{})
		reg.RegisterWrapper(imageComponentDackbox.Bucket, imageComponentIndex.Wrapper{})
		reg.RegisterWrapper(imageComponentEdgeDackbox.Bucket, imageComponentEdgeIndex.Wrapper{})
		reg.RegisterWrapper(imageCVEEdgeDackbox.Bucket, imageCVEEdgeIndex.Wrapper{})
		reg.RegisterWrapper(nodeDackbox.Bucket, nodeIndex.Wrapper{})
		reg.RegisterWrapper(nodeComponentEdgeDackbox.Bucket, nodeComponentEdgeIndex.Wrapper{})

		s.datastore, err = datastore.GetTestRocksBleveDataStore(s.T(), s.engine, s.index, s.dacky, s.keyFence)
		s.Require().NoError(err)
		s.imageVulnDatastore = imageCVEEdgeDataStore.GetTestRocksBleveDataStore(s.T(), s.index, s.dacky, s.keyFence)
		s.deploymentDatastore, err = deploymentDataStore.GetTestRocksBleveDataStore(s.T(), s.engine, s.index, s.dacky, s.keyFence)
		s.Require().NoError(err)
		s.namespaceDatastore, err = namespaceDataStore.GetTestRocksBleveDataStore(s.T(), s.engine, s.index, s.dacky, s.keyFence)
		s.Require().NoError(err)
		s.optionsMap = mappings.OptionsMap
	}

	s.testContexts = testutils.GetNamespaceScopedTestContexts(context.Background(), s.T(),
		resources.Image)

	s.extraImage = fixtures.GetImage()
}

func (s *imageDatastoreSACSuite) TearDownSuite() {
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		s.pgtestbase.Pool.Close()
	} else {
		s.Require().NoError(rocksdb.CloseAndRemove(s.engine))
		s.Require().NoError(s.index.Close())
	}
}

func (s *imageDatastoreSACSuite) SetupTest() {
	s.testImageIDs = make([]string, 0)
}

func (s *imageDatastoreSACSuite) TearDownTest() {
	for _, id := range s.testImageIDs {
		s.deleteImage(id)
	}
}

func (s *imageDatastoreSACSuite) waitForIndexing() {
	if !env.PostgresDatastoreEnabled.BooleanSetting() {
		// Some cases need to wait for dackbox indexing to complete.
		doneSignal := concurrency.NewSignal()
		s.indexQ.PushSignal(&doneSignal)
		<-doneSignal.Done()
	}
}

func (s *imageDatastoreSACSuite) deleteImage(id string) {
	s.Require().NoError(s.datastore.DeleteImages(s.testContexts[testutils.UnrestrictedReadWriteCtx], id))
}

func (s *imageDatastoreSACSuite) deleteDeployment(clusterid, id string) {
	s.Require().NoError(s.deploymentDatastore.RemoveDeployment(sac.WithAllAccess(context.Background()), clusterid, id))
}

func (s *imageDatastoreSACSuite) deleteNamespace(id string) {
	s.Require().NoError(s.namespaceDatastore.RemoveNamespace(sac.WithAllAccess(context.Background()), id))
}

func getImageCVEID(cve string) string {
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		return cve + "#crime-stories"
	}
	return cve
}

// getImageCVEEdgeID returns base 64 encoded Image:CVE ids
func getImageCVEEdgeID(image, cve string) string {
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		return postgres.IDFromPks([]string{image, getImageCVEID(cve)})
	}
	return edges.EdgeID{ParentID: image, ChildID: getImageCVEID(cve)}.ToString()
}

func (s *imageDatastoreSACSuite) verifyListImagesEqual(image1, image2 *storage.ListImage) {
	s.Equal(image1.GetId(), image2.GetId())
	s.Equal(image1.GetComponents(), image2.GetComponents())
	s.Equal(image1.GetCves(), image2.GetCves())
}

func (s *imageDatastoreSACSuite) verifyRawImagesEqual(image1, image2 *storage.Image) {
	s.Equal(image1.GetId(), image2.GetId())
	s.Equal(image1.GetComponents(), image2.GetComponents())
	s.Equal(image1.GetCves(), image2.GetCves())
}

func (s *imageDatastoreSACSuite) TestUpsertImage() {
	cases := testutils.GenericGlobalSACUpsertTestCases(s.T(), testutils.VerbUpsert)

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			image := fixtures.GetImageSherlockHolmes1()
			err := s.datastore.UpsertImage(ctx, image)
			defer s.deleteImage(image.GetId())
			if testCase.ExpectError {
				s.Error(err)
				s.ErrorIs(err, testCase.ExpectedError)
			} else {
				s.NoError(err)
				checkCtx := s.testContexts[testutils.UnrestrictedReadCtx]
				readImage, found, checkErr := s.datastore.GetImage(checkCtx, image.GetId())
				s.NoError(checkErr)
				s.True(found)
				s.Equal(*image.GetName(), *readImage.GetName())
			}
		})
	}
}

func (s *imageDatastoreSACSuite) TestUpdateVulnerabilityState() {
	cases := testutils.GenericGlobalSACUpsertTestCases(s.T(), "update vulnerability state")

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			writeCtx := s.testContexts[testutils.UnrestrictedReadWriteCtx]
			checkCtx := s.testContexts[testutils.UnrestrictedReadCtx]
			image := fixtures.GetImageSherlockHolmes1()
			cve1 := fixtures.GetEmbeddedImageCVE1234x0001()
			cve2 := fixtures.GetEmbeddedImageCVE4567x0002()
			cve3 := fixtures.GetEmbeddedImageCVE1234x0003()
			cve4 := fixtures.GetEmbeddedImageCVE3456x0004()
			cve5 := fixtures.GetEmbeddedImageCVE3456x0005()
			cve6 := fixtures.GetEmbeddedImageCVE2345x0006()
			foundCVEs := []*storage.EmbeddedVulnerability{cve1, cve2, cve3, cve4, cve5}
			missingCVEs := []*storage.EmbeddedVulnerability{cve6}
			insertErr := s.datastore.UpsertImage(writeCtx, image)
			defer s.deleteImage(image.GetId())
			s.Require().NoError(insertErr)
			for _, cve := range foundCVEs {
				edgeID := getImageCVEEdgeID(image.GetId(), cve.GetCve())
				edge, found, err := s.imageVulnDatastore.Get(checkCtx, edgeID)
				s.NoError(err)
				s.True(found)
				s.Equal(storage.VulnerabilityState_OBSERVED, edge.GetState())
			}
			for _, cve := range missingCVEs {
				edgeID := getImageCVEEdgeID(image.GetId(), cve.GetCve())
				edge, found, err := s.imageVulnDatastore.Get(checkCtx, edgeID)
				s.NoError(err)
				s.False(found)
				s.Nil(edge)
			}
			targetCve := cve3.GetCve()
			newState := storage.VulnerabilityState_DEFERRED
			updateErr := s.datastore.UpdateVulnerabilityState(ctx, targetCve, []string{image.GetId()}, newState)
			if testCase.ExpectError {
				s.Error(updateErr)
				s.ErrorIs(updateErr, testCase.ExpectedError)
			} else {
				s.NoError(updateErr)
				for _, cve := range foundCVEs {
					expectedState := storage.VulnerabilityState_OBSERVED
					if !env.PostgresDatastoreEnabled.BooleanSetting() && cve.GetCve() == cve3.GetCve() {
						// Currently, UpdateVulnerabilityState only updates the state column, but not the stored serialized object
						expectedState = storage.VulnerabilityState_DEFERRED
					}
					edgeID := getImageCVEEdgeID(image.GetId(), cve.GetCve())
					edge, found, err := s.imageVulnDatastore.Get(checkCtx, edgeID)
					s.NoError(err)
					s.True(found)
					s.Equal(expectedState, edge.GetState())
				}
				for _, cve := range missingCVEs {
					edgeID := getImageCVEEdgeID(image.GetId(), cve.GetCve())
					edge, found, err := s.imageVulnDatastore.Get(checkCtx, edgeID)
					s.NoError(err)
					s.False(found)
					s.Nil(edge)
				}
			}
		})
	}
}

func (s *imageDatastoreSACSuite) TestDeleteImages() {
	cases := testutils.GenericGlobalSACDeleteTestCases(s.T())

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			writeCtx := s.testContexts[testutils.UnrestrictedReadWriteCtx]
			checkCtx := s.testContexts[testutils.UnrestrictedReadCtx]
			image := fixtures.GetImageSherlockHolmes1()
			defer s.deleteImage(image.GetId())
			upsertErr := s.datastore.UpsertImage(writeCtx, image)
			s.Require().NoError(upsertErr)
			_, found, check1Err := s.datastore.GetImage(checkCtx, image.GetId())
			s.Require().NoError(check1Err)
			s.Require().True(found)
			deleteErr := s.datastore.DeleteImages(ctx, image.GetId())
			if testCase.ExpectError {
				s.Error(deleteErr)
				s.ErrorIs(deleteErr, testCase.ExpectedError)
				_, postRemovalFound, check2Err := s.datastore.GetImage(checkCtx, image.GetId())
				s.NoError(check2Err)
				s.True(postRemovalFound)
			} else {
				s.NoError(deleteErr)
				_, postRemovalFound, check2Err := s.datastore.GetImage(checkCtx, image.GetId())
				s.NoError(check2Err)
				s.False(postRemovalFound)
			}
		})
	}
}

func (s *imageDatastoreSACSuite) setupReadTest() ([]*storage.Image, func(), error) {
	var setupErr error
	namespacesToDelete := make([]string, 0, 1)
	deploymentsToDelete := make([]*storage.Deployment, 0, 1)
	imagesToDelete := make([]string, 0, 1)
	images := make([]*storage.Image, 0, 1)
	cleanup := func() {
		for _, img := range imagesToDelete {
			s.deleteImage(img)
		}
		for _, deployment := range deploymentsToDelete {
			s.deleteDeployment(deployment.GetClusterId(), deployment.GetId())
		}
		for _, ns := range namespacesToDelete {
			s.deleteNamespace(ns)
		}
	}
	setupCtx := sac.WithAllAccess(context.Background())
	namespace := fixtures.GetNamespace(testconsts.Cluster2, testconsts.Cluster2, testconsts.NamespaceB)
	namespacesToDelete = append(namespacesToDelete, namespace.GetId())
	setupErr = s.namespaceDatastore.AddNamespace(setupCtx, namespace)
	if setupErr != nil {
		return nil, cleanup, setupErr
	}
	deployment := fixtures.GetDeploymentSherlockHolmes1(uuid.NewV4().String(), namespace)
	deploymentsToDelete = append(deploymentsToDelete, deployment)
	setupErr = s.deploymentDatastore.UpsertDeployment(setupCtx, deployment)
	if setupErr != nil {
		return nil, cleanup, setupErr
	}
	deployment2 := fixtures.GetDeploymentDoctorJekyll2(uuid.NewV4().String(), namespace)
	deploymentsToDelete = append(deploymentsToDelete, deployment2)
	setupErr = s.deploymentDatastore.UpsertDeployment(setupCtx, deployment2)
	if setupErr != nil {
		return nil, cleanup, setupErr
	}
	image := fixtures.GetImageSherlockHolmes1()
	imagesToDelete = append(imagesToDelete, image.GetId())
	images = append(images, image)
	setupErr = s.datastore.UpsertImage(setupCtx, image)
	if setupErr != nil {
		return nil, cleanup, setupErr
	}
	image2 := fixtures.GetImageDoctorJekyll2()
	imagesToDelete = append(imagesToDelete, image2.GetId())
	images = append(images, image2)
	setupErr = s.datastore.UpsertImage(setupCtx, image2)
	if setupErr != nil {
		return nil, cleanup, setupErr
	}
	s.waitForIndexing()
	return images, cleanup, nil
}

func (s *imageDatastoreSACSuite) TestExists() {
	images, cleanup, setupErr := s.setupReadTest()
	defer cleanup()
	s.Require().NoError(setupErr)
	s.Require().NotZero(len(images))
	image := images[0]

	cases := testutils.GenericNamespaceSACGetTestCases(s.T())

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			exists, err := s.datastore.Exists(ctx, image.GetId())
			s.NoError(err)
			if testCase.ExpectedFound {
				s.True(exists)
			} else {
				s.False(exists)
			}
		})
	}
}

func (s *imageDatastoreSACSuite) TestListImage() {
	images, cleanup, setupErr := s.setupReadTest()
	defer cleanup()
	s.Require().NoError(setupErr)
	s.Require().NotZero(len(images))
	image := images[0]

	cases := testutils.GenericNamespaceSACGetTestCases(s.T())

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			readImage, found, err := s.datastore.ListImage(ctx, image.GetId())
			s.Require().NoError(err)
			if testCase.ExpectedFound {
				s.True(found)
				s.verifyListImagesEqual(types.ConvertImageToListImage(image), readImage)
			} else {
				s.False(found)
				s.Nil(readImage)
			}
		})
	}
}

func (s *imageDatastoreSACSuite) TestGetImage() {
	images, cleanup, setupErr := s.setupReadTest()
	defer cleanup()
	s.Require().NoError(setupErr)
	s.Require().NotZero(len(images))
	image := images[0]

	cases := testutils.GenericNamespaceSACGetTestCases(s.T())

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			readImage, found, err := s.datastore.GetImage(ctx, image.GetId())
			s.Require().NoError(err)
			if testCase.ExpectedFound {
				s.True(found)
				s.verifyRawImagesEqual(image, readImage)
			} else {
				s.False(found)
				s.Nil(readImage)
			}
		})
	}
}

func (s *imageDatastoreSACSuite) TestGetImageMetadata() {
	images, cleanup, setupErr := s.setupReadTest()
	defer cleanup()
	s.Require().NoError(setupErr)
	s.Require().NotZero(len(images))
	image := images[0]

	cases := testutils.GenericNamespaceSACGetTestCases(s.T())

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			readImageMeta, found, err := s.datastore.GetImageMetadata(ctx, image.GetId())
			s.Require().NoError(err)
			if testCase.ExpectedFound {
				s.True(found)
				s.Equal(image.GetId(), readImageMeta.GetId())
				s.Equal(image.GetComponents(), readImageMeta.GetComponents())
				s.Equal(image.GetCves(), readImageMeta.GetCves())
			} else {
				s.False(found)
				s.Nil(readImageMeta)
			}
		})
	}

	if env.PostgresDatastoreEnabled.BooleanSetting() {
		s.Require().True(len(images) > 1)
		image2 := images[1]
		// Test GetManyImageMetadata in postgres mode (only supported mode).
		for name, testCase := range cases {
			s.Run("Many_"+name, func() {
				ctx := s.testContexts[testCase.ScopeKey]
				readMeta, err := s.datastore.GetManyImageMetadata(ctx, []string{image.GetId(), image2.GetId()})
				s.Require().NoError(err)
				if testCase.ExpectedFound {
					s.Require().Equal(2, len(readMeta))
					readImageMeta1 := readMeta[0]
					readImageMeta2 := readMeta[1]
					if readImageMeta1.GetId() == image.GetId() {
						s.Equal(image.GetId(), readImageMeta1.GetId())
						s.Equal(image.GetComponents(), readImageMeta1.GetComponents())
						s.Equal(image.GetCves(), readImageMeta1.GetCves())
						s.Equal(image2.GetId(), readImageMeta2.GetId())
						s.Equal(image2.GetComponents(), readImageMeta2.GetComponents())
						s.Equal(image2.GetCves(), readImageMeta2.GetCves())
					} else {
						s.Equal(image2.GetId(), readImageMeta1.GetId())
						s.Equal(image2.GetComponents(), readImageMeta1.GetComponents())
						s.Equal(image2.GetCves(), readImageMeta1.GetCves())
						s.Equal(image.GetId(), readImageMeta2.GetId())
						s.Equal(image.GetComponents(), readImageMeta2.GetComponents())
						s.Equal(image.GetCves(), readImageMeta2.GetCves())
					}
				} else {
					s.Equal(0, len(readMeta))
				}
			})
		}
	}
}

func (s *imageDatastoreSACSuite) TestGetImagesBatch() {
	images, cleanup, setupErr := s.setupReadTest()
	defer cleanup()
	s.Require().NoError(setupErr)
	s.Require().True(len(images) > 1)
	image1 := images[0]
	image2 := images[1]

	cases := testutils.GenericNamespaceSACGetTestCases(s.T())

	for name, testCase := range cases {
		s.Run(name, func() {
			ctx := s.testContexts[testCase.ScopeKey]
			readMeta, err := s.datastore.GetImagesBatch(ctx, []string{image1.GetId(), image2.GetId()})
			s.Require().NoError(err)
			if testCase.ExpectedFound {
				s.Require().Equal(2, len(readMeta))
				readImageMeta1 := readMeta[0]
				readImageMeta2 := readMeta[1]
				if readImageMeta1.GetId() == image1.GetId() {
					s.Equal(image1.GetId(), readImageMeta1.GetId())
					s.Equal(image1.GetComponents(), readImageMeta1.GetComponents())
					s.Equal(image1.GetCves(), readImageMeta1.GetCves())
					s.Equal(image2.GetId(), readImageMeta2.GetId())
					s.Equal(image2.GetComponents(), readImageMeta2.GetComponents())
					s.Equal(image2.GetCves(), readImageMeta2.GetCves())
				} else {
					s.Equal(image2.GetId(), readImageMeta1.GetId())
					s.Equal(image2.GetComponents(), readImageMeta1.GetComponents())
					s.Equal(image2.GetCves(), readImageMeta1.GetCves())
					s.Equal(image1.GetId(), readImageMeta2.GetId())
					s.Equal(image1.GetComponents(), readImageMeta2.GetComponents())
					s.Equal(image1.GetCves(), readImageMeta2.GetCves())
				}
			} else {
				s.Equal(0, len(readMeta))
			}
		})
	}
}

func (s *imageDatastoreSACSuite) getSearchTestCases() map[string]map[string]bool {
	// The map structure is the mapping ScopeKey -> ImageID -> Visible
	cases := map[string]map[string]bool{
		testutils.UnrestrictedReadCtx: {
			s.extraImage.GetId():                       true,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   true,
		},
		testutils.UnrestrictedReadWriteCtx: {
			s.extraImage.GetId():                       true,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   true,
		},
		testutils.Cluster1ReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster1NamespaceAReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster1NamespaceBReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): false,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster1NamespacesABReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster1NamespacesBCReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): false,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster2ReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   true,
		},
		testutils.Cluster2NamespaceAReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): false,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster2NamespaceBReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   true,
		},
		testutils.Cluster2NamespacesACReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): false,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster2NamespacesBCReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   true,
		},
		testutils.Cluster3ReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): false,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster3NamespaceAReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): false,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.Cluster3NamespaceBReadWriteCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): false,
			fixtures.GetImageDoctorJekyll2().GetId():   false,
		},
		testutils.MixedClusterAndNamespaceReadCtx: {
			s.extraImage.GetId():                       false,
			fixtures.GetImageSherlockHolmes1().GetId(): true,
			fixtures.GetImageDoctorJekyll2().GetId():   true,
		},
	}
	return cases
}

func (s *imageDatastoreSACSuite) setupSearchTest() (func(), error) {
	var setupErr error

	namespacesToDelete := make([]string, 0, 1)
	deploymentsToDelete := make([]*storage.Deployment, 0, 1)
	imagesToDelete := make([]string, 0, 1)

	cleanup := func() {
		for _, img := range imagesToDelete {
			s.deleteImage(img)
		}
		for _, deployment := range deploymentsToDelete {
			s.deleteDeployment(deployment.GetClusterId(), deployment.GetId())
		}
		for _, ns := range namespacesToDelete {
			s.deleteNamespace(ns)
		}
	}

	image1 := fixtures.GetImageSherlockHolmes1()
	imagesToDelete = append(imagesToDelete, image1.GetId())
	image2 := fixtures.GetImageDoctorJekyll2()
	imagesToDelete = append(imagesToDelete, image2.GetId())
	imagesToDelete = append(imagesToDelete, s.extraImage.GetId())

	namespace1A := fixtures.GetNamespace(testconsts.Cluster1, testconsts.Cluster1, testconsts.NamespaceA)
	namespacesToDelete = append(namespacesToDelete, namespace1A.GetId())
	namespace2B := fixtures.GetNamespace(testconsts.Cluster2, testconsts.Cluster2, testconsts.NamespaceB)
	namespacesToDelete = append(namespacesToDelete, namespace2B.GetId())

	deployment1A1 := fixtures.GetDeploymentSherlockHolmes1(uuid.NewV4().String(), namespace1A)
	deploymentsToDelete = append(deploymentsToDelete, deployment1A1)
	deployment2B1 := fixtures.GetDeploymentSherlockHolmes1(uuid.NewV4().String(), namespace2B)
	deploymentsToDelete = append(deploymentsToDelete, deployment2B1)
	deployment2B2 := fixtures.GetDeploymentDoctorJekyll2(uuid.NewV4().String(), namespace2B)
	deploymentsToDelete = append(deploymentsToDelete, deployment2B2)

	setupCtx := sac.WithAllAccess(context.Background())

	setupErr = s.namespaceDatastore.AddNamespace(setupCtx, namespace1A)
	if setupErr != nil {
		return cleanup, setupErr
	}
	setupErr = s.namespaceDatastore.AddNamespace(setupCtx, namespace2B)
	if setupErr != nil {
		return cleanup, setupErr
	}

	setupErr = s.datastore.UpsertImage(setupCtx, s.extraImage)
	if setupErr != nil {
		return cleanup, setupErr
	}
	setupErr = s.datastore.UpsertImage(setupCtx, image1)
	if setupErr != nil {
		return cleanup, setupErr
	}
	setupErr = s.datastore.UpsertImage(setupCtx, image2)
	if setupErr != nil {
		return cleanup, setupErr
	}

	setupErr = s.deploymentDatastore.UpsertDeployment(setupCtx, deployment1A1)
	if setupErr != nil {
		return cleanup, setupErr
	}
	setupErr = s.deploymentDatastore.UpsertDeployment(setupCtx, deployment2B1)
	if setupErr != nil {
		return cleanup, setupErr
	}
	setupErr = s.deploymentDatastore.UpsertDeployment(setupCtx, deployment2B2)
	if setupErr != nil {
		return cleanup, setupErr
	}

	s.waitForIndexing()
	return cleanup, nil

}

func (s *imageDatastoreSACSuite) TestCountImages() {
	cleanup, setupErr := s.setupSearchTest()
	defer cleanup()
	s.Require().NoError(setupErr)

	cases := s.getSearchTestCases()
	for key, testCase := range cases {
		s.Run(key, func() {
			ctx := s.testContexts[key]
			expectedCount := 0
			for _, visible := range testCase {
				if visible {
					expectedCount++
				}
			}
			count, err := s.datastore.CountImages(ctx)
			s.NoError(err)
			s.Equal(expectedCount, count)
		})
	}
}

func (s *imageDatastoreSACSuite) TestCount() {
	cleanup, setupErr := s.setupSearchTest()
	defer cleanup()
	s.Require().NoError(setupErr)

	cases := s.getSearchTestCases()
	for key, testCase := range cases {
		s.Run(key, func() {
			ctx := s.testContexts[key]
			expectedCount := 0
			for _, visible := range testCase {
				if visible {
					expectedCount++
				}
			}
			count, err := s.datastore.Count(ctx, searchPkg.EmptyQuery())
			s.NoError(err)
			s.Equal(expectedCount, count)
		})
	}
}

func (s *imageDatastoreSACSuite) TestSearch() {
	cleanup, setupErr := s.setupSearchTest()
	defer cleanup()
	s.Require().NoError(setupErr)

	cases := s.getSearchTestCases()
	for key, testCase := range cases {
		s.Run(key, func() {
			ctx := s.testContexts[key]
			expectedIDs := make([]string, 0, len(testCase))
			for imageID, visible := range testCase {
				if visible {
					expectedIDs = append(expectedIDs, imageID)
				}
			}
			results, err := s.datastore.Search(ctx, searchPkg.EmptyQuery())
			s.NoError(err)
			resultIDHeap := make(map[string]struct{}, 0)
			for _, r := range results {
				resultIDHeap[r.ID] = struct{}{}
			}
			resultIDs := make([]string, 0, len(resultIDHeap))
			for k := range resultIDHeap {
				resultIDs = append(resultIDs, k)
			}
			s.ElementsMatch(expectedIDs, resultIDs)
		})
	}
}

func (s *imageDatastoreSACSuite) TestSearchImages() {
	cleanup, setupErr := s.setupSearchTest()
	defer cleanup()
	s.Require().NoError(setupErr)

	cases := s.getSearchTestCases()
	for key, testCase := range cases {
		s.Run(key, func() {
			ctx := s.testContexts[key]
			expectedIDs := make([]string, 0, len(testCase))
			for imageID, visible := range testCase {
				if visible {
					expectedIDs = append(expectedIDs, imageID)
				}
			}
			results, err := s.datastore.SearchImages(ctx, searchPkg.EmptyQuery())
			s.NoError(err)
			resultIDHeap := make(map[string]struct{}, 0)
			for _, r := range results {
				resultIDHeap[r.GetId()] = struct{}{}
			}
			resultIDs := make([]string, 0, len(resultIDHeap))
			for k := range resultIDHeap {
				resultIDs = append(resultIDs, k)
			}
			s.ElementsMatch(expectedIDs, resultIDs)
		})
	}
}

func (s *imageDatastoreSACSuite) TestSearchRawImages() {
	cleanup, setupErr := s.setupSearchTest()
	defer cleanup()
	s.Require().NoError(setupErr)
	refImages := map[string]*storage.Image{
		s.extraImage.GetId():                       s.extraImage,
		fixtures.GetImageSherlockHolmes1().GetId(): fixtures.GetImageSherlockHolmes1(),
		fixtures.GetImageDoctorJekyll2().GetId():   fixtures.GetImageDoctorJekyll2(),
	}

	cases := s.getSearchTestCases()
	for key, testCase := range cases {
		s.Run(key, func() {
			ctx := s.testContexts[key]
			expectedIDs := make([]string, 0, len(testCase))
			for imageID, visible := range testCase {
				if visible {
					expectedIDs = append(expectedIDs, imageID)
				}
			}
			results, err := s.datastore.SearchRawImages(ctx, searchPkg.EmptyQuery())
			s.NoError(err)
			resultImages := make(map[string]*storage.Image, 0)
			for _, r := range results {
				resultImages[r.GetId()] = r
			}
			resultIDs := make([]string, 0, len(resultImages))
			for k := range resultImages {
				resultIDs = append(resultIDs, k)
			}
			s.ElementsMatch(expectedIDs, resultIDs)
			for _, imageID := range expectedIDs {
				s.verifyRawImagesEqual(refImages[imageID], resultImages[imageID])
			}
		})
	}
}

func (s *imageDatastoreSACSuite) TestSearchListImages() {
	cleanup, setupErr := s.setupSearchTest()
	defer cleanup()
	s.Require().NoError(setupErr)
	refImages := map[string]*storage.Image{
		s.extraImage.GetId():                       s.extraImage,
		fixtures.GetImageSherlockHolmes1().GetId(): fixtures.GetImageSherlockHolmes1(),
		fixtures.GetImageDoctorJekyll2().GetId():   fixtures.GetImageDoctorJekyll2(),
	}

	cases := s.getSearchTestCases()
	for key, testCase := range cases {
		s.Run(key, func() {
			ctx := s.testContexts[key]
			expectedIDs := make([]string, 0, len(testCase))
			for imageID, visible := range testCase {
				if visible {
					expectedIDs = append(expectedIDs, imageID)
				}
			}
			results, err := s.datastore.SearchListImages(ctx, searchPkg.EmptyQuery())
			s.NoError(err)
			resultListImages := make(map[string]*storage.ListImage, 0)
			for _, r := range results {
				resultListImages[r.GetId()] = r
			}
			resultIDs := make([]string, 0, len(resultListImages))
			for k := range resultListImages {
				resultIDs = append(resultIDs, k)
			}
			s.ElementsMatch(expectedIDs, resultIDs)
			for _, imageID := range expectedIDs {
				s.verifyListImagesEqual(types.ConvertImageToListImage(refImages[imageID]), resultListImages[imageID])
			}
		})
	}
}
