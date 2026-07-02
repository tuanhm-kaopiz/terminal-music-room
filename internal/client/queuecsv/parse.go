package queuecsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// ParseURLs reads a CSV file with a single url column (header required: "url").
func ParseURLs(r io.Reader) ([]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM

	records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("csv file is empty")
	}

	header := strings.ToLower(strings.TrimSpace(records[0][0]))
	if header != "url" {
		return nil, fmt.Errorf(`csv header must be "url", got %q`, records[0][0])
	}

	var urls []string
	for i, row := range records[1:] {
		line := strings.TrimSpace(firstCell(row))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := validateYouTubeURL(line); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+2, err)
		}
		urls = append(urls, line)
	}
	if len(urls) == 0 {
		return nil, fmt.Errorf("no urls found in csv")
	}
	return urls, nil
}

func firstCell(row []string) string {
	if len(row) == 0 {
		return ""
	}
	return row[0]
}

func validateYouTubeURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty url")
	}
	lower := strings.ToLower(raw)
	if !strings.Contains(lower, "youtube.com") && !strings.Contains(lower, "youtu.be") {
		return fmt.Errorf("not a YouTube URL: %q", raw)
	}
	return nil
}
