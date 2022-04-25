package handlers

import (
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockStorage struct{}

func (s *mockStorage) SaveGuage(key string, val models.Gauge)          {}
func (s *mockStorage) SaveCounter(key string, val models.Counter)      {}
func (s *mockStorage) GetGuages() map[string]models.Gauge              { return nil }
func (s *mockStorage) GetCounters() map[string]models.Counter          { return nil }
func (s *mockStorage) IncrementCounter(key string, val models.Counter) {}

func Test_metricsSaver_SaveMetrics(t *testing.T) {
	mockStor := mockStorage{}
	testSaver := NewMetricsSaver(&mockStor)

	type want struct {
		method      string
		code        int
		response    string
		contentType string
		uri         string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive counter test",
			want: want{
				method:      http.MethodPost,
				code:        200,
				response:    "",
				contentType: "text/plain",
				uri:         "/update/counter/test/2",
			},
		},
		{
			name: "positive guage test",
			want: want{
				method:      http.MethodPost,
				code:        200,
				response:    "",
				contentType: "text/plain",
				uri:         "/update/guage/test/2.0",
			},
		},
		{
			name: "antipositive guage test",
			want: want{
				method:      http.MethodGet,
				code:        403,
				response:    "unsupported method",
				contentType: "text/plain",
				uri:         "/update/guage/test/2.0",
			},
		},
		{
			name: "antipositive test incorrect route",
			want: want{
				method:      http.MethodPost,
				code:        404,
				response:    "Not found this route",
				contentType: "text/plain",
				uri:         "/update/guage/test2.0",
			},
		},
		{
			name: "antipositive test incorrect route",
			want: want{
				method:      http.MethodPost,
				code:        404,
				response:    "Not found this route",
				contentType: "text/plain",
				uri:         "/update/counter/test/2/increment",
			},
		},
		{
			name: "antipositive test incorrect route",
			want: want{
				method:      http.MethodPost,
				code:        400,
				response:    "",
				contentType: "text/plain",
				uri:         "/update/counter/test/one",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.method, tt.want.uri, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(testSaver.SaveMetrics)
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.want.response && tt.want.response != "" {
				t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
			}

			// TODO: wait for Y.P. test is fixed
			//if res.Header.Get("Content-Type") != tt.want.contentType {
			//	t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			//}
		})
	}
}
