package models

import (
	"github.com/google/uuid"
)

type StreamName string

type StreamEventHeader struct {
	RequestID     uuid.UUID
	StreamName    StreamName
	EventType     EventType
	SchemaVersion int
}

func (h *StreamEventHeader) SetSchemaVersion(version int) {
	h.SchemaVersion = version
}

type StreamEvent interface {
	GetStreamEventHeader() StreamEventHeader
}
