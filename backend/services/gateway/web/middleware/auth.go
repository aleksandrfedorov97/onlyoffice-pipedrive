package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
)

func BuildHandleAuthMiddleware(
	clientID, clientSecret string,
	logger log.Logger,
) func(next http.Handler) http.HandlerFunc {
	logger.Debugf("zoom event middleware has been built with client_id %s and client_secret %s", clientID, clientSecret)
	return func(next http.Handler) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			signature := strings.ReplaceAll(r.Header.Get("Authorization"), "Basic ", "")
			if signature == "" {
				logger.Errorf("an unauthorized access to deauth endpoint")
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			validSignature := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))
			if signature != validSignature {
				logger.Errorf("invalid uninstall signature")
				logger.Debugf("valid signature is %s whereas signature is %s", validSignature, signature)
				rw.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(rw, r)
		}
	}
}
