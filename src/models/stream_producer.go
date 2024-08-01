package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	log "github.com/sirupsen/logrus"
)

type EsdbProducer struct {
	wg              *sync.WaitGroup
	db              *esdb.Client
	url             string
	lastEventNumber uint64
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
		return errors.New("db is nil")
	}

	// todo: verify that the stream is thread safe
	_, err := cli.db.AppendToStream(ctx, string(streamName), esdb.AppendToStreamOptions{}, eventData)
	if err != nil {
		return fmt.Errorf("failed to append event to stream: %w", err)
	}

	return nil
}

func (cli *EsdbProducer) insertData(ctx context.Context, event StreamEvent, data map[string]interface{}) error {
	// set the schema version
	eventHeader := event.GetStreamEventHeader()
	schemaVersion := event.GetStreamEventHeader().SchemaVersion
	eventHeader.SetSchemaVersion(schemaVersion)

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	header := event.GetStreamEventHeader()

	log.Debugf("saving %s to stream %s ...", header.EventType, header.StreamName)

	if err := cli.insertEvent(ctx, header.EventType, header.StreamName, nil, bytes); err != nil {
		return fmt.Errorf("EsdbProducer: failed to insert event: %w", err)
	}

	return nil
}

func (cli *EsdbProducer) SaveData(ctx context.Context, event StreamEvent, data map[string]interface{}) error {
	return cli.insertData(ctx, event, data)
}
