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

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client/model"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"go-micro.dev/v4/client"
)

var _ErrNotAdmin = errors.New("no admin access")

type apiController struct {
	namespace     string
	client        client.Client
	apiClient     pclient.PipedriveApiClient
	commandClient pclient.CommandClient
	jwtManager    crypto.JwtManager
	logger        log.Logger
}

func NewApiController(
	namespace string, client client.Client,
	jwtManager crypto.JwtManager, logger log.Logger) apiController {
	return apiController{
		namespace:     namespace,
		client:        client,
		apiClient:     pclient.NewPipedriveApiClient(),
		commandClient: pclient.NewCommandClient(jwtManager),
		jwtManager:    jwtManager,
		logger:        logger,
	}
}

func (c *apiController) getUser(ctx context.Context, id string) (response.UserResponse, int, error) {
	var ures response.UserResponse
	if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser", id), &ures); err != nil {
		c.logger.Errorf("could not get user access info: %s", err.Error())
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ures, http.StatusRequestTimeout, err
		}

		microErr := response.MicroError{}
		if err := json.Unmarshal([]byte(err.Error()), &microErr); err != nil {
			return ures, http.StatusUnauthorized, err
		}

		return ures, microErr.Code, err
	}

	return ures, http.StatusOK, nil
}

func (c *apiController) BuildGetMe() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		pctx, ok := r.Context().Value("X-Pipedrive-App-Context").(request.PipedriveTokenContext)
		if !ok {
			rw.WriteHeader(http.StatusForbidden)
			c.logger.Error("could not extract pipedrive context from the context")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		ures, status, _ := c.getUser(ctx, fmt.Sprint(pctx.UID+pctx.CID))
		if status != http.StatusOK {
			rw.WriteHeader(status)
			return
		}

		rw.Write(response.UserTokenResponse{
			ID:          ures.ID,
			AccessToken: ures.AccessToken,
			ExpiresAt:   ures.ExpiresAt,
		}.ToJSON())
	}
}

func (c *apiController) BuildPostSettings() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		pctx, ok := r.Context().Value("X-Pipedrive-App-Context").(request.PipedriveTokenContext)
		if !ok {
			rw.WriteHeader(http.StatusForbidden)
			c.logger.Error("could not extract pipedrive context from the context")
			return
		}

		len, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 0)
		if err != nil || (len/100000) > 10 {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		var settings request.DocSettings
		buf, _ := ioutil.ReadAll(r.Body)
		if err := json.Unmarshal(buf, &settings); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Errorf(err.Error())
			return
		}

		settings.CompanyID = pctx.CID
		if err := settings.Validate(); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Errorf("invalid settings format: %s", err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(2)
		errChan := make(chan error, 2)
		cidChan := make(chan int, 1)

		go func() {
			defer wg.Done()
			if err := c.commandClient.License(ctx, settings.DocAddress, settings.DocSecret); err != nil {
				c.logger.Errorf("could not validate ONLYOFFICE document server credentials: %s", err.Error())
				errChan <- err
				return
			}
		}()

		go func() {
			defer wg.Done()
			ures, _, err := c.getUser(ctx, fmt.Sprint(pctx.UID+pctx.CID))
			if err != nil {
				errChan <- err
				return
			}

			urs, err := c.apiClient.GetMe(ctx, model.Token{
				AccessToken:  ures.AccessToken,
				RefreshToken: ures.RefreshToken,
				TokenType:    ures.TokenType,
				Scope:        ures.Scope,
				ApiDomain:    ures.ApiDomain,
			})

			for _, access := range urs.Access {
				if access.App == "global" && !access.Admin {
					errChan <- _ErrNotAdmin
					return
				}
			}

			if err != nil {
				c.logger.Errorf("could not get pipedrive user or no user has admin permissions")
				errChan <- err
				return
			}

			cidChan <- urs.CompanyID
		}()

		wg.Wait()

		select {
		case <-errChan:
			rw.WriteHeader(http.StatusForbidden)
			return
		case <-ctx.Done():
			rw.WriteHeader(http.StatusRequestTimeout)
			return
		default:
		}

		msg := c.client.NewMessage("insert-settings", request.DocSettings{
			CompanyID:  <-cidChan,
			DocAddress: settings.DocAddress,
			DocSecret:  settings.DocSecret,
			DocHeader:  settings.DocHeader,
		})

		if err := c.client.Publish(ctx, msg); err != nil {
			c.logger.Errorf("could not insert settings: %s", err.Error())
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				rw.WriteHeader(http.StatusRequestTimeout)
				return
			}

			microErr := response.MicroError{}
			if err := json.Unmarshal([]byte(err.Error()), &microErr); err != nil {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			rw.WriteHeader(microErr.Code)
			return
		}

		rw.WriteHeader(http.StatusCreated)
	}
}

