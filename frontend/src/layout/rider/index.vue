<template>
  <div class="rider-layout">
    <div class="main">
      <Navbar />
      <div class="content">
        <div class="page-center-wrapper">
          <router-view />
        </div>
      </div>
    </div>
    <!-- 骑手聊天组件 -->
    <RiderChat />
  </div>
</template>

<script setup lang="ts">
import Navbar from "./components/Navbar.vue";
import RiderChat from "@/components/rider/RiderChat.vue";
// 不自动启动定位追踪，改为手动控制
// import useRiderLocation from "@/composables/useRiderLocation";
// const { isTracking, locationError } = useRiderLocation();
</script>

<style scoped lang="scss">
.main { flex: 1; display:flex; flex-direction:column; }

.rider-layout {
  height: 100vh;
  background: var(--rider-bg);
}

.content {
  flex: 1;
  padding: 18px 22px;
  overflow-y: auto;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.page-center-wrapper {
  width: 100%;
  max-width: 1180px;
  background: transparent;
  padding: 0;
  min-height: calc(100vh - 110px);
  box-shadow: none;
  border-radius: 0;
}

</style>
