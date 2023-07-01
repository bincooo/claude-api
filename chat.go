package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type B = map[string]string

func New(token, botId, channel string, pollTime, timeout time.Duration) *Chat {
	return NewChat(&Options{
		Retry:    2,
		BotId:    botId,
		Channel:  channel,
		PollTime: pollTime,
		Timeout:  timeout,
		Token:    token,
	})
}

func NewChat(opt *Options) *Chat {
	return &Chat{Options: opt}
}

// NewDefaultOptions 需要填写botId,channel,token
func NewDefaultOptions() *Options {
	return &Options{
		Retry:    2,
		BotId:    "",
		Channel:  "",
		PollTime: time.Second * 3,
		Timeout:  time.Second * 45,
		Token:    "",
	}
}

func (c *Chat) Reply(ctx context.Context, prompt string) (chan PartialResponse, error) {
	ctx1, cancel := context.WithTimeout(ctx, c.Timeout)
	go func(duration time.Duration, cancelFunc context.CancelFunc) {
		time.Sleep(duration)
		cancelFunc()
	}(c.Timeout*2, cancel)

	if c.Retry <= 0 {
		c.Retry = 1
	}
	var conversationId string

	for index := 1; index <= c.Retry; index++ {
		if id, err := c.PostMessage(prompt); err != nil {
			if index >= c.Retry {
				return nil, err
			}
		} else {
			conversationId = id
			break
		}
	}

	message := make(chan PartialResponse)
	go c.poll(ctx1, conversationId, message)
	return message, nil
}

// 轮询回复消息
func (c *Chat) poll(ctx context.Context, conversationId string, message chan PartialResponse) {
	defer close(message)
	limit := 1

	// true 结束循环 //false继续轮询
	handle := func() bool {
		replies, err := c.Replies(conversationId, limit)
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
			if value.BotId != c.BotId && !strings.HasPrefix(value.Text, "<@"+c.BotId+">") {
				slice = append(slice, value)
			}
		}

		// 没有轮询到回复消息
		if len(slice) == 0 {
			time.Sleep(c.PollTime)
			return false
		}

		value := slice[limit-1]
		if limit == 1 && value.Metadata.EventType == "claude_moderation" {
			time.Sleep(c.PollTime)
			limit = 2
			return false
		}

		for index := limit - 1; index >= 0; index-- {
			v := slice[index]
			if v.Metadata.EventType != "claude_moderation" && strings.Contains(v.Text, "I apologize, but I will not provide any responses") {
				message <- v
				return true
			}
			value = v
		}

		// 结尾没有了[_Typing…_]，结束接收
		if !strings.HasSuffix(value.Text, Typing) {
			return true
		}

		// 等待时间尽量避免触发限流
		time.Sleep(c.PollTime)
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

// Replies 获取回复
func (c *Chat) Replies(conversationId string, limit int) (*RepliesResponse, error) {
	r, err := c.newRequest(context.Background(), http.MethodGet, "conversations.replies", B{
		"channel": c.Channel,
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

// PostMessage 发送消息
func (c *Chat) PostMessage(prompt string) (conversationId string, err error) {
	conversationId = uuid.New().String()
	body := B{
		"channel":   c.Channel,
		"thread_ts": "",
		"text":      "<@" + c.BotId + ">\n" + prompt,
	}

	r, err1 := c.newRequest(context.Background(), http.MethodPost, "chat.postMessage", body)
	if err1 != nil {
		return "", err1
	}

	type postMessageResponse struct {
		ResponseClaude
		Ts string `json:"ts"`
	}
	marshal, err := io.ReadAll(r.Body)
	if err != nil {
		return "", nil
	}

	var pmr postMessageResponse
	if e := json.Unmarshal(marshal, &pmr); e != nil {
		return "", e
	}

	if !pmr.Ok {
		return "", errors.New(pmr.Error)
	}

	//if c.conversationId == "" {
	//	c.conversationId = pmr.Ts
	//}
	return pmr.Ts, nil
}

// NewChannel 创建频道
func (c *Chat) NewChannel(name string) error {

	r, err := c.newRequest(context.Background(), http.MethodGet, "conversations.list", B{"limit": "2000", "types": "public_channel,private_channel"})
	if err != nil {
		return err
	}

	type listResponse struct {
		ResponseClaude
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
			c.Channel = channel.Id
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
		ResponseClaude
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

	var rs ResponseClaude
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

	request.Header.Add("Authorization", "Bearer "+c.Token)

	if method != http.MethodGet {
		request.Header.Add("Content-Type", "application/json")
	}
	return http.DefaultClient.Do(request)
}
