package main

import (
	"context"
	"fmt"
	"github.com/Anyc66666666/claude-api"
	"time"
)

func main() {
	const (
		token = "xoxp-***"
		botId = "U05***"
	)
	chat := claude.New(token, botId, "C05EW***", time.Second*5)
	// 如果不手建频道，默认使用chat-9527
	//if err := chat.NewChannel("chat-7890"); err != nil {
	//	panic(err)
	//}

	prompt := "hi"
	fmt.Println("You: ", prompt)
	partialResponse, err := chat.Reply(context.Background(), prompt)
	if err != nil {
		panic(err)
	}
	select {
	case data := <-partialResponse:
		fmt.Println(data)

	}
	//fmt.Println(data)

	//prompt = "who are you?"
	//fmt.Println("You: ", prompt)
	//partialResponse, err = chat.Reply(context.Background(), prompt)
	//if err != nil {
	//	panic(err)
	//}
	//Println(partialResponse)
	//
	//prompt = "用中文讲个故事"
	//fmt.Println("You: ", prompt)
	//ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
	//defer cancel()
	//partialResponse, err = chat.Reply(ctx, prompt)
	//if err != nil {
	//	panic(err)
	//}
	//Println(partialResponse)
}

func Println(partialResponse chan claude.PartialResponse) {
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
