package adapters

import "errors"

type AdapterInterface interface {
	SendWebhook(input *Input, adapterType AdapterType) error
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
