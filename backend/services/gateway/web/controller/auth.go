package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"go-micro.dev/v4/client"
	"golang.org/x/sync/singleflight"
)

var group singleflight.Group

type authController struct {
	namespace     string
	redirectURI   string
	client        client.Client
	pipedriveAuth pclient.PipedriveAuthClient
	pipedriveAPI  pclient.PipedriveApiClient
	logger        log.Logger
}

func NewAuthController(
	namespace string,
	redirectURI string,
	client client.Client,
	pipedriveAuth pclient.PipedriveAuthClient,
	logger log.Logger,
) *authController {
	return &authController{
		namespace:     namespace,
		redirectURI:   redirectURI,
		client:        client,
		pipedriveAuth: pipedriveAuth,
		pipedriveAPI:  pclient.NewPipedriveApiClient(),
		logger:        logger,
	}
}

func (c authController) BuildGetAuth() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c.logger.Debug("a new auth request")
		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Debug("empty auth code parameter")
			return
		}

		group.Do(code, func() (interface{}, error) {
			token, err := c.pipedriveAuth.GetAccessToken(r.Context(), code, c.redirectURI)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				c.logger.Errorf("could not get pipedrive access token: %s", err.Error())
				return nil, err
			}

			usr, err := c.pipedriveAPI.GetMe(r.Context(), token)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				c.logger.Errorf("could not get pipedrive user: %s", err.Error())
				return nil, err
			}

			if err := c.client.Publish(r.Context(), client.NewMessage("insert-auth", response.UserResponse{
				ID:           fmt.Sprint(usr.ID + usr.CompanyID),
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				TokenType:    token.TokenType,
				Scope:        token.Scope,
				ApiDomain:    token.ApiDomain,
				ExpiresAt:    time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn-700)).UnixMilli(),
			})); err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				c.logger.Errorf("insert user error: %s", err.Error())
				return nil, err
			}

			c.logger.Debugf("redirecting to api domain: %s", token.ApiDomain)
			http.Redirect(rw, r, token.ApiDomain, http.StatusMovedPermanently)
			return nil, nil
		})
	}
}

func (c authController) BuildDeleteAuth() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c.logger.Debug("a new uninstall request")
		var ureq request.UninstallRequest

		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Errorf("could not parse request body: %s", err.Error())
			return
		}

		if err := json.Unmarshal(buf, &ureq); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			c.logger.Errorf("could not unmarshal request body: %s", err.Error())
			return
		}

		var res interface{}
		if err := c.client.Call(r.Context(), c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserDeleteHandler.DeleteUser", fmt.Sprint(ureq.UserID+ureq.CompanyID)), &res); err != nil {
			c.logger.Errorf("could not delete user: %s", err.Error())
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

		c.logger.Debugf("successfully published delete-auth message for user %s", fmt.Sprint(ureq.UserID+ureq.CompanyID))
		rw.WriteHeader(http.StatusOK)
	}
}
