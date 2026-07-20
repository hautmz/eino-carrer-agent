<template>
  <div :class="['message-bubble', `message-${message.role}`]">
    <div class="message-avatar">
      <el-avatar :size="32" :style="{ background: avatarColor }">
        {{ avatarText }}
      </el-avatar>
    </div>
    <div class="message-body">
      <div v-if="message.role === 'assistant' || message.role === 'tool'" class="message-content" v-html="renderedContent" />
      <div v-else class="message-content">{{ message.content }}</div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt({ html: false, breaks: true, linkify: true })

const props = defineProps({
  message: { type: Object, required: true },
})

const avatarText = computed(() => {
  switch (props.message.role) {
    case 'user': return '我'
    case 'assistant': return 'AI'
    case 'tool': return '🔧'
    case 'report': return '📋'
    default: return '?'
  }
})

const avatarColor = computed(() => {
  switch (props.message.role) {
    case 'user': return '#409EFF'
    case 'assistant': return '#67C23A'
    case 'tool': return '#E6A23C'
    case 'report': return '#F56C6C'
    default: return '#909399'
  }
})

const renderedContent = computed(() => {
  try {
    return md.render(props.message.content || '')
  } catch {
    return props.message.content
  }
})
</script>

<style scoped>
.message-bubble {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
}

.message-user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
}

.message-body {
  max-width: 75%;
}

.message-user .message-content {
  background: #409EFF;
  color: #fff;
  border-radius: 12px 12px 0 12px;
  padding: 10px 14px;
  word-break: break-word;
}

.message-assistant .message-content,
.message-tool .message-content {
  background: #f4f4f5;
  border-radius: 12px 12px 12px 0;
  padding: 10px 14px;
  word-break: break-word;
}

.message-assistant .message-content :deep(pre),
.message-tool .message-content :deep(pre) {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 8px 12px;
  border-radius: 6px;
  overflow-x: auto;
  font-size: 13px;
}

.message-assistant .message-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
}

.message-assistant .message-content :deep(th),
.message-assistant .message-content :deep(td) {
  border: 1px solid #ddd;
  padding: 6px 10px;
  text-align: left;
}
</style>
