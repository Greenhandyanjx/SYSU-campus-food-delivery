<template>
  <div class="order-card">
    <!-- 头部：店铺 + 状态 -->
    <div class="card-header">
      <div class="store" @click="$emit('open-store', order.storeId)">
        <img
          v-if="order.storeLogo"
          :src="order.storeLogo"
          alt="店铺logo"
          @error="onImgError"
        />
        <span class="store-name">{{ order.storeName }}</span>
      </div>
      <div class="status">{{ order.statusText }}</div>
    </div>

    <!-- 商品清单 -->
    <div class="card-body" @click="$emit('view', order)">
      <div class="items">
        <div class="item" v-for="(it, i) in order.items" :key="i">
          <img :src="itemImage(it)" @error="onImgError" />
          <div class="meta">
            <div class="name">{{ it.name }}</div>
            <div class="count">x{{ it.count }}</div>
          </div>
          <div class="price">¥{{ (it.price * it.count).toFixed(2) }}</div>
        </div>
      </div>

      <div class="summary">
        共 {{ totalCount }} 件商品，
        合计 <span class="total">¥{{ totalPrice.toFixed(2) }}</span>
        <span
          v-if="deliveryFee != null"
          class="fee-tip"
        >(含配送费 ¥{{ deliveryFee.toFixed(2) }})</span>
      </div>
    </div>

    <!-- 底部操作栏 -->
    <div class="card-footer">
      <div class="left-meta">
        下单时间：{{ formattedTime }}
        <ChatLauncher
          :merchant-id="order.storeId || order.merchantId"
          :merchant-name="order.storeName"
        />
      </div>

      <div class="actions">
        <!-- 待付款 -->
        <template v-if="status === 1">
          <div class="countdown">剩余支付时间 <strong>{{ countdown }}</strong></div>
          <el-button type="warning" round @click.stop="$emit('pay', order)">去付款</el-button>
          <el-button plain round @click.stop="$emit('cancel', order)">取消订单</el-button>
        </template>

        <!-- 待接单（可取消） -->
        <template v-else-if="status === 2">
          <span class="tip">等待商家接单</span>
          <el-button plain round @click.stop="$emit('cancel', order)">取消订单</el-button>
          <el-button plain round @click.stop="$emit('reorder', order)">再次购买</el-button>
        </template>

        <!-- 待派送（仅再次购买） -->
        <template v-else-if="status === 3">
          <span class="tip">正在分配骑手</span>
          <el-button plain round @click.stop="$emit('reorder', order)">再次购买</el-button>
        </template>

        <!-- 派送中 -->
        <template v-else-if="status === 4">
          <el-button type="primary" round @click.stop="$emit('contact-rider', order)">联系骑手</el-button>
          <el-button plain round @click.stop="$emit('reorder', order)">再次购买</el-button>
          <el-button type="primary" round @click.stop="$emit('confirm', order)">确认收货</el-button>
        </template>

        <!-- 已完成 -->
        <template v-else-if="status === 5">
          <el-button type="primary" round @click.stop="$emit('review', order)">去评价</el-button>
          <el-button plain round @click.stop="$emit('reorder', order)">再次购买</el-button>
        </template>

        <!-- 退款中 -->
        <template v-else-if="status === 7 || order.status === 'refund'">
          <el-tag type="info" effect="plain">退款中</el-tag>
          <el-button plain round @click.stop="$emit('view-refund', order)">查看详情</el-button>
        </template>

        <!-- 其他状态（已取消等） -->
        <template v-else>
          <el-button plain round @click.stop="$emit('view', order)">查看订单</el-button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import ChatLauncher from '@/components/Chat/ChatLauncher.vue'
import noImg from '@/assets/noImg.png'
import { safeImage } from '@/utils/asset'

const props = defineProps({
  order: { type: Object, required: true }
})

const emit = defineEmits([
  'pay', 'cancel', 'confirm', 'reorder', 'review',
  'view', 'view-refund', 'open-store', 'contact-rider', 'auto-cancel'
])

// 统一取 status（兼容 string/number）
const status = computed(() => Number(props.order.status))

