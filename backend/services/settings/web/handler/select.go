package handler

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/port"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"go-micro.dev/v4/client"
	"golang.org/x/sync/singleflight"
)

var group singleflight.Group

type SettingsSelectHandler struct {
	service port.DocSettingsService
	client  client.Client
	logger  log.Logger
}

func NewSettingsSelectHandler(
	service port.DocSettingsService,
	client client.Client,
	logger log.Logger,
) SettingsSelectHandler {
	return SettingsSelectHandler{
		service: service,
		client:  client,
		logger:  logger,
	}
}

func (u SettingsSelectHandler) GetSettings(ctx context.Context, cid *string, res *response.DocSettingsResponse) error {
	settings, err, _ := group.Do(*cid, func() (interface{}, error) {
		settings, err := u.service.GetSettings(ctx, *cid)
		if err != nil {
			u.logger.Warnf("could not get company %s settings. Reason: %s", *cid, err.Error())
			return settings, nil
		}

		return settings, nil
	})

	if set, ok := settings.(domain.DocSettings); ok {
		*res = response.DocSettingsResponse{
			DocAddress: set.DocAddress,
			DocSecret:  set.DocSecret,
			DocHeader:  set.DocHeader,
		}
		return nil
	}

	return err
}
