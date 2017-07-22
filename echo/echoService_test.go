package echo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

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

func TestEchoPost(t *testing.T) {

	const expected = "this is what we expect to be echo-ed"

	// create a request to pass to the handler
	req, err := http.NewRequest("POST", "/", strings.NewReader(expected))
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
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
