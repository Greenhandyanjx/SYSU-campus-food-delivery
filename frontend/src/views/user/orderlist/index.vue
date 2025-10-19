<template>
  <div class="order-list-page">
    <h2>我的订单</h2>
    <el-timeline>
      <el-timeline-item v-for="(o, i) in orders" :key="i" :timestamp="o.time">
        <div class="order-item">
          <div class="title">订单号: {{ o.id }} - 状态: {{ o.status }}</div>
          <div>骑手位置: {{ o.riderPos || '未定位' }}</div>
          <div>取餐码: <strong>{{ o.pickupCode || '—' }}</strong></div>
        </div>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const orders = ref([
  { id: 'ORD20251019001', status: '配送中', time: '2025-10-19 12:45', riderPos: '1.2km', pickupCode: 'A123' },
  { id: 'ORD20251019002', status: '已完成', time: '2025-10-18 18:20', riderPos: '', pickupCode: 'B456' },
])
</script>

<style scoped>
.order-list-page { padding: 12px }
.order-item { padding: 8px 0 }
</style>
