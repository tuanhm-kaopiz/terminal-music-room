package state

import (
	"sync"

	"github.com/terminal-music-room/music-room/internal/protocol"
)

const maxChatMessages = 100

// ConnStatus is the WebSocket connection lifecycle state.
type ConnStatus string

const (
	StatusDisconnected ConnStatus = "disconnected"
	StatusConnecting   ConnStatus = "connecting"
	StatusConnected    ConnStatus = "connected"
	StatusReconnecting ConnStatus = "reconnecting"
)

// Store holds client-visible session and room state updated from server events.
type Store struct {
	mu sync.RWMutex

	Status      ConnStatus
	SessionID   string
	DisplayName string

	InRoom bool
	Room   protocol.RoomSnapshot

	LastTick protocol.PlaybackTickPayload
	LastErr  *protocol.ErrorPayload
	KickedMessage string

	playbackListeners []chan struct{}
	roomListeners     []chan struct{}
}

// NewStore creates an empty state store.
func NewStore() *Store {
	return &Store{Status: StatusDisconnected}
}

// Snapshot returns a copy of the current store view.
func (s *Store) Snapshot() View {
	s.mu.RLock()
	defer s.mu.RUnlock()
	room := s.Room
	if s.Room.Chat != nil {
		room.Chat = append([]protocol.ChatMessage(nil), s.Room.Chat...)
	}
	if s.Room.Members != nil {
		room.Members = append([]protocol.Member(nil), s.Room.Members...)
	}
	if s.Room.Queue != nil {
		room.Queue = append([]protocol.QueueItem(nil), s.Room.Queue...)
	}
	if s.Room.Reactions != nil {
		room.Reactions = copyReactions(s.Room.Reactions)
	}
	var lastErr *protocol.ErrorPayload
	if s.LastErr != nil {
		errCopy := *s.LastErr
		lastErr = &errCopy
	}
	return View{
		Status:        s.Status,
		SessionID:     s.SessionID,
		DisplayName:   s.DisplayName,
		InRoom:        s.InRoom,
		Room:          room,
		LastTick:      s.LastTick,
		LastErr:       lastErr,
		KickedMessage: s.KickedMessage,
	}
}

// View is an immutable snapshot of store fields for readers.
type View struct {
	Status      ConnStatus
	SessionID   string
	DisplayName string
	InRoom      bool
	Room        protocol.RoomSnapshot
	LastTick    protocol.PlaybackTickPayload
	LastErr       *protocol.ErrorPayload
	KickedMessage string
}

// SubscribePlayback returns a coalesced notify channel for playback.state/tick updates.
func (s *Store) SubscribePlayback() <-chan struct{} {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.playbackListeners = append(s.playbackListeners, ch)
	s.mu.Unlock()
	return ch
}

