<template>
  <div class="problem-detail">
    <div class="loading-center" v-if="loading">
      <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else-if="problem">
      <!-- 题目头部 -->
      <div class="detail-header">
        <div class="detail-header-left">
          <div class="detail-breadcrumb">
            <router-link to="/">题库</router-link>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
            <span>{{ problem.Title || problem.title }}</span>
          </div>
          <h1>{{ problem.Title || problem.title }}</h1>
          <div class="detail-meta">
            <span class="meta-chip">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              {{ problem.TimeLimit || problem.time_limit }}ms
            </span>
            <span class="meta-chip">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
              {{ problem.MemoryLimit || problem.memory_limit }}MB
            </span>
            <span v-if="problem.IsAC || problem.is_ac" class="badge badge-success">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
              已通过
            </span>
          </div>
          <div class="detail-tags" v-if="(problem.Tags || problem.tags || []).length">
            <span class="mini-tag" v-for="tag in (problem.Tags || problem.tags || [])" :key="tag.ID || tag.id">
              {{ tag.Name || tag.name }}
            </span>
          </div>
        </div>
        <div class="detail-header-right">
          <div class="problem-stat-cards">
            <div class="stat-card">
              <span class="stat-card-value">{{ problem.SubmitCount ?? problem.submit_count ?? 0 }}</span>
              <span class="stat-card-label">提交</span>
            </div>
            <div class="stat-card">
              <span class="stat-card-value">{{ problem.AcceptedCount ?? problem.accepted_count ?? 0 }}</span>
              <span class="stat-card-label">通过</span>
            </div>
            <div class="stat-card">
              <span class="stat-card-value">{{ getRate }}%</span>
              <span class="stat-card-label">通过率</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 题目描述 -->
      <div class="card description-card">
        <div class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
          题目描述
        </div>
        <div class="markdown-body" v-html="renderedDescription"></div>
      </div>

      <!-- 代码提交区 -->
      <div class="card submit-card">
        <div class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
          提交代码
        </div>
        <div v-if="!store.isLoggedIn" class="login-hint">
          <p>请先 <router-link to="/login">登录</router-link> 后再提交代码</p>
        </div>
        <template v-else>
          <div class="submit-form">
            <div class="submit-bar">
              <div class="lang-select-wrapper">
                <svg class="lang-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                <select v-model="language" class="input lang-select">
                  <option value="go">Go</option>
                  <option value="python">Python</option>
                  <option value="java">Java</option>
                  <option value="cpp">C++</option>
                </select>
              </div>
              <button class="btn btn-primary" @click="handleSubmit" :disabled="submitting || !code.trim()">
                <span v-if="submitting" class="spinner"></span>
                <template v-else>
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
                  提交运行
                </template>
              </button>
            </div>
            <div class="editor-wrapper">
              <div class="editor-header">
                <span class="editor-lang-label">{{ languageLabel }}</span>
                <button class="btn btn-ghost btn-sm" @click="resetCode" title="重置代码">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                  重置
                </button>
              </div>
              <textarea
                v-model="code"
                class="input code-editor"
                :placeholder="codePlaceholder"
                rows="16"
                spellcheck="false"
              ></textarea>
            </div>
          </div>

          <!-- 提交结果反馈 -->
          <transition name="result-slide">
            <div class="submit-result" v-if="submitResult">
              <div class="result-card" :class="'result-' + (submitResult.status || 'Pending').toLowerCase()">
                <div class="result-icon">
                  <template v-if="submitResult.status === 'AC'">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                  </template>
                  <template v-else-if="submitResult.status === 'WA'">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
                  </template>
                  <template v-else-if="submitResult.status === 'Pending'">
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                  </template>
                  <template v-else>
                    <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                  </template>
                </div>
                <div class="result-info">
                  <span class="result-status" :class="'status-' + (submitResult.status || 'Pending')">
                    {{ submitResult.status || 'Pending' }}
                  </span>
                  <span class="result-msg">{{ submitResult.message }}</span>
                </div>
                <router-link
                  v-if="submitResult.submission_id"
                  :to="`/submissions/${submitResult.submission_id}`"
                  class="btn btn-outline btn-sm"
                >
                  查看详情
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
                </router-link>
              </div>
            </div>
          </transition>
        </template>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-icon">📭</span>
      <p class="empty-text">题目不存在或已被删除</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { getProblemDetail, submitCode } from '../api/index.js'
