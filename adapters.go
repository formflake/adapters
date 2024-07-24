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
				default:
					slog.Warn("unknown node type", "nodeType", node.NodeType, "adapter", "mattermost")
					continue
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
		slog.Warn("unknown event type", "eventType", eventType, "adapter", "mattermost")
		return nil, errors.New("unknown event type")
	}
}

type slackData struct {
	Blocks []slackMessageBlock `json:"blocks"`
}

type slackMessageBlockType string

const (
	slackMessageBlockTypeSection  slackMessageBlockType = "section"
	slackMessageBlockTypeDivider  slackMessageBlockType = "divider"
	slackMessageBlockTypeContext  slackMessageBlockType = "context"
	slackMessageBlockTypeImage    slackMessageBlockType = "image"
	slackMessageBlockTypeRichText slackMessageBlockType = "rich_text"
)

type slackMessageBlock struct {
	Type      slackMessageBlockType       `json:"type"`
	Text      *slackMessageBlockText      `json:"text,omitempty"`
	Accessory *slackMessageBlockAccessory `json:"accessory,omitempty"`
	Elements  *[]slackMessageBlockText    `json:"elements,omitempty"`
}

type slackMessageBlockTextType string

const (
	slackMessageBlockTextTypeText                 slackMessageBlockTextType = "text"
	slackMessageBlockTextTypeMarkdown             slackMessageBlockTextType = "mrkdwn"
	slackMessageBlockTextTypePlainText            slackMessageBlockTextType = "plain_text"
	slackMessageBlockTextTypeRichTextSection      slackMessageBlockTextType = "rich_text_section"
	slackMessageBlockTextTypeRichTextList         slackMessageBlockTextType = "rich_text_list"
	slackMessageBlockTextTypeRichTextPreformatted slackMessageBlockTextType = "rich_text_preformatted"
)

type slackMessageBlockElementStyle string

const (
	slackMessageBlockElementStyleBullet slackMessageBlockElementStyle = "bullet"
)

type slackMessageBlockText struct {
	Type     slackMessageBlockTextType     `json:"type"`
	Style    slackMessageBlockElementStyle `json:"style,omitempty"`
	Text     string                        `json:"text,omitempty"`
	Elements *[]slackMessageBlockText      `json:"elements,omitempty"`
}

type slackMessageBlockAccessory struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url,omitempty"`
	AltText  string `json:"alt_text,omitempty"`
}

func slackBlockBulletList(title string, list []string) slackMessageBlock {
	contactBlockElements := []slackMessageBlockText{}

	for _, item := range list {
		if item != "" {
			contactBlockElements = append(contactBlockElements, slackMessageBlockText{
				Type: slackMessageBlockTextTypeRichTextSection,
				Elements: &[]slackMessageBlockText{
					{
						Type: slackMessageBlockTextTypeText,
						Text: item,
					},
				},
			})
		}
	}

	contactBlock := slackMessageBlock{
		Type: slackMessageBlockTypeRichText,
		Elements: &[]slackMessageBlockText{
			{
				Type: slackMessageBlockTextTypeRichTextSection,
				Elements: &[]slackMessageBlockText{
					{
						Type: slackMessageBlockTextTypeText,
						Text: title,
					},
				},
			},
			{
				Type:     slackMessageBlockTextTypeRichTextList,
				Style:    slackMessageBlockElementStyleBullet,
				Elements: &contactBlockElements,
			},
		},
	}

	return contactBlock
}

func slackContactBlock(title string, contact *InputContactNode) slackMessageBlock {
	contactBlockElements := []slackMessageBlockText{}

	if contact.Firstname != "" {
		contactBlockElements = append(contactBlockElements, slackMessageBlockText{
			Type: slackMessageBlockTextTypeRichTextSection,
			Elements: &[]slackMessageBlockText{
				{
					Type: slackMessageBlockTextTypeText,
					Text: "First name: ",
				},
				{
					Type: slackMessageBlockTextTypeText,
					Text: contact.Firstname,
				},
			},
		})
	}

	if contact.Lastname != "" {
		contactBlockElements = append(contactBlockElements, slackMessageBlockText{
			Type: slackMessageBlockTextTypeRichTextSection,
			Elements: &[]slackMessageBlockText{
				{
					Type: slackMessageBlockTextTypeText,
					Text: "Last name: ",
				},
				{
					Type: slackMessageBlockTextTypeText,
					Text: contact.Lastname,
				},
			},
		})
	}

	if contact.Email != "" {
		contactBlockElements = append(contactBlockElements, slackMessageBlockText{
			Type: slackMessageBlockTextTypeRichTextSection,
			Elements: &[]slackMessageBlockText{
				{
					Type: slackMessageBlockTextTypeText,
					Text: "Email address: ",
				},
				{
					Type: slackMessageBlockTextTypeText,
					Text: contact.Email,
				},
			},
		})
	}

	if contact.Company != "" {
		contactBlockElements = append(contactBlockElements, slackMessageBlockText{
			Type: slackMessageBlockTextTypeRichTextSection,
			Elements: &[]slackMessageBlockText{
				{
					Type: slackMessageBlockTextTypeText,
					Text: "Company: ",
				},
				{
					Type: slackMessageBlockTextTypeText,
					Text: contact.Company,
				},
			},
		})
	}

	if contact.Phone != "" {
		contactBlockElements = append(contactBlockElements, slackMessageBlockText{
			Type: slackMessageBlockTextTypeRichTextSection,
			Elements: &[]slackMessageBlockText{
				{
					Type: slackMessageBlockTextTypeText,
					Text: "Phone: ",
				},
				{
					Type: slackMessageBlockTextTypeText,
					Text: contact.Phone,
				},
			},
		})
	}

	if contact.Details != "" {
		contactBlockElements = append(contactBlockElements, slackMessageBlockText{
			Type: slackMessageBlockTextTypeRichTextSection,
			Elements: &[]slackMessageBlockText{
				{
					Type: slackMessageBlockTextTypeText,
					Text: "Details: ",
				},
				{
					Type: slackMessageBlockTextTypeText,
					Text: contact.Details,
				},
			},
		})
	}

	if title == "" {
		title = "Contact Information"
	}

	contactBlock := slackMessageBlock{
		Type: slackMessageBlockTypeRichText,
		Elements: &[]slackMessageBlockText{
			{
				Type: slackMessageBlockTextTypeRichTextSection,
				Elements: &[]slackMessageBlockText{
					{
						Type: slackMessageBlockTextTypeText,
						Text: title,
					},
				},
			},
			{
				Type:     slackMessageBlockTextTypeRichTextList,
				Style:    slackMessageBlockElementStyleBullet,
				Elements: &contactBlockElements,
			},
		},
	}

	return contactBlock
}

