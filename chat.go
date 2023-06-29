package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type B = map[string]string

func New(token string, botId string) *Chat {
	return NewChat(Options{
		Retry: 2,
		BotId: botId,
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
		},
	})
}

func NewChat(opt Options) *Chat {
	return &Chat{Options: opt}
}

func (c *Chat) Reply(ctx context.Context, prompt string) (chan PartialResponse, error) {
	if c.Retry <= 0 {
		c.Retry = 1
	}

	if c.channel == "" {
		if err := c.NewChannel("chat-9527"); err != nil {
			return nil, err
		}
	}

	for index := 1; index <= c.Retry; index++ {
		if err := c.PostMessage(prompt, c.channel, c.conversationId); err != nil {
			if index >= c.Retry {
				return nil, err
			}
		} else {
			break
		}
	}

	message := make(chan PartialResponse)
	go c.poll(ctx, message)
	return message, nil
}

// 轮询回复消息
func (c *Chat) poll(ctx context.Context, message chan PartialResponse) {
	defer close(message)
	limit := 1

	// true 结束循环
	handle := func() bool {
		replies, err := c.Replies(c.conversationId, c.channel, limit)
		if err != nil {
			message <- PartialResponse{
				Error: err,
			}
			return true
		}

		if !replies.Ok {
			message <- PartialResponse{
				Error: errors.New(replies.Error),
			}
			return true
		}

		var slice []PartialResponse
		for _, value := range replies.Messages {
			//if value.BotId != c.BotId && !strings.HasPrefix(value.Text, "<@"+c.BotId+">") {
			if value.User == c.BotId {
				slice = append(slice, value)
			}
		}

		// 没有轮询到回复消息
		if len(slice) == 0 {
			time.Sleep(time.Second)
			return false
		}

		value := slice[len(slice)-1]
		if limit == 1 && strings.Contains("|claude_moderation|claude_error_message|", "|"+value.Metadata.EventType+"|") {
			time.Sleep(time.Second)
			limit = 3
			return false
		}

		for index := len(slice) - 1; index >= 0; index-- {
			v := slice[index]
			if v.Metadata.EventType != "claude_moderation" && v.Metadata.EventType != "claude_error_message" && strings.Contains(v.Text, "I apologize, but I will not provide any responses") {
				message <- v
				return true
			}
			value = v
		}

		// 结尾没有了[_Typing…_]，结束接收
		if !strings.HasSuffix(value.Text, Typing) {
			message <- value
			return true
		}

		if value.Text != Typing {
			message <- value
		}

		// 等待1秒尽量避免触发限流
		time.Sleep(time.Second)
		return false
	}

	for {
		select {
		case <-ctx.Done():
			message <- PartialResponse{
				Error: errors.New("请求超时"),
			}
			return
		default:
			if handle() {
				return
			}
		}
	}
}

// 获取回复
func (c *Chat) Replies(conversationId string, channel string, limit int) (*RepliesResponse, error) {
	r, err := c.newRequest(context.Background(), http.MethodGet, "conversations.replies", B{
		"channel": channel,
		"limit":   strconv.Itoa(limit),
		"ts":      conversationId,
	})
	if err != nil {
		return nil, err
	}

	marshal, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var rs RepliesResponse
	err = json.Unmarshal(marshal, &rs)
	if err != nil {
		return nil, err
	}

	return &rs, nil
}

// 发送消息
func (c *Chat) PostMessage(prompt string, channel string, conversationId string) error {
	var text string
	if strings.Contains(prompt, "[@claude]") {
		text = strings.Replace(prompt, "[@claude]", "<@"+c.BotId+">", -1)
	} else {
		text = "<@" + c.BotId + ">\n" + prompt
	}

	body := B{
		"text":      text,
		"channel":   channel,
		"thread_ts": conversationId,
	}

	r, err := c.newRequest(context.Background(), http.MethodPost, "chat.postMessage", body)
	if err != nil {
		return err
	}

	type postMessageResponse struct {
		ClaudeResponse
		Ts string `json:"ts"`
	}
	marshal, err := io.ReadAll(r.Body)
	if err != nil {
		return nil
	}

	var pmr postMessageResponse
	if e := json.Unmarshal(marshal, &pmr); e != nil {
		return e
	}

	if !pmr.Ok {
		return errors.New(pmr.Error)
	}

	if c.conversationId == "" {
		c.conversationId = pmr.Ts
	}
	return nil
}

// 创建频道
func (c *Chat) NewChannel(name string) error {

	r, err := c.newRequest(context.Background(), http.MethodGet, "conversations.list", B{"limit": "2000", "types": "public_channel,private_channel"})
	if err != nil {
		return err
	}

	type listResponse struct {
		ClaudeResponse
		Channels []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"channels"`
	}

	handle := func(resp *http.Response, model any) error {
		if resp.StatusCode != 200 {
			return errors.New("请求失败：" + resp.Status)
		}

		marshal, e := io.ReadAll(resp.Body)
		if e != nil {
			return e
		}

		if e = json.Unmarshal(marshal, model); e != nil {
			return e
		}
		return nil
	}

	var lrs listResponse
	if e := handle(r, &lrs); e != nil {
		return e
	}

	if !lrs.Ok {
		return errors.New(lrs.Error)
	}

	// 检查是否已存在频道
	for _, channel := range lrs.Channels {
		if channel.Name == name {
			c.channel = channel.Id
			return nil
		}
	}

	// 创建频道
	unescape, err := url.QueryUnescape(name)
	if err != nil {
		return err
	}

	r, err = c.newRequest(context.Background(), http.MethodPost, "conversations.create", B{"name": unescape})
	if err != nil {
		return err
	}

	type createResponse struct {
		ClaudeResponse
		Channel struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"channel"`
	}

	var crs createResponse
	if e := handle(r, &crs); e != nil {
		return e
	}
	if !crs.Ok {
		return errors.New(crs.Error)
	}

	// 邀请机器人加入频道
	r, err = c.newRequest(context.Background(), http.MethodPost, "conversations.invite", B{"channel": crs.Channel.Id, "users": c.BotId})
	if err != nil {
		return err
	}

	var rs ClaudeResponse
	if e := handle(r, &rs); e != nil {
		return e
	}
	if !rs.Ok {
		return errors.New(rs.Error)
	}
	return nil
}

func (c *Chat) newRequest(ctx context.Context, method string, route string, params map[string]string) (*http.Response, error) {
	marshal := make([]byte, 0)
	if method == http.MethodGet {
		var search []string
		for key, value := range params {
			search = append(search, key+"="+value)
		}

		if len(search) > 0 {
			route += "?" + strings.Join(search, "&")
		}

	} else {
		m, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		marshal = m
	}

	request, err := http.NewRequestWithContext(ctx, method, BU+"/"+route, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}

	for k, v := range c.Headers {
		request.Header.Add(k, v)
	}

	if method != http.MethodGet {
		request.Header.Add("Content-Type", "application/json")
	}
	return http.DefaultClient.Do(request)
}
