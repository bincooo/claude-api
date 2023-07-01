package main

import (
	"context"
	"fmt"
	"github.com/Anyc66666666/claude-api"
	"log"
	"time"
)

func main() {

	chat := claude.NewChat(&claude.Options{
		Token:    "xoxp-***f",
		Retry:    2,
		BotId:    "U05A5***",
		Channel:  "C05EW***",
		PollTime: time.Second * 2,
		Timeout:  time.Second * 45,
	})

	ctx := context.Background()
	res, err := chat.Reply(ctx, "什么是claude")
	if err != nil {
		log.Println(err)
	}

	for {

		select {

		case data := <-res:
			if data.Error != nil {
				log.Println(data.Error)
				return
			}
			fmt.Println(data.Text)
			return

		}

	}
}
