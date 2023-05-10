/**
 *
 * (c) Copyright Ascensio System SIA 2023
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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

			next.ServeHTTP(rw, r.WithContext(context.WithValue(r.Context(), "X-Pipedrive-App-Context", ctx)))
		}
	}
}
