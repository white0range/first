<template>
  <div class="home-page">
    <!-- Hero 区域 -->
    <div class="hero">
      <div class="hero-bg-shapes">
        <div class="shape shape-1"></div>
        <div class="shape shape-2"></div>
        <div class="shape shape-3"></div>
      </div>
      <div class="hero-content">
        <div class="hero-badge">⚡ 在线评测平台</div>
        <h1>用代码<span class="hero-highlight">书写未来</span></h1>
        <p class="hero-desc">刷题、竞赛、成长 — 在 GoJo 遇见更好的自己</p>
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
        <div class="hero-cta" v-if="!store.isLoggedIn">
          <router-link to="/register" class="btn btn-hero">🚀 立即开始</router-link>
          <router-link to="/login" class="btn btn-hero-outline">我已有账号</router-link>
        </div>
      </div>
    </div>

    <!-- 工具栏：标签筛选 + 搜索 -->
    <div class="toolbar">
      <div class="toolbar-left">
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
        <span class="result-count" v-if="!loading">共 {{ total }} 题</span>
      </div>
    </div>

    <!-- 题目列表 -->
    <div class="problems-section">
      <div class="loading-center" v-if="loading">
        <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
      </div>

      <template v-else-if="problems.length > 0">
        <div class="problem-list">
          <router-link
            v-for="(p, idx) in problems"
            :key="p.ID || p.id"
            :to="`/problems/${p.ID || p.id}`"
            class="problem-card"
          >
            <div class="problem-card-left">
              <div class="problem-idx" :class="{ 'problem-ac': p.IsAC || p.is_ac }">
                <template v-if="p.IsAC || p.is_ac">
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
                </template>
                <template v-else>
                  <span>{{ (page - 1) * limit + idx + 1 }}</span>
                </template>
              </div>
            </div>
            <div class="problem-card-body">
              <div class="problem-card-title">
                <span class="problem-name">{{ p.Title || p.title }}</span>
              </div>
              <div class="problem-card-meta">
                <div class="problem-tags" v-if="(p.Tags || p.tags || []).length">
                  <span class="mini-tag" v-for="tag in (p.Tags || p.tags || [])" :key="tag.ID || tag.id">
                    {{ tag.Name || tag.name }}
                  </span>
                </div>
                <div class="problem-stats">
                  <span class="stat-pass" :class="{ 'high-rate': getRate(p) >= 60, 'mid-rate': getRate(p) >= 30 && getRate(p) < 60 }">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
                    {{ getRate(p) }}%
                  </span>
                  <span class="stat-submit">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 11 12 14 22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
                    {{ p.SubmitCount ?? p.submit_count ?? 0 }} 次
                  </span>
                </div>
              </div>
            </div>
            <div class="problem-card-right">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
            </div>
          </router-link>
        </div>

        <!-- 分页 -->
        <div class="pagination" v-if="totalPages > 1">
          <button :disabled="page <= 1" @click="goPage(page - 1)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
            上一页
          </button>
          <button
            v-for="p in visiblePages"
            :key="p"
            :class="{ active: p === page }"
            @click="goPage(p)"
          >{{ p }}</button>
          <button :disabled="page >= totalPages" @click="goPage(page + 1)">
            下一页
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
          </button>
        </div>
      </template>

      <div class="empty-state" v-else>
        <span class="empty-icon">📭</span>
        <p class="empty-text">暂无题目</p>
        <p class="empty-hint">管理员快去发布第一道题目吧 🚀</p>
      </div>
    </div>

    <!-- 管理员浮动按钮 -->
    <transition name="fab-fade">
      <router-link
        v-if="store.isAdmin"
        to="/admin/problems"
        class="fab-admin"
        title="管理后台"
      >
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
        <span class="fab-label">管理</span>
      </router-link>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getProblems, getTags } from '../api/index.js'
