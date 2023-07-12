package utils

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestNewQueue(t *testing.T) {
	os.Setenv(QUEUE_TOPIC, "test-topic")
	os.Setenv(QUEUE_PARTITION_ID, "99")
	os.Setenv(QUEUE_HOST, "localhost:9022")
	os.Setenv(QUEUE_PUBLISH_TIMEOUT, "99s")
	os.Setenv(QUEUE_SUBSCRIBE_TIMEOUT, "999s")

	resultProtocol := ""
	resultHost := ""
	resultTopic := ""
	resultPartition := 0

	resultPublishTimeout, _ := time.ParseDuration("99s")
	resultSubscribeTimeout, _ := time.ParseDuration("999s")

	KafkaConnect = func(ctx context.Context, protocol, host, topic string, partition int) (*kafka.Conn, error) {
		resultProtocol = protocol
		resultHost = host
		resultTopic = topic
		resultPartition = partition
		return nil, nil
	}

	queue := NewQueue()

	assert.Equal(t, "test-topic", resultTopic, "Check topic")
	assert.Equal(t, "localhost:9022", resultHost, "Check host")
	assert.Equal(t, "tcp", resultProtocol, "Check protocol")
	assert.Equal(t, 99, resultPartition, "Check partition")
	assert.Equal(t, resultPublishTimeout, queue.PublishTimeout, "Check publish timeout")
	assert.Equal(t, resultSubscribeTimeout, queue.SubscribeTimeout, "Check subscribe timeout")
}

func TestPublish(t *testing.T) {
	os.Setenv(QUEUE_PUBLISH_TIMEOUT, "99s")
	queue := NewQueue()

	resultDeadline := time.Now()
	queue.SetWriteDeadline = func(t time.Time) error {
		resultDeadline = t
		return nil
	}

	resultMsg := ""
	queue.WriteMessages = func(m ...kafka.Message) (int, error) {
		resultMsg = string(m[0].Value)
		return 0, nil
	}

	msg := QueueMessage{
		WebHookUrl: "http://localhost:8888",
		Data:       "hello, world",
	}

	expectedMsg, _ := json.Marshal(msg)
	queue.Publish(msg)

	tm, _ := time.ParseDuration("99s")
	dl := time.Now().Add(tm)
	diff := dl.Sub(resultDeadline)
	expectedDeadline, _ := time.ParseDuration("1s")

	assert.Greater(t, expectedDeadline, diff, "Check deadline")
	assert.Equal(t, string(expectedMsg), resultMsg, "Check msg")
}

func TestSubscribe(t *testing.T) {
}

func TestClose(t *testing.T) {
}