func (c apiController) BuildGetSettings() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		pctx, ok := r.Context().Value("X-Pipedrive-App-Context").(request.PipedriveTokenContext)
		if !ok {
			rw.WriteHeader(http.StatusForbidden)
			c.logger.Error("could not extract pipedrive context from the context")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		ures, status, _ := c.getUser(ctx, fmt.Sprint(pctx.UID+pctx.CID))
		if status != http.StatusOK {
			rw.WriteHeader(status)
			return
		}

		urs, _ := c.apiClient.GetMe(ctx, model.Token{
			AccessToken:  ures.AccessToken,
			RefreshToken: ures.RefreshToken,
			TokenType:    ures.TokenType,
			Scope:        ures.Scope,
			ApiDomain:    ures.ApiDomain,
		})

		for _, access := range urs.Access {
			if access.App == "global" && !access.Admin {
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}

		var docs response.DocSettingsResponse
		if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:settings", c.namespace), "SettingsSelectHandler.GetSettings", fmt.Sprint(pctx.CID)), &docs); err != nil {
			c.logger.Errorf("could not get settings: %s", err.Error())
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				rw.WriteHeader(http.StatusRequestTimeout)
				return
			}

			microErr := response.MicroError{}
			if err := json.Unmarshal([]byte(err.Error()), &microErr); err != nil {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}

			rw.WriteHeader(microErr.Code)
			return
		}

		rw.Write(docs.ToJSON())
	}
}

func (c apiController) BuildGetConfig() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		id, filename, key, dealID := strings.TrimSpace(query.Get("id")), strings.TrimSpace(query.Get("name")),
			strings.TrimSpace(query.Get("key")), strings.TrimSpace(query.Get("deal_id"))

		pctx, ok := r.Context().Value("X-Pipedrive-App-Context").(request.PipedriveTokenContext)
		if !ok {
			rw.WriteHeader(http.StatusForbidden)
			c.logger.Error("could not extract pipedrive context from the context")
			return
		}

		if filename == "" {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Error("could not extract file name from URL Query")
			return
		}

		if len(filename) > 200 {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Error("file length is greater than 200")
			return
		}

		if key == "" {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Error("could not extract doc key from URL Query")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		var resp response.BuildConfigResponse
		if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:builder", c.namespace), "ConfigHandler.BuildConfig", request.BuildConfigRequest{
			UID:       pctx.UID,
			CID:       pctx.CID,
			Deal:      dealID,
			UserAgent: r.UserAgent(),
			Filename:  filename,
			FileID:    id,
			DocKey:    key,
		}), &resp); err != nil {
			c.logger.Errorf("could not build onlyoffice config: %s", err.Error())
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				rw.WriteHeader(http.StatusRequestTimeout)
				return
			}

			microErr := response.MicroError{}
			if err := json.Unmarshal([]byte(err.Error()), &microErr); err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}

			c.logger.Errorf("build config micro error: %s", microErr.Detail)
			rw.WriteHeader(microErr.Code)
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(resp.ToJSON())
	}
}
