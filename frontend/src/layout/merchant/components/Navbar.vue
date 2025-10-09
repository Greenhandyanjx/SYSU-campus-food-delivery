<template>
  <header class="navbar">
    <div class="navbar-left">
      <span class="title">中珠校园外卖 - 商家端</span>
    </div>
    <div class="navbar-right">
      <el-button type="text" @click="goHome">首页</el-button>
      <el-button type="text" @click="logout">退出</el-button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()

const goHome = () => {
  router.push('/merchant/dashboard')
}

const logout = () => {
  router.push('/merchant/login')
}
</script>

<style scoped>
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 56px;
  padding: 0 20px;
  background-color: #409eff;
  color: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.navbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
