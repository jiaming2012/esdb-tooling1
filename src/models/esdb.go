package models

import (
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
)

func NewEsdbClient(url string) (*esdb.Client, error) {
	settings, err := esdb.ParseConnectionString(url)
	if err != nil {
		return nil, fmt.Errorf("NewEsdbClient: failed to parse connection string: %v", err)
	}

	settings.Logger = func(level esdb.LogLevel, format string, args ...interface{}) {}

	return esdb.NewClient(settings)
}
