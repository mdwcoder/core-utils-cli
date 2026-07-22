package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mdwcoder/core-utils-cli/internal/archive"
	"github.com/mdwcoder/core-utils-cli/internal/checksum"
	"github.com/mdwcoder/core-utils-cli/internal/config"
	"github.com/mdwcoder/core-utils-cli/internal/output"
	"github.com/mdwcoder/core-utils-cli/internal/platform"
	"github.com/mdwcoder/core-utils-cli/internal/registry"
)

// InstalledRecord tracks an installed tool.
type InstalledRecord struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Binary      string    `json:"binary"`
	Platform    string    `json:"platform"`
	InstalledAt time.Time `json:"installed_at"`
}

// InstalledDB is the structure of installed.json.
type InstalledDB struct {
	Tools []InstalledRecord `json:"tools"`
}

func loadInstalled() (*InstalledDB, error) {
	data, err := os.ReadFile(config.InstalledFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &InstalledDB{}, nil
		}
		return nil, err
	}
	var db InstalledDB
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func saveInstalled(db *InstalledDB) error {
	if err := os.MkdirAll(config.BaseDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.InstalledFile, data, 0644)
}

func findInstalled(db *InstalledDB, id string) *InstalledRecord {
	for i := range db.Tools {
		if strings.EqualFold(db.Tools[i].ID, id) {
			return &db.Tools[i]
		}
	}
	return nil
}

func removeInstalled(db *InstalledDB, id string) bool {
	for i, t := range db.Tools {
		if strings.EqualFold(t.ID, id) {
			db.Tools = append(db.Tools[:i], db.Tools[i+1:]...)
			return true
		}
	}
	return false
}

func addOrUpdateInstalled(db *InstalledDB, rec InstalledRecord) {
	for i := range db.Tools {
		if strings.EqualFold(db.Tools[i].ID, rec.ID) {
			db.Tools[i] = rec
			return
		}
	}
	db.Tools = append(db.Tools, rec)
}

// Install installs a tool from the registry.
func Install(tool *registry.Tool, force bool) error {
	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to prepare directories: %w", err)
	}

	platKey := platform.Current()
	plat, err := tool.PlatformFor(platKey)
	if err != nil {
		return err
	}

	if plat.IsPlaceholderURL() {
		return fmt.Errorf("no download URL available for %s on %s (TODO/placeholder)", tool.ID, platKey)
	}

	instDB, err := loadInstalled()
	if err != nil {
		return err
	}
	if existing := findInstalled(instDB, tool.ID); existing != nil && !force {
		return fmt.Errorf("%s is already installed (%s). Use update to change version", tool.ID, existing.Version)
	}

	// Download to temp. The archive extractor picks its format from the file
	// extension, so the temp file must keep the asset's original extension
	// (e.g. .tar.gz, .zip) rather than a generic .tmp suffix.
	assetExt := path.Ext(plat.URL)
	if strings.HasSuffix(strings.ToLower(plat.URL), ".tar.gz") {
		assetExt = ".tar.gz"
	}
	tempFile := filepath.Join(config.BaseDir, fmt.Sprintf(".%s-%s.tmp%s", tool.ID, tool.Version, assetExt))
	output.Printf("Downloading %s for %s...\n", tool.ID, platKey)
	if err := registry.DownloadAsset(plat.URL, tempFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tempFile)

	// Checksum
	if !plat.IsPlaceholderChecksum() {
		if err := checksum.ValidateFile(tempFile, plat.Sha256); err != nil {
			return fmt.Errorf("checksum validation failed: %w", err)
		}
	} else {
		output.StatusWarn("Checksum", "sha256 is CHANGE_ME/placeholder — skipping validation")
	}

	// Extract
	toolDir := filepath.Join(config.ToolsDir, tool.ID, tool.Version)
	if err := os.RemoveAll(toolDir); err != nil {
		return fmt.Errorf("failed to clean install directory: %w", err)
	}
	if err := os.MkdirAll(toolDir, 0755); err != nil {
		return err
	}

	if plat.Type == "archive" {
		if err := archive.Extract(tempFile, toolDir); err != nil {
			os.RemoveAll(toolDir)
			return fmt.Errorf("extraction failed: %w", err)
		}
	} else if plat.Type == "binary" {
		destPath := filepath.Join(toolDir, filepath.Base(plat.BinaryPath))
		data, err := os.ReadFile(tempFile)
		if err != nil {
			os.RemoveAll(toolDir)
			return err
		}
		if err := os.WriteFile(destPath, data, 0755); err != nil {
			os.RemoveAll(toolDir)
			return err
		}
	} else {
		return fmt.Errorf("unsupported asset type: %s", plat.Type)
	}

	// Locate binary
	binPath := filepath.Join(toolDir, plat.BinaryPath)
	if plat.BinaryPath == "" {
		// Try to find by expected binary name
		binPath = filepath.Join(toolDir, tool.Binary)
	}
	if _, err := os.Stat(binPath); err != nil {
		// Fallback: search inside toolDir for the binary name
		found := ""
		filepath.Walk(toolDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.EqualFold(info.Name(), tool.Binary) || strings.EqualFold(info.Name(), tool.Binary+".exe") {
				found = path
				return filepath.SkipAll
			}
			return nil
		})
		if found != "" {
			binPath = found
		} else {
			os.RemoveAll(toolDir)
			return fmt.Errorf("cannot locate binary after extraction (expected %s)", binPath)
		}
	}

	// Ensure executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}

	// Link/copy to bin
	binLink := filepath.Join(config.BinDir, tool.Binary)
	if runtime.GOOS == "windows" && !strings.HasSuffix(binLink, ".exe") {
		binLink += ".exe"
	}
	os.Remove(binLink) // allow overwrite on update

	// Try symlink first, fall back to copy
	if err := os.Symlink(binPath, binLink); err != nil {
		data, err := os.ReadFile(binPath)
		if err != nil {
			return fmt.Errorf("failed to copy binary to bin: %w", err)
		}
		if err := os.WriteFile(binLink, data, 0755); err != nil {
			return fmt.Errorf("failed to copy binary to bin: %w", err)
		}
	}

	// Save record
	rec := InstalledRecord{
		ID:          tool.ID,
		Name:        tool.Name,
		Version:     tool.Version,
		Binary:      tool.Binary,
		Platform:    platKey,
		InstalledAt: time.Now().UTC(),
	}
	addOrUpdateInstalled(instDB, rec)
	if err := saveInstalled(instDB); err != nil {
		return fmt.Errorf("failed to save installed metadata: %w", err)
	}

	output.Printf("Installed %s (%s) -> %s\n", tool.ID, tool.Version, binLink)
	if !isInPath(config.BinDir) {
		output.StatusWarn("PATH", fmt.Sprintf("%s is not in your PATH. Add it to your shell profile.", config.BinDir))
	}
	return nil
}

