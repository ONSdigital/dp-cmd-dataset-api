package steps_test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-dataset-api/config"
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/mongo"
	"github.com/ONSdigital/dp-dataset-api/service"
	serviceMock "github.com/ONSdigital/dp-dataset-api/service/mock"
	"github.com/ONSdigital/dp-dataset-api/store"
	storeMock "github.com/ONSdigital/dp-dataset-api/store/datastoretest"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	"github.com/benweissmann/memongo"
	"github.com/cucumber/godog"
	"github.com/globalsign/mgo"
	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type DatasetFeature struct {
	ErrorFeature
	svc             *service.Service
	errorChan       chan error
	Datasets        []*models.Dataset
	MongoClient     *mongo.Mongo
	Config          *config.Configuration
	HTTPServer      *http.Server
	FakeAuthService *httpfake.HTTPFake
	ServiceRunning  bool
}

func NewDatasetFeature(mongoCapability *MongoCapability) *DatasetFeature {

	f := &DatasetFeature{
		HTTPServer:      &http.Server{},
		errorChan:       make(chan error),
		Datasets:        make([]*models.Dataset, 0),
		FakeAuthService: httpfake.New(),
		ServiceRunning:  false,
	}

	var err error

	f.Config, err = config.Get()
	if err != nil {
		panic(err)
	}

	f.Config.ZebedeeURL = f.FakeAuthService.ResolveURL("")

	mongodb := &mongo.Mongo{
		CodeListURL: "",
		Collection:  "datasets",
		Database:    memongo.RandomDatabase(),
		DatasetURL:  "datasets",
		URI:         mongoCapability.Server.URI(),
	}

	if err := mongodb.Init(); err != nil {
		panic(err)
	}

	f.MongoClient = mongodb

	initMock := &serviceMock.InitialiserMock{
		DoGetMongoDBFunc:       f.DoGetMongoDB,
		DoGetGraphDBFunc:       f.DoGetGraphDBOk,
		DoGetKafkaProducerFunc: f.DoGetKafkaProducerOk,
		DoGetHealthCheckFunc:   f.DoGetHealthcheckOk,
		DoGetHTTPServerFunc:    f.DoGetHTTPServer,
	}

	f.svc = service.New(f.Config, service.NewServiceList(initMock))

	return f
}

func (f *DatasetFeature) RegisterSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^private endpoints are enabled$`, f.PrivateEndpointsAreEnabled)
	ctx.Step(`^I am not identified$`, f.IAmNotIdentified)
	ctx.Step(`^I am identified as "([^"]*)"$`, f.IAmIdentifiedAs)
	ctx.Step(`^I have these datasets:$`, f.IHaveTheseDatasets)
	ctx.Step(`^the document in the database for id "([^"]*)" should be:$`, f.TheDocumentInTheDatabaseForIdShouldBe)
}

func (f *DatasetFeature) Reset() *DatasetFeature {
	f.Datasets = make([]*models.Dataset, 0)
	f.MongoClient.Database = memongo.RandomDatabase()
	f.MongoClient.Init()
	f.Config.EnablePrivateEndpoints = false
	f.FakeAuthService.Reset()
	return f
}

func (f *DatasetFeature) Close() error {
	if f.svc != nil && f.ServiceRunning {
		f.svc.Close(context.Background())
		f.ServiceRunning = false
	}
	f.FakeAuthService.Close()
	return nil
}

func funcClose(ctx context.Context) error {
	return nil
}

func (f *DatasetFeature) DoGetHealthcheckOk(cfg *config.Configuration, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return &serviceMock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (f *DatasetFeature) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	f.HTTPServer.Addr = bindAddr
	f.HTTPServer.Handler = router
	return f.HTTPServer
}

// DoGetMongoDB returns a MongoDB
func (f *DatasetFeature) DoGetMongoDB(ctx context.Context, cfg *config.Configuration) (store.MongoDB, error) {
	return f.MongoClient, nil
}

func (f *DatasetFeature) DoGetGraphDBOk(ctx context.Context) (store.GraphDB, service.Closer, error) {
	return &storeMock.GraphDBMock{CloseFunc: funcClose}, &serviceMock.CloserMock{CloseFunc: funcClose}, nil
}

func (f *DatasetFeature) DoGetKafkaProducerOk(ctx context.Context, cfg *config.Configuration) (kafka.IProducer, error) {
	return &kafkatest.IProducerMock{
		ChannelsFunc: func() *kafka.ProducerChannels {
			return &kafka.ProducerChannels{}
		},
		CloseFunc: funcClose,
	}, nil
}

func (f *DatasetFeature) BeforeRequestHook() error {
	if err := f.svc.Run(context.Background(), "1", "", "", f.errorChan); err != nil {
		return err
	}
	f.ServiceRunning = true
	return nil
}

func (f *DatasetFeature) IHaveTheseDatasets(datasetsJson *godog.DocString) error {

	datasets := []models.Dataset{}
	m := f.MongoClient

	err := json.Unmarshal([]byte(datasetsJson.Content), &datasets)
	if err != nil {
		return err
	}
	s := m.Session.Copy()
	defer s.Close()

	for _, datasetDoc := range datasets {
		f.putDatasetInDatabase(s, datasetDoc)
	}

	return nil
}

func (f *DatasetFeature) putDatasetInDatabase(s *mgo.Session, datasetDoc models.Dataset) {
	datasetID := datasetDoc.ID

	datasetUp := models.DatasetUpdate{
		ID:      datasetID,
		Next:    &datasetDoc,
		Current: &datasetDoc,
	}

	update := bson.M{
		"$set": datasetUp,
		"$setOnInsert": bson.M{
			"last_updated": time.Now(),
		},
	}
	_, err := s.DB(f.MongoClient.Database).C("datasets").UpsertId(datasetID, update)
	if err != nil {
		panic(err)
	}
}

func (f *DatasetFeature) IAmNotIdentified() error {
	f.FakeAuthService.NewHandler().Get("/identity").Reply(401)
	return nil
}

func (f *DatasetFeature) IAmIdentifiedAs(username string) error {
	f.FakeAuthService.NewHandler().Get("/identity").Reply(200).BodyString(`{ "identifier": "` + username + `"}`)
	return nil
}

func (f *DatasetFeature) PrivateEndpointsAreEnabled() error {
	f.Config.EnablePrivateEndpoints = true
	return nil
}

func (f *DatasetFeature) TheDocumentInTheDatabaseForIdShouldBe(documentId string, documentJson *godog.DocString) error {
	s := f.MongoClient.Session.Copy()
	defer s.Close()

	var expectedDataset models.Dataset

	err := json.Unmarshal([]byte(documentJson.Content), &expectedDataset)

	filterCursor := s.DB(f.MongoClient.Database).C("datasets").FindId(documentId)

	var document models.DatasetUpdate
	err = filterCursor.One(&document)
	if err != nil {
		return err
	}

	assert.Equal(f, documentId, document.ID)
	// FIXME: either test the intersection of the 2 JSONs, or use a table for the expected
	assert.Equal(f, expectedDataset.Title, document.Next.Title)
	assert.Equal(f, "created", document.Next.State)

	return f.StepError()
}