<template>
  <div class="submissions-page">
    <div class="page-header">
      <h1>我的提交记录</h1>
      <span class="subtitle" v-if="!loading && total > 0">共 {{ total }} 条记录</span>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else-if="submissions.length > 0">
      <div class="card" style="padding:0;overflow:hidden">
        <div class="sub-table">
          <div class="sub-header">
            <span class="col-id">#</span>
            <span class="col-problem">题目</span>
            <span class="col-lang">语言</span>
            <span class="col-status">状态</span>
            <span class="col-time">时间</span>
            <span class="col-action"></span>
          </div>
          <div class="sub-row" v-for="item in submissions" :key="item.ID || item.id">
            <span class="col-id">#{{ item.ID || item.id }}</span>
            <span class="col-problem">
              <router-link :to="`/problems/${item.ProblemID || item.problem_id}`">
                {{ item.ProblemTitle || item.problem_title || `#${item.ProblemID || item.problem_id}` }}
              </router-link>
            </span>
            <span class="col-lang">
              <span class="lang-badge">{{ item.Language || item.language }}</span>
            </span>
            <span class="col-status">
              <span class="status-indicator" :class="'indicator-' + (item.Status || item.status || 'Pending').toLowerCase()"></span>
              <span :class="'status-' + (item.Status || item.status || 'Pending')">
                {{ item.Status || item.status || 'Pending' }}
              </span>
            </span>
            <span class="col-time">{{ formatTime(item.CreatedAt || item.created_at) }}</span>
            <span class="col-action">
              <router-link
                :to="`/submissions/${item.ID || item.id}`"
                class="btn btn-ghost btn-sm"
              >
                详情
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
              </router-link>
            </span>
          </div>
        </div>
      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button :disabled="page <= 1" @click="goPage(page - 1)">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
          上一页
        </button>
        <button v-for="p in visiblePages" :key="p" :class="{ active: p === page }" @click="goPage(p)">{{ p }}</button>
        <button :disabled="page >= totalPages" @click="goPage(page + 1)">
          下一页
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
        </button>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-icon">📭</span>
      <p class="empty-text">还没有提交记录</p>
      <p class="empty-hint">快去刷题，提交你的第一份代码吧！</p>
      <router-link to="/" class="btn btn-primary" style="margin-top:16px">去刷题</router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getMySubmissions } from '../api/index.js'

const submissions = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const loading = ref(true)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit.value)))
const visiblePages = computed(() => {
  const pages = []
  const max = totalPages.value
  const start = Math.max(1, page.value - 2)
  const end = Math.min(max, page.value + 2)
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
})

function formatTime(t) {
  if (!t) return '—'
  return new Date(t).toLocaleString('zh-CN')
}

async function fetchSubmissions() {
  loading.value = true
  try {
    const res = await getMySubmissions({ page: page.value, limit: limit.value })
    const data = res.data.data || res.data
    submissions.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取提交记录失败', e)
  } finally {
    loading.value = false
  }
}

function goPage(p) { page.value = p; fetchSubmissions(); }

onMounted(fetchSubmissions)
</script>

<style scoped>
.subtitle { color: var(--text-light); font-size: 14px; }

.sub-table { }
.sub-header, .sub-row {
  display: grid;
  grid-template-columns: 70px 1fr 80px 110px 1fr 80px;
  align-items: center;
  padding: 14px 24px;
  gap: 8px;
}
.sub-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-light);
  border-bottom: 1px solid var(--border);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.sub-row {
  border-bottom: 1px solid var(--border);
  transition: all var(--transition);
  font-size: 14px;
}
.sub-row:last-child { border-bottom: none; }
.sub-row:hover { background: #f8faff; }
.col-id { color: var(--text-light); font-weight: 500; }
.col-problem a { font-weight: 600; }
.lang-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 5px;
  background: #f1f5f9;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.col-status { display: flex; align-items: center; gap: 6px; }
.status-indicator {
  width: 8px; height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.indicator-ac { background: var(--success); }
.indicator-wa, .indicator-error { background: var(--danger); }
.indicator-pending { background: var(--primary); animation: pulse 1.5s infinite; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: .4; } }
.indicator-tle, .indicator-mle { background: #f59e0b; }
.indicator-ce { background: #8b5cf6; }
.col-time { font-size: 13px; color: var(--text-light); }
.col-action { text-align: right; }
.col-action .btn-ghost { font-size: 13px; }
.col-action svg { width: 13px; height: 13px; }

@media (max-width: 768px) {
  .sub-header, .sub-row { grid-template-columns: 60px 1fr 80px 70px; }
  .col-lang, .col-time { display: none; }
}
@media (max-width: 640px) {
  .sub-header, .sub-row { padding: 12px 16px; grid-template-columns: 50px 1fr 60px; }
  .col-lang, .col-time { display: none; }
  .col-status { font-size: 12px; }
}
</style>
