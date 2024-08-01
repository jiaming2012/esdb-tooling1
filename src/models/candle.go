package models

import "time"

type CandleDTO struct {
	Symbol string `json:"symbol"`
	Open float64 `json:"o"`
	Close float64 `json:"c"`
	High float64 `json:"h"`
	Low float64 `json:"l"`
	Volume float64 `json:"v"`
	Timestamp int64 `json:"t"`
}

type Candle struct {
	Symbol string
	Open float64
	Close float64
	High float64
	Low float64
	Volume float64
	Timestamp time.Time
}