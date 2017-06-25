package authn

import (
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/afex/hystrix-go/hystrix/metric_collector"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	gsh "github.com/mchudgins/go-service-helper/handlers"
	"github.com/mchudgins/go-service-helper/hystrix"
	"github.com/mchudgins/playground/pkg/healthz"
	"github.com/prometheus/client_golang/prometheus"
)

type AuthResponse struct {
	JWT    string `json:"jwt"`
	UserID string `json:"userID"`
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
		mux := mux.NewRouter() //actuator.NewActuatorMux("")

		hc, err := healthz.NewConfig()
		healthzHandler, err := healthz.Handler(hc)
		if err != nil {
			log.Panic(err)
		}

		mux.Handle("/debug/vars", expvar.Handler())
		mux.Handle("/healthz", healthzHandler)
		mux.Handle("/metrics", prometheus.Handler())
		mux.HandleFunc("/login", loginGetHandler).Methods("GET")
		mux.HandleFunc("/login", loginPostHandler).Methods("POST")

		apiMux := http.NewServeMux()
		apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

			type data struct {
				Hostname string
				URL      string
				Handler  string
			}

			const authURL string = "/api/v1/authenticate/"

			logger, _ := gsh.FromContext(r.Context())

			if strings.HasPrefix(r.URL.Path, authURL) {
				uid := r.URL.Path[len(authURL):]
				now := time.Now()
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
					Subject:   uid,
					ExpiresAt: now.Add(time.Duration(5) * time.Minute).Unix(),
					Audience:  "*.dstcorp.net",
					Issuer:    "authn.dstcorp.net",
					IssuedAt:  now.Unix(),
					NotBefore: now.Add(0 - time.Duration(30)*time.Second).Unix(),
				}).SignedString([]byte("hello, world"))
				if err != nil {
					logger.WithError(err).Fatal("signing token")
				}

				m := &AuthResponse{
					JWT:    token,
					UserID: uid,
				}
				buf, err := json.Marshal(m)
				if err != nil {
					logger.WithError(err).WithField("authResponse", m.UserID).
						Error("while serializing auth response")
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(buf)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

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
		tracer = gsh.TracerFromHTTPRequest(gsh.NewTracer("authn"), "authn")
		chain := alice.New(tracer,
			gsh.HTTPMetricsCollector,
			gsh.HTTPLogrusLogger,
			handlers.CompressHandler,
			canonical).Then(mux)

		log.WithField("port", port).Info("HTTP service listening.")
		errc <- http.ListenAndServe(port, chain)
	}()

	// wait for somthin'
	log.Infof("exit: %s", <-errc)

	return nil
}
