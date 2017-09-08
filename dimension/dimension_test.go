package dimension_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-dataset-api/api-errors"
	"github.com/ONSdigital/dp-dataset-api/dimension"
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/store/datastoretest"
	. "github.com/smartystreets/goconvey/convey"
)

const secretKey = "coffee"

var internalError = errors.New("internal error")

func createRequestWithToken(method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	r.Header.Add("internal-token", secretKey)
	return r, err
}

func TestAddNodeIDToDimensionReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Add node id to a dimension returns ok", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/dimensions/age/options/55/node_id/11", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateDimensionNodeIDFunc: func(event *models.Dimension) error {
				return nil
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.AddNodeID(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.UpdateDimensionNodeIDCalls()), ShouldEqual, 1)
	})
}

func TestAddNodeIDToDimensionReturnsBadRequest(t *testing.T) {
	t.Parallel()
	Convey("Add node id to a dimension returns bad request", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/dimensions/age/options/55/node_id/11", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateDimensionNodeIDFunc: func(event *models.Dimension) error {
				return api_errors.DimensionNodeNotFound
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.AddNodeID(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.UpdateDimensionNodeIDCalls()), ShouldEqual, 1)
	})
}

func TestAddNodeIDToDimensionReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("Add node id to a dimension returns internal error", t, func() {
		r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123/dimensions/age/options/55/node_id/11", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			UpdateDimensionNodeIDFunc: func(event *models.Dimension) error {
				return internalError
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.AddNodeID(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.UpdateDimensionNodeIDCalls()), ShouldEqual, 1)
	})
}

func TestGetDimensionNodesReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("Get dimension nodes returns ok", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetDimensionNodesFromInstanceFunc: func(id string) (*models.DimensionNodeResults, error) {
				return &models.DimensionNodeResults{}, nil
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.GetNodes(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetDimensionNodesFromInstanceCalls()), ShouldEqual, 1)
	})
}

func TestGetDimensionNodesReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("Get dimension nodes returns not found", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetDimensionNodesFromInstanceFunc: func(id string) (*models.DimensionNodeResults, error) {
				return nil, api_errors.InstanceNotFound
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.GetNodes(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetDimensionNodesFromInstanceCalls()), ShouldEqual, 1)
	})
}

func TestGetDimensionNodesReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("Get dimension nodes returns internal error", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123/dimensions", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetDimensionNodesFromInstanceFunc: func(id string) (*models.DimensionNodeResults, error) {
				return nil, internalError
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.GetNodes(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetDimensionNodesFromInstanceCalls()), ShouldEqual, 1)
	})
}

func TestGetUniqueDimensionValuesReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("Get all unique dimensions returns ok", t, func() {
		r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123/dimensions/age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()

		mockedDataStore := &storetest.StorerMock{
			GetUniqueDimensionValuesFunc: func(id, dimension string) (*models.DimensionValues, error) {
				return &models.DimensionValues{}, nil
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.GetUnique(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetUniqueDimensionValuesCalls()), ShouldEqual, 1)
	})
}

func TestGetUniqueDimensionValuesReturnsNotFound(t *testing.T) {
	t.Parallel()
	Convey("Get all unique dimensions returns not found", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/instances/123/dimensions/age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetUniqueDimensionValuesFunc: func(id, dimension string) (*models.DimensionValues, error) {
				return nil, api_errors.InstanceNotFound
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.GetUnique(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetUniqueDimensionValuesCalls()), ShouldEqual, 1)
	})
}

func TestGetUniqueDimensionValuesReturnsInternalError(t *testing.T) {
	t.Parallel()
	Convey("Get all unique dimensions returns internal error", t, func() {
		r, err := http.NewRequest("GET", "http://localhost:21800/instances/123/dimensions/age/options", nil)
		So(err, ShouldBeNil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetUniqueDimensionValuesFunc: func(id, dimension string) (*models.DimensionValues, error) {
				return nil, internalError
			},
		}

		dimension := &dimension.Store{mockedDataStore}
		dimension.GetUnique(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetUniqueDimensionValuesCalls()), ShouldEqual, 1)
	})
}
