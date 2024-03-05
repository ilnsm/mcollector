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
