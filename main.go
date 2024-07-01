package adapters

import (
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
		return nil, err
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
