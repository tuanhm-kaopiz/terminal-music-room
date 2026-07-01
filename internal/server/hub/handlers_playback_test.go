package hub

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

type fixedResolver struct {
	track protocol.Track
	err   error
}

func (f fixedResolver) Resolve(_ context.Context, _, _ string) (protocol.Track, error) {
	if f.err != nil {
		return protocol.Track{}, f.err
	}
	return f.track, nil
}

func TestPlaybackPauseResumeSeekSkip(t *testing.T) {
	s := newTestServer(nil)
	s.resolver = fixedResolver{track: protocol.Track{
		VideoID:    "abc123xyz01",
		Title:      "Test Track",
		DurationMs: 120_000,
		SourceURL:  "https://youtube.com/watch?v=abc123xyz01",
	}}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	hostConn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"host"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, hostConn)

	guestConn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"guest"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer guestConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, guestConn)

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "playback-room"})
	if err := hostConn.Write(ctx, websocket.MessageText, create); err != nil {
		t.Fatal(err)
	}
	_, _, err = readEnvelope[protocol.RoomSnapshot](t, hostConn)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	join, _ := protocol.EncodeMessage(protocol.MsgRoomJoin, "j1", protocol.RoomJoinPayload{Slug: "playback-room"})
	if err := guestConn.Write(ctx, websocket.MessageText, join); err != nil {
		t.Fatal(err)
	}
	_, _, err = readEnvelope[protocol.RoomSnapshot](t, guestConn)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)
	_, _, err = readEnvelope[protocol.RoomMemberJoinedPayload](t, hostConn)
	if err != nil {
		t.Fatal(err)
	}

	play, _ := protocol.EncodeMessage(protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{
		URL: "https://youtube.com/watch?v=abc123xyz01",
	})
	if err := hostConn.Write(ctx, websocket.MessageText, play); err != nil {
		t.Fatal(err)
	}
	state, tick := readPlaybackUpdate(t, hostConn)
	if state.Status != protocol.PlaybackPlaying || state.Track == nil {
		t.Fatalf("play state %+v", state)
	}
	if tick.Status != protocol.PlaybackPlaying {
		t.Fatalf("play tick %+v", tick)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	pause, _ := protocol.EncodeMessage(protocol.MsgPlaybackPause, "pa1", protocol.PlaybackPausePayload{})
	if err := guestConn.Write(ctx, websocket.MessageText, pause); err != nil {
		t.Fatal(err)
	}
	state, tick = readPlaybackUpdate(t, hostConn)
	if state.Status != protocol.PlaybackPaused {
		t.Fatalf("pause state %+v", state)
	}
	if tick.Status != protocol.PlaybackPaused {
		t.Fatalf("pause tick %+v", tick)
	}

	resume, _ := protocol.EncodeMessage(protocol.MsgPlaybackResume, "r1", protocol.PlaybackResumePayload{})
	if err := hostConn.Write(ctx, websocket.MessageText, resume); err != nil {
		t.Fatal(err)
	}
	state, tick = readPlaybackUpdate(t, hostConn)
	if state.Status != protocol.PlaybackPlaying {
		t.Fatalf("resume state %+v", state)
	}
	if tick.Status != protocol.PlaybackPlaying {
		t.Fatalf("resume tick %+v", tick)
	}

	seek, _ := protocol.EncodeMessage(protocol.MsgPlaybackSeek, "s1", protocol.PlaybackSeekPayload{PositionMs: 30_000})
	if err := guestConn.Write(ctx, websocket.MessageText, seek); err != nil {
		t.Fatal(err)
	}
	state, tick = readPlaybackUpdate(t, hostConn)
	if state.PositionMs != 30_000 {
		t.Fatalf("seek state %+v", state)
	}
	if tick.PositionMs != 30_000 {
		t.Fatalf("seek tick %+v", tick)
	}

	skip, _ := protocol.EncodeMessage(protocol.MsgPlaybackSkip, "sk1", protocol.PlaybackSkipPayload{})
	if err := hostConn.Write(ctx, websocket.MessageText, skip); err != nil {
		t.Fatal(err)
	}
	state, tick = readPlaybackUpdate(t, hostConn)
	if state.Status != protocol.PlaybackEnded {
		t.Fatalf("skip state %+v", state)
	}
	if tick.Status != protocol.PlaybackEnded {
		t.Fatalf("skip tick %+v", tick)
	}
}

func readPlaybackUpdate(t *testing.T, conn *websocket.Conn) (protocol.PlaybackState, protocol.PlaybackTickPayload) {
	t.Helper()
	_, state, err := readEnvelope[protocol.PlaybackState](t, conn)
	if err != nil {
		t.Fatal(err)
	}
	_, tick, err := readEnvelope[protocol.PlaybackTickPayload](t, conn)
	if err != nil {
		t.Fatal(err)
	}
	return state, tick
}

func TestPlaybackPlayInvalidSource(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"solo"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn)

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "solo-room"})
	_ = conn.Write(ctx, websocket.MessageText, create)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, conn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, conn)

	play, _ := protocol.EncodeMessage(protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{URL: "https://example.com/not-youtube"})
	_ = conn.Write(ctx, websocket.MessageText, play)
	_, errPayload, err := readEnvelope[protocol.ErrorPayload](t, conn)
	if err != nil || errPayload.Code != protocol.ErrInvalidSource {
		t.Fatalf("error %+v err %v", errPayload, err)
	}
}
