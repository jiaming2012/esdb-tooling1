package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	log "github.com/sirupsen/logrus"
)

type eventStreamConsumer[T StreamEvent] struct {
	wg          *sync.WaitGroup
	db          *esdb.Client
	url         string
	mu          sync.Mutex
	savedEvents []T
	streamName  StreamName
}

func (cli *eventStreamConsumer[T]) replayEvents(ctx context.Context, stream StreamName, lastEventNumber uint64) error {
	if lastEventNumber == 0 {
		return nil
	}

	event, err := cli.db.ReadStream(ctx, string(stream), esdb.ReadStreamOptions{}, lastEventNumber)
	if err != nil {
		return fmt.Errorf("eventStoreDBClient: failed to read stream %s: %v", stream, err)
	}

	for {
		event, err := event.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("eventStoreDBClient: failed to read event from stream: %v", err)
		}

		if err := cli.processEvent(event.Event); err != nil {
			return fmt.Errorf("eventStoreDBClient: failed to process event: %v", err)
		}
	}

	return nil
}

func (cli *eventStreamConsumer[T]) subscribeToStream(ctx context.Context, wg *sync.WaitGroup, stream StreamName, initialEventNumber uint64) (chan error, error) {
	subscription, err := cli.db.SubscribeToStream(ctx, string(stream), esdb.SubscribeToStreamOptions{
		From: esdb.Revision(initialEventNumber),
	})

	if err != nil {
		return nil, fmt.Errorf("eventStoreDBClient: failed to subscribe to stream: %v", err)
	}

	log.Infof("esdbConsumer: subscribed to stream %s", stream)

	lastEventNumber := initialEventNumber

	errCh := make(chan error)

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			for {
				if subscription == nil {
					log.Debugf("subscription is nil")
					break
				}

				event := subscription.Recv()

				if event.SubscriptionDropped != nil {
					log.Infof("Subscription dropped: %v", event.SubscriptionDropped.Error)
					break
				}

				if event.EventAppeared == nil {
					continue
				}

				if event.CheckPointReached != nil {
					log.Infof("checkpoint reached: %v\n", event.CheckPointReached)
				}

				ev := event.EventAppeared.Event

				lastEventNumber = event.EventAppeared.OriginalEvent().EventNumber

				if err := cli.processEvent(ev); err != nil {
					errCh <- fmt.Errorf("eventStoreDBClient: failed to process event: %v", err)
					return
				}
			}

			log.Infof("re-subscribing subscription @ pos %v", lastEventNumber)

			subscription, err = cli.db.SubscribeToStream(ctx, string(stream), esdb.SubscribeToStreamOptions{
				From: esdb.Revision(lastEventNumber),
			})

			if err != nil {
				log.Errorf("eventStoreDBClient: failed to subscribe to stream: %v", err)
				return
			}
		}
	}()

	return errCh, nil
}

func (cli *eventStreamConsumer[T]) processEvent(event *esdb.RecordedEvent) error {
	var savedEvent T

	if err := json.Unmarshal(event.Data, &savedEvent); err != nil {
		return fmt.Errorf("esdbConsumer.processEvent: failed to unmarshal event data: %v", err)
	}

	cli.mu.Lock()
	cli.savedEvents = append(cli.savedEvents, savedEvent)
	cli.mu.Unlock()

	return nil
}

func (cli *eventStreamConsumer[T]) Start(ctx context.Context, wg *sync.WaitGroup) {
	settings, err := esdb.ParseConnectionString(cli.url)
	if err != nil {
		log.Panicf("failed to parse connection string: %v", err)
	}

	cli.db, err = esdb.NewClient(settings)
	if err != nil {
		log.Panicf("failed to create client: %v", err)
	}

	lastEventNumber, err := FindStreamLastEventNumber(ctx, cli.db, cli.streamName)
	if err != nil {
		log.Panicf("eventStoreDBClient: failed to find last event number: %v", err)
	}

	if err := cli.replayEvents(ctx, cli.streamName, lastEventNumber); err != nil {
		log.Panicf("eventStoreDBClient: failed to replay events: %v", err)
	}

	if _, err := cli.subscribeToStream(ctx, wg, cli.streamName, lastEventNumber); err != nil {
		log.Panicf("eventStoreDBClient: failed to subscribe to stream: %v", err)
	}
}

func NewEventStreamConsumer[T StreamEvent](db *esdb.Client, url string, instance T) *eventStreamConsumer[T] {
	return &eventStreamConsumer[T]{
		wg:          &sync.WaitGroup{},
		db:          db,
		url:         url,
		savedEvents: make([]T, 0),
		streamName:  instance.GetStreamEventHeader().StreamName,
	}
}
