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

func TestReconnectRestoresRoomSnapshot(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	conn1, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"host"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ack := readSessionAck(t, conn1)

	create, err := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "reconnect-room"})
	if err != nil {
		t.Fatal(err)
	}
	if err := conn1.Write(ctx, websocket.MessageText, create); err != nil {
		t.Fatal(err)
	}
	_, snap, err := readEnvelope[protocol.RoomSnapshot](t, conn1)
	if err != nil || snap.Slug != "reconnect-room" {
		t.Fatalf("create snap %+v err %v", snap, err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, conn1)

	_ = conn1.Close(websocket.StatusNormalClosure, "bye")
	time.Sleep(50 * time.Millisecond)

	conn2, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Session-Id": []string{ack.SessionID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")

	ack2 := readSessionAck(t, conn2)
	if ack2.SessionID != ack.SessionID {
		t.Fatalf("session changed %s -> %s", ack.SessionID, ack2.SessionID)
	}
	_, restored, err := readEnvelope[protocol.RoomSnapshot](t, conn2)
	if err != nil || restored.Slug != "reconnect-room" || len(restored.Members) != 1 {
		t.Fatalf("restored snap %+v err %v", restored, err)
	}
}

func TestReconnectExpiredRequiresJoin(t *testing.T) {
	s := newTestServer(nil)
	s.reconnectTTL = 20 * time.Millisecond
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

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "expire-room"})
	if err := hostConn.Write(ctx, websocket.MessageText, create); err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, hostConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	guestConn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"guest"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ack := readSessionAck(t, guestConn)

	join, _ := protocol.EncodeMessage(protocol.MsgRoomJoin, "j1", protocol.RoomJoinPayload{Slug: "expire-room"})
	if err := guestConn.Write(ctx, websocket.MessageText, join); err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)
	_, _, _ = readEnvelope[protocol.RoomMemberJoinedPayload](t, hostConn)

	_ = guestConn.Close(websocket.StatusNormalClosure, "bye")
	time.Sleep(50 * time.Millisecond)

	conn2, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Session-Id": []string{ack.SessionID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close(websocket.StatusNormalClosure, "done")

	_ = readSessionAck(t, conn2)

	joinAgain, _ := protocol.EncodeMessage(protocol.MsgRoomJoin, "j2", protocol.RoomJoinPayload{Slug: "expire-room"})
	if err := conn2.Write(ctx, websocket.MessageText, joinAgain); err != nil {
		t.Fatal(err)
	}
	_, joinedSnap, err := readEnvelope[protocol.RoomSnapshot](t, conn2)
	if err != nil || joinedSnap.Slug != "expire-room" {
		t.Fatalf("manual join snap %+v err %v", joinedSnap, err)
	}
}

func TestReactionSendAggregated(t *testing.T) {
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

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "react-room"})
	if err := hostConn.Write(ctx, websocket.MessageText, create); err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, hostConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	join, _ := protocol.EncodeMessage(protocol.MsgRoomJoin, "j1", protocol.RoomJoinPayload{Slug: "react-room"})
	if err := guestConn.Write(ctx, websocket.MessageText, join); err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)
	_, _, _ = readEnvelope[protocol.RoomMemberJoinedPayload](t, hostConn)

	play, _ := protocol.EncodeMessage(protocol.MsgPlaybackPlay, "p1", protocol.PlaybackPlayPayload{
		URL: "https://youtube.com/watch?v=abc123xyz01",
	})
	if err := hostConn.Write(ctx, websocket.MessageText, play); err != nil {
		t.Fatal(err)
	}
	_, _ = readPlaybackUpdate(t, hostConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	react, _ := protocol.EncodeMessage(protocol.MsgReactionSend, "r1", protocol.ReactionSendPayload{Emoji: "🔥"})
	if err := guestConn.Write(ctx, websocket.MessageText, react); err != nil {
		t.Fatal(err)
	}
	_, updated, err := readEnvelope[protocol.ReactionUpdatedPayload](t, hostConn)
	if err != nil || updated.Counts["🔥"] != 1 {
		t.Fatalf("reaction updated %+v err %v", updated, err)
	}

	if err := guestConn.Write(ctx, websocket.MessageText, react); err != nil {
		t.Fatal(err)
	}
	_, updated, err = readEnvelope[protocol.ReactionUpdatedPayload](t, hostConn)
	if err != nil || updated.Counts["🔥"] != 2 {
		t.Fatalf("aggregated counts %+v err %v", updated, err)
	}
}

func TestReactionSendNoTrack(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"host"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn)

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "no-track"})
	if err := conn.Write(ctx, websocket.MessageText, create); err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, conn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, conn)

	react, _ := protocol.EncodeMessage(protocol.MsgReactionSend, "r1", protocol.ReactionSendPayload{Emoji: "🔥"})
	if err := conn.Write(ctx, websocket.MessageText, react); err != nil {
		t.Fatal(err)
	}
	_, errPayload, err := readEnvelope[protocol.ErrorPayload](t, conn)
	if err != nil || errPayload.Message != "no track playing" {
		t.Fatalf("error %+v err %v", errPayload, err)
	}
}
