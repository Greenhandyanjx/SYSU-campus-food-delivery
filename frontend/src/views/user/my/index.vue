<template>
  <div class="my-page">
    <div class="profile">
      <el-avatar size="64" icon="User" />
      <div class="info">
        <div class="name">{{ username }}</div>
        <div class="sub">普通用户</div>
      </div>
    </div>

    <el-card class="options" shadow="hover">
      <el-row>
        <el-col :span="12"><el-button type="text" @click="go('profile')">资料</el-button></el-col>
        <el-col :span="12"><el-button type="text" @click="go('address')">地址</el-button></el-col>
        <el-col :span="12"><el-button type="text" @click="go('favorites')">收藏</el-button></el-col>
        <el-col :span="12"><el-button type="text" @click="go('settings')">设置</el-button></el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const username = ref(localStorage.getItem('username') || '游客')

function go(key: string) {
  console.log('go', key)
}
</script>

<style scoped>
.my-page { padding: 12px }
.profile { display:flex; gap:12px; align-items:center }
.info .name { font-weight:600 }
.options { margin-top:12px }
</style>
