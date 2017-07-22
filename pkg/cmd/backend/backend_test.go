package backend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var getTests = []struct {
	url          string
	status       int
	responseBody *string
}{
	{"/", http.StatusOK, nil},
	{"/test", http.StatusOK, nil},
	{"/apis-explorer", http.StatusOK, nil},
	//	{"/api/v1/echo/fubar", http.StatusOK, nil},
	//	{"/api/v1/", http.StatusOK, nil},
	//	{"/api/v1", http.StatusOK, nil},
}

func TestWalkURLs(t *testing.T) {
	logger := getLogger()

	handler, ok := newServer(logger).(*mux.Router)
	if !ok {
		t.Fatalf("unable to cast to mux.Router")
	}

	handler.Walk(func(r *mux.Route, m *mux.Router, ancestors []*mux.Route) error {
		path, err := r.GetPathTemplate()
		if err != nil {
			t.Fatalf("GetPathTemplate failed -- %s", err)
		}

		t.Logf("Walker: %s\n", path)
		return nil
	})
}

func TestGetURLs(t *testing.T) {

	logger := getLogger()

	for _, tt := range getTests {
		// create a request to pass to the handler
		req, err := http.NewRequest("GET", tt.url, nil)
		if err != nil {
			t.Fatalf("unable to create request (http.NewRequest) -- %s", err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()

		// the handler under test
		handler := newServer(logger)

		// perform test
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != tt.status {
			t.Errorf("'%s' handler returned wrong status code: got %v want %v",
				tt.url, status, tt.status)
		}

		// Check the response body is what we expect.
		if tt.responseBody != nil && rr.Body.String() != *tt.responseBody {
			t.Errorf("'%s' handler returned unexpected body: got %v want %v",
				tt.url, rr.Body.String(), tt.responseBody)
		}
	}
}

var postTests = []struct {
	url          string
	status       int
	responseBody *string
}{
	{"/", http.StatusOK, nil},
}

func TestPostURLs(t *testing.T) {
	var expected string

	logger := getLogger()

	for _, tt := range postTests {
		if tt.responseBody != nil {
			expected = *tt.responseBody
		} else {
			expected = ""
		}

		// create a request to pass to the handler
		req, err := http.NewRequest("POST", tt.url, strings.NewReader(expected))
		if err != nil {
			t.Fatalf("unable to create request (http.NewRequest) -- %s", err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()

		// the handler under test
		handler := newServer(logger)

		// perform test
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != tt.status {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, tt.status)
		}

		// Check the response body is what we expect.
		if tt.responseBody != nil && rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	}
}

func getLogger() *zap.Logger {
	//config := zap.NewProductionConfig()
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build(zap.AddStacktrace(zapcore.PanicLevel))

	return logger //.With(log.String("x-request-id", "01234"))
}
