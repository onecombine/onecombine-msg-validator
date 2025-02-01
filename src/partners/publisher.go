package partners

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
	Bootstrap                string
	Username                 string
	Password                 string
	IssuerProfileTopicName   string
	AcquirerProfileTopicName string
	QueueType                string
	Timeout                  string
}

type EventPublisher struct {
	a *kafka.Writer
	i *kafka.Writer
}

func (p *EventPublisher) PublishAcquirerProfileChangedEvent(ctx context.Context, acq *AcquirerProfile) error {

	val, err := json.Marshal(acq)
	if err != nil {
		fmt.Printf("Unable to marshal acquirer profile for publish to kafka, error: %v\n", err)
		return err
	}
	msg := kafka.Message{
		Key:   []byte(acq.AcqID),
		Value: val,
	}
	err = p.a.WriteMessages(ctx, msg)
	if err != nil {
		fmt.Printf("Unable to publish acquirer profile to kafka, error: %v\n", err)
		return err
	}

	return nil
}

func (p *EventPublisher) PublishIssuerProfileChangedEvent(ctx context.Context, iss *IssuerProfile) error {

	val, err := json.Marshal(iss)
	if err != nil {
		fmt.Printf("Unable to marshal issuer profile for publish to kafka, error: %v\n", err)
		return err
	}
	msg := kafka.Message{
		Key:   []byte(iss.IssuerID),
		Value: val,
	}
	err = p.i.WriteMessages(ctx, msg)
	if err != nil {
		fmt.Printf("Unable to publish issuer profile to kafka, error: %v\n", err)
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

	acqKafkaConfig := kafka.WriterConfig{
		Brokers:      hosts,
		Topic:        cfg.AcquirerProfileTopicName,
		Balancer:     &kafka.Hash{},
		Dialer:       dialer,
		WriteTimeout: writeTimeout,
		//Logger:       kafka.LoggerFunc(logInfo),
		//ErrorLogger: kafka.LoggerFunc(logError),
	}

	issKafkaConfig := kafka.WriterConfig{
		Brokers:      hosts,
		Topic:        cfg.IssuerProfileTopicName,
		Balancer:     &kafka.Hash{},
		Dialer:       dialer,
		WriteTimeout: writeTimeout,
		//Logger:       kafka.LoggerFunc(logInfo),
		//ErrorLogger: kafka.LoggerFunc(logError),
	}

	return &EventPublisher{
		i: kafka.NewWriter(issKafkaConfig),
		a: kafka.NewWriter(acqKafkaConfig),
	}
}
