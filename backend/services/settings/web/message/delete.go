package message

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/port"
)

type DeleteMessageHandler struct {
	service port.DocSettingsService
}

func BuildDeleteMessageHandler(service port.DocSettingsService) DeleteMessageHandler {
	return DeleteMessageHandler{
		service: service,
	}
}

func (i DeleteMessageHandler) GetHandler() func(context.Context, interface{}) error {
	return func(ctx context.Context, payload interface{}) error {
		if cid, ok := payload.(string); !ok {
			return _ErrInvalidMessagePayload
		} else {
			return i.service.DeleteSettings(ctx, cid)
		}
	}
}
