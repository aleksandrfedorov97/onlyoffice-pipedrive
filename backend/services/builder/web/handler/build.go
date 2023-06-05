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

package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ONLYOFFICE/onlyoffice-integration-adapters/config"
	"github.com/ONLYOFFICE/onlyoffice-integration-adapters/crypto"
	plog "github.com/ONLYOFFICE/onlyoffice-integration-adapters/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client/model"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/constants"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"github.com/mileusna/useragent"
	"go-micro.dev/v4/client"
)

type ConfigHandler struct {
	client     client.Client
	apiClient  pclient.PipedriveApiClient
	jwtManager crypto.JwtManager
	config     *config.ServerConfig
	onlyoffice *shared.OnlyofficeConfig
	logger     plog.Logger
}

func NewConfigHandler(
	client client.Client,
	jwtManager crypto.JwtManager,
	apiClient pclient.PipedriveApiClient,
	config *config.ServerConfig,
	onlyoffice *shared.OnlyofficeConfig,
	logger plog.Logger,
) ConfigHandler {
	return ConfigHandler{
		client:     client,
		apiClient:  apiClient,
		jwtManager: jwtManager,
		config:     config,
		onlyoffice: onlyoffice,
		logger:     logger,
	}
}

func (c ConfigHandler) processConfig(user response.UserResponse, req request.BuildConfigRequest, ctx context.Context) (response.BuildConfigResponse, error) {
	var config response.BuildConfigResponse
	var wg sync.WaitGroup
	wg.Add(2)
	usrChan := make(chan model.User, 1)
	settingsChan := make(chan response.DocSettingsResponse, 1)
	errorsChan := make(chan error, 2)

	go func() {
		defer wg.Done()
		u, err := c.apiClient.GetMe(ctx, model.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
			TokenType:    user.TokenType,
			Scope:        user.Scope,
			ApiDomain:    user.ApiDomain,
		})

		if err != nil {
			c.logger.Debugf("could not get pipedrive user: %s", err.Error())
			errorsChan <- err
			return
		}

		c.logger.Debugf("populating pipedrive user %d channel", u.ID)
		usrChan <- u
		c.logger.Debugf("successfully populated pipedrive channel")
	}()

	go func() {
		defer wg.Done()
		var docs response.DocSettingsResponse
		if err := c.client.Call(
			ctx,
			c.client.NewRequest(
				fmt.Sprintf("%s:settings", c.config.Namespace),
				"SettingsSelectHandler.GetSettings",
				fmt.Sprint(req.CID),
			),
			&docs,
		); err != nil {
			c.logger.Debugf("could not document server settings: %s", err.Error())
			errorsChan <- err
			return
		}

		if docs.DocAddress == "" || docs.DocSecret == "" || docs.DocHeader == "" {
			c.logger.Debugf("no settings found")
			errorsChan <- ErrNoSettingsFound
			return
		}

		c.logger.Debugf("populating document server %d settings channel", req.CID)
		settingsChan <- docs
		c.logger.Debugf("successfully populated document server settings channel")
	}()

	c.logger.Debugf("waiting for goroutines to finish execution")
	wg.Wait()
	c.logger.Debugf("goroutines have finished the execution")

	select {
	case err := <-errorsChan:
		return config, err
	case <-ctx.Done():
		return config, ErrOperationTimeout
	default:
		c.logger.Debugf("select default")
	}

	usr := <-usrChan
	settings := <-settingsChan
	t := "desktop"
	ua := useragent.Parse(req.UserAgent)

	if ua.Mobile || ua.Tablet {
		t = "mobile"
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	dreq, _ := http.NewRequest("GET", fmt.Sprintf("%s/files/%s/download", user.ApiDomain, req.FileID), nil)
	dreq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", user.AccessToken))
	resp, err := client.Do(dreq)
	if err != nil {
		return config, err
	}

	filename := shared.EscapeFilename(req.Filename)
	config = response.BuildConfigResponse{
		Document: response.Document{
			Key:   req.DocKey,
			Title: filename,
			URL:   resp.Header.Get("Location"),
		},
		EditorConfig: response.EditorConfig{
			User: response.User{
				ID:   fmt.Sprint(usr.ID + usr.CompanyID),
				Name: usr.Name,
			},
			CallbackURL: fmt.Sprintf(
				"%s/callback?cid=%d&did=%s&fid=%s&filename=%s",
				c.onlyoffice.Onlyoffice.Builder.CallbackURL,
				usr.CompanyID, req.Deal, req.FileID,
				url.QueryEscape(filename),
			),
			Customization: response.Customization{
				Goback: response.Goback{
					RequestClose: false,
				},
				Plugins:       false,
				HideRightMenu: false,
			},
			Lang: usr.Language.Lang,
		},
		Type:      t,
		ServerURL: settings.DocAddress,
	}

	if strings.TrimSpace(filename) != "" {
		ext := strings.ReplaceAll(filepath.Ext(filename), ".", "")
		fileType, err := constants.GetFileType(ext)
		if err != nil {
			return config, err
		}
		config.Document.FileType = ext
		config.Document.Permissions = response.Permissions{
			Edit:                 constants.IsExtensionEditable(ext),
			Comment:              true,
			Download:             true,
			Print:                false,
			Review:               false,
			Copy:                 true,
			ModifyContentControl: true,
			ModifyFilter:         true,
		}
		config.DocumentType = fileType
	}

	token, err := c.jwtManager.Sign(settings.DocSecret, config)
	if err != nil {
		c.logger.Debugf("could not sign document server config: %s", err.Error())
		return config, err
	}

	config.Token = token
	return config, nil
}

func (c ConfigHandler) BuildConfig(ctx context.Context, payload request.BuildConfigRequest, res *response.BuildConfigResponse) error {
	c.logger.Debugf("processing a docs config: %s", payload.Filename)

	config, err, _ := group.Do(fmt.Sprint(payload.UID+payload.CID), func() (interface{}, error) {
		req := c.client.NewRequest(
			fmt.Sprintf("%s:auth", c.config.Namespace), "UserSelectHandler.GetUser",
			fmt.Sprint(payload.UID+payload.CID),
		)

		var ures response.UserResponse
		if err := c.client.Call(ctx, req, &ures); err != nil {
			c.logger.Debugf("could not get user %d access info: %s", payload.UID+payload.CID, err.Error())
			return nil, err
		}

		config, err := c.processConfig(ures, payload, ctx)
		if err != nil {
			return nil, err
		}

		return config, nil
	})

	if cfg, ok := config.(response.BuildConfigResponse); ok {
		*res = cfg
		return nil
	}

	return err
}
