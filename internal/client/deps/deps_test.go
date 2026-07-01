package deps

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestCheckAllPresent(t *testing.T) {
	lookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	t.Cleanup(func() { lookPath = exec.LookPath })

	r := Check()
	if len(r.Missing) != 0 {
		t.Fatalf("expected none missing, got %v", r.Missing)
	}
	if err := FormatError(r); err != nil {
		t.Fatalf("FormatError: %v", err)
	}
}

func TestCheckMissingMPV(t *testing.T) {
	lookPath = func(name string) (string, error) {
		if name == "mpv" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + name, nil
	}
	t.Cleanup(func() { lookPath = exec.LookPath })

	r := Check()
	if len(r.Missing) != 1 || r.Missing[0] != "mpv" {
		t.Fatalf("missing: %v", r.Missing)
	}
	err := FormatError(r)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "mpv") {
		t.Fatalf("message: %q", msg)
	}
	if !strings.Contains(msg, "install") {
		t.Fatalf("expected install hint in %q", msg)
	}
}

func TestEnsurePlayback(t *testing.T) {
	lookPath = func(name string) (string, error) {
		return "", errors.New("not found")
	}
	t.Cleanup(func() { lookPath = exec.LookPath })

	if err := EnsurePlayback(); err == nil {
		t.Fatal("expected error")
	}
}
