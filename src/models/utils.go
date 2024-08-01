package models

import (
	"context"
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
)

func FindStreamLastEventNumber(ctx context.Context, db *esdb.Client, streamName StreamName) (uint64, error) {
	stream, err := db.ReadStream(ctx, string(streamName), esdb.ReadStreamOptions{
		Direction: esdb.Backwards,
		From:      esdb.End{},
	}, 1)

	if err != nil {
		if esdbError, ok := err.(*esdb.Error); ok && esdbError.IsErrorCode(esdb.ErrorCodeResourceNotFound) {
			return 0, nil
		}

		return 0, fmt.Errorf("FindStreamLastEventNumber: failed to read stream %s: %w", streamName, err)
	}

	event, err := stream.Recv()
	if err != nil {
		if esdbError, ok := err.(*esdb.Error); ok && esdbError.IsErrorCode(esdb.ErrorCodeResourceNotFound) {
			return 0, nil
		}

		return 0, fmt.Errorf("FindStreamLastEventNumber: failed to read event from stream %s: %w", streamName, err)
	}

	return event.Event.EventNumber, nil
}
