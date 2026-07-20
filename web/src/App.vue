<template>
  <div class="app-container">
    <div v-if="!userStore.isLoggedIn" class="login-prompt">
      <div class="login-card">
        <h1>Eino Career Agent</h1>
        <p>AI 职业规划师 — 智能对话，精准规划</p>
        <el-button type="primary" size="large" @click="loginDialogRef?.open()">
          登录 / 注册
        </el-button>
      </div>
      <LoginDialog ref="loginDialogRef" @success="onLoginSuccess" />
    </div>

    <div v-else class="chat-layout">
      <aside class="sidebar">
        <div class="sidebar-header">
          <span class="app-title">Eino Career</span>
          <el-button :icon="Plus" size="small" circle @click="newChat" />
        </div>
        <div class="conv-list">
          <div
            v-for="conv in chatStore.conversations"
            :key="conv.id"
            :class="['conv-item', { active: conv.id === chatStore.currentConvId }]"
            @click="selectConv(conv.id)"
          >
            <span class="conv-title">{{ conv.title || '新对话' }}</span>
            <el-button
              :icon="Delete"
              size="small"
              circle
              text
              @click.stop="removeConv(conv.id)"
            />
          </div>
        </div>
        <div class="sidebar-footer">
          <span class="username">{{ userStore.username }}</span>
          <el-button text @click="handleLogout">退出</el-button>
        </div>
      </aside>

      <main class="chat-main">
        <ChatWindow />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import { useUserStore } from './stores/user'
import { useChatStore } from './stores/chat'
import LoginDialog from './components/LoginDialog.vue'
import ChatWindow from './components/ChatWindow.vue'

const userStore = useUserStore()
const chatStore = useChatStore()
const loginDialogRef = ref(null)

onMounted(() => {
  if (userStore.isLoggedIn) {
    chatStore.loadConversations()
  }
})

function onLoginSuccess() {
  chatStore.loadConversations()
}

function newChat() {
  chatStore.clearCurrentChat()
}

function selectConv(id) {
  chatStore.selectConversation(id)
}

async function removeConv(id) {
  await chatStore.removeConversation(id)
}

function handleLogout() {
  userStore.logout()
  chatStore.conversations = []
  chatStore.clearCurrentChat()
}
</script>

<style scoped>
.app-container {
  height: 100vh;
  width: 100vw;
}

.login-prompt {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  background: #fff;
  border-radius: 16px;
  padding: 48px 40px;
  text-align: center;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}

.login-card h1 {
  font-size: 28px;
  margin-bottom: 8px;
  color: #303133;
}

.login-card p {
  color: #909399;
  margin-bottom: 24px;
}

.chat-layout {
  display: flex;
  height: 100%;
}

.sidebar {
  width: 260px;
  background: #f5f7fa;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.app-title {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
}

.conv-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.conv-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 4px;
  transition: background 0.2s;
}

.conv-item:hover {
  background: #e8eaed;
}

.conv-item.active {
  background: #409EFF;
  color: #fff;
}

.conv-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  font-size: 14px;
}

.sidebar-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-top: 1px solid #e4e7ed;
}

.username {
  font-size: 14px;
  color: #606266;
}

.chat-main {
  flex: 1;
  overflow: hidden;
}
</style>
