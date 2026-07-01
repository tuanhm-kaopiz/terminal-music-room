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

func TestRoomCreateJoinLeaveFlow(t *testing.T) {
	s := newTestServer(nil)
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

	create, err := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "backend-team"})
	if err != nil {
		t.Fatal(err)
	}
	if err := hostConn.Write(ctx, websocket.MessageText, create); err != nil {
		t.Fatal(err)
	}
	_, snap, err := readEnvelope[protocol.RoomSnapshot](t, hostConn)
	if err != nil || snap.Slug != "backend-team" || len(snap.Members) != 1 {
		t.Fatalf("snapshot %+v err %v", snap, err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	guestConn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{"guest"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer guestConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, guestConn)

	join, err := protocol.EncodeMessage(protocol.MsgRoomJoin, "j1", protocol.RoomJoinPayload{Slug: "backend-team"})
	if err != nil {
		t.Fatal(err)
	}
	if err := guestConn.Write(ctx, websocket.MessageText, join); err != nil {
		t.Fatal(err)
	}
	_, guestSnap, err := readEnvelope[protocol.RoomSnapshot](t, guestConn)
	if err != nil || len(guestSnap.Members) != 2 {
		t.Fatalf("guest snap %+v err %v", guestSnap, err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	_, joined, err := readEnvelope[protocol.RoomMemberJoinedPayload](t, hostConn)
	if err != nil || joined.Member.Nickname != "guest" {
		t.Fatalf("joined %+v err %v", joined, err)
	}

	leave, err := protocol.EncodeMessage(protocol.MsgRoomLeave, "l1", protocol.RoomLeavePayload{})
	if err != nil {
		t.Fatal(err)
	}
	if err := guestConn.Write(ctx, websocket.MessageText, leave); err != nil {
		t.Fatal(err)
	}
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)
	_, left, err := readEnvelope[protocol.RoomMemberLeftPayload](t, hostConn)
	if err != nil || left.SessionID == "" {
		t.Fatalf("left %+v err %v", left, err)
	}
}

func readEnvelope[T any](t *testing.T, conn *websocket.Conn) (protocol.Envelope, T, error) {
	t.Helper()
	var zero T
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, data, err := conn.Read(ctx)
	if err != nil {
		return protocol.Envelope{}, zero, err
	}
	return protocol.DecodePayload[T](data)
}