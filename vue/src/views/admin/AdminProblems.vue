<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>⚙️ 题目管理</h1>
      <div style="display:flex;gap:8px">
        <router-link to="/admin/tags" class="btn btn-outline btn-sm">🏷 标签管理</router-link>
        <router-link to="/admin/problems/new" class="btn btn-primary">+ 新建题目</router-link>
      </div>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <template v-else-if="problems.length > 0">
      <div class="card">
        <div class="admin-table">
          <div class="admin-header">
            <span class="col-id">ID</span>
            <span class="col-title">标题</span>
            <span class="col-stat">提交/通过</span>
            <span class="col-action">操作</span>
          </div>
          <div class="admin-row" v-for="p in problems" :key="p.ID || p.id">
            <span class="col-id">{{ p.ID || p.id }}</span>
            <span class="col-title">{{ p.Title || p.title }}</span>
            <span class="col-stat">{{ p.SubmitCount || p.submit_count }} / {{ p.AcceptedCount || p.accepted_count }}</span>
            <span class="col-action">
              <router-link :to="`/admin/problems/${p.ID || p.id}/edit`" class="btn btn-outline btn-sm">编辑</router-link>
              <button class="btn btn-danger btn-sm" @click="confirmDelete(p)">删除</button>
            </span>
          </div>
        </div>
      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button :disabled="page <= 1" @click="goPage(page - 1)">上一页</button>
        <button v-for="p in visiblePages" :key="p" :class="{ active: p === page }" @click="goPage(p)">{{ p }}</button>
        <button :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</button>
      </div>
    </template>

    <div class="empty" v-else>
      <span>📭</span>
      <p>暂无题目</p>
    </div>

    <!-- 删除确认弹窗 -->
    <div class="modal-overlay" v-if="deleteTarget" @click.self="deleteTarget=null">
      <div class="modal-card">
        <h3>确认删除</h3>
        <p>确定要删除题目「{{ deleteTarget.Title || deleteTarget.title }}」吗？此操作不可撤销。</p>
        <div class="modal-actions">
          <button class="btn btn-outline" @click="deleteTarget=null">取消</button>
          <button class="btn btn-danger" @click="handleDelete" :disabled="deleting">
            <span v-if="deleting" class="spinner"></span>
            <span v-else>确认删除</span>
          </button>
        </div>
      </div>
    </div>
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
.admin-header, .admin-row {
  display: grid;
  grid-template-columns: 60px 1fr 120px 160px;
  align-items: center;
  padding: 12px 20px;
  gap: 12px;
}
.admin-header { font-size: 12px; font-weight: 600; color: var(--text-secondary); border-bottom: 1px solid var(--border); }
.admin-row { border-bottom: 1px solid var(--border); }
.admin-row:hover { background: #f8faff; }
.col-action { display: flex; gap: 6px; }

.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,.4);
  display: flex; align-items: center; justify-content: center;
  z-index: 200;
}
.modal-card {
  background: var(--bg-card);
  border-radius: var(--radius);
  padding: 28px;
  max-width: 420px;
  width: 90%;
  box-shadow: var(--shadow-lg);
}
.modal-card h3 { margin-bottom: 12px; }
.modal-card p { color: var(--text-secondary); font-size: 14px; margin-bottom: 20px; }
.modal-actions { display: flex; gap: 8px; justify-content: flex-end; }

.empty { text-align: center; padding: 80px; color: var(--text-light); }
.empty span { font-size: 48px; display: block; margin-bottom: 8px; }
</style>
