package testServer

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"io/ioutil"

	"fmt"

	"github.com/mchudgins/go-service-helper/server"
	"go.uber.org/zap"
)

type TestServer struct {
	httputil.ReverseProxy
	address     string
	commandName string
	logger      *zap.Logger
	defaultCSP  string
	target      url.URL
	insecure    bool
}

func New(target *url.URL, defaultCSP string, logger *zap.Logger, listenPort string, fInsecure bool) (*TestServer, error) {

	director := func(req *http.Request) {
		req.Host = req.Header.Get("Host")
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	}

	ts := &TestServer{
		address:      listenPort,
		target:       *target,
		commandName:  target.Host,
		defaultCSP:   defaultCSP,
		ReverseProxy: httputil.ReverseProxy{Director: director},
		logger:       logger,
		insecure:     fInsecure,
	}

	ts.ReverseProxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: fInsecure},
	}

	return ts, nil
}

func (p *TestServer) Run(ctx context.Context, certFile, keyFile string) {

	index := 0
	if p.address[0] == ':' {
		index = index + 1
	}
	listenPort, err := strconv.Atoi(p.address[index:])
	if err != nil {
		return
	}

	server.Run(ctx,
		server.WithLogger(p.logger),
		server.WithHTTPListenPort(listenPort),
		server.WithCertificate(certFile, keyFile),
		server.WithHTTPServer(p.NewHTTPServer()),
	)
}

func (p *TestServer) NewHTTPServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p.logger.Info(r.URL.String(),
			zap.Any("Headers", r.Header))

		/*
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
		*/

		// the whole reason we're proxy'ing, is to test the app with various
		// security-related headers

		if len(p.defaultCSP) > 0 {
			w.Header().Set("Content-Security-Policy", p.defaultCSP)
		}

		if len(r.Header["X-Bounce"]) > 0 {
			w.WriteHeader(http.StatusOK)
			return
		}

		origin := r.Header["Origin"]

		if len(origin) > 0 &&
			r.Method == "OPTIONS" {

			originURL, err := url.Parse(origin[0])
			if err != nil {
				p.logger.Info("unable to parse ORIGIN header as URL", zap.Error(err),
					zap.String("Origin", origin[0]))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if originURL.Hostname() != "localhost" &&
				strings.HasSuffix(originURL.Hostname(), ".dstcorp.io") == false {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			acrHeaders := r.Header["Access-Control-Request-Headers"]
			acrMethod := r.Header["Access-Control-Request-Method"]
			if len(acrHeaders) == 0 || len(acrMethod) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if acrMethod[0] != "POST" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.Header().Set("Access-Control-Allow-Origin", originURL.String())
		}

		if len(origin) > 0 &&
			(r.Method == "GET" || r.Method == "HEAD") {

			originURL, err := url.Parse(origin[0])
			if err != nil {
				p.logger.Info("unable to parse ORIGIN header as URL", zap.Error(err),
					zap.String("Origin", origin[0]))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			p.logger.Info("Origin",
				zap.String("host", originURL.Host),
				zap.String("Hostname()", originURL.Hostname()),
				zap.String("port", originURL.Port()))
			if len(r.Header["Referer"]) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if originURL.Hostname() == "localhost" ||
				strings.HasSuffix(originURL.Hostname(), ".dstcorp.io") == true {
				w.Header().Set("Access-Control-Allow-Origin", originURL.String())
			}
		}

		if len(origin) > 0 &&
			r.Method == "POST" {
			originURL, err := url.Parse(origin[0])
			if err != nil {
				p.logger.Info("unable to parse ORIGIN header as URL", zap.Error(err),
					zap.String("Origin", origin[0]))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			p.logger.Info("Origin",
				zap.String("host", originURL.Host),
				zap.String("Hostname()", originURL.Hostname()),
				zap.String("port", originURL.Port()))
			/*
				if len(r.Header["Referer"]) == 0 {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			*/
			if originURL.Hostname() == "localhost" ||
				strings.HasSuffix(originURL.Hostname(), ".dstcorp.io") == true {
				w.Header().Set("Access-Control-Allow-Origin", originURL.String())
			}

			data, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				p.logger.Error("unable to read body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			p.logger.Info("data", zap.String("Body", string(data)))
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "{ \"status\" : \"ok\", \"body\": %s }", string(data))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{ "status": "ok" }`))

		//		p.ReverseProxy.ServeHTTP(w, r)
	})

	return mux
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
