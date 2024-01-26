package types

import (
	"context"
	"fmt"
)

type Chat interface {
	NewChannel(name string) error
	Reply(ctx context.Context, prompt string, attrs []Attachment) (chan PartialResponse, error)
	Delete()
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
	BaseURL string            // 可代理转发
}

type PartialResponse struct {
	Error   error  `json:"-"`
	Text    string `json:"text"`
	RawData []byte `json:"-"`
}

type ErrorWrapper struct {
	ErrorType ErrorType `json:"error"`
	Detail    string    `json:"detail"`
}

type ErrorType struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c ErrorWrapper) Error() string {
	return fmt.Sprintf("[ClaudeError::%s]%s: %s", c.ErrorType.Type, c.ErrorType.Message, c.Detail)
}
