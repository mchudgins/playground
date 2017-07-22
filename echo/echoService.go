package echo

import (
	"io/ioutil"
	"net/http"

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

func NewHTTPServer(logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		logger.Info("@handler",
			zap.String("URL", req.URL.Path),
			zap.String("Method", req.Method))

		switch req.Method {
		case "GET":
			w.WriteHeader(http.StatusOK)
			break

		case "POST":
			buf, err := ioutil.ReadAll(req.Body)
			if err != nil {
				logger.Error("failed to read POST data", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Write(buf)
			break

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return mux
}
