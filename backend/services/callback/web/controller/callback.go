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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	plog "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client/model"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/message"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/util/backoff"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type callbackController struct {
	namespace     string
	maxSize       int64
	uploadTimeout int
	logger        plog.Logger
	client        client.Client
	pipedriveAPI  pclient.PipedriveApiClient
	jwtManager    crypto.JwtManager
}

func NewCallbackController(
	namespace string,
	maxSize int64,
	uploadTimeout int,
	logger plog.Logger,
	client client.Client,
) *callbackController {
	return &callbackController{
		namespace:     namespace,
		maxSize:       maxSize,
		uploadTimeout: uploadTimeout,
		logger:        logger,
		client:        client,
		pipedriveAPI:  pclient.NewPipedriveApiClient(),
		jwtManager:    crypto.NewOnlyofficeJwtManager(),
	}
}

func (c callbackController) UploadFile(ctx context.Context, msg message.JobMessage) error {
	var wg sync.WaitGroup
	wg.Add(2)
	userChan := make(chan response.UserResponse, 1)
	sizeChan := make(chan int64, 1)
	errChan := make(chan error, 2)

	go func() {
		defer wg.Done()

		c.logger.Debugf("trying to get an access token")
		req := c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser", msg.UID)
		var ures response.UserResponse
		if err := c.client.Call(ctx, req, &ures, client.WithRetries(3), client.WithBackoff(func(ctx context.Context, req client.Request, attempts int) (time.Duration, error) {
			return backoff.Do(attempts), nil
		})); err != nil {
			errChan <- err
			return
		}

		c.logger.Debugf("populating user channel")
		userChan <- ures
		c.logger.Debugf("successfully populated user channel")
	}()

	go func() {
		defer wg.Done()

		headResp, err := otelhttp.Head(ctx, msg.Url)
		if err != nil {
			errChan <- err
			return
		}

		size, err := strconv.ParseInt(headResp.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			errChan <- err
			return
		}

		c.logger.Debugf("populating file size channel")
		sizeChan <- size
		c.logger.Debugf("successfully populated file size channel")
	}()

	c.logger.Debugf("callback is waiting for waitgroup")
	wg.Wait()
	c.logger.Debugf("callback waitgroup ok")

	select {
	case err := <-errChan:
		c.logger.Debugf("an error from the channel: %s", err.Error())
		return err
	default:
		c.logger.Debugf("select default")
	}

	ures := <-userChan
	if err := c.pipedriveAPI.UploadFile(ctx, msg.Url, msg.Deal, msg.FileID, msg.Filename, <-sizeChan, model.Token{
		AccessToken:  ures.AccessToken,
		RefreshToken: ures.RefreshToken,
		TokenType:    ures.TokenType,
		Scope:        ures.Scope,
		ApiDomain:    ures.ApiDomain,
	}); err != nil {
		c.logger.Debugf("could not upload an onlyoffice file to pipedrive: %s", err.Error())
		return err
	}

	return nil
}

func (c callbackController) BuildPostHandleCallback() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		cid, did, fid := strings.TrimSpace(query.Get("cid")), strings.TrimSpace(query.Get("did")), strings.TrimSpace(query.Get("fid"))
		rw.Header().Set("Content-Type", "application/json")

		var body request.CallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			c.logger.Errorf("could not decode a callback body")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(response.CallbackResponse{
				Error: 1,
			}.ToJSON())
			return
		}

		if body.Token == "" {
			c.logger.Error("invalid callback body token")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(response.CallbackResponse{
				Error: 1,
			}.ToJSON())
			return
		}

		if cid == "" || did == "" || fid == "" {
			c.logger.Error("invalid query parameter")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(response.CallbackResponse{
				Error: 1,
			}.ToJSON())
			return
		}

		req := c.client.NewRequest(fmt.Sprintf("%s:settings", c.namespace), "SettingsSelectHandler.GetSettings", cid)
		var res response.DocSettingsResponse
		if err := c.client.Call(r.Context(), req, &res); err != nil {
			c.logger.Errorf("could not extract doc server settings %s", cid)
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(response.CallbackResponse{
				Error: 1,
			}.ToJSON())
			return
		}

		if err := c.jwtManager.Verify(res.DocSecret, body.Token, &body); err != nil {
			c.logger.Errorf("could not verify callback jwt (%s). Reason: %s", body.Token, err.Error())
			rw.WriteHeader(http.StatusForbidden)
			rw.Write(response.CallbackResponse{
				Error: 1,
			}.ToJSON())
			return
		}

		if err := body.Validate(); err != nil {
			c.logger.Errorf("invalid callback body. Reason: %s", err.Error())
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(response.CallbackResponse{
				Error: 1,
			}.ToJSON())
			return
		}

		if body.Status == 2 {
			filename := strings.TrimSpace(r.URL.Query().Get("filename"))
			if filename == "" {
				rw.WriteHeader(http.StatusInternalServerError)
				c.logger.Errorf("callback request %s does not contain a filename", body.Key)
				rw.Write(response.CallbackResponse{
					Error: 1,
				}.ToJSON())
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 7*time.Second)
			defer cancel()

			usr := body.Users[0]
			if usr != "" {
				if err := c.pipedriveAPI.ValidateFileSize(ctx, c.maxSize, body.URL); err != nil {
					c.logger.Errorf("could not validate file %s: %s", filename, err.Error())
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write(response.CallbackResponse{
						Error: 1,
					}.ToJSON())
					return
				}

				uctx, cancel := context.WithTimeout(ctx, time.Duration(c.uploadTimeout)*time.Second)
				defer cancel()
				if err := c.UploadFile(uctx, message.JobMessage{
					UID:      usr,
					Deal:     did,
					FileID:   fid,
					Filename: filename,
					Url:      body.URL,
				}); err != nil {
					c.logger.Errorf("could not upload file %s: %s", filename, err.Error())
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write(response.CallbackResponse{
						Error: 1,
					}.ToJSON())
				}
			}
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(response.CallbackResponse{
			Error: 0,
		}.ToJSON())
	}
}
