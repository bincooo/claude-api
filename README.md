

### ClaudeAI for [Node.js](./README_node.md)/GoLang

Slack Conversation Library for ClaudeAI.

Web Conversation Library for ClaudeAI.  [link](https://claude.ai/chat)

[Service For SillyTavern](https://github.com/bincooo/MiaoX)

### Usage
```bash
go get github.com/bincooo/claude-api@[commit]
```

使用slack for claude
```go
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
// 不支持附件
partialResponse, err := chat.Reply(context.Background(), prompt, nil)
if err != nil {
    panic(err)
}
Println(partialResponse)

// ======

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

使用web for claude

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



### New 🎉🎉🎉

（2023-09-01）自动刷取token凭证失效，添加临时方案（不保证可用性，也许会抽风）

<span style="color:red">*</span>tips：<span style="color:red">对电脑要求比较高，吃性能</span>, 手机啥的就不要想了

[视频教程](https://www.bilibili.com/video/BV1Sw411S7hZ)

step 1:

电脑需安装docker，自行研究安装。

安装完成后执行命令，可查看是否安装成功

```bash
docker info
```

step 2:

同级目录下创建`.env`文件，填写你的电脑ip和vpn （根据个人需要填写，英美地区电脑就不需要填写，留空）。

ip是你本机的ip，不要填写127.0.0.1，不然容器无法识别

```tex
PROXY="http://[你电脑的ip]:7890"
```

step 3:

运行镜像：docker compose和 指令二选一

docker compose

```vim
version: '3'
services:
  app:
    restart: always
    image: bincooo/claude-helper:v1.0.1
    volumes:
     - ./.env:/code/.env
    environment:
     - ENABLED_X11VNC=no
    ports:
     - 8088:8080
```

docker command

```bash
docker run --name claude-helper -p 8088:8080 -v ./.env:/code/.env -d bincooo/claude-helper:v1.0.1
```



（2023-07-28）提供自动刷取token凭证
`RECAPTCHA_KEY` 、`RECAPTCHA_TOKEN` 的值在claude.ai的登陆页面随意填写邮箱，点击发送后在开发者工具中的`https://claude.ai/api/auth/send_code` 请求中获取

    1. 是否有过期时间未知？？？？
    2. 是否与IP绑定未知？？？？
（实验性功能）请自行测试研究...

+++++++++++<br>
添加了web新出的claude-2 🎉

食用方法，在浏览器内登陆，打开开发者工具（F12），复制Cookie中的sessionKey即可。

sessionKey便是程序中的token，appid就不需要了，具体使用参考示例：examples/claude-2/main.go



### 授权以及获取user-token

网页([登录](https://app.slack.com))后, 进入api配置页面([点我跳转](https://api.slack.com/))。

〉》点击 【Create an app】

​	〉》主页看见Your Apps并弹出窗口【Create an app】  〉》  点击【From scratch】

​	〉》填写app名称以及选择工作空间（例：name: Bot, workspace: chat）	 〉》  点击【Create App】

​	〉》点击左侧边栏上的【OAuth & Permissions】	 〉》  下拉至【Scopes】卡片，在 【User Token Scopes】 项下添加权限，如下：

​							channels:history,  channels:read,  channels:write,  groups:history,  groups:read,  groups:write, 

​							chat:write,  im:history,  im:write,  mpim:history,  mpim:write

​	〉》回到顶部【OAuth Tokens for Your Workspace】栏，点击【Install to Workspace】，然后确认授权即可


至此，获得拥有一定权限的user-token

<img src="static/截屏2023-04-18 09.10.56.png" alt="截屏2023-04-18 09.10.56" style="zoom:50%;" />



<img src="static/截屏2023-04-18 09.14.41.png" alt="截屏2023-04-18 09.14.41" style="zoom:50%;" />



### 获取 claude appid

<img src="static/截屏2023-04-18 08.49.20.png" alt="截屏2023-04-18 08.49.20" style="zoom:50%;" />

### 注意事项
由于是Slack转接Claude，Slack是有限流机制[读我](https://api.slack.com/docs/rate-limits#tier_t5)。
目前使用的是web协议对接，文档说明似乎是1秒一个请求，后面可以尝试使用sock对接可拥有更多的请求流量。

Credits
Thank you to:

- https://github.com/ikechan8370/chatgpt-plugin original NodeJS implementation

