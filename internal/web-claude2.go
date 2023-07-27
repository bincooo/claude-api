package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/bincooo/claude-api/types"
	"github.com/bincooo/requests"
	"github.com/bincooo/requests/models"
	"github.com/bincooo/requests/url"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	WebClaude2BU = "https://claude.ai/api"
	JA3          = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0"
	UA           = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.79"
)

type webClaude2Response struct {
	Completion   string `json:"completion"`
	StopReason   string `json:"stop_reason"`
	Model        string `json:"model"`
	Truncated    bool   `json:"truncated"`
	Stop         string `json:"stop"`
	LogId        string `json:"log_id"`
	Exception    any    `json:"exception"`
	MessageLimit struct {
		Type string `json:"type"`
	} `json:"messageLimit"`
}

type WebClaude2 struct {
	mu sync.Mutex
	types.Options

	organizationId string
	conversationId string
}

func NewWebClaude2(opt types.Options) types.Chat {
	return &WebClaude2{Options: opt}
}

func (wc *WebClaude2) NewChannel(string) error {
	return nil
}

func (wc *WebClaude2) Reply(ctx context.Context, prompt string, attr *types.Attachment) (chan types.PartialResponse, error) {
	wc.mu.Lock()
	if wc.Retry <= 0 {
		wc.Retry = 1
	}

	//if wc.Headers["cookie"] == "sessionKey=auto" {
	//	token, err := util.Login(wc.Agency)
	//	if err != nil {
	//		return nil, err
	//	}
	//	wc.Headers["cookie"] = "sessionKey=" + token
	//	logrus.Info("自动生成sessionKey: " + token)
	//}

	if wc.organizationId == "" {
		if err := wc.getOrganization(); err != nil {
			wc.mu.Unlock()
			return nil, err
		}
	}

	if wc.conversationId == "" {
		if err := wc.createConversation(); err != nil {
			wc.mu.Unlock()
			return nil, err
		}
	}

	var response *models.Response
	for index := 1; index <= wc.Retry; index++ {
		r, err := wc.PostMessage(5*time.Minute, prompt, attr)
		if err != nil {
			if index >= wc.Retry {
				wc.mu.Unlock()
				return nil, err
			}
		} else {
			response = r
			break
		}
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Text)
	}

	message := make(chan types.PartialResponse)
	go wc.resolve(ctx, response, message)
	return message, nil
}

func (wc *WebClaude2) resolve(ctx context.Context, r *models.Response, message chan types.PartialResponse) {
	defer wc.mu.Unlock()
	defer close(message)
	reader := bufio.NewReader(r.Body)
	block := []byte("data: ")
	original := make([]byte, 0)

	// return true 结束轮询
	handle := func() bool {
		line, hasMore, err := reader.ReadLine()
		original = append(original, line...)
		if hasMore {
			return false
		}
		//fmt.Println(string(original))
		if err == io.EOF {
			return true
		}

		if err != nil {
			message <- types.PartialResponse{
				Error: err,
			}
			return true
		}

		dst := make([]byte, len(original))
		copy(dst, original)
		original = make([]byte, 0)

		if !bytes.HasPrefix(dst, block) {
			return false
		}
		if !bytes.HasSuffix(dst, []byte("}")) {
			return false
		}

		dst = bytes.TrimPrefix(dst, block)
		var response webClaude2Response
		if e := IgnorePanicUnmarshal(dst, &response); e != nil {
			//fmt.Println(e)
			return false
		}

		message <- types.PartialResponse{
			Text:    response.Completion,
			RawData: dst,
		}

		if response.StopReason == "stop_sequence" {
			return true
		}

		return false
	}

	for {
		select {
		case <-ctx.Done():
			message <- types.PartialResponse{
				Error: errors.New("resolve timeout"),
			}
			return
		default:
			if handle() {
				return
			}
		}
	}
}

