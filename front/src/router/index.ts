import { createRouter, createWebHistory } from 'vue-router'
import { Home, Download, Extract, Setting, MangaDetail, History } from '../views'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home,
    },
    {
      path: '/download',
      name: 'download',
      component: Download,
    },
    {
      path: '/extract',
      name: 'extract',
      component: Extract,
    },
    {
      path: '/setting',
      name: 'setting',
      component: Setting,
    },
    {
      path: '/manga/:path',
      name: 'manga',
      component: MangaDetail,
    },
    {
      path: '/history',
      name: 'history',
      component: History,
    }
  ],
})

export default router
