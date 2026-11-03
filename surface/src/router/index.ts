import { createRouter, createWebHistory } from 'vue-router'
import { isLoggedIn } from '../auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      component: () => import('../components/AppLayout.vue'),
      children: [
        {
          path: '',
          name: 'chat',
          component: () => import('../views/ChatView.vue')
        },
        {
          path: 'documents',
          name: 'documents',
          component: () => import('../views/DocumentsView.vue')
        }
      ]
    }
  ]
})

router.beforeEach((to) => {
  const loggedIn = isLoggedIn()

  // 未登录时只允许访问公开页面，其余跳转登录页并记录来源
  if (!to.meta.public && !loggedIn) {
    return {
      name: 'login',
      query: to.fullPath === '/' ? {} : { redirect: to.fullPath }
    }
  }

  // 已登录时不再展示登录页
  if (to.name === 'login' && loggedIn) {
    return { name: 'chat' }
  }
})

export default router
