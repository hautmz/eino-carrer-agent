<template>
  <div class="chat-window">
    <div ref="messageList" class="message-list">
      <div v-if="messages.length === 0" class="empty-hint">
        <el-empty description="开始你的职业规划对话吧！" />
      </div>
      <MessageBubble v-for="msg in messages" :key="msg.id" :message="msg" />
      <div v-if="reportProgress" class="report-progress">
        <el-progress :percentage="reportProgress.percentage || 0" :format="() => reportProgress.status || '生成中...'" />
      </div>
    </div>
    <div class="input-area">
      <div class="input-toolbar">
        <FileUpload @uploaded="onFileUploaded" />
      </div>
      <div class="input-row">
        <el-input
          v-model="inputText"
          type="textarea"
          :rows="2"
          placeholder="输入你的问题..."
          :disabled="isStreaming"
          @keydown.enter.exact.prevent="handleSend"
        />
        <el-button
          type="primary"
          :icon="Promotion"
          :loading="isStreaming"
          :disabled="!inputText.trim()"
          @click="handleSend"
        >
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import { Promotion } from '@element-plus/icons-vue'
import { storeToRefs } from 'pinia'
import { useChatStore } from '../stores/chat'
import MessageBubble from './MessageBubble.vue'
import FileUpload from './FileUpload.vue'

const chatStore = useChatStore()
const { messages, isStreaming, reportProgress } = storeToRefs(chatStore)

const inputText = ref('')
const currentFileId = ref(null)
const messageList = ref(null)

function onFileUploaded(fileId) {
  currentFileId.value = fileId
}

async function handleSend() {
  const text = inputText.value.trim()
  if (!text || isStreaming.value) return

  inputText.value = ''
  const fileId = currentFileId.value
  currentFileId.value = null

  await chatStore.sendMessage(text, fileId)
}

watch(
  () => messages.value.length,
  () => {
    nextTick(() => {
      if (messageList.value) {
        messageList.value.scrollTop = messageList.value.scrollHeight
      }
    })
  }
)
</script>

<style scoped>
.chat-window {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.empty-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.report-progress {
  padding: 12px 16px;
}

.input-area {
  border-top: 1px solid #e4e7ed;
  padding: 12px;
  background: #fff;
}

.input-toolbar {
  margin-bottom: 8px;
}

.input-row {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.input-row .el-textarea {
  flex: 1;
}
</style>
