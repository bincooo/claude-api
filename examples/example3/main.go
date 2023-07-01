package main

import (
	"context"
	"fmt"
	"github.com/Anyc66666666/claude-api"
	"log"
)

func main() {
	options := claude.NewDefaultOptions()
	options.BotId = "U05A5***"
	options.Channel = "C05EW***"
	options.Token = "xoxp-***f"

	chat := claude.NewChat(options)

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
