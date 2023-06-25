package utils

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const QUEUE_TOPIC string = "QUEUE_TOPIC"
const QUEUE_PARTITION_ID string = "QUEUE_PARTITION_ID"
const QUEUE_HOST string = "QUEUE_HOST"
const QUEUE_PUBLISH_TIMEOUT string = "QUEUE_PUBLISH_TIMEOUT"
const QUEUE_SUBSCRIBE_TIMEOUT string = "QUEUE_SUBSCRIBE_TIMEOUT"

type Queue struct {
	Connection       *kafka.Conn
	PublishTimeout   time.Duration
	SubscribeTimeout time.Duration
}

type QueueMessage struct {
	WebHookUrl string `json:"webhook_url" binding:"required"`
	Data       string `json:"data" binding:"required"`
}

type QueueMessageConsumer interface {
	ProcessMessage(msg QueueMessage)
}

func NewQueue() *Queue {
	topic := GetEnv(QUEUE_TOPIC, "notification-queue")
	partition, _ := strconv.Atoi(GetEnv(QUEUE_PARTITION_ID, "0"))
	host := GetEnv(QUEUE_HOST, "")

	var queue Queue
	conn, err := kafka.DialLeader(context.Background(), "tcp", host, topic, partition)

	if err != nil {
		log.Fatal("Queue: failed to dial leader:", err)
	}

	queue.Connection = conn
	queue.PublishTimeout, _ = time.ParseDuration(GetEnv(QUEUE_PUBLISH_TIMEOUT, "20s"))
	queue.SubscribeTimeout, _ = time.ParseDuration(GetEnv(QUEUE_SUBSCRIBE_TIMEOUT, "20s"))

	return &queue
}

func (queue Queue) Publish(msg QueueMessage) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	queue.Connection.SetWriteDeadline(time.Now().Add(queue.PublishTimeout))
	_, err = queue.Connection.WriteMessages(kafka.Message{Value: raw})
	return err
}

func (queue Queue) Subscribe(consumer QueueMessageConsumer) {
	queue.Connection.SetReadDeadline(time.Now().Add(queue.SubscribeTimeout))
	batch := queue.Connection.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	buffer := make([]byte, 10e3) // 10KB max per message
	for {
		bytes, err := batch.Read(buffer)
		if err != nil {
			break
		}

		var message QueueMessage
		err = json.Unmarshal(buffer[:bytes], &message)
		if err != nil {
			log.Printf("%v\n", err)
		}

		consumer.ProcessMessage(message)
	}

	batch.Close()
}

func (queue Queue) Close() error {
	err := queue.Connection.Close()
	return err
}
