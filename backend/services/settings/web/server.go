package web

import (
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/service/rpc"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/adapter"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/port"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/service"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/handler"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/message"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"go-micro.dev/v4/cache"
	mclient "go-micro.dev/v4/client"
)

type DocserverRPCServer struct {
	service port.DocSettingsService
	logger  log.Logger
}

func NewDocserverRPCServer(
	persistenceConfig *config.PersistenceConfig,
	credentialsConfig *config.OAuthCredentialsConfig,
	logger log.Logger,
) rpc.RPCEngine {
	adptr := adapter.NewMemoryDocserverAdapter()
	if persistenceConfig.Persistence.URL != "" {
		adptr = adapter.NewMongoDocserverAdapter(persistenceConfig.Persistence.URL)
	}

	service := service.NewSettingsService(adptr, crypto.NewAesEncryptor([]byte(credentialsConfig.Credentials.ClientSecret)), logger)
	return DocserverRPCServer{
		service: service,
		logger:  logger,
	}
}

func (a DocserverRPCServer) BuildMessageHandlers() []rpc.RPCMessageHandler {
	return []rpc.RPCMessageHandler{
		{
			Topic:   "insert-settings",
			Queue:   "pipedrive-docserver",
			Handler: message.BuildInsertMessageHandler(a.service).GetHandler(),
		},
		{
			Topic:   "delete-settings",
			Queue:   "pipedrive-docserver",
			Handler: message.BuildDeleteMessageHandler(a.service).GetHandler(),
		},
	}
}

func (a DocserverRPCServer) BuildHandlers(client mclient.Client, cache cache.Cache) []interface{} {
	return []interface{}{
		handler.NewSettingsSelectHandler(a.service, client, a.logger),
	}
}
