package main

import (
	"context"
	"flag"
	"fulfillment-entitlements-api/internal/config"
	"fulfillment-entitlements-api/internal/dependencies"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Producer struct {
	Dependencies *dependencies.Dependencies
}

func (p *Producer) HandleRequest(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	dep := p.Dependencies
	for _, record := range kinesisEvent.Records {
		dep.GetLogger().Infof("received message from kinesis. partition key: %s", record.Kinesis.PartitionKey)
		data := record.Kinesis.Data
		recordDataText := string(data)
		dep.GetLogger().Infof("data: %v", recordDataText)
	}

	return nil
}

func main() {
	configLocation := flag.String("config", "internal/config/config.json", "path to the config file")
	flag.Parse()

	cfg := &config.Config{}

	if err := cfg.Load(context.TODO(), *configLocation); err != nil {
		panic(err)
	}

	dep := dependencies.New(cfg)
	if err := dep.Initialize(); err != nil {
		panic(err)
	}

	producer := &Producer{
		Dependencies: dep,
	}

	lambda.Start(producer.HandleRequest)
}
