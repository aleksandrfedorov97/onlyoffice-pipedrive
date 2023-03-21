package port

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
)

type DocSettingsServiceAdapter interface {
	InsertSettings(ctx context.Context, settings domain.DocSettings) error
	SelectSettings(ctx context.Context, cid string) (domain.DocSettings, error)
	UpsertSettings(ctx context.Context, settings domain.DocSettings) (domain.DocSettings, error)
	DeleteSettings(ctx context.Context, cid string) error
}
