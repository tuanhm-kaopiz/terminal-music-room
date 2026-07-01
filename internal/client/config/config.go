package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultServerURL is the music-roomd base URL when none is configured.
	DefaultServerURL = "http://localhost:8080"
	// EnvServerURL overrides the default server URL.
	EnvServerURL = "MUSIC_ROOM_SERVER_URL"
	// EnvConfigPath overrides the config file location.
	EnvConfigPath = "MUSIC_ROOM_CONFIG"
)

// Config is persisted client identity and server connection settings.
type Config struct {
	Nickname  string `yaml:"nickname"`
	ServerURL string `yaml:"server_url"`
	SessionID string `yaml:"session_id"`
}

// LoggedIn reports whether the client has a saved session (AC-003).
func (c Config) LoggedIn() bool {
	return c.Nickname != "" && c.SessionID != ""
}

// DefaultPath returns ~/.config/music-room/config.yaml.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".config", "music-room", "config.yaml"), nil
}

// ResolvePath returns the config file path from env or the default location.
func ResolvePath() (string, error) {
	if p := strings.TrimSpace(os.Getenv(EnvConfigPath)); p != "" {
		return p, nil
	}
	return DefaultPath()
}

// Load reads config from path. Missing files return an empty config.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// Save writes config to path, creating parent directories as needed.
func Save(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// ResolveServerURL picks server URL from flag, saved config, env, or default.
func ResolveServerURL(flag, saved string) string {
	if s := strings.TrimSpace(flag); s != "" {
		return s
	}
	if s := strings.TrimSpace(saved); s != "" {
		return s
	}
	if s := strings.TrimSpace(os.Getenv(EnvServerURL)); s != "" {
		return s
	}
	return DefaultServerURL
}

// WebSocketURL converts a music-roomd base URL to the /v1/ws endpoint.
func WebSocketURL(serverURL string) (string, error) {
	raw := strings.TrimSpace(serverURL)
	if raw == "" {
		return "", fmt.Errorf("server URL is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %w", err)
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", fmt.Errorf("unsupported URL scheme %q (use http, https, ws, or wss)", u.Scheme)
	}
	basePath := strings.TrimSuffix(u.Path, "/")
	u.Path = basePath + "/v1/ws"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
