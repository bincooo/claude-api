import Authenticator from '../src'
async function main() {

  let
    token = 'xoxp-5xxxxx',
    bot = 'U0xxxx',
    text = '讲个故事'

  const authenticator = new Authenticator(token, bot)
  const channel = await authenticator.newChannel('chat-7890')
  let result = await authenticator.sendMessage({
    text, channel, onMessage: (originalMessage) => {
      // console.log(originalMessage)
    }
  })
  console.log('==============1\n', result)

  text = '接着讲，接下来进入修仙情节'
  result = await authenticator.sendMessage({
    text, channel,
    conversationId: result.conversationId,
    onMessage: (originalMessage) => {
      // console.log(originalMessage)
    }
  })
  console.log('==============2\n', result)
}

main()