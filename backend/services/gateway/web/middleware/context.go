package middleware

import (
	"context"
	"net/http"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
)

func BuildHandleContextMiddleware(
	clientSecret string,
	jwtManager crypto.JwtManager,
	logger log.Logger,
) func(next http.Handler) http.HandlerFunc {
	logger.Debugf("pipedrive context middleware has been built with client_secret %s", clientSecret)
	return func(next http.Handler) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			var ctx request.PipedriveTokenContext
			token := r.Header.Get("X-Pipedrive-App-Context")
			if token == "" {
				logger.Errorf("unauthorized access to an api endpoint")
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			if err := jwtManager.Verify(clientSecret, token, &ctx); err != nil {
				logger.Errorf("could not verify X-Pipedrive-App-Context: %s", err.Error())
				rw.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), request.PipedriveTokenContext{}, ctx)))
		}
	}
}
