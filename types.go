package claude

import (
	"fmt"
	"github.com/bincooo/emit.io"
	"net/http"
	"sync"
)

type Chat struct {
	mu   sync.Mutex
	opts *Options

	oid string
	cid string

	session *emit.Session
}

type Attachment struct {
	Content  string `json:"extracted_content"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	FileType string `json:"file_type"`
}

type Options struct {
	Retry   int    // 重试次数
	BotId   string // slack里的claude-id
	Model   string // 提供两个模型：slack 、 web-claude-2
	Proxies string // 本地代理
	BaseURL string // 可代理转发
	jar     http.CookieJar
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