import { store } from '../store/index.js'

const route = useRoute()
const problem = ref(null)
const loading = ref(true)
const language = ref('go')
const code = ref('')
const submitting = ref(false)
const submitResult = ref(null)
let ws = null

const languageLabel = computed(() => {
  const labels = { go: 'Go', python: 'Python', java: 'Java', cpp: 'C++' }
  return labels[language.value] || 'Go'
})

const getRate = computed(() => {
  if (!problem.value) return 0
  const sub = problem.value.SubmitCount ?? problem.value.submit_count ?? 0
  const acc = problem.value.AcceptedCount ?? problem.value.accepted_count ?? 0
  if (sub === 0) return 0
  return Math.round((acc / sub) * 100)
})

const codePlaceholder = computed(() => {
  const templates = {
    go: 'package main\n\nimport "fmt"\n\nfunc main() {\n\t// 在这里写下你的代码\n\tfmt.Println("Hello GoJo!")\n}',
    python: '# 在这里写下你的 Python 代码\nprint("Hello GoJo!")',
    java: 'public class Main {\n    public static void main(String[] args) {\n        // 在这里写下你的代码\n        System.out.println("Hello GoJo!");\n    }\n}',
    cpp: '#include <iostream>\nusing namespace std;\n\nint main() {\n    // 在这里写下你的代码\n    cout << "Hello GoJo!" << endl;\n    return 0;\n}',
  }
  return templates[language.value] || '// 在这里写下你的代码'
})

const renderedDescription = computed(() => {
  const desc = problem.value?.Description || problem.value?.description || ''
  return desc.replace(/\n/g, '<br>').replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
})

function resetCode() {
  code.value = ''
}

async function fetchProblem() {
  loading.value = true
  try {
    const res = await getProblemDetail(route.params.id)
    problem.value = res.data.data || res.data
  } catch (e) {
    problem.value = null
  } finally {
    loading.value = false
  }
}

