import { WebClient } from '@slack/web-api'
import delay from 'delay'
import { v4 as uuidv4 } from 'uuid'
import * as types from './types'

const DAY_MS = 1000 * 60 * 60 * 24
const TYPING = '_Typing…_'

function dat() {
  return new Date()
    .getTime()
}

function str(json: Object) {
  return JSON.stringify(json)
}



class Authenticator {
  private bot?: string
  private token?: string
  private channelTs = new Map<string, number>()
  private client?: WebClient

  constructor(token: string, bot: string) {
    this.bot = bot
    this.token = token
    this.client = new WebClient( this.token )
  }


  async newChannel(name: string): string {
    const conversations = await this.client.conversations.list({ limit: 2000 })
    if (!conversations.ok) {
      const error = new types.ClaudeError(conversations.error)
      error.statusCode = 5001
      error.statusText = 'method `conversations.list` error.'
      throw error
    }

    const conversation = conversations.channels.find(it => it.name === name)
    if (conversation) {
      return conversation.id
    }

    const result = await this.client.conversations.create({ name })

    if (result.ok) {
      this._joinChannel(result.channel.id, this.bot)
      return result.channel.id
    }

    const error = new types.ClaudeError(result.error)
    error.statusCode = 5002
    error.statusText = 'method `conversations.create` error.'
    throw error
  }

  private async _joinChannel(channel: string, users: string) {
    const result = await this.client.conversations.invite({ channel, users })
    if (!result.ok) {
      await this._deleteChannel(channel)
      const error = new types.ClaudeError(result.error)
      error.statusCode = 5003
      error.statusText = 'method `conversations.invite` error.'
      throw error
    }
  }

  private async _deleteChannel(channel: string) {
    const result = await this.client.conversations.rename({
      channel: result.channel_id,
      name: name + dat()
    })
    if (result.ok) {
      await this.client.conversations.leave({
        channel: result.channel_id
      })
    }
  }


  async sendMessage(opt: {
    text: string,
    channel: string
    conversationId?: string
    onMessage?: (partialResponse: types.ChatResponse) => void
  }): types.ChatResponse {
    const {
      text,
      channel,
      conversationId = uuidv4(),
      onMessage
    } = opt

    let ts = this.channelTs.get(conversationId)
    const result = await this.client.chat.postMessage({
      text: `<@${this.bot}> ${text}`,
      thread_ts: ts,
      channel
    })

    let resultMessage = ''
    if (!this.channelTs.has(conversationId)) {
      this.channelTs.set(conversationId, result.ts)
      ts = result.ts
    }

    while(true) {
      const partialResponse = await this.client.conversations.replies({ channel, ts, limit: 1 })
      if (!partialResponse.ok) {
        await delay(500)
        continue
      }
      const messages = partialResponse.messages.filter(it => result.message.bot_id !== it.bot_id)
      const message = messages[messages.length - 1]
      if (message) {
        if (message.text) resultMessage = message.text
        if (onMessage && message.text !== TYPING) {
          onMessage({
            text: message.text?.replace(TYPING, ''),
            conversationId,
            channel
          })
        }
        if (!message.text || !message.text.endsWith(TYPING)) {
          break
        }
      }
      await delay(500)
    }

    return {
      text: resultMessage,
      conversationId,
      channel
    }
  }
}

export default Authenticator