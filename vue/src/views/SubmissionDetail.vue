<template>
  <div class="submission-detail-page">
    <div class="page-header">
      <h1>📝 提交详情</h1>
      <router-link to="/my-submissions" class="btn btn-outline btn-sm">← 返回列表</router-link>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <template v-else-if="submission">
      <!-- 状态卡片 -->
      <div class="card status-card" :class="'card-' + (submission.Status || submission.status || 'Pending').toLowerCase()">
        <div class="status-main">
          <span class="status-big-icon">
            <template v-if="submission.Status === 'AC' || submission.status === 'AC'">🎉</template>
            <template v-else-if="submission.Status === 'WA' || submission.status === 'WA'">❌</template>
            <template v-else-if="submission.Status === 'Pending' || submission.status === 'Pending'">⏳</template>
            <template v-else>⚠️</template>
          </span>
          <div>
            <h2 :class="'status-' + (submission.Status || submission.status || 'Pending')">
              {{ submission.Status || submission.status || 'Pending' }}
            </h2>
            <p class="status-desc">{{ statusDesc }}</p>
          </div>
        </div>
        <div class="status-meta">
          <div class="meta-item">
            <span class="meta-label">语言</span>
            <span class="meta-value">{{ submission.Language || submission.language }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">题目</span>
            <router-link :to="`/problems/${submission.ProblemID || submission.problem_id}`">
              #{{ submission.ProblemID || submission.problem_id }}
            </router-link>
          </div>
          <div class="meta-item">
            <span class="meta-label">提交时间</span>
            <span class="meta-value">{{ formatTime(submission.CreatedAt || submission.created_at) }}</span>
          </div>
          <div class="meta-item" v-if="submission.TimeCost || submission.time_cost">
            <span class="meta-label">耗时</span>
            <span class="meta-value">{{ submission.TimeCost || submission.time_cost }}ms</span>
          </div>
          <div class="meta-item" v-if="submission.MemoryCost || submission.memory_cost">
            <span class="meta-label">内存</span>
            <span class="meta-value">{{ submission.MemoryCost || submission.memory_cost }}KB</span>
          </div>
        </div>
      </div>

      <!-- 代码展示 -->
      <div class="card">
        <h3>💻 提交代码</h3>
        <pre class="code-block"><code>{{ submission.Code || submission.code }}</code></pre>
      </div>

      <!-- 错误输出（如果有） -->
      <div class="card" v-if="submission.ActualOutput || submission.actual_output">
        <h3>📤 运行输出</h3>
        <pre class="output-block"><code>{{ submission.ActualOutput || submission.actual_output }}</code></pre>
      </div>

      <!-- AI 帮助 -->
      <div class="card ai-card" v-if="showAIButton">
        <h3>🤖 AI 导师诊断</h3>
        <p v-if="!aiResult && !aiLoading">让 AI 帮你分析代码中的问题，给出改进建议。</p>
        <button class="btn btn-primary" @click="fetchAIHelp" :disabled="aiLoading">
          <span v-if="aiLoading" class="spinner"></span>
          <span v-else>🔍 让 AI 帮你分析</span>
        </button>
        <button v-if="aiResult" class="btn btn-outline btn-sm" style="margin-left:8px" @click="aiResult=''">清除</button>
        <div class="ai-result" v-if="aiResult || aiLoading">
          <div class="ai-content" v-text="aiResult || 'AI 正在思考中...'"></div>
          <span v-if="aiLoading" class="typing-cursor">▌</span>
        </div>
      </div>
    </template>

    <div class="empty" v-else>
      <span>📭</span>
      <p>找不到该提交记录</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getSubmission, getAIHelp } from '../api/index.js'

const route = useRoute()
const submission = ref(null)
const loading = ref(true)
const aiLoading = ref(false)
const aiResult = ref('')

const statusDesc = computed(() => {
  const s = submission.value?.Status || submission.value?.status
  const map = {
    'AC': '恭喜！代码通过了所有测试用例 ✅',
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
    // 后端返回 {message, submission_id, problem_id, status, actual_output, language} 在顶层
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
      // 尝试从 SSE error 事件中提取错误
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
      // 解析 SSE 事件
      const lines = buffer.split('\n')
      buffer = lines.pop() || '' // 保留未完成的行

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
.status-card {
  margin-bottom: 24px;
}
.status-main {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}
.status-big-icon { font-size: 40px; }
.status-main h2 { font-size: 24px; }
.status-desc { font-size: 14px; color: var(--text-secondary); }

.status-meta {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
  padding-top: 16px;
  border-top: 1px solid var(--border);
}
.meta-item { display: flex; flex-direction: column; gap: 2px; }
.meta-label { font-size: 11px; color: var(--text-light); text-transform: uppercase; }
.meta-value { font-size: 14px; font-weight: 500; }

.card-ac { border-left: 4px solid var(--success); }
.card-wa { border-left: 4px solid var(--danger); }
.card-pending { border-left: 4px solid var(--primary); }

.code-block, .output-block {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.7;
  margin-top: 12px;
}
.output-block { background: #0f172a; color: #94a3b8; }

.ai-card { margin-top: 24px; }
.ai-card h3 { margin-bottom: 8px; }
.ai-card p { color: var(--text-secondary); font-size: 14px; margin-bottom: 12px; }
.ai-result { margin-top: 16px; }
.ai-content {
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
  font-size: 14px;
  line-height: 1.8;
  white-space: pre-wrap;
  border: 1px solid var(--border);
}
.typing-cursor {
  animation: blink 1s step-end infinite;
  color: var(--primary);
  font-weight: 700;
}
@keyframes blink { 0%,100% { opacity:1; } 50% { opacity:0; } }

.empty { text-align: center; padding: 80px; color: var(--text-light); }
.empty span { font-size: 48px; display: block; margin-bottom: 8px; }
</style>
