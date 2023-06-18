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
	Acquirer01SecretKey string `json:"ACQUIRER01_SECRETKEY" binding:"required"`
	Acquirer02SecretKey string `json:"ACQUIRER02_SECRETKEY" binding:"required"`
	Acquirer03SecretKey string `json:"ACQUIRER03_SECRETKEY" binding:"required"`
	Acquirer01ApiKey    string `json:"ACQUIRER01_APIKEY" binding:"required"`
	Acquirer02ApiKey    string `json:"ACQUIRER02_APIKEY" binding:"required"`
	Acquirer03ApiKey    string `json:"ACQUIRER03_APIKEY" binding:"required"`
	XnapSecretKey       string `json:"XNAP_SECRETKEY" binding:"required"`
	XnapApiKey          string `json:"XNAP_APIKEY" binding:"required"`
}

type AwsUtils struct {
	SecretsManager *secretsmanager.Client
	SecretValues   AwsSecretValues
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func NewAwsUtils() *AwsUtils {
	var instance AwsUtils

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(GetEnv(AWS_REGION, "ap-southeast-1")))
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	instance.SecretsManager = secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(GetEnv(AWS_SECRET_ID, "")),
	}

	result, err := instance.SecretsManager.GetSecretValue(context.TODO(), input)
	if err != nil {
		fmt.Println(err.Error())
	}

	if result.SecretString != nil {
		err := json.Unmarshal([]byte(*result.SecretString), &instance.SecretValues)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return &instance
}

func (awsUtils AwsUtils) GetApiKeysMap() map[string]string {
	result := make(map[string]string)
	result[awsUtils.SecretValues.Acquirer01ApiKey] = awsUtils.SecretValues.Acquirer01SecretKey
	result[awsUtils.SecretValues.Acquirer02ApiKey] = awsUtils.SecretValues.Acquirer02SecretKey
	result[awsUtils.SecretValues.Acquirer03ApiKey] = awsUtils.SecretValues.Acquirer03SecretKey
	return result
}
