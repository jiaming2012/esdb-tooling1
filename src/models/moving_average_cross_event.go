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
	RequestID uuid.UUID
	Symbol    string
	Period    int
	State     MovingAverageState
}

func NewMovingAverageCrossEvent(requestID uuid.UUID, symbol string, period int, state MovingAverageState) *MovingAverageCrossEvent {
	return &MovingAverageCrossEvent{
		RequestID: requestID,
		Symbol:    symbol,
		Period:    period,
		State:     state,
	}
}

func (ma *MovingAverageCrossEvent) GetStreamEventHeader() StreamEventHeader {
	return StreamEventHeader{
		RequestID:     ma.RequestID,
		StreamName:    MovingAverageStream,
		EventType:     MovingAverageCrossEventType,
		SchemaVersion: 1,
	}
}
