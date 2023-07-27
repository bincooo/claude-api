package main

import (
	"fmt"
	"github.com/bincooo/claude-api/internal/util"
)

func main() {
	token, err := util.Login("http://127.0.0.1:7890")
	if err != nil {
		panic(err)
	}
	fmt.Println(token)
}
