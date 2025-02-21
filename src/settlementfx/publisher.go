package settlementfx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/onecombine/onecombine-msg-validator/src/utils"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/aws_msk_iam_v2"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type EventPublisherConfig struct {
	Bootstrap string
	Username  string
	Password  string
	TopicName string
	QueueType string
	Timeout   string
}

type EventPublisher struct {
	p *kafka.Writer
}

func (p *EventPublisher) PublishSettlementFXChangedEvent(ctx context.Context, fx *SettlementFX) error {
	val, err := json.Marshal(fx)
	if err != nil {
		fmt.Printf("Unable to marshal settlement fx data for publish, error: %v", err)
		return err
	}

	msg := kafka.Message{
		Key:   []byte(fx.Pair),
		Value: val,
	}

	if err = p.p.WriteMessages(ctx, msg); err != nil {
		fmt.Printf("Unable to publish settlement fx change event, error: %v", err)
		return err
	}

	return nil
}

func NewEventPublisher(cfg *EventPublisherConfig) *EventPublisher {
	var dialer *kafka.Dialer

	switch cfg.QueueType {
	case utils.MSK:
		awsConfig, err := awsConfig.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil
		}
		dialer = &kafka.Dialer{
			DualStack:     false,
			SASLMechanism: aws_msk_iam_v2.NewMechanism(awsConfig),
		}
	case utils.SASL_PLAIN:
		dialer = &kafka.Dialer{
			DualStack: false,
			SASLMechanism: plain.Mechanism{
				Username: cfg.Username,
				Password: cfg.Password,
			},
		}
	}

	hosts := strings.Split(cfg.Bootstrap, ",")
	writeTimeout, _ := time.ParseDuration(cfg.Timeout)

	kafkaConfig := kafka.WriterConfig{
		Brokers:      hosts,
		Topic:        cfg.TopicName,
		Balancer:     &kafka.Hash{},
		Dialer:       dialer,
		WriteTimeout: writeTimeout,
		//Logger:       kafka.LoggerFunc(logInfo),
		//ErrorLogger: kafka.LoggerFunc(logError),
	}

	return &EventPublisher{
		p: kafka.NewWriter(kafkaConfig),
	}

}
