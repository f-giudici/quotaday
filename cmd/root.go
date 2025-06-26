/*
Copyright © 2025 Francesco Giudici <dev@foggy.day>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/fgday/quotaday/api"
)

func Execute() {
	app := &cli.App{
		Usage: "start Quotaday webserver",
		Commands: []*cli.Command{
			newVersionCommand(),
		},
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "port to listen to",
				Value:   80,
			},
		},
		Action: func(cCtx *cli.Context) error {
			port := fmt.Sprintf(":%d", cCtx.Uint("port"))
			log.Printf("Starting Quotaday %s on port %s\n", versionString(), port)

			server := api.NewServer()
			r := http.NewServeMux()
			h := api.HandlerFromMux(server, r)

			s := &http.Server{
				Handler: h,
				Addr:    "0.0.0.0" + port,
			}

			return s.ListenAndServe()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
