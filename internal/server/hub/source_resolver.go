package hub

import (
	"context"

	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/youtube"
)

// SourceResolver resolves play/queue sources to tracks.
type SourceResolver interface {
	Resolve(ctx context.Context, rawURL, query string) (protocol.Track, error)
}

type poolResolver struct {
	pool *youtube.Pool
}

func (p poolResolver) Resolve(ctx context.Context, rawURL, query string) (protocol.Track, error) {
	return p.pool.Resolve(ctx, rawURL, query)
}

func newDefaultResolver() SourceResolver {
	yt := youtube.NewResolver(youtube.Config{})
	return poolResolver{pool: youtube.NewPool(yt, 0)}
}
