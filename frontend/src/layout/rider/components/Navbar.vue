<template>
  <div class="nav">
    <div class="left">
      <div class="brand">中珠校园外卖 - 骑手端</div>

      <div class="tabs">
        <div class="tab" :class="{ active: isActive('/rider/dashboard') }" @click="go('/rider/dashboard')">
          <i class="iconfont icon-dashboard"></i>
          <span>工作台</span>
        </div>
        <div class="tab" :class="{ active: isActive('/rider/new') }" @click="go('/rider/new')">
          <i class="iconfont icon-notification"></i>
          <span>待接单</span>
        </div>
        <div class="tab" :class="{ active: isActive('/rider/ongoing') }" @click="go('/rider/ongoing')">
          <i class="iconfont icon-truck"></i>
          <span>进行中</span>
        </div>
        <div class="tab" :class="{ active: isActive('/rider/history') }" @click="go('/rider/history')">
          <i class="iconfont icon-history"></i>
          <span>历史订单</span>
        </div>
        <div class="tab" :class="{ active: isActive('/rider/me') }" @click="go('/rider/me')">
          <i class="iconfont icon-user"></i>
          <span>我的</span>
        </div>
      </div>
    </div>

    <div class="right">
      <!-- 定位状态指示器 -->
      <div class="location-status" :class="{ 'tracking': isTracking, 'error': hasLocationError }" @click="handleLocationClick">
        <i class="iconfont" :class="locationIcon"></i>
        <span class="location-text">{{ locationText }}</span>
      </div>

      <div class="user">{{ username }}</div>
      <el-button size="small" type="info" plain @click="logout">退出</el-button>
    </div>
  </div>

  <!-- 定位权限提示 -->
  <LocationPermissionPrompt
    :error="hasLocationError ? locationError : null"
    @dismiss="handlePermissionPromptDismiss"
    ref="permissionPromptRef"
  />

  <!-- 地图弹窗 -->
  <AmapModal
    v-model="showMapModal"
    :title="mapModalTitle"
    :address="mapModalAddress"
    :default-location="currentLocation"
    :show-rider-location="true"
    :rider-location="{
      lng: currentLocation[0],
      lat: currentLocation[1]
    }"
  />
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import AmapModal from "@/components/AmapModal.vue";
import LocationPermissionPrompt from "@/components/LocationPermissionPrompt.vue";
import locationTracker from "@/utils/locationTracker";

const router = useRouter();
const route = useRoute();

const username = computed(() => localStorage.getItem("username") || "rider");

const isActive = (path: string) => route.path === path;

const go = (path: string) => {
  router.push(path);
};

const logout = () => {
  localStorage.removeItem("token");
  localStorage.removeItem("role");
  localStorage.removeItem("username");
  router.push("/login");
};

// 权限提示相关
const permissionPromptRef = ref();

const handlePermissionPromptDismiss = () => {
  console.log('Permission prompt dismissed');
};

// 定位状态相关
const isTracking = ref(false);
const hasLocationError = ref(false);
const currentLocation = ref<[number, number]>([113.299, 23.099]); // 默认中山大学位置

const locationIcon = computed(() => {
  if (hasLocationError.value) return "icon-location-error";
  if (isTracking.value) return "icon-location-on";
  return "icon-location-off";
});

const locationText = computed(() => {
  if (hasLocationError.value) return "定位失败(点重试)";
  if (isTracking.value) return "已定位";
  return "未定位";
});

// 地图弹窗相关
const showMapModal = ref(false);
const mapModalTitle = ref('');
const mapModalAddress = ref('');

// 监听定位状态变化（通过全局事件）
const setupLocationListener = () => {
  // 监听定位状态
  window.addEventListener('rider:locationStatus', (event: any) => {
    console.log('📍 [骑手定位] 收到定位状态事件:', event.detail);
    isTracking.value = event.detail.isTracking;
    hasLocationError.value = !!event.detail.error;

    // 更新当前位置
    if (event.detail.position && event.detail.position.latitude && event.detail.position.longitude) {
      currentLocation.value = [event.detail.position.longitude, event.detail.position.latitude];
      console.log('📍 [骑手定位] 更新当前位置:', currentLocation.value);
    }
  });

  // 监听位置更新
  window.addEventListener('rider:locationUpdate', (event: any) => {
    console.log('📍 [骑手定位] 收到位置更新事件:', event.detail);
    if (event.detail.position && event.detail.position.latitude && event.detail.position.longitude) {
      currentLocation.value = [event.detail.position.longitude, event.detail.position.latitude];
      console.log('📍 [骑手定位] 位置更新后:', currentLocation.value);
    }
  });
};

