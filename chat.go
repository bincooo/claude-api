package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bincooo/emit.io"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

var (
	ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0"
)

const (
	baseURL   = "https://claude.ai/api"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.79"
)

type webClaude2Response struct {
	Id           string `json:"id"`
	Completion   string `json:"completion"`
	StopReason   string `json:"stop_reason"`
	Model        string `json:"model"`
	Type         string `json:"type"`
	Truncated    bool   `json:"truncated"`
	Stop         string `json:"stop"`
	LogId        string `json:"log_id"`
	Exception    any    `json:"exception"`
	MessageLimit struct {
		Type string `json:"type"`
	} `json:"messageLimit"`
}

func Ja3(j string) {
	ja3 = j
}

func NewDefaultOptions(cookies string, model string) (*Options, error) {
	options := Options{
		Retry: 2,
		Model: model,
	}

	if cookies != "" {
		if !strings.Contains(cookies, "sessionKey=") {
			cookies = "sessionKey=" + cookies
		}

		jar, err := emit.NewCookieJar("https://claude.ai", cookies)
		if err != nil {
			return nil, err
		}
		options.jar = jar
	}

	return &options, nil
}

func New(opts *Options) (*Chat, error) {
	if opts.Model != "" && !strings.HasPrefix(opts.Model, "claude-") {
		return nil, errors.New("claude-model cannot has `claude-` prefix")
	}
	return &Chat{
		opts: opts,
	}, nil
}

func (c *Chat) Client(session *emit.Session) {
	c.session = session
}

func (c *Chat) Reply(ctx context.Context, message string, attrs []Attachment) (chan PartialResponse, error) {
	if c.opts.Model == "" {
		// 动态加载 model
		model, err := c.loadModel()
		if err != nil {
			return nil, err
		}
		c.opts.Model = model
	}

	c.mu.Lock()
	logrus.Info("curr model: ", c.opts.Model)
	var response *http.Response
	for index := 1; index <= c.opts.Retry; index++ {
		r, err := c.PostMessage(message, attrs)
		if err != nil {
			if index >= c.opts.Retry {
				c.mu.Unlock()
				return nil, err
			}

			var wap *ErrorWrapper
			ok := errors.As(err, &wap)

			if ok && wap.ErrorType.Message == "Invalid model" {
				c.mu.Unlock()
				return nil, errors.New(wap.ErrorType.Message)
			} else {
				logrus.Error("[retry] ", err)
			}
		} else {
			response = r
			break
		}
	}

	ch := make(chan PartialResponse)
	go c.resolve(ctx, response, ch)
	return ch, nil
}

func (c *Chat) PostMessage(message string, attrs []Attachment) (*http.Response, error) {
	var (
		organizationId string
		conversationId string
	)

	// 获取组织ID
	{
		oid, err := c.getO()
		if err != nil {
			return nil, fmt.Errorf("fetch organization failed: %v", err)
		}
		organizationId = oid
	}

	// 获取会话ID
	{
		cid, err := c.getC(organizationId)
		if err != nil {
			return nil, fmt.Errorf("fetch conversation failed: %v", err)
		}
		conversationId = cid
	}

	payload := map[string]interface{}{
		"rendering_mode": "raw",
		"files":          make([]string, 0),
		"timezone":       "America/New_York",
		"model":          c.opts.Model,
		"prompt":         message,
	}
	if len(attrs) > 0 {
		payload["attachments"] = attrs
	} else {
		payload["attachments"] = []any{}
	}

	return emit.ClientBuilder(c.session).
		Ja3(ja3).
		CookieJar(c.opts.jar).
		POST(baseURL+"/organizations/"+organizationId+"/chat_conversations/"+conversationId+"/completion").
		Header("referer", "https://claude.ai").
		Header("accept", "text/event-stream").
		Header("user-agent", userAgent).
		JHeader().
		Body(payload).
		DoC(emit.Status(http.StatusOK), emit.IsSTREAM)
}

func (c *Chat) Delete() {
	if c.oid == "" {
		return
	}

	if c.cid == "" {
		return
	}

	_, err := emit.ClientBuilder(c.session).
		Ja3(ja3).
		CookieJar(c.opts.jar).
		DELETE(baseURL+"/organizations"+c.oid+"/chat_conversations/"+c.cid).
		Header("Origin", "https://claude.ai").
		Header("Referer", "https://claude.ai/").
		Header("Accept-Language", "en-US,en;q=0.9").
		Header("user-agent", userAgent).
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		c.cid = ""
	}
}

