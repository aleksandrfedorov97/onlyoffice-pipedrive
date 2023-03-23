package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"golang.org/x/sync/semaphore"
)

var _ErrNotAdmin = errors.New("no admin access")

type apiController struct {
	namespace        string
	client           client.Client
	apiClient        pclient.PipedriveApiClient
	commandClient    pclient.CommandClient
	jwtManager       crypto.JwtManager
	logger           log.Logger
	allowedDownloads int
}

func NewApiController(
	namespace string, client client.Client,
	jwtManager crypto.JwtManager, allowedDownloads int, logger log.Logger) apiController {
	return apiController{
		namespace:        namespace,
		client:           client,
		apiClient:        pclient.NewPipedriveApiClient(),
		commandClient:    pclient.NewCommandClient(jwtManager),
		jwtManager:       jwtManager,
		logger:           logger,
		allowedDownloads: allowedDownloads,
	}
}

func (c *apiController) BuildGetMe() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		pctx, ok := r.Context().Value(request.PipedriveTokenContext{}).(request.PipedriveTokenContext)
		if !ok {
			rw.WriteHeader(http.StatusForbidden)
			c.logger.Error("could not extract pipedrive context from the context")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		var ures response.UserResponse
		if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser", fmt.Sprint(pctx.UID+pctx.CID)), &ures); err != nil {
			c.logger.Errorf("could not get user access info: %s", err.Error())
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

		rw.Write(response.UserTokenResponse{
			ID:          ures.ID,
			AccessToken: ures.AccessToken,
			ExpiresAt:   ures.ExpiresAt,
		}.ToJSON())
	}
}

func (c *apiController) BuildPostSettings() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		pctx, ok := r.Context().Value(request.PipedriveTokenContext{}).(request.PipedriveTokenContext)
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
		errChan := make(chan error, 2)
		cidChan := make(chan int, 1)

		go func() {
			wg.Add(1)
			defer wg.Done()
			if err := c.commandClient.License(ctx, settings.DocAddress, settings.DocSecret); err != nil {
				c.logger.Errorf("could not validate ONLYOFFICE document server credentials: %s", err.Error())
				errChan <- err
				return
			}
		}()

		go func() {
			var ures response.UserResponse
			if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser", fmt.Sprint(pctx.UID+pctx.CID)), &ures); err != nil {
				c.logger.Errorf("could not get user access info: %s", err.Error())
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
		default:
		}

		msg := c.client.NewMessage("insert-settings", request.DocSettings{
			CompanyID:  <-cidChan,
			DocAddress: settings.DocAddress,
			DocSecret:  settings.DocSecret,
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
		pctx, ok := r.Context().Value(request.PipedriveTokenContext{}).(request.PipedriveTokenContext)
		if !ok {
			rw.WriteHeader(http.StatusForbidden)
			c.logger.Error("could not extract pipedrive context from the context")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		var ures response.UserResponse
		if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser", fmt.Sprint(pctx.UID+pctx.CID)), &ures); err != nil {
			c.logger.Errorf("could not get user access info: %s", err.Error())
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

		pctx, ok := r.Context().Value(request.PipedriveTokenContext{}).(request.PipedriveTokenContext)
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

func (c apiController) BuildGetFile() http.HandlerFunc {
	sem := semaphore.NewWeighted(int64(c.allowedDownloads))
	return func(rw http.ResponseWriter, r *http.Request) {
		if ok := sem.TryAcquire(1); !ok {
			c.logger.Warn("too many download requests")
			rw.WriteHeader(http.StatusTooManyRequests)
			return
		}

		defer sem.Release(1)

		fid, cid, token := strings.TrimSpace(r.URL.Query().Get("fid")),
			strings.TrimSpace(r.URL.Query().Get("cid")), strings.TrimSpace(r.URL.Query().Get("token"))

		var pctx request.PipedriveTokenContext
		if token == "" {
			c.logger.Errorf("unauthorized access to an api endpoint")
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}

		var docs response.DocSettingsResponse
		if err := c.client.Call(r.Context(), c.client.NewRequest(fmt.Sprintf("%s:settings", c.namespace), "SettingsSelectHandler.GetSettings", cid), &docs); err != nil {
			c.logger.Debugf("could not document server settings: %s", err.Error())
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := c.jwtManager.Verify(docs.DocSecret, token, &pctx); err != nil {
			c.logger.Errorf("could not verify X-Pipedrive-App-Context: %s", err.Error())
			rw.WriteHeader(http.StatusForbidden)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		var ures response.UserResponse
		if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser", fmt.Sprint(pctx.UID+pctx.CID)), &ures); err != nil {
			c.logger.Errorf("could not get user access info: %s", err.Error())
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

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/files/%s/download", ures.ApiDomain, fid), nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ures.AccessToken))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		defer resp.Body.Close()
		io.Copy(rw, resp.Body)
	}
}
