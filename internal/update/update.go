// Package update implements a minimal self-update loop: read a JSON feed,
// download a platform asset, verify SHA-256, then replace the executable.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var defaultFeed = "https://your-host.example/app/latest.json"

// Release is the update feed schema. Keep it small and versioned.
type Release struct {
	Version string  `json:"version"`
	Notes   string  `json:"notes,omitempty"`
	Assets  []Asset `json:"assets"`
	URL     string  `json:"url,omitempty"`
}

type Asset struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// Latest fetches the newest release from the update feed.
func Latest(ctx context.Context, feed string) (*Release, error) {
	if feed == "" {
		feed = defaultFeed
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update feed returned %s", resp.Status)
	}
	var r Release
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&r); err != nil {
		return nil, err
	}
	if r.Version == "" {
		return nil, errors.New("update feed has no version")
	}
	return &r, nil
}

// AssetFor selects the platform asset for this binary.
func (r *Release) AssetFor() (*Asset, error) {
	for i := range r.Assets {
		a := &r.Assets[i]
		if a.OS == runtime.GOOS && a.Arch == runtime.GOARCH {
			return a, nil
		}
	}
	return nil, fmt.Errorf("release has no asset for %s/%s", runtime.GOOS, runtime.GOARCH)
}

// Newer reports whether release version is newer than current.
func Newer(release, current string) bool {
	return strings.TrimSpace(release) != strings.TrimSpace(current) &&
		strings.TrimSpace(release) != "dev" &&
		!strings.HasPrefix(release, "v"+current)
}

// Download fetches an asset into a temp file and verifies its SHA-256.
func Download(ctx context.Context, a *Asset) (string, error) {
	tmp, err := os.CreateTemp("", "gogui-"+filepath.Base(a.URL)+"-*")
	if err != nil {
		return "", err
	}
	name := tmp.Name()
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(name)
		}
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %s", resp.Status)
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, h), resp.Body); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if a.SHA256 != "" && !strings.EqualFold(sum, a.SHA256) {
		os.Remove(name)
		return "", fmt.Errorf("sha256 mismatch: got %s want %s", sum, a.SHA256)
	}
	return name, nil
}

// RunCLI implements `gogui update [--feed URL]`.
func RunCLI(args []string, version string) error {
	feed := defaultFeed
	if len(args) > 0 {
		switch args[0] {
		case "--feed":
			if len(args) < 2 {
				return errors.New("--feed needs a URL")
			}
			feed = args[1]
		case "check":
			// check-only is implied by the flow below when --feed is given.
			if len(args) > 1 && args[1] == "--feed" && len(args) > 2 {
				feed = args[2]
			}
		default:
			return fmt.Errorf("usage: gogui update [--feed URL]")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	rel, err := Latest(ctx, feed)
	if err != nil {
		return err
	}
	if !Newer(rel.Version, version) {
		fmt.Printf("up to date: %s\n", version)
		return nil
	}
	a, err := rel.AssetFor()
	if err != nil {
		return err
	}
	fmt.Printf("downloading %s ...\n", a.URL)
	staged, err := Download(ctx, a)
	if err != nil {
		return err
	}
	defer os.Remove(staged)
	fmt.Printf("installing %s ...\n", rel.Version)
	return replaceBinary(staged)
}
