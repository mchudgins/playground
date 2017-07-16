package echo

import (
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	afex "github.com/afex/hystrix-go/hystrix"
	rpc "github.com/dstcorp/rpc-golang/service"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/justinas/alice"
	"github.com/mchudgins/go-service-helper/correlationID"
	gsh "github.com/mchudgins/go-service-helper/handlers"
	"github.com/mchudgins/playground/pkg/healthz"
	"github.com/mwitkow/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
	xcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type echoService struct {
	Insecure    bool
	address     string
	gRPCAddress string
	commandName string
	logger      *zap.Logger
}

func grpcEndpointLog(logger *zap.Logger, s string) grpc.UnaryServerInterceptor {
	return func(ctx xcontext.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		//			log.Debug("grpcEndpointLog+", zap.String("", s))
		//			defer log.Debug("grpcEndpointLog-", zap.String("", s))
		return handler(ctx, req)
	}
}

func (p *echoService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
}

func Run(logger *zap.Logger, port, certFile, keyFile string) {
	p := &echoService{
		Insecure:    false,
		logger:      logger,
		address:     ":6060",
		gRPCAddress: port,
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

	// gRPC server
	go func() {

		lis, err := net.Listen("tcp", p.gRPCAddress)
		if err != nil {
			errc <- err
			return
		}

		var s *grpc.Server

		if p.Insecure {
			s = grpc.NewServer(
				grpc_middleware.WithUnaryServerChain(
					grpc_prometheus.UnaryServerInterceptor,
					grpcEndpointLog(p.logger, "certMgr")))
		} else {
			tlsCreds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
			if err != nil {
				p.logger.Fatal("Failed to generate grpc TLS credentials", zap.Error(err))
			}
			s = grpc.NewServer(
				grpc.Creds(tlsCreds),
				grpc.RPCCompressor(grpc.NewGZIPCompressor()),
				grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
				grpc_middleware.WithUnaryServerChain(
					grpc_prometheus.UnaryServerInterceptor,
					grpcEndpointLog(p.logger, "Echo RPC server")))
		}

		echoServer, err := NewServer(p.logger)
		if err != nil {
			p.logger.Panic("while creating new EchoServer", zap.Error(err))
		}
		rpc.RegisterEchoServiceServer(s, echoServer)

		if p.Insecure {
			log.Warnf("gRPC service listening insecurely on %s", p.gRPCAddress)
		} else {
			log.Infof("gRPC service listening on %s", p.gRPCAddress)
		}
		errc <- s.Serve(lis)
	}()

	// health & metrics via https
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

		canonical := handlers.CanonicalHost("http://fubar.local.dstcorp.io:7070", http.StatusPermanentRedirect)
		var tracer func(http.Handler) http.Handler
		tracer = gsh.TracerFromHTTPRequest(gsh.NewTracer(p.commandName), "proxy")

		rootMux.PathPrefix("/").Handler(p)

		chain := alice.New(tracer,
			gsh.HTTPMetricsCollector,
			gsh.HTTPLogrusLogger,
			canonical,
			handlers.CompressHandler)

		//errc <- http.ListenAndServe(p.address, chain.Then(rootMux))
		tls := &http.Server{
			Addr:              p.address,
			Handler:           chain.Then(rootMux),
			ReadTimeout:       time.Duration(5) * time.Second,
			ReadHeaderTimeout: time.Duration(2) * time.Second,
		}

		errc <- tls.ListenAndServeTLS(certFile, keyFile)
	}()

	// start the hystrix stream provider
	go func() {
		hystrixStreamHandler := afex.NewStreamHandler()
		hystrixStreamHandler.Start()
		errc <- http.ListenAndServe(":7071", hystrixStreamHandler)
	}()

	// wait for somthin'
	p.logger.Info("Echo Server",
		zap.String("local address", p.address),
		zap.String("gRPC address", p.gRPCAddress))
	p.logger.Info("exit", zap.Error(<-errc))
}
