// Package server serves the embedded GUI assets and JSON API endpoints used
// by the browser and by the Wails WebView.
package server

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"
)

// Assets must be imported by the package that owns the embedded files.
type Options struct {
	Version  string
	Assets   embed.FS
	Updater  Updater
	AutoStart AutoStarter
}

type Updater interface {
	Check(version string) (string, bool, error)
}

type AutoStarter interface {
	Enable() error
	Disable() error
	Enabled() (bool, error)
}

func Handler(o Options) http.Handler {
	if o.Version == "" {
		o.Version = "dev"
	}
	mux := http.NewServeMux()
	assets := o.Assets
	mux.Handle("/", http.FileServer(http.FS(assets)))
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		jsonResponse(w, map[string]any{
			"version":    o.Version,
			"namedPipe":  false,
			"server":     "running",
		})
	})
	mux.HandleFunc("/api/update/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if o.Updater == nil {
			jsonResponse(w, map[string]any{"upToDate": true, "version": o.Version})
			return
		}
		v, ok, err := o.Updater.Check(o.Version)
		if err != nil {
			jsonError(w, err)
			return
		}
		jsonResponse(w, map[string]any{"upToDate": ok, "version": v})
	})
	mux.HandleFunc("/api/autostart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Enable bool `json:"enable"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, err)
			return
		}
		if o.AutoStart == nil {
			http.Error(w, "autostart not implemented", http.StatusNotImplemented)
			return
		}
		var err error
		if req.Enable {
			err = o.AutoStart.Enable()
		} else {
			err = o.AutoStart.Disable()
		}
		if err != nil {
			jsonError(w, err)
			return
		}
		enabled, _ := o.AutoStart.Enabled()
		jsonResponse(w, map[string]any{"enabled": enabled})
	})
	return mux
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": strings.TrimSpace(err.Error())})
}
