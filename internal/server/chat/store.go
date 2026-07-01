package chat

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// Store persists chat as compact tab-separated lines (low disk use, no DB).
// Format: RFC3339Nano \t kind \t author \t body
type Store struct {
	path string
	mu   sync.Mutex
}

// NewStore creates an append-only log at {dataDir}/{slug}.chat.log.
func NewStore(dataDir, slug string) *Store {
	return &Store{path: filepath.Join(dataDir, slug+".chat.log")}
}

// Path returns the on-disk log path (for tests).
func (s *Store) Path() string {
	return s.path
}

// Append writes one message line to the log file.
func (s *Store) Append(msg protocol.ChatMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, encodeLine(msg))
	return err
}

// LoadRecent reads up to n most recent messages from the log.
func (s *Store) LoadRecent(n int) ([]protocol.ChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	msgs := make([]protocol.ChatMessage, 0, len(lines))
	for _, line := range lines {
		msg, err := decodeLine(line)
		if err != nil {
			continue
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func encodeLine(msg protocol.ChatMessage) string {
	at := msg.At.UTC().Format(time.RFC3339Nano)
	kind := string(msg.Kind)
	author := sanitizeField(msg.Author)
	body := sanitizeField(msg.Body)
	return strings.Join([]string{at, kind, author, body}, "\t")
}

func decodeLine(line string) (protocol.ChatMessage, error) {
	parts := strings.SplitN(line, "\t", 4)
	if len(parts) != 4 {
		return protocol.ChatMessage{}, fmt.Errorf("invalid chat log line")
	}
	at, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return protocol.ChatMessage{}, err
	}
	kind := protocol.ChatKind(parts[1])
	msg := protocol.ChatMessage{
		Kind: kind,
		At:   at,
	}
	if parts[2] != "" {
		msg.Author = parts[2]
	}
	msg.Body = parts[3]
	msg.ID = newMessageID()
	return msg, nil
}

func sanitizeField(s string) string {
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
