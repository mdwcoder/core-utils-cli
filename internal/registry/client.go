package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mdwcoder/core-utils-cli/internal/config"
)

// Fetch downloads the registry from the backend and returns it.
func Fetch() (*Registry, error) {
	url := config.GetRegistryURL()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "core-utils-cli/"+config.AppVersion)

	client := &http.Client{Timeout: 10 * time.Second}
	if config.CUDManaged() {
		client.CheckRedirect = rejectUntrustedRedirect
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("registry returned %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("invalid registry JSON: %w", err)
	}
	return &r, nil
}

// DownloadAsset downloads a platform asset to a temporary file path.
func DownloadAsset(url string, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "core-utils-cli/"+config.AppVersion)

	client := &http.Client{Timeout: 60 * time.Second}
	if config.CUDManaged() {
		if !allowedCUDAssetURL(req.URL) {
			return fmt.Errorf("CUD-managed downloads require an approved HTTPS release host")
		}
		client.CheckRedirect = rejectUntrustedRedirect
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download returned %d: %s", resp.StatusCode, string(body))
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return out.Close()
}

func rejectUntrustedRedirect(req *http.Request, _ []*http.Request) error {
	if config.CUDManaged() && !allowedCUDAssetURL(req.URL) && req.URL.String() != config.OfficialRegistryURL {
		return fmt.Errorf("redirect outside approved CoreUtils sources")
	}
	return nil
}

func allowedCUDAssetURL(value *url.URL) bool {
	if value == nil || value.Scheme != "https" || value.User != nil {
		return false
	}
	switch value.Hostname() {
	case "api.core-utils.dev", "github.com", "objects.githubusercontent.com", "release-assets.githubusercontent.com":
		return true
	default:
		return false
	}
}
