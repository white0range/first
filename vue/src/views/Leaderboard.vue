<template>
  <div class="leaderboard-page">
    <div class="page-header">
      <h1>🏆 全服排行榜</h1>
      <p class="subtitle">卷王争霸，谁与争锋</p>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner"></div>
    </div>

    <template v-else>
      <!-- 前三名特殊展示 -->
      <div class="top-three" v-if="leaderboard.length >= 3">
        <div class="top-card top-2" v-if="leaderboard[1]">
          <div class="top-avatar">🥈</div>
          <div class="top-name">{{ leaderboard[1].username }}</div>
          <div class="top-score">{{ leaderboard[1].score }} 分</div>
        </div>
        <div class="top-card top-1" v-if="leaderboard[0]">
          <div class="top-crown">👑</div>
          <div class="top-avatar">🥇</div>
          <div class="top-name">{{ leaderboard[0].username }}</div>
          <div class="top-score">{{ leaderboard[0].score }} 分</div>
        </div>
        <div class="top-card top-3" v-if="leaderboard[2]">
          <div class="top-avatar">🥉</div>
          <div class="top-name">{{ leaderboard[2].username }}</div>
          <div class="top-score">{{ leaderboard[2].score }} 分</div>
        </div>
      </div>

      <!-- 完整列表 -->
      <div class="card" v-if="leaderboard.length > 0">
        <!-- 自己的排名提示 -->
        <div class="my-rank-bar" v-if="store.isLoggedIn && myRank > 0">
          🎯 我的排名：<strong>第 {{ myRank }} 名</strong>，战力值：<strong>{{ myScore }}</strong>
        </div>
        <div class="my-rank-bar no-rank" v-else-if="store.isLoggedIn && myRank === -1">
          🏃 你还没有上榜，快去刷题吧！
        </div>
        <div class="rank-table">
          <div class="rank-header">
            <span class="col-rank">排名</span>
            <span class="col-user">玩家</span>
            <span class="col-score">战力值</span>
          </div>
          <div
            class="rank-row"
            v-for="item in leaderboard"
            :key="item.user_id"
            :class="{ 'is-me': item.user_id === myUserId }"
          >
            <span class="col-rank">
              <span class="rank-num" :class="'rank-' + item.rank">{{ item.rank }}</span>
            </span>
            <span class="col-user">
              <span class="user-dot" :style="{ background: avatarColor(item.user_id) }"></span>
              {{ item.username }}
              <span v-if="item.user_id === myUserId" class="me-badge">我</span>
            </span>
            <span class="col-score">{{ item.score }}</span>
          </div>
        </div>
      </div>

      <div class="empty" v-else>
        <span>🏜️</span>
        <p>还没有人上榜，快来提交代码吧！</p>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getLeaderboard } from '../api/index.js'
import { store } from '../store/index.js'

const leaderboard = ref([])
const loading = ref(true)
const myRank = ref(-1)
const myScore = ref(0)
const myUserId = ref(null)

// 从 localStorage 解析 token 获取自己的用户 ID（简化处理）
function parseMyUserId() {
  const token = localStorage.getItem('token')
  if (!token) return
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    myUserId.value = payload.user_id
  } catch (e) { /* ignore */ }
}

const colors = ['#6366f1', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#8b5cf6', '#ef4444', '#14b8a6']
function avatarColor(id) { return colors[id % colors.length] }

async function fetchLeaderboard() {
  loading.value = true
  try {
    // 如果已登录，自动带上 token
    const config = {}
    const token = localStorage.getItem('token')
    if (token) config.headers = { Authorization: `Bearer ${token}` }
    const res = await getLeaderboard(config)
    const data = res.data.data || res.data
    leaderboard.value = data.top_50 || []
    myRank.value = data.my_rank || -1
    myScore.value = data.my_score || 0
  } catch (e) {
    console.error('获取排行榜失败', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  parseMyUserId()
  fetchLeaderboard()
})
</script>

<style scoped>
.subtitle { color: var(--text-secondary); font-size: 14px; margin-top: 4px; }

/* 前三名 */
.top-three {
  display: flex;
  justify-content: center;
  align-items: flex-end;
  gap: 20px;
  margin-bottom: 32px;
  flex-wrap: wrap;
}
.top-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 32px;
  border-radius: var(--radius);
  background: var(--bg-card);
  border: 1px solid var(--border);
  min-width: 140px;
  position: relative;
}
.top-1 { order: 1; transform: scale(1.1); border-color: #fbbf24; box-shadow: 0 0 24px rgba(251,191,36,.2); }
.top-2 { order: 0; }
.top-3 { order: 2; }
.top-crown { font-size: 32px; position: absolute; top: -20px; }
.top-avatar { font-size: 40px; margin-bottom: 8px; }
.top-name { font-weight: 700; font-size: 16px; }
.top-score { font-size: 20px; font-weight: 800; color: var(--primary); }

/* 列表 */
.my-rank-bar {
  text-align: center;
  padding: 12px 20px;
  margin-bottom: 8px;
  background: linear-gradient(135deg, #eef2ff, #f0f9ff);
  border-radius: var(--radius-sm);
  font-size: 14px;
  color: var(--text-secondary);
}
.my-rank-bar strong { color: var(--primary); }
.my-rank-bar.no-rank { background: #fefce8; }

.rank-table { }
.rank-header, .rank-row {
  display: grid;
  grid-template-columns: 80px 1fr 100px;
  align-items: center;
  padding: 12px 20px;
}
.rank-header { font-size: 12px; font-weight: 600; color: var(--text-secondary); border-bottom: 1px solid var(--border); }
.rank-row { border-bottom: 1px solid var(--border); transition: var(--transition); }
.rank-row:hover { background: #f8faff; }
.rank-row.is-me { background: #eef2ff; }
.rank-num { font-weight: 700; font-size: 15px; }
.rank-1 { color: #f59e0b; }
.rank-2 { color: #94a3b8; }
.rank-3 { color: #d97706; }
.col-user { display: flex; align-items: center; gap: 8px; }
.user-dot { width: 10px; height: 10px; border-radius: 50%; }
.me-badge { font-size: 10px; background: var(--primary); color: #fff; padding: 1px 6px; border-radius: 4px; }
.col-score { font-weight: 600; }

.empty { text-align: center; padding: 80px; color: var(--text-light); }
.empty span { font-size: 48px; display: block; margin-bottom: 8px; }
</style>
