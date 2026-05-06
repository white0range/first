<template>
  <div class="problem-detail">
    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <template v-else-if="problem">
      <!-- 题目头部 -->
      <div class="detail-header">
        <div>
          <h1>{{ problem.Title || problem.title }}</h1>
          <div class="detail-meta">
            <span>⏱ {{ problem.TimeLimit || problem.time_limit }}ms</span>
            <span>💾 {{ problem.MemoryLimit || problem.memory_limit }}MB</span>
            <span
              v-if="problem.IsAC || problem.is_ac"
              class="badge badge-success"
            >✅ 已通过</span>
          </div>
        </div>
        <div class="detail-tags" v-if="(problem.Tags || problem.tags || []).length">
          <span class="mini-tag" v-for="tag in (problem.Tags || problem.tags || [])" :key="tag.ID || tag.id">
            {{ tag.Name || tag.name }}
          </span>
        </div>
      </div>

      <!-- 题目描述 -->
      <div class="card description-card">
        <h3>📖 题目描述</h3>
        <div class="markdown-body" v-html="renderedDescription"></div>
      </div>

      <!-- 代码提交区 -->
      <div class="card submit-card">
        <h3>💻 提交代码</h3>
        <div v-if="!store.isLoggedIn" class="login-hint">
          <p>请先 <router-link to="/login">登录</router-link> 后再提交代码</p>
        </div>
        <template v-else>
          <div class="submit-form">
            <div class="submit-bar">
              <select v-model="language" class="input" style="width:160px">
                <option value="go">Go</option>
                <option value="python">Python</option>
                <option value="java">Java</option>
                <option value="cpp">C++</option>
              </select>
              <button class="btn btn-primary" @click="handleSubmit" :disabled="submitting || !code.trim()">
                <span v-if="submitting" class="spinner"></span>
                <span v-else>🚀 提交运行</span>
              </button>
            </div>
            <textarea
              v-model="code"
              class="input code-editor"
              :placeholder="codePlaceholder"
              rows="16"
            ></textarea>
          </div>

          <!-- 提交结果反馈 -->
          <div class="submit-result" v-if="submitResult">
            <div class="result-card" :class="'result-' + (submitResult.status || 'Pending').toLowerCase()">
              <div class="result-icon">
                <template v-if="submitResult.status === 'AC'">🎉</template>
                <template v-else-if="submitResult.status === 'WA'">❌</template>
                <template v-else-if="submitResult.status === 'Pending'">⏳</template>
                <template v-else>⚠️</template>
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
                查看详情 →
              </router-link>
            </div>
          </div>
        </template>
      </div>
    </template>

    <div class="empty" v-else>
      <span>📭</span>
      <p>题目不存在或已被删除</p>
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
// WebSocket
let ws = null
let wsRetry = 0

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
      wsRetry = 0
      // 发送心跳
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
          if (msg.status === 'AC') {
            // 刷新题目信息以更新 AC 标记
            fetchProblem()
          }
          // 评测完成后关闭 WebSocket
          closeWS()
        }
      } catch {}
    }
    ws.onerror = () => { /* ignore */ }
    ws.onclose = () => {
      if (ws?._pingInterval) clearInterval(ws._pingInterval)
    }
  } catch { /* ignore */ }
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
    // 连接 WebSocket 等待实时推送
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
.detail-header { margin-bottom: 24px; }
.detail-header h1 { font-size: 24px; margin-bottom: 8px; }
.detail-meta { display: flex; gap: 16px; align-items: center; color: var(--text-secondary); font-size: 13px; flex-wrap: wrap; }
.detail-tags { margin-top: 12px; display: flex; gap: 6px; flex-wrap: wrap; }

.mini-tag {
  padding: 4px 10px;
  border-radius: 100px;
  background: #eef2ff;
  color: var(--primary);
  font-size: 12px;
  font-weight: 500;
}

.description-card { margin-bottom: 24px; }
.description-card h3 { margin-bottom: 16px; font-size: 16px; }
.markdown-body { font-size: 15px; line-height: 1.8; }
.markdown-body pre { background: #1e293b; color: #e2e8f0; padding: 16px; border-radius: 8px; overflow-x: auto; margin: 12px 0; }

.submit-card h3 { margin-bottom: 16px; font-size: 16px; }

.submit-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.code-editor {
  font-family: 'Fira Code', 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.7;
  background: #1e293b;
  color: #e2e8f0;
  border-color: #334155;
}
.code-editor:focus { border-color: var(--primary); }

.login-hint { text-align: center; padding: 32px; color: var(--text-secondary); }

/* 提交结果 */
.submit-result { margin-top: 16px; }
.result-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  border-radius: var(--radius);
  flex-wrap: wrap;
}
.result-ac { background: #f0fdf4; border: 1px solid #bbf7d0; }
.result-wa, .result-error { background: #fef2f2; border: 1px solid #fecaca; }
.result-pending { background: #eef2ff; border: 1px solid #c7d2fe; }
.result-icon { font-size: 28px; }
.result-info { display: flex; flex-direction: column; flex: 1; }
.result-status { font-size: 18px; }
.result-msg { font-size: 13px; color: var(--text-secondary); }

.empty { text-align: center; padding: 80px 20px; color: var(--text-light); }
.empty span { font-size: 64px; display: block; margin-bottom: 12px; }
</style>
