<template>
  <div class="submission-detail-page">
    <div class="page-header">
      <h1>提交详情</h1>
      <router-link to="/my-submissions" class="btn btn-ghost btn-sm">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
        返回列表
      </router-link>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else-if="submission">
      <!-- 状态卡片 -->
      <div class="card status-card" :class="'card-' + (submission.Status || 'Pending').toLowerCase()">
        <div class="status-main">
          <div class="status-icon-wrapper" :class="'icon-' + (submission.Status || 'Pending').toLowerCase()">
            <template v-if="submission.Status === 'AC'">
              <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
            </template>
            <template v-else-if="submission.Status === 'WA'">
              <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
            </template>
            <template v-else-if="submission.Status === 'Pending'">
              <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            </template>
            <template v-else>
              <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            </template>
          </div>
          <div class="status-text-group">
            <h2 :class="'status-' + (submission.Status || 'Pending')">
              {{ submission.Status || 'Pending' }}
              <span class="status-desc-inline">{{ statusDesc }}</span>
            </h2>
          </div>
        </div>
        <div class="status-meta">
          <div class="meta-item">
            <span class="meta-label">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
              语言
            </span>
            <span class="meta-value lang-tag">{{ submission.Language || submission.language }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
              题目
            </span>
            <router-link :to="`/problems/${submission.ProblemID || submission.problem_id}`" class="meta-value">
              #{{ submission.ProblemID || submission.problem_id }}
            </router-link>
          </div>
          <div class="meta-item">
            <span class="meta-label">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              提交时间
            </span>
            <span class="meta-value">{{ formatTime(submission.CreatedAt || submission.created_at) }}</span>
          </div>
          <div class="meta-item" v-if="submission.TimeCost || submission.time_cost">
            <span class="meta-label">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
              耗时
            </span>
            <span class="meta-value">{{ submission.TimeCost || submission.time_cost }}ms</span>
          </div>
          <div class="meta-item" v-if="submission.MemoryCost || submission.memory_cost">
            <span class="meta-label">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>
              内存
            </span>
            <span class="meta-value">{{ submission.MemoryCost || submission.memory_cost }}KB</span>
          </div>
        </div>
      </div>

      <!-- 代码展示 -->
      <div class="card">
        <div class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
          提交代码
        </div>
        <pre class="code-block"><code>{{ submission.Code || submission.code }}</code></pre>
      </div>

      <!-- 运行输出 -->
      <div class="card" v-if="submission.ActualOutput || submission.actual_output">
        <div class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 15 15 9 21 9"/><path d="M17 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14"/></svg>
          运行输出
        </div>
        <pre class="output-block"><code>{{ submission.ActualOutput || submission.actual_output }}</code></pre>
      </div>

      <!-- AI 帮助 -->
      <div class="card ai-card" v-if="showAIButton">
        <div class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5"/><path d="M8.5 8.5v.01"/><path d="M16 15.5v.01"/><path d="M12 12v.01"/><path d="M11 17v.01"/><path d="M7 14v.01"/></svg>
          AI 导师诊断
        </div>
        <p class="ai-desc" v-if="!aiResult && !aiLoading">让 AI 帮你分析代码中的问题，给出改进建议。</p>
        <div class="ai-actions">
          <button class="btn btn-primary" @click="fetchAIHelp" :disabled="aiLoading">
            <span v-if="aiLoading" class="spinner"></span>
            <template v-else>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-5 4 4 0 0 1-5-5"/><path d="M8.5 8.5v.01"/><path d="M16 15.5v.01"/><path d="M12 12v.01"/><path d="M11 17v.01"/><path d="M7 14v.01"/></svg>
              让 AI 帮我分析
            </template>
          </button>
          <button v-if="aiResult" class="btn btn-outline btn-sm" @click="aiResult=''">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="4 17 9 22 17 9"/><path d="M14 14l-3 3"/></svg>
            清除结果
          </button>
        </div>
        <transition name="ai-fade">
          <div class="ai-result" v-if="aiResult || aiLoading">
            <div class="ai-content" v-text="aiResult || 'AI 正在思考中...'"></div>
            <span v-if="aiLoading" class="typing-cursor">▌</span>
          </div>
        </transition>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-icon">📭</span>
      <p class="empty-text">找不到该提交记录</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getSubmission } from '../api/index.js'

const route = useRoute()
const submission = ref(null)
const loading = ref(true)
const aiLoading = ref(false)
const aiResult = ref('')

const statusDesc = computed(() => {
  const s = submission.value?.Status || submission.value?.status
  const map = {
    'AC': '恭喜！代码通过了所有测试用例',
    'WA': '答案错误，输出与预期不符',
    'TLE': '运行超时，请优化算法复杂度',
    'MLE': '内存超限，请优化空间使用',
    'CE': '编译错误，请检查语法',
    'RE': '运行时错误',
    'SE': '系统错误，请联系管理员',
    'Pending': '评测进行中，请稍候...',
  }
  return map[s] || '未知状态'
})