function connectWS(submissionId) {
  const token = store.token
  if (!token) return
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${window.location.host}/api/ws?token=${encodeURIComponent(token)}`
  try {
    ws = new WebSocket(url)
    ws.onopen = () => {
      const ping = setInterval(() => { if (ws?.readyState === 1) ws.send('ping') }, 15000)
      ws._pingInterval = ping
    }
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data)
        if (msg.type === 'JUDGE_RESULT' && msg.submission_id === submissionId) {
          submitResult.value = {
            ...submitResult.value,
            status: msg.status,
            message: msg.message || '评测完成！',
          }
          if (msg.status === 'AC') fetchProblem()
          closeWS()
        }
      } catch {}
    }
    ws.onerror = () => {}
    ws.onclose = () => {
      if (ws?._pingInterval) clearInterval(ws._pingInterval)
    }
  } catch {}
}

function closeWS() {
  if (ws) {
    if (ws._pingInterval) clearInterval(ws._pingInterval)
    ws.close()
    ws = null
  }
}

async function handleSubmit() {
  submitting.value = true
  submitResult.value = null
  try {
    const res = await submitCode({
      problem_id: parseInt(route.params.id),
      language: language.value,
      code: code.value,
    })
    submitResult.value = { status: 'Pending', message: res.data.message, submission_id: res.data.submission_id }
    connectWS(res.data.submission_id)
  } catch (err) {
    submitResult.value = {
      status: 'Error',
      message: err.response?.data?.error || '提交失败，请稍后重试',
    }
  } finally {
    submitting.value = false
  }
}

onMounted(fetchProblem)
onUnmounted(closeWS)
</script>

<style scoped>
.detail-header {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}
.detail-header-left { flex: 1; min-width: 0; }
.detail-breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-light);
  margin-bottom: 8px;
}
.detail-breadcrumb a { color: var(--text-light) !important; }
.detail-breadcrumb a:hover { color: var(--primary) !important; }
.detail-breadcrumb svg { width: 12px; height: 12px; flex-shrink: 0; }
.detail-header h1 { font-size: 26px; font-weight: 800; margin-bottom: 12px; letter-spacing: -0.3px; }
.detail-meta { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
.meta-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 6px;
  background: #f1f5f9;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
}
.meta-chip svg { flex-shrink: 0; }
.detail-tags { margin-top: 12px; display: flex; gap: 6px; flex-wrap: wrap; }

.detail-header-right { flex-shrink: 0; }
.problem-stat-cards {
  display: flex;
  gap: 12px;
}
.stat-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 14px 20px;
  background: #f8fafc;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  min-width: 70px;
}
.stat-card-value { font-size: 22px; font-weight: 800; color: var(--primary); line-height: 1.2; }
.stat-card-label { font-size: 11px; color: var(--text-light); font-weight: 500; }

.description-card { margin-bottom: 24px; }
.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 16px;
}
.card-title svg { color: var(--primary); flex-shrink: 0; }
.markdown-body { font-size: 15px; line-height: 1.8; color: var(--text); }
.markdown-body pre {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 12px 0;
  font-size: 13px;
}

.submit-card { margin-bottom: 24px; }
.login-hint {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
  font-size: 15px;
  background: #f8fafc;
  border-radius: var(--radius-sm);
}
.submit-form { }
.submit-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
  align-items: center;
}
.lang-select-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}
.lang-select-wrapper .lang-icon {
  position: absolute;
  left: 12px;
  color: var(--text-light);
  pointer-events: none;
}
.lang-select {
  padding-left: 36px !important;
  width: 150px;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%2394a3b8' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px !important;
}

.editor-wrapper {
  border: 1.5px solid #334155;
  border-radius: var(--radius-sm);
  overflow: hidden;
  transition: border-color var(--transition);
}
.editor-wrapper:focus-within { border-color: var(--primary); }
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 14px;
  background: #1a1f2e;
  border-bottom: 1px solid #2d3548;
}
.editor-lang-label {
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.editor-header .btn-ghost { color: #94a3b8; padding: 4px 10px; font-size: 12px; }
.editor-header .btn-ghost:hover { color: #e2e8f0; background: #2d3548; }
.code-editor {
  font-family: 'Fira Code', 'Consolas', 'Monaco', 'Cascadia Code', monospace;
  font-size: 13px;
  line-height: 1.7;
  background: #1e293b;
  color: #e2e8f0;
  border: none !important;
  border-radius: 0 !important;
  resize: vertical;
  padding: 14px !important;
  tab-size: 4;
}
.code-editor:focus { box-shadow: none !important; }
.code-editor::placeholder { color: #475569; }

/* Submit Result */
.submit-result { margin-top: 16px; }
.result-slide-enter-active { animation: slideUp .3s ease; }
.result-slide-leave-active { animation: slideUp .2s ease reverse; }
@keyframes slideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
.result-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 22px;
  border-radius: var(--radius);
  flex-wrap: wrap;
}
.result-card svg { flex-shrink: 0; }
.result-ac { background: #f0fdf4; border: 1px solid #bbf7d0; }
.result-ac svg { color: var(--success); }
.result-wa, .result-error { background: #fef2f2; border: 1px solid #fecaca; }
.result-wa svg, .result-error svg { color: var(--danger); }
.result-pending { background: #eef2ff; border: 1px solid #c7d2fe; }
.result-pending svg { color: var(--primary); }
.result-icon { display: flex; align-items: center; }
.result-info { display: flex; flex-direction: column; flex: 1; min-width: 0; }
.result-status { font-size: 20px; font-weight: 700; }
.result-msg { font-size: 13px; color: var(--text-secondary); margin-top: 2px; }

@media (max-width: 768px) {
  .detail-header { flex-direction: column; }
  .problem-stat-cards { justify-content: flex-start; }
  .stat-card { min-width: 60px; padding: 10px 14px; }
  .stat-card-value { font-size: 18px; }
}
@media (max-width: 640px) {
  .detail-header h1 { font-size: 20px; }
  .submit-bar { flex-direction: column; align-items: stretch; }
  .lang-select-wrapper { width: 100%; }
  .lang-select { width: 100%; }
  .submit-bar .btn { width: 100%; }
}
</style>
