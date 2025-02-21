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

/* for reference
{
	"id":2,
	"acqId":"100090",
	"name":"Legacy Mock Acquirer",
	"description":"legacy mock acquirer for integration testing",
	"apiKey":"ABCD-ABCD-ABCD",
	"secret":"aaaa",
	"hook":"",
	"orgId":4,
	"settlement_fee":"0.60",
	"settlement_type":"PCT",
	"settlement_waived":false,
	"switching_fee":"0.61",
	"switching_type":"ABS",
	"switching_waived":true,
	"settlement_currency_code":"THB",
	"settlement_report_bucket":"aws-1cb-th-dev-s3-sandbox-acquirer-report-100090",
	"created":"","modified":"2025-02-21T06:39:00Z"}

*/

type AcquirerProfileEvent struct {
	ID                     uint   `json:"id"`
	AcqID                  string `json:"acqId"`
	Name                   string `json:"name"`
	Description            string `json:"description"`
	ApiKey                 string `json:"apiKey"`
	Secret                 string `json:"secret"`
	NotificationHook       string `json:"hook"`
	OrganizationID         uint   `json:"orgId"`
	SettlementFee          string `json:"settlement_fee"`
	SettlementType         string `json:"settlement_type"`
	SettlementWaived       bool   `json:"settlement_waived"`
	SwitchingFee           string `json:"switching_fee"`
	SwitchingType          string `json:"switching_type"`
	SwitchingWaived        bool   `json:"switching_waived"`
	SettlementCurrencyCode string `json:"settlement_currency_code"`
	SettlementReportBucket string `json:"settlement_report_bucket"`
	Created                string `json:"created"`
	Modified               string `json:"modified"`
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
	a.store.Set(e.AcqID, eventToAcquirerProfile(e))
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
	return &acquirerConsumer{
		kreader: reader,
		store:   store,
		cfg:     cfg,
	}
}

func eventToAcquirerProfile(e *AcquirerProfileEvent) *AcquirerProfile {
	return &AcquirerProfile{
		ID:                     e.ID,
		AcqID:                  e.AcqID,
		Name:                   e.Name,
		Description:            e.Description,
		ApiKey:                 e.ApiKey,
		Secret:                 e.Secret,
		OrganizationID:         e.OrganizationID,
		SettlementFee:          e.SettlementFee,
		SettlementType:         e.SettlementType,
		SettlementWaived:       e.SettlementWaived,
		SwitchingFee:           e.SwitchingFee,
		SwitchingType:          e.SwitchingType,
		SwitchingWaived:        e.SettlementWaived,
		SettlementCurrencyCode: e.SettlementCurrencyCode,
		SettlementReportBucket: e.SettlementReportBucket,
		Created:                e.Created,
		Modified:               e.Modified,
	}
}
