package reverseProxy

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

	"github.com/mchudgins/go-service-helper/server"
	"go.uber.org/zap"
)

type Proxy struct {
	httputil.ReverseProxy
	address     string
	commandName string
	logger      *zap.Logger
	defaultCSP  string
	target      url.URL
	insecure    bool
}

func NewProxy(target *url.URL, defaultCSP string, logger *zap.Logger, listenPort string, fInsecure bool) (*Proxy, error) {

	director := func(req *http.Request) {
		req.Host = req.Header.Get("Host")
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)

		logger.Info("new path", zap.String("path", req.URL.Path))
	}

	proxy := &Proxy{
		address:      listenPort,
		target:       *target,
		commandName:  target.Host,
		defaultCSP:   defaultCSP,
		ReverseProxy: httputil.ReverseProxy{Director: director},
		logger:       logger,
		insecure:     fInsecure,
	}

	proxy.ReverseProxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: fInsecure},
	}

	return proxy, nil
}

func (p *Proxy) Run(ctx context.Context, certFile, keyFile string) {

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

func (p *Proxy) NewHTTPServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		p.ReverseProxy.ServeHTTP(w, r)
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
