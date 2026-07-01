package hub

import (
	"os"
	"strings"
)

// Config holds music-roomd runtime settings from the environment.
type Config struct {
	ListenAddr string
	DataDir    string
}

// LoadConfig reads configuration from the environment.
// MUSIC_ROOM_LISTEN defaults to ":8080".
// MUSIC_ROOM_DATA_DIR defaults to "./data/chat" for on-disk chat logs.
func LoadConfig() Config {
	addr := strings.TrimSpace(os.Getenv("MUSIC_ROOM_LISTEN"))
	if addr == "" {
		addr = ":8080"
	}
	dataDir := strings.TrimSpace(os.Getenv("MUSIC_ROOM_DATA_DIR"))
	if dataDir == "" {
		dataDir = "./data/chat"
	}
	return Config{ListenAddr: addr, DataDir: dataDir}
}
