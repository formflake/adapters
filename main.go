package integrations

import (
	"errors"
)

type IntegrationInterface interface {
	MapWebhook(input *Input, adapterType IntegrationType) (*Webhook, error)
	GetIntegrationDetails() IntegrationDetailMap
}

type adapterService struct{}

var _ IntegrationInterface = &adapterService{}

type adapterDetail struct {
	Name  string
	Icon  string
	Color string
}

type IntegrationDetailMap map[IntegrationType]adapterDetail

type IntegrationType int64

type Input struct {
	Title   string
	Message string
}

type EventType string

type Webhook struct {
	Data    interface{}
	Headers map[string][]string
}

const (
	IntegrationGeneric    IntegrationType = 0
	IntegrationMattermost IntegrationType = 1
	IntegrationSlack      IntegrationType = 2
	IntegrationNtfy       IntegrationType = 3

	MaxTypeID int64 = int64(IntegrationNtfy)

	EventFormFinished EventType = "form.finished"
)

var adapterDetails = IntegrationDetailMap{
	IntegrationGeneric: {
		Name: "Generic Webhook",
		Icon: "logos:webhooks",
	},
	IntegrationMattermost: {
		Name: "Mattermost",
		Icon: "logos:mattermost-icon",
	},
	IntegrationSlack: {
		Name: "Slack",
		Icon: "logos:slack-icon",
	},
	IntegrationNtfy: {
		Name:  "Ntfy",
		Icon:  "simple-icons:ntfy",
		Color: "#10b981",
	},
}

func NewIntegration() *adapterService {
	return &adapterService{}
}

func (ad *adapterService) GetIntegrationDetails() IntegrationDetailMap {
	return adapterDetails
}

var sendWebhookMap = map[IntegrationType]func(*Input) *Webhook{
	IntegrationGeneric:    generic,
	IntegrationMattermost: mattermost,
	IntegrationSlack:      slack,
	IntegrationNtfy:       ntfy,
}

func (ad *adapterService) MapWebhook(input *Input, adapterType IntegrationType) (*Webhook, error) {
	if input == nil {
		return nil, errors.New("input not defined")
	}
	if sendWebhookFunc, ok := sendWebhookMap[adapterType]; ok {
		return sendWebhookFunc(input), nil
	} else {
		return nil, errors.New("function not defined")
	}
}
