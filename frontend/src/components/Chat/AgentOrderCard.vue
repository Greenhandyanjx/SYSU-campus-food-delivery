<template>
  <div class="agent-order-card">
    <!-- 卡片头部：商家 + 金额 -->
    <div class="card-header">
      <div class="merchant-info">
        <img v-if="data.logo" :src="data.logo" class="merchant-logo" alt="logo"
          @error="onLogoError" />
        <span v-else class="merchant-icon">🏪</span>
        <span class="merchant-name">{{ data.merchant || '商家' }}</span>
      </div>
      <div class="order-amount">¥{{ formatPrice(data.amount) }}</div>
    </div>

    <!-- 订单编号 -->
    <div class="order-id-row">
      <span class="label">订单号</span>
      <span class="value">#{{ data.orderId }}</span>
    </div>

    <!-- 菜品列表（简要） -->
    <div class="dish-list" v-if="data.dishes && data.dishes.length > 0">
      <div class="dish-item" v-for="(dish, i) in data.dishes.slice(0, 4)" :key="i">
        <span class="dish-name">{{ dish.name }}</span>
        <span class="dish-qty">×{{ dish.qty }}</span>
        <span class="dish-price">¥{{ formatPrice(dish.price * dish.qty) }}</span>
      </div>
      <div class="dish-more" v-if="data.dishes.length > 4">
        ...还有 {{ data.dishes.length - 4 }} 个菜品
      </div>
    </div>

    <!-- 收货信息 -->
    <div class="consignee-info" v-if="data.consignee || data.address">
      <div class="info-row" v-if="data.consignee">
        <span class="info-label">👤</span>
        <span>{{ data.consignee }}</span>
        <span v-if="data.phone" class="info-phone">{{ data.phone }}</span>
      </div>
      <div class="info-row" v-if="data.address">
        <span class="info-label">📍</span>
        <span class="info-addr">{{ data.address }}</span>
      </div>
    </div>

    <!-- 订单时间 -->
    <div class="order-time" v-if="data.orderTime">
      🕐 {{ data.orderTime }}
    </div>

    <!-- 状态标签 -->
    <div class="status-bar">
      <span class="status-tag" :class="statusClass">{{ statusText }}</span>
    </div>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <!-- 通用：查看详情 -->
      <button class="action-btn primary" @click="viewDetail">
        📄 查看详情
      </button>

      <!-- 待支付 (status=1) -->
      <template v-if="data.status === 1">
        <button class="action-btn pay" @click="sendToAgent('帮我支付订单 #' + data.orderId)">
          💳 支付订单
        </button>
        <button class="action-btn outline" @click="sendToAgent('帮我取消订单 #' + data.orderId)">
          ❌ 取消订单
        </button>
        <button class="action-btn outline" @click="modifyAddress">
          ✏️ 修改地址
        </button>
      </template>

      <!-- 已支付/待接单 (status=2) -->
      <template v-else-if="data.status === 2">
        <button class="action-btn outline" @click="sendToAgent('帮我取消订单 #' + data.orderId)">
          ❌ 取消订单
        </button>
      </template>

      <!-- 已接单/准备中 (status=3) -->
      <template v-else-if="data.status === 3">
        <button class="action-btn outline" @click="contactMerchant">
          📞 联系商家
        </button>
      </template>

      <!-- 配送中 (status=4) -->
      <template v-else-if="data.status === 4">
        <button class="action-btn outline" @click="sendToAgent('查询订单 #' + data.orderId + ' 的配送进度')">
          🚚 配送进度
        </button>
        <button class="action-btn outline" @click="contactMerchant">
          📞 联系商家
        </button>
      </template>

      <!-- 已完成 (status=5) -->
      <template v-else-if="data.status === 5">
        <button class="action-btn highlight" @click="reviewOrder">
          ⭐ 评价订单
        </button>
        <button class="action-btn outline" @click="sendToAgent('再来一单，和订单 #' + data.orderId + ' 一样')">
          🔄 再来一单
        </button>
      </template>

      <!-- 已取消 (status=6) -->
      <template v-else-if="data.status === 6">
        <button class="action-btn outline" @click="sendToAgent('想重新下单，和之前 #' + data.orderId + ' 一样')">
          🔄 重新下单
        </button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * AgentOrderCard.vue — AI 助手消息中的订单卡片组件
 * ==================================================
 * 在聊天气泡中渲染一个可交互的订单卡片，包含根据订单状态的按钮组。
 * 按钮操作包括：页面跳转（查看详情/修改地址/评价）和
 * 向助手发消息（支付/取消/联系商家/再来一单）。
 */
import { computed } from 'vue'
import { useRouter } from 'vue-router'

export interface OrderDish {
  name: string
  qty: number
  price: number
}

export interface OrderCardData {
  orderId: number | string
  status: number
  merchant: string
  amount: number
  deliveryFee?: number
  dishes: OrderDish[]
  consignee?: string
  address?: string
  phone?: string
  orderTime?: string
  logo?: string
}