func (s *Store) emitPlaybackChange() {
	for _, ch := range s.playbackListeners {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// SubscribeRoom returns a coalesced notify channel for any room state change (AC-054).
func (s *Store) SubscribeRoom() <-chan struct{} {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.roomListeners = append(s.roomListeners, ch)
	s.mu.Unlock()
	return ch
}

func (s *Store) emitRoomChange() {
	s.emitPlaybackChange()
	for _, ch := range s.roomListeners {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// SetStatus updates connection status.
func (s *Store) SetStatus(status ConnStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

// ClearRoom resets room state after leave or expired reconnect (AC-050).
func (s *Store) ClearRoom() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InRoom = false
	s.Room = protocol.RoomSnapshot{}
	s.LastTick = protocol.PlaybackTickPayload{}
	s.KickedMessage = ""
	s.emitPlaybackChange()
	s.emitRoomChange()
}

// Apply dispatches a server envelope into the store.
func (s *Store) Apply(env protocol.Envelope) error {
	switch env.Type {
	case protocol.MsgSessionAck:
		var p protocol.SessionAckPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applySessionAck(p)
	case protocol.MsgRoomSnapshot:
		var p protocol.RoomSnapshot
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyRoomSnapshot(p)
	case protocol.MsgRoomMemberJoined:
		var p protocol.RoomMemberJoinedPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyMemberJoined(p)
	case protocol.MsgRoomMemberLeft:
		var p protocol.RoomMemberLeftPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyMemberLeft(p)
	case protocol.MsgRoomHostChanged:
		var p protocol.RoomHostChangedPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyHostChanged(p)
	case protocol.MsgRoomKicked:
		var p protocol.RoomKickedPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyRoomKicked(p)
	case protocol.MsgPlaybackState:
		var p protocol.PlaybackState
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyPlaybackState(p)
	case protocol.MsgPlaybackTick:
		var p protocol.PlaybackTickPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyPlaybackTick(p)
	case protocol.MsgQueueUpdated:
		var p protocol.QueueUpdatedPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyQueueUpdated(p)
	case protocol.MsgChatMessage:
		var p protocol.ChatMessage
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyChatMessage(p)
	case protocol.MsgVoteUpdated:
		var p protocol.VoteUpdatedPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyVoteUpdated(p)
	case protocol.MsgReactionUpdated:
		var p protocol.ReactionUpdatedPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyReactionUpdated(p)
	case protocol.MsgError:
		var p protocol.ErrorPayload
		if err := env.UnmarshalPayload(&p); err != nil {
			return err
		}
		s.applyError(p)
	}
	return nil
}

func (s *Store) applySessionAck(p protocol.SessionAckPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SessionID = p.SessionID
	s.DisplayName = p.DisplayName
}

func (s *Store) applyRoomSnapshot(p protocol.RoomSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InRoom = p.Slug != ""
	s.Room = p
	if len(s.Room.Chat) > maxChatMessages {
		s.Room.Chat = s.Room.Chat[len(s.Room.Chat)-maxChatMessages:]
	}
	s.emitRoomChange()
}

func (s *Store) applyMemberJoined(p protocol.RoomMemberJoinedPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range s.Room.Members {
		if m.SessionID == p.Member.SessionID {
			return
		}
	}
	s.Room.Members = append(s.Room.Members, p.Member)
	s.emitRoomChange()
}

func (s *Store) applyMemberLeft(p protocol.RoomMemberLeftPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.Room.Members[:0]
	for _, m := range s.Room.Members {
		if m.SessionID != p.SessionID {
			out = append(out, m)
		}
	}
	s.Room.Members = out
	s.emitRoomChange()
}

func (s *Store) applyRoomKicked(p protocol.RoomKickedPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InRoom = false
	s.Room = protocol.RoomSnapshot{}
	s.LastTick = protocol.PlaybackTickPayload{}
	s.KickedMessage = p.Message
	if s.KickedMessage == "" {
		s.KickedMessage = "Removed from room by host"
	}
	s.emitPlaybackChange()
	s.emitRoomChange()
}

func (s *Store) applyHostChanged(p protocol.RoomHostChangedPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Room.HostID = p.HostSessionID
	for i := range s.Room.Members {
		s.Room.Members[i].IsHost = s.Room.Members[i].SessionID == p.HostSessionID
	}
	s.emitRoomChange()
}

func (s *Store) applyPlaybackState(p protocol.PlaybackState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Room.Playback = p
	s.emitRoomChange()
}

func (s *Store) applyPlaybackTick(p protocol.PlaybackTickPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastTick = p
	if p.Status == protocol.PlaybackPlaying {
		s.Room.Playback.Status = p.Status
		s.Room.Playback.PositionMs = p.PositionMs
	}
	s.emitRoomChange()
}

func (s *Store) applyQueueUpdated(p protocol.QueueUpdatedPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Room.Queue = append([]protocol.QueueItem(nil), p.Items...)
	s.emitRoomChange()
}

func (s *Store) applyChatMessage(p protocol.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Room.Chat = append(s.Room.Chat, p)
	if len(s.Room.Chat) > maxChatMessages {
		s.Room.Chat = s.Room.Chat[len(s.Room.Chat)-maxChatMessages:]
	}
	s.emitRoomChange()
}

func (s *Store) applyVoteUpdated(p protocol.VoteUpdatedPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.Vote == nil {
		s.Room.Vote = nil
		s.emitRoomChange()
		return
	}
	v := *p.Vote
	s.Room.Vote = &v
	s.emitRoomChange()
}

func (s *Store) applyReactionUpdated(p protocol.ReactionUpdatedPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Room.Reactions = copyReactions(p.Counts)
	s.emitRoomChange()
}

func (s *Store) applyError(p protocol.ErrorPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	errCopy := p
	s.LastErr = &errCopy
	s.emitRoomChange()
}

func copyReactions(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
