package partners

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type IssuerProfile struct {
	ID             uint   `json:"id"`
	IssID          string `json:"issId"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ApiKey         string `json:"apiKey"`
	Secret         string `json:"secret"`
	OrganizationID uint   `json:"orgId"`
}

type AcquirerProfile struct {
	ID               uint   `json:"id"`
	AcqID            string `json:"acqId"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ApiKey           string `json:"apiKey"`
	Secret           string `json:"secret"`
	NotificationHook string `json:"hook"`
	OrganizationID   uint   `json:"orgId"`
}

type PartnerService struct {
	baseUrl  string
	acqStore *MemoryStore
	issStore *MemoryStore
}

var (
	API_LIST_ACQUIRER_PATH = "/v1/profile/acquirers"
)

const REFRESH_ACQUIRERS_SECS string = "REFRESH_ACQUIRERS_SECS"
const REFRESH_ISSUERS_SECS string = "REFRESH_ISSUERS_SECS"

func NewPartnerService(baseUrl string) *PartnerService {
	return &PartnerService{
		baseUrl:  baseUrl,
		acqStore: NewMemoryStore(),
		issStore: NewMemoryStore(),
	}
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
