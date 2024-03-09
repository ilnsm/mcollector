package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
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

func TestListAllMetrics(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	type testCase struct {
		wantBody        string
		wantStatus      int
		setup           func(*testCase)
		storage         *mock_transport.MockStorage
		wantContentType string
	}

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "No metrics",
			tc: testCase{
				wantBody:        "\n<!DOCTYPE html>\n<html>\n<head>\n\n\t<title>Metric's' Data</title>\n\n</head>\n<body>\n\n\t   <h1>Data</h1>\n\t   <ul>\n\t   \n\t   </ul>\n\n\n</body>\n</html>\n",
				wantStatus:      http.StatusOK,
				wantContentType: "text/html",
				setup: func(tc *testCase) {
					gauges := make(map[string]float64)
					counters := make(map[string]int64)
					tc.storage.EXPECT().GetGauges(gomock.Any()).Return(gauges, nil).Times(1)
					tc.storage.EXPECT().GetCounters(gomock.Any()).Return(counters, nil).Times(1)
				},
			},
		},
		{
			name: "Positive test",
			tc: testCase{
				wantBody:        "\n<!DOCTYPE html>\n<html>\n<head>\n\n\t<title>Metric's' Data</title>\n\n</head>\n<body>\n\n\t   <h1>Data</h1>\n\t   <ul>\n\t   \n\t       <li>couner_1: 534</li>\n\t   \n\t       <li>couner_2: 11</li>\n\t   \n\t       <li>gauge_1: 54.12</li>\n\t   \n\t       <li>gauge_2: 1092.2</li>\n\t   \n\t   </ul>\n\n\n</body>\n</html>",
				wantStatus:      http.StatusOK,
				wantContentType: "text/html",
				setup: func(tc *testCase) {
					gauges := make(map[string]float64, 2)
					gauges["gauge_1"] = 54.12
					gauges["gauge_2"] = 1092.2
					//
					counters := make(map[string]int64, 2)
					counters["couner_1"] = 534
					counters["couner_2"] = 11

					tc.storage.EXPECT().GetGauges(gomock.Any()).Return(gauges, nil).Times(1)
					tc.storage.EXPECT().GetCounters(gomock.Any()).Return(counters, nil).Times(1)
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			test.tc.storage = mock_transport.NewMockStorage(mockCtl)
			test.tc.setup(&test.tc)

			a := &API{Storage: test.tc.storage}
			handler := ListAllMetrics(a)
			handler.ServeHTTP(w, request)

			if ct := w.Header().Get("Content-Type"); ct != test.tc.wantContentType {
				t.Errorf("handler returned wrong content-type: got %v\n want %v", ct, test.tc.wantContentType)
			}
			if body := normalizeHTML(w.Body.String()); body != normalizeHTML(test.tc.wantBody) {
				t.Errorf("handler returned wrong body: got %v want %v", body, test.tc.wantBody)
			}
		})
	}
}

func normalizeHTML(html string) string {
	html = strings.TrimSpace(html)
	html = strings.ReplaceAll(html, "\n", "")
	html = strings.ReplaceAll(html, "\t", "")
	return html
}

