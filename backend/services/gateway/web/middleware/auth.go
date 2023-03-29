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
	logger.Debugf("pipedrive event middleware has been built with client_id %s and client_secret %s", clientID, clientSecret)
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