import { store } from '../store/index.js'

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
/* ===== Hero ===== */
.hero {
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #1e1b4b 0%, #312e81 25%, #3730a3 50%, #4338ca 75%, #4f46e5 100%);
  border-radius: var(--radius);
  padding: 56px 40px;
  text-align: center;
  color: #fff;
  margin-bottom: 28px;
  isolation: isolate;
}
.hero-bg-shapes { position: absolute; inset: 0; overflow: hidden; pointer-events: none; z-index: 0; }
.shape {
  position: absolute;
  border-radius: 50%;
  opacity: .08;
}
.shape-1 {
  width: 400px; height: 400px;
  background: radial-gradient(circle, #818cf8, transparent);
  top: -100px; right: -100px;
  animation: float 8s ease-in-out infinite;
}
.shape-2 {
  width: 300px; height: 300px;
  background: radial-gradient(circle, #22d3ee, transparent);
  bottom: -80px; left: -80px;
  animation: float 10s ease-in-out infinite reverse;
}
.shape-3 {
  width: 200px; height: 200px;
  background: radial-gradient(circle, #a78bfa, transparent);
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  animation: float 12s ease-in-out infinite;
}
@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(30px, -20px) scale(1.05); }
  66% { transform: translate(-20px, 10px) scale(.95); }
}

.hero-content { position: relative; z-index: 1; }
.hero-badge {
  display: inline-block;
  padding: 6px 16px;
  border-radius: 100px;
  background: rgba(255,255,255,.12);
  backdrop-filter: blur(4px);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
  margin-bottom: 16px;
  border: 1px solid rgba(255,255,255,.15);
}
.hero h1 {
  font-size: 40px;
  font-weight: 800;
  margin-bottom: 8px;
  letter-spacing: -1px;
  line-height: 1.2;
}
.hero-highlight {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.hero-desc { font-size: 17px; opacity: .85; font-weight: 400; margin-bottom: 24px; }
.hero-stats {
  display: flex;
  justify-content: center;
  gap: 48px;
  margin-bottom: 28px;
  flex-wrap: wrap;
}
.stat-item { display: flex; flex-direction: column; align-items: center; }
.stat-value { font-size: 28px; font-weight: 800; line-height: 1; }
.stat-label { font-size: 12px; opacity: .75; margin-top: 4px; font-weight: 500; }

.hero-cta { display: flex; gap: 12px; justify-content: center; flex-wrap: wrap; }
.btn-hero {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 14px 32px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 700;
  background: #fff;
  color: #4338ca !important;
  box-shadow: 0 4px 16px rgba(0,0,0,.2);
  transition: all var(--transition);
}
.btn-hero:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,.3); }
.btn-hero-outline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 14px 32px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  background: transparent;
  color: #fff !important;
  border: 2px solid rgba(255,255,255,.25);
  transition: all var(--transition);
}
.btn-hero-outline:hover { border-color: #fff; background: rgba(255,255,255,.08); }

/* ===== Toolbar ===== */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}
.toolbar-left { display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
.filter-bar { display: flex; gap: 6px; flex-wrap: wrap; }
.tag-chip {
  padding: 6px 14px;
  border-radius: 100px;
  border: 1.5px solid var(--border);
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition);
  font-family: inherit;
}
.tag-chip:hover { border-color: var(--primary); color: var(--primary); }
.tag-chip.active {
  background: var(--primary);
  border-color: var(--primary);
  color: #fff;
  box-shadow: 0 2px 8px rgba(99,102,241,.25);
}
.result-count { font-size: 13px; color: var(--text-light); font-weight: 500; }

/* ===== Problem List ===== */
.problem-list { display: flex; flex-direction: column; gap: 8px; }
.problem-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition: all var(--transition);
  color: var(--text) !important;
}
.problem-card:hover {
  border-color: var(--primary-light);
  box-shadow: 0 2px 12px rgba(99,102,241,.10);
  transform: translateY(-1px);
}
.problem-card-left { flex-shrink: 0; }
.problem-idx {
  width: 36px; height: 36px;
  display: flex; align-items: center; justify-content: center;
  border-radius: var(--radius-xs);
  font-size: 14px; font-weight: 700;
  background: #f1f5f9;
  color: var(--text-secondary);
}
.problem-idx.problem-ac {
  background: #d1fae5;
  color: #065f46;
}
.problem-card-body { flex: 1; min-width: 0; }
.problem-card-title { margin-bottom: 6px; }
.problem-name { font-size: 15px; font-weight: 600; }
.problem-card-meta { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.problem-tags { display: flex; gap: 4px; flex-wrap: wrap; }
.problem-stats { display: flex; align-items: center; gap: 12px; margin-left: auto; flex-shrink: 0; }
.stat-pass, .stat-submit {
  display: flex; align-items: center; gap: 4px;
  font-size: 12px; font-weight: 600;
  color: var(--text-light);
}
.stat-pass svg, .stat-submit svg { width: 13px; height: 13px; }
.stat-pass.high-rate { color: var(--success); }
.stat-pass.mid-rate { color: #f59e0b; }
.problem-card-right {
  flex-shrink: 0;
  color: var(--border);
  transition: all var(--transition);
}
.problem-card:hover .problem-card-right { color: var(--primary); transform: translateX(4px); }

@media (max-width: 640px) {
  .problem-stats { display: none; }
  .hero { padding: 40px 20px; }
  .hero h1 { font-size: 28px; }
  .hero-stats { gap: 28px; }
}

/* ===== FAB Admin ===== */
.fab-admin {
  position: fixed;
  bottom: 32px;
  right: 32px;
  z-index: 50;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 20px;
  border-radius: 100px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: #fff !important;
  font-weight: 700;
  font-size: 14px;
  box-shadow: 0 4px 16px rgba(245,158,11,.35);
  transition: all var(--transition);
  text-decoration: none;
}
.fab-admin:hover {
  transform: translateY(-2px) scale(1.03);
  box-shadow: 0 8px 28px rgba(245,158,11,.45);
}
.fab-label { display: inline; }
.fab-fade-enter-active, .fab-fade-leave-active { transition: all .3s ease; }
.fab-fade-enter-from, .fab-fade-leave-to { opacity: 0; transform: translateY(20px) scale(.8); }

@media (max-width: 640px) {
  .fab-admin { bottom: 20px; right: 20px; padding: 12px 16px; }
  .fab-label { display: none; }
}

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
