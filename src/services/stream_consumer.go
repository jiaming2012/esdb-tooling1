package services

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"jiaming2012/esdb-tooling-1/src/models"
)

func ConsumeMACrossoverEvents(consumer *models.EventStreamConsumer[*models.MovingAverageCrossEvent], wg *sync.WaitGroup) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		eventCh := consumer.GetEvents()
		for ev := range eventCh {
			meta := ev.GetStreamEventMeta()

			if meta.IsReplay {
				continue
			}

			log.Infof("Received event: RequestID:%s, State: %s", ev.RequestID, ev.State)
		}

		log.Info("esdbConsumer: stopped consuming events")
	}()
}