const showAIButton = computed(() => {
  const s = submission.value?.Status || submission.value?.status
  return s && s !== 'AC' && s !== 'Pending'
})

function formatTime(t) {
  if (!t) return '—'
  return new Date(t).toLocaleString('zh-CN')
}

async function fetchSubmission() {
  loading.value = true
  try {
    const res = await getSubmission(route.params.id)
    const s = res.data
    submission.value = {
      ID: s.submission_id,
      ProblemID: s.problem_id,
      Status: s.status,
      Language: s.language,
      ActualOutput: s.actual_output,
      Code: s.code || '',
      TimeCost: s.time_cost,
      MemoryCost: s.memory_cost,
    }
  } catch (e) {
    submission.value = null
  } finally {
    loading.value = false
  }
}

async function fetchAIHelp() {
  aiLoading.value = true
  aiResult.value = ''
  try {
    const token = localStorage.getItem('token')
    const response = await fetch(`/api/submissions/${route.params.id}/ai-help`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    if (!response.ok) {
      const errText = await response.text()
      const errMatch = errText.match(/event:\s*error\s*\n*data:\s*(.*)/)
      throw new Error(errMatch ? errMatch[1] : 'AI 导师暂时无法连接')
    }
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''
      for (const line of lines) {
        if (line.startsWith('data:') && !line.includes('error')) {
          const chunk = line.slice(5).trim()
          if (chunk && chunk !== '[DONE]') {
            aiResult.value += chunk
          }
        }
      }
    }
  } catch (e) {
    aiResult.value = e.message || 'AI 导师暂时无法连接，请稍后重试'
  } finally {
    aiLoading.value = false
  }
}

onMounted(fetchSubmission)
</script>

<style scoped>
.status-card { margin-bottom: 24px; }
.status-main {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 24px;
}
.status-icon-wrapper {
  width: 64px; height: 64px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.icon-ac { background: #d1fae5; color: var(--success); }
.icon-wa { background: #fee2e2; color: var(--danger); }
.icon-pending { background: #eef2ff; color: var(--primary); animation: pulse 2s infinite; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: .7; } }
.icon-tle, .icon-mle { background: #fef3c7; color: #d97706; }
.icon-ce { background: #ede9fe; color: #7c3aed; }
.icon-error { background: #fee2e2; color: var(--danger); }

.status-text-group h2 { font-size: 24px; display: flex; flex-direction: column; gap: 4px; }
.status-desc-inline { font-size: 14px; font-weight: 400; color: var(--text-secondary); }

.status-meta {
  display: flex;
  gap: 28px;
  flex-wrap: wrap;
  padding-top: 20px;
  border-top: 1px solid var(--border);
}
.meta-item { display: flex; flex-direction: column; gap: 4px; }
.meta-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--text-light);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.meta-label svg { width: 12px; height: 12px; }
.meta-value { font-size: 14px; font-weight: 600; }
.lang-tag {
  display: inline-block;
  padding: 2px 10px;
  background: #f1f5f9;
  border-radius: 4px;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  font-weight: 700;
  color: var(--text-secondary);
}

.card-ac { border-left: 4px solid var(--success); }
.card-wa { border-left: 4px solid var(--danger); }
.card-pending { border-left: 4px solid var(--primary); }
.card-tle, .card-mle { border-left: 4px solid #f59e0b; }
.card-ce { border-left: 4px solid #8b5cf6; }

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 16px;
}
.card-title svg { color: var(--primary); flex-shrink: 0; }

.code-block, .output-block {
  background: #1e293b;
  color: #e2e8f0;
  padding: 20px;
  border-radius: var(--radius-sm);
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.7;
  font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
}
.output-block { background: #0f172a; color: #94a3b8; }

.ai-card { margin-top: 24px; }
.ai-desc { color: var(--text-secondary); font-size: 14px; margin-bottom: 16px; }
.ai-actions { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.ai-result { margin-top: 16px; }
.ai-fade-enter-active, .ai-fade-leave-active { transition: all .3s ease; }
.ai-fade-enter-from, .ai-fade-leave-to { opacity: 0; transform: translateY(8px); }
.ai-content {
  background: #f8fafc;
  border-radius: var(--radius-sm);
  padding: 20px;
  font-size: 14px;
  line-height: 1.8;
  white-space: pre-wrap;
  border: 1px solid var(--border);
}
.typing-cursor { animation: blink 1s step-end infinite; color: var(--primary); }
@keyframes blink { 50% { opacity: 0; } }
</style>