<template>
  <el-menu router :default-active="$route.path" class="sidebar">
    <el-menu-item index="/merchant/dashboard">工作台</el-menu-item>
    <el-menu-item index="/merchant/orders">订单管理</el-menu-item>
    <el-menu-item index="/merchant/menu">菜品管理</el-menu-item>
    <el-menu-item index="/merchant/statistics">数据统计</el-menu-item>
  </el-menu>
</template>

<style scoped>
.sidebar {
  width: 200px;
  height: 100vh;
  border-right: 1px solid #eee;
}
</style>