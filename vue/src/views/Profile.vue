<template>
  <div class="profile-page">
    <div class="page-header">
      <h1>个人中心</h1>
      <button class="btn btn-ghost btn-sm" @click="handleLogout">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
        退出登录
      </button>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else-if="profile">
      <div class="profile-grid">
        <!-- 个人信息卡片 -->
        <div class="card profile-card">
          <div class="profile-avatar" :class="{ 'admin-avatar-big': profile.role === 1 }">
            {{ (profile.username || '?')[0]?.toUpperCase() }}
            <span v-if="profile.role === 1" class="avatar-crown-big">👑</span>
          </div>
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

        <!-- 快捷入口 + 管理面板 -->
        <div class="profile-side">
          <div class="card quick-links">
            <h3>快捷入口</h3>
            <div class="link-grid">
              <router-link to="/my-submissions" class="quick-link">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
                我的提交记录
              </router-link>
              <router-link to="/leaderboard" class="quick-link">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 22V2h4v20"/></svg>
                全服排行榜
              </router-link>
              <router-link to="/" class="quick-link">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
                继续刷题
              </router-link>
            </div>
          </div>

          <!-- 👑 管理员专属面板 -->
          <div class="card admin-panel" v-if="store.isAdmin">
            <div class="admin-panel-header">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
              <h3>管理后台</h3>
            </div>
            <p class="admin-panel-desc">管理题目、测试用例和标签</p>
            <div class="admin-links">
              <router-link to="/admin/problems" class="admin-link-btn">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
                题目管理
                <svg class="link-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
              </router-link>
              <router-link to="/admin/tags" class="admin-link-btn">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
                标签管理
                <svg class="link-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
              </router-link>
              <router-link to="/admin/problems/new" class="admin-link-btn new">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                发布新题目
                <svg class="link-arrow" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-icon">😵</span>
      <p class="empty-text">获取用户信息失败</p>
      <p class="empty-hint">请先登录后再查看个人中心</p>
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

/* Profile Card */
.profile-card {
  text-align: center;
  position: relative;
}
.profile-avatar {
  position: relative;
  width: 80px; height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--accent));
  color: #fff;
  display: inline-flex; align-items: center; justify-content: center;
  font-size: 36px; font-weight: 800;
  margin-bottom: 14px;
  box-shadow: 0 4px 14px rgba(99,102,241,.35);
  transition: all var(--transition);
}
.admin-avatar-big {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 4px 14px rgba(245,158,11,.35);
}
.avatar-crown-big {
  position: absolute;
  top: -10px;
  right: -8px;
  font-size: 22px;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,.2));
}
.profile-card h2 { font-size: 22px; margin-bottom: 6px; }
.profile-role { margin-bottom: 22px; }
.profile-stats { display: flex; justify-content: center; gap: 40px; }
.p-stat { display: flex; flex-direction: column; align-items: center; }
.p-stat-value { font-size: 26px; font-weight: 800; color: var(--primary); line-height: 1.2; }
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
  transition: all var(--transition);
}
.solved-tag:hover { background: #a7f3d0; transform: scale(1.05); }

/* Quick Links */
.profile-side { display: flex; flex-direction: column; gap: 20px; }
.quick-links h3 { margin-bottom: 16px; font-size: 16px; }
.link-grid { display: flex; flex-direction: column; gap: 8px; }
.quick-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  background: #f8fafc;
  transition: all var(--transition);
  color: var(--text) !important;
  font-weight: 500;
  font-size: 14px;
}
.quick-link:hover { background: #eef2ff; transform: translateX(4px); }
.quick-link svg { flex-shrink: 0; color: var(--text-light); }

/* Admin Panel */
.admin-panel {
  border: 1.5px solid #fde68a;
  background: linear-gradient(135deg, #fffbeb, #fefce8);
}
.admin-panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.admin-panel-header svg { color: #d97706; }
.admin-panel-header h3 { font-size: 16px; color: #92400e; margin-bottom: 0; }
.admin-panel-desc { font-size: 13px; color: #a16207; margin-bottom: 16px; }
.admin-links { display: flex; flex-direction: column; gap: 8px; }
.admin-link-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  background: rgba(255,255,255,.7);
  border: 1px solid #fde68a;
  transition: all var(--transition);
  color: #92400e !important;
  font-weight: 600;
  font-size: 14px;
}
.admin-link-btn:hover { background: rgba(255,255,255,.9); border-color: #f59e0b; transform: translateX(4px); }
.admin-link-btn.new { border-color: #fbbf24; background: #fef3c7; }
.admin-link-btn.new:hover { background: #fde68a; }
.admin-link-btn svg { flex-shrink: 0; }
.link-arrow { margin-left: auto; color: #d97706; }
.admin-link-btn:hover .link-arrow { transform: translateX(4px); transition: transform var(--transition); }
</style>
