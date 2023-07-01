### 这是一个go调用claude api的库
#### example1
```
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

```
#### 其他的例子可以详情看[这里](https://github.com/Anyc66666666/claude-api/tree/main/examples)


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

