package hub

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/chat"
	"github.com/terminal-music-room/music-room/internal/server/room"
)

func TestChatSendAndPersist(t *testing.T) {
	dir := t.TempDir()
	s := newTestServer(nil)
	s.cfg.DataDir = dir
	s.rooms = room.NewManager(chat.Options{DataDir: dir})

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	conn := dialWithNick(t, ctx, wsURL, "alice")
	defer conn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn)

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "chat-room"})
	_ = conn.Write(ctx, websocket.MessageText, create)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, conn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, conn)

	send, _ := protocol.EncodeMessage(protocol.MsgChatSend, "m1", protocol.ChatSendPayload{Body: "hello 🔥"})
	_ = conn.Write(ctx, websocket.MessageText, send)
	_, msg, err := readEnvelope[protocol.ChatMessage](t, conn)
	if err != nil || msg.Author != "alice" || msg.Body != "hello 🔥" {
		t.Fatalf("msg %+v err %v", msg, err)
	}

	logPath := filepath.Join(dir, "chat-room.chat.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log missing: %v", err)
	}
}

func TestChatSendEmptyRejected(t *testing.T) {
	s := newTestServer(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	conn := dialWithNick(t, ctx, wsURL, "bob")
	defer conn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, conn)

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "empty-chat"})
	_ = conn.Write(ctx, websocket.MessageText, create)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, conn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, conn)

	send, _ := protocol.EncodeMessage(protocol.MsgChatSend, "m1", protocol.ChatSendPayload{Body: "   "})
	_ = conn.Write(ctx, websocket.MessageText, send)
	_, errPayload, err := readEnvelope[protocol.ErrorPayload](t, conn)
	if err != nil || errPayload.Code != protocol.ErrInvalidMessage {
		t.Fatalf("error %+v err %v", errPayload, err)
	}
}
