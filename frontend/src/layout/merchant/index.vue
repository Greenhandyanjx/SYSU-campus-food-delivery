<template>
  <div :class="classObj" class="app-wrapper">
    <div
      v-if="classObj.mobile && sidebarOpened"
      class="drawer-bg"
      @click="handleClickOutside"
    />
    <Sidebar class="sidebar-container" />
    <div class="main-container">
      <Navbar />
      <main>
        <router-view />
      </main>
      <AppMain />
    </div>
  </div>
</template>

<script setup lang="ts">
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
</script>

<style lang="scss" scoped>
/* 你的原样式可以保留 */
.app-wrapper {
  position: relative;
  height: 100%;
  width: 100%;
  min-width: 1366px;
  overflow-x: auto;
  overflow-y: hidden;
}

.main-container {
  height: 100%;
  background: #f3f4f7;
  position: relative;
  width: calc(100% - 190px);
}

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

.mobile.hideSidebar .sidebar-container {
  pointer-events: none;
  transition-duration: 0.3s;
  transform: translate3d(-190px, 0, 0);
}

.withoutAnimation .main-container,
.withoutAnimation .sidebar-container {
  transition: none;
}
</style>
