package lib

import "time"

type Metrics struct {
	Temperature float64 `json:"temperature" dynamodbav:"temperature"`
	Humidity float64 `json:"humidity" dynamodbav:"humidity"`
	MeasuredAt int64 `json:"measured_at" dynamodbav:"measured_at"`
	MeasuredAtMinutes int64 `json:"measured_at_minutes" dynamodbav:"measured_at_minutes"`
}

func NewMetrics(temperature, humidity float64, measuredAt time.Time) *Metrics {
	measuredAtMinutes := time.Date(
		measuredAt.Year(), measuredAt.Month(), measuredAt.Day(), measuredAt.Hour(), (measuredAt.Minute() / 10) * 10, 0, 0,
		measuredAt.Location())
	return &Metrics{
		Temperature: temperature,
		Humidity: humidity,
		MeasuredAt: measuredAt.UnixNano(),
		MeasuredAtMinutes: measuredAtMinutes.UnixNano(),
	}
}