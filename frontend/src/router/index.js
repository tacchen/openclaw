import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫：检查 localStorage 中的 token
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const isLoggedIn = !!token
  
  console.log('Router guard:', to.name, 'isLoggedIn:', isLoggedIn, 'requiresAuth:', to.meta.requiresAuth)
  
  // 需要认证但未登录，跳转到登录页
  if (to.meta.requiresAuth && !isLoggedIn) {
    console.log('Redirecting to login')
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } 
  // 已登录但访问登录页，跳转到首页
  else if (isLoggedIn && (to.name === 'Login' || to.name === 'Register')) {
    console.log('Already logged in, redirecting to home')
    next({ name: 'Home' })
  } 
  else {
    next()
  }
})

export default router
