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
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	plog "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client/model"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/constants"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"github.com/mileusna/useragent"
	"go-micro.dev/v4/client"
	"golang.org/x/sync/singleflight"
)

var _ErrNoSettingsFound = errors.New("could not find document server settings")
var _ErrOperationTimeout = errors.New("operation timeout")

type ConfigHandler struct {
	namespace   string
	logger      plog.Logger
	client      client.Client
	apiClient   pclient.PipedriveApiClient
	jwtManager  crypto.JwtManager
	gatewayURL  string
	callbackURL string
	group       singleflight.Group
}

func NewConfigHandler(
	namespace string,
	logger plog.Logger,
	client client.Client,
	jwtManager crypto.JwtManager,
	gatewayURL string,
	callbackURL string,
) ConfigHandler {
	return ConfigHandler{
		namespace:   namespace,
		logger:      logger,
		client:      client,
		apiClient:   pclient.NewPipedriveApiClient(),
		jwtManager:  jwtManager,
		gatewayURL:  gatewayURL,
		callbackURL: callbackURL,
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
		if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:settings", c.namespace), "SettingsSelectHandler.GetSettings", fmt.Sprint(req.CID)), &docs); err != nil {
			c.logger.Debugf("could not document server settings: %s", err.Error())
			errorsChan <- err
			return
		}

		if docs.DocAddress == "" || docs.DocSecret == "" || docs.DocHeader == "" {
			c.logger.Debugf("no settings found")
			errorsChan <- _ErrNoSettingsFound
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
		return config, _ErrOperationTimeout
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

	downloadToken := request.PipedriveTokenContext{
		UID: usr.ID,
		CID: usr.CompanyID,
	}

	downloadToken.IssuedAt = 0
	downloadToken.ExpiresAt = time.Now().Add(4 * time.Minute).UnixMilli()
	tkn, _ := c.jwtManager.Sign(settings.DocSecret, downloadToken)

	filename := shared.EscapeFilename(req.Filename)
	config = response.BuildConfigResponse{
		Document: response.Document{
			Key:   req.DocKey,
			Title: filename,
			URL:   fmt.Sprintf("%s/files/download?cid=%d&fid=%s&token=%s", c.gatewayURL, usr.CompanyID, req.FileID, tkn),
		},
		EditorConfig: response.EditorConfig{
			User: response.User{
				ID:   fmt.Sprint(usr.ID + usr.CompanyID),
				Name: usr.Name,
			},
			CallbackURL: fmt.Sprintf(
				"%s/callback?cid=%d&did=%s&fid=%s&filename=%s",
				c.callbackURL, usr.CompanyID, req.Deal, req.FileID, url.QueryEscape(filename),
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

	config, err, _ := c.group.Do(fmt.Sprint(payload.UID+payload.CID), func() (interface{}, error) {
		req := c.client.NewRequest(
			fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser",
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
