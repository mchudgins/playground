package backend

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/opentracing/opentracing-go"
)

const (
	idpEndpoint string = "http://localhost:9090/api/v1/authenticate"
)

func VerifyIdentity(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			httpClient := &http.Client{}
			httpReq, _ := http.NewRequest("GET", idpEndpoint, nil)

			httpReq.WithContext(ctx)

			// Transmit the span's TraceContext as HTTP headers on our
			// outbound request.
			opentracing.GlobalTracer().Inject(
				span.Context(),
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(httpReq.Header))

			resp, err := httpClient.Do(httpReq)
			if err != nil {
				log.WithError(err).WithField("idpEndpoint",idpEndpoint).
					Error("error contacting ipdEndpoint")
			} else if resp.StatusCode != 200 {
					log.WithField("StatusCode", resp.StatusCode).
						Warn("not authenticated")
			}
		} else {
			log.Error("no span")
		}
		defer span.Finish()

		fn.ServeHTTP(w,r)
	})
}
