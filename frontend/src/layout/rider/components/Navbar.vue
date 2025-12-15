<template>
  <div class="nav">
    <div class="left">
      <div class="brand">中珠校园外卖 - 骑手端</div>

      <div class="tabs">
        <div class="tab" :class="{ active: isActive('/rider/dashboard') }" @click="go('/rider/dashboard')">工作台</div>
        <div class="tab" :class="{ active: isActive('/rider/new') }" @click="go('/rider/new')">待接单</div>
        <div class="tab" :class="{ active: isActive('/rider/ongoing') }" @click="go('/rider/ongoing')">进行中</div>
        <div class="tab" :class="{ active: isActive('/rider/history') }" @click="go('/rider/history')">历史订单</div>
        <div class="tab" :class="{ active: isActive('/rider/me') }" @click="go('/rider/me')">我的</div>
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
  height: 56px;
  background: #fff;
  border-bottom: 1px solid var(--rider-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 18px;
}

.brand {
  font-weight: 900;
  color: var(--rider-primary);
  letter-spacing: .5px;
  white-space: nowrap;
}

.tabs {
  display: flex;
  gap: 18px;
  margin-left: 18px;
}

.tab {
  position: relative;
  padding: 10px 2px;
  cursor: pointer;
  color: #606266;
  font-weight: 600;
  user-select: none;
}

.tab:hover {
  color: var(--rider-primary);
}

.tab.active {
  color: var(--rider-primary);
}

.tab.active::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 2px;
  height: 3px;
  border-radius: 99px;
  background: var(--rider-primary);
}

.right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user {
  color: #606266;
  font-size: 12px;
  background: #f2f6ff;
  border: 1px solid #dbe7ff;
  padding: 6px 10px;
  border-radius: 999px;
}

</style>
