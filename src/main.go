package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	log "github.com/sirupsen/logrus"

	"jiaming2012/esdb-tooling-1/src/models"
)

func generateRandomCandles(*esdb.Client) {

}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Create channel for shutdown signals.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	url := "esdb://localhost:2113?tls=false"

	settings, err := esdb.ParseConnectionString(url)
	if err != nil {
		log.Panicf("failed to parse connection string: %v", err)
	}

	esdbClient, err := esdb.NewClient(settings)
	if err != nil {
		log.Fatalf("failed to create EventStoreDB client: %v", err)
	}

	esdbConsumer := models.NewEventStreamConsumer(esdbClient, url, &models.MovingAverageCrossEvent{})

	esdbConsumer.Start(ctx, wg)

	time.Sleep(2 * time.Second)

	generateRandomCandles(esdbClient)

	// Block here until program is shut down
	<-stop

	// Shut down event client
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Info("main: gracefully shut down")
}
