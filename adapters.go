package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type AdapterType int64

const (
	AdapterGeneric    AdapterType = 0
	AdapterMattermost AdapterType = 1
	AdapterSlack      AdapterType = 2
	AdapterNtfy       AdapterType = 3
)

type EventType string

const (
	EventFormFinished EventType = "form.finished"
)

type Input struct {
	Title      string
	Message    string
	Project    string // optional
	EndpointID string
	EventType  EventType
}

type webhook struct {
	data    webhookData
	headers map[string][]string
}

type webhookData struct {
	Data       interface{} `json:"data"`
	EventType  EventType   `json:"event_type"`
	EndpointID string      `json:"endpoint_id"`
}

func generic(input *Input) *webhook {
	return &webhook{
		webhookData{
			EventType:  input.EventType,
			EndpointID: input.EndpointID,
			Data:       input.Message,
		},
		nil,
	}
}

type mattermostData struct {
	Text        string `json:"text"`
	Attachments []struct {
		Text  string `json:"text"`
		Color string `json:"color"`
	} `json:"attachments"`
}

func mattermost(input *Input) *webhook {
	return &webhook{
		webhookData{
			EventType:  input.EventType,
			EndpointID: input.EndpointID,
			Data: mattermostData{
				Text: input.Title,
				Attachments: []struct {
					Text  string "json:\"text\""
					Color string "json:\"color\""
				}{
					{
						Color: "#1B5495",
						Text:  input.Message,
					},
				},
			},
		},
		nil,
	}
}

type slackData struct {
	Test string `json:"test"`
}

func slack(input *Input) *webhook {
	return &webhook{
		webhookData{
			EventType:  input.EventType,
			EndpointID: input.EndpointID,
			Data: slackData{
				Test: input.Title,
			},
		},
		nil,
	}
}

func ntfy(input *Input) *webhook {
	return &webhook{
		webhookData{
			EventType:  input.EventType,
			EndpointID: input.EndpointID,
			Data:       input.Message,
		},
		map[string][]string{
			"X-Title": {input.Title},
		},
	}
}

func (ad *adapterService) send(webhook *webhook, project string) error {
	if webhook == nil {
		return errors.New("webhook data undefined")
	}

	jsonBytes, err := json.Marshal(webhook.data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprint(ad.url, "/api/v1/projects/", project, "/events"),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return err
	}
	if webhook.headers != nil {
		req.Header = map[string][]string(webhook.headers)
	}
	req.Header.Set("Authorization", fmt.Sprint("Bearer ", ad.key))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			slog.Error("error closing response body", "err", err)
		}
	}(resp.Body)

	if body, err := io.ReadAll(resp.Body); err == nil {
		slog.Info(string(body)) // TODO
	}

	if resp.StatusCode >= 400 {
		return errors.New("error status code " + strconv.FormatInt(int64(resp.StatusCode), 10))
	}

	return nil
}
