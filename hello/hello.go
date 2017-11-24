package hello

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/mchudgins/go-service-helper/server"
	"go.uber.org/zap"
)

func Run(ctx context.Context, logger *zap.Logger, port, certFile, keyFile string) {

	index := 0
	if port[0] == ':' {
		index = index + 1
	}
	listenPort, err := strconv.Atoi(port[index:])
	if err != nil {
		return
	}

	server.Run(ctx,
		server.WithLogger(logger),
		server.WithHTTPListenPort(listenPort),
		server.WithHTTPServer(NewHTTPServer(logger)),
	)
}

func NewHTTPServer(logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		logger.Info("@handler",
			zap.String("URL", req.URL.Path),
			zap.String("Method", req.Method))

		hostname, err := os.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("%s", err)
		}

		switch req.Method {
		case "GET":
			w.Header().Add("X-Host", hostname)
			w.WriteHeader(http.StatusOK)
			break

		case "POST":
			w.Header().Add("X-Host", hostname)
			buf, err := ioutil.ReadAll(req.Body)
			if err != nil {
				logger.Error("failed to read POST data", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Write(buf)
			break

		default:
			w.Header().Add("X-Host", hostname)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return mux
}
