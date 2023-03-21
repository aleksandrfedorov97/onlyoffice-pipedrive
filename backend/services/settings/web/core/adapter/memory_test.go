package adapter

import (
	"context"
	"testing"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestMemoryAdapter(t *testing.T) {
	adapter := NewMemoryDocserverAdapter()

	t.Run("save settings", func(t *testing.T) {
		assert.NoError(t, adapter.InsertSettings(context.Background(), settings))
	})

	t.Run("save the same settings object", func(t *testing.T) {
		assert.NoError(t, adapter.InsertSettings(context.Background(), settings))
	})

	t.Run("get settings by cid", func(t *testing.T) {
		s, err := adapter.SelectSettings(context.Background(), "mock")
		assert.NoError(t, err)
		assert.Equal(t, settings, s)
	})

	t.Run("update settings by cid", func(t *testing.T) {
		s, err := adapter.UpsertSettings(context.Background(), domain.DocSettings{
			CompanyID:  "mock",
			DocAddress: "mock",
			DocSecret:  "mock",
		})
		assert.NoError(t, err)
		assert.NotNil(t, s)
	})

	t.Run("delete settings by cid", func(t *testing.T) {
		assert.NoError(t, adapter.DeleteSettings(context.Background(), "mock"))
	})

	t.Run("get invalid settings", func(t *testing.T) {
		_, err := adapter.SelectSettings(context.Background(), "mock")
		assert.Error(t, err)
	})
}
