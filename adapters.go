package integrations

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/nao1215/markdown"
)

// func generic(input *Input) *Webhook {
// 	return &Webhook{
// 		Data:    input.Message,
// 		Headers: nil,
// 	}
// }

type mattermostData struct {
	Text        string `json:"text"`
	Attachments []struct {
		Text  string `json:"text"`
		Color string `json:"color"`
	} `json:"attachments"`
}

func mattermost(input interface{}, eventType EventType) (*Webhook, error) {
	if input == nil {
		return nil, errors.New("input undefined")
	}
	switch eventType {
	case EventFormFinished:
		if data, ok := input.(*InputFormFinished); ok {
			message := &strings.Builder{}
			md := markdown.NewMarkdown(message).
				H2(data.FormTranslation).
				PlainText("\n")

			if data.Contact != nil {
				// TODO i18n
				md.BulletList(
					"**First name**: "+data.Contact.Firstname,
					"**Last name**: "+data.Contact.Lastname,
					"**Email address**: "+data.Contact.Email,
					"**Company**: "+data.Contact.Company,
					"**Phone**: "+data.Contact.Phone,
					"**Details**: "+data.Contact.Details,
				).
					PlainText("\n")
			}

			for _, node := range data.Nodes {
				md.H3(node.NodeTranslation)
				switch node.NodeType {
				case 0:
					for _, element := range node.ChoiceNode.Elements {
						md.BulletList(element.Label)
						if element.AnswerShort != "" {
							md.BlueBadge(element.AnswerShort)
						}
						if element.AnswerLong != "" {
							md.BlueBadge(element.AnswerLong)
						}
					}
					md.PlainText("\n")
				case 1:
					md.H4(node.SelectNode.Label).
						BulletList(node.SelectNode.Selected...).PlainText("\n")
				case 2:
					md.BulletList(
						"**First name**: "+node.ContactNode.Firstname,
						"**Last name**: "+node.ContactNode.Lastname,
						"**Email address**: "+node.ContactNode.Email,
						"**Company**: "+node.ContactNode.Company,
						"**Phone**: "+node.ContactNode.Phone,
						"**Details**: "+node.ContactNode.Details,
					).PlainText("\n")
				case 3:
					md.H4(node.RatingNode.Label)
					rows := [][]string{}
					for _, element := range node.RatingNode.Elements {
						rows = append(rows, []string{
							element.Label,
							strconv.FormatInt(element.Value, 10) + "/10 :star:",
						})
					}
					md.Table(markdown.TableSet{
						Header: []string{"Label", "Rating"},
						Rows:   rows,
					})
				}
				md.PlainText("\n")
			}

			md.Build()

			return &Webhook{
				Data: mattermostData{
					Text: fmt.Sprintf("%s [%s](%s)", data.Title, data.LinkText, data.LinkUrl),
					Attachments: []struct {
						Text  string `json:"text"`
						Color string `json:"color"`
					}{
						{
							Color: "#1B5495",
							Text:  message.String(),
						},
					},
				},
				Headers: nil,
			}, nil
		} else {
			return nil, errors.New("type assertion failed for InputFormFinished")
		}
	default:
		slog.Warn("unknown event in adapters", "eventType", eventType)
		return nil, errors.New("unknown event type")
	}
}

// type slackData struct {
// 	Blocks []slackMessageBlock `json:"blocks"`
// }

// type slackMessageBlock struct {
// 	Type      string                     `json:"type"`
// 	Text      slackMessageBlockText      `json:"text,omitempty"`
// 	Accessory slackMessageBlockAccessory `json:"accessory,omitempty"`
// }

// type slackMessageBlockText struct {
// 	Type string `json:"type"`
// 	Text string `json:"text"`
// }

// type slackMessageBlockAccessory struct {
// 	Type     string `json:"type"`
// 	ImageURL string `json:"image_url,omitempty"`
// 	AltText  string `json:"alt_text,omitempty"`
// }

// func slack(input interface{}, eventType EventType) (*Webhook, error) {
// 	switch eventType {
// 	case EventFormFinished:
// 		return &Webhook{
// 			Data: slackData{
// 				// Blocks: []slackMessageBlock{ // TODO
// 				// 	{
// 				// 		Type: "section",
// 				// 		Text: slackMessageBlockText{
// 				// 			Type: "mrkdwn",ch
// 				// 			Text: input.Message,
// 				// 		},
// 				// 	},
// 				// },
// 			},
// 			Headers: nil,
// 		}, nil
// 	}
// 	return nil, errors.New("unknown event type")
// }

// func ntfy(input *Input, eventType EventType) (*Webhook, error) {
// 	switch eventType {
// 	case EventFormFinished:
// 		return &Webhook{
// 			Data: input.Message,
// 			Headers: map[string][]string{
// 				"X-Title": {input.Title},
// 			},
// 		}, nil
// 	}
// 	return nil, errors.New("unknown event type")
// }
