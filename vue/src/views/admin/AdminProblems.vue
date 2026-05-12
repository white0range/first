<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>题目管理</h1>
      <div class="page-actions">
        <router-link to="/admin/tags" class="btn btn-ghost btn-sm">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
          标签管理
        </router-link>
        <router-link to="/admin/problems/new" class="btn btn-primary">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新建题目
        </router-link>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-box">
        <span class="stat-box-value">{{ total }}</span>
        <span class="stat-box-label">题目总数</span>
      </div>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else-if="problems.length > 0">
      <div class="card" style="padding:0;overflow:hidden">
        <div class="admin-table">
          <div class="admin-header">
            <span class="col-id">ID</span>
            <span class="col-title">标题</span>
            <span class="col-tags">标签</span>
            <span class="col-stat">提交/通过</span>
            <span class="col-action">操作</span>
          </div>
          <div class="admin-row" v-for="p in problems" :key="p.ID || p.id">
            <span class="col-id">{{ p.ID || p.id }}</span>
            <span class="col-title">
              <span class="problem-title-text">{{ p.Title || p.title }}</span>
            </span>
            <span class="col-tags">
              <span class="mini-tag" v-for="tag in (p.Tags || p.tags || []).slice(0, 2)" :key="tag.ID || tag.id">{{ tag.Name || tag.name }}</span>
              <span v-if="(p.Tags || p.tags || []).length > 2" class="mini-tag more">+{{ (p.Tags || p.tags || []).length - 2 }}</span>
            </span>
            <span class="col-stat">{{ p.SubmitCount || p.submit_count || 0 }} / {{ p.AcceptedCount || p.accepted_count || 0 }}</span>
            <span class="col-action">
              <router-link :to="`/admin/problems/${p.ID || p.id}/edit`" class="btn btn-ghost btn-sm">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑
              </router-link>
              <button class="btn btn-danger btn-sm" @click="confirmDelete(p)">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                删除
              </button>
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
      <p class="empty-text">暂无题目</p>
      <p class="empty-hint">点击「新建题目」发布你的第一道题目</p>
    </div>

    <!-- 删除确认弹窗 -->
    <transition name="modal">
      <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget=null">
        <div class="modal-card">
          <div class="modal-icon">🗑️</div>
          <h3>确认删除</h3>
          <p>确定要删除题目「<strong>{{ deleteTarget.Title || deleteTarget.title }}</strong>」吗？此操作不可撤销，相关的提交记录也会被删除。</p>
          <div class="modal-actions">
            <button class="btn btn-ghost btn-lg" @click="deleteTarget=null">取消</button>
            <button class="btn btn-danger btn-lg" @click="handleDelete" :disabled="deleting">
              <span v-if="deleting" class="spinner"></span>
              <span v-else>确认删除</span>
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getProblems, adminDeleteProblem } from '../../api/index.js'

const problems = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(20)
const loading = ref(true)
const deleteTarget = ref(null)
const deleting = ref(false)

const totalPages = computed(() => Math.ceil(total.value / limit.value))
const visiblePages = computed(() => {
  const pages = []
  for (let i = Math.max(1, page.value - 2); i <= Math.min(totalPages.value, page.value + 2); i++) pages.push(i)
  return pages
})

async function fetchProblems() {
  loading.value = true
  try {
    const res = await getProblems({ page: page.value, limit: limit.value })
    problems.value = res.data.items || res.data.data?.items || []
    total.value = res.data.total || res.data.data?.total || 0
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

function goPage(p) { page.value = p; fetchProblems() }
function confirmDelete(p) { deleteTarget.value = p }

async function handleDelete() {
  deleting.value = true
  try {
    const id = deleteTarget.value.ID || deleteTarget.value.id
    await adminDeleteProblem(id)
    deleteTarget.value = null
    fetchProblems()
  } catch (e) { alert('删除失败：' + (e.response?.data?.error || e.message)) }
  finally { deleting.value = false }
}

onMounted(fetchProblems)
</script>

<style scoped>
.page-actions { display: flex; gap: 8px; }

.stats-row { margin-bottom: 20px; }
.stat-box {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 32px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
}
.stat-box-value { font-size: 28px; font-weight: 800; color: var(--primary); line-height: 1; }
.stat-box-label { font-size: 12px; color: var(--text-light); margin-top: 4px; }

.admin-table { }
.admin-header, .admin-row {
  display: grid;
  grid-template-columns: 60px 1fr 120px 120px 140px;
  align-items: center;
  padding: 14px 24px;
  gap: 12px;
}
.admin-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-light);
  border-bottom: 1px solid var(--border);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.admin-row {
  border-bottom: 1px solid var(--border);
  transition: all var(--transition);
  font-size: 14px;
}
.admin-row:last-child { border-bottom: none; }
.admin-row:hover { background: #f8faff; }
.col-id { color: var(--text-light); font-weight: 500; }
.problem-title-text { font-weight: 600; }
.col-tags { display: flex; gap: 4px; flex-wrap: wrap; }
.col-tags .mini-tag.more { background: #f1f5f9; color: var(--text-light); }
.col-stat { font-size: 13px; color: var(--text-secondary); }
.col-action { display: flex; gap: 4px; }
.col-action .btn-ghost { font-size: 12px; padding: 5px 10px; }
.col-action .btn-danger { font-size: 12px; padding: 5px 10px; }
.col-action svg { width: 13px; height: 13px; }

/* Modal */
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,.45);
  display: flex; align-items: center; justify-content: center;
  z-index: 200;
  backdrop-filter: blur(4px);
}
.modal-card {
  background: var(--bg-card);
  border-radius: var(--radius);
  padding: 32px;
  max-width: 420px;
  width: 90%;
  box-shadow: var(--shadow-xl);
  text-align: center;
  animation: modalIn .25s ease;
}
@keyframes modalIn {
  from { opacity: 0; transform: scale(.92) translateY(12px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
.modal-icon { font-size: 40px; margin-bottom: 12px; }
.modal-card h3 { margin-bottom: 8px; font-size: 18px; }
.modal-card p { color: var(--text-secondary); font-size: 14px; margin-bottom: 24px; line-height: 1.6; }
.modal-actions { display: flex; gap: 10px; justify-content: center; }
.modal-actions .btn { min-width: 100px; }

.modal-enter-active, .modal-leave-active { transition: all .2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .modal-card, .modal-leave-to .modal-card { transform: scale(.92) translateY(12px); }

@media (max-width: 768px) {
  .admin-header, .admin-row { grid-template-columns: 50px 1fr 100px 100px; }
  .col-tags { display: none; }
}
@media (max-width: 640px) {
  .admin-header, .admin-row { grid-template-columns: 40px 1fr 80px; }
  .col-tags, .col-stat { display: none; }
  .col-action { flex-direction: column; }
}
</style>
