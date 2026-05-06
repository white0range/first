<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <div class="auth-icon">🚀</div>
        <h2>创建账号</h2>
        <p>加入 GoJo，开启编程之旅</p>
      </div>

      <form @submit.prevent="handleRegister" class="auth-form">
        <div class="form-group">
          <label>用户名</label>
          <input v-model="form.username" class="input" placeholder="请设置用户名" required />
        </div>
        <div class="form-group">
          <label>密码</label>
          <input v-model="form.password" class="input" type="password" placeholder="请设置密码（至少6位）" required minlength="6" />
        </div>
        <div class="form-group">
          <label>确认密码</label>
          <input v-model="confirmPassword" class="input" type="password" placeholder="请再次输入密码" required />
        </div>

        <div class="error-msg" v-if="error">{{ error }}</div>
        <div class="success-msg" v-if="success">{{ success }}</div>

        <button type="submit" class="btn btn-primary auth-btn" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          <span v-else>注 册</span>
        </button>
      </form>

      <div class="auth-footer">
        已有账号？
        <router-link to="/login">立即登录 →</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { register } from '../api/index.js'

const router = useRouter()
const form = reactive({ username: '', password: '' })
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

async function handleRegister() {
  error.value = ''
  success.value = ''
  if (form.password !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }
  if (form.password.length < 6) {
    error.value = '密码长度至少6位'
    return
  }
  loading.value = true
  try {
    await register(form)
    success.value = '注册成功！即将跳转到登录页...'
    setTimeout(() => router.push('/login'), 1500)
  } catch (err) {
    error.value = err.response?.data?.error || '注册失败，该用户名可能已存在'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { display: flex; justify-content: center; align-items: center; min-height: calc(100vh - 160px); }
.auth-card { width: 100%; max-width: 400px; background: var(--bg-card); border-radius: var(--radius); border: 1px solid var(--border); padding: 40px 32px; box-shadow: var(--shadow-lg); }
.auth-header { text-align: center; margin-bottom: 32px; }
.auth-icon { font-size: 48px; margin-bottom: 8px; }
.auth-header h2 { font-size: 22px; margin-bottom: 4px; }
.auth-header p { color: var(--text-secondary); font-size: 14px; }
.auth-form { display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group label { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.auth-btn { width: 100%; padding: 12px; font-size: 15px; margin-top: 8px; }
.error-msg { color: var(--danger); font-size: 13px; text-align: center; background: #fef2f2; padding: 10px; border-radius: 8px; }
.success-msg { color: var(--success); font-size: 13px; text-align: center; background: #f0fdf4; padding: 10px; border-radius: 8px; }
.auth-footer { text-align: center; margin-top: 20px; font-size: 14px; color: var(--text-secondary); }
</style>
