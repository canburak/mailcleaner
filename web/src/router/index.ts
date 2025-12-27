import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      redirect: '/accounts',
    },
    {
      path: '/accounts',
      name: 'accounts',
      component: () => import('../views/AccountsView.vue'),
    },
    {
      path: '/accounts/:id',
      name: 'account-detail',
      component: () => import('../views/AccountDetailView.vue'),
      props: true,
    },
    {
      path: '/accounts/:id/rules',
      name: 'account-rules',
      component: () => import('../views/RulesView.vue'),
      props: true,
    },
    {
      path: '/accounts/:id/preview',
      name: 'account-preview',
      component: () => import('../views/PreviewView.vue'),
      props: true,
    },
  ],
})

export default router
