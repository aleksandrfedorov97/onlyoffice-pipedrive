/**
 *
 * (c) Copyright Ascensio System SIA 2023
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
			DocHeader:  settings.DocHeader,
		})
		return err
	}
}
