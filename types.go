package claude

type Chat struct {
	Options

	channel        string
	conversationId string
}

type Options struct {
	Headers map[string]string
	Retry   int
	BotId   string
}

type ClaudeResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type RepliesResponse struct {
	ClaudeResponse
	Messages []PartialResponse `json:"messages"`
}

type PartialResponse struct {
	Error error `json:"-"`

	Text  string `json:"text"`
	BotId string `json:"bot_id"`
	User  string `json:"user"`

	Metadata struct {
		EventType string `json:"event_type"`
	} `json:"metadata"`
}
