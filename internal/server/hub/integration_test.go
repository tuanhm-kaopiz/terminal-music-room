package hub

import (
	"context"
	"fmt"
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func TestIntegrationCreateJoinPlayPauseChat(t *testing.T) {
	s := newTestServer(nil)
	s.resolver = fixedResolver{track: protocol.Track{
		VideoID:    "abc123xyz01",
		Title:      "Integration Track",
		DurationMs: 180_000,
		SourceURL:  "https://youtube.com/watch?v=abc123xyz01",
	}}
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	hostConn := dialWithNick(t, ctx, wsURL, "host")
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	hostAck := readSessionAck(t, hostConn)

	guestConn := dialWithNick(t, ctx, wsURL, "guest")
	defer guestConn.Close(websocket.StatusNormalClosure, "done")
	guestAck := readSessionAck(t, guestConn)

	snap := createRoom(t, ctx, hostConn, "integration-flow")
	if snap.Slug != "integration-flow" || len(snap.Members) != 1 {
		t.Fatalf("create snap %+v", snap)
	}

	guestSnap := joinRoom(t, ctx, guestConn, hostConn, "integration-flow")
	if len(guestSnap.Members) != 2 {
		t.Fatalf("join snap %+v", guestSnap)
	}

	writeMsg(t, ctx, hostConn, protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{
		URL: "https://youtube.com/watch?v=abc123xyz01",
	})
	state, tick := readPlaybackUpdate(t, hostConn)
	if state.Status != protocol.PlaybackPlaying || state.Track == nil || state.Track.Title != "Integration Track" {
		t.Fatalf("play state %+v", state)
	}
	if tick.Status != protocol.PlaybackPlaying {
		t.Fatalf("play tick %+v", tick)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)
	readPlaybackUpdate(t, guestConn)

	writeMsg(t, ctx, guestConn, protocol.MsgPlaybackPause, "pa1", protocol.PlaybackPausePayload{})
	state, tick = readPlaybackUpdate(t, hostConn)
	if state.Status != protocol.PlaybackPaused {
		t.Fatalf("pause state %+v", state)
	}
	if tick.Status != protocol.PlaybackPaused {
		t.Fatalf("pause tick %+v", tick)
	}
	readPlaybackUpdate(t, guestConn)

	writeMsg(t, ctx, guestConn, protocol.MsgChatSend, "m1", protocol.ChatSendPayload{Body: "hello room"})
	guestMsg := readUserChat(t, guestConn)
	if guestMsg.Body != "hello room" || guestMsg.Author != "guest" {
		t.Fatalf("guest chat %+v", guestMsg)
	}
	hostMsg := readUserChat(t, hostConn)
	if hostMsg.Body != "hello room" {
		t.Fatalf("host chat %+v", hostMsg)
	}

	_ = hostAck
	_ = guestAck
}

func TestIntegrationPlayBySearchQuery(t *testing.T) {
	s := newTestServer(nil)
	s.resolver = queryResolver{track: protocol.Track{
		VideoID:    "search00123",
		Title:      "Search Hit",
		DurationMs: 180_000,
		SourceURL:  "https://youtube.com/watch?v=search00123",
	}}
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	conn := dialWithNick(t, ctx, wsURL, "host")
	defer conn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn)
	_ = createRoom(t, ctx, conn, "search-play")

	writeMsg(t, ctx, conn, protocol.MsgPlaybackPlay, "q1", protocol.PlaybackPlayPayload{
		Query: "lofi beats",
	})
	state, tick := readPlaybackUpdate(t, conn)
	if state.Status != protocol.PlaybackPlaying || state.Track == nil || state.Track.Title != "Search Hit" {
		t.Fatalf("play state %+v", state)
	}
	if tick.Status != protocol.PlaybackPlaying {
		t.Fatalf("play tick %+v", tick)
	}
}

