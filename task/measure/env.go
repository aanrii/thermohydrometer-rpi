package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/xerrors"
)

type Spec struct {
	Debug bool `default:"false"`

	RootCAPath     string `split_words:"true" default:"certs/AmazonRootCA1.cer"`
	CertPath       string `split_words:"true" default:"certs/certificate.pem.crt"`
	PrivateKeyPath string `split_words:"true" default:"certs/private.pem.key"`

	PinName            string `split_words:"true" default:"GPIO26"`
	MaxMeasureAttempts uint   `split_words:"true" default:"10"`

	BrokerEndpoint  string        `split_words:"true" required:"true"`
	ThingName       string        `split_words:"true" required:"true"`
	ConnectTimeout  time.Duration `split_words:"true" default:"10s"`
	RequestTimeout  time.Duration `split_words:"true" default:"5s"`
	ResponseTimeout time.Duration `split_words:"true" default:"10s"`

	SentryEndpoint string `split_words:"true" required:"true"`
}

func ReadSpec() (*Spec, error) {
	var s Spec

	err := envconfig.Process("", &s)
	if err != nil {
		return nil, xerrors.Errorf("Failed to read spec values: %w", err)
	}

	return &s, nil
}
