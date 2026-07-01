package youtube

import (
	"context"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

const defaultWorkers = 4

// Pool limits concurrent yt-dlp calls so WebSocket handlers stay non-blocking.
type Pool struct {
	resolver *Resolver
	jobs     chan poolJob
}

type poolJob struct {
	ctx    context.Context
	url    string
	query  string
	result chan poolResult
}

type poolResult struct {
	track protocol.Track
	err   error
}

// NewPool starts worker goroutines that delegate to resolver.
func NewPool(resolver *Resolver, workers int) *Pool {
	if workers <= 0 {
		workers = defaultWorkers
	}
	p := &Pool{
		resolver: resolver,
		jobs:     make(chan poolJob, workers*2),
	}
	for i := 0; i < workers; i++ {
		go p.worker()
	}
	return p
}

// Resolve enqueues work and waits for the result (call from a background goroutine).
func (p *Pool) Resolve(ctx context.Context, rawURL, query string) (protocol.Track, error) {
	resCh := make(chan poolResult, 1)
	job := poolJob{ctx: ctx, url: rawURL, query: query, result: resCh}
	select {
	case p.jobs <- job:
	case <-ctx.Done():
		return protocol.Track{}, ctx.Err()
	}
	select {
	case res := <-resCh:
		return res.track, res.err
	case <-ctx.Done():
		return protocol.Track{}, ctx.Err()
	}
}

func (p *Pool) worker() {
	for job := range p.jobs {
		track, err := p.resolver.Resolve(job.ctx, job.url, job.query)
		job.result <- poolResult{track: track, err: err}
	}
}
