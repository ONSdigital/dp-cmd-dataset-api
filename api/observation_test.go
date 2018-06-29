package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	errs "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/mocks"
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/store/datastoretest"
	"github.com/ONSdigital/dp-filter/observation"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/audit/audit_mock"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	dimension1 = models.CodeList{Name: "aggregate"}
	dimension2 = models.CodeList{Name: "geography"}
	dimension3 = models.CodeList{Name: "time"}
	dimension4 = models.CodeList{Name: "age"}
)

func TestGetObservationsReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Given a request to get a single observation for a version of a dataset returns 200 OK response", t, func() {

		dimensions := []models.CodeList{
			models.CodeList{
				Name: "aggregate",
				HRef: "http://localhost:8081/code-lists/cpih1dim1aggid",
			},
			models.CodeList{
				Name: "geography",
				HRef: "http://localhost:8081/code-lists/uk-only",
			},
			models.CodeList{
				Name: "time",
				HRef: "http://localhost:8081/code-lists/time",
			},
		}
		usagesNotes := &[]models.UsageNote{models.UsageNote{Title: "data_marking", Note: "this marks the obsevation with a special character"}}

		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(string, string, string, string) (*models.Version, error) {
				return &models.Version{
					Dimensions: dimensions,
					Headers:    []string{"v4_2", "data_marking", "confidence_interval", "aggregate_code", "aggregate", "geography_code", "geography", "time", "time"},
					Links: &models.VersionLinks{
						Version: &models.LinkObject{
							HRef: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:   "1",
						},
					},
					State:      models.PublishedState,
					UsageNotes: usagesNotes,
				}, nil
			},
		}

		count := 0
		mockRowReader := &mocks.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				count++
				if count == 1 {
					return "v4_2,data_marking,confidence_interval,time,time,geography_code,geography,aggregate_code,aggregate", nil
				} else if count == 2 {
					return "146.3,p,2,Month,Aug-16,K02000001,,cpi1dim1G10100,01.1 Food", nil
				}
				return "", io.EOF
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockedObservationStore := &mocks.ObservationStoreMock{
			GetCSVRowsFunc: func(*observation.Filter, *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)

		Convey("When request contains query parameters where the dimension name is in lower casing", func() {
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()
			api.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, getTestData("expectedDocWithSingleObservation"))

			So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
			So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
			So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
			So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 1)
			So(len(mockRowReader.ReadCalls()), ShouldEqual, 3)

			ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
			So(len(auditMock.RecordCalls()), ShouldEqual, 2)
			auditMock.AssertRecordCalls(
				audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
				audit_mock.Expected{Action: getObservationsAction, Result: audit.Successful, Params: ap},
			)
		})

		Convey("When request contains query parameters where the dimension name is in upper casing", func() {
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&AggregaTe=cpi1dim1S40403&GEOGRAPHY=K02000001", nil)
			w := httptest.NewRecorder()
			api.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, getTestData("expectedSecondDocWithSingleObservation"))

			So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
			So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
			So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
			So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 1)
			So(len(mockRowReader.ReadCalls()), ShouldEqual, 3)

			ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
			So(len(auditMock.RecordCalls()), ShouldEqual, 2)
			auditMock.AssertRecordCalls(
				audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
				audit_mock.Expected{Action: getObservationsAction, Result: audit.Successful, Params: ap},
			)
		})
	})

	Convey("A successful request to get multiple observations via a wildcard for a version of a dataset returns 200 OK response", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=*&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dimensions := []models.CodeList{
			models.CodeList{
				Name: "aggregate",
				HRef: "http://localhost:8081/code-lists/cpih1dim1aggid",
			},
			models.CodeList{
				Name: "geography",
				HRef: "http://localhost:8081/code-lists/uk-only",
			},
			models.CodeList{
				Name: "time",
				HRef: "http://localhost:8081/code-lists/time",
			},
		}
		usagesNotes := &[]models.UsageNote{models.UsageNote{Title: "data_marking", Note: "this marks the observation with a special character"}}

		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(string, string, string, string) (*models.Version, error) {
				return &models.Version{
					Dimensions: dimensions,
					Headers:    []string{"v4_2", "data_marking", "confidence_interval", "aggregate_code", "aggregate", "geography_code", "geography", "time", "time"},
					Links: &models.VersionLinks{
						Version: &models.LinkObject{
							HRef: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:   "1",
						},
					},
					State:      models.PublishedState,
					UsageNotes: usagesNotes,
				}, nil
			},
		}

		count := 0
		mockRowReader := &mocks.CSVRowReaderMock{
			ReadFunc: func() (string, error) {
				count++
				if count == 1 {
					return "v4_2,data_marking,confidence_interval,time,time,geography_code,geography,aggregate_code,aggregate", nil
				} else if count == 2 {
					return "146.3,p,2,Month,Aug-16,K02000001,,cpi1dim1G10100,01.1 Food", nil
				} else if count == 3 {
					return "112.1,,,Month,Aug-16,K02000001,,cpi1dim1G10101,01.2 Waste", nil
				}
				return "", io.EOF
			},
			CloseFunc: func() error {
				return nil
			},
		}

		mockedObservationStore := &mocks.ObservationStoreMock{
			GetCSVRowsFunc: func(*observation.Filter, *int) (observation.CSVRowReader, error) {
				return mockRowReader, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Body.String(), ShouldContainSubstring, getTestData("expectedDocWithMultipleObservations"))

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
		So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 1)
		So(len(mockRowReader.ReadCalls()), ShouldEqual, 4)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Successful, Params: ap},
		)
	})
}

func TestGetObservationsReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to mongo datastore return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return nil, errs.ErrInternalServer
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldResemble, "internal error\n")

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 0)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When the dataset does not exist return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return nil, errs.ErrDatasetNotFound
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 0)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When the dataset exists but is unpublished return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 0)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When the edition of a dataset does not exist return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return errs.ErrEditionNotFound
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrEditionNotFound.Error())

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionsCalls()), ShouldEqual, 0)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When version does not exist for an edition of a dataset returns status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return nil, errs.ErrVersionNotFound
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When an unpublished version has an incorrect state for an edition of a dataset return an internal error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{State: "gobbly-gook"}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		assertInternalServerErr(w)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When a version document has not got a headers field return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{
					Dimensions: []models.CodeList{dimension1, dimension2, dimension3},
					State:      models.PublishedState,
				}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When a version document has not got any dimensions field return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{
					Headers: []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate"},
					State:   models.PublishedState,
				}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When the first header in array does not describe the header row correctly return internal error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{
					Dimensions: []models.CodeList{dimension1, dimension2, dimension3},
					Headers:    []string{"v4"},
					State:      models.PublishedState,
				}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		assertInternalServerErr(w)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When an invalid query parameter is set in request return 400 bad request with an error message containing a list of incorrect query parameters", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{
					Dimensions: []models.CodeList{dimension1, dimension3},
					Headers:    []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate"},
					State:      models.PublishedState,
				}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "Incorrect selection of query parameters: [geography], these dimensions do not exist for this version of the dataset\n")

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When there is a missing query parameter that is expected to be set in request return 400 bad request with an error message containing a list of missing query parameters", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{
					Dimensions: []models.CodeList{dimension1, dimension2, dimension3, dimension4},
					Headers:    []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate", "geography_code", "geography", "age_code", "age"},
					State:      models.PublishedState,
				}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "Missing query parameters for the following dimensions: [age]\n")

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When there are too many query parameters that are set to wildcard (*) value request returns 400 bad request", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=*&aggregate=*&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{
					Dimensions: []models.CodeList{dimension1, dimension2, dimension3},
					Headers:    []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate", "geography_code", "geography"},
					State:      models.PublishedState,
				}, nil
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, genericMockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "only one wildcard (*) is allowed as a value in selected query parameters\n")

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When requested query does not find a unique observation return no observations found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(datasetID string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{
						Dimensions: []models.CodeList{dimension1, dimension2, dimension3},
						Headers:    []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate", "geography_code", "geography"},
						State:      models.PublishedState,
					},
					nil
			},
		}

		mockedObservationStore := &mocks.ObservationStoreMock{
			GetCSVRowsFunc: func(*observation.Filter, *int) (observation.CSVRowReader, error) {
				return nil, errs.ErrObservationsNotFound
			},
		}

		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrObservationsNotFound.Error())

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
		So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 1)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})

	Convey("When requested query has a multi-valued dimension return bad request", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001&geography=K02000002", nil)
		w := httptest.NewRecorder()

		dimensions := []models.CodeList{
			models.CodeList{
				Name: "aggregate",
				HRef: "http://localhost:8081/code-lists/cpih1dim1aggid",
			},
			models.CodeList{
				Name: "geography",
				HRef: "http://localhost:8081/code-lists/uk-only",
			},
			models.CodeList{
				Name: "time",
				HRef: "http://localhost:8081/code-lists/time",
			},
		}
		usagesNotes := &[]models.UsageNote{models.UsageNote{Title: "data_marking", Note: "this marks the obsevation with a special character"}}

		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
			},
			CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
				return nil
			},
			GetVersionFunc: func(string, string, string, string) (*models.Version, error) {
				return &models.Version{
					Dimensions: dimensions,
					Headers:    []string{"v4_2", "data_marking", "confidence_interval", "aggregate_code", "aggregate", "geography_code", "geography", "time", "time"},
					Links: &models.VersionLinks{
						Version: &models.LinkObject{
							HRef: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:   "1",
						},
					},
					State:      models.PublishedState,
					UsageNotes: usagesNotes,
				}, nil
			},
		}

		mockedObservationStore := &mocks.ObservationStoreMock{}
		auditMock := audit_mock.New()
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "Multi-valued query parameters for the following dimensions: [geography]\n")

		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
		So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 0)

		ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
		So(len(auditMock.RecordCalls()), ShouldEqual, 2)
		auditMock.AssertRecordCalls(
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
			audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
		)
	})
}