func TestUpdateTheMetricWithJSON(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	type testCase struct {
		sendBody        string
		wantBody        string
		wantStatus      int
		setup           func(*testCase)
		storage         *mock_transport.MockStorage
		wantContentType string
		sendContentType string
	}

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Positive test gauge",
			tc: testCase{
				sendBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().InsertGauge(gomock.Any(), "gaugeNameJSON", 38.988).Times(1)
					tc.storage.EXPECT().SelectGauge(gomock.Any(), "gaugeNameJSON").Return(38.988, nil).Times(1)
				},
				wantContentType: applicationJSON,
				sendContentType: applicationJSON,
			},
		},
		{
			name: "Positive test counter",
			tc: testCase{
				sendBody: `{"id": "counter_foo",
					"type": "counter",
					"delta": 92}`,
				wantBody: `{"id": "counter_foo",
					"type": "counter",
					"delta": 92}`,
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().InsertCounter(gomock.Any(), "counter_foo", int64(92)).Times(1)
					tc.storage.EXPECT().SelectCounter(gomock.Any(), "counter_foo").Return(int64(92), nil).Times(1)
				},
				wantContentType: applicationJSON,
				sendContentType: applicationJSON,
			},
		},
		{
			name: "Content type test",
			tc: testCase{
				sendBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().InsertGauge(gomock.Any(), "gaugeNameJSON", 38.988).Times(0)
				},
				wantContentType: "text/plain; charset=utf-8",
				sendContentType: "bad-content-type",
			},
		}, {
			name: "Wrong metric type",
			tc: testCase{
				sendBody: `{"id": "counter_foo",
					"type": "metric",
					"delta": 92}`,
				wantBody:   ``,
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().InsertCounter(gomock.Any(), "counter_foo", int64(92)).Times(0)
					tc.storage.EXPECT().SelectCounter(gomock.Any(), "counter_foo").Return(int64(92), nil).Times(0)
				},
				wantContentType: "text/plain; charset=utf-8",
				sendContentType: applicationJSON,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.tc.storage = mock_transport.NewMockStorage(mockCtl)
			test.tc.setup(&test.tc)
			a := &API{Storage: test.tc.storage}

			w := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/update", nil)

			reader := strings.NewReader(test.tc.sendBody)
			body := io.NopCloser(reader)
			request.Body = body
			request.Header.Set(contentType, test.tc.sendContentType)
			handler := UpdateTheMetricWithJSON(a)

			handler.ServeHTTP(w, request)

			if reflect.DeepEqual(w.Body, body) {
				t.Errorf("handler returned wrong body: got %v want %v", w.Body, body)
			}

			if ct := w.Header().Get("Content-Type"); ct != test.tc.wantContentType {
				t.Errorf("handler returned wrong content-type: got %v\n want %v", ct, test.tc.wantContentType)

				if code := w.Code; code != test.tc.wantStatus {
					t.Errorf("handler returned wrong status code: got %v want %v", code, test.tc.wantStatus)
				}
			}
		})
	}
}

func TestGetTheMetricWithJSON(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	type testCase struct {
		sendBody        string
		wantBody        string
		wantStatus      int
		setup           func(*testCase)
		storage         *mock_transport.MockStorage
		wantContentType string
		sendContentType string
	}

	tests := []struct {
		name string
		tc   testCase
	}{
		{
			name: "Positive test gauge",
			tc: testCase{
				sendBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().SelectGauge(gomock.Any(), "gaugeNameJSON").Return(38.988, nil).Times(1)
				},
				wantContentType: applicationJSON,
				sendContentType: applicationJSON,
			},
		},
		{
			name: "Positive test counter",
			tc: testCase{
				sendBody: `{"id": "counter_foo",
					"type": "counter",
					"delta": 92}`,
				wantBody: `{"id": "counter_foo",
					"type": "counter",
					"delta": 92}`,
				wantStatus: http.StatusOK,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().SelectCounter(gomock.Any(), "counter_foo").Return(int64(92), nil).Times(1)
				},
				wantContentType: applicationJSON,
				sendContentType: applicationJSON,
			},
		},
		{
			name: "Content type test",
			tc: testCase{
				sendBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantBody: `{"id": "gaugeNameJSON",
					"type": "gauge",
					"value": 38.988}`,
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().SelectGauge(gomock.Any(), "gaugeNameJSON").Return(38.988, nil).Times(0)
				},
				wantContentType: "text/plain; charset=utf-8",
				sendContentType: "bad-content-type",
			},
		}, {
			name: "Wrong metric type",
			tc: testCase{
				sendBody: `{"id": "counter_foo",
					"type": "metric",
					"delta": 92}`,
				wantBody:   ``,
				wantStatus: http.StatusBadRequest,
				setup: func(tc *testCase) {
					tc.storage.EXPECT().SelectCounter(gomock.Any(), "counter_foo").Return(int64(92), nil).Times(0)
				},
				wantContentType: "text/plain; charset=utf-8",
				sendContentType: applicationJSON,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.tc.storage = mock_transport.NewMockStorage(mockCtl)
			test.tc.setup(&test.tc)
			a := &API{Storage: test.tc.storage}

			w := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)

			reader := strings.NewReader(test.tc.sendBody)
			body := io.NopCloser(reader)
			request.Body = body
			request.Header.Set(contentType, test.tc.sendContentType)
			handler := GetTheMetricWithJSON(a)

			handler.ServeHTTP(w, request)

			if reflect.DeepEqual(w.Body, body) {
				t.Errorf("handler returned wrong body: got %v want %v", w.Body, body)
			}

			if ct := w.Header().Get("Content-Type"); ct != test.tc.wantContentType {
				t.Errorf("handler returned wrong content-type: got %v\n want %v", ct, test.tc.wantContentType)

				if code := w.Code; code != test.tc.wantStatus {
					t.Errorf("handler returned wrong status code: got %v want %v", code, test.tc.wantStatus)
				}
			}
		})
	}
}
