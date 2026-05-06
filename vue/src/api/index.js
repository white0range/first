import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
})

// 请求拦截：自动带 token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：统一处理错误
api.interceptors.response.use(
  res => res,
  err => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('role')
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(err)
  }
)

// ========== 用户 ==========
export const register = (data) => api.post('/register', data)
export const login = (data) => api.post('/login', data)
export const getProfile = () => api.get('/profile')

// ========== 题目 ==========
export const getProblems = (params) => api.get('/problems', { params })
export const getProblemDetail = (id) => api.get(`/problems/${id}`)

// ========== 标签 ==========
export const getTags = () => api.get('/tags')

// ========== 提交 ==========
export const submitCode = (data) => api.post('/submit', data)
export const getSubmission = (id) => api.get(`/submissions/${id}`)
export const getMySubmissions = (params) => api.get('/my-submissions', { params })
export const getAIHelp = (id) => api.get(`/submissions/${id}/ai-help`)

// ========== 排行榜 ==========
export const getLeaderboard = (config) => api.get('/leaderboard', config)

// ========== 管理员 ==========
export const adminCreateProblem = (data) => api.post('/admin/problems', data)
export const adminUpdateProblem = (id, data) => api.put(`/admin/problems/${id}`, data)
export const adminDeleteProblem = (id) => api.delete(`/admin/problems/${id}`)
export const adminGetTestCases = (id) => api.get(`/admin/problems/${id}/cases`)
export const adminAddTestCase = (id, data) => api.post(`/admin/problems/${id}/cases`, data)
export const adminDeleteTestCase = (caseId) => api.delete(`/admin/problems/cases/${caseId}`)
export const adminCreateTag = (data) => api.post('/admin/tags', data)
export const adminDeleteTag = (id) => api.delete(`/admin/tags/${id}`)
export const adminUpdateProblemTags = (id, data) => api.put(`/admin/problems/${id}/tags`, data)

export default api
