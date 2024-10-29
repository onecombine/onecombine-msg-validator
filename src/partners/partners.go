package partners

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type partnerService struct {
	baseUrl string
}

func NewPartnerService(baseUrl string) *partnerService {
	return &partnerService{
		baseUrl: baseUrl,
	}
}

func (s partnerService) ListAcquirers() ([]*AcquirerProfile, error) {
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

func (s partnerService) ListIssuers() ([]*IssuerProfile, error) {
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
