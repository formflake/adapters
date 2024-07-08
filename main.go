package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type AdapterInterface interface {
	SendWebhook(input *Input, adapterType AdapterType) error
	GetEndpoint(projectID, endpointID string) (*Endpoint, error)
	GetAdapterDetails() AdapterDetailMap
	CreateEndpoint(projectID string, params UpsertEndpointParams) (*CreateEndpointResponse, error)
	UpdateEndpoint(projectID, endpointID string, params UpsertEndpointParams) (*UpdateEndpointResponse, error)
	TogglePause(projectID, endpointID string) (string, error)
}

type adapterService struct {
	AdapterInterface
	url            string
	key            string
	defaultProject string
}

var _ AdapterInterface = &adapterService{}

func NewAdapter(url, key, defaultProject string) *adapterService {
	return &adapterService{
		url:            url,
		key:            key,
		defaultProject: defaultProject,
	}
}

func (ad *adapterService) GetAdapterDetails() AdapterDetailMap {
	return adapterDetails
}

type UpsertEndpointParams struct {
	Name               string `json:"name"`
	URL                string `json:"url"`
	AdvancedSignatures bool   `json:"advanced_signatures"`
	AppID              string `json:"appID"` // deprecated but required
	// Authentication
	Description       string `json:"description"`
	HttpTimeout       int64  `json:"http_timeout"`
	IsDisabled        bool   `json:"is_disabled"`
	OwnerID           string `json:"owner_id"`
	RateLimit         int64  `json:"rate_limit"`
	RateLimitDuration int64  `json:"rate_limit_duration"`
	Secret            string `json:"secret"`
	SlackWebhookURL   string `json:"slack_webhook_url"`
	SupportEmail      string `json:"support_email"`
}

type CreateEndpointResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Uid    string `json:"uid"`
		Status string `json:"status"`
	} `json:"data"`
}

func (ad *adapterService) CreateEndpoint(projectID string, params UpsertEndpointParams) (*CreateEndpointResponse, error) {
	if projectID == "" {
		projectID = ad.defaultProject
	}

	buff := new(bytes.Buffer)
	err := json.NewEncoder(buff).Encode(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprint(ad.url, "/api/v1/projects/", projectID, "/endpoints"),
		buff,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", ad.key))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > http.StatusBadRequest {
		return nil, fmt.Errorf("response code %d invalid", resp.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Error("error closing response body", "err", err)
		}
	}(resp.Body)

	var response CreateEndpointResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

type UpdateEndpointResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func (ad *adapterService) UpdateEndpoint(projectID, endpointID string, params UpsertEndpointParams) (*UpdateEndpointResponse, error) {
	if projectID == "" {
		projectID = ad.defaultProject
	}

	buff := new(bytes.Buffer)
	err := json.NewEncoder(buff).Encode(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprint(ad.url, "/api/v1/projects/", projectID, "/endpoints/", endpointID),
		buff,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", ad.key))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > http.StatusBadRequest {
		return nil, fmt.Errorf("response code %d invalid", resp.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Error("error closing response body", "err", err)
		}
	}(resp.Body)

	var response UpdateEndpointResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (ad *adapterService) TogglePause(projectID, endpointID string) (string, error) {
	if projectID == "" {
		projectID = ad.defaultProject
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprint(ad.url, "/api/v1/projects/", projectID, "/endpoints/", endpointID, "/pause"),
		nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", ad.key))

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("response code %d invalid", resp.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Error("error closing response body", "err", err)
		}
	}(resp.Body)

	var endpoint EndpointToggleStatus
	if err := json.NewDecoder(resp.Body).Decode(&endpoint); err != nil {
		return "", err
	}

	return endpoint.Data.Status, nil
}

var sendWebhookMap = map[AdapterType]func(*Input) *webhook{
	AdapterGeneric:    generic,
	AdapterMattermost: mattermost,
	AdapterSlack:      slack,
	AdapterNtfy:       ntfy,
}

func (ad *adapterService) SendWebhook(input *Input, adapterType AdapterType) error {
	if input == nil {
		return errors.New("input not defined")
	}
	if sendWebhookFunc, ok := sendWebhookMap[adapterType]; ok {
		if input.Project == "" {
			input.Project = ad.defaultProject
		}
		return ad.send(sendWebhookFunc(input), input.Project)
	} else {
		return errors.New("function not defined")
	}
}

type Endpoint struct {
	Message string       `json:"message"`
	Status  bool         `json:"status"`
	Data    EndpointData `json:"data"`
}

type EndpointData struct {
	// Authentication
	// Secrets
	SlackWebhookURL   string     `json:"slack_webhook_url"`
	Status            string     `json:"status"`
	SupportEmail      string     `json:"support_email"`
	UID               string     `json:"uid"`
	UpdatedAt         time.Time  `json:"updated_at"`
	URL               string     `json:"url"`
	CreatedAt         time.Time  `json:"created_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	Description       string     `json:"description"`
	Events            int64      `json:"events"`
	HttpTimeout       int64      `json:"http_timeout"`
	Name              string     `json:"name"`
	OwnerID           string     `json:"owner_id"`
	ProjectID         string     `json:"project_id"`
	RateLimit         int64      `json:"rate_limit"`
	RateLimitDuration int64      `json:"rate_limit_duration"`
}

type EndpointToggleStatus struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
}

func (ad *adapterService) GetEndpoint(projectID, endpointID string) (*Endpoint, error) {
	if projectID == "" {
		projectID = ad.defaultProject
	}

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprint(ad.url, "/api/v1/projects/", projectID, "/endpoints/", endpointID),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", ad.key))

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response code %d invalid", resp.StatusCode)
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Error("error closing response body", "err", err)
		}
	}(resp.Body)

	var endpoint Endpoint
	if err := json.NewDecoder(resp.Body).Decode(&endpoint); err != nil {
		return nil, err
	}

	return &endpoint, nil
}
