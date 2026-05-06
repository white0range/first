<template>
  <div class="admin-edit-page">
    <div class="page-header">
      <h1>{{ isNew ? '📝 新建题目' : '✏️ 编辑题目' }}</h1>
      <router-link to="/admin/problems" class="btn btn-outline btn-sm">← 返回列表</router-link>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <form v-else class="edit-form" @submit.prevent="handleSave">
      <!-- 基本信息 -->
      <div class="card">
        <h3>基本信息</h3>
        <div class="form-row">
          <div class="form-group" style="flex:1">
            <label>题目标题 <span class="required">*</span></label>
            <input v-model="form.title" class="input" placeholder="例如：A+B Problem" required />
          </div>
        </div>
        <div class="form-row two-col">
          <div class="form-group">
            <label>时间限制 (ms)</label>
            <input v-model.number="form.time_limit" class="input" type="number" placeholder="1000" />
          </div>
          <div class="form-group">
            <label>内存限制 (MB)</label>
            <input v-model.number="form.memory_limit" class="input" type="number" placeholder="256" />
          </div>
        </div>
        <div class="form-group">
          <label>题目描述 <span class="required">*</span></label>
          <textarea v-model="form.description" class="input" rows="10" placeholder="支持 Markdown 格式..." required></textarea>
        </div>
      </div>

      <!-- 标签 -->
      <div class="card">
        <h3>🏷 标签</h3>
        <div class="tag-select">
          <label v-for="tag in allTags" :key="tag.ID || tag.id" class="tag-checkbox">
            <input type="checkbox" :value="tag.ID || tag.id" v-model="form.tag_ids" />
            <span>{{ tag.Name || tag.name }}</span>
          </label>
        </div>
        <p class="hint" v-if="allTags.length === 0">暂无标签，请先去标签管理创建</p>
      </div>

      <!-- 测试用例 -->
      <div class="card">
        <div class="testcase-header">
          <h3>🧪 测试用例</h3>
          <button type="button" class="btn btn-outline btn-sm" @click="addCase">+ 添加</button>
        </div>

        <div v-if="form.test_cases.length === 0" class="no-cases">
          还没有测试用例，点击"添加"按钮创建
        </div>

        <div v-for="(tc, idx) in form.test_cases" :key="idx" class="testcase-item">
          <div class="tc-header">
            <span>用例 #{{ idx + 1 }}</span>
            <button type="button" class="btn btn-danger btn-sm" @click="removeCase(idx)">删除</button>
          </div>
          <div class="tc-body">
            <div class="form-group">
              <label>输入</label>
              <textarea v-model="tc.input" class="input mono" rows="3" placeholder="标准输入..." required></textarea>
            </div>
            <div class="form-group">
              <label>期望输出</label>
              <textarea v-model="tc.expected_output" class="input mono" rows="3" placeholder="期望输出..." required></textarea>
            </div>
          </div>
        </div>
      </div>

      <!-- 已存在的测试用例（编辑模式） -->
      <div class="card" v-if="!isNew && existingCases.length > 0">
        <h3>📦 已有测试用例 ({{ existingCases.length }})</h3>
        <div class="existing-case" v-for="tc in existingCases" :key="tc.ID || tc.id">
          <div class="ec-header">
            <span>#{{ tc.ID || tc.id }}</span>
            <button type="button" class="btn btn-danger btn-sm" @click="deleteExistingCase(tc)">删除</button>
          </div>
          <div class="ec-body">
            <div><strong>输入：</strong><code>{{ tc.Input || tc.input }}</code></div>
            <div><strong>输出：</strong><code>{{ tc.ExpectedOutput || tc.expected_output }}</code></div>
          </div>
        </div>
      </div>

      <div class="error-msg" v-if="error">{{ error }}</div>

      <div class="form-actions">
        <button type="submit" class="btn btn-primary" :disabled="saving">
          <span v-if="saving" class="spinner"></span>
          <span v-else>{{ isNew ? '发布题目' : '保存修改' }}</span>
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTags, getProblemDetail, adminCreateProblem, adminUpdateProblem, adminUpdateProblemTags, adminAddTestCase, adminGetTestCases, adminDeleteTestCase } from '../../api/index.js'

const route = useRoute()
const router = useRouter()
const isNew = computed(() => route.name === 'AdminProblemNew')
const problemId = computed(() => parseInt(route.params.id))