// Update updates all or one installed tool.
func Update(specificID string) error {
	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to prepare directories: %w", err)
	}

	instDB, err := loadInstalled()
	if err != nil {
		return err
	}
	if len(instDB.Tools) == 0 {
		output.Println("No tools installed.")
		return nil
	}

	// Load registry
	var reg *registry.Registry
	reg, err = registry.Fetch()
	if err != nil {
		output.StatusWarn("Registry", "remote fetch failed, trying cache")
		reg, err = registry.LoadCache()
		if err != nil {
			return fmt.Errorf("cannot load registry: %w", err)
		}
	}

	updated := 0
	for _, rec := range instDB.Tools {
		if specificID != "" && !strings.EqualFold(rec.ID, specificID) {
			continue
		}
		tool := reg.FindTool(rec.ID)
		if tool == nil {
			output.StatusWarn(rec.ID, "not found in registry")
			continue
		}
		if tool.Version == rec.Version {
			output.Printf("%s is already up to date (%s)\n", rec.ID, rec.Version)
			continue
		}
		output.Printf("Updating %s: %s -> %s\n", rec.ID, rec.Version, tool.Version)
		if err := Install(tool, true); err != nil {
			output.Errorf("Update failed for %s: %v\n", rec.ID, err)
			continue
		}
		updated++
	}

	if specificID != "" && updated == 0 {
		return fmt.Errorf("no update performed for %s", specificID)
	}
	output.Printf("Updated %d tool(s).\n", updated)
	return nil
}

// Remove uninstalls a tool.
func Remove(id string) error {
	instDB, err := loadInstalled()
	if err != nil {
		return err
	}
	rec := findInstalled(instDB, id)
	if rec == nil {
		return fmt.Errorf("tool %s is not installed", id)
	}

	// Remove tool directory
	toolDir := filepath.Join(config.ToolsDir, rec.ID)
	if err := os.RemoveAll(toolDir); err != nil {
		return fmt.Errorf("failed to remove tool directory: %w", err)
	}

	// Remove binary link
	binLink := filepath.Join(config.BinDir, rec.Binary)
	if runtime.GOOS == "windows" && !strings.HasSuffix(binLink, ".exe") {
		binLink += ".exe"
	}
	os.Remove(binLink)

	// Update installed.json
	removeInstalled(instDB, rec.ID)
	if err := saveInstalled(instDB); err != nil {
		return fmt.Errorf("failed to save installed metadata: %w", err)
	}

	output.Printf("Removed %s\n", rec.ID)
	return nil
}

// List prints installed tools.
func List() error {
	instDB, err := loadInstalled()
	if err != nil {
		return err
	}
	if len(instDB.Tools) == 0 {
		output.Println("No tools installed.")
		return nil
	}
	output.Println(output.Bold("Installed tools:"))
	for _, t := range instDB.Tools {
		output.Printf("  %-20s %s (%s) %s\n", t.ID, t.Name, t.Version, t.InstalledAt.Format("2006-01-02"))
	}
	return nil
}

func isInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	sep := string(filepath.ListSeparator)
	for _, p := range strings.Split(pathEnv, sep) {
		if strings.TrimSpace(p) == dir {
			return true
		}
	}
	return false
}

// IsInPath reports whether dir is present in the PATH environment variable.
func IsInPath(dir string) bool {
	return isInPath(dir)
}

// LoadInstalled returns the installed tools database.
func LoadInstalled() (*InstalledDB, error) {
	return loadInstalled()
}
