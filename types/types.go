package types

import (
	"context"
)

type Chat interface {
	NewChannel(name string) error
	Reply(ctx context.Context, prompt string, attrs []Attachment) (chan PartialResponse, error)
}

type Attachment struct {
	Content  string `json:"extracted_content"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	FileType string `json:"file_type"`
}

type Options struct {
	Headers map[string]string // 请求头
	Retry   int               // 重试次数
	BotId   string            // slack里的claude-id
	Model   string            // 提供两个模型：slack 、 web-claude-2
	Agency  string            // 本地代理
}

type PartialResponse struct {
	Error   error  `json:"-"`
	Text    string `json:"text"`
	RawData []byte `json:"-"`
}
