<template>
  <div class="cart-page">
    <h2>购物车</h2>
    <el-table :data="cartItems" style="width:100%">
      <el-table-column prop="name" label="商品" />
      <el-table-column prop="price" label="单价" />
      <el-table-column prop="qty" label="数量" />
      <el-table-column label="小计">
        <template #default="{ row }">
          ¥{{ (row.price * row.qty).toFixed(2) }}
        </template>
      </el-table-column>
    </el-table>
    <div class="cart-footer">
      <div>合计: ¥{{ total }}</div>
      <el-button type="primary">去结算</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

const cartItems = ref([
  { name: '鸡腿饭', price: 22.5, qty: 1 },
  { name: '可乐', price: 5.0, qty: 2 },
])

const total = computed(() => cartItems.value.reduce((s, i) => s + i.price * i.qty, 0).toFixed(2))
</script>

<style scoped>
.cart-page { padding: 12px }
.cart-footer { display:flex; justify-content:space-between; align-items:center; margin-top:12px }
</style>
