<template>
  <nav class="navbar" :class="{ scrolled }">
    <div class="navbar-inner">
      <router-link to="/" class="logo">
        <span class="logo-icon">⚡</span>
        <span class="logo-text">GoJo</span>
      </router-link>

      <div class="nav-links">
        <router-link to="/" class="nav-link" active-class="active">📚 题库</router-link>
        <router-link to="/leaderboard" class="nav-link" active-class="active">🏆 排行榜</router-link>
        <template v-if="store.isAdmin">
          <router-link to="/admin/problems" class="nav-link admin-link" active-class="active">⚙️ 管理</router-link>
        </template>
      </div>

      <div class="nav-actions">
        <template v-if="store.isLoggedIn">
          <div class="user-menu" @click="toggleMenu">
            <span class="user-avatar">{{ store.username?.[0]?.toUpperCase() }}</span>
            <span class="user-name">{{ store.username }}</span>
            <span class="arrow">▾</span>
          </div>
          <div class="dropdown" v-if="menuOpen" @click.stop>
            <router-link to="/profile" class="dropdown-item" @click="menuOpen=false">👤 个人中心</router-link>
            <router-link to="/my-submissions" class="dropdown-item" @click="menuOpen=false">📋 我的提交</router-link>
            <div class="dropdown-divider"></div>
            <a class="dropdown-item logout" @click="handleLogout">🚪 退出登录</a>
          </div>
        </template>
        <template v-else>
          <router-link to="/login" class="btn btn-outline btn-sm">登录</router-link>
          <router-link to="/register" class="btn btn-primary btn-sm">注册</router-link>
        </template>
      </div>
    </div>
    <!-- 点击空白关闭下拉 -->
    <div v-if="menuOpen" class="dropdown-overlay" @click="menuOpen=false"></div>
  </nav>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { store } from '../store/index.js'

const router = useRouter()
const menuOpen = ref(false)
const scrolled = ref(false)

function toggleMenu() { menuOpen.value = !menuOpen.value }
function handleLogout() {
  store.logout()
  menuOpen.value = false
  router.push('/')
}

function onScroll() { scrolled.value = window.scrollY > 8 }
onMounted(() => window.addEventListener('scroll', onScroll, { passive: true }))
onUnmounted(() => window.removeEventListener('scroll', onScroll))
</script>

<style scoped>
.navbar {
  position: sticky; top: 0; z-index: 100;
  background: rgba(255,255,255,.82);
  backdrop-filter: blur(16px) saturate(180%);
  -webkit-backdrop-filter: blur(16px) saturate(180%);
  border-bottom: 1px solid transparent;
  transition: all var(--transition);
}
.navbar.scrolled {
  border-bottom-color: var(--border);
  box-shadow: 0 1px 4px rgba(0,0,0,.04);
}

.navbar-inner {
  max-width: 1200px; margin: 0 auto;
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 24px; height: 60px;
}

.logo {
  display: flex; align-items: center; gap: 8px;
  font-size: 22px; font-weight: 800;
  color: var(--text) !important;
  letter-spacing: -0.5px;
}
.logo-icon { font-size: 26px; }
.logo-text {
  background: linear-gradient(135deg, var(--primary) 0%, var(--accent) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav-links { display: flex; gap: 2px; }
.nav-link {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 10px;
  font-size: 14px; font-weight: 600;
  color: var(--text-secondary) !important;
  transition: all var(--transition);
}
.nav-link:hover { background: #f1f5f9; color: var(--text) !important; }
.nav-link.active { background: #eef2ff; color: var(--primary) !important; }
.admin-link.active { background: #fef3c7; color: #92400e !important; }

.nav-actions { display: flex; align-items: center; gap: 10px; position: relative; }

.user-menu {
  display: flex; align-items: center; gap: 10px;
  padding: 6px 12px; border-radius: 10px; cursor: pointer;
  transition: all var(--transition);
}
.user-menu:hover { background: #f1f5f9; }

.user-avatar {
  width: 36px; height: 36px; border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--accent));
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-weight: 800; font-size: 15px;
  box-shadow: 0 2px 6px rgba(99,102,241,.3);
}
.user-name { font-size: 14px; font-weight: 600; }
.arrow { font-size: 10px; color: var(--text-light); margin-left: -2px; }

.dropdown {
  position: absolute; top: calc(100% + 10px); right: 0;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); box-shadow: var(--shadow-lg);
  min-width: 170px; padding: 6px; z-index: 200;
  animation: dropIn .18s ease;
}
@keyframes dropIn { from { opacity:0; transform:translateY(-8px) scale(.96); } to { opacity:1; transform:translateY(0) scale(1); } }

.dropdown-item {
  display: block; padding: 10px 14px; border-radius: var(--radius-sm);
  font-size: 14px; color: var(--text) !important; transition: all var(--transition);
}
.dropdown-item:hover { background: #f1f5f9; }
.logout { color: var(--danger) !important; }
.logout:hover { background: #fef2f2 !important; }
.dropdown-divider { height: 1px; background: var(--border); margin: 4px 0; }

.dropdown-overlay { position: fixed; inset: 0; z-index: 99; }

@media (max-width: 640px) {
  .nav-links { display: none; }
  .user-name { display: none; }
  .navbar-inner { padding: 0 14px; }
}
</style>
