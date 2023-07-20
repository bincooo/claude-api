package main

import (
	"context"
	"fmt"
	"github.com/bincooo/claude-api"
	"github.com/bincooo/claude-api/types"
	"github.com/bincooo/claude-api/vars"
	"time"
)

func main() {
	const (
		token = "xoxp-xxx"
		botId = "U05382WAQ1M"
	)
	options := claude.NewDefaultOptions(token, botId, vars.Model4Slack)
	chat, err := claude.New(options)
	if err != nil {
		panic(err)
	}

	// 如果不手建频道，默认使用chat-9527
	if err := chat.NewChannel("chat-7890"); err != nil {
		panic(err)
	}

	prompt := "hi"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	prompt = "who are you?"
	fmt.Println("You: ", prompt)
	partialResponse, err = chat.Reply(context.Background(), prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	prompt = "用中文讲个故事"
	fmt.Println("You: ", prompt)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	partialResponse, err = chat.Reply(ctx, prompt, nil)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)
}

func Println(partialResponse chan types.PartialResponse) {
	for {
		message, ok := <-partialResponse
		if !ok {
			return
		}

		if message.Error != nil {
			panic(message.Error)
		}

		fmt.Println(message.Text)
		fmt.Println("===============")
	}
}
