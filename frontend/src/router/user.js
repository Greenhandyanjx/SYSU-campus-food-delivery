import Layout from "@/layout/user/index.vue";
const routes = [
  {
    path: "/user",
    component: () => import("@/views/login/Login.vue"),
    meta: { title: "商家登录", hidden: true, notNeedAuth: true },
  },
  {
    path: "/user/404",
    component: () => import("@/views/404.vue"),
    meta: { title: "页面不存在", hidden: true, notNeedAuth: true },
  },
  {
    path: "/user",
    component: Layout,
    redirect: "/user/home",
    children: [
      {
        path: "home",
        component: () => import("@/views/user/main/index.vue"),
        name: "userHome",
        meta: {
          title: "工作台",
          icon: "home",
          affix: true,
        },
      },
      { path: 'store/:name', component: () => import('@/views/user/store/index.vue'), name: 'userStore' },
      { path: 'orderlist', component: () => import('@/views/user/orderlist/index.vue') },
      { path: 'cart', component: () => import('@/views/user/cart/index.vue') },
      { path: 'my', component: () => import('@/views/user/my/index.vue') },
    ]
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
// export default [
//   { path: '/user/home', component: () => import('@/views/user/main/index.vue') },
//   { path: '/user/orderlist', component: () => import('@/views/user/orderlist/index.vue') },
//   { path: '/user/cart', component: () => import('@/views/user/cart/index.vue') },
//   { path: '/user/my', component: () => import('@/views/user/my/index.vue') },
// ]