// 金额相关
const totalPrice = computed(() => {
  const itemsTotal = (props.order.items || []).reduce((s, it) => s + Number(it.price || 0) * Number(it.count || 0), 0)
  const fee = Number(props.order.deliveryFee ?? props.order.delivery_fee ?? props.order.deliveryAmount ?? 0) || 0
  return itemsTotal + fee
})
const totalCount = computed(() =>
  (props.order.items || []).reduce((s, it) => s + Number(it.count || 0), 0)
)
const deliveryFee = computed(() => {
  console.log('Calculating delivery fee for order:', props.order);
  const v = props.order.deliveryFee ?? props.order.delivery_fee ?? props.order.deliveryAmount ?? 0
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
})

// 时间格式化（优先 payDeadline → time → orderTime）
const formattedTime = computed(() => {
  const t = props.order.payDeadline || props.order.time || props.order.orderTime || props.order.createdAt
  if (!t) return ''
  const d = new Date(t)
  if (isNaN(d)) return t
  const M = d.getMonth() + 1
  const D = d.getDate()
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${M}月${D}日 ${hh}:${mm}`
})

// 待付款倒计时
const countdown = ref('00:00:00')
let timer = null

const updateCountdown = () => {
  if (status.value !== 1 || !props.order.payDeadline) {
    countdown.value = ''
    return
  }
  const diff = new Date(props.order.payDeadline).getTime() - Date.now()
  if (diff <= 0) {
    countdown.value = '已超时'
    emit('auto-cancel', props.order)
    clearInterval(timer)
    return
  }
  const s = Math.floor(diff / 1000)
  const h = String(Math.floor(s / 3600)).padStart(2, '0')
  const m = String(Math.floor((s % 3600) / 60)).padStart(2, '0')
  const sec = String(s % 60).padStart(2, '0')
  countdown.value = `${h}:${m}:${sec}`
}

onMounted(() => {
  if (status.value === 1) {
    updateCountdown()
    timer = setInterval(updateCountdown, 1000)
  }
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})

// 图片错误兜底
const onImgError = (e) => {
  if (e && e.target) e.target.src = noImg
}

const itemImage = (it) => safeImage(it.img || it.image || it.picture || it.pic || it.thumb || it.imgUrl || '', noImg)
</script>

<style scoped>
.order-card {
  background: #fff;
  border-radius: 14px;
  margin: 16px 0;
  box-shadow: 0 2px 10px rgba(0,0,0,0.05);
  overflow: hidden;
  transition: all .2s;
}
.order-card:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
  transform: translateY(-2px);
}

/* Header */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  background: #fffceb;
  border-bottom: 1px solid #f4f4f4;
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
}
.store-name {
  font-weight: 600;
  font-size: 15px;
  color: #333;
}
.status {
  color: #ff8c00;
  font-weight: 700;
  font-size: 14px;
}

/* Body */
.card-body {
  padding: 14px 18px;
  cursor: pointer;
}
.items { gap: 10px; }
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
.name { font-size: 14px; color: #333; font-weight: 500; }
.count { color: #999; }
.price { color: #ff7b00; font-weight: 700; }
.summary {
  text-align: right;
  margin-top: 8px;
  font-size: 13px;
  color: #666;
}
.total {
  color: #ff6b00;
  font-size: 16px;
  font-weight: 700;
  margin-left: 4px;
}
.fee-tip { margin-left: 8px; color: #999; }

/* Footer */
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 18px;
  background: #fafafa;
  border-top: 1px solid #f2f2f2;
}
.left-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: #999;
}
.actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.tip { margin-right: 12px; color: #999; font-size: 13px; }
.countdown { color: #ff9800; font-size: 13px; margin-right: 8px; }

/* 按钮美化 */
.el-button--warning {
  background: linear-gradient(90deg, #ffce00, #ffb300);
  color: #333;
  border: none;
}
.el-button--primary {
  background: linear-gradient(90deg, #ffb300, #ff8c00);
  color: #fff;
  border: none;
}
.el-button:hover { opacity: .9; transform: translateY(-1px); }
</style>