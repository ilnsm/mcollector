package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	mock_transport "github.com/ospiem/mcollector/internal/mock"
	"github.com/ospiem/mcollector/internal/models"
	"go.uber.org/mock/gomock"
)

var errNotFound = errors.New("value not found")

func TestUpdateTheMetric(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	type testCase struct {
		setup      func(*testCase)
		storage    *mock_transport.MockStorage
		mType      string
		mName      string
		mValue     string
		wantStatus int
	}

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Positive test with gauge 1",
			tc: testCase{
				mType:      models.Gauge,
				mName:      "test_gauge",
				mValue:     "39.2",
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					v, _ := strconv.ParseFloat(tc.mValue, 64)
					tc.storage.EXPECT().InsertGauge(gomock.Any(), tc.mName, v).Return(nil).Times(1)
				},
			},
		},
		{
			name: "Positive test with counter 1",
			tc: testCase{
				mType:      models.Counter,
				mName:      "test_counter",
				mValue:     "190",
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					v, _ := strconv.ParseInt(tc.mValue, 10, 64)
					tc.storage.EXPECT().InsertCounter(gomock.Any(), tc.mName, v).Return(nil).Times(1)
				},
			},
		},
		{
			name: "Negative test with gauge 1",
			tc: testCase{
				mType:      models.Gauge,
				mName:      "test_gauge",
				mValue:     "dfe",
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().InsertGauge(gomock.Any(), tc.mName, float64(0)).Return(nil).Times(1)
				},
			},
		},
		{
			name: "Negative test with counter 1",
			tc: testCase{
				mType:      models.Counter,
				mName:      "test_counter",
				mValue:     "190.134",
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().InsertCounter(gomock.Any(), tc.mName, int64(0)).Return(nil).Times(1)
				},
			},
		},

		{
			name: "Invalid metric type",
			tc: testCase{
				mType:      "ggauge",
				mName:      "test_gauge",
				mValue:     "10.22",
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().InsertGauge(gomock.Any(), tc.mName, float64(0)).Return(nil).Times(0)
				},
			},
		},
	}
	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/update/{mType}/{mName}/{mValue}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("mType", test.tc.mType)
			rctx.URLParams.Add("mName", test.tc.mName)
			rctx.URLParams.Add("mValue", test.tc.mValue)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			test.tc.storage = mock_transport.NewMockStorage(mockCtl)
			test.tc.setup(&test.tc)
			a := &API{Storage: test.tc.storage}
			handler := UpdateTheMetric(a)
			handler.ServeHTTP(w, request)

			if status := w.Code; status != test.tc.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.tc.wantStatus)
			}
		})
	}
}

func TestGetTheMetric(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	type testCase struct {
		setup      func(*testCase)
		storage    *mock_transport.MockStorage
		mType      string
		mName      string
		wantBody   string
		wantStatus int
	}

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Positive test with gauge 1",
			tc: testCase{
				mType:      models.Gauge,
				mName:      "test_gauge",
				wantBody:   "112.5",
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					v, _ := strconv.ParseFloat(tc.wantBody, 64)
					tc.storage.EXPECT().SelectGauge(gomock.Any(), tc.mName).Return(v, nil).Times(1)
				},
			},
		},
		{
			name: "Positive test with counter 1",
			tc: testCase{
				mType:      models.Counter,
				mName:      "test_counter",
				wantBody:   "1982",
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					v, _ := strconv.ParseInt(tc.wantBody, 10, 64)
					tc.storage.EXPECT().SelectCounter(gomock.Any(), tc.mName).Return(v, nil).Times(1)
				},
			},
		},
		{
			name: "Negative test with gauge 1",
			tc: testCase{
				mType:      models.Gauge,
				mName:      "gauge2",
				wantBody:   "404 page not found\n",
				wantStatus: http.StatusNotFound,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().SelectGauge(gomock.Any(), tc.mName).Return(float64(0), errNotFound).Times(1)
				},
			},
		},
		{
			name: "Negative test with counter 1",
			tc: testCase{
				mType:      models.Counter,
				mName:      "test_counter",
				wantBody:   "404 page not found\n",
				wantStatus: http.StatusNotFound,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().SelectCounter(gomock.Any(), tc.mName).Return(int64(0), errNotFound).Times(1)
				},
			},
		},
		{
			name: "Invalid metric type",
			tc: testCase{
				mType:      "ccounter",
				mName:      "test_counter",
				wantBody:   "",
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().SelectCounter(gomock.Any(), tc.mName).Return(int64(0), errNotFound).Times(0)
				},
			},
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/value/{mType}/{mName}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("mType", test.tc.mType)
			rctx.URLParams.Add("mName", test.tc.mName)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			test.tc.storage = mock_transport.NewMockStorage(mockCtl)
			test.tc.setup(&test.tc)
			a := &API{Storage: test.tc.storage}
			handler := GetTheMetric(a)
			handler.ServeHTTP(w, request)

			if status := w.Code; status != test.tc.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.tc.wantStatus)
			}

			if body := w.Body.String(); body != test.tc.wantBody {
				t.Errorf("handler returned wrong body: got %v want %v", body, test.tc.wantBody)
			}
		})
	}
}

func TestPingDB(t *testing.T) {

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	type testCase struct {
		storage  *mock_transport.MockStorage
		setup    func(testCase2 *testCase)
		wantCode int
	}

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "DB available",
			tc: testCase{
				setup: func(tc *testCase) {
					tc.storage.EXPECT().Ping(gomock.Any()).Return(nil)
				},
				wantCode: http.StatusOK,
			},
		},
		{
			name: "DB unavailable",
			tc: testCase{
				setup: func(tc *testCase) {
					tc.storage.EXPECT().Ping(gomock.Any()).Return(errors.New("DB unavailable"))
				},
				wantCode: http.StatusInternalServerError,
			},
		},
	}

	for _, test := range tests {
		t.Run("ping", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			test.tc.storage = mock_transport.NewMockStorage(mockCtl)
			test.tc.setup(&test.tc)
			a := &API{Storage: test.tc.storage}
			handler := PingDB(a)
			handler.ServeHTTP(w, request)

			if status := w.Code; status != test.tc.wantCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.tc.wantCode)
			}
		})
	}
}
