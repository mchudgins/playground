package backend

//go:generate go run ../../../main.go htmlGen ../../../cmd/htmlGen/test.yaml

import (
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"
	afex "github.com/afex/hystrix-go/hystrix"
	"github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/mchudgins/go-service-helper/actuator"
	gsh "github.com/mchudgins/go-service-helper/handlers"
	"github.com/mchudgins/go-service-helper/hystrix"
	"github.com/mchudgins/go-service-helper/serveSwagger"
	"github.com/mchudgins/playground/pkg/cmd/backend/htmlGen"
	"github.com/mchudgins/playground/pkg/healthz"
	"github.com/mchudgins/playground/tmp"
	"github.com/prometheus/client_golang/prometheus"
)

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
)

func init() {
	indexTemplate = template.Must(template.New("/").Parse(html))
}

func Run(port, host string) error {
	log.Printf("backend.Run()")

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	if len(host) == 0 {
		host = hostname
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

		})
		circuitBreaker, err := hystrix.NewHystrixHelper("grpc-backend")
		if err != nil {
			log.WithError(err).
				Fatalf("Error creating circuitBreaker")
		}
		metricCollector.Registry.Register(circuitBreaker.NewPrometheusCollector)
		mux.Handle("/api/v1/", circuitBreaker.Handler(apiMux))

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

			type data struct {
				Hostname string
				URL      string
				Handler  string
			}

			switch r.URL.Path {
			case "/apis-explorer":
				r.URL.Path = "/apiList.html"
				htmlGen.Server.ServeHTTP(w, r)
				break

			case "/test":
				err = indexTemplate.Execute(w, data{Hostname: hostname, URL: r.URL.Path, Handler: "/"})
				if err != nil {
					log.WithError(err).
						WithField("template", indexTemplate.Name()).
						WithField("path", r.URL.Path).
						Error("Unable to execute template")
				}
				break

			default:
				if r.URL.Path == "/" {
					r.URL.Path = "/index.html"
				}

				tmp.ServeHTTPWithIndexes(w, r)
				//				status = http.StatusNotFound
				//				http.NotFound(w, r)
			}

		})

		canonical := handlers.CanonicalHost(host, http.StatusPermanentRedirect)
		var tracer func(http.Handler) http.Handler
		tracer = gsh.TracerFromHTTPRequest(gsh.NewTracer("playground"), "playground")
		chain := alice.New(tracer, gsh.HTTPMetricsCollector, gsh.HTTPLogrusLogger, canonical, VerifyIdentity).Then(mux)

		log.WithField("port", port).Info("HTTP service listening.")
		errc <- http.ListenAndServe(port, chain)
	}()

	// start the hystrix stream provider
	go func() {
		hystrixStreamHandler := afex.NewStreamHandler()
		hystrixStreamHandler.Start()
		errc <- http.ListenAndServe(":8081", hystrixStreamHandler)
	}()

	// wait for somthin'
	log.Infof("exit: %s", <-errc)

	return nil
}
