package main

import (
	"golang.org/x/xerrors"
)

func Task(s *Spec) error {
	tlsc, err := TLSConfig(s.RootCAPath, s.CertPath, s.PrivateKeyPath)
	if err != nil {
		return xerrors.Errorf("Failed to initialize a tls config: %w", err)
	}

	client, err := NewDeviceShadowClient(&DeviceShadowClientConfig{
		TLS:             tlsc,
		BrokerEndpoint:  s.BrokerEndpoint,
		ClientID:        s.ThingName,
		ThingName:       s.ThingName,
		ConnectTimeout:  s.ConnectTimeout,
		RequestTimeout:  s.RequestTimeout,
		ResponseTimeout: s.ResponseTimeout,
	})
	if err != nil {
		return xerrors.Errorf("Failed to initialize a device shadow client: %w", err)
	}
	defer client.Close()

	m, err := NewThermoHydrometer(s.PinName)
	if err != nil {
		return xerrors.Errorf("Failed to initialize a thermo hydrometer: %w", err)
	}
	ms, err := m.Measure(s.MaxMeasureAttempts)
	if err != nil {
		return xerrors.Errorf("Failed to measure temperature and humidity: %w", err)
	}

	ds := NewDeviceShadow(ms)
	if err := client.Upload(ds); err != nil {
		return xerrors.Errorf("Failed to upload the device shadow: %w", err)
	}

	return nil
}
