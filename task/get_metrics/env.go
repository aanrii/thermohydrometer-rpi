package main

import (
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/xerrors"
)

type Spec struct {
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