package backend

import (
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/template"
	"time"

	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/mchudgins/certMgr/pkg/healthz"
	"github.com/mchudgins/go-service-helper/actuator"
	"github.com/mchudgins/go-service-helper/hystrix"
	"github.com/mchudgins/go-service-helper/loggingWriter"
	"github.com/mchudgins/go-service-helper/serveSwagger"
	"github.com/prometheus/client_golang/prometheus"
)

type promWriter struct {
	w             http.ResponseWriter
	statusCode    int
	contentLength int
}

var (
	indexTemplate *template.Template
	html          = `
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
  <title>Welcome to OpenShift</title>
  <p>This is {{.Hostname}}</p>
  <p>Page: {{.URL}}</p>
  <p>Handler: {{.Handler}}</p>
</body>
</html>`

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

func init() {
	prometheus.MustRegister(httpRequestsReceived)
	prometheus.MustRegister(httpRequestsProcessed)
	prometheus.MustRegister(httpRequestDuration)

	indexTemplate = template.Must(template.New("/").Parse(html))
}

func NewPromWriter(w http.ResponseWriter) *promWriter {
	return &promWriter{w: w}
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

func httpCounter(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func Run() error {
	log.Printf("backend.Run()")

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

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

		swaggerProxy, _ := serveSwagger.NewSwaggerProxy("/swagger-ui/")
		mux.Handle("/swagger-ui/", swaggerProxy)

		mux.Handle("/swagger/",
			http.StripPrefix("/swagger/", Server))

		apiMux := http.NewServeMux()
		apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			httpRequestsReceived.With(prometheus.Labels{"url": r.URL.Path}).Inc()

			type data struct {
				Hostname string
				URL      string
				Handler  string
			}

			type echo struct {
				Message string `json:"message"`
			}

			if strings.HasPrefix(r.URL.Path, "/api/v1/echo/") {
				m := &echo{
					Message: "hello, " + r.URL.Path[len("/api/v1/echo/"):],
				}
				buf, err := json.Marshal(m)
				if err != nil {
					log.WithError(err).WithField("message", m.Message).
						Error("while serializing echo response")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(buf)
				}
			} else {
				err = indexTemplate.Execute(w, data{Hostname: hostname, URL: r.URL.Path, Handler: "/api/v1"})
				if err != nil {
					log.WithError(err).
						WithField("template", indexTemplate.Name()).
						WithField("path", r.URL.Path).
						Error("Unable to execute template")
				}
			}

			httpRequestsProcessed.With(prometheus.Labels{"url": r.URL.Path, "status": "200"}).Inc()
		})
		circuitBreaker, err := hystrix.NewHystrixHelper("grpc-backend")
		if err != nil {
			log.WithError(err).
				Fatalf("Error creating circuitBreaker")
		}
		metricCollector.Registry.Register(circuitBreaker.NewPrometheusCollector)
		mux.Handle("/api/v1/", circuitBreaker.Handler(apiMux))

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			httpRequestsReceived.With(prometheus.Labels{"url": "/"}).Inc()

			status := http.StatusOK

			type data struct {
				Hostname string
				URL      string
				Handler  string
			}

			switch r.URL.Path {
			case "/":
				err = indexTemplate.Execute(w, data{Hostname: hostname, URL: r.URL.Path, Handler: "/"})
				if err != nil {
					log.WithError(err).
						WithField("template", indexTemplate.Name()).
						WithField("path", r.URL.Path).
						Error("Unable to execute template")
					status = http.StatusServiceUnavailable
				}
				break

			default:
				status = http.StatusNotFound
				http.NotFound(w, r)
			}

			httpRequestsProcessed.With(prometheus.Labels{"url": "/", "status": strconv.Itoa(status)}).Inc()
		})

		log.WithField("port", ":8080").Info("HTTPS service listening.")
		errc <- http.ListenAndServe(":8080", loggingWriter.HTTPLogrusLogger(httpCounter(mux)))
	}()

	// wait for somthin'
	log.Infof("exit: %s", <-errc)

	return nil
}
