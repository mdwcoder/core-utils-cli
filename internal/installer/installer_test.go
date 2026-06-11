package installer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mdwcoder/core-utils-cli/internal/config"
	"github.com/mdwcoder/core-utils-cli/internal/registry"
)

func TestLoadSaveInstalled(t *testing.T) {
	oldBase := config.BaseDir
	config.BaseDir = t.TempDir()
	config.InstalledFile = filepath.Join(config.BaseDir, "installed.json")
	defer func() {
		config.BaseDir = oldBase
		config.InstalledFile = filepath.Join(oldBase, "installed.json")
	}()

	db, err := loadInstalled()
	if err != nil {
		t.Fatal(err)
	}
	if len(db.Tools) != 0 {
		t.Error("expected empty db")
	}

	rec := InstalledRecord{ID: "test", Name: "Test", Version: "1.0.0", Binary: "test", InstalledAt: time.Now().UTC()}
	addOrUpdateInstalled(db, rec)
	if err := saveInstalled(db); err != nil {
		t.Fatal(err)
	}

	db2, err := loadInstalled()
	if err != nil {
		t.Fatal(err)
	}
	if len(db2.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(db2.Tools))
	}
	if db2.Tools[0].ID != "test" {
		t.Errorf("expected test, got %s", db2.Tools[0].ID)
	}
}

func TestFindRemoveInstalled(t *testing.T) {
	db := &InstalledDB{
		Tools: []InstalledRecord{
			{ID: "a", Version: "1.0.0"},
			{ID: "b", Version: "2.0.0"},
		},
	}
	if findInstalled(db, "a") == nil {
		t.Error("expected to find a")
	}
	if !removeInstalled(db, "a") {
		t.Error("expected remove to succeed")
	}
	if len(db.Tools) != 1 {
		t.Errorf("expected 1 tool after removal, got %d", len(db.Tools))
	}
	if removeInstalled(db, "c") {
		t.Error("expected remove to fail for unknown tool")
	}
}

func TestIsInPath(t *testing.T) {
	if !isInPath("/usr/bin") {
		// PATH may not contain /usr/bin on all test envs; skip assertion if not present
		for _, p := range filepath.SplitList(os.Getenv("PATH")) {
			if isInPath(p) {
				return
			}
		}
		t.Error("expected isInPath to find at least one directory in PATH")
	}
}

func TestInstallPlaceholder(t *testing.T) {
	oldBase := config.BaseDir
	config.BaseDir = t.TempDir()
	config.BinDir = filepath.Join(config.BaseDir, "bin")
	config.ToolsDir = filepath.Join(config.BaseDir, "tools")
	config.InstalledFile = filepath.Join(config.BaseDir, "installed.json")
	defer func() {
		config.BaseDir = oldBase
		config.BinDir = filepath.Join(oldBase, "bin")
		config.ToolsDir = filepath.Join(oldBase, "tools")
		config.InstalledFile = filepath.Join(oldBase, "installed.json")
	}()

	config.EnsureDirs()
	tool := &registry.Tool{
		ID:      "test-tool",
		Name:    "TestTool",
		Version: "1.0.0",
		Binary:  "test-tool",
		Platforms: map[string]registry.Platform{
			"linux-amd64": {
				Type:   "archive",
				URL:    "TODO",
				Sha256: "CHANGE_ME",
			},
		},
	}
	err := Install(tool, false)
	if err == nil {
		t.Error("expected error for placeholder URL")
	}
}
