<template>
  <div class="leaderboard-page">
    <div class="page-header">
      <h1>全服排行榜</h1>
      <span class="subtitle">卷王争霸，谁与争锋</span>
    </div>

    <div class="loading-center" v-if="loading">
      <div class="spinner spinner-dark" style="width:32px;height:32px;border-width:3px"></div>
    </div>

    <template v-else>
      <!-- 前三名特殊展示 -->
      <div class="top-three" v-if="leaderboard.length >= 3">
        <div class="top-card top-2" v-if="leaderboard[1]">
          <div class="top-medal">🥈</div>
          <div class="top-avatar">{{ leaderboard[1].username[0]?.toUpperCase() }}</div>
          <div class="top-name">{{ leaderboard[1].username }}</div>
          <div class="top-score">{{ leaderboard[1].score }}</div>
          <div class="top-label">积分</div>
        </div>
        <div class="top-card top-1" v-if="leaderboard[0]">
          <div class="top-crown">👑</div>
          <div class="top-medal">🥇</div>
          <div class="top-avatar gold">{{ leaderboard[0].username[0]?.toUpperCase() }}</div>
          <div class="top-name">{{ leaderboard[0].username }}</div>
          <div class="top-score">{{ leaderboard[0].score }}</div>
          <div class="top-label">积分</div>
        </div>
        <div class="top-card top-3" v-if="leaderboard[2]">
          <div class="top-medal">🥉</div>
          <div class="top-avatar">{{ leaderboard[2].username[0]?.toUpperCase() }}</div>
          <div class="top-name">{{ leaderboard[2].username }}</div>
          <div class="top-score">{{ leaderboard[2].score }}</div>
          <div class="top-label">积分</div>
        </div>
      </div>

      <!-- 完整列表 -->
      <div class="card" v-if="leaderboard.length > 0" style="padding:0;overflow:hidden">
        <!-- 自己的排名提示 -->
        <div class="my-rank-bar" v-if="store.isLoggedIn && myRank > 0">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
          <span>我的排名：<strong>第 {{ myRank }} 名</strong>，战力值：<strong>{{ myScore }}</strong></span>
        </div>
        <div class="my-rank-bar no-rank" v-else-if="store.isLoggedIn && myRank === -1">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 22V2h4v20"/></svg>
          <span>你还没有上榜，快去刷题吧！</span>
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
              <span class="rank-num" :class="'rank-' + item.rank" v-if="item.rank <= 3">
                <svg v-if="item.rank === 1" width="18" height="18" viewBox="0 0 24 24" fill="#f59e0b" stroke="#f59e0b" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                <svg v-else-if="item.rank === 2" width="18" height="18" viewBox="0 0 24 24" fill="#94a3b8" stroke="#94a3b8" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                <svg v-else-if="item.rank === 3" width="18" height="18" viewBox="0 0 24 24" fill="#d97706" stroke="#d97706" stroke-width="1"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
              </span>
              <span class="rank-num rank-default" v-else>{{ item.rank }}</span>
            </span>
            <span class="col-user">
              <span class="user-avatar-sm" :style="{ background: avatarColor(item.user_id) }">{{ item.username[0]?.toUpperCase() }}</span>
              <span class="user-name-text">{{ item.username }}</span>
              <span v-if="item.user_id === myUserId" class="me-badge">我</span>
            </span>
            <span class="col-score">{{ item.score }}</span>
          </div>
        </div>
      </div>

      <div class="empty-state" v-else>
        <span class="empty-icon">🏜️</span>
        <p class="empty-text">还没有人上榜</p>
        <p class="empty-hint">快来提交代码，争夺榜首吧！</p>
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

function parseMyUserId() {
  const token = localStorage.getItem('token')
  if (!token) return
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    myUserId.value = payload.user_id
  } catch (e) {}
}

