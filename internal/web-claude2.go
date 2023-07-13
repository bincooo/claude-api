package internal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/bincooo/claude-api/types"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
	"io"
	"net/http"
	"strings"
	"sync"
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

func (wc *WebClaude2) Reply(ctx context.Context, prompt string) (chan types.PartialResponse, error) {
	wc.mu.Lock()
	if wc.Retry <= 0 {
		wc.Retry = 1
	}

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
		r, err := wc.PostMessage(ctx, prompt)
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
	go wc.resolve(response, message)
	return message, nil
}

func (wc *WebClaude2) resolve(r *models.Response, message chan types.PartialResponse) {
	defer wc.mu.Unlock()
	defer close(message)
	reader := bufio.NewReader(r.Body)
	block := []byte("data: ")
	for {
		original, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			message <- types.PartialResponse{
				Error: err,
			}
			return
		}

		if !bytes.HasPrefix(original, block) {
			continue
		}
		if !bytes.HasSuffix(original, []byte("}")) {
			continue
		}

		original = bytes.TrimPrefix(original, block)
		var response webClaude2Response
		if e := IgnorePanicUnmarshal(original, &response); e != nil {
			//fmt.Println(e)
			continue
		}

		message <- types.PartialResponse{
			Text: response.Completion,
		}

		if response.StopReason == "stop_sequence" {
			return
		}
	}
}

func (wc *WebClaude2) getOrganization() error {
	//headers := make(Kv)
	//headers["user-agent"] = UA
	response, err := wc.newRequest(context.Background(), http.MethodGet, "organizations", nil, nil)
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
	response, err := wc.newRequest(context.Background(), http.MethodGet, "organizations/"+wc.organizationId+"/chat_conversations", headers, nil)
	if err != nil {
		return err
	}

	marshal, e := io.ReadAll(response.Body)
	if e != nil {
		return e
	}
	result := make([]Kv, 0)
	if e = json.Unmarshal(marshal, &result); e != nil {
		return e
	}

	if uid, _ := result[0]["uuid"]; uid != "" {
		wc.conversationId = uid
		return nil
	}
	return errors.New("failed to fetch the `conversation-id`")
}

func (wc *WebClaude2) PostMessage(ctx context.Context, prompt string) (*models.Response, error) {
	if wc.organizationId == "" {
		return nil, errors.New("there is no corresponding `organization-id`")
	}
	if wc.conversationId == "" {
		return nil, errors.New("there is no corresponding `conversation-id`")
	}

	params := make(map[string]any)
	params["attachments"] = []any{}
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
	return wc.newRequest(ctx, http.MethodPost, "append_message", headers, params)
}

func (wc *WebClaude2) newRequest(ctx context.Context, method string, route string, headers map[string]string, params map[string]any) (*models.Response, error) {
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
		return requests.Post(WebClaude2BU+"/"+route, req)
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
