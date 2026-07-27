package config

import "testing"

func TestCUDManagedForcesOfficialRegistry(t *testing.T) {
	t.Setenv("CORE_UTILS_CUD_MANAGED", "1")
	t.Setenv("CORE_UTILS_REGISTRY_URL", "https://evil.example/registry")
	if got := GetRegistryURL(); got != OfficialRegistryURL {
		t.Fatalf("managed registry = %q, want %q", got, OfficialRegistryURL)
	}
}
