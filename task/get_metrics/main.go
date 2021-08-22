package main

import (
	"context"
	"github.com/aanrii/thermohydrometer-rpi/lib"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/getsentry/sentry-go"
	"golang.org/x/xerrors"
	"log"
	"time"
)

var db *dynamodb.DynamoDB

func init() {
	s, err := ReadSpec()
	if err != nil {
		log.Fatalf("Failed to read spec: %+v", err)
	}

	opts := sentry.ClientOptions{ Dsn: s.SentryEndpoint }
	if err := sentry.Init(opts); err != nil {
		log.Fatalf("Failed to initialize sentry: %+v", err)
	}

	se, err := session.NewSession()
	if err != nil {
		sentry.CaptureException(err)
		log.Fatalf("Failed to initialize AWS session: %+v", err)
	}
	db = dynamodb.New(se)
}

type handler func (context.Context, getMetricsRequest) ([]lib.Metrics, error)

func handleError(h handler) handler {
	return func(ctx context.Context, r getMetricsRequest) ([]lib.Metrics, error) {
		res, err := h(ctx, r)
		if err != nil {
			sentry.CaptureException(err)
		}
		return res, err
	}
}

type getMetricsRequest struct {
	From int64 `json:"from"`
	To int64 `json:"to"`
}

const defaultFrom = 4

func handle(ctx context.Context, r getMetricsRequest) ([]lib.Metrics, error) {
	to := time.Now()
	if r.To > 0 {
		to = time.Unix(r.To / int64(time.Second), r.To % int64(time.Second))
	}
	if to.After(time.Now()) {
		return nil, xerrors.New(`Query parameter "to" must be before now`)
	}
	from := to.Add(time.Hour * time.Duration(-defaultFrom))
	if r.From > 0 {
		from = time.Unix(r.From / int64(time.Second), r.From % int64(time.Second))
	}
	if from.After(to) {
		return nil, xerrors.New(`Query parameter "from" must be before "to"`)
	}

	return GetMetricsFromDB(ctx, from, to)
}

func main() {
	lambda.Start(handleError(handle))
}
