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

package cmd

import (
	"context"
	"net/http"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/log"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/service/rpc"
	"github.com/ONLYOFFICE/onlyoffice-pipedrive/services/auth/web"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
)

func Server() *cli.Command {
	return &cli.Command{
		Name:     "server",
		Usage:    "starts a new rpc server instance",
		Category: "server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config_path",
				Usage:   "sets custom configuration path",
				Aliases: []string{"config", "conf", "c"},
			},
			&cli.StringFlag{
				Name:    "environment",
				Usage:   "sets servers environment (development, testing, production)",
				Aliases: []string{"env", "e"},
			},
		},
		Action: func(c *cli.Context) error {
			var (
				CONFIG_PATH = c.String("config_path")
				// ENVIRONMENT = c.String("environment")
			)

			fx.New(
				pkg.Bootstrap(CONFIG_PATH),
				fx.Provide(rpc.NewService),
				fx.Provide(web.NewAuthRPCServer),
				fx.Invoke(func(lifecycle fx.Lifecycle, service micro.Service, repl *http.Server, logger log.Logger) {
					lifecycle.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							go repl.ListenAndServe()
							go service.Run()
							return nil
						},
						OnStop: func(ctx context.Context) error {
							g, gCtx := errgroup.WithContext(ctx)
							g.Go(func() error {
								return repl.Shutdown(gCtx)
							})
							return g.Wait()
						},
					})
				}),
				fx.NopLogger,
			).Run()

			return nil
		},
	}
}
