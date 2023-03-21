package adapter

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/domain"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/settings/web/core/port"
)

var _ErrNoCompanySettings = errors.New("no company settings")

type memoryDocserverAdapter struct {
	kvs map[string][]byte
}

func NewMemoryDocserverAdapter() port.DocSettingsServiceAdapter {
	return &memoryDocserverAdapter{
		kvs: make(map[string][]byte),
	}
}

func (m *memoryDocserverAdapter) save(settings domain.DocSettings) error {
	buffer, err := json.Marshal(settings)

	if err != nil {
		return err
	}

	m.kvs[settings.CompanyID] = buffer

	return nil
}

func (m *memoryDocserverAdapter) InsertSettings(ctx context.Context, settings domain.DocSettings) error {
	return m.save(settings)
}

func (m *memoryDocserverAdapter) SelectSettings(ctx context.Context, cid string) (domain.DocSettings, error) {
	buffer, ok := m.kvs[cid]
	var settings domain.DocSettings

	if !ok {
		return settings, _ErrNoCompanySettings
	}

	if err := json.Unmarshal(buffer, &settings); err != nil {
		return settings, err
	}

	return settings, nil
}

func (m *memoryDocserverAdapter) UpsertSettings(ctx context.Context, settings domain.DocSettings) (domain.DocSettings, error) {
	if err := m.save(settings); err != nil {
		return domain.DocSettings{}, err
	}

	return settings, nil
}

func (m *memoryDocserverAdapter) DeleteSettings(ctx context.Context, cid string) error {
	if _, ok := m.kvs[cid]; !ok {
		return _ErrNoCompanySettings
	}

	delete(m.kvs, cid)

	return nil
}