func TestGetListOfValidDimensionNames(t *testing.T) {
	t.Parallel()
	Convey("Given a list of valid dimension codelist objects", t, func() {
		Convey("When getListOfValidDimensionNames is called", func() {
			dimension1 := models.CodeList{
				Name: "time",
			}

			dimension2 := models.CodeList{
				Name: "aggregate",
			}

			dimension3 := models.CodeList{
				Name: "geography",
			}

			version := &models.Version{
				Dimensions: []models.CodeList{dimension1, dimension2, dimension3},
			}

			Convey("Then func returns the correct number of dimensions", func() {
				validDimensions := getListOfValidDimensionNames(version.Dimensions)

				So(len(validDimensions), ShouldEqual, 3)
				So(validDimensions[0], ShouldEqual, "time")
				So(validDimensions[1], ShouldEqual, "aggregate")
				So(validDimensions[2], ShouldEqual, "geography")
			})
		})
	})
}

func TestGetDimensionOffsetInHeaderRow(t *testing.T) {
	t.Parallel()
	Convey("Given the version headers are valid", t, func() {
		Convey("When the version has no metadata headers", func() {
			version := &models.Version{
				Headers: []string{
					"v4_0",
					"time_codelist",
					"time",
					"aggregate_codelist",
					"Aggregate",
					"geography_codelist",
					"geography",
				},
			}

			Convey("Then getListOfValidDimensionNames func returns the correct number of headers", func() {
				dimensionOffset, err := getDimensionOffsetInHeaderRow(version.Headers)

				So(err, ShouldBeNil)
				So(dimensionOffset, ShouldEqual, 0)
			})
		})

		Convey("When the version has metadata headers", func() {
			version := &models.Version{
				Headers: []string{
					"V4_2",
					"data_marking",
					"confidence_interval",
					"time_codelist",
					"time",
				},
			}

			Convey("Then getListOfValidDimensionNames func returns the correct number of headers", func() {
				dimensionOffset, err := getDimensionOffsetInHeaderRow(version.Headers)

				So(err, ShouldBeNil)
				So(dimensionOffset, ShouldEqual, 2)
			})
		})
	})

	Convey("Given the first value in the header does not have an underscore `_` in value", t, func() {
		Convey("When the getListOfValidDimensionNames func is called", func() {
			version := &models.Version{
				Headers: []string{
					"v4",
					"time_codelist",
					"time",
					"aggregate_codelist",
					"aggregate",
					"geography_codelist",
					"geography",
				},
			}
			Convey("Then function returns error, `index out of range`", func() {
				dimensionOffset, err := getDimensionOffsetInHeaderRow(version.Headers)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "index out of range")
				So(dimensionOffset, ShouldEqual, 0)
			})
		})
	})

	Convey("Given the first value in the header does not follow the format `v4_1`", t, func() {
		Convey("When the getListOfValidDimensionNames func is called", func() {
			version := &models.Version{
				Headers: []string{
					"v4_one",
					"time_codelist",
					"time",
					"aggregate_codelist",
					"aggregate",
					"geography_codelist",
					"geography",
				},
			}
			Convey("Then function returns error, `index out of range`", func() {
				dimensionOffset, err := getDimensionOffsetInHeaderRow(version.Headers)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "strconv.Atoi: parsing \"one\": invalid syntax")
				So(dimensionOffset, ShouldEqual, 0)
			})
		})
	})
}

