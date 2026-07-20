import { chatSSE } from '../utils/sse'

export function streamChat(params, callbacks) {
  return chatSSE('/api/chat/stream', params, callbacks)
}
