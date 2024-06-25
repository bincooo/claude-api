

### ClaudeAI for [Node.js](./README_node.md)/GoLang

Slack Conversation Library for ClaudeAI.

Web Conversation Library for ClaudeAI.  [link](https://claude.ai/chat)

[Service For SillyTavern](https://github.com/bincooo/MiaoX)

### Usage
```bash
go get github.com/bincooo/claude-api@[commit]
```

```go
var (
    token = "sk-ant-xxx"
    attrCtx = "==附件内容=="
)

// 可自动获取token，无需手动注册
tk, err := util.Login("http://127.0.0.1:7890")
if err != nil {
    panic(err)
}
token = tk
options := claude.NewDefaultOptions(token, "", vars.Model4WebClaude2)
options.Agency = "http://127.0.0.1:7890"
chat, err := claude.New(options)
if err != nil {
    panic(err)
}

prompt := "who are you?"
fmt.Println("You: ", prompt)
// 正常对话
partialResponse, err = chat.Reply(context.Background(), prompt, nil)
if err != nil {
    panic(err)
}
Println(partialResponse)
// 附件上传
prompt = "总结附件内容："
fmt.Println("You: ", prompt)
ctx, cancel := context.WithTimeout(context.TODO(), time.Second*20)
defer cancel()
partialResponse, err = chat.Reply(ctx, prompt, []types.Attachment{
    {
        Content:  attrCtx,
        FileName: "paste.txt",
        FileSize: 999999,
        FileType: "txt",
    }
})
if err != nil {
    panic(err)
}
Println(partialResponse)

// =========

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
```
