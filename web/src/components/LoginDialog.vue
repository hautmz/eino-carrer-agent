<template>
  <el-dialog v-model="visible" title="登录 / 注册" width="400px" :close-on-click-modal="false">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="登录" name="login">
        <el-form :model="loginForm" @submit.prevent="handleLogin">
          <el-form-item label="用户名">
            <el-input v-model="loginForm.username" placeholder="请输入用户名" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" show-password />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%">登录</el-button>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="注册" name="register">
        <el-form :model="registerForm" @submit.prevent="handleRegister">
          <el-form-item label="用户名">
            <el-input v-model="registerForm.username" placeholder="3-50字符" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="registerForm.password" type="password" placeholder="6-50字符" show-password />
          </el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" style="width: 100%">注册</el-button>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'

const emit = defineEmits(['success'])
const userStore = useUserStore()

const visible = ref(false)
const activeTab = ref('login')
const loading = ref(false)

const loginForm = ref({ username: '', password: '' })
const registerForm = ref({ username: '', password: '' })

function open() {
  visible.value = true
}

async function handleLogin() {
  if (!loginForm.value.username || !loginForm.value.password) {
    ElMessage.warning('请填写用户名和密码')
    return
  }
  loading.value = true
  try {
    await userStore.login(loginForm.value.username, loginForm.value.password)
    ElMessage.success('登录成功')
    visible.value = false
    emit('success')
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  if (!registerForm.value.username || !registerForm.value.password) {
    ElMessage.warning('请填写用户名和密码')
    return
  }
  if (registerForm.value.password.length < 6) {
    ElMessage.warning('密码至少6个字符')
    return
  }
  loading.value = true
  try {
    await userStore.register(registerForm.value.username, registerForm.value.password)
    ElMessage.success('注册成功')
    visible.value = false
    emit('success')
  } catch {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

defineExpose({ open })
</script>
