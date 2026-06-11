package registry

import (
	"testing"
)

func TestRegistryFindTool(t *testing.T) {
	r := &Registry{
		Tools: []Tool{
			{ID: "memory-note-cli", Name: "MemoryNoteCLI"},
			{ID: "pushguard", Name: "PushGuard"},
		},
	}
	if r.FindTool("memory-note-cli") == nil {
		t.Error("expected to find memory-note-cli")
	}
	if r.FindTool("MemoryNoteCLI") == nil {
		t.Error("expected to find by name case-insensitive")
	}
	if r.FindTool("notfound") != nil {
		t.Error("expected nil for unknown tool")
	}
}

func TestRegistrySearchTools(t *testing.T) {
	r := &Registry{
		Tools: []Tool{
			{ID: "memory-note-cli", Name: "MemoryNoteCLI", Description: "Fast terminal note manager", Category: "Productivity", Language: "Python"},
			{ID: "pushguard", Name: "PushGuard", Description: "Git push guard", Category: "Git", Language: "Python"},
		},
	}

	results := r.SearchTools("note")
	if len(results) != 1 || results[0].ID != "memory-note-cli" {
		t.Errorf("expected 1 result for 'note', got %v", results)
	}

	results = r.SearchTools("python")
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'python', got %d", len(results))
	}

	results = r.SearchTools("git")
	if len(results) != 1 || results[0].ID != "pushguard" {
		t.Errorf("expected 1 result for 'git', got %v", results)
	}
}

func TestToolPlatformFor(t *testing.T) {
	tool := Tool{
		ID: "memory-note-cli",
		Platforms: map[string]Platform{
			"linux-amd64": {Type: "archive", URL: "http://example.com/a.tar.gz"},
		},
	}
	p, err := tool.PlatformFor("linux-amd64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type != "archive" {
		t.Errorf("expected type archive, got %s", p.Type)
	}
	_, err = tool.PlatformFor("darwin-arm64")
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
}

func TestPlatformIsPlaceholder(t *testing.T) {
	p := Platform{URL: "TODO", Sha256: "CHANGE_ME"}
	if !p.IsPlaceholderURL() {
		t.Error("expected placeholder URL")
	}
	if !p.IsPlaceholderChecksum() {
		t.Error("expected placeholder checksum")
	}

	p2 := Platform{URL: "https://example.com/file.tar.gz", Sha256: "abc123"}
	if p2.IsPlaceholderURL() {
		t.Error("expected non-placeholder URL")
	}
	if p2.IsPlaceholderChecksum() {
		t.Error("expected non-placeholder checksum")
	}
}

func TestLoadSaveCache(t *testing.T) {
	r := &Registry{
		Version: 1,
		Tools: []Tool{
			{ID: "test-tool", Name: "TestTool"},
		},
	}
	if err := SaveCache(r); err != nil {
		t.Fatalf("save cache failed: %v", err)
	}
	loaded, err := LoadCache()
	if err != nil {
		t.Fatalf("load cache failed: %v", err)
	}
	if loaded.Version != 1 || len(loaded.Tools) != 1 {
		t.Errorf("unexpected loaded cache: %+v", loaded)
	}
}
