package claude

import (
	"context"
	"github.com/bincooo/emit.io"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

const (
	cookies = "xxx"
)

func init() {
	logrus.SetLevel(logrus.TraceLevel)
}

func TestDelete(t *testing.T) {
	session, err := emit.NewJa3Session("http://127.0.0.1:7890", 180)
	if err != nil {
		t.Fatal(err)
	}

	options, err := NewDefaultOptions(cookies, "")
	if err != nil {
		t.Fatal(err)
	}

	chat, err := New(options)
	if err != nil {
		t.Fatal(err)
	}

	chat.Client(session)
	_, _ = chat.IsPro()
	chat.cid = "0821da02-1697-4df8-87b3-5b059b6cfec3"
	chat.Delete()
}

func TestChat(t *testing.T) {
	session, err := emit.NewJa3Session("http://127.0.0.1:7890", 180)
	if err != nil {
		t.Fatal(err)
	}

	options, err := NewDefaultOptions(cookies, "")
	if err != nil {
		t.Fatal(err)
	}

	timeout, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	chat, err := New(options)
	if err != nil {
		t.Fatal(err)
	}

	chat.Client(session)
	isP, _ := chat.IsPro()
	t.Logf("account is pro: %v", isP)
	partialResponse, err := chat.Reply(timeout, "hi ~ who are you?", nil)
	if err != nil {
		t.Fatal(err)
	}

	echo(t, partialResponse)
	chat.Delete()
}

func echo(t *testing.T, response chan PartialResponse) {
	content := ""
	for {
		message, ok := <-response
		if !ok {
			break
		}

		if message.Error != nil {
			t.Fatal(message.Error)
		}

		t.Log(message.Text)
		t.Log("===============")
		content += message.Text
	}

	t.Log(content)
}
