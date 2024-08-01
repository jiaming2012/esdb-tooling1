package services

import (
	"context"
	"fmt"
	"math"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"jiaming2012/esdb-tooling-1/src/models"
)

func ProduceMACrossoverEvents(ctx context.Context, cli *esdb.Client, count int) error {
	esdbProducer := models.NewEsdbProducer(cli)

	log.Infof("Producing %v MA crossover events", count)

	for i := 0; i < count; i++ {
		var state models.MovingAverageState
		if math.Mod(float64(i), 2) == 0 {
			state = models.MovingAverageStateAbove
		} else {
			state = models.MovingAverageStateBelow
		}

		candle := &models.MovingAverageCrossEvent{
			RequestID: uuid.New(),
			Symbol:    "AAPL",
			Period:    1440,
			State:     state,
		}

		if err := esdbProducer.SaveData(ctx, candle); err != nil {
			return fmt.Errorf("failed to save data: %v", err)
		}

		log.Infof("Produced MA crossover event: %s", candle.RequestID)
	}

	return nil
}
