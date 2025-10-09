import Vue from "vue";
import Router from "vue-router";
import Layout from "@/layout/merchant/components/Navbar.vue";

Vue.use(Router);

const router = new Router({
  scrollBehavior: (to, from, savedPosition) => {
    if (savedPosition) {
      return savedPosition;
    }
    return { x: 0, y: 0 };
  },
  base: process.env.BASE_URL,
  routes: [
    {
      path: "/merchant/login",
      component: () => import("@/views/login/Login.vue"),
      meta: { title: "商家登录", hidden: true, notNeedAuth: true }
    },
    // 404页
    {
      path: "/merchant/404",
      component: () => import("@/views/404.vue"),
      meta: { title: "页面不存在", hidden: true, notNeedAuth: true }
    },

    // 主体布局页
    {
      path: "/merchant",
      component: Layout,
      redirect: "/merchant/dashboard",
      children: [
        {
          path: "dashboard",
          component: () => import("@/views/merchant/Dashboard.vue"),
          name: "MerchantDashboard",
          meta: {
            title: "工作台",
            icon: "dashboard",
            affix: true
          }
        },
        {
          path: "statistics",
          component: () => import("@/views/merchant/Statistics.vue"),
          meta: {
            title: "数据统计",
            icon: "icon-statistics"
          }
        },
        {
          path: "orders",
          component: () => import("@/views/merchant/Orders.vue"),
          meta: {
            title: "订单管理",
            icon: "icon-order"
          }
        },
        {
          path: "menu",
          component: () => import("@/views/merchant/Menu.vue"),
          meta: {
            title: "菜品管理",
            icon: "icon-dish"
          }
        }
      ]
    },

    // 未匹配路由跳404
    {
      path: "*",
      redirect: "/merchant/404",
      meta: { hidden: true }
    }
  ]
});

export default router;
