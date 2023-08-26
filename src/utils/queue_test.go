package utils

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type Result struct {
	ResultHost         string
	ResultGroupId      string
	ResultTopic        string
	ResultReadTimeout  time.Duration
	ResultWriteTimeout time.Duration
	Msg                []byte
}

type MockKafkaWriter struct {
	result Result
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.result.Msg = msgs[0].Value
	return nil
}

type MockKafkaReader struct {
	result Result
	msgs   []kafka.Message
	index  int
}

func (m *MockKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if m.index < len(m.msgs) {
		msg := m.msgs[m.index]
		m.index = m.index + 1
		return msg, nil
	} else {
		return kafka.Message{}, errors.New("EOF")
	}
}
func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return nil
}
func (m *MockKafkaReader) Close() error {
	return nil
}

func TestNewQueue(t *testing.T) {
	os.Setenv(QUEUE_TOPIC, "test-topic")
	os.Setenv(QUEUE_HOST, "localhost1:9022,localhost2:9022,localhost3:9022")
	os.Setenv(QUEUE_PUBLISH_TIMEOUT, "99s")
	os.Setenv(QUEUE_SUBSCRIBE_TIMEOUT, "999s")
	os.Setenv(QUEUE_CONSUMERGROUP_ID, "1-2-3")
	os.Setenv(QUEUE_READOFFSET, "EARLIEST")

	expectedReadTimeout, _ := time.ParseDuration("999s")
	expectedWriteTimeout, _ := time.ParseDuration("99s")

	mockWriter := MockKafkaWriter{}
	mockReader := MockKafkaReader{}

	CreateWriter = func(config kafka.WriterConfig) interface{} {
		mockWriter.result.ResultHost = strings.Join(config.Brokers[:], ",")
		mockWriter.result.ResultTopic = config.Topic
		mockWriter.result.ResultWriteTimeout = config.WriteTimeout
		return &mockWriter
	}

	CreateReader = func(config kafka.ReaderConfig) interface{} {
		mockReader.result.ResultHost = strings.Join(config.Brokers[:], ",")
		mockReader.result.ResultGroupId = config.GroupID
		mockReader.result.ResultTopic = config.Topic
		mockReader.result.ResultReadTimeout = config.ReadBatchTimeout
		return &mockReader
	}

	NewQueue(QUEUE_MODE_PUBLISHER)

	assert.Equal(t, "test-topic", mockWriter.result.ResultTopic, "Check topic")
	assert.Equal(t, "localhost1:9022,localhost2:9022,localhost3:9022", mockWriter.result.ResultHost, "Check host")
	assert.Equal(t, expectedWriteTimeout, mockWriter.result.ResultWriteTimeout, "Check publish timeout")

	NewQueue(QUEUE_MODE_SUBSCRIBER)

	assert.Equal(t, "test-topic", mockReader.result.ResultTopic, "Check topic")
	assert.Equal(t, "1-2-3", mockReader.result.ResultGroupId, "Check group id")
	assert.Equal(t, "localhost1:9022,localhost2:9022,localhost3:9022", mockReader.result.ResultHost, "Check host")
	assert.Equal(t, expectedReadTimeout, mockReader.result.ResultReadTimeout, "Check subscribe timeout")
}

func TestPublish(t *testing.T) {
	mockWriter := MockKafkaWriter{}

	CreateWriter = func(config kafka.WriterConfig) interface{} {
		mockWriter.result.ResultHost = strings.Join(config.Brokers[:], ",")
		mockWriter.result.ResultTopic = config.Topic
		mockWriter.result.ResultWriteTimeout = config.WriteTimeout
		return &mockWriter
	}

	queue := NewQueue(QUEUE_MODE_PUBLISHER)

	msg := QueueMessage{
		WebHookUrl: "http://localhost:8888",
		Data:       "hello, world",
	}

	expectedMsg, _ := json.Marshal(msg)
	queue.Publish(context.TODO(), msg)

	assert.Equal(t, string(expectedMsg), string(mockWriter.result.Msg), "Check msg")
}

type MockQueueMessageConsumer struct {
	msgs []QueueMessage
}

func NewMockQueueMessageConsumer() interface{} {
	var mock MockQueueMessageConsumer
	return &mock
}

func (m *MockQueueMessageConsumer) ProcessMessage(msg QueueMessage) {
	m.msgs = append(m.msgs, msg)
}

func TestSubscribe(t *testing.T) {
	qmsg := QueueMessage{WebHookUrl: "abcd", Data: "mnop"}
	data, _ := json.Marshal(qmsg)
	msg := kafka.Message{Value: []byte(data)}

	mockReader := MockKafkaReader{}
	mockReader.index = 0
	mockReader.msgs = []kafka.Message{msg}

	CreateReader = func(config kafka.ReaderConfig) interface{} {
		mockReader.result.ResultHost = strings.Join(config.Brokers[:], ",")
		mockReader.result.ResultGroupId = config.GroupID
		mockReader.result.ResultTopic = config.Topic
		mockReader.result.ResultReadTimeout = config.ReadBatchTimeout
		return &mockReader
	}

	queue := NewQueue(QUEUE_MODE_SUBSCRIBER)
	consumer := NewMockQueueMessageConsumer()

	queue.Subscribe(context.TODO(), consumer.(QueueMessageConsumer))

	cons := consumer.(*MockQueueMessageConsumer)
	out, _ := json.Marshal(cons.msgs[0])
	assert.Equal(t, string(data), string(out), "Check data")
}

func TestClose(t *testing.T) {
}
