import Layout from "@/layout/rider/index.vue";

const routes = [
  {
    path: "/rider",
    component: Layout,
    redirect: "/rider/dashboard",
    children: [
      {
        path: "dashboard",
        component: () => import("@/views/rider/dashboard/index.vue"),
        name: "RiderDashboard",
        meta: { title: "工作台" },
      },
      {
        path: "new",
        component: () => import("@/views/rider/RiderNewOrders.vue"),
        name: "RiderNewOrders",
        meta: { title: "待接单" },
      },
      {
        path: "ongoing",
        component: () => import("@/views/rider/RiderOngoing.vue"),
        name: "RiderOngoing",
        meta: { title: "进行中" },
      },
      {
        path: "history",
        component: () => import("@/views/rider/RiderHistory.vue"),
        name: "RiderHistory",
        meta: { title: "历史订单" },
      },
      {
        path: "me",
        component: () => import("@/views/rider/RiderMe.vue"),
        name: "RiderMe",
        meta: { title: "个人信息" },
      },
    ],
  },
];

export default routes;
