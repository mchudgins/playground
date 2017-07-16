package echo

import (
	rpc "github.com/dstcorp/rpc-golang/service"
	"github.com/mchudgins/go-service-helper/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

/*
type sourcetype int

const (
	interrupt sourcetype = iota
	httpServer
	metricsServer
	rpcServer
)

type errorSource struct {
	source sourcetype
	err    error
}

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
		logger.Info("grpcEndpointLog+", zap.String("", s))
		defer func() {
			logger.Info("grpcEndpointLog-", zap.String("", s))
			logger.Sync()
		}()

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

	errc := make(chan errorSource)

	// interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- errorSource{
			source: interrupt,
			err:    fmt.Errorf("%s", <-c),
		}
	}()

	// gRPC server
	go func() {

		lis, err := net.Listen("tcp", p.gRPCAddress)
		if err != nil {
			errc <- errorSource{
				err:    err,
				source: rpcServer,
			}
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
		errc <- errorSource{
			err:    s.Serve(lis),
			source: rpcServer,
		}
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

		errc <- errorSource{
			err:    tls.ListenAndServeTLS(certFile, keyFile),
			source: httpServer,
		}
	}()

	// start the hystrix stream provider
	go func() {
		hystrixStreamHandler := afex.NewStreamHandler()
		hystrixStreamHandler.Start()
		errc <- errorSource{
			err:    http.ListenAndServe(":7071", hystrixStreamHandler),
			source: metricsServer,
		}
	}()

	// wait for somthin'
	p.logger.Info("Echo Server",
		zap.String("local address", p.address),
		zap.String("gRPC address", p.gRPCAddress))
	rc := <-errc
	p.logger.Info("exit", zap.Error(rc.err), zap.Int("source", int(rc.source)))
	if rc.source == interrupt {
	}
}
*/

func Run(logger *zap.Logger, port, certFile, keyFile string) {
	server.Run(
		server.WithLogger(logger),
		server.WithCertificate(certFile, keyFile),
		server.WithRPCServer(func(s *grpc.Server) error {
			echoServer, err := NewServer(logger)
			if err != nil {
				logger.Panic("while creating new EchoServer", zap.Error(err))
			}
			rpc.RegisterEchoServiceServer(s, echoServer)
			return nil
		}),
	)
}
