<template>
  <div class="home-page">
    <!-- Hero 区域 -->
    <div class="hero">
      <div class="hero-content">
        <h1>⚡ GoJo 在线评测平台</h1>
        <p>刷题、竞赛、成长 — 用代码书写你的未来</p>
        <div class="hero-stats">
          <div class="stat-item">
            <span class="stat-value">{{ total }}</span>
            <span class="stat-label">精选题目</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">4</span>
            <span class="stat-label">编程语言</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">Docker</span>
            <span class="stat-label">安全沙箱</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 标签筛选 -->
    <div class="filter-bar" v-if="tags.length > 0">
      <button
        v-for="tag in tags"
        :key="tag.ID || tag.id"
        class="tag-chip"
        :class="{ active: selectedTagId === (tag.ID || tag.id) }"
        @click="toggleTag(tag.ID || tag.id)"
      >
        {{ tag.Name || tag.name }}
      </button>
    </div>

    <!-- 题目列表 -->
    <div class="problems-section">
      <div class="loading-center" v-if="loading">
        <div class="spinner"></div>
      </div>

      <template v-else-if="problems.length > 0">
        <div class="problem-table">
          <div class="table-header">
            <span class="col-id">#</span>
            <span class="col-title">题目</span>
            <span class="col-tags">标签</span>
            <span class="col-stat">通过率</span>
            <span class="col-action"></span>
          </div>
          <router-link
            v-for="(p, idx) in problems"
            :key="p.ID || p.id"
            :to="`/problems/${p.ID || p.id}`"
            class="table-row"
          >
            <span class="col-id">
              <span v-if="p.IsAC || p.is_ac" class="ac-mark">✓</span>
              <span v-else class="row-num">{{ (page - 1) * limit + idx + 1 }}</span>
            </span>
            <span class="col-title">
              <span class="problem-name">{{ p.Title || p.title }}</span>
            </span>
            <span class="col-tags">
              <span class="mini-tag" v-for="tag in (p.Tags || p.tags || [])" :key="tag.ID || tag.id">
                {{ tag.Name || tag.name }}
              </span>
            </span>
            <span class="col-stat">
              <span class="pass-rate">
                {{ getRate(p) }}%
              </span>
            </span>
            <span class="col-action">
              <span class="btn btn-outline btn-sm">开始挑战 →</span>
            </span>
          </router-link>
        </div>

        <!-- 分页 -->
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
        <p>暂无题目</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getProblems, getTags } from '../api/index.js'

const problems = ref([])
const tags = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const loading = ref(true)
const selectedTagId = ref(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit.value)))
const visiblePages = computed(() => {
  const pages = []
  const max = totalPages.value
  const start = Math.max(1, page.value - 2)
  const end = Math.min(max, page.value + 2)
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
})

function getRate(p) {
  const sub = p.SubmitCount ?? p.submit_count ?? 0
  const acc = p.AcceptedCount ?? p.accepted_count ?? 0
  if (sub === 0) return 0
  return Math.round((acc / sub) * 100)
}

async function fetchProblems() {
  loading.value = true
  try {
    const params = { page: page.value, limit: limit.value }
    if (selectedTagId.value) params.tag_id = selectedTagId.value
    const res = await getProblems(params)
    const data = res.data
    problems.value = data.items || []
    total.value = data.total || 0
  } catch (e) {
    console.error('获取题目列表失败', e)
  } finally {
    loading.value = false
  }
}

async function fetchTags() {
  try {
    const res = await getTags()
    tags.value = res.data.data || res.data || []
  } catch (e) { /* ignore */ }
}

function toggleTag(tagId) {
  if (selectedTagId.value === tagId) {
    selectedTagId.value = null
  } else {
    selectedTagId.value = tagId
  }
  page.value = 1
  fetchProblems()
}

function goPage(p) {
  if (p < 1 || p > totalPages.value) return
  page.value = p
  fetchProblems()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  fetchProblems()
  fetchTags()
})
</script>

<style scoped>
/* Hero */
.hero {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: var(--radius);
  padding: 48px 32px;
  text-align: center;
  color: #fff;
  margin-bottom: 24px;
}
.hero h1 { font-size: 32px; margin-bottom: 8px; }
.hero p { font-size: 16px; opacity: .9; }
.hero-stats {
  display: flex;
  justify-content: center;
  gap: 40px;
  margin-top: 24px;
  flex-wrap: wrap;
}
.stat-item { display: flex; flex-direction: column; }
.stat-value { font-size: 22px; font-weight: 700; }
.stat-label { font-size: 12px; opacity: .8; }

/* 标签筛选 */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 20px;
}
.tag-chip {
  padding: 6px 14px;
  border-radius: 100px;
  border: 1.5px solid var(--border);
  background: var(--bg-card);
  font-size: 13px;
  cursor: pointer;
  transition: var(--transition);
  color: var(--text-secondary);
}
.tag-chip:hover { border-color: var(--primary); color: var(--primary); }
.tag-chip.active { background: var(--primary); color: #fff; border-color: var(--primary); }

/* 表格 */
.problem-table {
  background: var(--bg-card);
  border-radius: var(--radius);
  border: 1px solid var(--border);
  overflow: hidden;
}
.table-header, .table-row {
  display: grid;
  grid-template-columns: 60px 1fr 200px 100px 120px;
  align-items: center;
  padding: 14px 20px;
  gap: 12px;
}
.table-header {
  background: #f8fafc;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.table-row {
  border-top: 1px solid var(--border);
  color: var(--text) !important;
  transition: var(--transition);
  cursor: pointer;
}
.table-row:hover { background: #f8faff; }

.ac-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px; height: 24px;
  background: #d1fae5;
  color: #065f46;
  border-radius: 6px;
  font-weight: 700;
  font-size: 14px;
}
.row-num { color: var(--text-light); font-weight: 500; }
.problem-name { font-weight: 600; }
.mini-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  background: #f1f5f9;
  font-size: 11px;
  color: var(--text-secondary);
  margin-right: 4px;
}
.pass-rate { font-size: 13px; color: var(--text-secondary); }
.empty { text-align: center; padding: 60px; color: var(--text-light); font-size: 16px; }
.empty span { font-size: 48px; display: block; margin-bottom: 8px; }

@media (max-width: 768px) {
  .table-header, .table-row { grid-template-columns: 40px 1fr 80px; }
  .col-tags, .col-action { display: none; }
}
</style>
