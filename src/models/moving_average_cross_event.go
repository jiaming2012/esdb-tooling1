package models

import "github.com/google/uuid"

type MovingAverageState string

const (
	MovingAverageStateBelow MovingAverageState = "crossed_below"
	MovingAverageStateAbove MovingAverageState = "crossed_above"
	MovingAverageStateNone  MovingAverageState = "none"
)

const (
	MovingAverageStream StreamName = "moving_average"
)

type MovingAverageCrossEvent struct {
	Meta      StreamEventMeta
	RequestID uuid.UUID
	Symbol    string
	Period    int
	State     MovingAverageState
}

func NewMovingAverageCrossEvent(symbol string, period int, state MovingAverageState) *MovingAverageCrossEvent {
	return &MovingAverageCrossEvent{
		Symbol: symbol,
		Period: period,
		State:  state,
	}
}

func (ma *MovingAverageCrossEvent) GetStreamEventMeta() *StreamEventMeta {
	return &ma.Meta
}

func (ma *MovingAverageCrossEvent) GetStreamEventHeader() *StreamEventHeader {
	return &StreamEventHeader{
		StreamName:    MovingAverageStream,
		EventType:     MovingAverageCrossEventType,
		SchemaVersion: 1,
	}
}
