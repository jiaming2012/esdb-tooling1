package models

import (
	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	log "github.com/sirupsen/logrus"
)

func NewEsdbClient(url string) (*esdb.Client, error) {
	settings, err := esdb.ParseConnectionString(url)
	if err != nil {
		log.Panicf("failed to parse connection string: %v", err)
	}

	settings.Logger = func(level esdb.LogLevel, format string, args ...interface{}) {}

	return esdb.NewClient(settings)
}
