package echo

import (
	"context"

	rpc "github.com/dstcorp/rpc-golang/service"
	"github.com/mchudgins/go-service-helper/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, logger *zap.Logger, port, certFile, keyFile string) {
	server.Run(ctx,
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
