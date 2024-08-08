package integrations

import (
	"errors"
	"fmt"
)

type IntegrationInterface interface {
	MapWebhook(input interface{}, adapterType IntegrationType, eventType EventType) (*Webhook, error)
	GetIntegrationDetails() IntegrationDetailMap
}

type adapterService struct {
	IntegrationInterface
}

type adapterData struct{}

var _ IntegrationInterface = &adapterService{}

type adapterDetail struct {
	Name  string
	Icon  string
	Color string
	Help  string
}

type IntegrationDetailMap map[IntegrationType]adapterDetail

type IntegrationType int64

type InputFormFinished struct {
	LinkText        string
	LinkUrl         string
	Title           string
	FormTranslation string
	Nodes           []InputFormFinishedNode
	Contact         *InputContactNode
}

type InputFormFinishedNode struct {
	Relation        int64
	NodeType        int64
	NodeTranslation string
	ContactNode     InputContactNode
	SelectNode      InputSelectNode
	RatingNode      InputRatingNode
	ChoiceNode      InputChoiceNode
}

type InputContactNode struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Company   string `json:"company"`
	Phone     string `json:"phone"`
	Details   string `json:"details"`
}

type InputSelectNode struct {
	Label    string   `json:"label"`
	Selected []string `json:"selected"`
}

type InputRatingNode struct {
	Label    string `json:"label"`
	Elements []struct {
		Label string `json:"label"`
		Value int64  `json:"value"`
	} `json:"elements"`
}

type InputChoiceNode struct {
	Elements []struct {
		Label       string `json:"label"`
		AnswerShort string `json:"answerShort"`
		AnswerLong  string `json:"answerLong"`
	} `json:"elements"`
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
	IntegrationTeams      IntegrationType = 4

	MinTypeID int64 = int64(IntegrationGeneric)
	MaxTypeID int64 = int64(IntegrationTeams)

	EventFormFinished EventType = "form.finished"
)

var adapterDetails = IntegrationDetailMap{
	// IntegrationGeneric: {
	// 	Name: "Generic Webhook",
	// 	Icon: "logos:webhooks",
	// },
	IntegrationMattermost: {
		Name: "Mattermost",
		Icon: "logos:mattermost-icon",
		Help: "https://developers.mattermost.com/integrate/webhooks/incoming/#create-an-incoming-webhook",
	},
	IntegrationSlack: {
		Name: "Slack",
		Icon: "logos:slack-icon",
		Help: "https://api.slack.com/messaging/webhooks",
	},
	// IntegrationNtfy: {
	// 	Name:  "Ntfy",
	// 	Icon:  "simple-icons:ntfy",
	// 	Color: "#10b981",
	// },
	IntegrationTeams: {
		Name: "Teams",
		Icon: "logos:microsoft-teams",
		Help: "https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook",
	},
}

func NewIntegration() *adapterService {
	return &adapterService{
		&adapterData{},
	}
}

func (ad *adapterData) GetIntegrationDetails() IntegrationDetailMap {
	return adapterDetails
}

var sendWebhookMap = map[IntegrationType]func(input interface{}, eventType EventType) (*Webhook, error){
	// IntegrationGeneric:    generic,
	IntegrationMattermost: mattermost,
	IntegrationSlack:      slack,
	// IntegrationNtfy:       ntfy,
	IntegrationTeams: teams,
}

func (ad *adapterData) MapWebhook(input interface{}, adapterType IntegrationType, eventType EventType) (*Webhook, error) {
	if input == nil {
		return nil, errors.New("input not defined")
	}
	if sendWebhookFunc, ok := sendWebhookMap[adapterType]; ok {
		return sendWebhookFunc(input, eventType)
	} else {
		return nil, errors.New(fmt.Sprint("map function not defined, type: ", adapterType))
	}
}
