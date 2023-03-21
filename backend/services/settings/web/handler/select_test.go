package handler

import (
	"context"
	"testing"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/adapter"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/service"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/shared/response"
	"github.com/stretchr/testify/assert"
)

type mockEncryptor struct{}

func (e mockEncryptor) Encrypt(text string) (string, error) {
	return string(text), nil
}

func (e mockEncryptor) Decrypt(ciphertext string) (string, error) {
	return string(ciphertext), nil
}

func TestSelectCaching(t *testing.T) {
	adapter := adapter.NewMemoryDocserverAdapter()
	service := service.NewSettingsService(adapter, mockEncryptor{}, log.NewEmptyLogger())

	sel := NewSettingsSelectHandler(service, nil, log.NewEmptyLogger())

	service.CreateSettings(context.Background(), domain.DocSettings{
		CompanyID:  "mock",
		DocAddress: "mock",
		DocSecret:  "mock",
	})

	t.Run("get settings", func(t *testing.T) {
		var res response.DocSettingsResponse
		id := "mock"
		assert.NoError(t, sel.GetSettings(context.Background(), &id, &res))
		assert.NotEmpty(t, res)
	})
}
