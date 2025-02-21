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

type KafkaConfig struct {
	Bootstrap       string
	Username        string
	Password        string
	TopicName       string
	ConsumerGroupID string
	QueueType       string
	Timeout         string
	ReadOffset      string
}

/*
{"id":4,
"issuer_id":"200099",
"name":"mock issuer for test",
"description":"Mock issuer",
"apiKey":"0439BA10-205B-46AE-9963-1FB22642D8AF",
"secret":"6dc010a7952422a",
"orgId":4,
"fx_name":"RUBTHB",
"fx_value":"0.35",
"settlement_fee":"0.6",
"settlement_type":"PCT",
"settlement_waived":false,
"switching_fee":"0.65",
"switching_type":"ABS",
"switching_waived":false,
"settlement_currency_code":"RUB",
"settlement_report_bucket":"aws-1cb-th-dev-s3-sandbox-issuer-report-200099",
"created":"2025-01-01T00:00:00Z",
"modified":"2025-02-21T05:21:36Z"}

*/

type IssuerProfileEvent struct {
	ID                        uint   `json:"id"`
	IssuerID                  string `json:"issuer_id"`
	Name                      string `json:"name"`
	Description               string `json:"description"`
	ApiKey                    string `json:"apiKey"`
	Secret                    string `json:"secret"`
	OrganizationID            uint   `json:"orgId"`
	FxName                    string `json:"fx_name"`
	FXValue                   string `json:"fx_value"`
	SettlementFee             string `json:"settlement_fee"`
	SettlementFeeType         string `json:"settlement_type"`
	SettlementFeeWaived       bool   `json:"settlement_waived"`
	SwitchingFee              string `json:"switching_fee"`
	SwitchingFeeType          string `json:"switching_type"`
	SwitchingFeeWaived        bool   `json:"switching_waived"`
	SettlementCurrencyCode    string `json:"settlement_currency_code"`
	SettlementReportBucket    string `json:"settlement_report_bucket"`
	RefundNotificationWebHook string `json:"refund_notification_web_hook"`
	Created                   string `json:"created"`
	Modified                  string `json:"modified"`
}

type IssuerProfileConsumer interface {
	Subscribe(wg *sync.WaitGroup) chan string
	Process(e *IssuerProfileEvent) error
}

type issuerConsumer struct {
	store   *MemoryStore
	kreader *kafka.Reader
	cfg     *KafkaConfig
}

// Process implements IssuerProfileConsumer.
func (i *issuerConsumer) Process(e *IssuerProfileEvent) error {

	var key string
	for k, v := range i.store.GetAll() {
		if v.(*IssuerProfile).IssuerID == e.IssuerID {
			key = k
			break
		}
	}

	i.store.Set(key, eventToIssuerProfile(e))
	fmt.Printf("Update issuer profile in memory storage (IssuerID: %s)\n", e.IssuerID)
	return nil
}

// Subscribe implements IssuerProfileConsumer.
func (i *issuerConsumer) Subscribe(wg *sync.WaitGroup) chan string {

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

				var event IssuerProfileEvent
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
					fmt.Printf("Process issuer profile (id: %s) successfully\n", event.IssuerID)
				}

				i.kreader.CommitMessages(context.TODO(), msg)
				time.Sleep(100 * time.Millisecond)
			}
		}

	}()

	wg.Add(1)
	return cls
}

func NewKafkaIssuerProfileConsumer(store *MemoryStore, cfg *KafkaConfig) IssuerProfileConsumer {
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
	return &issuerConsumer{
		kreader: reader,
		store:   store,
		cfg:     cfg,
	}
}

func eventToIssuerProfile(e *IssuerProfileEvent) *IssuerProfile {
	return &IssuerProfile{
		ID:                        e.ID,
		IssuerID:                  e.IssuerID,
		Name:                      e.Name,
		Description:               e.Description,
		ApiKey:                    e.ApiKey,
		Secret:                    e.Secret,
		OrganizationID:            e.OrganizationID,
		FXName:                    e.FxName,
		FXValue:                   e.FXValue,
		SettlementFee:             e.SettlementFee,
		SettlementType:            e.SettlementFeeType,
		SettlementWaived:          e.SettlementFeeWaived,
		SwitchingFee:              e.SwitchingFee,
		SwitchingType:             e.SwitchingFeeType,
		SwitchingWaived:           e.SettlementFeeWaived,
		SettlementCurrencyCode:    e.SettlementCurrencyCode,
		SettlementReportBucket:    e.SettlementReportBucket,
		RefundNotificationWebHook: e.RefundNotificationWebHook,
		Created:                   e.Created,
		Modified:                  e.Modified,
	}
}

func printDebug(e *IssuerProfile) {
	fmt.Println("====================== DEBUG  =========================")
	fmt.Printf("IssuerID: %s", e.IssuerID)
	fmt.Printf("Settlement Fee: %s", e.SettlementFee)
	// TODO: complete all fields
	fmt.Println("=======================================================")
}