const colors = ['#6366f1', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#8b5cf6', '#ef4444', '#14b8a6']
function avatarColor(id) { return colors[(id || 1) % colors.length] }

async function fetchLeaderboard() {
  loading.value = true
  try {
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
.subtitle { color: var(--text-light); font-size: 14px; margin-top: 4px; font-weight: 400; }

/* ===== Top 3 ===== */
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
  padding: 28px 36px 24px;
  border-radius: var(--radius);
  background: var(--bg-card);
  border: 1.5px solid var(--border);
  min-width: 150px;
  position: relative;
  transition: all var(--transition);
}
.top-1 {
  transform: scale(1.08);
  border-color: #fbbf24;
  box-shadow: 0 0 30px rgba(251,191,36,.15);
  background: linear-gradient(180deg, #fffbeb, #fff);
}
.top-2 { background: linear-gradient(180deg, #f8fafc, #fff); }
.top-3 { background: linear-gradient(180deg, #fff7ed, #fff); }
.top-card:hover { transform: translateY(-4px) scale(1.08); }
.top-2:hover { transform: translateY(-4px); }
.top-3:hover { transform: translateY(-4px); }
.top-crown {
  position: absolute;
  top: -18px;
  font-size: 32px;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,.15));
  animation: crownBounce 2s ease-in-out infinite;
}
@keyframes crownBounce {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-4px) rotate(5deg); }
}
.top-medal { font-size: 28px; margin-bottom: 8px; }
.top-avatar {
  width: 48px; height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--accent));
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-weight: 800; font-size: 20px;
  margin-bottom: 10px;
}
.top-avatar.gold { background: linear-gradient(135deg, #f59e0b, #d97706); }
.top-name { font-weight: 700; font-size: 16px; margin-bottom: 4px; }
.top-score { font-size: 26px; font-weight: 800; background: linear-gradient(135deg, var(--primary), var(--accent)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.top-1 .top-score { background: linear-gradient(135deg, #f59e0b, #d97706); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.top-label { font-size: 11px; color: var(--text-light); font-weight: 500; }

/* List */
.my-rank-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px 20px;
  background: linear-gradient(135deg, #eef2ff, #f0f9ff);
  border-bottom: 1px solid var(--border);
  font-size: 14px;
  color: var(--text-secondary);
}
.my-rank-bar strong { color: var(--primary); font-weight: 700; }
.my-rank-bar svg { flex-shrink: 0; color: var(--primary); }
.my-rank-bar.no-rank { background: #fefce8; }
.my-rank-bar.no-rank svg { color: #d97706; }

.rank-table { }
.rank-header, .rank-row {
  display: grid;
  grid-template-columns: 70px 1fr 100px;
  align-items: center;
  padding: 14px 24px;
  gap: 8px;
}
.rank-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-light);
  border-bottom: 1px solid var(--border);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.rank-row {
  border-bottom: 1px solid var(--border);
  transition: all var(--transition);
}
.rank-row:last-child { border-bottom: none; }
.rank-row:hover { background: #f8faff; }
.rank-row.is-me { background: #eef2ff; }
.rank-num { display: flex; align-items: center; font-weight: 700; font-size: 15px; }
.rank-default { color: var(--text-secondary); }

.col-user { display: flex; align-items: center; gap: 10px; }
.user-avatar-sm {
  width: 32px; height: 32px;
  border-radius: 50%;
  color: #fff;
  display: flex; align-items: center; justify-content: center;
  font-weight: 700; font-size: 13px;
  flex-shrink: 0;
}
.user-name-text { font-weight: 600; font-size: 14px; }
.me-badge {
  font-size: 10px;
  font-weight: 700;
  background: var(--primary);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  margin-left: 2px;
}
.col-score { font-weight: 700; font-size: 16px; color: var(--text); }

@media (max-width: 640px) {
  .top-three { gap: 12px; }
  .top-card { min-width: 110px; padding: 20px 16px; }
  .top-1 { transform: scale(1.05); }
  .rank-header, .rank-row { padding: 12px 16px; grid-template-columns: 50px 1fr 70px; }
  .user-name-text { max-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
}
</style>
