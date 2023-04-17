# ClaudeAI for Node.js

slack Authentication Library for ClaudeAI.

### Usage

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

Credits
Thank you to:

- https://github.com/ikechan8370/chatgpt-plugin original NodeJS implementation

