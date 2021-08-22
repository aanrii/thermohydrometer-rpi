package main

import (
	"context"
	"github.com/aanrii/thermohydrometer-rpi/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"golang.org/x/xerrors"
	"strconv"
	"time"
)

const maxKeys = 100

func interval(from, to time.Time) time.Duration {
	r := to.Sub(from)
	if r / (time.Minute * 10) < maxKeys {
		return time.Minute * 10
	}
	return r / maxKeys
}

func round10Minutes(t time.Time) time.Time {
	minutes := (t.Minute() / 10) * 10
	return time.Date(
		t.Year(), t.Month(), t.Day(), t.Hour(), minutes, 0, 0, t.Location())
}

func keys(from, to time.Time) []map[string]*dynamodb.AttributeValue {
	attrs := make([]map[string]*dynamodb.AttributeValue, 0)

	i := interval(from, to)
	for t := from; t.Before(to); t = t.Add(i) {
		roundedT := round10Minutes(t)
		roundedTStr := strconv.Itoa(int(roundedT.UnixNano()))
		attrs = append(attrs, map[string]*dynamodb.AttributeValue{
			"measured_at_minutes": { N: &roundedTStr },
		})
	}

	return attrs
}

func GetMetricsFromDB(ctx context.Context, from, to time.Time) ([]lib.Metrics, error) {
	initialInput := dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			lib.TableName: {
				Keys: keys(from, to),
				ProjectionExpression: aws.String(lib.AllColumnsProjectionExpr),
			},
		},
	}

	items := make([]map[string]*dynamodb.AttributeValue, 0)
	for input := initialInput;; {
		output, err := db.BatchGetItemWithContext(ctx, &input)
		if err != nil {
			return nil, xerrors.Errorf("Failed from get items: %w", err)
		}
		items = append(items, output.Responses[lib.TableName]...)

		if len(output.UnprocessedKeys) == 0 {
			break
		}
		input = dynamodb.BatchGetItemInput{ RequestItems: output.UnprocessedKeys }
	}

	mArr := make([]lib.Metrics, len(items))
	for i, item := range items {
		var m lib.Metrics
		if err := dynamodbattribute.UnmarshalMap(item, &m); err != nil {
			return nil, xerrors.Errorf("Failed from unmarshal the response: %w", err)
		}
		mArr[i] = m
	}

	return mArr, nil
}