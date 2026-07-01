package protocol

import (
	"encoding/json"
	"fmt"
)

// Envelope is the on-wire WebSocket JSON frame.
type Envelope struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// UnmarshalPayload decodes the envelope payload into v.
func (e *Envelope) UnmarshalPayload(v any) error {
	if len(e.Payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(e.Payload, v); err != nil {
		return fmt.Errorf("protocol: decode %q payload: %w", e.Type, err)
	}
	return nil
}

// NewEnvelope builds an envelope with a typed payload.
func NewEnvelope(msgType, id string, payload any) (Envelope, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return Envelope{}, fmt.Errorf("protocol: encode %q payload: %w", msgType, err)
		}
		raw = b
	}
	return Envelope{Type: msgType, ID: id, Payload: raw}, nil
}

// Encode serializes an envelope to JSON bytes.
func Encode(env Envelope) ([]byte, error) {
	b, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("protocol: marshal envelope: %w", err)
	}
	return b, nil
}

// EncodeMessage serializes a typed message as an envelope.
func EncodeMessage(msgType, id string, payload any) ([]byte, error) {
	env, err := NewEnvelope(msgType, id, payload)
	if err != nil {
		return nil, err
	}
	return Encode(env)
}

// Decode parses JSON bytes into an Envelope.
func Decode(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, fmt.Errorf("protocol: unmarshal envelope: %w", err)
	}
	if env.Type == "" {
		return Envelope{}, fmt.Errorf("protocol: missing message type")
	}
	return env, nil
}

// DecodePayload decodes JSON bytes into an envelope and typed payload.
func DecodePayload[T any](data []byte) (Envelope, T, error) {
	var zero T
	env, err := Decode(data)
	if err != nil {
		return Envelope{}, zero, err
	}
	var payload T
	if err := env.UnmarshalPayload(&payload); err != nil {
		return env, zero, err
	}
	return env, payload, nil
}

// NewErrorEnvelope builds a server error message.
func NewErrorEnvelope(id string, code ErrorCode, message string, retryAfter *int) (Envelope, error) {
	return NewEnvelope(MsgError, id, ErrorPayload{
		Code:       code,
		Message:    message,
		RetryAfter: retryAfter,
	})
}
