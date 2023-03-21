package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	plog "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/worker"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/message"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"go-micro.dev/v4/client"
)

type callbackController struct {
	namespace    string
	maxSize      int64
	logger       plog.Logger
	client       client.Client
	pipedriveAPI pclient.PipedriveApiClient
	jwtManager   crypto.JwtManager
}

func NewCallbackController(
	namespace string,
	maxSize int64,
	logger plog.Logger,
	client client.Client,
) *callbackController {
	return &callbackController{
		namespace:    namespace,
		maxSize:      maxSize,
		logger:       logger,
		client:       client,
		pipedriveAPI: pclient.NewPipedriveApiClient(),
		jwtManager:   crypto.NewOnlyofficeJwtManager(),
	}
}

func (c callbackController) BuildPostHandleCallback(enqueuer worker.BackgroundEnqueuer) http.HandlerFunc {
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
				if err := enqueuer.EnqueueContext(r.Context(), "pipedrive-callback-upload", message.JobMessage{
					UID:      usr,
					Deal:     did,
					FileID:   fid,
					Filename: filename,
					Url:      body.URL,
				}.ToJSON(), worker.WithMaxRetry(3)); err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					c.logger.Errorf("could not enqueue a new job with key %s", body.Key)
					rw.Write(response.CallbackResponse{
						Error: 1,
					}.ToJSON())
					return
				}
			}
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(response.CallbackResponse{
			Error: 0,
		}.ToJSON())
	}
}