const form = reactive({
  title: '',
  description: '',
  time_limit: 1000,
  memory_limit: 256,
  tag_ids: [],
  test_cases: [],
})

const allTags = ref([])
const existingCases = ref([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')

function addCase() {
  form.test_cases.push({ input: '', expected_output: '' })
}
function removeCase(idx) { form.test_cases.splice(idx, 1) }

async function fetchTags() {
  try {
    const res = await getTags()
    allTags.value = res.data.data || res.data || []
  } catch (e) { /* ignore */ }
}

async function fetchProblem() {
  if (isNew.value) return
  loading.value = true
  try {
    const [problemRes, casesRes] = await Promise.all([
      getProblemDetail(problemId.value),
      adminGetTestCases(problemId.value).catch(() => ({ data: { data: [] } })),
    ])
    const p = problemRes.data.data || problemRes.data
    form.title = p.Title || p.title || ''
    form.description = p.Description || p.description || ''
    form.time_limit = p.TimeLimit || p.time_limit || 1000
    form.memory_limit = p.MemoryLimit || p.memory_limit || 256
    form.tag_ids = (p.Tags || p.tags || []).map(t => t.ID || t.id)
    existingCases.value = casesRes.data.data || casesRes.data || []
  } catch (e) {
    error.value = '获取题目信息失败'
    console.error(e)
  } finally {
    loading.value = false
  }
}

async function deleteExistingCase(tc) {
  if (!confirm('确定删除该测试用例？')) return
  try {
    await adminDeleteTestCase(tc.ID || tc.id)
    existingCases.value = existingCases.value.filter(c => (c.ID || c.id) !== (tc.ID || tc.id))
  } catch (e) { alert('删除失败') }
}

async function handleSave() {
  error.value = ''
  saving.value = true
  try {
    const payload = {
      title: form.title,
      description: form.description,
      time_limit: form.time_limit,
      memory_limit: form.memory_limit,
      tag_ids: form.tag_ids,
      test_cases: form.test_cases,
    }
    if (isNew.value) {
      await adminCreateProblem(payload)
    } else {
      await adminUpdateProblem(problemId.value, payload)
      // 单独更新标签关联
      await adminUpdateProblemTags(problemId.value, { tag_ids: form.tag_ids }).catch(() => {})
      // 如果有新增测试用例，逐个添加
      for (const tc of form.test_cases) {
        if (tc.input || tc.expected_output) {
          await adminAddTestCase(problemId.value, tc).catch(() => {})
        }
      }
    }
    router.push('/admin/problems')
  } catch (e) {
    error.value = e.response?.data?.error || '保存失败'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchTags()
  fetchProblem()
})
</script>

<style scoped>
.edit-form { display: flex; flex-direction: column; gap: 20px; }
.edit-form h3 { font-size: 16px; margin-bottom: 16px; }
.form-row { display: flex; gap: 16px; }
.two-col > * { flex: 1; }
.form-group { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
.form-group label { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.required { color: var(--danger); }
.mono { font-family: 'Fira Code', 'Consolas', monospace; font-size: 13px; }

.tag-select { display: flex; flex-wrap: wrap; gap: 8px; }
.tag-checkbox { display: flex; align-items: center; gap: 6px; font-size: 14px; cursor: pointer; padding: 4px 12px; border-radius: 6px; border: 1px solid var(--border); }
.tag-checkbox:has(input:checked) { background: #eef2ff; border-color: var(--primary); }
.hint { color: var(--text-light); font-size: 13px; }

.testcase-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.testcase-header h3 { margin-bottom: 0; }
.testcase-item { border: 1px solid var(--border); border-radius: 8px; padding: 16px; margin-bottom: 12px; }
.tc-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; font-weight: 600; }
.tc-body { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.no-cases { color: var(--text-light); font-size: 14px; text-align: center; padding: 24px; }

.existing-case { border: 1px solid var(--border); border-radius: 8px; padding: 12px 16px; margin-bottom: 8px; }
.ec-header { display: flex; justify-content: space-between; margin-bottom: 8px; font-weight: 600; }
.ec-body { font-size: 13px; display: flex; flex-direction: column; gap: 4px; }
.ec-body code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; font-size: 12px; }

.form-actions { display: flex; justify-content: flex-end; padding-top: 8px; }
.error-msg { color: var(--danger); font-size: 13px; background: #fef2f2; padding: 12px; border-radius: 8px; }
</style>
