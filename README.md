

### WEB ClaudeAI for GoLang

Web Conversation Library for ClaudeAI.  [link](https://claude.ai/chat)

### Usage
```bash
go get github.com/bincooo/claude-api@[commit]
```

```go
const (
	cookies = "xxx"
)

fun main() {
	session := emit.NewJa3Session("http://127.0.0.1:7890", 180*time.Second)
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
```
