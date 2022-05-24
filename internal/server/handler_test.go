package server

import (
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct{}

func (s *mockStorage) SaveGauge(key string, val models.Gauge)          {}
func (s *mockStorage) SaveCounter(key string, val models.Counter)      {}
func (s *mockStorage) GetGauges() map[string]models.Gauge              { return nil }
func (s *mockStorage) GetCounters() map[string]models.Counter          { return nil }
func (s *mockStorage) IncrementCounter(key string, val models.Counter) {}
func (s *mockStorage) GetCounter(string) (models.Counter, error)       { return models.Counter(0), nil }
func (s *mockStorage) GetGauge(string) (models.Gauge, error)           { return models.Gauge(0), nil }

func TestRouter(t *testing.T) {
	storage := mockStorage{}
	r := Router(&storage)
	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		method      string
		code        int
		response    string
		contentType string
		uri         string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive counter test",
			want: want{
				method:      http.MethodPost,
				code:        200,
				contentType: "text/plain",
				uri:         "/update/counter/test/2",
			},
		},
		{
			name: "positive gauge test",
			want: want{
				method:      http.MethodPost,
				code:        200,
				contentType: "text/plain",
				uri:         "/update/gauge/test/2.0",
			},
		},
		{
			name: "antipositive gauge test",
			want: want{
				method:      http.MethodGet,
				code:        405,
				contentType: "text/plain",
				uri:         "/update/gauge/test/2.0",
			},
		},
		{
			name: "antipositive test incorrect route",
			want: want{
				method:      http.MethodPost,
				code:        501,
				contentType: "text/plain",
				uri:         "/update/gauge/test2.0",
			},
		},
		{
			name: "antipositive test incorrect route 2",
			want: want{
				method:      http.MethodPost,
				code:        501,
				response:    "",
				contentType: "text/plain",
				uri:         "/update/counter/test/2/increment",
			},
		},
		{
			name: "antipositive test incorrect route 3",
			want: want{
				method:      http.MethodPost,
				code:        400,
				contentType: "text/plain",
				uri:         "/update/counter/test/one",
			},
		},
		{
			name: "positive gauge test value",
			want: want{
				method:      http.MethodGet,
				code:        200,
				response:    "0",
				contentType: "text/plain",
				uri:         "/value/gauge/test",
			},
		},
		{
			name: "positive counter test value",
			want: want{
				method:      http.MethodGet,
				code:        200,
				response:    "0",
				contentType: "text/plain",
				uri:         "/value/counter/test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.want.method, tt.want.uri)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.response != "" {
				assert.Equal(t, tt.want.response, body)
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	t.Log(method, path)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}
