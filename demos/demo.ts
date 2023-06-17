import Authenticator from '../src'
async function main() {

  let
    token = 'xoxp-5137262897089-5124636131074-5142120975890-ddeaf55a79dcf72fe0e246e754ed0841',
    bot = 'U05382WAQ1M',
    debug = true,
    text = '讲个故事'

  const authenticator = new Authenticator(token, bot, debug)
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