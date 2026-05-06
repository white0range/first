<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <div class="auth-icon">⚡</div>
        <h2>欢迎回来</h2>
        <p>登录 GoJo，继续你的编程之旅</p>
      </div>

      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="form-group">
          <label>👤 用户名</label>
          <input v-model="form.username" class="input" placeholder="请输入用户名" autocomplete="username" required />
        </div>
        <div class="form-group">
          <label>🔒 密码</label>
          <input v-model="form.password" class="input" type="password" placeholder="请输入密码" autocomplete="current-password" required />
        </div>

        <div class="error-msg" v-if="error">⚠️ {{ error }}</div>

        <button type="submit" class="btn btn-primary auth-btn" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          <span v-else>🚀 登 录</span>
        </button>
      </form>

      <div class="auth-footer">
        还没有账号？
        <router-link to="/register">立即注册 →</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../api/index.js'
import { store } from '../store/index.js'

const router = useRouter()
const form = reactive({ username: '', password: '' })
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    const res = await login(form)
    store.login(res.data.token, form.username)
    // 重定向到之前想访问的页面
    const redirect = router.currentRoute.value.query.redirect || '/'
    router.push(redirect)
  } catch (err) {
    error.value = err.response?.data?.error || '登录失败，请检查账号密码'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 160px);
}
.auth-card {
  width: 100%;
  max-width: 400px;
  background: var(--bg-card);
  border-radius: var(--radius);
  border: 1px solid var(--border);
  padding: 40px 32px;
  box-shadow: var(--shadow-lg);
}
.auth-header { text-align: center; margin-bottom: 32px; }
.auth-icon { font-size: 48px; margin-bottom: 8px; }
.auth-header h2 { font-size: 22px; margin-bottom: 4px; }
.auth-header p { color: var(--text-secondary); font-size: 14px; }
.auth-form { display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.auth-btn { width: 100%; padding: 12px; font-size: 15px; margin-top: 8px; }
.error-msg { color: var(--danger); font-size: 13px; text-align: center; background: #fef2f2; padding: 10px; border-radius: 8px; }
.auth-footer { text-align: center; margin-top: 20px; font-size: 14px; color: var(--text-secondary); }
</style>
