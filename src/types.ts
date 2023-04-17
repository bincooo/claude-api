export class ClaudeError extends Error {
  statusCode?: number
  statusText?: string
  originalError?: Error
}

export type ChatResponse = {
  text: string
  channel: string
  conversationId?: string
}