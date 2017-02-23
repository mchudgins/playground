package backend

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/mchudgins/certMgr/pkg/healthz"
	"github.com/mchudgins/go-service-helper/pkg/loggingWriter"
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
)

func init() {
	prometheus.MustRegister(httpRequestsReceived)
	prometheus.MustRegister(httpRequestsProcessed)

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
		u := r.URL.Path
		httpRequestsReceived.With(prometheus.Labels{"url": u}).Inc()
		pw := NewPromWriter(w)
		defer func() {
			httpRequestsProcessed.With(prometheus.Labels{"url": u, "status": strconv.Itoa(pw.statusCode)}).Inc()
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
		hc, err := healthz.NewConfig()
		healthzHandler, err := healthz.Handler(hc)
		if err != nil {
			log.Panic(err)
		}

		mux := http.NewServeMux()

		mux.Handle("/healthz", healthzHandler)
		mux.Handle("/metrics", prometheus.Handler())
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			httpRequestsReceived.With(prometheus.Labels{"url": "/"}).Inc()

			type data struct {
				Hostname string
			}

			err = indexTemplate.Execute(w, data{Hostname: hostname})
			if err != nil {
				log.WithError(err).Error("Unable to execute template")
			}

			httpRequestsProcessed.With(prometheus.Labels{"url": "/", "status": "200"}).Inc()
		})

		log.Infof("HTTPS service listening on %s", ":8080")
		errc <- http.ListenAndServe(":8080", loggingWriter.HTTPLogrusLogger(httpCounter(mux)))
	}()

	// wait for somthin'
	log.Infof("exit: %s", <-errc)

	return nil
}
