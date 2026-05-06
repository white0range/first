import { reactive } from 'vue'

function parseToken(token) {
  if (!token) return {}
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return {
      username: payload.username || '',
      role: payload.role || 0,
      user_id: payload.user_id || 0,
    }
  } catch { return {} }
}

const saved = parseToken(localStorage.getItem('token') || '')

// 全局共享的用户状态
export const store = reactive({
  token: localStorage.getItem('token') || '',
  username: localStorage.getItem('username') || saved.username || '',
  role: parseInt(localStorage.getItem('role') || String(saved.role || '0')),

  get isLoggedIn() {
    return !!this.token
  },

  get isAdmin() {
    return this.role === 1
  },

  login(token, username, role) {
    // 如果没有传入 username/role，从 token 中解析
    const parsed = parseToken(token)
    this.token = token
    this.username = username || parsed.username || ''
    this.role = role ?? parsed.role ?? 0
    localStorage.setItem('token', token)
    localStorage.setItem('username', this.username)
    localStorage.setItem('role', String(this.role))
  },

  logout() {
    this.token = ''
    this.username = ''
    this.role = 0
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('role')
  }
})
