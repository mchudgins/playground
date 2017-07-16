package echo

import (
	echo "github.com/dstcorp/rpc-golang/service"
	"go.uber.org/zap"
	context "golang.org/x/net/context"
)

type echoServer struct {
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) (*echoServer, error) {
	return &echoServer{
		logger: logger,
	}, nil
}

func (s *echoServer) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	resp := &echo.EchoResponse{
		Message: req.GetMessage(),
	}

	return resp, nil
}

func (s *echoServer) Diagnostics(ctx context.Context, req *echo.DiagnosticsRequest) (*echo.DiagnosticsResponse, error) {
	resp := &echo.DiagnosticsResponse{}

	return resp, nil
}
