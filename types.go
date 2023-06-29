package claude

import "time"

type Chat struct {
	Options

	//Channel        string
	//conversationId string
}

type Options struct {
	Headers  map[string]string
	Retry    int
	BotId    string
	Channel  string
	PollTime time.Duration
}

type ResponseClaude struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type RepliesResponse struct {
	ResponseClaude
	Messages []PartialResponse `json:"messages"`
}

type PartialResponse struct {
	Error error `json:"-"`

	Text  string `json:"text"`
	BotId string `json:"bot_id"`

	Metadata struct {
		EventType string `json:"event_type"`
	} `json:"metadata"`
}
