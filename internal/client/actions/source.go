package actions

import (
	"fmt"
	"strings"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

// ParseSourceArgs classifies REPL-style args as a YouTube URL or search query.
func ParseSourceArgs(args []string) (url, query string, err error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("provide a YouTube URL or search query")
	}
	joined := strings.Join(args, " ")
	return ParseSource(joined)
}

// ParseSource classifies a single string as a YouTube URL or search query.
func ParseSource(input string) (url, query string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("provide a YouTube URL or search query")
	}
	if strings.Contains(input, "youtube.com") || strings.Contains(input, "youtu.be") {
		return input, "", nil
	}
	return "", input, nil
}

// PlaybackPlayPayload builds a play payload from explicit url/query flags.
func PlaybackPlayPayload(url, query string) (protocol.PlaybackPlayPayload, error) {
	url = strings.TrimSpace(url)
	query = strings.TrimSpace(query)
	switch {
	case url != "" && query != "":
		return protocol.PlaybackPlayPayload{}, fmt.Errorf("use either --url or --query")
	case url != "":
		return protocol.PlaybackPlayPayload{URL: url}, nil
	case query != "":
		return protocol.PlaybackPlayPayload{Query: query}, nil
	default:
		return protocol.PlaybackPlayPayload{}, fmt.Errorf("provide --url or --query")
	}
}

// QueueAddPayload builds a queue-add payload from explicit url/query flags.
func QueueAddPayload(url, query string) (protocol.QueueAddPayload, error) {
	url = strings.TrimSpace(url)
	query = strings.TrimSpace(query)
	switch {
	case url != "" && query != "":
		return protocol.QueueAddPayload{}, fmt.Errorf("use either --url or --query")
	case url != "":
		return protocol.QueueAddPayload{URL: url}, nil
	case query != "":
		return protocol.QueueAddPayload{Query: query}, nil
	default:
		return protocol.QueueAddPayload{}, fmt.Errorf("provide --url or --query")
	}
}

func playPayloadFromSource(urlOrQuery string) (protocol.PlaybackPlayPayload, error) {
	url, query, err := ParseSource(urlOrQuery)
	if err != nil {
		return protocol.PlaybackPlayPayload{}, err
	}
	if url != "" {
		return protocol.PlaybackPlayPayload{URL: url}, nil
	}
	return protocol.PlaybackPlayPayload{Query: query}, nil
}

func queueAddPayloadFromSource(urlOrQuery string) (protocol.QueueAddPayload, error) {
	url, query, err := ParseSource(urlOrQuery)
	if err != nil {
		return protocol.QueueAddPayload{}, err
	}
	if url != "" {
		return protocol.QueueAddPayload{URL: url}, nil
	}
	return protocol.QueueAddPayload{Query: query}, nil
}
