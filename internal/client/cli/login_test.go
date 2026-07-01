package cli

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terminal-music-room/music-room/internal/client/config"
	"github.com/terminal-music-room/music-room/internal/server/hub"
)

func TestLoginSuccess(t *testing.T) {
	s := hub.New(hub.Config{ListenAddr: ":0", DataDir: t.TempDir()}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	var out bytes.Buffer
	err := Login(context.Background(), &out, cfgPath, "kaopiz", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Logged in as kaopiz") {
		t.Fatalf("output %q", out.String())
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Nickname != "kaopiz" || cfg.ServerURL != ts.URL || cfg.SessionID == "" {
		t.Fatalf("cfg %+v", cfg)
	}
	if !cfg.LoggedIn() {
		t.Fatal("expected logged in config")
	}
}

func TestLoginInvalidNickname(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	err := Login(context.Background(), io.Discard, cfgPath, "", "http://localhost:8080")
	if err == nil || !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("err %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LoggedIn() {
		t.Fatal("config should not be saved on validation failure")
	}
}

func TestLoginNicknameTooLong(t *testing.T) {
	long := strings.Repeat("a", 33)
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	err := Login(context.Background(), io.Discard, cfgPath, long, "http://localhost:8080")
	if err == nil || !strings.Contains(err.Error(), "1–32 characters") {
		t.Fatalf("err %v", err)
	}
}
