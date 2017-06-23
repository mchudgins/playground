package backend

import (
	"context"
	"net/http"

	"strings"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/mchudgins/go-service-helper/user"
	"github.com/mchudgins/go-service-helper/zipkin"
	"github.com/mchudgins/playground/pkg/cmd/authn"
)

const (
	idpEndpoint   string = "http://localhost:9090/api/v1/authenticate"
	loginEndpoint string = "http://localhost:9090/login"

	authCookieName string = "Authentication"
	authHeaderName string = "Authentication"
)

func getTokenFromRequest(r *http.Request) string {
	// if the cookie is present
	cookie, err := r.Cookie(authCookieName)
	if cookie != nil && err == nil {
		return cookie.Value
	}

	hdr := r.Header.Get(authHeaderName)
	if len(hdr) > 0 {
		str := strings.Split(hdr, " ")
		if len(str) == 2 && strings.Compare("token", strings.ToLower(str[0])) == 0 {
			return str[1]
		}
	}

	return ""
}

func validateWithIDP(ctx context.Context, token string) string {
	var resp *http.Response
	var err error

	httpClient := zipkin.NewClient("authn")

	httpReq, _ := http.NewRequest("GET", idpEndpoint+"/"+token, nil)

	resp, err = httpClient.Do(httpReq.WithContext(ctx))
	if err != nil {
		log.WithError(err).WithField("idpEndpoint", idpEndpoint).
			Error("error contacting ipdEndpoint")
	} else if resp.StatusCode != 200 {
		log.WithField("StatusCode", resp.StatusCode).
			Warn("not authenticated")
	}

	decoder := json.NewDecoder(resp.Body)
	var authResponse authn.AuthResponse
	err = decoder.Decode(&authResponse)
	if err != nil {
		log.WithError(err).Fatal("decoding authn response")
		return ""
	}

	log.WithFields(log.Fields{"userID": authResponse.UserID, "jwt": authResponse.JWT}).Info("auth response")

	return authResponse.UserID
}

func VerifyIdentity(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		verified := false
		ctx := r.Context()
		token = getTokenFromRequest(r)
		log.WithField("token", token).Info("VerifyIdentity")
		if len(token) != 0 {
			uid := validateWithIDP(ctx, token)
			if len(uid) > 0 {
				r = r.WithContext(user.NewContext(ctx, uid))
				r.Header.Set(user.USERID, uid)
			}
			verified = len(uid) > 0
		}

		if !verified {
			w.Header().Set("Location", loginEndpoint)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		fn.ServeHTTP(w, r)
	})
}
