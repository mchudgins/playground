package backend

//go:generate go run ../../../main.go htmlGen ../../../cmd/htmlGen/test.yaml

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	gsh "github.com/mchudgins/go-service-helper/handlers"
	"github.com/mchudgins/go-service-helper/hystrix"
	"github.com/mchudgins/go-service-helper/server"
	"github.com/mchudgins/playground/pkg/cmd/backend/htmlGen"
	"github.com/mchudgins/playground/tmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func newServer(logger *zap.Logger) http.Handler {
	hostname, err := os.Hostname()
	if err != nil {
		logger.Panic("unable to obtain hostname", zap.Error(err))
	}

	mux := mux.NewRouter()
	/*
		swaggerProxy, _ := serveSwagger.NewSwaggerProxy("/swagger-ui/")
		mux.Handle("/swagger-ui/", swaggerProxy)

		mux.Handle("/swagger/",
			http.StripPrefix("/swagger/", Server))
	*/
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		logger, ok := gsh.FromContext(r.Context())
		if ok {
			logger.WithField("url", r.URL.Path).Info("api called")
		}

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

	mux.PathPrefix("/api/v1/").Handler(alice.New(circuitBreaker.Handler, VerifyIdentity).Then(apiMux))

	mux.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("@switch", zap.String("URL.Path", r.URL.Path))

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

	return mux
}

func Run(ctx context.Context, port, host string) error {
	logger := GetLogger()
	defer logger.Sync()

	mux := newServer(logger)

	server.Run(ctx,
		server.WithLogger(logger),
		server.WithCertificate("../certMgr/cert.pem", "../certMgr/key.pem"),
		server.WithHTTPServer(mux))

	return nil
}

func GetLogger() *zap.Logger {
	//config := zap.NewProductionConfig()
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build(zap.AddStacktrace(zapcore.PanicLevel))

	return logger //.With(log.String("x-request-id", "01234"))
}
