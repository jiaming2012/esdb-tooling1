package models

import (
	"github.com/google/uuid"
)

type StreamName string

type StreamEventHeader struct {
	StreamName    StreamName
	EventType     EventType
	SchemaVersion int
}

type StreamEventMeta struct {
	RequestID     uuid.UUID
	IsReplay      bool
}

func (m *StreamEventMeta) SetMeta(requstID uuid.UUID, isReplay bool) {
	m.RequestID = requstID
	m.IsReplay = isReplay
}

type StreamEvent interface {
	GetStreamEventHeader() *StreamEventHeader
	GetStreamEventMeta() *StreamEventMeta
}
