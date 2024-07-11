package integrations

func generic(input *Input) *Webhook {
	return &Webhook{
		Data:    input.Message,
		Headers: nil,
	}
}

type mattermostData struct {
	Text        string `json:"text"`
	Attachments []struct {
		Text  string `json:"text"`
		Color string `json:"color"`
	} `json:"attachments"`
}

func mattermost(input *Input) *Webhook {
	return &Webhook{
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
		Headers: nil,
	}
}

type slackData struct {
	Text   string              `json:"text"`
	Blocks []slackMessageBlock `json:"blocks"`
}

type slackMessageBlock struct {
	Type string                `json:"type"`
	Text slackMessageBlockText `json:"text"`
}

type slackMessageBlockText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func slack(input *Input) *Webhook {
	return &Webhook{
		Data: slackData{
			Text: input.Title,
			// Blocks: []slackMessageBlock{ // TODO
			// 	{
			// 		Type: "section",
			// 		Text: slackMessageBlockText{
			// 			Type: "mrkdwn",
			// 			Text: input.Message,
			// 		},
			// 	},
			// },
		},
		Headers: nil,
	}
}

func ntfy(input *Input) *Webhook {
	return &Webhook{
		Data: input.Message,
		Headers: map[string][]string{
			"X-Title": {input.Title},
		},
	}
}
