package hub

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/queuehistory"
)

func TestQueueAddRemoveForbidden(t *testing.T) {
	dir := t.TempDir()
	s := newTestServer(nil)
	s.cfg.DataDir = dir
	s.cfg.QueueHistoryDir = filepath.Join(dir, "queue")
	s.queueHistory = queuehistory.NewStore(s.cfg.QueueHistoryDir)
	s.resolver = fixedResolver{track: protocol.Track{
		VideoID:    "queuevid001",
		Title:      "Queued Song",
		DurationMs: 120_000,
		SourceURL:  "https://youtube.com/watch?v=queuevid001",
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

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "queue-room"})
	_ = hostConn.Write(ctx, websocket.MessageText, create)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, hostConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	join, _ := protocol.EncodeMessage(protocol.MsgRoomJoin, "j1", protocol.RoomJoinPayload{Slug: "queue-room"})
	_ = guestConn.Write(ctx, websocket.MessageText, join)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, guestConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	add, _ := protocol.EncodeMessage(protocol.MsgQueueAdd, "qa1", protocol.QueueAddPayload{
		URL: "https://youtube.com/watch?v=queuevid001",
	})
	_ = guestConn.Write(ctx, websocket.MessageText, add)
	_, updated, err := readEnvelope[protocol.QueueUpdatedPayload](t, guestConn)
	if err != nil || len(updated.Items) != 1 {
		t.Fatalf("queue updated %+v err %v", updated, err)
	}
	if updated.Items[0].Title != "Queued Song" || updated.Items[0].AddedBy == "" {
		t.Fatalf("item %+v", updated.Items[0])
	}
	historyPath := s.queueHistory.Path("queue-room")
	data, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("queue history: %v", err)
	}
	if !strings.Contains(string(data), "queuevid001") || !strings.Contains(string(data), "Queued Song") {
		t.Fatalf("history %q", string(data))
	}
	itemID := updated.Items[0].ID

	remove, _ := protocol.EncodeMessage(protocol.MsgQueueRemove, "qr1", protocol.QueueRemovePayload{ItemID: itemID})
	_ = guestConn.Write(ctx, websocket.MessageText, remove)
	_, forbidden, err := readEnvelope[protocol.ErrorPayload](t, guestConn)
	if err != nil || forbidden.Code != protocol.ErrForbidden {
		t.Fatalf("error %+v err %v", forbidden, err)
	}

	_ = hostConn.Write(ctx, websocket.MessageText, remove)
	_, afterRemove, err := readEnvelope[protocol.QueueUpdatedPayload](t, guestConn)
	if err != nil || len(afterRemove.Items) != 0 {
		t.Fatalf("queue %+v err %v", afterRemove, err)
	}
}

func TestQueueReorderHostOnly(t *testing.T) {
	s := newTestServer(nil)
	s.resolver = fixedResolver{track: protocol.Track{
		VideoID: "reorderv001",
		Title:   "Reorder",
	}}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	wsURL := "ws" + ts.URL[4:] + "/v1/ws"

	hostConn := dialWithNick(t, ctx, wsURL, "host")
	defer hostConn.Close(websocket.StatusNormalClosure, "done")
	_ = readSessionAck(t, hostConn)

	create, _ := protocol.EncodeMessage(protocol.MsgRoomCreate, "c1", protocol.RoomCreatePayload{Slug: "reorder-room"})
	_ = hostConn.Write(ctx, websocket.MessageText, create)
	_, _, _ = readEnvelope[protocol.RoomSnapshot](t, hostConn)
	_, _, _ = readEnvelope[protocol.ChatMessage](t, hostConn)

	for i, id := range []string{"reorderv001", "reorderv002", "reorderv003"} {
		add, _ := protocol.EncodeMessage(protocol.MsgQueueAdd, "a"+string(rune('0'+i)), protocol.QueueAddPayload{
			URL: "https://youtube.com/watch?v=" + id,
		})
		_ = hostConn.Write(ctx, websocket.MessageText, add)
		_, _, _ = readEnvelope[protocol.QueueUpdatedPayload](t, hostConn)
	}

	r, ok := s.rooms.Get("reorder-room")
	if !ok || len(r.Queue) < 3 {
		t.Fatal("expected 3 queue items")
	}
	thirdID := r.Queue[2].ID
	firstID := r.Queue[0].ID

	reorder, _ := protocol.EncodeMessage(protocol.MsgQueueReorder, "ro1", protocol.QueueReorderPayload{
		ItemID:  thirdID,
		AfterID: firstID,
	})
	_ = hostConn.Write(ctx, websocket.MessageText, reorder)
	_, updated, err := readEnvelope[protocol.QueueUpdatedPayload](t, hostConn)
	if err != nil || len(updated.Items) != 3 {
		t.Fatalf("updated %+v err %v", updated, err)
	}
	if updated.Items[0].ID != firstID || updated.Items[1].ID != thirdID {
		t.Fatalf("order %+v", updated.Items)
	}
}

func dialWithNick(t *testing.T, ctx context.Context, wsURL, nick string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{"X-Nickname": []string{nick}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return conn
}
