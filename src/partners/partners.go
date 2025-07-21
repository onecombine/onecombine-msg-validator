package partners

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type IssuerProfile struct {
	ID                           uint   `json:"id"`
	IssuerID                     string `json:"issuer_id"`
	Name                         string `json:"name"`
	Description                  string `json:"description"`
	ApiKey                       string `json:"apiKey"`
	Secret                       string `json:"secret"`
	OrganizationID               uint   `json:"orgId"`
	FXName                       string `json:"fx_name"`
	FXValue                      string `json:"fx_value"`
	SettlementFee                string `json:"settlement_fee"`
	SettlementType               string `json:"settlement_type"`
	SettlementWaived             bool   `json:"settlement_waived"`
	SwitchingFee                 string `json:"switching_fee"`
	SwitchingType                string `json:"switching_type"`
	SwitchingWaived              bool   `json:"switching_waived"`
	SettlementCurrencyCode       string `json:"settlement_currency_code"`
	SettlementReportBucket       string `json:"settlement_report_bucket"`
	RefundNotificationWebHook    string `json:"refund_notification_webhook"`
	CancelledNotificationWebHook string `json:"cancelled_notification_webhook"`
	Created                      string `json:"created"`
	Modified                     string `json:"modified"`
}

type AcquirerProfile struct {
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

type PartnerService struct {
	baseUrl     string
	acqStore    *MemoryStore
	issStore    *MemoryStore
	issConsumer IssuerProfileConsumer
	acqConsumer AcquirerProfileConsumer
	wg          *sync.WaitGroup
}

const (
	API_LIST_ACQUIRER_PATH = "/v1/profile/acquirers"
	API_LIST_ISSUER_PATH   = "/v1/profile/issuers"
)

const REFRESH_ACQUIRERS_SECS string = "REFRESH_ACQUIRERS_SECS"
const REFRESH_ISSUERS_SECS string = "REFRESH_ISSUERS_SECS"

func NewPartnerService(baseUrl string, issKConfig, acqKConfig *KafkaConfig) *PartnerService {
	issStore := NewMemoryStore()
	acqStore := NewMemoryStore()

	issuerConsumer := NewKafkaIssuerProfileConsumer(issStore, issKConfig)
	acquirerConsumer := NewKafkaAcquirerProfileConsumer(acqStore, acqKConfig)

	var wg sync.WaitGroup

	service := &PartnerService{
		baseUrl:     baseUrl,
		acqStore:    acqStore,
		issStore:    issStore,
		issConsumer: issuerConsumer,
		acqConsumer: acquirerConsumer,
		wg:          &wg,
	}

	service.refreshAcquirers()
	service.refreshIssuers()

	issuerConsumer.Subscribe(service.wg)
	acquirerConsumer.Subscribe(service.wg)

	return service
}

func NewPartnewServiceWithoutEvent(baseUrl string) *PartnerService {

	issStore := NewMemoryStore()
	acqStore := NewMemoryStore()

	service := &PartnerService{
		baseUrl:     baseUrl,
		acqStore:    acqStore,
		issStore:    issStore,
		issConsumer: nil,
		acqConsumer: nil,
		wg:          nil,
	}

	service.refreshAcquirers()
	service.refreshIssuers()

	return service
}

func (s PartnerService) ListAcquirers() ([]*AcquirerProfile, error) {
	response, err := http.Get(fmt.Sprintf("%s/v1/profile/acquirers", s.baseUrl))

	if err != nil {
		return []*AcquirerProfile{}, err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return []*AcquirerProfile{}, err
	}

	var acquirers []*AcquirerProfile
	json.Unmarshal(responseData, &acquirers)
	return acquirers, nil
}

func (s PartnerService) ListIssuers() ([]*IssuerProfile, error) {
	response, err := http.Get(fmt.Sprintf("%s/v1/profile/issuers", s.baseUrl))

	if err != nil {
		return []*IssuerProfile{}, err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return []*IssuerProfile{}, err
	}

	var issuers []*IssuerProfile
	json.Unmarshal(responseData, &issuers)
	return issuers, nil
}

func (s PartnerService) refreshAcquirers() error {
	response, err := http.Get(fmt.Sprintf("%s%s", s.baseUrl, API_LIST_ACQUIRER_PATH))

	if err != nil {
		return err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var acquirers []*AcquirerProfile
	json.Unmarshal(responseData, &acquirers)

	for _, acq := range acquirers {
		s.acqStore.Set(acq.ApiKey, acq)
	}
	return nil
}

func (s PartnerService) refreshIssuers() error {
	response, err := http.Get(fmt.Sprintf("%s%s", s.baseUrl, API_LIST_ISSUER_PATH))

	if err != nil {
		return err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var issuers []*IssuerProfile
	json.Unmarshal(responseData, &issuers)

	for _, iss := range issuers {
		s.issStore.Set(iss.ApiKey, iss)
	}
	return nil
}

func (s PartnerService) StartAcquirerScheduler() {
	period, err := strconv.Atoi(os.Getenv(REFRESH_ACQUIRERS_SECS))
	if err != nil {
		period = 15 // Default
	}

	go func() {
		for {
			s.refreshAcquirers()
			time.Sleep(time.Duration(period) * time.Second)
		}
	}()
}

func (s PartnerService) StartIssuerScheduler() {
	period, err := strconv.Atoi(os.Getenv(REFRESH_ISSUERS_SECS))
	if err != nil {
		period = 15 // Default
	}

	go func() {
		for {
			s.refreshIssuers()
			time.Sleep(time.Duration(period) * time.Second)
		}
	}()
}

func (s PartnerService) GetAcquirerStore() *MemoryStore {
	return s.acqStore
}

func (s PartnerService) GetIssuerStore() *MemoryStore {
	return s.issStore
}

func (s PartnerService) WaitForCompletion() {
	s.wg.Wait()
}
