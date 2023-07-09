package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockAwsUtils struct {
	Value string
}

func NewMockAwsUtils(v string) interface{} {
	var instance MockAwsUtils
	instance.Value = v
	return &instance
}

func (mock MockAwsUtils) GetSecretValue(region, secretId string) string {
	return mock.Value
}

func TestGetEnv(t *testing.T) {
	os.Setenv("KEY001", "VALUE001")

	assert.Equal(t, "VALUE001", GetEnv("KEY001", ""), "Basic get env - found")
	assert.Equal(t, "N/A", GetEnv("KEY002", "N/A"), "Basic get env - not found")
	assert.Equal(t, "", GetEnv("KEY003", ""), "Basic get env - no default")
}

func TestNewAwsUtils(t *testing.T) {
	mock := NewMockAwsUtils(`{"ACQUIRER01_SECRETKEY":"A1SK","ACQUIRER02_SECRETKEY":"A2SK","ACQUIRER03_SECRETKEY":"A3SK","ACQUIRER01_APIKEY":"A1AK","ACQUIRER02_APIKEY":"A2AK","ACQUIRER03_APIKEY":"A3AK","ACQUIRER01_IDEMPOTENCYKEY":"A1IK","ACQUIRER02_IDEMPOTENCYKEY":"A2IK","ACQUIRER03_IDEMPOTENCYKEY":"A3IK","ACQUIRER01_ID":"A1ID","ACQUIRER02_ID":"A2ID","ACQUIRER03_ID":"A3ID","XNAP_SECRETKEY":"XSK","XNAP_APIKEY":"XAK"}`)
	loader := mock.(AwsSecretStringLoader)
	secret := NewAwsSecretValues(&loader)
	expected := &AwsSecretValues{
		Acquirer01SecretKey:      "A1SK",
		Acquirer02SecretKey:      "A2SK",
		Acquirer03SecretKey:      "A3SK",
		Acquirer01ApiKey:         "A1AK",
		Acquirer02ApiKey:         "A2AK",
		Acquirer03ApiKey:         "A3AK",
		Acquirer01IdempotencyKey: "A1IK",
		Acquirer02IdempotencyKey: "A2IK",
		Acquirer03IdempotencyKey: "A3IK",
		Acquirer01Id:             "A1ID",
		Acquirer02Id:             "A2ID",
		Acquirer03Id:             "A3ID",
		XnapSecretKey:            "XSK",
		XnapApiKey:               "XAK",
	}
	assert.Equal(t, expected, secret, "All is well scenario")

	mock = NewMockAwsUtils(`{"ACQUIRER02_SECRETKEY":"A2SK","ACQUIRER03_SECRETKEY":"A3SK","ACQUIRER02_APIKEY":"A2AK","ACQUIRER03_APIKEY":"A3AK","ACQUIRER02_IDEMPOTENCYKEY":"A2IK","ACQUIRER03_IDEMPOTENCYKEY":"A3IK","ACQUIRER02_ID":"A2ID","ACQUIRER03_ID":"A3ID","XNAP_SECRETKEY":"XSK","XNAP_APIKEY":"XAK"}`)
	loader = mock.(AwsSecretStringLoader)
	secret = NewAwsSecretValues(&loader)
	expected = &AwsSecretValues{
		Acquirer01SecretKey:      "",
		Acquirer02SecretKey:      "A2SK",
		Acquirer03SecretKey:      "A3SK",
		Acquirer01ApiKey:         "",
		Acquirer02ApiKey:         "A2AK",
		Acquirer03ApiKey:         "A3AK",
		Acquirer01IdempotencyKey: "",
		Acquirer02IdempotencyKey: "A2IK",
		Acquirer03IdempotencyKey: "A3IK",
		Acquirer01Id:             "",
		Acquirer02Id:             "A2ID",
		Acquirer03Id:             "A3ID",
		XnapSecretKey:            "XSK",
		XnapApiKey:               "XAK",
	}
	assert.Equal(t, expected, secret, "Missing some definition in AWS Secrets Manager")
}