func TestExtractQueryParameters(t *testing.T) {
	t.Parallel()
	Convey("Given a list of valid dimension headers for version", t, func() {
		headers := []string{
			"time",
			"aggregate",
			"geography",
		}

		Convey("When a request is made containing query parameters for each dimension/header", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&aggregate=Overall Index&geography=wales",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns a list of query parameters and their corresponding value", func() {
				queryParameters, err := extractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldBeNil)
				So(len(queryParameters), ShouldEqual, 3)
				So(queryParameters["time"], ShouldEqual, "JAN08")
				So(queryParameters["aggregate"], ShouldEqual, "Overall Index")
				So(queryParameters["geography"], ShouldEqual, "wales")
			})
		})

		Convey("When a request is made containing query parameters for 2/3 dimensions/headers", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&geography=wales",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns an error", func() {
				queryParameters, err := extractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errorMissingQueryParameters([]string{"aggregate"}))
				So(queryParameters, ShouldBeNil)
			})
		})

		Convey("When a request is made containing all query parameters for each dimensions/headers but also an invalid one", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&aggregate=Food&geography=wales&age=52",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns an error", func() {
				queryParameters, err := extractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errorIncorrectQueryParameters([]string{"age"}))
				So(queryParameters, ShouldBeNil)
			})
		})

		Convey("When a request is made containing all query parameters for each dimensions/headers but there is a duplicate", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&aggregate=Food&geography=wales&time=JAN0",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns an error", func() {
				queryParameters, err := extractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errorMultivaluedQueryParameters([]string{"time"}))
				So(queryParameters, ShouldBeNil)
			})
		})
	})
}

func TestGetObservationAuditAttemptedError(t *testing.T) {
	Convey("given audit action attempted returns an error", t, func() {
		auditMock := audit_mock.NewErroring(getObservationsAction, audit.Attempted)

		mockedDataStore := &storetest.StorerMock{}
		mockedObservationStore := &mocks.ObservationStoreMock{}
		api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)

		Convey("when get observation is called", func() {
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()

			api.Router.ServeHTTP(w, r)

			Convey("then a 500 response status is returned", func() {
				assertInternalServerErr(w)
				So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.GetVersionsCalls()), ShouldEqual, 0)

				So(len(auditMock.RecordCalls()), ShouldEqual, 1)
				auditMock.AssertRecordCalls(
					audit_mock.Expected{
						Action: getObservationsAction,
						Result: audit.Attempted,
						Params: common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}},
				)
			})
		})
	})
}

func TestGetObservationAuditUnsuccessfulError(t *testing.T) {
	ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}

	Convey("given audit action unsuccessful returns an error", t, func() {

		Convey("when datastore.getDataset returns an error", func() {
			auditMock := audit_mock.NewErroring(getObservationsAction, audit.Unsuccessful)

			mockedDataStore := &storetest.StorerMock{
				GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
					return nil, errs.ErrDatasetNotFound
				},
			}

			mockedObservationStore := &mocks.ObservationStoreMock{}
			api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()

			api.Router.ServeHTTP(w, r)

			Convey("then a 500 response status is returned", func() {
				assertInternalServerErr(w)
				So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 0)
				So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 0)

				So(len(auditMock.RecordCalls()), ShouldEqual, 2)
				auditMock.AssertRecordCalls(
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
				)
			})
		})

		Convey("when datastore.getEdition returns an error", func() {
			auditMock := audit_mock.NewErroring(getObservationsAction, audit.Unsuccessful)

			mockedDataStore := &storetest.StorerMock{
				GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
					return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
				},
				CheckEditionExistsFunc: func(ID string, editionID string, state string) error {
					return errs.ErrEditionNotFound
				},
			}

			mockedObservationStore := &mocks.ObservationStoreMock{}
			api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()

			api.Router.ServeHTTP(w, r)

			Convey("then a 500 response status is returned", func() {
				assertInternalServerErr(w)
				So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 0)
				So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 0)

				So(len(auditMock.RecordCalls()), ShouldEqual, 2)
				auditMock.AssertRecordCalls(
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
				)
			})
		})

		Convey("when datastore.getVersion returns an error", func() {
			auditMock := audit_mock.NewErroring(getObservationsAction, audit.Unsuccessful)

			mockedDataStore := &storetest.StorerMock{
				GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
					return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
				},
				CheckEditionExistsFunc: func(ID string, editionID string, state string) error {
					return nil
				},
				GetVersionFunc: func(datasetID string, editionID string, version string, state string) (*models.Version, error) {
					return nil, errs.ErrVersionNotFound
				},
			}

			mockedObservationStore := &mocks.ObservationStoreMock{}
			api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()

			api.Router.ServeHTTP(w, r)

			Convey("then a 500 response status is returned", func() {
				assertInternalServerErr(w)
				So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
				So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 0)

				So(len(auditMock.RecordCalls()), ShouldEqual, 2)
				auditMock.AssertRecordCalls(
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
				)
			})
		})

		Convey("when the version does not have no header data", func() {
			auditMock := audit_mock.NewErroring(getObservationsAction, audit.Unsuccessful)

			mockedDataStore := &storetest.StorerMock{
				GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
					return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
				},
				CheckEditionExistsFunc: func(ID string, editionID string, state string) error {
					return nil
				},
				GetVersionFunc: func(datasetID string, editionID string, version string, state string) (*models.Version, error) {
					return &models.Version{}, nil
				},
			}

			mockedObservationStore := &mocks.ObservationStoreMock{}
			api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()

			api.Router.ServeHTTP(w, r)

			Convey("then a 500 response status is returned", func() {
				assertInternalServerErr(w)
				So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
				So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 0)

				So(len(auditMock.RecordCalls()), ShouldEqual, 2)
				auditMock.AssertRecordCalls(
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Unsuccessful, Params: ap},
				)
			})
		})
	})
}

