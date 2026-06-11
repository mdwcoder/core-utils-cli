package platform

import (
	"runtime"
	"testing"
)

func TestCurrent(t *testing.T) {
	p := Current()
	if p == "" {
		t.Error("expected non-empty platform")
	}
	expected := runtime.GOOS + "-" + runtime.GOARCH
	if p != expected {
		t.Errorf("expected %s, got %s", expected, p)
	}
}

func TestValidate(t *testing.T) {
	if !Validate("linux-amd64") {
		t.Error("expected linux-amd64 to be valid")
	}
	if Validate("") {
		t.Error("expected empty string to be invalid")
	}
}
