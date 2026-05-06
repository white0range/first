<template>
  <div class="admin-tags-page">
    <div class="page-header">
      <h1>🏷 标签管理</h1>
      <router-link to="/admin/problems" class="btn btn-outline btn-sm">← 题目管理</router-link>
    </div>

    <!-- 新建标签 -->
    <div class="card" style="margin-bottom:20px">
      <h3>创建新标签</h3>
      <form @submit.prevent="handleCreate" class="tag-form">
        <input v-model="newTagName" class="input" placeholder="输入标签名称，如：动态规划" required />
        <button type="submit" class="btn btn-primary" :disabled="creating">
          <span v-if="creating" class="spinner"></span>
          <span v-else>创建</span>
        </button>
      </form>
      <div class="error-msg" v-if="createError">{{ createError }}</div>
    </div>

    <!-- 标签列表 -->
    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <template v-else-if="tags.length > 0">
      <div class="card">
        <div class="tags-grid">
          <div class="tag-row" v-for="tag in tags" :key="tag.ID || tag.id">
            <span class="tag-name">{{ tag.Name || tag.name }}</span>
            <button class="btn btn-danger btn-sm" @click="handleDelete(tag)">删除</button>
          </div>
        </div>
      </div>
    </template>

    <div class="empty" v-else>
      <span>🏷️</span>
      <p>暂无标签，创建一个吧</p>
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
.tag-form { display: flex; gap: 12px; margin-top: 12px; }
.tag-form .input { flex: 1; }
.tags-grid { }
.tag-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}
.tag-row:last-child { border-bottom: none; }
.tag-name { font-weight: 600; font-size: 15px; }
.error-msg { color: var(--danger); font-size: 13px; margin-top: 8px; }
.empty { text-align: center; padding: 80px; color: var(--text-light); }
.empty span { font-size: 48px; display: block; margin-bottom: 8px; }
</style>
