<template>
  <div class="auth-page">
    <div class="auth-bg-shapes">
      <div class="auth-shape auth-shape-1"></div>
      <div class="auth-shape auth-shape-2"></div>
    </div>
    <div class="auth-card">
      <div class="auth-header">
        <router-link to="/" class="auth-logo">⚡ GoJo</router-link>
        <h2>欢迎回来</h2>
        <p>登录账号，继续你的编程之旅</p>
      </div>

      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="form-group">
          <label for="username">用户名</label>
          <div class="input-wrapper">
            <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            <input id="username" v-model="form.username" class="input" placeholder="请输入用户名" autocomplete="username" required />
          </div>
        </div>
        <div class="form-group">
          <label for="password">密码</label>
          <div class="input-wrapper">
            <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
            <input id="password" v-model="form.password" class="input" type="password" placeholder="请输入密码" autocomplete="current-password" required />
          </div>
        </div>

        <div class="error-msg" v-if="error">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
          {{ error }}
        </div>

        <button type="submit" class="btn btn-primary auth-btn" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          <span v-else>
            登录
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
          </span>
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
  position: relative;
  padding: 20px;
}
.auth-bg-shapes {
  position: fixed; inset: 0; overflow: hidden; pointer-events: none; z-index: 0;
}
.auth-shape {
  position: absolute;
  border-radius: 50%;
  opacity: .06;
}
.auth-shape-1 {
  width: 500px; height: 500px;
  background: radial-gradient(circle, var(--primary), transparent);
  top: -150px; right: -100px;
}
.auth-shape-2 {
  width: 400px; height: 400px;
  background: radial-gradient(circle, var(--accent), transparent);
  bottom: -100px; left: -80px;
}

.auth-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  background: var(--bg-card);
  border-radius: var(--radius);
  border: 1px solid var(--border);
  padding: 44px 36px;
  box-shadow: var(--shadow-xl);
  animation: cardFloat 0.5s ease;
}
@keyframes cardFloat {
  from { opacity: 0; transform: translateY(20px) scale(.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
.auth-logo {
  display: inline-block;
  font-size: 18px;
  font-weight: 800;
  color: var(--text) !important;
  margin-bottom: 20px;
}
.auth-header { text-align: center; margin-bottom: 32px; }
.auth-header h2 { font-size: 24px; font-weight: 800; margin-bottom: 6px; }
.auth-header p { color: var(--text-secondary); font-size: 14px; }

.auth-form { display: flex; flex-direction: column; gap: 18px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 13px; font-weight: 600; color: var(--text-secondary); }

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}
.input-wrapper .input {
  padding-left: 40px;
}
.input-icon {
  position: absolute;
  left: 12px;
  color: var(--text-light);
  pointer-events: none;
  flex-shrink: 0;
}

.auth-btn {
  width: 100%;
  padding: 13px;
  font-size: 15px;
  margin-top: 4px;
  border-radius: 12px;
}
.auth-btn svg { transition: transform var(--transition); }
.auth-btn:hover:not(:disabled) svg { transform: translateX(4px); }

.error-msg {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--danger);
  font-size: 13px;
  background: #fef2f2;
  padding: 12px 14px;
  border-radius: var(--radius-sm);
  border: 1px solid #fecaca;
}
.error-msg svg { flex-shrink: 0; }

.auth-footer {
  text-align: center;
  margin-top: 24px;
  font-size: 14px;
  color: var(--text-secondary);
}
</style>