func (c *Chat) resolve(ctx context.Context, r *http.Response, message chan PartialResponse) {
	defer c.mu.Unlock()
	defer close(message)
	defer r.Body.Close()

	var (
		prefix1 = "event: "
		prefix2 = []byte("data: ")
	)

	scanner := bufio.NewScanner(r.Body)
	scanner.Split(func(data []byte, eof bool) (advance int, token []byte, err error) {
		if eof && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if eof {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	// return true 结束轮询
	handler := func() bool {
		if !scanner.Scan() {
			return true
		}

		var event string
		data := scanner.Text()
		logrus.Trace("--------- ORIGINAL MESSAGE ---------")
		logrus.Trace(data)

		if len(data) < 7 || data[:7] != prefix1 {
			return false
		}
		event = data[7:]

		if !scanner.Scan() {
			return true
		}

		dataBytes := scanner.Bytes()
		logrus.Trace("--------- ORIGINAL MESSAGE ---------")
		logrus.Trace(string(dataBytes))
		if len(dataBytes) < 6 || !bytes.HasPrefix(dataBytes, prefix2) {
			return false
		}

		if event != "completion" {
			return false
		}

		var response webClaude2Response
		if err := json.Unmarshal(dataBytes[6:], &response); err != nil {
			return false
		}

		message <- PartialResponse{
			Text:    response.Completion,
			RawData: dataBytes[6:],
		}

		return response.StopReason == "stop_sequence"
	}

	for {
		select {
		case <-ctx.Done():
			message <- PartialResponse{
				Error: errors.New("resolve timeout"),
			}
			return
		default:
			if handler() {
				return
			}
		}
	}
}

// 加载默认模型
func (c *Chat) IsPro() (bool, error) {
	o, err := c.getO()
	if err != nil {
		return false, err
	}

	response, err := emit.ClientBuilder(c.session).
		GET(baseURL+"/bootstrap/"+o+"/statsig").
		Ja3(ja3).
		CookieJar(c.opts.jar).
		Header("Origin", "https://claude.ai").
		Header("Referer", "https://claude.ai/").
		Header("Accept-Language", "en-US,en;q=0.9").
		Header("user-agent", userAgent).
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		return false, err
	}

	value := emit.TextResponse(response)
	compileRegex := regexp.MustCompile(`"custom":{"isPro":true,`)
	matchArr := compileRegex.FindStringSubmatch(value)
	return len(matchArr) > 0, nil
}

// 加载默认模型
func (c *Chat) loadModel() (string, error) {
	o, err := c.getO()
	if err != nil {
		return "", err
	}

	response, err := emit.ClientBuilder(c.session).
		GET(baseURL+"/bootstrap/"+o+"/statsig").
		Ja3(ja3).
		CookieJar(c.opts.jar).
		Header("Origin", "https://claude.ai").
		Header("Referer", "https://claude.ai/").
		Header("Accept-Language", "en-US,en;q=0.9").
		Header("user-agent", userAgent).
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		return "", err
	}

	value := emit.TextResponse(response)
	compileRegex := regexp.MustCompile(`"value":{"model":"(claude-[^"]+)"}`)
	matchArr := compileRegex.FindStringSubmatch(value)
	if len(matchArr) == 0 {
		return "", errors.New("failed to fetch the conversation")
	}

	return matchArr[len(matchArr)-1], nil
}

func (c *Chat) getO() (string, error) {
	if c.oid != "" {
		return c.oid, nil
	}

	response, err := emit.ClientBuilder(c.session).
		GET(baseURL+"/organizations").
		Ja3(ja3).
		CookieJar(c.opts.jar).
		Header("Origin", "https://claude.ai").
		Header("Referer", "https://claude.ai/").
		Header("Accept-Language", "en-US,en;q=0.9").
		Header("user-agent", userAgent).
		DoC(emit.Status(http.StatusOK), emit.IsJSON)
	if err != nil {
		return "", err
	}

	results, err := emit.ToSlice(response)
	if err != nil {
		return "", err
	}

	if uid, _ := results[0]["uuid"]; uid != nil && uid != "" {
		c.oid = uid.(string)
		return c.oid, nil
	}

	return "", errors.New("failed to fetch the organization")
}

func (c *Chat) getC(o string) (string, error) {
	if c.cid != "" {
		return c.cid, nil
	}

	payload := map[string]interface{}{
		"name": "",
		"uuid": uuid.New().String(),
	}

	pro, err := c.IsPro()
	if err != nil {
		return "", err
	}

	if pro {
		// 尊贵的pro
		payload["model"] = c.opts.Model
	} else {
		if strings.Contains(c.opts.Model, "opus") {
			return "", errors.New("failed to used pro model: " + c.opts.Model)
		}
	}

	response, err := emit.ClientBuilder(c.session).
		POST(baseURL+"/organizations/"+o+"/chat_conversations").
		Ja3(ja3).
		CookieJar(c.opts.jar).
		JHeader().
		Header("Origin", "https://claude.ai").
		Header("Referer", "https://claude.ai/").
		Header("Accept-Language", "en-US,en;q=0.9").
		Header("user-agent", userAgent).
		Body(payload).
		DoC(emit.Status(http.StatusCreated), emit.IsJSON)
	if err != nil {
		return "", err
	}

	result, err := emit.ToMap(response)
	if err != nil {
		return "", err
	}

	if uid, ok := result["uuid"]; ok {
		if u, okey := uid.(string); okey {
			c.cid = u
			return u, nil
		}
	}

	return "", errors.New("failed to fetch the conversation")
}
