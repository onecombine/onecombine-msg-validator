package partners

import (
	"context"
	"strings"
	"time"

	"github.com/onecombine/onecombine-msg-validator/src/utils"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/aws_msk_iam_v2"
	"github.com/segmentio/kafka-go/sasl/plain"

	"google.golang.org/protobuf/proto"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	pb "github.com/onecombine/onecombine-msg-validator/src/messages"
)

type pbEventPublisher struct {
	a *kafka.Writer
	i *kafka.Writer
}

// PublishAcquirerProfileChangedEvent implements EventPublisher.
func (p *pbEventPublisher) PublishAcquirerProfileChangedEvent(ctx context.Context, acq *AcquirerProfile) error {
	profile := &pb.AcquirerProfile{
		Version:                "v1.0",
		Id:                     int32(acq.ID),
		AcqId:                  acq.AcqID,
		Name:                   acq.Name,
		Description:            acq.Description,
		ApiKey:                 acq.ApiKey,
		Secret:                 acq.Secret,
		OrgId:                  int32(acq.OrganizationID),
		SettlementFee:          acq.SettlementFee,
		SettlementFeeType:      acq.SettlementType,
		SettlementFeeWaived:    acq.SettlementWaived,
		SwitchingFee:           acq.SwitchingFee,
		SwitchingFeeType:       acq.SwitchingType,
		SwitchingFeeWaived:     acq.SwitchingWaived,
		SettlementCurrencyCode: acq.SettlementCurrencyCode,
		SettlementReportBucket: acq.SettlementReportBucket,
		Created:                acq.Created,
		Modified:               acq.Modified,
	}

	val, err := proto.Marshal(profile)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(acq.AcqID),
		Value: val,
	}

	err = p.i.WriteMessages(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

// PublishIssuerProfileChangedEvent implements EventPublisher.
func (p *pbEventPublisher) PublishIssuerProfileChangedEvent(ctx context.Context, iss *IssuerProfile) error {

	profile := &pb.IssuerProfile{
		Version:                         "v1.0",
		Id:                              int32(iss.ID),
		IssuerId:                        iss.IssuerID,
		Name:                            iss.Name,
		Description:                     iss.Description,
		ApiKey:                          iss.ApiKey,
		Secret:                          iss.Secret,
		OrgId:                           int32(iss.OrganizationID),
		FxName:                          iss.FXName,
		FxValue:                         iss.FXValue,
		SettlementFee:                   iss.SettlementFee,
		SettlementFeeType:               iss.SettlementType,
		SettlementFeeWaived:             iss.SettlementWaived,
		SwitchingFee:                    iss.SwitchingFee,
		SwitchingFeeType:                iss.SwitchingType,
		SwitchingFeeWaived:              iss.SwitchingWaived,
		SettlementCurrencyCode:          iss.SettlementCurrencyCode,
		SettlementReportBucket:          iss.SettlementReportBucket,
		RefundNotificationWebhook:       iss.RefundNotificationWebHook,
		CalcellationNotificationWebhook: iss.CancelledNotificationWebHook,
		Created:                         iss.Created,
		Modified:                        iss.Modified,
	}

	val, err := proto.Marshal(profile)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(iss.IssuerID),
		Value: val,
	}

	err = p.i.WriteMessages(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

func NewPBEventPublisher(cfg *EventPublisherConfig) EventPublisher {
	var dialer *kafka.Dialer

	switch cfg.QueueType {
	case utils.MSK:
		awsConfig, err := awsConfig.LoadDefaultConfig(context.Background())
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

	hosts := strings.Split(cfg.Bootstrap, ",")
	writeTimeout, _ := time.ParseDuration(cfg.Timeout)

	acqKafkaConfig := kafka.WriterConfig{
		Brokers:      hosts,
		Topic:        cfg.AcquirerProfileTopicName,
		Balancer:     &kafka.Hash{},
		Dialer:       dialer,
		WriteTimeout: writeTimeout,
		//Logger:       kafka.LoggerFunc(logInfo),
		//ErrorLogger: kafka.LoggerFunc(logError),
	}

	issKafkaConfig := kafka.WriterConfig{
		Brokers:      hosts,
		Topic:        cfg.IssuerProfileTopicName,
		Balancer:     &kafka.Hash{},
		Dialer:       dialer,
		WriteTimeout: writeTimeout,
		//Logger:       kafka.LoggerFunc(logInfo),
		//ErrorLogger: kafka.LoggerFunc(logError),
	}

	return &pbEventPublisher{
		i: kafka.NewWriter(issKafkaConfig),
		a: kafka.NewWriter(acqKafkaConfig),
	}
}
