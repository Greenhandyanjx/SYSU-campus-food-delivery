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
  },
  // 个人资料相关路由
  {
    path: "/rider/profile/edit",
    component: () => import('@/views/rider/profile/edit.vue'),
    name: "RiderProfileEdit",
    meta: {
      title: "个人资料",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/security",
    component: () => import('@/views/rider/profile/security.vue'),
    name: "RiderProfileSecurity",
    meta: {
      title: "账户安全",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/payment",
    component: () => import('@/views/rider/profile/payment.vue'),
    name: "RiderProfilePayment",
    meta: {
      title: "收款设置",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/work",
    component: () => import('@/views/rider/profile/work.vue'),
    name: "RiderProfileWork",
    meta: {
      title: "工作偏好",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/notification",
    component: () => import('@/views/rider/profile/notification.vue'),
    name: "RiderProfileNotification",
    meta: {
      title: "消息通知",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/map",
    component: () => import('@/views/rider/profile/map.vue'),
    name: "RiderProfileMap",
    meta: {
      title: "地图设置",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/help",
    component: () => import('@/views/rider/profile/help.vue'),
    name: "RiderProfileHelp",
    meta: {
      title: "帮助中心",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/feedback",
    component: () => import('@/views/rider/profile/feedback.vue'),
    name: "RiderProfileFeedback",
    meta: {
      title: "意见反馈",
      requireAuth: true,
      role: "rider"
    }
  },
  {
    path: "/rider/profile/contact",
    component: () => import('@/views/rider/profile/contact.vue'),
    name: "RiderProfileContact",
    meta: {
      title: "联系客服",
      requireAuth: true,
      role: "rider"
    }
  }
]

export default routes