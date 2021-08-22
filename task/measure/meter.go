package main

import (
	"time"

	"github.com/MichaelS11/go-dht"
	"github.com/aanrii/thermohydrometer-rpi/lib"
	"golang.org/x/xerrors"
)

type ThermoHydrometer struct {
	d *dht.DHT
}

func NewThermoHydrometer(pinName string) (*ThermoHydrometer, error) {
	if err := dht.HostInit(); err != nil {
		return nil, xerrors.Errorf("Failed to initialize a host: %w", err)
	}

	d, err := dht.NewDHT(pinName, dht.Celsius, "DHT22")
	if err != nil {
		return nil, xerrors.Errorf("Failed to initialize a DHT22: %w", err)
	}

	return &ThermoHydrometer{d: d}, nil
}

func (m *ThermoHydrometer) Measure(maxAttempts uint) (*lib.Metrics, error) {
	var lastErr error
	for i := uint(0); i < maxAttempts; i++ {
		h, t, err := m.d.Read()
		if err == nil {
			return lib.NewMetrics(t, h, time.Now()), nil
		}
		lastErr = err
	}
	return nil, xerrors.Errorf("The number of attempts reading values reaches its limit. Last error: %w", lastErr)
}
