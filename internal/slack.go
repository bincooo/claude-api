package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/bincooo/claude-api/types"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ClackBU = "https://slack.com/api"

	ClackTyping = "_Typing…_"
)

type BasicResponse struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"error"`
	RawData []byte `json:"-"`
}

type RepliesResponse struct {
	BasicResponse
	Messages []repliesMessage `json:"messages"`
}

type repliesMessage struct {
	types.PartialResponse
	BotId    string `json:"bot_id"`
	User     string `json:"user"`
	Metadata struct {
		EventType string `json:"event_type"`
	} `json:"metadata"`
}

type Kv = map[string]string

type Slack struct {
	types.Options

	channel        string
	conversationId string
}

func NewSlack(opt types.Options) types.Chat {
	return &Slack{Options: opt}
}

func (s *Slack) Reply(ctx context.Context, prompt string, attrs []types.Attachment) (chan types.PartialResponse, error) {
	if s.Retry <= 0 {
		s.Retry = 1
	}

	if s.channel == "" {
		if err := s.NewChannel("chat-9527"); err != nil {
			return nil, err
		}
	}

	for index := 1; index <= s.Retry; index++ {
		if err := s.PostMessage(prompt); err != nil {
			if index >= s.Retry {
				return nil, err
			}
		} else {
			break
		}
	}

	message := make(chan types.PartialResponse)
	go s.poll(ctx, message)
	return message, nil
}

func (*Slack) Delete() {}

// 轮询回复消息
func (s *Slack) poll(ctx context.Context, message chan types.PartialResponse) {
	defer close(message)
	limit := 1

	// true 结束循环
	handle := func() bool {
		replies, err := s.Replies(limit)
		if err != nil {
			message <- types.PartialResponse{
				Error: err,
			}
			return true
		}

		if !replies.Ok {
			message <- types.PartialResponse{
				Error:   errors.New(replies.Error),
				RawData: replies.RawData,
			}
			return true
		}

		var slice []repliesMessage
		for _, value := range replies.Messages {
			//if value.BotId != s.BotId && !strings.HasPrefix(value.Text, "<@"+s.BotId+">") {
			if value.User == s.BotId {
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
				message <- types.PartialResponse{
					Error:   v.PartialResponse.Error,
					Text:    v.PartialResponse.Text,
					RawData: replies.RawData,
				}
				return true
			}
			value = v
		}

		// 结尾没有了[_Typing…_]，结束接收
		if !strings.HasSuffix(value.Text, ClackTyping) {
			message <- types.PartialResponse{
				Error:   value.PartialResponse.Error,
				Text:    value.PartialResponse.Text,
				RawData: replies.RawData,
			}
			return true
		}

		if value.Text != ClackTyping {
			message <- types.PartialResponse{
				Error:   value.PartialResponse.Error,
				Text:    value.PartialResponse.Text,
				RawData: replies.RawData,
			}
		}

		// 等待1秒尽量避免触发限流
		time.Sleep(time.Second)
		return false
	}

	for {
		select {
		case <-ctx.Done():
			message <- types.PartialResponse{
				Error: errors.New("polling is timeout"),
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
func (s *Slack) Replies(limit int) (*RepliesResponse, error) {
	r, err := s.newRequest(context.Background(), http.MethodGet, "conversations.replies", Kv{
		"channel": s.channel,
		"limit":   strconv.Itoa(limit),
		"ts":      s.conversationId,
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

	rs.RawData = marshal
	return &rs, nil
}

// 发送消息
func (s *Slack) PostMessage(prompt string) error {
	var text string
	if strings.Contains(prompt, "[@claude]") {
		text = strings.Replace(prompt, "[@claude]", "<@"+s.BotId+">", -1)
	} else {
		text = "<@" + s.BotId + ">\n" + prompt
	}

	body := Kv{
		"text":      text,
		"channel":   s.channel,
		"thread_ts": s.conversationId,
	}

	r, err := s.newRequest(context.Background(), http.MethodPost, "chat.postMessage", body)
	if err != nil {
		return err
	}

	type postMessageResponse struct {
		BasicResponse
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

	if s.conversationId == "" {
		s.conversationId = pmr.Ts
	}
	return nil
}

// 创建频道,先查询已存在直接返回，否则创建并加入bot
func (s *Slack) NewChannel(name string) error {

	r, err := s.newRequest(context.Background(), http.MethodGet, "conversations.list", Kv{"limit": "2000", "types": "public_channel,private_channel"})
	if err != nil {
		return err
	}

	type listResponse struct {
		BasicResponse
		Channels []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"channels"`
	}

	handle := func(resp *http.Response, model any) error {
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
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
			s.channel = channel.Id
			return nil
		}
	}

	// 创建频道
	unescape, err := url.QueryUnescape(name)
	if err != nil {
		return err
	}

	r, err = s.newRequest(context.Background(), http.MethodPost, "conversations.create", Kv{"name": unescape})
	if err != nil {
		return err
	}

	type createResponse struct {
		BasicResponse
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
	r, err = s.newRequest(context.Background(), http.MethodPost, "conversations.invite", Kv{"channel": crs.Channel.Id, "users": s.BotId})
	if err != nil {
		return err
	}

	var rs BasicResponse
	if e := handle(r, &rs); e != nil {
		return e
	}
	if !rs.Ok {
		return errors.New(rs.Error)
	}
	return nil
}

func (s *Slack) newRequest(ctx context.Context, method string, route string, params map[string]string) (*http.Response, error) {
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

	request, err := http.NewRequestWithContext(ctx, method, ClackBU+"/"+route, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}

	for k, v := range s.Headers {
		request.Header.Add(k, v)
	}

	if method != http.MethodGet {
		request.Header.Add("Content-Type", "application/json")
	}
	return http.DefaultClient.Do(request)
}
