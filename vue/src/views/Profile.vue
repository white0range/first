<template>
  <div class="profile-page">
    <div class="page-header">
      <h1>👤 个人中心</h1>
      <button class="btn btn-outline btn-sm" @click="handleLogout">🚪 退出登录</button>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <template v-else-if="profile">
      <div class="profile-grid">
        <!-- 个人信息卡片 -->
        <div class="card profile-card">
          <div class="profile-avatar">{{ (profile.username || '?')[0]?.toUpperCase() }}</div>
          <h2>{{ profile.username }}</h2>
          <div class="profile-role">
            <span v-if="profile.role === 1" class="badge badge-warning">👑 管理员</span>
            <span v-else class="badge badge-info">普通用户</span>
          </div>
          <div class="profile-stats">
            <div class="p-stat">
              <span class="p-stat-value">{{ profile.solved_count ?? 0 }}</span>
              <span class="p-stat-label">已解决</span>
            </div>
            <div class="p-stat">
              <span class="p-stat-value">{{ solvedList.length }}</span>
              <span class="p-stat-label">不同题目</span>
            </div>
            <div class="p-stat">
              <span class="p-stat-value">{{ profile.role === 1 ? '管理员' : '用户' }}</span>
              <span class="p-stat-label">身份</span>
            </div>
          </div>
          <!-- 已解决题目列表 -->
          <div class="solved-section" v-if="solvedList.length > 0">
            <h4>✅ 已解决的题目 ({{ solvedList.length }})</h4>
            <div class="solved-tags">
              <router-link
                v-for="pid in solvedList" :key="pid"
                :to="`/problems/${pid}`"
                class="solved-tag"
              >#{{ pid }}</router-link>
            </div>
          </div>
        </div>

        <!-- 快捷入口 -->
        <div class="card quick-links">
          <h3>快捷入口</h3>
          <div class="link-grid">
            <router-link to="/my-submissions" class="quick-link">📋 我的提交记录</router-link>
            <router-link to="/leaderboard" class="quick-link">🏆 全服排行榜</router-link>
            <router-link to="/" class="quick-link">📚 继续刷题</router-link>
          </div>
        </div>
      </div>
    </template>

    <div class="empty" v-else>
      <span>😵</span>
      <p>获取用户信息失败，请先登录</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getProfile } from '../api/index.js'
import { store } from '../store/index.js'

const router = useRouter()
const profile = ref(null)
const solvedList = ref([])
const loading = ref(true)

function handleLogout() {
  store.logout()
  router.push('/')
}

onMounted(async () => {
  try {
    const res = await getProfile()
    const data = res.data.data || res.data
    profile.value = data.user_info || {}
    solvedList.value = data.solved_list || []
  } catch (e) {
    console.error('获取个人信息失败', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.profile-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}
@media (max-width: 768px) { .profile-grid { grid-template-columns: 1fr; } }

.profile-card { text-align: center; }
.profile-avatar {
  width: 80px; height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--accent));
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-size: 36px; font-weight: 800;
  margin: 0 auto 14px;
  box-shadow: 0 4px 14px rgba(99,102,241,.35);
}
.profile-card h2 { font-size: 22px; margin-bottom: 6px; }
.profile-role { margin-bottom: 22px; }
.profile-stats { display: flex; justify-content: center; gap: 36px; }
.p-stat { display: flex; flex-direction: column; }
.p-stat-value { font-size: 26px; font-weight: 800; color: var(--primary); }
.p-stat-label { font-size: 12px; color: var(--text-secondary); font-weight: 500; }

.solved-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--border);
}
.solved-section h4 { font-size: 14px; color: var(--text-secondary); margin-bottom: 10px; }
.solved-tags { display: flex; flex-wrap: wrap; gap: 6px; justify-content: center; }
.solved-tag {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 100px;
  background: #d1fae5;
  color: #065f46 !important;
  font-size: 13px;
  font-weight: 600;
  transition: var(--transition);
}
.solved-tag:hover { background: #a7f3d0; transform: scale(1.05); }

.quick-links h3 { margin-bottom: 16px; font-size: 16px; }
.link-grid { display: flex; flex-direction: column; gap: 8px; }
.quick-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 8px;
  background: #f8fafc;
  transition: var(--transition);
  color: var(--text) !important;
  font-weight: 500;
}
.quick-link:hover { background: #eef2ff; }
.empty { text-align: center; padding: 80px; color: var(--text-light); }
.empty span { font-size: 48px; display: block; margin-bottom: 8px; }
</style>
