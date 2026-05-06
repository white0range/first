<template>
  <div class="submissions-page">
    <div class="page-header">
      <h1>📋 我的提交记录</h1>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <template v-else-if="submissions.length > 0">
      <div class="card">
        <div class="sub-table">
          <div class="sub-header">
            <span class="col-id">#</span>
            <span class="col-problem">题目ID</span>
            <span class="col-lang">语言</span>
            <span class="col-status">状态</span>
            <span class="col-time">时间</span>
            <span class="col-action"></span>
          </div>
          <div class="sub-row" v-for="item in submissions" :key="item.ID || item.id">
            <span class="col-id">{{ item.ID || item.id }}</span>
            <span class="col-problem">
              <router-link :to="`/problems/${item.ProblemID || item.problem_id}`">
                #{{ item.ProblemID || item.problem_id }}
              </router-link>
            </span>
            <span class="col-lang">
              <span class="lang-badge">{{ item.Language || item.language }}</span>
            </span>
            <span class="col-status">
              <span :class="'status-' + (item.Status || item.status || 'Pending')">
                {{ item.Status || item.status || 'Pending' }}
              </span>
            </span>
            <span class="col-time">{{ formatTime(item.CreatedAt || item.created_at) }}</span>
            <span class="col-action">
              <router-link
                :to="`/submissions/${item.ID || item.id}`"
                class="btn btn-outline btn-sm"
              >
                详情
              </router-link>
            </span>
          </div>
        </div>
      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button :disabled="page <= 1" @click="goPage(page - 1)">上一页</button>
        <button
          v-for="p in visiblePages"
          :key="p"
          :class="{ active: p === page }"
          @click="goPage(p)"
        >{{ p }}</button>
        <button :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</button>
      </div>
    </template>

    <div class="empty" v-else>
      <span>📭</span>
      <p>还没有提交记录，快去刷题吧！</p>
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
.sub-header, .sub-row {
  display: grid;
  grid-template-columns: 60px 100px 80px 100px 1fr 80px;
  align-items: center;
  padding: 12px 20px;
  gap: 8px;
}
.sub-header { font-size: 12px; font-weight: 600; color: var(--text-secondary); border-bottom: 1px solid var(--border); }
.sub-row { border-bottom: 1px solid var(--border); transition: var(--transition); }
.sub-row:hover { background: #f8faff; }
.lang-badge {
  padding: 2px 8px;
  border-radius: 4px;
  background: #f1f5f9;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  text-transform: uppercase;
}
.col-time { font-size: 12px; color: var(--text-light); }
.empty { text-align: center; padding: 80px; color: var(--text-light); }
.empty span { font-size: 48px; display: block; margin-bottom: 8px; }

@media (max-width: 640px) {
  .sub-header, .sub-row { grid-template-columns: 60px 1fr 80px; }
  .col-lang, .col-time { display: none; }
}
</style>
