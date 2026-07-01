package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/panels"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

func chatTestStore(t *testing.T) *state.Store {
	t.Helper()
	store := state.NewStore()
	ack, err := protocol.NewEnvelope(protocol.MsgSessionAck, "", protocol.SessionAckPayload{
		SessionID:   "sess-host",
		DisplayName: "host#1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(ack); err != nil {
		t.Fatal(err)
	}
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", panels.FixtureView().Room)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestChatSendViaActions(t *testing.T) {
	store := chatTestStore(t)
	var sentType string
	var sentBody string
	act := actions.New(func(_ context.Context, msgType string, payload any) error {
		sentType = msgType
		p, ok := payload.(protocol.ChatSendPayload)
		if !ok {
			t.Fatalf("payload type %T", payload)
		}
		sentBody = p.Body
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	m.input.SetValue("hello crew")
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(*Model)
	if sentType != protocol.MsgChatSend {
		t.Fatalf("msgType = %q, want chat.send", sentType)
	}
	if sentBody != "hello crew" {
		t.Fatalf("body = %q", sentBody)
	}
	if got.input.Value() != "" {
		t.Fatalf("input should clear after send, got %q", got.input.Value())
	}
	if got.errMsg != "" {
		t.Fatalf("unexpected errMsg: %q", got.errMsg)
	}
}

func TestChatRejectEmpty(t *testing.T) {
	store := chatTestStore(t)
	sent := false
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = true
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	m.input.SetValue("   ")
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sent {
		t.Fatal("empty chat must not be sent (AC-027)")
	}
}

func TestChatSendErrorToast(t *testing.T) {
	store := chatTestStore(t)
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		return context.Canceled
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	m.input.SetValue("oops")
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := next.(*Model)
	if got.errMsg == "" {
		t.Fatal("expected error toast on send failure")
	}
	if got.input.Value() != "oops" {
		t.Fatal("input should remain when send fails")
	}
}

func TestChatLastErrToast(t *testing.T) {
	store := chatTestStore(t)
	env, err := protocol.NewEnvelope(protocol.MsgError, "", protocol.ErrorPayload{Message: "forbidden"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}

	m := NewModel(context.Background(), Config{Store: store})
	m.refresh()
	if m.errMsg != "forbidden" {
		t.Fatalf("errMsg = %q, want forbidden", m.errMsg)
	}
}

func TestChatScrollOnNewMessage(t *testing.T) {
	store := chatTestStore(t)
	m := NewModel(context.Background(), Config{Store: store})
	m.chatScroll = 3
	m.view = store.Snapshot()

	room := m.view.Room
	room.Chat = append(room.Chat, protocol.ChatMessage{Kind: protocol.ChatKindUser, Author: "x", Body: "new"})
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", room)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	m.refresh()
	if m.chatScroll != 0 {
		t.Fatalf("chatScroll = %d, want 0 after new message", m.chatScroll)
	}
}

func playbackTestStore(t *testing.T, playback protocol.PlaybackState) *state.Store {
	t.Helper()
	store := state.NewStore()
	room := panels.FixtureView().Room
	room.Playback = playback
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", room)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestPlaybackPauseWhenPlaying(t *testing.T) {
	store := playbackTestStore(t, protocol.PlaybackState{
		Status: protocol.PlaybackPlaying,
		Track:  &protocol.Track{Title: "Neon Nights"},
	})
	var sent string
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = msgType
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeySpace})
	if sent != protocol.MsgPlaybackPause {
		t.Fatalf("sent %q, want pause", sent)
	}
}

func TestPlaybackResumeWhenPaused(t *testing.T) {
	store := playbackTestStore(t, protocol.PlaybackState{
		Status: protocol.PlaybackPaused,
		Track:  &protocol.Track{Title: "Neon Nights"},
	})
	var sent string
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = msgType
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeySpace})
	if sent != protocol.MsgPlaybackResume {
		t.Fatalf("sent %q, want resume", sent)
	}
}

