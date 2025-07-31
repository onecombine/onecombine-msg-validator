package settlementfx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/onecombine/onecombine-msg-validator/src/partners"
	"github.com/onecombine/onecombine-msg-validator/src/utils"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/aws_msk_iam_v2"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type SettlementFxEvent struct {
	Pair     string `json:"pair"`
	Value    string `json:"value"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
}

type SettlementFXConsumer interface {
	Subscribe(wg *sync.WaitGroup) chan string
	Process(e *SettlementFxEvent) error
}

type fxConsumer struct {
	store   *partners.MemoryStore
	kreader *kafka.Reader
	cfg     *partners.KafkaConfig
}

// Process implements SettlementFXConsumer.
func (f *fxConsumer) Process(e *SettlementFxEvent) error {
	fx := &SettlementFX{
		Pair:     e.Pair,
		Value:    e.Value,
		Created:  e.Created,
		Modified: e.Modified,
	}
	f.store.Set(fx.Pair, fx)
	return nil
}

// Subscribe implements SettlementFXConsumer.
func (f *fxConsumer) Subscribe(wg *sync.WaitGroup) chan string {
	cls := make(chan string)
	topic := f.cfg.TopicName

	go func() {
		fmt.Printf("Subscribe to settlement fx event stream, topic: %s\n", topic)

		for {
			select {
			case <-cls:
				fmt.Printf("Stop subscribe topic: %s\n", topic)
				f.kreader.Close()
				wg.Done()
				close(cls)
				return
			default:
				fmt.Printf("Fetching message (topic: %s)\n", topic)
				msg, err := f.kreader.FetchMessage(context.TODO())
				if err != nil {
					fmt.Printf("Error fetch message from kafka (topic: %s), error: %v", topic, err)
					continue
				}

				var event SettlementFxEvent
				if err = json.Unmarshal(msg.Value, &event); err != nil {
					fmt.Printf("Unable to unmarshal event message (settlementFx), error: %v", err)
					f.kreader.CommitMessages(context.TODO(), msg)
					continue
				}

				if err = f.Process(&event); err != nil {
					fmt.Printf("Unable to process settlement fx event, error: %v", err)
				} else {
					fmt.Printf("Process settlement fx event successfully")
				}

				f.kreader.CommitMessages(context.TODO(), msg)
				time.Sleep(1000 * time.Millisecond)
			}
		}
	}()

	wg.Add(1)

	return cls
}

func NewKafkaSettlementFXConsumer(store *partners.MemoryStore, cfg *partners.KafkaConfig) SettlementFXConsumer {
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
	if offset = kafka.LastOffset; cfg.ReadOffset == "LATEST" {
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

	return &fxConsumer{
		kreader: reader,
		store:   store,
		cfg:     cfg,
	}
}
