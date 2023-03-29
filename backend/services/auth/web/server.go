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

package web

import (
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/service/rpc"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/core/adapter"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/core/port"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/core/service"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/handler"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/message"
	pclient "github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/client"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"go-micro.dev/v4/cache"
	mclient "go-micro.dev/v4/client"
)

type AuthRPCServer struct {
	service       port.UserAccessService
	pipedriveAuth pclient.PipedriveAuthClient
	logger        log.Logger
}

func NewAuthRPCServer(persistenceConfig *config.PersistenceConfig, oauthConfig *config.OAuthCredentialsConfig, logger log.Logger) rpc.RPCEngine {
	adptr := adapter.NewMemoryUserAdapter()
	if persistenceConfig.Persistence.URL != "" {
		adptr = adapter.NewMongoUserAdapter(persistenceConfig.Persistence.URL)
	}

	service := service.NewUserService(adptr, crypto.NewAesEncryptor([]byte(oauthConfig.Credentials.ClientSecret)), logger)
	return AuthRPCServer{
		service: service,
		pipedriveAuth: pclient.NewPipedriveAuthClient(
			oauthConfig.Credentials.ClientID, oauthConfig.Credentials.ClientSecret,
		),
		logger: logger,
	}
}

func (a AuthRPCServer) BuildMessageHandlers() []rpc.RPCMessageHandler {
	return []rpc.RPCMessageHandler{
		{
			Topic:   "insert-auth",
			Queue:   "pipedrive-auth",
			Handler: message.BuildInsertMessageHandler(a.service).GetHandler(),
		},
	}
}

func (a AuthRPCServer) BuildHandlers(client mclient.Client, cache cache.Cache) []interface{} {
	return []interface{}{
		handler.NewUserSelectHandler(a.service, client, a.pipedriveAuth, a.logger),
		handler.NewUserDeleteHandler(a.service, client, a.logger),
	}
}
