import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'chat',
      component: () => import('../views/ChatView.vue')
    },
    {
      path: '/documents',
      name: 'documents',
      component: () => import('../views/DocumentsView.vue')
    }
  ]
})

export default router