func TestGetApiKeysMap(t *testing.T) {
	os.Setenv("ACQUIRER01_WEBHOOKURL", "A1WH")
	os.Setenv("ACQUIRER02_WEBHOOKURL", "A2WH")
	os.Setenv("ACQUIRER03_WEBHOOKURL", "A3WH")

	mock := NewMockAwsUtils(`{"ACQUIRER01_SECRETKEY":"A1SK","ACQUIRER02_SECRETKEY":"A2SK","ACQUIRER03_SECRETKEY":"A3SK","ACQUIRER01_APIKEY":"A1AK","ACQUIRER02_APIKEY":"A2AK","ACQUIRER03_APIKEY":"A3AK","ACQUIRER01_IDEMPOTENCYKEY":"A1IK","ACQUIRER02_IDEMPOTENCYKEY":"A2IK","ACQUIRER03_IDEMPOTENCYKEY":"A3IK","ACQUIRER01_ID":"A1ID","ACQUIRER02_ID":"A2ID","ACQUIRER03_ID":"A3ID","XNAP_SECRETKEY":"XSK","XNAP_APIKEY":"XAK"}`)
	loader := mock.(AwsSecretStringLoader)
	secret := NewAwsSecretValues(&loader)
	result := secret.GetApiKeysMap()

	expected := make(map[string]*ApiKeyMapValue)
	expected["A1AK"] = &ApiKeyMapValue{ApiKey: "A1AK", SecretKey: "A1SK", IdempotencyKey: "A1IK", Id: "A1ID", WebhookUrl: "A1WH"}
	expected["A2AK"] = &ApiKeyMapValue{ApiKey: "A2AK", SecretKey: "A2SK", IdempotencyKey: "A2IK", Id: "A2ID", WebhookUrl: "A2WH"}
	expected["A3AK"] = &ApiKeyMapValue{ApiKey: "A3AK", SecretKey: "A3SK", IdempotencyKey: "A3IK", Id: "A3ID", WebhookUrl: "A3WH"}

	assert.Equal(t, expected, result, "All is well scenario")
}

func TestGetWebhookKeysMap(t *testing.T) {
	os.Setenv("ACQUIRER01_WEBHOOKURL", "A1WH")
	os.Setenv("ACQUIRER02_WEBHOOKURL", "A2WH")
	os.Setenv("ACQUIRER03_WEBHOOKURL", "A3WH")

	mock := NewMockAwsUtils(`{"ACQUIRER01_SECRETKEY":"A1SK","ACQUIRER02_SECRETKEY":"A2SK","ACQUIRER03_SECRETKEY":"A3SK","ACQUIRER01_APIKEY":"A1AK","ACQUIRER02_APIKEY":"A2AK","ACQUIRER03_APIKEY":"A3AK","ACQUIRER01_IDEMPOTENCYKEY":"A1IK","ACQUIRER02_IDEMPOTENCYKEY":"A2IK","ACQUIRER03_IDEMPOTENCYKEY":"A3IK","ACQUIRER01_ID":"A1ID","ACQUIRER02_ID":"A2ID","ACQUIRER03_ID":"A3ID","XNAP_SECRETKEY":"XSK","XNAP_APIKEY":"XAK"}`)
	loader := mock.(AwsSecretStringLoader)
	secret := NewAwsSecretValues(&loader)
	result := secret.GetWebHookKeysMap()

	expected := make(map[string]*ApiKeyMapValue)
	expected["A1WH"] = &ApiKeyMapValue{ApiKey: "A1AK", SecretKey: "A1SK", IdempotencyKey: "A1IK", Id: "A1ID", WebhookUrl: "A1WH"}
	expected["A2WH"] = &ApiKeyMapValue{ApiKey: "A2AK", SecretKey: "A2SK", IdempotencyKey: "A2IK", Id: "A2ID", WebhookUrl: "A2WH"}
	expected["A3WH"] = &ApiKeyMapValue{ApiKey: "A3AK", SecretKey: "A3SK", IdempotencyKey: "A3IK", Id: "A3ID", WebhookUrl: "A3WH"}

	assert.Equal(t, expected, result, "All is well scenario")
}
