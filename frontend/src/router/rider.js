import Dashboard from '@/views/rider/Dashboard.vue'
import Home from '@/views/rider/main/index.vue'
import { createRouter, createWebHistory } from "vue-router"
const routes = [
  {
    path: "/rider",
    component: () => import('@/views/rider/main/index.vue')
  },
  {
    path: "/rider/dashboard",
    component: () => import('@/views/rider/Dashboard.vue')
  },
  { path: "/rider/home", component: { template: '<div>测试 Rider Home</div>' } }
]
export default routes