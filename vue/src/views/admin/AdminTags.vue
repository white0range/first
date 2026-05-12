<template>
  <div class="admin-tags-page">
    <div class="page-header">
      <h1>标签管理</h1>
      <router-link to="/admin/problems" class="btn btn-ghost btn-sm">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>
        题目管理
      </router-link>
    </div>

    <!-- 新建标签 -->
    <div class="card" style="margin-bottom:20px">
      <div class="card-title">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        创建新标签
      </div>
      <form @submit.prevent="handleCreate" class="tag-form">
        <div class="input-wrapper">
          <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
          <input v-model="newTagName" class="input" placeholder="输入标签名称，如：动态规划" required />
        </div>
        <button type="submit" class="btn btn-primary" :disabled="creating || !newTagName.trim()">
          <span v-if="creating" class="spinner"></span>
          <template v-else>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            创建
          </template>
        </button>
      </form>
      <div class="error-msg" v-if="createError">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        {{ createError }}
      </div>
    </div>

    <!-- 标签列表 -->
    <div class="loading-center" v-if="loading">
      <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else-if="tags.length > 0">
      <div class="card" style="padding:0;overflow:hidden">
        <div class="tags-grid">
          <div class="tag-row" v-for="tag in tags" :key="tag.ID || tag.id">
            <div class="tag-left">
              <span class="tag-icon">🏷️</span>
              <span class="tag-name">{{ tag.Name || tag.name }}</span>
            </div>
            <button class="btn btn-danger btn-sm" @click="handleDelete(tag)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              删除
            </button>
          </div>
        </div>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-icon">🏷️</span>
      <p class="empty-text">暂无标签</p>
      <p class="empty-hint">在上方输入框创建第一个标签吧</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getTags, adminCreateTag, adminDeleteTag } from '../../api/index.js'

const tags = ref([])
const newTagName = ref('')
const loading = ref(true)
const creating = ref(false)
const createError = ref('')

async function fetchTags() {
  loading.value = true
  try {
    const res = await getTags()
    tags.value = res.data.data || res.data || []
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function handleCreate() {
  createError.value = ''
  creating.value = true
  try {
    await adminCreateTag({ name: newTagName.value })
    newTagName.value = ''
    fetchTags()
  } catch (e) {
    createError.value = e.response?.data?.error || '创建失败，标签可能已存在'
  } finally {
    creating.value = false
  }
}

async function handleDelete(tag) {
  if (!confirm(`确定删除标签「${tag.Name || tag.name}」吗？`)) return
  try {
    await adminDeleteTag(tag.ID || tag.id)
    fetchTags()
  } catch (e) { alert('删除失败') }
}

onMounted(fetchTags)
</script>

<style scoped>
.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 16px;
}
.card-title svg { color: var(--primary); flex-shrink: 0; }

.tag-form {
  display: flex;
  gap: 12px;
  align-items: center;
}
.tag-form .input-wrapper {
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
}
.tag-form .input-wrapper .input { padding-left: 40px; }
.input-icon {
  position: absolute;
  left: 12px;
  color: var(--text-light);
  pointer-events: none;
  flex-shrink: 0;
}

.tags-grid { }
.tag-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 24px;
  border-bottom: 1px solid var(--border);
  transition: all var(--transition);
}
.tag-row:last-child { border-bottom: none; }
.tag-row:hover { background: #f8faff; }
.tag-left { display: flex; align-items: center; gap: 10px; }
.tag-icon { font-size: 16px; }
.tag-name { font-weight: 600; font-size: 15px; }
.tag-row .btn-danger {
  opacity: 0;
  transition: opacity var(--transition);
}
.tag-row:hover .btn-danger { opacity: 1; }

.error-msg {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--danger);
  font-size: 13px;
  margin-top: 12px;
}
.error-msg svg { flex-shrink: 0; }
</style>
