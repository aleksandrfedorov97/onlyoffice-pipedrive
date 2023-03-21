package handler

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/core/port"
	"go-micro.dev/v4/client"
)

type UserDeleteHandler struct {
	service port.UserAccessService
	client  client.Client
	logger  log.Logger
}

func NewUserDeleteHandler(
	service port.UserAccessService,
	client client.Client,
	logger log.Logger,
) UserDeleteHandler {
	return UserDeleteHandler{
		service: service,
		client:  client,
		logger:  logger,
	}
}

func (u UserDeleteHandler) DeleteUser(ctx context.Context, uid *string, res *interface{}) error {
	u.logger.Debugf("removing user %s", *uid)
	return u.service.DeleteUser(ctx, *uid)
}
