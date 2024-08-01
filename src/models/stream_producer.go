package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	log "github.com/sirupsen/logrus"
)

type EsdbProducer struct {
	db *esdb.Client
}

func (cli *EsdbProducer) insertEvent(ctx context.Context, eventType EventType, streamName StreamName, meta []byte, data []byte) error {
	eventData := esdb.EventData{
		ContentType: esdb.ContentTypeJson,
		EventType:   string(eventType),
		Data:        data,
	}

	if meta != nil {
		eventData.Metadata = meta
	}

	if cli.db == nil {
		return fmt.Errorf("EsdbProducer.insertEvent: db is nil")
	}

	// todo: verify that the stream is thread safe
	_, err := cli.db.AppendToStream(ctx, string(streamName), esdb.AppendToStreamOptions{}, eventData)
	if err != nil {
		return fmt.Errorf("EsdbProducer.insertEvent: failed to append event to stream: %w", err)
	}

	return nil
}

func (cli *EsdbProducer) insertData(ctx context.Context, event StreamEvent) error {
	header := event.GetStreamEventHeader()

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	log.Debugf("saving %s to stream %s ...", header.EventType, header.StreamName)

	if err := cli.insertEvent(ctx, header.EventType, header.StreamName, nil, bytes); err != nil {
		return fmt.Errorf("EsdbProducer.insertData: failed to insert event: %w", err)
	}

	return nil
}

func (cli *EsdbProducer) SaveData(ctx context.Context, event StreamEvent) error {
	return cli.insertData(ctx, event)
}

func NewEsdbProducer(db *esdb.Client) *EsdbProducer {
	return &EsdbProducer{
		db: db,
	}
}