const props = defineProps<{
  data: OrderCardData
}>()

const emit = defineEmits<{
  sendMessage: [text: string]
}>()

const router = useRouter()

// 状态文本映射
const statusText = computed(() => {
  const map: Record<number, string> = {
    1: '⏳ 待支付',
    2: '✅ 已支付',
    3: '👨‍🍳 准备中',
    4: '🚴 配送中',
    5: '✅ 已完成',
    6: '❌ 已取消',
  }
  return map[props.data.status] || '未知状态'
})

// 状态 CSS class
const statusClass = computed(() => {
  const map: Record<number, string> = {
    1: 'status-pending',
    2: 'status-paid',
    3: 'status-preparing',
    4: 'status-delivering',
    5: 'status-completed',
    6: 'status-cancelled',
  }
  return map[props.data.status] || ''
})

// logo 加载失败时隐藏图片，由 v-if="data.logo" 控制回退到 emoji
function onLogoError(e: Event) {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
}

// 格式化价格
function formatPrice(price: number): string {
  return (price || 0).toFixed(2)
}

// 查看详情
function viewDetail() {
  router.push(`/user/order/${props.data.orderId}`)
}

// 评价订单
function reviewOrder() {
  router.push(`/user/review/${props.data.orderId}`)
}

// 修改地址
function modifyAddress() {
  router.push(`/user/address`)
}

// 联系商家
function contactMerchant() {
  sendToAgent('我想联系商家，订单 #' + props.data.orderId)
}

// 向 AI 助手发送消息
function sendToAgent(text: string) {
  emit('sendMessage', text)
}
</script>

<style scoped>
.agent-order-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e8e8e8;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  padding: 12px;
  margin-top: 8px;
  margin-bottom: 4px;
  font-size: 13px;
  line-height: 1.5;
}

/* 头部 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.merchant-info {
  display: flex;
  align-items: center;
  gap: 4px;
}

.merchant-icon {
  font-size: 16px;
}

.merchant-logo {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}

.merchant-name {
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.order-amount {
  font-size: 16px;
  font-weight: 700;
  color: #FF6B35;
}

/* 订单号 */
.order-id-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #f0f0f0;
}

.order-id-row .value {
  color: #666;
  font-family: monospace;
}

/* 菜品列表 */
.dish-list {
  margin-bottom: 8px;
}

.dish-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 3px 0;
  font-size: 13px;
}

.dish-name {
  flex: 1;
  color: #555;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dish-qty {
  color: #999;
  margin: 0 8px;
  min-width: 24px;
  text-align: right;
}

.dish-price {
  color: #333;
  font-weight: 500;
  min-width: 56px;
  text-align: right;
}

.dish-more {
  color: #aaa;
  font-size: 12px;
  padding: 2px 0;
}

/* 收货信息 */
.consignee-info {
  background: #f9f9f9;
  border-radius: 8px;
  padding: 8px 10px;
  margin-bottom: 8px;
  font-size: 12px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 2px;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  min-width: 20px;
  text-align: center;
}

.info-phone {
  margin-left: 8px;
  color: #999;
}

.info-addr {
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 订单时间 */
.order-time {
  font-size: 11px;
  color: #bbb;
  margin-bottom: 8px;
}

/* 状态栏 */
.status-bar {
  margin-bottom: 10px;
}

.status-tag {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.status-pending { background: #FFF3E0; color: #E65100; }
.status-paid { background: #E8F5E9; color: #2E7D32; }
.status-preparing { background: #FFF8E1; color: #F57F17; }
.status-delivering { background: #E3F2FD; color: #1565C0; }
.status-completed { background: #E8F5E9; color: #1B5E20; }
.status-cancelled { background: #FBE9E7; color: #BF360C; }

/* 操作按钮 */
.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.action-btn {
  padding: 6px 12px;
  border-radius: 16px;
  border: 1px solid #e0e0e0;
  background: white;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  color: #555;
}

.action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.action-btn:active {
  transform: translateY(0);
}

.action-btn.primary {
  border-color: #FF6B35;
  color: #FF6B35;
  background: #FFF3E8;
}

.action-btn.primary:hover {
  background: #FF6B35;
  color: white;
}

.action-btn.pay {
  border-color: #FF9800;
  color: #E65100;
  background: #FFF3E0;
}

.action-btn.pay:hover {
  background: #FF9800;
  color: white;
}

.action-btn.highlight {
  border-color: #4CAF50;
  color: #2E7D32;
  background: #E8F5E9;
}

.action-btn.highlight:hover {
  background: #4CAF50;
  color: white;
}

.action-btn.outline {
  border-color: #ddd;
  color: #666;
  background: #fafafa;
}

.action-btn.outline:hover {
  border-color: #bbb;
  background: #f0f0f0;
}
</style>
