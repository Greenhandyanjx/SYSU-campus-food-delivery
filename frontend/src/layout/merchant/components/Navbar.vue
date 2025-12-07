<template>
  <header class="navbar">
    <!-- 左侧标题 -->
    <div class="navbar-left">
      <span class="title">中珠校园外卖 - 商家端</span>
    </div>

    <!-- 中间菜单 -->
    <div class="navbar-center">
      <el-menu
        mode="horizontal"
        :ellipsis="false"
        :default-active="activePath"
        background-color="transparent"
        text-color="#fff"
        active-text-color="#ffd04b"
        @select="handleSelect"
      >
        <el-menu-item index="/merchant/dashboard">工作台</el-menu-item>
        <el-menu-item index="/merchant/orders">订单管理</el-menu-item>
        <el-menu-item index="/merchant/menu">菜品管理</el-menu-item>
        <el-menu-item index="/merchant/statistics">数据统计</el-menu-item>
        <el-menu-item index="/merchant/meal">套餐管理</el-menu-item>
      </el-menu>
    </div>

    <!-- 右侧操作区 -->
    <div class="navbar-right">
      <el-dropdown trigger="click" @command="handleCommand">
            <span class="el-dropdown-link" style="display:flex;align-items:center;gap:8px;cursor:pointer">
              <el-avatar :size="32" :src="logoSrc" />
              <span class="username">{{ username || '用户' }}</span>
            </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">个人信息</el-dropdown-item>
            <el-dropdown-item command="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const activePath = ref(route.path)

watch(
  () => route.path,
  (newPath) => {
    activePath.value = newPath
  }
)

const handleSelect = (path: string) => {
  router.push(path)
}

const username = ref(localStorage.getItem('username') || '')
import { onMounted } from 'vue'
import { getMerchantProfile } from '@/api/merchant/profile'
const logoSrc = ref('/src/assets/merchant.svg')

onMounted(async () => {
  try {
    const r: any = await getMerchantProfile()
    if (r && r.data && r.data.data) {
      const d = r.data.data
      logoSrc.value = d.logo || d.logoUrl || '/src/assets/merchant.svg'
    }
  } catch (e) {}
})

const handleCommand = (command: string) => {
  if (command === 'logout') {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/merchant/profile')
  }
}

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  router.push('/login')
}
</script>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 30px;
  background-color: #409eff;
  color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.navbar-left .title {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}
.navbar-center {
  flex: 1;
  display: flex;
  justify-content: center;
  overflow: visible !important;
  white-space: nowrap;
  min-width: 0; 
}


.navbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: 30px;
}

.navbar-right .el-button {
  font-size: 15px;
  font-weight: 500;
  color: #fff !important;
  background-color: transparent !important;
  border: none;
  border-radius: 0;
  height: 56px;
  padding: 0 20px;
  transition: background-color 0.3s ease;
}
.navbar-right .el-button:hover {
  background-color: rgba(255, 255, 255, 0.15) !important;
  color: #fff !important;
}


::v-deep(.el-menu.el-menu--horizontal) {
  flex-shrink: 0 !important;
  border-bottom: none;
  background-color: transparent !important;
  overflow: visible !important;
}

::v-deep(.el-menu-item) {
  font-size: 15px;
  font-weight: 500;
  padding: 0 22px !important;
  overflow: visible !important;
  text-overflow: unset !important;
}

/* 鼠标悬停颜色 */
::v-deep(.el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.2) !important;
}
</style>
