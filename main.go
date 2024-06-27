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

type AdapterInterface interface {
	Send(input *Input, adapterType AdapterType) error
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

func (ad *adapterService) Send(input *Input, adapterType AdapterType) error {
	switch adapterType {
	case AdapterGeneric:
		return ad.generic(input)
	case AdapterMattermost:
		return ad.mattermost(input)
	case AdapterSlack:
		return ad.slack(input)
	case AdapterNtfy:
		return ad.ntfy(input)
	default:
		return errors.New("invalid adapter type")
	}
}

func (ad *adapterService) sendWebhook(webhook *webhookData, headers *headerMap, project string) error {
	if webhook == nil {
		return errors.New("webhook data undefined")
	}

	if project == "" {
		project = ad.defaultProject
	}

	jsonBytes, err := json.Marshal(webhook)
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
	if headers != nil {
		req.Header = map[string][]string(*headers)
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
