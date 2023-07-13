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
		token = "sk-ant-xxxx"
	)
	options := claude.NewDefaultOptions(token, "", vars.Model4WebClaude2)
	options.Agency = "http://127.0.0.1:7890"
	chat, err := claude.New(options)
	if err != nil {
		panic(err)
	}

	prompt := "hi"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	prompt = "who are you?"
	fmt.Println("You: ", prompt)
	partialResponse, err = chat.Reply(context.Background(), prompt)
	if err != nil {
		panic(err)
	}
	Println(partialResponse)

	prompt = "用中文讲个故事"
	fmt.Println("You: ", prompt)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	defer cancel()
	partialResponse, err = chat.Reply(ctx, prompt)
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
