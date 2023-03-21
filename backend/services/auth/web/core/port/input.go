package port

import (
	"context"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web/core/domain"
)

type UserAccessService interface {
	CreateUser(ctx context.Context, user domain.UserAccess) error
	GetUser(ctx context.Context, uid string) (domain.UserAccess, error)
	UpdateUser(ctx context.Context, user domain.UserAccess) (domain.UserAccess, error)
	DeleteUser(ctx context.Context, uid string) error
}