// 处理定位点击
const handleLocationClick = async () => {
  // 如果定位失败，尝试重新启动定位
  if (hasLocationError.value) {
    console.log('🔄 [骑手定位] 用户点击重试，重新启动定位追踪');
    const success = await locationTracker.startTracking();
    if (success) {
      ElMessage.success('✅ 定位重启成功');
    } else {
      ElMessage.error('❌ 定位重启失败，请检查浏览器权限设置');
    }
  }

  // 永远弹出地图弹窗
  showMapModal.value = true;
  mapModalTitle.value = '骑手当前位置';

  if (hasLocationError.value) {
    mapModalAddress.value = '定位失败，无法获取具体地址';
  } else if (isTracking.value) {
    const currentPosition = locationTracker.getCurrentPosition();
    if (currentPosition) {
      mapModalAddress.value = `定位成功 (精度: ${Math.round(currentPosition.accuracy || 0)}m)`;
    } else {
      mapModalAddress.value = '正在获取地址信息...';
    }
  } else {
    mapModalAddress.value = '暂无位置信息';
  }
};

// 启动定位追踪
const startLocationTracking = async () => {
  console.log('🚀 [骑手定位] 开始启动定位追踪');

  try {
    const success = await locationTracker.startTracking();
    if (success) {
      console.log('✅ [骑手定位] 定位追踪启动成功');
      ElMessage.success('✅ 定位服务已启动');
    } else {
      console.warn('⚠️ [骑手定位] 定位追踪启动失败');
      ElMessage.warning('⚠️ 定位服务启动失败，部分功能可能受限');
    }
  } catch (error) {
    console.error('❌ [骑手定位] 启动定位追踪时发生错误:', error);
    ElMessage.error('❌ 定位服务异常');
  }
};

// 组件挂载时启动定位追踪
onMounted(async () => {
  console.log('🔧 [骑手定位] Navbar组件已挂载，启动定位追踪');
  setupLocationListener();

  // 等待一小段时间确保事件监听器设置完成
  setTimeout(async () => {
    await startLocationTracking();
  }, 100);
});

// 组件卸载时清理定位追踪
onUnmounted(() => {
  console.log('🧹 [骑手定位] Navbar组件将卸载，保持定位追踪运行（其他页面可能需要）');
  // 注意：这里不停止定位追踪，因为用户可能切换到其他骑手页面
  // 如果需要完全停止，可以在用户退出登录时调用
});
</script>

<style scoped lang="scss">
.nav {
  height: 60px;
  background: linear-gradient(to right, #FFB302, #FFC200);
  border-bottom: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(255, 179, 2, 0.2);
}

.brand {
  font-weight: 900;
  color: #fff;
  letter-spacing: .5px;
  white-space: nowrap;
  font-size: 18px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.tabs {
  display: flex;
  gap: 8px;
  margin-left: 32px;
}

.tab {
  position: relative;
  padding: 8px 16px;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.9);
  font-weight: 600;
  user-select: none;
  border-radius: 20px;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s ease;
  font-size: 14px;
}

.tab i {
  font-size: 16px;
}

.tab:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.15);
}

.tab.active {
  color: var(--rider-primary);
  background: #fff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.tab.active::after {
  display: none;
}

.right {
  display: flex;
  align-items: center;
  gap: 12px;
}

// 定位状态指示器
.location-status {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
  font-weight: 600;

  &:hover {
    background: rgba(255, 255, 255, 0.25);
    border-color: rgba(255, 255, 255, 0.5);
  }

  &.tracking {
    background: rgba(103, 194, 58, 0.2);
    border-color: rgba(103, 194, 58, 0.4);
    color: #fff;

    &:hover {
      background: rgba(103, 194, 58, 0.3);
      border-color: rgba(103, 194, 58, 0.6);
    }
  }

  &.error {
    background: rgba(245, 108, 108, 0.2);
    border-color: rgba(245, 108, 108, 0.4);
    color: #fff;
    animation: pulse 2s infinite;

    &:hover {
      background: rgba(245, 108, 108, 0.3);
      border-color: rgba(245, 108, 108, 0.6);
    }
  }

  .iconfont {
    font-size: 14px;
  }

  .location-text {
    font-size: 12px;
    white-space: nowrap;
  }
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.6; }
  100% { opacity: 1; }
}

.user {
  color: var(--rider-primary);
  font-size: 13px;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 8px 16px;
  border-radius: 20px;
  font-weight: 600;
}

:deep(.el-button--small) {
  border-radius: 20px;
  font-weight: 600;
  border-color: rgba(255, 255, 255, 0.5);
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

:deep(.el-button--small:hover) {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.7);
}

/* Icon font styles */
.iconfont {
  font-family: "iconfont" !important;
  font-size: 16px;
  font-style: normal;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

// Define icons if iconfont is not available
.icon-dashboard:before { content: "📊"; }
.icon-notification:before { content: "🔔"; }
.icon-truck:before { content: "🚚"; }
.icon-history:before { content: "📋"; }
.icon-user:before { content: "👤"; }

// 定位相关图标
.icon-location-off:before { content: "📍"; }
.icon-location-on:before { content: "🟢"; }
.icon-location-error:before { content: "🔴"; }

</style>
