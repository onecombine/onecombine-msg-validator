package utils

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/aws_msk_iam_v2"
)

const QUEUE_TOPIC string = "QUEUE_TOPIC"
const QUEUE_HOST string = "QUEUE_HOST"
const QUEUE_PUBLISH_TIMEOUT string = "QUEUE_PUBLISH_TIMEOUT"
const QUEUE_SUBSCRIBE_TIMEOUT string = "QUEUE_SUBSCRIBE_TIMEOUT"
const QUEUE_CONSUMERGROUP_ID string = "QUEUE_CONSUMERGROUP_ID"
const QUEUE_READOFFSET string = "QUEUE_READOFFSET" // EARLIEST, LATEST

const QUEUE_MODE_PUBLISHER string = "QUEUE_PUB"
const QUEUE_MODE_SUBSCRIBER string = "QUEUE_SUB"

type IKafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type IKafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type Queue struct {
	KafkaReader *IKafkaReader
	KafkaWriter *IKafkaWriter
}

type QueueMessage struct {
	WebHookUrl string `json:"webhook_url" binding:"required"`
	Data       string `json:"data" binding:"required"`
}

type QueueMessageConsumer interface {
	ProcessMessage(msg QueueMessage)
}

var QueueReaderConnect = kafka.NewReader
var QueueWriterConnect = kafka.NewWriter

func createReader(config kafka.ReaderConfig) interface{} {
	return QueueReaderConnect(config)
}

func createWriter(config kafka.WriterConfig) interface{} {
	return QueueWriterConnect(config)
}

var CreateReader = createReader
var CreateWriter = createWriter

func NewQueue(mode string) *Queue {
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil
	}

	topic := GetEnv(QUEUE_TOPIC, "notification-queue")
	hosts := strings.Split(GetEnv(QUEUE_HOST, ""), ",")
	group := GetEnv(QUEUE_CONSUMERGROUP_ID, "CG-0")
	publishTimeout, _ := time.ParseDuration(GetEnv(QUEUE_PUBLISH_TIMEOUT, "20s"))
	subscribeTimeout, _ := time.ParseDuration(GetEnv(QUEUE_SUBSCRIBE_TIMEOUT, "20s"))
	var offset int64
	if offset = kafka.LastOffset; GetEnv(QUEUE_READOFFSET, "LATEST") == "EARLIEST" {
		offset = kafka.FirstOffset
	}

	var queue Queue
	switch mode {
	case QUEUE_MODE_SUBSCRIBER:
		{
			kafkaConfig := kafka.ReaderConfig{
				Brokers:          hosts,
				GroupID:          group,
				Topic:            topic,
				ReadBatchTimeout: subscribeTimeout,
				StartOffset:      offset,
			}
			kafkaConfig.Dialer = &kafka.Dialer{
				DualStack:     false,
				SASLMechanism: aws_msk_iam_v2.NewMechanism(awsConfig),
				TLS:           &tls.Config{},
			}
			reader := CreateReader(kafkaConfig).(IKafkaReader)
			queue.KafkaReader = &reader
		}
	case QUEUE_MODE_PUBLISHER:
		{
			kafkaConfig := kafka.WriterConfig{
				Brokers:      hosts,
				Topic:        topic,
				WriteTimeout: publishTimeout,
				Async:        true,
			}
			kafkaConfig.Dialer = &kafka.Dialer{
				DualStack:     false,
				SASLMechanism: aws_msk_iam_v2.NewMechanism(awsConfig),
				TLS:           &tls.Config{},
			}
			writer := CreateWriter(kafkaConfig).(IKafkaWriter)
			queue.KafkaWriter = &writer
		}
	}

	if err != nil {
		log.Fatal("Queue: failed to dial leader:", err)
	}

	return &queue
}

func (queue Queue) Publish(ctx context.Context, msg QueueMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = (*queue.KafkaWriter).WriteMessages(ctx, kafka.Message{Value: raw})
	return err
}

func (queue Queue) Subscribe(ctx context.Context, consumer QueueMessageConsumer) {
	for {
		m, err := (*queue.KafkaReader).FetchMessage(ctx)
		fmt.Println("KReader :", m)
		if err != nil {
			return
		}

		if len(m.Value) == 0 {
			continue
		}

		var message QueueMessage
		err = json.Unmarshal(m.Value, &message)
		if err != nil {
			log.Printf("%v\n", err)
		}

		consumer.ProcessMessage(message)
		err = (*queue.KafkaReader).CommitMessages(ctx, m)
		if err != nil {
			log.Printf("%v\n", err)
		}
	}
}

func (queue Queue) Close() error {
	err := (*queue.KafkaReader).Close()
	return err
}
