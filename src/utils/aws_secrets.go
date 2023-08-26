package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const AWS_SECRET_ID = "AWS_SECRET_ID"
const AWS_REGION string = "AWS_REGION"

type AwsSecretValues struct {
	Acquirer01SecretKey      string `json:"ACQUIRER01_SECRETKEY" binding:"required"`
	Acquirer02SecretKey      string `json:"ACQUIRER02_SECRETKEY" binding:"required"`
	Acquirer03SecretKey      string `json:"ACQUIRER03_SECRETKEY" binding:"required"`
	Acquirer01ApiKey         string `json:"ACQUIRER01_APIKEY" binding:"required"`
	Acquirer02ApiKey         string `json:"ACQUIRER02_APIKEY" binding:"required"`
	Acquirer03ApiKey         string `json:"ACQUIRER03_APIKEY" binding:"required"`
	Acquirer01IdempotencyKey string `json:"ACQUIRER01_IDEMPOTENCYKEY" binding:"required"`
	Acquirer02IdempotencyKey string `json:"ACQUIRER02_IDEMPOTENCYKEY" binding:"required"`
	Acquirer03IdempotencyKey string `json:"ACQUIRER03_IDEMPOTENCYKEY" binding:"required"`
	Acquirer01Id             string `json:"ACQUIRER01_ID" binding:"required"`
	Acquirer02Id             string `json:"ACQUIRER02_ID" binding:"required"`
	Acquirer03Id             string `json:"ACQUIRER03_ID" binding:"required"`
	XnapSecretKey            string `json:"XNAP_SECRETKEY" binding:"required"`
	XnapApiKey               string `json:"XNAP_APIKEY" binding:"required"`
}

type ApiKeyMapValue struct {
	ApiKey         string
	SecretKey      string
	IdempotencyKey string
	Id             string
	WebhookUrl     string
}

type IAwsSecretStringLoader interface {
	GetSecretValue(region, secretId string) string
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func NewAwsSecretValues(loader *IAwsSecretStringLoader) *AwsSecretValues {
	var instance AwsSecretValues

	if loader == nil {
		u := NewAwsUtils().(IAwsSecretStringLoader)
		loader = &u
	}

	value := (*loader).GetSecretValue(GetEnv(AWS_REGION, "ap-southeast-1"), GetEnv(AWS_SECRET_ID, ""))
	if value != "" {
		err := json.Unmarshal([]byte(value), &instance)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return &instance
}

type AwsUtils struct {
}

func NewAwsUtils() interface{} {
	var instance AwsUtils
	return &instance
}

func (autils AwsUtils) GetSecretValue(region, secretId string) string {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	secretMgr := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{SecretId: aws.String(secretId)}

	result, err := secretMgr.GetSecretValue(context.TODO(), input)
	if err != nil {
		fmt.Println(err.Error())
	}
	return *result.SecretString
}

func (sv AwsSecretValues) GetApiKeysMap() map[string]*ApiKeyMapValue {
	result := make(map[string]*ApiKeyMapValue)
	result[sv.Acquirer01ApiKey] = &ApiKeyMapValue{ApiKey: sv.Acquirer01ApiKey, SecretKey: sv.Acquirer01SecretKey, IdempotencyKey: sv.Acquirer01IdempotencyKey, Id: sv.Acquirer01Id, WebhookUrl: GetEnv("ACQUIRER01_WEBHOOKURL", "")}
	result[sv.Acquirer02ApiKey] = &ApiKeyMapValue{ApiKey: sv.Acquirer02ApiKey, SecretKey: sv.Acquirer02SecretKey, IdempotencyKey: sv.Acquirer02IdempotencyKey, Id: sv.Acquirer02Id, WebhookUrl: GetEnv("ACQUIRER02_WEBHOOKURL", "")}
	result[sv.Acquirer03ApiKey] = &ApiKeyMapValue{ApiKey: sv.Acquirer03ApiKey, SecretKey: sv.Acquirer03SecretKey, IdempotencyKey: sv.Acquirer03IdempotencyKey, Id: sv.Acquirer03Id, WebhookUrl: GetEnv("ACQUIRER03_WEBHOOKURL", "")}
	return result
}

func (sv AwsSecretValues) GetWebHookKeysMap() map[string]*ApiKeyMapValue {
	result := make(map[string]*ApiKeyMapValue)
	result[GetEnv("ACQUIRER01_WEBHOOKURL", "")] = &ApiKeyMapValue{ApiKey: sv.Acquirer01ApiKey, SecretKey: sv.Acquirer01SecretKey, IdempotencyKey: sv.Acquirer01IdempotencyKey, Id: sv.Acquirer01Id, WebhookUrl: GetEnv("ACQUIRER01_WEBHOOKURL", "")}
	result[GetEnv("ACQUIRER02_WEBHOOKURL", "")] = &ApiKeyMapValue{ApiKey: sv.Acquirer02ApiKey, SecretKey: sv.Acquirer02SecretKey, IdempotencyKey: sv.Acquirer02IdempotencyKey, Id: sv.Acquirer02Id, WebhookUrl: GetEnv("ACQUIRER02_WEBHOOKURL", "")}
	result[GetEnv("ACQUIRER03_WEBHOOKURL", "")] = &ApiKeyMapValue{ApiKey: sv.Acquirer03ApiKey, SecretKey: sv.Acquirer03SecretKey, IdempotencyKey: sv.Acquirer03IdempotencyKey, Id: sv.Acquirer03Id, WebhookUrl: GetEnv("ACQUIRER03_WEBHOOKURL", "")}
	return result
}
