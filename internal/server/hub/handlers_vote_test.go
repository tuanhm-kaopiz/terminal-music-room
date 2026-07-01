package hub

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestVoteSkipPasses(t *testing.T) {
	s := newTestServer(nil)
	s.resolver = fixedResolver{track: protocol.Track{
		VideoID:    "votevid0001",
		Title:      "Vote Track",
		DurationMs: 120_000,
		SourceURL:  "https://youtube.com/watch?v=votevid0001",
	}}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	hostConn := dialWithNick(t, ctx, wsURL, "host")
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, hostConn)

	guestConn := dialWithNick(t, ctx, wsURL, "guest")
	defer guestConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, guestConn)

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "vote-room"})
	_ = hostConn.Write(ctx, websocket.MessageText, create)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, hostConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	join, _ := protocol.EncodeMessage(protocol.MsgRoomJoin, "j1", protocol.RoomJoinPayload{Slug: "vote-room"})
	_ = guestConn.Write(ctx, websocket.MessageText, join)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	play, _ := protocol.EncodeMessage(protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{
		URL: "https://youtube.com/watch?v=votevid0001",
	})
	_ = hostConn.Write(ctx, websocket.MessageText, play)
	_, _ = readPlaybackUpdate(t, hostConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	skipVote, _ := protocol.EncodeMessage(protocol.MsgVoteSkip, "v1", protocol.VoteSkipPayload{})
	_ = hostConn.Write(ctx, websocket.MessageText, skipVote)
	updated := waitForVoteUpdate(t, hostConn, 1)
	if updated.Progress == nil || updated.Progress.Votes != 1 {
		t.Fatalf("updated %+v", updated)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	_ = guestConn.Write(ctx, websocket.MessageText, skipVote)
	_, _ = readPlaybackUpdate(t, hostConn)
}

func waitForVoteUpdate(t *testing.T, conn *websocket.Conn, wantVotes int) protocol.VoteUpdatedPayload {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, data, err := conn.Read(ctx)
		cancel()
		if err != nil {
			continue
		}
		env, err := protocol.Decode(data)
		if err != nil || env.Type != protocol.MsgVoteUpdated {
			continue
		}
		var payload protocol.VoteUpdatedPayload
		if err := env.UnmarshalPayload(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Progress != nil && payload.Progress.Votes == wantVotes {
			return payload
		}
	}
	t.Fatalf("timeout waiting for vote.updated with %d votes", wantVotes)
	return protocol.VoteUpdatedPayload{}
}
