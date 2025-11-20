<template>
  <div class="order-card">
    <!-- 头部：店铺信息 -->
    <div class="card-header">
      <div class="store" @click="$emit('open-store', order.storeId)">
        <img v-if="order.storeLogo" :src="order.storeLogo" alt="logo" @error="onImgError" />
        <span class="store-name">{{ order.storeName }}</span>
        <i class="arrow el-icon-arrow-right"></i>
      </div>
      <div class="status">{{ order.statusText }}</div>
    </div>

    <!-- 内容：商品清单 -->
    <div class="card-body" @click="$emit('view', order)">
      <div class="items">
        <div class="item" v-for="(it, idx) in order.items" :key="idx">
          <img :src="itemImage(it)" alt="item" @error="onImgError" />
          <div class="meta">
            <div class="name">{{ it.name }}</div>
            <div class="count">x{{ it.count }}</div>
          </div>
          <div class="price">¥{{ (it.price * it.count).toFixed(2) }}</div>
        </div>
      </div>
      <div class="summary">
        共 {{ totalCount }} 件商品，
        合计：<span class="total">¥{{ totalPrice.toFixed(2) }}</span>
      </div>
    </div>

    <!-- 底部：操作按钮 -->
    <div class="card-footer">
      <div class="left-meta" style="display: flex; align-items: center; gap: 16px;">下单时间：{{ order.time }}
      <ChatLauncher :merchant-id="order.storeId || order.merchantId" :merchant-name="order.storeName" />
      </div>

      <div class="actions">
        
        <template v-if="order.status === 'pending'">
          <div class="countdown">剩余支付时间 <strong>{{ countdown }}</strong></div>
          <el-button type="warning" round size="default" @click.stop="$emit('pay', order)">去付款</el-button>
          <el-button round plain size="default" @click.stop="$emit('cancel', order)">取消订单</el-button>
        </template>

        <template v-else-if="order.status === 'shipping'">
          <el-button type="primary" round size="default" @click.stop="$emit('confirm', order)">确认收货</el-button>
          <el-button round plain size="default" @click.stop="$emit('reorder', order)">再次购买</el-button>
        </template>

        <template v-else-if="order.status === 'completed'">
          <el-button type="primary" round size="default" @click.stop="$emit('review', order)">去评价</el-button>
          <el-button round plain size="default" @click.stop="$emit('reorder', order)">再次购买</el-button>
        </template>

        <template v-else-if="order.status === 'refund'">
          <el-tag type="info" effect="plain">退款中</el-tag>
          <el-button round plain size="default" @click.stop="$emit('view-refund', order)">查看详情</el-button>
        </template>

        <template v-else>
          <el-button round plain size="default" @click.stop="$emit('view', order)">查看订单</el-button>
          <!-- 聊天入口，封装为组件 -->

        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import ChatLauncher from '@/components/Chat/ChatLauncher.vue'
const props = defineProps({ order: { type: Object, required: true } })
const emit = defineEmits(['pay','cancel','confirm','reorder','review','view','view-refund','open-store','auto-cancel'])

const totalPrice = computed(() => (props.order.items || []).reduce((s, it) => s + (it.price * it.count), 0))
const totalCount = computed(() => (props.order.items || []).reduce((s, it) => s + it.count, 0))

// 倒计时逻辑
const countdown = ref('00:00:00')
let timer = null
function formatDiff(diffMs) {
  if (diffMs <= 0) return '00:00:00'
  const s = Math.floor(diffMs / 1000)
  const hh = String(Math.floor(s / 3600)).padStart(2, '0')
  const mm = String(Math.floor((s % 3600) / 60)).padStart(2, '0')
  const ss = String(s % 60).padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}
function updateCountdown() {
  if (props.order.status !== 'pending') return
  const end = new Date(props.order.payDeadline).getTime()
  const now = Date.now()
  const diff = end - now
  countdown.value = formatDiff(diff)
  if (diff <= 0) {
    emit('auto-cancel', props.order)
    clearInterval(timer)
    timer = null
  }
}
onMounted(() => {
  if (props.order.status === 'pending') {
    updateCountdown()
    timer = setInterval(updateCountdown, 1000)
  }
})
onBeforeUnmount(() => timer && clearInterval(timer))

function onImgError(e) {
  try { e.target && (e.target.src = '/src/assets/noImg.png') } catch (err) {}
}

function itemImage(it) {
  // support multiple possible fields from backend: img, image, picture, pic, thumb, imgUrl
  return (
    it.img || it.image || it.picture || it.pic || it.thumb || it.imgUrl || '/src/assets/noImg.png'
  )
}
</script>

<style scoped>
.order-card {
  border-radius: 14px;
  margin: 16px 0;
  background: #fff;
  box-shadow: 0 2px 10px rgba(0,0,0,0.05);
  overflow: hidden;
  transition: all 0.2s ease;
}
.order-card:hover { box-shadow: 0 4px 16px rgba(0,0,0,0.08); transform: translateY(-2px); }

/* Header */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  border-bottom: 1px solid #f4f4f4;
  background: #fffceb;
}
.store {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.store img {
  width: 38px;
  height: 38px;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid #f5f5f5;
}
.store-name {
  font-weight: 600;
  color: #333;
  font-size: 15px;
}
.arrow {
  color: #bbb;
  font-size: 16px;
}
.status {
  color: #ff8c00;
  font-weight: 700;
  font-size: 14px;
}

/* Body */
.card-body {
  padding: 14px 18px;
}
.items {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.item {
  display: flex;
  align-items: center;
  gap: 12px;
}
.item img {
  width: 70px;
  height: 70px;
  border-radius: 8px;
  object-fit: cover;
}
.meta {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.name {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}
.count {
  color: #999;
}
.price {
  color: #ff7b00;
  font-weight: 700;
}
.summary {
  text-align: right;
  color: #666;
  margin-top: 8px;
  font-size: 13px;
}
.total {
  color: #ff6b00;
  font-size: 16px;
  font-weight: 700;
  margin-left: 4px;
}

/* Footer */
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 18px;
  border-top: 1px solid #f2f2f2;
  background: #fafafa;
}
.left-meta {
  font-size: 13px;
  color: #999;
}
.actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.countdown {
  color: #ff9800;
  font-size: 13px;
  margin-right: 6px;
}

/* 按钮风格强化 */
.el-button--warning {
  background: linear-gradient(90deg, #ffce00, #ffb300);
  color: #333;
  border: none;
}
.el-button--primary {
  background: linear-gradient(90deg, #ffb300, #ff8c00);
  border: none;
  color: #fff;
}
.el-button--plain {
  color: #666;
  border-color: #ddd;
}
.el-button:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}
</style>
