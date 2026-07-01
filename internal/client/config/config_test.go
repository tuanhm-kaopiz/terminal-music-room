package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	want := Config{
		Nickname:  "kaopiz",
		ServerURL: "http://localhost:8080",
		SessionID: "abc123",
	}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg != (Config{}) {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoggedIn(t *testing.T) {
	if (Config{Nickname: "a", SessionID: "s"}).LoggedIn() != true {
		t.Fatal("expected logged in")
	}
	if (Config{Nickname: "a"}).LoggedIn() {
		t.Fatal("expected not logged in without session")
	}
}

func TestWebSocketURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"http://localhost:8080", "ws://localhost:8080/v1/ws"},
		{"https://music.example.com", "wss://music.example.com/v1/ws"},
		{"http://127.0.0.1:18080/", "ws://127.0.0.1:18080/v1/ws"},
	}
	for _, tc := range cases {
		got, err := WebSocketURL(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("%q -> %q err %v", tc.in, got, err)
		}
	}
}

func TestResolveServerURL(t *testing.T) {
	t.Setenv(EnvServerURL, "http://env.example.com")
	if got := ResolveServerURL("http://flag.example.com", "http://saved.example.com"); got != "http://flag.example.com" {
		t.Fatalf("flag: %s", got)
	}
	if got := ResolveServerURL("", "http://saved.example.com"); got != "http://saved.example.com" {
		t.Fatalf("saved: %s", got)
	}
	if got := ResolveServerURL("", ""); got != "http://env.example.com" {
		t.Fatalf("env: %s", got)
	}
}
