package authn

import (
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"encoding/json"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/mchudgins/go-service-helper/actuator"
	gsh "github.com/mchudgins/go-service-helper/handlers"
	"github.com/mchudgins/go-service-helper/hystrix"
	"github.com/mchudgins/playground/pkg/healthz"
	"github.com/prometheus/client_golang/prometheus"
)

type promWriter struct {
	w             http.ResponseWriter
	statusCode    int
	contentLength int
}

var (
	httpRequestsReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "httpRequestsReceived_total",
			Help: "Number of HTTP requests received.",
		},
		[]string{"url"},
	)
	httpRequestsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "httpRequestsProcessed_total",
			Help: "Number of HTTP requests processed.",
		},
		[]string{"url", "status"},
	)
	httpRequestDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_duration",
			Help: "Duration of HTTP responses.",
		},
		[]string{"url", "status"},
	)
)

func NewPromWriter(w http.ResponseWriter) *promWriter {
	return &promWriter{w: w, statusCode: 200}
}

func (l *promWriter) Header() http.Header {
	return l.w.Header()
}

func (l *promWriter) Write(data []byte) (int, error) {
	l.contentLength += len(data)
	return l.w.Write(data)
}

func (l *promWriter) WriteHeader(status int) {
	l.statusCode = status
	l.w.WriteHeader(status)
}

func (l *promWriter) Length() int {
	return l.contentLength
}

func (l *promWriter) StatusCode() int {

	// if nobody set the status, but data has been written
	// then all must be well.
	if l.statusCode == 0 && l.contentLength > 0 {
		return http.StatusOK
	}

	return l.statusCode
}

func httpCounter(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		u := r.URL.Path
		httpRequestsReceived.With(prometheus.Labels{"url": u}).Inc()
		pw := NewPromWriter(w)
		defer func() {
			status := strconv.Itoa(pw.statusCode)
			httpRequestsProcessed.With(prometheus.Labels{"url": u, "status": status}).Inc()
			end := time.Now()
			duration := end.Sub(start)
			httpRequestDuration.With(prometheus.Labels{"url": u, "status": status}).Observe(float64(duration.Nanoseconds()))
		}()

		fn.ServeHTTP(pw, r)
	})
}

func Run(port, host string) error {
	log.Printf("authn.Run()")

	// make a channel to listen on events,
	// then launch the servers.

	errc := make(chan error)

	// interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// http server
	go func() {
		mux := actuator.NewActuatorMux("")

		hc, err := healthz.NewConfig()
		healthzHandler, err := healthz.Handler(hc)
		if err != nil {
			log.Panic(err)
		}

		mux.Handle("/debug/vars", expvar.Handler())
		mux.Handle("/healthz", healthzHandler)
		mux.Handle("/metrics", prometheus.Handler())

		apiMux := http.NewServeMux()
		apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			httpRequestsReceived.With(prometheus.Labels{"url": r.URL.Path}).Inc()

			type data struct {
				Hostname string
				URL      string
				Handler  string
			}

			type authResponse struct {
				JWT    string `json:"jwt"`
				UserID string `json:"userID"`
			}

			if strings.HasPrefix(r.URL.Path, "/api/v1/authenticate") {
				m := &authResponse{
					JWT:    "asldgk45cvmop8avppM",
					UserID: "bob@example.com",
				}
				buf, err := json.Marshal(m)
				if err != nil {
					log.WithError(err).WithField("authResponse", m.UserID).
						Error("while serializing auth response")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(buf)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

			httpRequestsProcessed.With(prometheus.Labels{"url": r.URL.Path, "status": "200"}).Inc()
		})
		circuitBreaker, err := hystrix.NewHystrixHelper("authn-api-backend")
		if err != nil {
			log.WithError(err).
				Fatalf("Error creating circuitBreaker")
		}
		metricCollector.Registry.Register(circuitBreaker.NewPrometheusCollector)
		mux.Handle("/api/v1/", circuitBreaker.Handler(apiMux))

		canonical := handlers.CanonicalHost(host, http.StatusPermanentRedirect)
		var tracer func(http.Handler) http.Handler
		tracer = gsh.TracerFromInternalHTTPRequest(gsh.NewTracer("authn"), "authn")
		chain := alice.New(tracer, gsh.HTTPLogrusLogger, httpCounter, canonical).Then(mux)

		log.WithField("port", port).Info("HTTP service listening.")
		errc <- http.ListenAndServe(port, chain)
	}()

	// wait for somthin'
	log.Infof("exit: %s", <-errc)

	return nil
}
