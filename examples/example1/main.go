package main

import (
	"context"
	"fmt"
	"github.com/Anyc66666666/claude-api"
	"log"
	"time"
)

func main() {
	chat := claude.New("xoxp-***f", "U05A5***", "C05EW***", time.Second*2, time.Second*45)
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
