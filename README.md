### ClaudeAI for Node.js

Slack Conversation Library for ClaudeAI.



### Usage
```bash
npm install claude-api
// or
yarn install claude-api
```

```js
import Authenticator, { type ChatResponse } from 'claude-api'
// ==========
let
    // user-token
    token = 'xoxp-xxxxx',
    // claude appid
    bot = 'U0xxxx',
    text = '讲个故事'

  const authenticator = new Authenticator(token, bot)
  // 创建一个频道，已存在则直接返回频道ID
  const channel = await authenticator.newChannel('chat-7890')
  let result: ChatResponse = await authenticator.sendMessage({
    text, channel, onMessage: (originalMessage: ChatResponse) => {
      // console.log(originalMessage)
    }
  })
  console.log('==============1\n', result)

  text = '接着讲，接下来进入修仙情节'
  result = await authenticator.sendMessage({
    text, channel,
    conversationId: result.conversationId,
    onMessage: (originalMessage: ChatResponse) => {
      // console.log(originalMessage)
    }
  })
```



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



Credits
Thank you to:

- https://github.com/ikechan8370/chatgpt-plugin original NodeJS implementation

