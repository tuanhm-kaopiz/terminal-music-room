package hub

import (
	"os"
	"strings"
)

// Config holds music-roomd runtime settings from the environment.
type Config struct {
	ListenAddr      string
	DataDir         string
	QueueHistoryDir string
}

// LoadConfig reads configuration from the environment.
// MUSIC_ROOM_LISTEN defaults to ":8080".
// MUSIC_ROOM_DATA_DIR defaults to "./data/chat" for on-disk chat logs.
// MUSIC_ROOM_QUEUE_HISTORY_DIR defaults to "./data/queue" for queue URL history.
func LoadConfig() Config {
	addr := strings.TrimSpace(os.Getenv("MUSIC_ROOM_LISTEN"))
	if addr == "" {
		addr = ":8080"
	}
	dataDir := strings.TrimSpace(os.Getenv("MUSIC_ROOM_DATA_DIR"))
	if dataDir == "" {
		dataDir = "./data/chat"
	}
	queueHistoryDir := strings.TrimSpace(os.Getenv("MUSIC_ROOM_QUEUE_HISTORY_DIR"))
	if queueHistoryDir == "" {
		queueHistoryDir = "./data/queue"
	}
	return Config{ListenAddr: addr, DataDir: dataDir, QueueHistoryDir: queueHistoryDir}
}
