package web

import (
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	plog "github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/service/rpc"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/builder/web/handler"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/crypto"
	"go-micro.dev/v4/cache"
	mclient "go-micro.dev/v4/client"
)

type ConfigRPCServer struct {
	namespace  string
	jwtManager crypto.JwtManager
	logger     plog.Logger
	gatewayURL string
}

func NewConfigRPCServer(
	serverConfig *config.ServerConfig,
	credentialsConfig *config.OAuthCredentialsConfig,
	onlyofficeConfig *shared.OnlyofficeConfig,
	logger log.Logger,
) rpc.RPCEngine {
	jwtManager := crypto.NewOnlyofficeJwtManager()
	return ConfigRPCServer{
		namespace:  serverConfig.Namespace,
		jwtManager: jwtManager,
		logger:     logger,
		gatewayURL: onlyofficeConfig.Onlyoffice.Builder.GatewayURL,
	}
}

func (a ConfigRPCServer) BuildMessageHandlers() []rpc.RPCMessageHandler {
	return nil
}

func (a ConfigRPCServer) BuildHandlers(c mclient.Client, cache cache.Cache) []interface{} {
	return []interface{}{
		handler.NewConfigHandler(a.namespace, a.logger, c, a.jwtManager, a.gatewayURL),
	}
}