func TestIntegrationPlayAfterDisconnect(t *testing.T) {
	s := newTestServer(nil)
	s.resolver = fixedResolver{track: protocol.Track{
		VideoID:    "abc123xyz01",
		Title:      "Reconnect Track",
		DurationMs: 60_000,
		SourceURL:  "https://youtube.com/watch?v=abc123xyz01",
	}}
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	conn1, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"X-Nickname": {"host"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ack := readSessionAck(t, conn1)
	createRoom(t, ctx, conn1, "reconnect-play")
	_ = conn1.Close(websocket.StatusNormalClosure, "phase1")

	time.Sleep(50 * time.Millisecond)

	conn2, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"X-Session-Id": {ack.SessionID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn2)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, conn2)

	writeMsg(t, ctx, conn2, protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{
		URL: "https://youtube.com/watch?v=abc123xyz01",
	})
	state, tick := readPlaybackUpdate(t, conn2)
	if state.Status != protocol.PlaybackPlaying || state.Track == nil {
		t.Fatalf("play state %+v", state)
	}
	if tick.Status != protocol.PlaybackPlaying {
		t.Fatalf("play tick %+v", tick)
	}
}

func TestIntegrationPlayAfterDisconnectRealYTDLP(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	s := newTestServer(nil)
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	conn1, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"X-Nickname": {"host"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ack := readSessionAck(t, conn1)
	createRoom(t, ctx, conn1, "yt-play")
	_ = conn1.Close(websocket.StatusNormalClosure, "phase1")

	time.Sleep(50 * time.Millisecond)

	conn2, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"X-Session-Id": {ack.SessionID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn2)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, conn2)

	writeMsg(t, ctx, conn2, protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{
		URL: "https://www.youtube.com/watch?v=jNQXAC9IVRw",
	})
	state, tick := readPlaybackUpdate(t, conn2)
	if state.Status != protocol.PlaybackPlaying || state.Track == nil {
		t.Fatalf("play state %+v", state)
	}
	if tick.Status != protocol.PlaybackPlaying {
		t.Fatalf("play tick %+v", tick)
	}
}

func TestIntegrationListenAndServePlayRealYTDLP(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	s := newTestServer(nil)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	s.startPlaybackTicks()
	go func() { _ = s.server.Serve(ln) }()
	t.Cleanup(func() { _ = s.Shutdown(context.Background()) })

	addr := ln.Addr().(*net.TCPAddr)
	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/v1/ws", addr.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	conn1, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"X-Nickname": {"host"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ack := readSessionAck(t, conn1)
	createRoom(t, ctx, conn1, "listen-serve-play")
	_ = conn1.Close(websocket.StatusNormalClosure, "phase1")

	time.Sleep(50 * time.Millisecond)

	conn2, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"X-Session-Id": {ack.SessionID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn2)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, conn2)

	writeMsg(t, ctx, conn2, protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{
		URL: "https://www.youtube.com/watch?v=jNQXAC9IVRw",
	})
	state, tick := readPlaybackUpdate(t, conn2)
	if state.Status != protocol.PlaybackPlaying || state.Track == nil {
		t.Fatalf("play state %+v", state)
	}
	if tick.Status != protocol.PlaybackPlaying {
		t.Fatalf("play tick %+v", tick)
	}
}

func TestIntegrationSlugTaken(t *testing.T) {
	s := newTestServer(nil)
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	hostConn := dialWithNick(t, ctx, wsURL, "host")
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, hostConn)
	createRoom(t, ctx, hostConn, "taken-slug")

	otherConn := dialWithNick(t, ctx, wsURL, "other")
	defer otherConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, otherConn)

	writeMsg(t, ctx, otherConn, protocol.MsgRoomCreate, "c2", protocol.RoomCreatePayload{Slug: "taken-slug"})
	_, errPayload, err := readEnvelope[protocol.ErrorPayload](t, otherConn)
	if err != nil || errPayload.Code != protocol.ErrSlugTaken {
		t.Fatalf("error %+v err %v", errPayload, err)
	}
}

