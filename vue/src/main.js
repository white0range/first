import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'

// 懒加载所有页面
const Home = () => import('./views/Home.vue')
const Login = () => import('./views/Login.vue')
const Register = () => import('./views/Register.vue')
const ProblemDetail = () => import('./views/ProblemDetail.vue')
const Leaderboard = () => import('./views/Leaderboard.vue')
const Profile = () => import('./views/Profile.vue')
const MySubmissions = () => import('./views/MySubmissions.vue')
const SubmissionDetail = () => import('./views/SubmissionDetail.vue')
const AdminProblems = () => import('./views/admin/AdminProblems.vue')
const AdminProblemEdit = () => import('./views/admin/AdminProblemEdit.vue')
const AdminTags = () => import('./views/admin/AdminTags.vue')

const routes = [
  { path: '/', name: 'Home', component: Home, meta: { title: '题库 - GoJo' } },
  { path: '/login', name: 'Login', component: Login, meta: { title: '登录 - GoJo' } },
  { path: '/register', name: 'Register', component: Register, meta: { title: '注册 - GoJo' } },
  { path: '/problems/:id', name: 'ProblemDetail', component: ProblemDetail, meta: { title: '题目详情' } },
  { path: '/leaderboard', name: 'Leaderboard', component: Leaderboard, meta: { title: '排行榜 - GoJo' } },
  { path: '/profile', name: 'Profile', component: Profile, meta: { title: '个人中心 - GoJo', auth: true } },
  { path: '/my-submissions', name: 'MySubmissions', component: MySubmissions, meta: { title: '我的提交 - GoJo', auth: true } },
  { path: '/submissions/:id', name: 'SubmissionDetail', component: SubmissionDetail, meta: { title: '提交详情', auth: true } },
  { path: '/admin/problems', name: 'AdminProblems', component: AdminProblems, meta: { title: '题目管理', auth: true, admin: true } },
  { path: '/admin/problems/new', name: 'AdminProblemNew', component: AdminProblemEdit, meta: { title: '新建题目', auth: true, admin: true } },
  { path: '/admin/problems/:id/edit', name: 'AdminProblemEdit', component: AdminProblemEdit, meta: { title: '编辑题目', auth: true, admin: true } },
  { path: '/admin/tags', name: 'AdminTags', component: AdminTags, meta: { title: '标签管理', auth: true, admin: true } },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
})

// 路由守卫：检查登录状态
router.beforeEach((to, from, next) => {
  document.title = to.meta.title || 'GoJo'
  const token = localStorage.getItem('token')
  const role = localStorage.getItem('role')
  if (to.meta.auth && !token) {
    next(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  } else if (to.meta.admin && role !== '1') {
    next('/')
  } else {
    next()
  }
})

const app = createApp(App)
app.use(router)
app.mount('#app')
