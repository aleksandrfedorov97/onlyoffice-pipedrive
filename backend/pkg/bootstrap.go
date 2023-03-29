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

package pkg

import (
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/cache"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/messaging"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/registry"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/service/repl"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/trace"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/worker"
	"go.uber.org/fx"
)

func Bootstrap(path string) fx.Option {
	return fx.Module(
		"pkg",
		fx.Provide(config.BuildNewCacheConfig(path)),
		fx.Provide(config.BuildNewCorsConfig(path)),
		fx.Provide(config.BuildNewCredentialsConfig(path)),
		fx.Provide(config.BuildNewLoggerConfig(path)),
		fx.Provide(config.BuildNewMessagingConfig(path)),
		fx.Provide(config.BuildNewPersistenceConfig(path)),
		fx.Provide(config.BuildNewRegistryConfig(path)),
		fx.Provide(config.BuildNewResilienceConfig(path)),
		fx.Provide(config.BuildNewServerConfig(path)),
		fx.Provide(config.BuildNewTracerConfig(path)),
		fx.Provide(config.BuildNewWorkerConfig(path)),
		fx.Provide(cache.NewCache),
		fx.Provide(log.NewLogrusLogger),
		fx.Provide(registry.NewRegistry),
		fx.Provide(messaging.NewBroker),
		fx.Provide(trace.NewTracer),
		fx.Provide(worker.NewBackgroundWorker),
		fx.Provide(worker.NewBackgroundEnqueuer),
		fx.Provide(repl.NewService),
	)
}