func TestIntegrationRoomFull(t *testing.T) {
	s := newTestServer(nil)
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	hostConn := dialWithNick(t, ctx, wsURL, "host")
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, hostConn)
	createRoom(t, ctx, hostConn, "full-room")

	now := time.Now()
	for i := 0; i < 18; i++ {
		_, err := s.rooms.Join("full-room", protocol.Member{
			SessionID: fmt.Sprintf("offline-%d", i),
			Nickname:  "member",
		}, now, "")
		if err != nil {
			t.Fatalf("prefill join %d: %v", i, err)
		}
	}

	nineteenth := dialWithNick(t, ctx, wsURL, "nineteenth")
	defer nineteenth.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, nineteenth)
	joinRoom(t, ctx, nineteenth, hostConn, "full-room")

	overflow := dialWithNick(t, ctx, wsURL, "overflow")
	defer overflow.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, overflow)

	writeMsg(t, ctx, overflow, protocol.MsgRoomJoin, "j-full", protocol.RoomJoinPayload{Slug: "full-room"})
	_, errPayload, err := readEnvelope[protocol.ErrorPayload](t, overflow)
	if err != nil || errPayload.Code != protocol.ErrRoomFull {
		t.Fatalf("error %+v err %v", errPayload, err)
	}
}

func TestIntegrationHostLeaveTransfer(t *testing.T) {
	s := newTestServer(nil)
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	hostConn := dialWithNick(t, ctx, wsURL, "host")
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, hostConn)

	guestConn := dialWithNick(t, ctx, wsURL, "guest")
	defer guestConn.Close(websocket.StatusNormalClosure, "done")
	guestAck := readSessionAck(t, guestConn)

	createRoom(t, ctx, hostConn, "host-transfer")
	joinRoom(t, ctx, guestConn, hostConn, "host-transfer")

	writeMsg(t, ctx, hostConn, protocol.MsgRoomLeave, "l1", protocol.RoomLeavePayload{})
	_, _, _ = readEnvelope[protocol.ChatMessage](t, guestConn)
	_, left, err := readEnvelope[protocol.RoomMemberLeftPayload](t, guestConn)
	if err != nil || left.SessionID == "" {
		t.Fatalf("left %+v err %v", left, err)
	}
	_, hostChanged, err := readEnvelope[protocol.RoomHostChangedPayload](t, guestConn)
	if err != nil || hostChanged.HostSessionID != guestAck.SessionID {
		t.Fatalf("host changed %+v guest %s err %v", hostChanged, guestAck.SessionID, err)
	}
}