func TestPlaybackSkipSendsMessage(t *testing.T) {
	store := playbackTestStore(t, protocol.PlaybackState{
		Status: protocol.PlaybackPlaying,
		Track:  &protocol.Track{Title: "Neon Nights"},
	})
	var sent string
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = msgType
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if sent != protocol.MsgPlaybackSkip {
		t.Fatalf("sent %q, want skip", sent)
	}
}

func TestPlaybackNoTrackGuard(t *testing.T) {
	store := playbackTestStore(t, protocol.PlaybackState{Status: protocol.PlaybackEnded})
	sent := false
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		sent = true
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeySpace})
	got := next.(*Model)
	if sent {
		t.Fatal("space should not send without track")
	}
	if got.errMsg == "" {
		t.Fatal("expected error toast for no track")
	}

	next, _ = got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	got = next.(*Model)
	if sent {
		t.Fatal("skip should not send without track")
	}
	if got.errMsg == "" {
		t.Fatal("expected error toast for skip without track")
	}
}

func TestVoteSkipShortcut(t *testing.T) {
	store := playbackTestStore(t, protocol.PlaybackState{
		Status: protocol.PlaybackPlaying,
		Track:  &protocol.Track{Title: "Neon Nights"},
	})
	var sent string
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = msgType
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	if sent != protocol.MsgVoteSkip {
		t.Fatalf("sent %q, want vote.skip", sent)
	}
}

