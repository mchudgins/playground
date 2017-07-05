package reverseProxy

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	afex "github.com/afex/hystrix-go/hystrix"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/mchudgins/go-service-helper/correlationID"
	gsh "github.com/mchudgins/go-service-helper/handlers"
	"github.com/mchudgins/playground/pkg/healthz"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Proxy struct {
	httputil.ReverseProxy
	address     string
	commandName string
	logger      *zap.Logger
	defaultCSP  string
}

func NewProxy(target *url.URL, defaultCSP string, logger *zap.Logger) (*Proxy, error) {

	director := func(req *http.Request) {
		req.Host = req.Header.Get("Host")
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)

		logger.Info("new path", zap.String("path", req.URL.Path))
	}

	return &Proxy{
		address:      ":7070",
		commandName:  target.Host,
		defaultCSP:   defaultCSP,
		ReverseProxy: httputil.ReverseProxy{Director: director},
		logger:       logger}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// send any correlation ID on to the servers we contact
	corrID := correlationID.FromContext(ctx)
	if len(corrID) > 0 {
		r.Header.Set(correlationID.CORRID, corrID)
	}

	// enable tracing
	var childSpan opentracing.Span
	span := opentracing.SpanFromContext(ctx)

	if span == nil {
		childSpan = opentracing.StartSpan(p.commandName)
	} else {
		childSpan = opentracing.StartSpan(p.commandName,
			opentracing.ChildOf(span.Context()))

	}
	defer childSpan.Finish()

	ext.SpanKindRPCClient.Set(childSpan)

	//r = r.WithContext(ctx)

	// Transmit the span's TraceContext as HTTP headers on our
	// outbound request.
	opentracing.GlobalTracer().Inject(
		childSpan.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	// the whole reason we're proxy'ing, is to test the app with various
	// security-related headers

	if len(p.defaultCSP) > 0 {
		w.Header().Set("Content-Security-Policy", p.defaultCSP)
	}

	p.ReverseProxy.ServeHTTP(w, r)
}

func (p *Proxy) Run() error {
	// make a channel to listen on events,
	// then launch the servers.

	errc := make(chan error)

	// interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// proxy
	go func() {
		rootMux := mux.NewRouter() //actuator.NewActuatorMux("")

		hc, err := healthz.NewConfig()
		healthzHandler, err := healthz.Handler(hc)
		if err != nil {
			p.logger.Panic("Constructing healthz.Handler", zap.Error(err))
		}

		// set up handlers for THIS instance
		// (these are not expected to be proxied)
		rootMux.Handle("/debug/vars", expvar.Handler())
		rootMux.Handle("/healthz", healthzHandler)
		rootMux.Handle("/metrics", prometheus.Handler())
		rootMux.Handle("/cspReport", NewCSPReporter()).Methods("POST")

		canonical := handlers.CanonicalHost("localhost", http.StatusPermanentRedirect)
		var tracer func(http.Handler) http.Handler
		tracer = gsh.TracerFromHTTPRequest(gsh.NewTracer(p.commandName), "proxy")

		rootMux.PathPrefix("/").Handler(p)

		chain := alice.New(tracer,
			gsh.HTTPMetricsCollector,
			gsh.HTTPLogrusLogger,
			canonical,
			handlers.CompressHandler)

		errc <- http.ListenAndServe(p.address, chain.Then(rootMux))
	}()

	// start the hystrix stream provider
	go func() {
		hystrixStreamHandler := afex.NewStreamHandler()
		hystrixStreamHandler.Start()
		errc <- http.ListenAndServe(":7071", hystrixStreamHandler)
	}()

	// wait for somthin'
	p.logger.Info("Reverse Proxy Serving", zap.String("address", p.address))
	p.logger.Info("exit", zap.Error(<-errc))

	return nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