func slack(input interface{}, eventType EventType) (*Webhook, error) {
	if input == nil {
		return nil, errors.New("input undefined")
	}
	switch eventType {
	case EventFormFinished:
		if data, ok := input.(*InputFormFinished); ok {
			blocks := []slackMessageBlock{
				{
					Type: slackMessageBlockTypeSection,
					Text: &slackMessageBlockText{
						Type: slackMessageBlockTextTypeMarkdown,
						Text: fmt.Sprintf("%s <%s|%s>", data.Title, data.LinkUrl, data.LinkText),
					},
				},
				{
					Type: slackMessageBlockTypeDivider,
				},
			}

			if data.Contact != nil {
				blocks = append(blocks, slackContactBlock("", data.Contact))
			}

			for _, node := range data.Nodes {
				if node.NodeTranslation == "" {
					node.NodeTranslation = "Missing Translation"
				}
				switch node.NodeType {
				case 0:
					richTextElements := []slackMessageBlockText{
						{
							Type: slackMessageBlockTextTypeRichTextSection,
							Elements: &[]slackMessageBlockText{
								{
									Type: slackMessageBlockTextTypeText,
									Text: node.NodeTranslation,
								},
							},
						},
					}

					for _, element := range node.ChoiceNode.Elements {
						if element.Label == "" {
							continue
						}
						richTextElements = append(richTextElements, slackMessageBlockText{
							Type:  slackMessageBlockTextTypeRichTextList,
							Style: slackMessageBlockElementStyleBullet,
							Elements: &[]slackMessageBlockText{
								{
									Type: slackMessageBlockTextTypeRichTextSection,
									Elements: &[]slackMessageBlockText{
										{
											Type: slackMessageBlockTextTypeText,
											Text: element.Label,
										},
									},
								},
							},
						})
						if element.AnswerShort != "" {
							richTextElements = append(richTextElements, slackMessageBlockText{
								Type: slackMessageBlockTextTypeRichTextPreformatted,
								Elements: &[]slackMessageBlockText{
									{
										Type: slackMessageBlockTextTypeText,
										Text: element.AnswerShort,
									},
								},
							})
						}
						if element.AnswerLong != "" {
							richTextElements = append(richTextElements, slackMessageBlockText{
								Type: slackMessageBlockTextTypeRichTextPreformatted,
								Elements: &[]slackMessageBlockText{
									{
										Type: slackMessageBlockTextTypeText,
										Text: element.AnswerLong,
									},
								},
							})
						}
					}

					richTextBlock := slackMessageBlock{
						Type:     slackMessageBlockTypeRichText,
						Elements: &richTextElements,
					}
					blocks = append(blocks, richTextBlock)
				case 1:
					blocks = append(blocks, slackBlockBulletList(node.SelectNode.Label, node.SelectNode.Selected))
				case 2:
					blocks = append(blocks, slackContactBlock(node.NodeTranslation, &node.ContactNode))
				case 3:
					rows := make([]string, len(node.RatingNode.Elements))
					for idx, element := range node.RatingNode.Elements {
						rows[idx] = fmt.Sprint(element.Label, ": ", strconv.FormatInt(element.Value, 10), "/10 ⭐")
					}
					blocks = append(blocks, slackBlockBulletList(node.RatingNode.Label, rows))
				default:
					slog.Warn("unknown node type", "nodeType", node.NodeType, "adapter", "slack")
					continue
				}
				blocks = append(blocks, slackMessageBlock{
					Type: slackMessageBlockTypeDivider,
				})
			}

			return &Webhook{
				Data: slackData{
					Blocks: blocks,
				},
				Headers: nil,
			}, nil
		} else {
			return nil, errors.New("type assertion failed for InputFormFinished")
		}
	default:
		slog.Warn("unknown event type", "eventType", eventType, "adapter", "slack")
		return nil, errors.New("unknown event type")
	}
}

// func ntfy(input interface{}, eventType EventType) (*Webhook, error) {
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
