package panels

import (
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

// FixtureView returns a rich room snapshot for render tests.
func FixtureView() state.View {
	return state.View{
		Status:      state.StatusConnected,
		SessionID:   "sess-host",
		DisplayName: "host#1",
		InRoom:      true,
		Room: protocol.RoomSnapshot{
			Slug:   "backend-team",
			HostID: "sess-host",
			Members: []protocol.Member{
				{SessionID: "sess-host", Nickname: "host", DisplayName: "host#1", IsHost: true},
				{SessionID: "sess-guest", Nickname: "guest", DisplayName: "guest#2"},
			},
			Playback: protocol.PlaybackState{
				Status:     protocol.PlaybackPlaying,
				PositionMs: 60000,
				DurationMs: 180000,
				Track:      &protocol.Track{Title: "Neon Nights", DurationMs: 180000},
			},
			Queue: []protocol.QueueItem{
				{ID: "q1", Title: "Track Two", AddedBy: "guest#2"},
				{ID: "q2", Title: "Track Three", AddedBy: "host#1"},
			},
			Chat: []protocol.ChatMessage{
				{Kind: protocol.ChatKindUser, Author: "guest#2", Body: "hello"},
			},
			Vote: &protocol.Vote{
				Kind:      protocol.VoteKindSkip,
				Voters:    []string{"sess-guest"},
				Threshold: 2,
			},
			Reactions: map[string]int{"🔥": 1},
		},
	}
}