func TestIntegrationPasswordJoinAndKick(t *testing.T) {
	s := newTestServer(nil)
	ts, wsURL, ctx := startIntegrationHub(t, s)
	defer ts.Close()

	hostConn := dialWithNick(t, ctx, wsURL, "host")
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	hostAck := readSessionAck(t, hostConn)

	guestConn := dialWithNick(t, ctx, wsURL, "guest")
	defer guestConn.Close(websocket.StatusNormalClosure, "done")
	guestAck := readSessionAck(t, guestConn)

	writeMsg(t, ctx, hostConn, protocol.MsgRoomCreate, "c-pw", protocol.RoomCreatePayload{
		Slug:     "pw-room",
		Password: "secret",
	})
	_, snap, err := readEnvelope[protocol.RoomSnapshot](t, hostConn)
	if err != nil || !snap.PasswordProtected {
		t.Fatalf("create snap %+v err %v", snap, err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	writeMsg(t, ctx, guestConn, protocol.MsgRoomJoin, "j-bad", protocol.RoomJoinPayload{
		Slug:     "pw-room",
		Password: "wrong",
	})
	_, errPayload, err := readEnvelope[protocol.ErrorPayload](t, guestConn)
	if err != nil || errPayload.Code != protocol.ErrAuthFailed {
		t.Fatalf("error %+v err %v", errPayload, err)
	}

	guestSnap := joinRoomWithPassword(t, ctx, guestConn, hostConn, "pw-room", "secret")
	if len(guestSnap.Members) != 2 {
		t.Fatalf("guest snap %+v", guestSnap)
	}

	writeMsg(t, ctx, hostConn, protocol.MsgRoomKick, "kick1", protocol.RoomKickPayload{
		TargetSessionID: guestAck.SessionID,
	})
	_, kicked, err := readEnvelope[protocol.RoomKickedPayload](t, guestConn)
	if err != nil || kicked.Message == "" {
		t.Fatalf("kicked %+v err %v", kicked, err)
	}
	_, left, err := readEnvelope[protocol.RoomMemberLeftPayload](t, hostConn)
	if err != nil || left.SessionID != guestAck.SessionID {
		t.Fatalf("left %+v err %v", left, err)
	}

	joinSnap := joinRoomWithPassword(t, ctx, guestConn, hostConn, "pw-room", "secret")
	if len(joinSnap.Members) != 2 {
		t.Fatalf("re-join snap %+v", joinSnap)
	}

	writeMsg(t, ctx, guestConn, protocol.MsgRoomKick, "kick2", protocol.RoomKickPayload{
		TargetSessionID: hostAck.SessionID,
	})
	_, forbid, err := readEnvelope[protocol.ErrorPayload](t, guestConn)
	if err != nil || forbid.Code != protocol.ErrForbidden {
		t.Fatalf("forbid %+v err %v", forbid, err)
	}
}

func joinRoomWithPassword(t *testing.T, ctx context.Context, joiner, host *websocket.Conn, slug, password string) protocol.RoomSnapshot {
	t.Helper()
	writeMsg(t, ctx, joiner, protocol.MsgRoomJoin, "join-"+slug, protocol.RoomJoinPayload{Slug: slug, Password: password})
	_, snap, err := readEnvelope[protocol.RoomSnapshot](t, joiner)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, joiner)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, host)
	_, _, err = readEnvelope[protocol.RoomMemberJoinedPayload](t, host)
	if err != nil {
		t.Fatal(err)
	}
	return snap
}

func startIntegrationHub(t *testing.T, s *Server) (*httptest.Server, string, context.Context) {
	t.Helper()
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(func() { ts.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"
	return ts, wsURL, ctx
}

func writeMsg(t *testing.T, ctx context.Context, conn *websocket.Conn, msgType, id string, payload any) {
	t.Helper()
	data, err := protocol.EncodeMessage(msgType, id, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatal(err)
	}
}

func createRoom(t *testing.T, ctx context.Context, conn *websocket.Conn, slug string) protocol.RoomSnapshot {
	t.Helper()
	writeMsg(t, ctx, conn, protocol.MsgRoomCreate, "create-"+slug, protocol.RoomCreatePayload{Slug: slug})
	_, snap, err := readEnvelope[protocol.RoomSnapshot](t, conn)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, conn)
	return snap
}

func joinRoom(t *testing.T, ctx context.Context, joiner, host *websocket.Conn, slug string) protocol.RoomSnapshot {
	t.Helper()
	writeMsg(t, ctx, joiner, protocol.MsgRoomJoin, "join-"+slug, protocol.RoomJoinPayload{Slug: slug})
	_, snap, err := readEnvelope[protocol.RoomSnapshot](t, joiner)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, joiner)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, host)
	_, _, err = readEnvelope[protocol.RoomMemberJoinedPayload](t, host)
	if err != nil {
		t.Fatal(err)
	}
	return snap
}

func readUserChat(t *testing.T, conn *websocket.Conn) protocol.ChatMessage {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		env, msg, err := readEnvelope[protocol.ChatMessage](t, conn)
		if err != nil {
			t.Fatal(err)
		}
		if env.Type == protocol.MsgChatMessage && msg.Kind == protocol.ChatKindUser && msg.Body != "" {
			return msg
		}
	}
	t.Fatal("timeout waiting for user chat message")
	return protocol.ChatMessage{}
}
