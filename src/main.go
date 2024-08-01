package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"jiaming2012/esdb-tooling-1/src/models"
	"jiaming2012/esdb-tooling-1/src/services"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Create channel for shutdown signals.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	// Create a new EventStoreDB client
	url := "esdb://localhost:2113?tls=false"
	esdbClient, err := models.NewEsdbClient(url)
	if err != nil {
		log.Fatalf("failed to create EventStoreDB client: %v", err)
	}

	// Create a new event stream consumer
	streamConsumer := models.NewEventStreamConsumer(esdbClient, url, &models.MovingAverageCrossEvent{})

	// Start a goroutine to consume events
	services.ConsumeMACrossoverEvents(streamConsumer, wg)

	if err := streamConsumer.Start(ctx, wg); err != nil {
		log.Fatalf("failed to start consumer: %v", err)
	}

	// Produce some MA crossover events
	if err := services.ProduceMACrossoverEvents(ctx, esdbClient, 4); err != nil {
		log.Fatalf("failed to produce MA crossover events: %v", err)
	}

	log.Info("main: waiting for shutdown signal: press Ctrl+C to stop")

	// Block here until program is shut down
	<-stop

	// Shut down event client
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Info("main: gracefully shut down")
}
