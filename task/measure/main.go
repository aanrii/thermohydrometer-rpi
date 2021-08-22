package main

import (
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	s, err := ReadSpec()
	if err != nil {
		log.Fatalf("Failed to read spec: %+v", err)
	}

	c := zap.NewProductionConfig()
	if s.Debug {
		c = zap.NewDevelopmentConfig()
	}
	l, err := c.Build()
	if err != nil {
		log.Fatalf("Failed to initialize a logger: %v", err)
	}
	defer l.Sync()

	so := sentry.ClientOptions{
		Dsn: s.SentryEndpoint,
	}
	if err := sentry.Init(so); err != nil {
		l.Fatal("Failed to initialize sentry", zap.Error(err))
	}
	defer sentry.Flush(5 * time.Second)

	if err := Task(s); err != nil {
		l.Error("Failed to execute task", zap.Error(err))
		sentry.CaptureException(err)
	}
}
