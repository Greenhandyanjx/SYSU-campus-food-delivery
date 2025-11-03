import { createRouter, createWebHistory } from 'vue-router'
import userRoutes from './user'
// import riderRoutes from './rider'
import merchantRoutes from './merchant'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: () => import('@/views/login/Login.vue') },
    { path: '/register', component: () => import('@/views/login/register.vue') },
    // ...riderRoutes,
    ...merchantRoutes,
    ...userRoutes,
  ],
})

export default router
// import { createRouter, createWebHistory } from 'vue-router'

// const router = createRouter({
//   history: createWebHistory(),
//   routes: [
//     {
//       path: '/',
//       redirect: '/login'
//     },
//     {
//       path: '/login',
//       component: () => import('@/views/login/Login.vue')
//     },
//     {
//       path: '/merchant/dashboard',
//       component: () => import('@/views/merchant/Dashboard.vue')
//     }
//   ]
// })

// export default router
