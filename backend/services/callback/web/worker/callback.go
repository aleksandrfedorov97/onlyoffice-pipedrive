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

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client/model"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/message"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/util/backoff"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type workerContext struct{}

type callbackWorker struct {
	namespace     string
	client        client.Client
	pipedriveAPI  pclient.PipedriveApiClient
	uploadTimeout int
	logger        log.Logger
}

func NewWorkerContext() workerContext {
	return workerContext{}
}

func NewCallbackWorker(namespace string, client client.Client, uploadTimeout int, logger log.Logger) callbackWorker {
	return callbackWorker{
		namespace:     namespace,
		client:        client,
		pipedriveAPI:  pclient.NewPipedriveApiClient(),
		uploadTimeout: uploadTimeout,
		logger:        logger,
	}
}
func (c callbackWorker) UploadFile(ctx context.Context, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.uploadTimeout)*time.Second)
	defer cancel()

	tracer := otel.GetTracerProvider().Tracer("pipedrive-onlyoffice/pool")
	tctx, span := tracer.Start(ctx, "upload")
	defer span.End()

	var msg message.JobMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		logger.Errorf("could not notify ws clients. Reason: %s", err.Error())
		return err
	}

	c.logger.Debugf("got a new file %s upload job (%s)", msg.Filename, msg.UID)

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
		if err := c.client.Call(tctx, req, &ures, client.WithRetries(3), client.WithBackoff(func(ctx context.Context, req client.Request, attempts int) (time.Duration, error) {
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

		headResp, err := otelhttp.Head(tctx, msg.Url)
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

	c.logger.Debugf("worker is waiting for waitgroup")
	wg.Wait()
	c.logger.Debugf("worker waitgroup ok")

	select {
	case err := <-errChan:
		c.logger.Debugf("an error from the channel: %s", err.Error())
		return err
	default:
		c.logger.Debugf("select default")
	}

	ures := <-userChan
	if err := c.pipedriveAPI.UploadFile(tctx, msg.Url, msg.Deal, msg.FileID, msg.Filename, <-sizeChan, model.Token{
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
