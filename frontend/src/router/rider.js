// 骑手端路由配置
const routes = [
  {
    path: "/rider",
    component: () => import('@/views/rider/main/index.vue'),
    name: "RiderHome",
    meta: {
      title: "骑手首页",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/dashboard",
    component: () => import('@/views/rider/Dashboard.vue'),
    name: "RiderDashboard",
    meta: {
      title: "骑手工作台",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/orders",
    component: () => import('@/views/rider/orders/index.vue'),
    name: "RiderOrders",
    meta: {
      title: "历史订单",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/wallet",
    component: () => import('@/views/rider/wallet/index.vue'),
    name: "RiderWallet",
    meta: {
      title: "我的钱包",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/stats",
    component: () => import('@/views/rider/stats/index.vue'),
    name: "RiderStats",
    meta: {
      title: "数据统计",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile",
    component: () => import('@/views/rider/profile/index.vue'),
    name: "RiderProfile",
    meta: {
      title: "个人中心",
      requireAuth: true,
      role: "rider"
    }
  }
]

export default routes