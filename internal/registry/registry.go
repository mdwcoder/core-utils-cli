package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mdwcoder/core-utils-cli/internal/config"
)

// Registry represents the full registry JSON.
type Registry struct {
	Version   int    `json:"version"`
	UpdatedAt string `json:"updatedAt"`
	Tools     []Tool `json:"tools"`
}

// Tool represents a single tool in the registry.
type Tool struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Category    string              `json:"category,omitempty"`
	Language    string              `json:"language,omitempty"`
	Version     string              `json:"version,omitempty"`
	Binary      string              `json:"binary,omitempty"`
	Homepage    string              `json:"homepage,omitempty"`
	Repository  string              `json:"repository,omitempty"`
	Platforms   map[string]Platform `json:"platforms,omitempty"`
}

// Platform represents a platform-specific asset.
type Platform struct {
	Type       string `json:"type"`
	URL        string `json:"url"`
	Sha256     string `json:"sha256,omitempty"`
	BinaryPath string `json:"binaryPath,omitempty"`
}

// IsPlaceholderURL returns true if the URL is not a real download link.
func (p Platform) IsPlaceholderURL() bool {
	return p.URL == "" || p.URL == "TODO" || p.URL == "CHANGE_ME"
}

// IsPlaceholderChecksum returns true if the checksum is not a real value.
func (p Platform) IsPlaceholderChecksum() bool {
	return p.Sha256 == "" || p.Sha256 == "CHANGE_ME"
}

// LoadCache reads the local registry cache.
func LoadCache() (*Registry, error) {
	data, err := os.ReadFile(config.CacheFile)
	if err != nil {
		return nil, err
	}
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// SaveCache writes the registry cache to disk.
func SaveCache(r *Registry) error {
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.CacheFile, data, 0644)
}

// FindTool searches by exact id or case-insensitive name match.
func (r *Registry) FindTool(query string) *Tool {
	query = strings.ToLower(query)
	for _, t := range r.Tools {
		if strings.ToLower(t.ID) == query || strings.ToLower(t.Name) == query {
			return &t
		}
	}
	return nil
}

// SearchTools searches across id, name, description, category, and language.
func (r *Registry) SearchTools(query string) []Tool {
	query = strings.ToLower(query)
	var results []Tool
	for _, t := range r.Tools {
		if strings.Contains(strings.ToLower(t.ID), query) ||
			strings.Contains(strings.ToLower(t.Name), query) ||
			strings.Contains(strings.ToLower(t.Description), query) ||
			strings.Contains(strings.ToLower(t.Category), query) ||
			strings.Contains(strings.ToLower(t.Language), query) {
			results = append(results, t)
		}
	}
	return results
}

// PlatformFor returns the asset for a given platform key, or an error.
func (t *Tool) PlatformFor(platformKey string) (Platform, error) {
	if t.Platforms == nil {
		return Platform{}, fmt.Errorf("no platforms available for %s", t.ID)
	}
	p, ok := t.Platforms[platformKey]
	if !ok {
		return Platform{}, fmt.Errorf("platform %s not supported by %s", platformKey, t.ID)
	}
	return p, nil
}
