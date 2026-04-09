import { createRouter, createWebHistory } from 'vue-router'
import { Home, Download, Online, OnlineDetail, OnlineReader, Extract, Setting, MangaDetail, History } from '../views'

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
      path: '/online',
      name: 'online',
      component: Online,
    },
    {
      path: '/online/detail',
      name: 'online-detail',
      component: OnlineDetail,
    },
    {
      path: '/online/read',
      name: 'online-read',
      component: OnlineReader,
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
