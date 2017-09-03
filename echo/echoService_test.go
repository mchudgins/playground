package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getLogger() *zap.Logger {
	//config := zap.NewProductionConfig()
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build(zap.AddStacktrace(zapcore.PanicLevel))

	return logger //.With(log.String("x-request-id", "01234"))
}

var getTests = []struct {
	url          string
	status       int
	responseBody *string
}{
	{"/", http.StatusOK, nil},
	{"/test", http.StatusOK, nil},
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
		handler := NewHTTPServer(logger)

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

const expected = "this is what we expect to be echo-ed"

var methodTests = []struct {
	url          string
	method       string
	responseBody string
	statusCode   int
}{
	{"/", "POST", expected, http.StatusOK},
	{"/", "GET", "", http.StatusOK},
	{"/", "PUT", "", http.StatusMethodNotAllowed},
	{"/", "HEAD", "", http.StatusMethodNotAllowed},
	{"/", "DELETE", "", http.StatusMethodNotAllowed},
	{"/", "OPTIONS", "", http.StatusMethodNotAllowed},
}

func TestEchoPost(t *testing.T) {

	for _, tt := range methodTests {
		// create a request to pass to the handler
		var reader io.Reader
		if len(tt.responseBody) != 0 {
			reader = strings.NewReader(tt.responseBody)
		} else {
			reader = nil
		}
		req, err := http.NewRequest(tt.method, tt.url, reader)
		if err != nil {
			t.Fatalf("unable to create request (http.NewRequest) -- %s", err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()

		// the handler under test
		handler := NewHTTPServer(getLogger())

		// perform test
		handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != tt.statusCode {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, tt.statusCode)
		}

		// Check the response body is what we expect.
		if len(tt.responseBody) != 0 && rr.Body.String() != tt.responseBody {
			t.Errorf("%s handler returned unexpected body: got '%v' want '%v'",
				tt.method, rr.Body.String(), tt.responseBody)
		}
	}
}
