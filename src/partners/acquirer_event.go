package partners

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/onecombine/onecombine-msg-validator/src/utils"
	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/aws_msk_iam_v2"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type AcquirerProfileEvent struct {
	AcquirerID string
}

type AcquirerProfileConsumer interface {
	Subscribe(wg *sync.WaitGroup) chan string
	Process(e *AcquirerProfileEvent) error
}

type acquirerConsumer struct {
	store   *MemoryStore
	kreader *kafka.Reader
	cfg     *KafkaConfig
}

// Process implements IssuerProfileConsumer.
func (a *acquirerConsumer) Process(e *AcquirerProfileEvent) error {
	a.store.Set(e.AcquirerID, e)
	return nil
}

// Subscribe implements IssuerProfileConsumer.
func (i *acquirerConsumer) Subscribe(wg *sync.WaitGroup) chan string {

	cls := make(chan string)
	topic := i.cfg.TopicName

	go func() {
		fmt.Printf("Subscribe to issuer profile event stream, topic: %s\n", topic)

		for {
			select {
			case <-cls:
				fmt.Printf("Stop subscribe topic: %s\n", i.cfg.TopicName)
				i.kreader.Close()
				wg.Done()
				close(cls)
				return
			default:
				fmt.Println("Fetching message")
				msg, err := i.kreader.FetchMessage(context.TODO())
				if err != nil {
					fmt.Printf("Error fetch message from kafka, error: %v\n", err)
					continue
				}

				var event AcquirerProfileEvent
				err = json.Unmarshal(msg.Value, &event)
				if err != nil {
					fmt.Printf("Unable to unmarshal event message, error: %v\n", err)
					i.kreader.CommitMessages(context.TODO(), msg)
					continue
				}

				err = i.Process(&event)
				if err != nil {
					fmt.Printf("Unable to process profile event, error: %v\n", err)
				} else {
					fmt.Printf("Process acquirer profile (id: %s) successfully\n", event.AcquirerID)
				}

				i.kreader.CommitMessages(context.TODO(), msg)
				time.Sleep(100 * time.Millisecond)
			}
		}

	}()

	wg.Add(1)
	return cls
}

func NewKafkaAcquirerProfileConsumer(store *MemoryStore, cfg *KafkaConfig) AcquirerProfileConsumer {
	hosts := strings.Split(cfg.Bootstrap, ",")

	var dialer *kafka.Dialer
	switch cfg.QueueType {
	case utils.MSK:
		awsConfig, err := config.LoadDefaultConfig(context.Background())
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

	var offset int64
	if offset = kafka.LastOffset; cfg.ReadOffset == "EARLIEST" {
		offset = kafka.LastOffset
	}

	var timeout string
	if timeout = cfg.Timeout; cfg.Timeout == "" {
		timeout = "20s"
	}

	subscribeTimeout, _ := time.ParseDuration(timeout)

	kafkaConfig := kafka.ReaderConfig{
		Brokers:          hosts,
		GroupID:          cfg.ConsumerGroupID,
		Topic:            cfg.TopicName,
		ReadBatchTimeout: subscribeTimeout,
		StartOffset:      offset,
	}

	kafkaConfig.Dialer = dialer
	reader := kafka.NewReader(kafkaConfig)
	return &acquirerConsumer{
		kreader: reader,
		store:   store,
	}
}
