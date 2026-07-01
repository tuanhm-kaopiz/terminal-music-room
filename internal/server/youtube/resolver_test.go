package youtube

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type mockRunner struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (m mockRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	return m.fn(ctx, name, args...)
}

func sampleJSON(id, title string, duration float64) []byte {
	return []byte(`{"id":"` + id + `","title":"` + title + `","duration":` + strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll("", "", ""), "", "")) + `}`)
}

func TestResolveURLMetadata(t *testing.T) {
	payload := []byte(`{"id":"abc123xyz01","title":"Lofi Mix","duration":120.5,"webpage_url":"https://www.youtube.com/watch?v=abc123xyz01"}`)
	r := NewResolver(Config{
		Runner: mockRunner{fn: func(_ context.Context, name string, args ...string) ([]byte, error) {
			if name != "yt-dlp" {
				t.Fatalf("bin %q", name)
			}
			joined := strings.Join(args, " ")
			if !strings.Contains(joined, "abc123xyz01") {
				t.Fatalf("args %v", args)
			}
			return payload, nil
		}},
	})

	track, err := r.Resolve(context.Background(), "https://youtube.com/watch?v=abc123xyz01", "")
	if err != nil {
		t.Fatal(err)
	}
	if track.VideoID != "abc123xyz01" || track.Title != "Lofi Mix" || track.DurationMs != 120_500 {
		t.Fatalf("track %+v", track)
	}
}

func TestResolveSearchFirstResult(t *testing.T) {
	payload := []byte(`{"id":"search00123","title":"Search Hit","duration":90,"webpage_url":"https://www.youtube.com/watch?v=search00123"}`)
	r := NewResolver(Config{
		Runner: mockRunner{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
			joined := strings.Join(args, " ")
			if !strings.Contains(joined, "ytsearch5:lofi beats") || !strings.Contains(joined, "-I") {
				t.Fatalf("args %v", args)
			}
			return payload, nil
		}},
	})

	track, err := r.Resolve(context.Background(), "", "lofi beats")
	if err != nil {
		t.Fatal(err)
	}
	if track.VideoID != "search00123" {
		t.Fatalf("track %+v", track)
	}
}

func TestResolveSearchCache(t *testing.T) {
	var calls atomic.Int32
	payload := []byte(`{"id":"cached12345","title":"Cached","duration":60}`)
	r := NewResolver(Config{
		SearchCache: time.Minute,
		Runner: mockRunner{fn: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
			calls.Add(1)
			return payload, nil
		}},
	})

	if _, err := r.Resolve(context.Background(), "", "same query"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(context.Background(), "", "same query"); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatalf("yt-dlp calls = %d want 1", calls.Load())
	}
}

func TestResolveInvalidURL(t *testing.T) {
	r := NewResolver(Config{})
	_, err := r.Resolve(context.Background(), "https://example.com/video", "")
	if err != ErrInvalidSource {
		t.Fatalf("got %v", err)
	}
}

func TestResolveYTDLPFailure(t *testing.T) {
	r := NewResolver(Config{
		Runner: mockRunner{fn: func(context.Context, string, ...string) ([]byte, error) {
			return nil, context.DeadlineExceeded
		}},
	})
	_, err := r.Resolve(context.Background(), "https://youtube.com/watch?v=dQw4w9WgXcQ", "")
	if err == nil || !strings.Contains(err.Error(), "source unavailable") {
		t.Fatalf("got %v", err)
	}
}

func TestResolveBothURLAndQueryRejected(t *testing.T) {
	r := NewResolver(Config{})
	_, err := r.Resolve(context.Background(), "https://youtube.com/watch?v=dQw4w9WgXcQ", "lofi")
	if err != ErrInvalidSource {
		t.Fatalf("got %v", err)
	}
}

func TestPoolLimitsConcurrentJobs(t *testing.T) {
	block := make(chan struct{})
	var active atomic.Int32
	var peak atomic.Int32
	r := NewResolver(Config{
		Runner: mockRunner{fn: func(context.Context, string, ...string) ([]byte, error) {
			cur := active.Add(1)
			for {
				old := peak.Load()
				if cur <= old || peak.CompareAndSwap(old, cur) {
					break
				}
			}
			<-block
			active.Add(-1)
			return []byte(`{"id":"pool1234567","title":"Pool","duration":30}`), nil
		}},
	})
	pool := NewPool(r, 2)

	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		go func() {
			_, _ = pool.Resolve(context.Background(), "", "q"+string(rune('a'+i)))
			done <- struct{}{}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	if peak.Load() > 2 {
		t.Fatalf("peak concurrency %d want <= 2", peak.Load())
	}
	close(block)
	for i := 0; i < 4; i++ {
		<-done
	}
}
