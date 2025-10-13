// import { createRouter, createWebHistory } from "vue-router";
import Layout from "@/layout/merchant/index.vue";

const routes = [
  {
    path: "/merchant",
    component: () => import("@/views/login/Login.vue"),
    meta: { title: "商家登录", hidden: true, notNeedAuth: true },
  },
  {
    path: "/merchant/404",
    component: () => import("@/views/404.vue"),
    meta: { title: "页面不存在", hidden: true, notNeedAuth: true },
  },
  {
    path: "/merchant",
    component: Layout,
    redirect: "/merchant/dashboard",
    children: [
      {
        path: "dashboard",
        component: () => import("@/views/merchant/dashboard/index.vue"),
        name: "MerchantDashboard",
        meta: {
          title: "工作台",
          icon: "dashboard",
          affix: true,
        },
      },
      {
        path: "statistics",
        component: () => import("@/views/merchant/statistics/index.vue"),
        meta: {
          title: "数据统计",
          icon: "icon-statistics",
        },
      },
      {
        path: "orders",
        component: () => import("@/views/merchant/orders/index.vue"),
        meta: {
          title: "订单管理",
          icon: "icon-order",
        },
      },
      {
        path: "menu",
        component: () => import("@/views/merchant/menu/index.vue"),
        meta: {
          title: "菜品管理",
          icon: "icon-dish",
        },
      },
      {
        path: "meal",
        component: () => import("@/views/merchant/meal/index.vue"),
        meta: {
          title: "套餐管理",
          icon: "icon-setmeal",
        },
      },
      {
        path: "menu/add",
        component: () => import("@/views/merchant/menu/addDishType.vue"),
        meta: {
          title: "添加菜品",
          icon: "icon-dish",
        }
      },
      {
        path: "meal/add",
        component: () => import("@/views/merchant/meal/addSetmeal.vue"),
        meta: {
          title: "添加套餐",
          icon: "icon-setmeal",
        }
      },

    ],
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/merchant/404",
    meta: { hidden: true },
  },
];

// const router = createRouter({
//   history: createWebHistory(import.meta.env.BASE_URL),
//   routes,
//   scrollBehavior: () => ({ left: 0, top: 0 }),
// });

export default routes;
