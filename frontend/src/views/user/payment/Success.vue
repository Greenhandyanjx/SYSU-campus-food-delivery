<template>
  <div style="padding:40px;text-align:center;">
    <h2>支付成功 🎉</h2>
    <p>感谢您的订单，我们已收到付款，会尽快为您处理配送。</p>
    <div style="margin-top:20px;">
      <el-button type="primary" @click="goHome">回到首页</el-button>
      <el-button style="margin-left:8px;" @click="goOrders">查看订单</el-button>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
function goHome() { router.push('/user/home') }
function goOrders() { router.push('/user/orderlist') }
</script>
