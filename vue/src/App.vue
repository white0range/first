<template>
  <div class="app-container">
    <Navbar />
    <main class="main-content">
      <router-view />
    </main>
    <footer class="app-footer">
      <div class="footer-inner">
        <span>GoJo Online Judge 2026</span>
        <span class="footer-desc">用代码改变世界</span>
      </div>
    </footer>
  </div>
</template>

<script setup>
import Navbar from './components/Navbar.vue'
</script>

<style>
:root {
  --primary: #6366f1;
  --primary-light: #818cf8;
  --primary-dark: #4f46e5;
  --accent: #06b6d4;
  --success: #10b981;
  --warning: #f59e0b;
  --danger: #ef4444;
  --bg: #fafbfc;
  --bg-card: #ffffff;
  --text: #1e293b;
  --text-secondary: #64748b;
  --text-light: #94a3b8;
  --border: #e2e8f0;
  --radius: 14px;
  --radius-sm: 8px;
  --shadow: 0 1px 3px rgba(0,0,0,.05), 0 1px 2px rgba(0,0,0,.03);
  --shadow-md: 0 4px 6px rgba(0,0,0,.06), 0 2px 4px rgba(0,0,0,.04);
  --shadow-lg: 0 12px 28px rgba(0,0,0,.07), 0 4px 12px rgba(0,0,0,.04);
  --transition: 0.25s cubic-bezier(.4,0,.2,1);
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.7;
  -webkit-font-smoothing: antialiased;
}

a { color: var(--primary); text-decoration: none; }
a:hover { color: var(--primary-dark); }

.app-container { min-height: 100vh; display: flex; flex-direction: column; }

.main-content {
  flex: 1;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 28px 24px;
}

.app-footer {
  border-top: 1px solid var(--border);
  background: var(--bg-card);
  padding: 18px 24px;
  margin-top: 48px;
}
.footer-inner {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: var(--text-light);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 20px;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition);
}
.btn-primary { background: linear-gradient(135deg, var(--primary), var(--primary-dark)); color: #fff; box-shadow: 0 2px 8px rgba(99,102,241,.25); }
.btn-primary:hover { transform: translateY(-1px); box-shadow: 0 4px 16px rgba(99,102,241,.35); }
.btn-success { background: var(--success); color: #fff; }
.btn-danger { background: var(--danger); color: #fff; }
.btn-outline { background: transparent; border: 1.5px solid var(--border); color: var(--text); }
.btn-outline:hover { border-color: var(--primary); color: var(--primary); }
.btn-sm { padding: 6px 14px; font-size: 12px; border-radius: 6px; }
.btn:disabled { opacity: 0.45; cursor: not-allowed; transform: none !important; }

.card {
  background: var(--bg-card);
  border-radius: var(--radius);
  border: 1px solid var(--border);
  padding: 28px;
  box-shadow: var(--shadow);
}

.input {
  width: 100%;
  padding: 11px 16px;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 14px;
  transition: all var(--transition);
  outline: none;
  background: var(--bg-card);
  color: var(--text);
}
.input:focus { border-color: var(--primary); box-shadow: 0 0 0 3px rgba(99,102,241,.12); }
.input::placeholder { color: var(--text-light); }

textarea.input { resize: vertical; min-height: 100px; font-family: 'Fira Code', 'Consolas', 'Monaco', monospace; }

select.input { cursor: pointer; }

.badge {
  display: inline-flex; align-items: center; gap: 3px;
  padding: 3px 10px; border-radius: 100px;
  font-size: 11px; font-weight: 700;
}
.badge-success { background: #d1fae5; color: #065f46; }
.badge-warning { background: #fef3c7; color: #92400e; }
.badge-danger { background: #fee2e2; color: #991b1b; }
.badge-info { background: #dbeafe; color: #1e40af; }
.badge-pending { background: #e0e7ff; color: #3730a3; }

.status-AC { color: #10b981; font-weight: 700; }
.status-WA { color: #ef4444; font-weight: 700; }
.status-TLE { color: #f59e0b; font-weight: 700; }
.status-CE { color: #8b5cf6; font-weight: 700; }
.status-RE { color: #ef4444; font-weight: 700; }
.status-SE { color: #6b7280; font-weight: 700; }
.status-Pending { color: #6366f1; font-weight: 700; }

@keyframes spin { to { transform: rotate(360deg); } }

.spinner {
  width: 20px; height: 20px;
  border: 2.5px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin .65s linear infinite;
}

.loading-center {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 80px 20px;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 6px;
  margin-top: 28px;
}
.pagination button {
  min-width: 38px; height: 38px;
  padding: 0 12px;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all var(--transition);
  color: var(--text-secondary);
}
.pagination button:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
.pagination button.active { background: var(--primary); color: #fff; border-color: var(--primary); font-weight: 700; }
.pagination button:disabled { opacity: 0.35; cursor: not-allowed; }

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
  flex-wrap: wrap;
  gap: 14px;
}
.page-header h1 {
  font-size: 26px;
  font-weight: 800;
  color: var(--text);
}

@media (max-width: 640px) {
  .main-content { padding: 16px 12px; }
  .card { padding: 18px; }
}
</style>
