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
      <div class="user">{{ username }}</div>
      <el-button size="small" type="info" plain @click="logout">退出</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";

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

</style>