func TestVotePriorityShortcut(t *testing.T) {
	store := chatTestStore(t)
	var itemID string
	act := actions.New(func(_ context.Context, msgType string, payload any) error {
		if msgType != protocol.MsgVotePriority {
			t.Fatalf("msgType = %q", msgType)
		}
		p, ok := payload.(protocol.VotePriorityPayload)
		if !ok {
			t.Fatalf("payload type %T", payload)
		}
		itemID = p.ItemID
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	m.selectedQueueIdx = 1
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	if itemID != "q2" {
		t.Fatalf("itemID = %q, want q2", itemID)
	}
}

func TestVotePriorityNoSelection(t *testing.T) {
	store := seekStoreForVote(t)
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		t.Fatal("should not vote priority without queue item")
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	got := next.(*Model)
	if got.errMsg == "" {
		t.Fatal("expected error when queue empty")
	}
}

func TestVoteReactionShortcut(t *testing.T) {
	store := playbackTestStore(t, protocol.PlaybackState{
		Status: protocol.PlaybackPlaying,
		Track:  &protocol.Track{Title: "Neon Nights"},
	})
	var emoji string
	act := actions.New(func(_ context.Context, msgType string, payload any) error {
		if msgType != protocol.MsgReactionSend {
			t.Fatalf("msgType = %q", msgType)
		}
		p, ok := payload.(protocol.ReactionSendPayload)
		if !ok {
			t.Fatalf("payload type %T", payload)
		}
		emoji = p.Emoji
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if emoji != "🔥" {
		t.Fatalf("emoji = %q, want fire", emoji)
	}
}

func TestVoteReactionNoTrackGuard(t *testing.T) {
	store := playbackTestStore(t, protocol.PlaybackState{Status: protocol.PlaybackEnded})
	sent := false
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		sent = true
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	got := next.(*Model)
	if sent {
		t.Fatal("reaction should not send without track")
	}
	if got.errMsg == "" {
		t.Fatal("expected error toast (AC-036)")
	}
}

func seekStoreForVote(t *testing.T) *state.Store {
	t.Helper()
	store := state.NewStore()
	room := panels.FixtureView().Room
	room.Queue = nil
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", room)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	return store
}

func memberTestStore(t *testing.T) *state.Store {
	t.Helper()
	store := state.NewStore()
	ack, err := protocol.NewEnvelope(protocol.MsgSessionAck, "", protocol.SessionAckPayload{
		SessionID:   "sess-guest",
		DisplayName: "guest#2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(ack); err != nil {
		t.Fatal(err)
	}
	env, err := protocol.NewEnvelope(protocol.MsgRoomSnapshot, "", panels.FixtureView().Room)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestQueueRemoveHost(t *testing.T) {
	store := chatTestStore(t)
	var removed string
	act := actions.New(func(_ context.Context, msgType string, payload any) error {
		if msgType != protocol.MsgQueueRemove {
			t.Fatalf("msgType = %q", msgType)
		}
		p, ok := payload.(protocol.QueueRemovePayload)
		if !ok {
			t.Fatalf("payload type %T", payload)
		}
		removed = p.ItemID
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if removed != "q1" {
		t.Fatalf("removed %q, want q1", removed)
	}
}

func TestQueueRemoveMemberDenied(t *testing.T) {
	store := memberTestStore(t)
	sent := false
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		sent = true
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := next.(*Model)
	if sent {
		t.Fatal("member must not remove queue items (AC-022)")
	}
	if got.errMsg != "host only" {
		t.Fatalf("errMsg = %q", got.errMsg)
	}
}

func TestQueueReorderDownHost(t *testing.T) {
	store := chatTestStore(t)
	var itemID, afterID string
	act := actions.New(func(_ context.Context, msgType string, payload any) error {
		if msgType != protocol.MsgQueueReorder {
			t.Fatalf("msgType = %q", msgType)
		}
		p, ok := payload.(protocol.QueueReorderPayload)
		if !ok {
			t.Fatalf("payload type %T", payload)
		}
		itemID = p.ItemID
		afterID = p.AfterID
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyCtrlDown})
	if itemID != "q1" || afterID != "q2" {
		t.Fatalf("reorder %q after %q", itemID, afterID)
	}
}

func TestQueueReorderMemberDenied(t *testing.T) {
	store := memberTestStore(t)
	sent := false
	act := actions.New(func(_ context.Context, _ string, _ any) error {
		sent = true
		return nil
	}, store)

	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyCtrlDown})
	got := next.(*Model)
	if sent {
		t.Fatal("member must not reorder queue (AC-022)")
	}
	if got.errMsg != "host only" {
		t.Fatalf("errMsg = %q", got.errMsg)
	}
}

func TestQueueHostChangedGating(t *testing.T) {
	store := memberTestStore(t)
	m := NewModel(context.Background(), Config{Store: store})
	if IsHost(m.view) {
		t.Fatal("expected member before host change")
	}

	env, err := protocol.NewEnvelope(protocol.MsgRoomHostChanged, "", protocol.RoomHostChangedPayload{
		HostSessionID: "sess-guest",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Apply(env); err != nil {
		t.Fatal(err)
	}
	m.refresh()
	if !IsHost(m.view) {
		t.Fatal("expected host after host_changed (AC-039)")
	}
}

func TestLeaveOpensModal(t *testing.T) {
	store := chatTestStore(t)
	m := NewModel(context.Background(), Config{Store: store})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	got := next.(*Model)
	if got.mode != ModeModalLeave {
		t.Fatalf("mode = %v, want leave modal", got.mode)
	}
}

func TestLeaveCancelEsc(t *testing.T) {
	store := chatTestStore(t)
	m := NewModel(context.Background(), Config{Store: store})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = *next.(*Model)
	next, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyEscape})
	got := next.(*Model)
	if got.mode != ModeDashboard {
		t.Fatalf("mode = %v, want dashboard after esc", got.mode)
	}
}

func TestLeaveConfirmCallsLeave(t *testing.T) {
	store := chatTestStore(t)
	left := false
	m := NewModel(context.Background(), Config{
		Store: store,
		Leave: func(ctx context.Context) error {
			left = true
			return nil
		},
	})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = *next.(*Model)
	next, cmd := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	got := next.(*Model)
	if !left {
		t.Fatal("expected cfg.Leave on confirm")
	}
	if !got.quit || cmd == nil {
		t.Fatal("expected TUI quit after leave")
	}
}

func TestLeaveConfirmViaActions(t *testing.T) {
	store := chatTestStore(t)
	var sent string
	act := actions.New(func(_ context.Context, msgType string, _ any) error {
		sent = msgType
		return nil
	}, store)
	m := NewModel(context.Background(), Config{Store: store, Actions: act})
	next, _ := (&m).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = *next.(*Model)
	_, _ = (&m).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sent != protocol.MsgRoomLeave {
		t.Fatalf("sent %q, want room.leave", sent)
	}
}