func (wc *WebClaude2) getOrganization() error {
	//headers := make(Kv)
	//headers["user-agent"] = UA
	response, err := wc.newRequest(30*time.Second, http.MethodGet, "organizations", nil, nil)
	if err != nil {
		return err
	}
	marshal, e := io.ReadAll(response.Body)
	if e != nil {
		return e
	}
	result := make([]map[string]any, 0)
	if e = json.Unmarshal(marshal, &result); e != nil {
		return e
	}
	if uid, _ := result[0]["uuid"]; uid != nil && uid != "" {
		wc.organizationId = uid.(string)
		return nil
	}
	return errors.New("failed to fetch the `organization-id`")
}

func (wc *WebClaude2) createConversation() error {
	if wc.organizationId == "" {
		return errors.New("there is no corresponding `organization-id`")
	}

	headers := make(Kv)
	headers["user-agent"] = UA

	params := make(map[string]any)
	params["name"] = ""
	params["uuid"] = uuid.NewString()
	response, err := wc.newRequest(30*time.Second, http.MethodPost, "organizations/"+wc.organizationId+"/chat_conversations", headers, params)
	if err != nil {
		return err
	}

	marshal, e := io.ReadAll(response.Body)
	if e != nil {
		return e
	}
	result := make(Kv, 0)
	if e = json.Unmarshal(marshal, &result); e != nil {
		return e
	}

	if uid, _ := result["uuid"]; uid != "" {
		wc.conversationId = uid
		return nil
	}
	return errors.New("failed to fetch the `conversation-id`")
}

func (wc *WebClaude2) PostMessage(timeout time.Duration, prompt string, attr *types.Attachment) (*models.Response, error) {
	if wc.organizationId == "" {
		return nil, errors.New("there is no corresponding `organization-id`")
	}
	if wc.conversationId == "" {
		return nil, errors.New("there is no corresponding `conversation-id`")
	}

	params := make(map[string]any)
	if attr != nil {
		params["attachments"] = []any{
			map[string]any{
				"extracted_content": attr.Content,
				"file_size":         attr.FileSize,
				"file_name":         attr.FileName,
				"file_type":         attr.FileType,
			},
		}
	} else {
		params["attachments"] = []any{}
	}
	params["conversation_uuid"] = wc.conversationId
	params["organization_uuid"] = wc.organizationId
	params["text"] = prompt
	params["completion"] = Kv{
		"model":    "claude-2",
		"prompt":   prompt,
		"timezone": "Asia/Shanghai",
	}

	headers := make(Kv)
	headers["user-agent"] = UA
	headers["accept"] = "text/event-stream"
	return wc.newRequest(timeout, http.MethodPost, "append_message", headers, params)
}

func (wc *WebClaude2) newRequest(timeout time.Duration, method string, route string, headers map[string]string, params map[string]any) (*models.Response, error) {
	if method == http.MethodGet {
		var search []string
		for key, value := range params {
			if v, ok := value.(string); ok {
				search = append(search, key+"="+v)
			}
		}

		if len(search) > 0 {
			route += "?" + strings.Join(search, "&")
		}

		params = nil
	}

	req := url.NewRequest()
	req.Timeout = timeout
	if method != http.MethodGet && params != nil {
		req.Json = params
	}

	if wc.Agency != "" {
		req.Proxies = wc.Agency
	}

	uHeaders := url.NewHeaders()
	for k, v := range headers {
		uHeaders.Set(k, v)
	}

	for k, v := range wc.Headers {
		uHeaders.Set(k, v)
	}

	req.Headers = uHeaders
	req.Ja3 = JA3
	switch method {
	case http.MethodGet:
		return requests.Get(WebClaude2BU+"/"+route, req)
	default:
		return requests.RequestStream(http.MethodPost, WebClaude2BU+"/"+route, req)
	}
}

// ====

func IgnorePanicUnmarshal(data []byte, v any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			//fmt.Println("发生了panic:", r)
			if rec, ok := r.(string); ok {
				err = errors.New(rec)
			}
		}
	}()
	//fmt.Println(string(data))
	return json.Unmarshal(data, v)
}
