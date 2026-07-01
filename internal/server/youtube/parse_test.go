package youtube

import (
	"testing"
)

func TestParseVideoID(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ", true},
		{"https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ", true},
		{"https://youtube.com/shorts/dQw4w9WgXcQ", "dQw4w9WgXcQ", true},
		{"dQw4w9WgXcQ", "dQw4w9WgXcQ", true},
		{"https://example.com/watch?v=dQw4w9WgXcQ", "", false},
		{"not-a-url", "", false},
	}
	for _, tc := range cases {
		got, ok := ParseVideoID(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("ParseVideoID(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
