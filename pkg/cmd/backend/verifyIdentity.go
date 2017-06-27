package backend

import (
	"context"
	"net/http"

	"strings"

	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	gsh "github.com/mchudgins/go-service-helper/handlers"
	"github.com/mchudgins/go-service-helper/user"
	"github.com/mchudgins/go-service-helper/zipkin"
	"github.com/mchudgins/playground/pkg/cmd/authn"
)

const (
	idpEndpoint   string = "http://localhost:9090/api/v1/authenticate"
	loginEndpoint string = "http://localhost:9090/login"

	authCookieName string = "Authentication"
	authHeaderName string = "Authorization"
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

	logger, _ := gsh.FromContext(ctx)

	httpClient := zipkin.NewClient("authn")

	httpReq, _ := http.NewRequest("GET", idpEndpoint+"/"+token, nil)

	resp, err = httpClient.Do(httpReq.WithContext(ctx))
	if err != nil {
		logger.WithError(err).WithField("idpEndpoint", idpEndpoint).
			Error("error contacting ipdEndpoint")
		return ""
	} else if resp.StatusCode != 200 {
		logger.WithField("StatusCode", resp.StatusCode).
			Warn("not authenticated")
		return ""
	}

	decoder := json.NewDecoder(resp.Body)
	var authResponse authn.AuthResponse
	err = decoder.Decode(&authResponse)
	if err != nil {
		logger.WithError(err).Fatal("decoding authn response")
		return ""
	}

	jwtToken, err := jwt.ParseWithClaims(authResponse.JWT, &jwt.StandardClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("hello, world"), nil
	})
	if err != nil {
		logger.WithError(err).Warn("invalid JWT")
		return ""
	}

	claims := jwtToken.Claims.(*jwt.StandardClaims)

	/*
		w, err := jws.ParseJWT([]byte(authResponse.JWT))
		if err != nil {
			log.WithError(err).Warn("unable to parse token")
			return ""
		}

		x, ok := w.(*jws.JWS)


		v := &jwt.Validator{}
		v.SetIssuer("authn.dstcorp.net")
		v.SetAudience("*.dstcorp.net")
		err = w.Validate([]byte("hello, world"), jws.GetSigningMethod("HS256"), v)
		if err != nil {
			log.WithError(err).Warn("invalid JWT Token")
			return ""
		}
		claims := w.Claims()
	*/

	subject := claims.Subject
	if len(subject) == 0 {
		logger.Warn("no subject found in JWT claims")
		return ""
	}

	alg := jwtToken.Header["alg"].(string)
	if strings.Compare(alg, jwt.SigningMethodHS256.Name) != 0 {
		logger.WithFields(log.Fields{"alg": jwtToken.Header["alg"],
			"name": jwt.SigningMethodHS256.Name,
		}).Info()
		return ""
	}

	logger.WithFields(log.Fields{"userID": authResponse.UserID,
		"jwt":         authResponse.JWT,
		"alg":         jwtToken.Header["alg"],
		"kid":         jwtToken.Header["kid"],
		"jwt.Subject": subject}).Info("auth response")

	return authResponse.UserID
}

func VerifyIdentity(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		verified := false
		ctx := r.Context()
		logger, _ := gsh.FromContext(ctx)
		token = getTokenFromRequest(r)
		logger.WithField("token", token).Info("VerifyIdentity")
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
