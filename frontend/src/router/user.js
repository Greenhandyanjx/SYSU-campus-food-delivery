export default [
  { path: '/user/home', component: () => import('@/views/user/main/index.vue') },
  { path: '/user/orderlist', component: () => import('@/views/user/orderlist/index.vue') },
  { path: '/user/cart', component: () => import('@/views/user/chart/index.vue') },
  { path: '/user/my', component: () => import('@/views/user/my/index.vue') },
]