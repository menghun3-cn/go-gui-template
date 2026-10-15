//go:build !nogui

package gui

import (
	"net"
	"net/http"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/menghun3-cn/go-gui-template/internal/server"
)

// Run starts the Wails WebView window pointing at the local embedded server.
func Run(opts Options) error {
	handler := server.Handler(server.Options{
		Version:   opts.Version,
		Assets:    Assets,
		Updater:   opts.Updater,
		AutoStart: opts.AutoStart,
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: handler}
	go srv.Serve(ln) //nolint:errcheck
	url := "http://" + ln.Addr().String()

	app := application.New(application.Options{
		Name:        "go-gui-template",
		Description: "Single-exe Go GUI template",
		Assets: application.AssetOptions{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, url+r.URL.Path, http.StatusTemporaryRedirect)
			}),
		},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Go GUI Template",
		Width:  900,
		Height: 680,
		URL:    url,
	})
	return app.Run()
}

type Options struct {
	Version   string
	Updater   server.Updater
	AutoStart server.AutoStarter
}
