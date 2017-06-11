package backend

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

const (
	CORRID string = "X-Correlation-ID"
)

var (
	debugMode          = true
	serviceName        = "playground"
	serviceHostPort    = "localhost:8080"
	zipkinHTTPEndpoint = "http://localhost:9411/api/v1/spans"
)

type traceLogger struct{}

func (logger traceLogger) Log(keyval ...interface{}) error {
	log.Warn("Log was called!")
	return nil
}

func NewTracer() opentracing.Tracer {
	collector, err := zipkin.NewHTTPCollector(zipkinHTTPEndpoint,
		zipkin.HTTPLogger(traceLogger{}),
		zipkin.HTTPBatchSize(1))
	if err != nil {
		log.WithError(err).Fatal("zipkin.NewHTTPCollector failed")
	}

	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, debugMode, serviceHostPort, serviceName),
		zipkin.WithLogger(traceLogger{}),
		zipkin.DebugMode(true),
		zipkin.ClientServerSameSpan(true),
	)

	if err != nil {
		log.WithError(err).Warn("unable to construct zipkin.Tracer")
	}

	opentracing.InitGlobalTracer(tracer)

	return tracer
}

// HandlerFunc is a middleware function for incoming HTTP requests.
type HandlerFunc func(next http.Handler) http.Handler

// FromHTTPRequest returns a Middleware HandlerFunc that tries to join with an
// OpenTracing trace found in the HTTP request headers and starts a new Span
// called `operationName`. If no trace could be found in the HTTP request
// headers, the Span will be a trace root. The Span is incorporated in the
// HTTP Context object and can be retrieved with
// opentracing.SpanFromContext(ctx).
func TracerFromHTTPRequest(tracer opentracing.Tracer, operationName string,
) HandlerFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			span, ctx := opentracing.StartSpanFromContext(req.Context(), operationName)
			defer span.Finish()

			// tag this request with a correlation ID, so we can troubleshoot it later, if necessary
			corrID := req.Header.Get(CORRID)
			if len(corrID) == 0 {
				corrID = uuid.New().String()
			}

			w.Header().Add(CORRID, corrID)
			span.SetTag(CORRID, corrID)

			// store span in context
			ctx = opentracing.ContextWithSpan(req.Context(), span)

			// update request context to include our new span
			req = req.WithContext(ctx)

			// next middleware or actual request handler
			next.ServeHTTP(w, req)
		})
	}
}
