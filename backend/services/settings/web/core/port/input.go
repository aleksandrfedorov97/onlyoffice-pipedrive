package port

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
)

type DocSettingsService interface {
	CreateSettings(ctx context.Context, settings domain.DocSettings) error
	GetSettings(ctx context.Context, cid string) (domain.DocSettings, error)
	UpdateSettings(ctx context.Context, settings domain.DocSettings) (domain.DocSettings, error)
	DeleteSettings(ctx context.Context, cid string) error
}
