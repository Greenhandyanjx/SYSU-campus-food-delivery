<template>
  <div class="order-detail-page">
    <div class="header">
      <el-button type="text" @click="$router.back()">← 返回</el-button>
      <h2>订单详情 - {{ order?.id || id }}</h2>
    </div>

    <div v-if="order" class="detail-card">
      <div class="section">
        <h3>基本信息</h3>
        <p>店铺：<strong>{{ order.storeName }}</strong></p>
        <p>状态：<strong>{{ order.statusText }}</strong></p>
        <p>下单时间：{{ order.time }}</p>
      </div>

      <div class="section">
        <h3>商品清单</h3>
        <ul class="goods">
          <li v-for="(it, i) in order.items" :key="i">
            <span class="gname">{{ it.name }}</span>
            <span class="gcount">x{{ it.count }}</span>
            <span class="gprice">¥{{ (it.price * it.count).toFixed(2) }}</span>
          </li>
        </ul>
        <div class="total">合计：<strong>¥{{ totalPrice.toFixed(2) }}</strong></div>
      </div>

      <div class="section">
        <h3>物流信息</h3>
        <p>配送状态：{{ order.statusText }}</p>
        <p>预计到达：大约 20-40 分钟（示例）</p>
      </div>

      <div class="section">
        <h3>发票 / 支付</h3>
        <p>支付方式：在线支付（示例）</p>
        <p>发票信息：暂无（示例）</p>
      </div>
    </div>

    <div v-else class="empty">未能加载订单详情</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import orderApi from '@/api/user/order'

const route = useRoute()
const id = route.params.id
const order = ref(null)

const totalPrice = computed(() => (order.value?.items || []).reduce((s, it) => s + (it.price * it.count), 0))

async function fetch() {
  try {
    // 尝试调用后端接口，若失败则使用前端 mock（兼容当前项目）
    const res = await orderApi.getOrderDetail(id)
    order.value = res?.data || res || null
  } catch (e) {
    // fallback: 从 window 全局查找 mock（orderlist 页面使用 rawOrders）
    try { order.value = window.__RAW_ORDERS__?.find(o=>o.id===id) || null } catch(e){ order.value = null }
  }
}

onMounted(() => { fetch() })
</script>

<style scoped>
.order-detail-page { padding: 12px }
.header { display:flex; align-items:center; gap:12px }
.detail-card { background:#fff; border-radius:8px; padding:16px; margin-top:12px }
.section { border-bottom:1px dashed #eee; padding:12px 0 }
.goods { list-style:none; padding:0; margin:0 }
.goods li { display:flex; justify-content:space-between; padding:6px 0 }
.total { text-align:right; margin-top:8px; font-size:16px }
.empty { padding:40px; color:#999 }
</style>
