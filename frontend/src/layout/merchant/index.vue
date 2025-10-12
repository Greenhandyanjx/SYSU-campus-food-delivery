<template>
  <div class="merchant-layout">
    <!-- 左侧侧边栏有点难看，舍弃 -->
    <!-- <Sidebar class="sidebar" /> -->

    <!-- 右侧主体内容 -->
    <div class="main-container">
      <Navbar />
      <div class="content">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script>
import Navbar from "./components/Navbar.vue";
// import Sidebar from "./components/Sidebar.vue"; // 假如你有侧边栏

export default {
  name: "MerchantLayout",
  components: { Navbar}
};
</script>

<style lang="scss" scoped>
/* 你的原样式可以保留 */
.merchant-layout {
  display: flex;
  height: 100vh;
}

/* 左侧栏 */
.sidebar {
  width: 200px;
  background-color: #2c3e50;
  color: white;
  min-height: 100vh;
}

/* 主内容区域 */
.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #f4f6f9;
}

/* 顶部导航栏 */
.main-container > nav {
  height: 60px;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

/* 路由内容区域 */
.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
.app-wrapper {
  position: relative;
  height: 100%;
  width: 100%;
  min-width: 1366px;
  overflow-x: auto;
  overflow-y: hidden;
}

// .main-container {
//   height: 100%;
//   background: #f3f4f7;
//   position: relative;
//   width: calc(100% - 190px);
// }

.drawer-bg {
  background: #000;
  opacity: 0.3;
  width: 100%;
  top: 0;
  height: 100%;
  position: absolute;
  z-index: 999;
}

.sidebar-container {
  transition: width 0.28s;
  width: 190px !important;
  height: 100%;
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  z-index: 1001;
  overflow: hidden;
}

.hideSidebar .main-container {
  margin-left: 80px;
  width: calc(100% - 80px);
}

.hideSidebar .sidebar-container {
  width: 80px !important;
}

.mobile .main-container {
  margin-left: 0px;
}

.mobile .sidebar-container {
  transition: transform 0.28s;
  width: 190px !important;
}

.mobile.openSidebar {
  position: fixed;
  top: 0;
}

// .mobile.hideSidebar .sidebar-container {
//   pointer-events: none;
//   transition-duration: 0.3s;
//   transform: translate3d(-190px, 0, 0);
// }

.withoutAnimation .main-container,
.withoutAnimation .sidebar-container {
  transition: none;
}
</style>
<!-- <script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'
import AppMain from './components/AppMain.vue'

/* Sidebar 状态管理（你可以先用 ref 简单实现） */
const sidebarOpened = ref(true)
const withoutAnimation = ref(false)
const device = ref<'desktop' | 'mobile'>('desktop')

/* 响应式切换设备类型 */
const checkDevice = () => {
  if (window.innerWidth < 992) {
    device.value = 'mobile'
  } else {
    device.value = 'desktop'
  }
}

/* 监听页面宽度变化 */
onMounted(() => {
  checkDevice()
  window.addEventListener('resize', checkDevice)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', checkDevice)
})

/* 动态 class 绑定 */
const classObj = computed(() => ({
  hideSidebar: !sidebarOpened.value,
  openSidebar: sidebarOpened.value,
  withoutAnimation: withoutAnimation.value,
  mobile: device.value === 'mobile'
}))

/* 点击遮罩层关闭侧边栏（移动端） */
const handleClickOutside = () => {
  sidebarOpened.value = false
}
</script> -->

