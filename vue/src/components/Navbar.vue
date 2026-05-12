<template>
  <nav class="navbar" :class="{ scrolled }">
    <div class="navbar-inner">
      <router-link to="/" class="logo">
        <span class="logo-icon">⚡</span>
        <span class="logo-text">GoJo</span>
      </router-link>

      <div class="nav-links">
        <router-link to="/" class="nav-link" active-class="active">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
          题库
        </router-link>
        <router-link to="/leaderboard" class="nav-link" active-class="active">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 22V2h4v20"/></svg>
          排行榜
        </router-link>
        <template v-if="store.isAdmin">
          <router-link to="/admin/problems" class="nav-link admin-link" active-class="active">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
            管理
            <span class="admin-badge">Admin</span>
          </router-link>
        </template>
      </div>

      <div class="nav-actions">
        <template v-if="store.isLoggedIn">
          <div class="user-menu" @click="toggleMenu">
            <span class="user-avatar" :class="{ 'admin-avatar': store.isAdmin }">
              {{ store.username?.[0]?.toUpperCase() }}
              <span v-if="store.isAdmin" class="avatar-crown">👑</span>
            </span>
            <span class="user-name">{{ store.username }}</span>
            <svg class="arrow-icon" :class="{ rotated: menuOpen }" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>
          </div>
          <div class="dropdown" v-if="menuOpen" @click.stop>
            <router-link to="/profile" class="dropdown-item" @click="menuOpen=false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
              个人中心
            </router-link>
            <router-link to="/my-submissions" class="dropdown-item" @click="menuOpen=false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
              我的提交
            </router-link>
            <template v-if="store.isAdmin">
              <div class="dropdown-divider"></div>
              <div class="dropdown-label">管理</div>
              <router-link to="/admin/problems" class="dropdown-item" @click="menuOpen=false">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
                题目管理
              </router-link>
              <router-link to="/admin/tags" class="dropdown-item" @click="menuOpen=false">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/></svg>
                标签管理
              </router-link>
            </template>
            <div class="dropdown-divider"></div>
            <a class="dropdown-item logout" @click="handleLogout">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
              退出登录
            </a>
          </div>
        </template>
        <template v-else>
          <router-link to="/login" class="btn btn-ghost btn-sm">登录</router-link>
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
  background: rgba(255,255,255,.85);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
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
  padding: 0 24px; height: 64px;
}

/* Logo */
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

/* Nav Links */
.nav-links { display: flex; gap: 4px; }
.nav-link {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 10px;
  font-size: 14px; font-weight: 600;
  color: var(--text-secondary) !important;
  transition: all var(--transition);
  position: relative;
}
.nav-link svg { flex-shrink: 0; }
.nav-link:hover { background: #f1f5f9; color: var(--text) !important; }
.nav-link.active { background: #eef2ff; color: var(--primary) !important; }

.admin-link { position: relative; }
.admin-badge {
  font-size: 9px;
  font-weight: 700;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: #fff;
  padding: 1px 6px;
  border-radius: 4px;
  letter-spacing: 0.3px;
  text-transform: uppercase;
}
.admin-link.active { background: #fef3c7; color: #92400e !important; }

/* User Actions */
.nav-actions { display: flex; align-items: center; gap: 8px; position: relative; }

.user-menu {
  display: flex; align-items: center; gap: 10px;
  padding: 6px 12px; border-radius: 12px; cursor: pointer;
  transition: all var(--transition);
}
.user-menu:hover { background: #f1f5f9; }

.user-avatar {
  position: relative;
  width: 36px; height: 36px; border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--accent));
  color: #fff; display: flex; align-items: center; justify-content: center;
  font-weight: 800; font-size: 15px;
  box-shadow: 0 2px 6px rgba(99,102,241,.3);
  transition: all var(--transition);
}
.admin-avatar {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 2px 6px rgba(245,158,11,.3);
}
.avatar-crown {
  position: absolute;
  top: -8px;
  right: -6px;
  font-size: 12px;
  filter: drop-shadow(0 1px 2px rgba(0,0,0,.2));
}
.user-name { font-size: 14px; font-weight: 600; max-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.arrow-icon { color: var(--text-light); transition: transform var(--transition); flex-shrink: 0; }
.arrow-icon.rotated { transform: rotate(180deg); }

/* Dropdown */
.dropdown {
  position: absolute; top: calc(100% + 10px); right: 0;
  background: var(--bg-card); border: 1px solid var(--border);
  border-radius: var(--radius); box-shadow: var(--shadow-xl);
  min-width: 200px; padding: 6px; z-index: 200;
  animation: dropIn .18s ease;
}
@keyframes dropIn {
  from { opacity: 0; transform: translateY(-8px) scale(.96); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.dropdown-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border-radius: var(--radius-sm);
  font-size: 14px; color: var(--text) !important; transition: all var(--transition);
}
.dropdown-item svg { flex-shrink: 0; color: var(--text-light); }
.dropdown-item:hover { background: #f1f5f9; }
.dropdown-item:hover svg { color: var(--text-secondary); }
.logout { color: var(--danger) !important; }
.logout:hover { background: #fef2f2 !important; }
.logout svg { color: var(--danger) !important; }
.dropdown-divider { height: 1px; background: var(--border); margin: 4px 0; }
.dropdown-label {
  padding: 6px 14px 4px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-light);
}

.dropdown-overlay { position: fixed; inset: 0; z-index: 99; }

/* Mobile */
@media (max-width: 768px) {
  .nav-links { display: none; }
}
@media (max-width: 640px) {
  .user-name { display: none; }
  .navbar-inner { padding: 0 14px; }
}
</style>
