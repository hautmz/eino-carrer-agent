import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getConversationList, deleteConversation } from '../api/conversation'
import { streamChat } from '../api/chat'

export const useChatStore = defineStore('chat', () => {
  const conversations = ref([])
  const currentConvId = ref('')
  const messages = ref([])
  const isStreaming = ref(false)
  const reportProgress = ref(null)

  async function loadConversations() {
    try {
      const res = await getConversationList({ page: 1, page_size: 50 })
      conversations.value = res.data.list || []
    } catch {
      conversations.value = []
    }
  }

  async function removeConversation(id) {
    await deleteConversation(id)
    if (currentConvId.value === id) {
      currentConvId.value = ''
      messages.value = []
    }
    await loadConversations()
  }

  function selectConversation(id) {
    currentConvId.value = id
    messages.value = []
  }

  function clearCurrentChat() {
    currentConvId.value = ''
    messages.value = []
  }

  function addMessage(role, content, extra = {}) {
    messages.value.push({ role, content, ...extra, id: Date.now() + Math.random() })
  }

  function appendToLastMessage(chunk) {
    const idx = messages.value.length - 1
    if (idx >= 0 && messages.value[idx].role === 'assistant') {
      const msg = { ...messages.value[idx], content: messages.value[idx].content + chunk }
      messages.value.splice(idx, 1, msg)
    }
  }

  async function sendMessage(text, fileId = null) {
    if (!text.trim() || isStreaming.value) return

    addMessage('user', text, fileId ? { file_id: fileId } : {})
    isStreaming.value = true

    const params = {
      conversation_id: currentConvId.value || '',
      message: text,
      file_id: fileId,
    }

    addMessage('assistant', '')

    return new Promise((resolve) => {
      streamChat(params, {
        onMessage: (chunk) => {
          appendToLastMessage(chunk)
        },
        onToolCall: (data) => {
          addMessage('tool', typeof data === 'string' ? data : JSON.stringify(data))
        },
        onReportProgress: (data) => {
          reportProgress.value = data
        },
        onReportResult: (data) => {
          reportProgress.value = null
          addMessage('report', typeof data === 'string' ? data : JSON.stringify(data))
        },
        onError: (err) => {
          const idx = messages.value.length - 1
          if (idx >= 0 && messages.value[idx].role === 'assistant' && !messages.value[idx].content) {
            messages.value.splice(idx, 1, { ...messages.value[idx], content: `错误: ${err}` })
          }
          isStreaming.value = false
          resolve()
        },
        onDone: () => {
          isStreaming.value = false
          loadConversations()
          resolve()
        },
      })
    })
  }

  return {
    conversations,
    currentConvId,
    messages,
    isStreaming,
    reportProgress,
    loadConversations,
    removeConversation,
    selectConversation,
    clearCurrentChat,
    addMessage,
    sendMessage,
  }
})
