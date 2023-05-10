package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/gateway/assets"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client/model"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"go-micro.dev/v4/client"
	"golang.org/x/sync/semaphore"
)

type fileController struct {
	namespace        string
	allowedDownloads int
	client           client.Client
	apiClient        pclient.PipedriveApiClient
	jwtManager       crypto.JwtManager
	logger           log.Logger
}

func NewFileController(
	namespace string, allowedDownloads int, client client.Client,
	jwtManager crypto.JwtManager, logger log.Logger) fileController {
	return fileController{
		namespace:        namespace,
		client:           client,
		apiClient:        pclient.NewPipedriveApiClient(),
		jwtManager:       jwtManager,
		logger:           logger,
		allowedDownloads: allowedDownloads,
	}
}

func (c *fileController) getUser(ctx context.Context, id string) (response.UserResponse, int) {
	var ures response.UserResponse
	if err := c.client.Call(ctx, c.client.NewRequest(fmt.Sprintf("%s:auth", c.namespace), "UserSelectHandler.GetUser", id), &ures); err != nil {
		c.logger.Errorf("could not get user access info: %s", err.Error())
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ures, http.StatusRequestTimeout
		}

		microErr := response.MicroError{}
		if err := json.Unmarshal([]byte(err.Error()), &microErr); err != nil {
			return ures, http.StatusUnauthorized
		}

		return ures, microErr.Code
	}

	return ures, http.StatusOK
}

func (c fileController) BuildGetFile() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()
		lang, fileType, dealID, filename := strings.TrimSpace(query.Get("lang")),
			strings.TrimSpace(query.Get("type")), strings.TrimSpace(query.Get("deal")),
			strings.TrimSpace(query.Get("filename"))
		if lang == "" || fileType == "" || dealID == "" || filename == "" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		pctx, ok := r.Context().Value("X-Pipedrive-App-Context").(request.PipedriveTokenContext)
		if !ok {
			rw.WriteHeader(http.StatusForbidden)
			c.logger.Error("could not extract pipedrive context from the context")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()

		ures, status := c.getUser(ctx, fmt.Sprint(pctx.UID+pctx.CID))
		if status != http.StatusOK {
			rw.WriteHeader(status)
			return
		}

		file, err := assets.Files.Open(fmt.Sprintf("assets/%s/new.%s", lang, fileType))
		if err != nil {
			lang = "en-US"
			file, err = assets.Files.Open(fmt.Sprintf("assets/%s/new.%s", lang, fileType))
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				c.logger.Errorf("could not get a new file: %s", err.Error())
				return
			}
			res, ferr := c.apiClient.CreateFile(ctx, dealID, filename, file, model.Token{
				AccessToken:  ures.AccessToken,
				RefreshToken: ures.AccessToken,
				TokenType:    ures.TokenType,
				Scope:        ures.Scope,
				ApiDomain:    ures.ApiDomain,
			})

			if ferr != nil {
				rw.WriteHeader(http.StatusBadRequest)
				c.logger.Errorf("could not upload a pipedrive file: %s", ferr.Error())
				return
			}

			rw.Write(res.ToJSON())
			return
		}

		res, ferr := c.apiClient.CreateFile(ctx, dealID, filename, file, model.Token{
			AccessToken:  ures.AccessToken,
			RefreshToken: ures.AccessToken,
			TokenType:    ures.TokenType,
			Scope:        ures.Scope,
			ApiDomain:    ures.ApiDomain,
		})

		if ferr != nil {
			rw.WriteHeader(http.StatusBadRequest)
			c.logger.Errorf("could not upload a pipedrive file: %s", ferr.Error())
			return
		}

		rw.Write(res.ToJSON())
	}
}

func (c fileController) BuildDownloadFile() http.HandlerFunc {
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

		var wg sync.WaitGroup
		wg.Add(2)
		errChan := make(chan error, 2)

		go func() {
			defer wg.Done()
			if err := c.jwtManager.Verify(docs.DocSecret, token, &pctx); err != nil {
				c.logger.Errorf("could not verify X-Pipedrive-App-Context: %s", err.Error())
				errChan <- err
				return
			}
		}()

		go func() {
			defer wg.Done()
			var tkn interface{}
			if err := c.jwtManager.Verify(docs.DocSecret, strings.ReplaceAll(r.Header.Get(docs.DocHeader), "Bearer ", ""), &tkn); err != nil {
				c.logger.Errorf("could not verify docs header: %s", err.Error())
				errChan <- err
				return
			}
		}()

		wg.Wait()

		select {
		case <-errChan:
			rw.WriteHeader(http.StatusForbidden)
			return
		default:
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		ures, status := c.getUser(ctx, fmt.Sprint(pctx.UID+pctx.CID))
		if status != http.StatusOK {
			rw.WriteHeader(status)
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
