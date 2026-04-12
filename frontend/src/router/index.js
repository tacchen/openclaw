import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue')
  },
  {
    path: '/push-config',
    name: 'PushConfig',
    component: () => import('../views/PushConfig.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