func TestGetObservationAuditSuccessfulError(t *testing.T) {
	Convey("given audit action successful returns an error", t, func() {
		auditMock := audit_mock.NewErroring(getObservationsAction, audit.Successful)

		Convey("when get observations is called with a valid request", func() {

			dimensions := []models.CodeList{
				models.CodeList{
					Name: "aggregate",
					HRef: "http://localhost:8081/code-lists/cpih1dim1aggid",
				},
				models.CodeList{
					Name: "geography",
					HRef: "http://localhost:8081/code-lists/uk-only",
				},
				models.CodeList{
					Name: "time",
					HRef: "http://localhost:8081/code-lists/time",
				},
			}
			usagesNotes := &[]models.UsageNote{models.UsageNote{Title: "data_marking", Note: "this marks the obsevation with a special character"}}

			mockedDataStore := &storetest.StorerMock{
				GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
					return &models.DatasetUpdate{Current: &models.Dataset{State: models.PublishedState}}, nil
				},
				CheckEditionExistsFunc: func(datasetID, editionID, state string) error {
					return nil
				},
				GetVersionFunc: func(string, string, string, string) (*models.Version, error) {
					return &models.Version{
						Dimensions: dimensions,
						Headers:    []string{"v4_2", "data_marking", "confidence_interval", "aggregate_code", "aggregate", "geography_code", "geography", "time", "time"},
						Links: &models.VersionLinks{
							Version: &models.LinkObject{
								HRef: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
								ID:   "1",
							},
						},
						State:      models.PublishedState,
						UsageNotes: usagesNotes,
					}, nil
				},
			}

			count := 0
			mockRowReader := &mocks.CSVRowReaderMock{
				ReadFunc: func() (string, error) {
					count++
					if count == 1 {
						return "v4_2,data_marking,confidence_interval,time,time,geography_code,geography,aggregate_code,aggregate", nil
					} else if count == 2 {
						return "146.3,p,2,Month,Aug-16,K02000001,,cpi1dim1G10100,01.1 Food", nil
					}
					return "", io.EOF
				},
				CloseFunc: func() error {
					return nil
				},
			}

			mockedObservationStore := &mocks.ObservationStoreMock{
				GetCSVRowsFunc: func(*observation.Filter, *int) (observation.CSVRowReader, error) {
					return mockRowReader, nil
				},
			}

			api := GetAPIWithMockedDatastore(mockedDataStore, &mocks.DownloadsGeneratorMock{}, auditMock, mockedObservationStore)
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()

			api.Router.ServeHTTP(w, r)

			Convey("then a 500 status response is returned", func() {
				assertInternalServerErr(w)
				So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.CheckEditionExistsCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
				So(len(mockedObservationStore.GetCSVRowsCalls()), ShouldEqual, 1)

				ap := common.Params{"dataset_id": "cpih012", "edition": "2017", "version": "1"}
				So(len(auditMock.RecordCalls()), ShouldEqual, 2)
				auditMock.AssertRecordCalls(
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Attempted, Params: ap},
					audit_mock.Expected{Action: getObservationsAction, Result: audit.Successful, Params: ap},
				)
			})

		})
	})
}

func getTestData(filename string) string {
	jsonBytes, err := ioutil.ReadFile("./observation_test_data/" + filename + ".json")
	if err != nil {
		log.ErrorC("unable to read json file into bytes", err, log.Data{"filename": filename})
		os.Exit(1)
	}
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, jsonBytes); err != nil {
		log.ErrorC("unable to remove whitespace from json bytes", err, log.Data{"filename": filename})
		os.Exit(1)
	}

	return buffer.String()
}
