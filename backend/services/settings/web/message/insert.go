package message

import (
	"context"
	"fmt"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/port"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/request"
	"github.com/mitchellh/mapstructure"
)

type InsertMessageHandler struct {
	service port.DocSettingsService
}

func BuildInsertMessageHandler(service port.DocSettingsService) InsertMessageHandler {
	return InsertMessageHandler{
		service: service,
	}
}

func (i InsertMessageHandler) GetHandler() func(context.Context, interface{}) error {
	return func(ctx context.Context, payload interface{}) error {
		var settings request.DocSettings
		if err := mapstructure.Decode(payload, &settings); err != nil {
			return err
		}
		_, err := i.service.UpdateSettings(ctx, domain.DocSettings{
			CompanyID:  fmt.Sprint(settings.CompanyID),
			DocAddress: settings.DocAddress,
			DocSecret:  settings.DocSecret,
		})
		return err
	}
}
