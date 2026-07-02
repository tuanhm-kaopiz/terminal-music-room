package queuecsv

import (
	"strings"
	"testing"
)

func TestParseURLs(t *testing.T) {
	csv := `url
https://www.youtube.com/watch?v=dQw4w9WgXcQ
https://youtu.be/abc123xyz01
`
	urls, err := ParseURLs(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Fatalf("got %d urls", len(urls))
	}
}

func TestParseURLsSkipsBlankAndComments(t *testing.T) {
	csv := `url
https://www.youtube.com/watch?v=aaaaaaaaaaa

# skipped
https://youtu.be/bbbbbbbbbbb
`
	urls, err := ParseURLs(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Fatalf("got %v", urls)
	}
}

func TestParseURLsRejectsBadHeader(t *testing.T) {
	_, err := ParseURLs(strings.NewReader("link\nhttps://youtu.be/x\n"))
	if err == nil {
		t.Fatal("expected header error")
	}
}

func TestParseURLsRejectsNonYouTube(t *testing.T) {
	csv := `url
https://example.com/watch?v=abc123xyz01
`
	_, err := ParseURLs(strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestParseURLsEmpty(t *testing.T) {
	_, err := ParseURLs(strings.NewReader("url\n"))
	if err == nil {
		t.Fatal("expected error for no urls")
	}
}
