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

	"github.com/f-giudici/quotaday/pkg/quote"
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
			http.HandleFunc("/", serveHTTP)
			return http.ListenAndServe(port, nil)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(getRemoteHostInfo(r))
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	qb := quote.New()
	qb.FillExample()
	q := qb.RandomQuotation()
	if err := q.WriteHTML(w); err != nil {
		log.Fatal(err)
	}
}

func getRemoteHostInfo(r *http.Request) string {
	// Sample Headers:
	// Accept:[*/*]
	// Accept-Encoding:[gzip, br]
	// Cdn-Loop:[cloudflare; loops=1]
	// Cf-Connecting-Ip:[1.2.3.4]
	// Cf-Ipcountry:[IT]
	// Cf-Ray:[93804dd5edd859f5-MXP]
	// Cf-Visitor:[{"scheme":"https"}]
	// User-Agent:[curl/7.88.1]
	// X-Forwarded-For:[2.3.4.5]
	// X-Forwarded-Host:[quote.example.com]
	// X-Forwarded-Port:[80]
	// X-Forwarded-Proto:[http]
	// X-Forwarded-Server:[traefik-32bfd46sce-74c3h]
	// X-Real-Ip:[2.3.4.5]]

	remoteAddr := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")

	// Proxied through Cloudflare?
	if remote := r.Header.Get("Cf-Connecting-Ip"); remote != "" {
		remoteAddr = fmt.Sprintf("%s (%s)", remote, r.Header.Get("Cf-Ipcountry"))
	} else if remote := r.Header.Get("X-Real-Ip"); remote != "" {
		remoteAddr = remote
	} else if remote := r.Header.Get("X-Forwarded-For"); remote != "" {
		remoteAddr = remote
	}

	return fmt.Sprintf("%s %q - %s %s %q", remoteAddr, userAgent, r.Method, r.Proto, r.URL.String())
}
