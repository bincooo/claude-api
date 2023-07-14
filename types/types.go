package types

import (
	"context"
)

type Chat interface {
	NewChannel(name string) error
	Reply(ctx context.Context, prompt string) (chan PartialResponse, error)
}

type Options struct {
	Headers map[string]string // 请求头
	Retry   int               // 重试次数
	BotId   string            // slack里的claude-id
	Model   string            // 提供两个模型：slack 、 web-claude-2
	Agency  string            // 本地代理
}

type BasicResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type RepliesResponse struct {
	BasicResponse
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
