package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type EventStreamConsumer[T StreamEvent] struct {
	wg          *sync.WaitGroup
	db          *esdb.Client
	url         string
	savedEvents chan T
	streamName  StreamName
}

func (cli *EventStreamConsumer[T]) GetEvents() <-chan T {
	return cli.savedEvents
}

func (cli *EventStreamConsumer[T]) replayEvents(ctx context.Context, stream StreamName, lastEventNumber uint64) error {
	if lastEventNumber == 0 {
		return nil
	}

	event, err := cli.db.ReadStream(ctx, string(stream), esdb.ReadStreamOptions{}, lastEventNumber)
	if err != nil {
		return fmt.Errorf("EventStreamConsumer.replayEvents: failed to read stream %s: %v", stream, err)
	}

	for {
		event, err := event.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("EventStreamConsumer.replayEvents: failed to read event from stream: %v", err)
		}

		if err := cli.processEvent(event.Event, true); err != nil {
			return fmt.Errorf("EventStreamConsumer.replayEvents: failed to process event: %v", err)
		}
	}

	return nil
}

func (cli *EventStreamConsumer[T]) subscribeToStream(ctx context.Context, wg *sync.WaitGroup, stream StreamName, initialEventNumber uint64) (chan error, error) {
	subscription, err := cli.db.SubscribeToStream(ctx, string(stream), esdb.SubscribeToStreamOptions{
		From: esdb.Revision(initialEventNumber),
	})

	if err != nil {
		return nil, fmt.Errorf("EventStreamConsumer.subscribeToStream: failed to subscribe to stream: %v", err)
	}

	log.Infof("EventStreamConsumer: subscribed to stream %s @ positon %v", stream, initialEventNumber)

	lastEventNumber := initialEventNumber

	errCh := make(chan error)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			for {
				if subscription == nil {
					log.Debugf("EventStreamConsumer.subscribeToStream: subscription is nil")
					break
				}

				event := subscription.Recv()

				if event.SubscriptionDropped != nil {
					log.Infof("EventStreamConsumer.subscribeToStream: subscription dropped: %v", event.SubscriptionDropped.Error)
					break
				}

				if event.EventAppeared == nil {
					continue
				}

				if event.CheckPointReached != nil {
					log.Infof("EventStreamConsumer.subscribeToStream: checkpoint reached: %v\n", event.CheckPointReached)
				}

				ev := event.EventAppeared.Event

				lastEventNumber = event.EventAppeared.OriginalEvent().EventNumber

				if err := cli.processEvent(ev, false); err != nil {
					errCh <- fmt.Errorf("EventStreamConsumer.subscribeToStream: failed to process event: %v", err)
					return
				}
			}

			select {
			case <-ctx.Done():
				// Check if the context is cancelled or expired
				log.Infof("EventStreamConsumer.subscribeToStream: context cancelled or deadline exceeded")
				cli.stop()
				return
			default:
				// Context is not cancelled, proceed with the operation
				log.Infof("EventStreamConsumer.subscribeToStream: re-subscribing subscription @ position %v", lastEventNumber)

				subscription, err = cli.db.SubscribeToStream(ctx, string(stream), esdb.SubscribeToStreamOptions{
					From: esdb.Revision(lastEventNumber),
				})

				if err != nil {
					log.Errorf("EventStreamConsumer.subscribeToStream: failed to subscribe to stream: %v", err)
				}
			}
		}
	}()

	return errCh, nil
}

func (cli *EventStreamConsumer[T]) processEvent(event *esdb.RecordedEvent, isReplay bool) error {
	var savedEvent T

	if err := json.Unmarshal(event.Data, &savedEvent); err != nil {
		return fmt.Errorf("EventStreamConsumer.subscribeToStream: failed to unmarshal event data: %v", err)
	}

	meta := savedEvent.GetStreamEventMeta()

	meta.SetMeta(uuid.Nil, isReplay)

	cli.savedEvents <- savedEvent

	return nil
}

func (cli *EventStreamConsumer[T]) stop() {
	cli.db.Close()
	close(cli.savedEvents)
}

func (cli *EventStreamConsumer[T]) Start(ctx context.Context, wg *sync.WaitGroup) error {
	settings, err := esdb.ParseConnectionString(cli.url)
	if err != nil {
		return fmt.Errorf("eventStreamConsumer.Start: failed to parse connection string: %v", err)
	}

	cli.db, err = esdb.NewClient(settings)
	if err != nil {
		return fmt.Errorf("eventStreamConsumer.Start: failed to create client: %v", err)
	}

	lastEventNumber, err := FindStreamLastEventNumber(ctx, cli.db, cli.streamName)
	if err != nil {
		return fmt.Errorf("eventStreamConsumer.Start: failed to find last event number: %v", err)
	}

	if err := cli.replayEvents(ctx, cli.streamName, lastEventNumber); err != nil {
		return fmt.Errorf("eventStreamConsumer.Start: failed to replay events: %v", err)
	}

	if _, err := cli.subscribeToStream(ctx, wg, cli.streamName, lastEventNumber); err != nil {
		return fmt.Errorf("eventStreamConsumer.Start: failed to subscribe to stream: %v", err)
	}

	return nil
}

func NewEventStreamConsumer[T StreamEvent](db *esdb.Client, url string, instance T) *EventStreamConsumer[T] {
	return &EventStreamConsumer[T]{
		wg:          &sync.WaitGroup{},
		db:          db,
		url:         url,
		savedEvents: make(chan T),
		streamName:  instance.GetStreamEventHeader().StreamName,
	}
}
