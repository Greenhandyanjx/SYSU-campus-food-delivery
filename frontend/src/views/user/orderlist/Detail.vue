<template>
  <div class="order-detail-bg">
    <div class="order-detail-page">
      <div class="header">
        <el-button type="text" @click="$router.back()" class="back-btn" size="large">
          <i class="el-icon-arrow-left"></i> 返回
        </el-button>
        <h2>订单详情 #{{ order?.id || id }}</h2>
        <div class="status-badge" :class="order?.status">{{ order?.statusText }}</div>
      </div>

      <div v-if="order" class="detail-card">
        <div class="section">
          <div class="section-header">
            <h3>基本信息</h3>
          </div>
          <div class="store-info" @click="openStore(order.storeId)">
            <img :src="order.storeLogo || '/src/assets/noImg.png'" @error="onImgError" class="store-logo" />
            <div class="store-detail">
              <div class="store-name">{{ order.storeName }}</div>
              <div class="order-time">下单时间：{{ order.time }}</div>
            </div>
            <i class="el-icon-arrow-right goto-icon"></i>
          </div>
        </div>

        <div class="section">
          <div class="section-header">
            <h3>商品清单</h3>
          </div>
          <div class="goods">
            <div v-for="(it, i) in order.items" :key="i" class="goods-item">
              <img :src="it.image || '/src/assets/noImg.png'" @error="onImgError" class="goods-img" />
              <div class="goods-info">
                <div class="goods-name">{{ it.name }}</div>
                <div class="goods-price">¥{{ it.price.toFixed(2) }}</div>
              </div>
              <div class="goods-count">x{{ it.count }}</div>
              <div class="goods-total">¥{{ (it.price * it.count).toFixed(2) }}</div>
            </div>
          </div>
          <div class="total">
            <span class="total-label">合计</span>
            <span class="total-price">¥{{ totalPrice.toFixed(2) }}</span>
          </div>
        </div>

        <div class="section">
          <div class="section-header">
            <h3>配送信息</h3>
          </div>
          <div class="delivery-info">
            <div class="delivery-status">
              <i class="el-icon-location status-icon"></i>
              <div class="status-text">{{ getDeliveryStatus(order.status) }}</div>
            </div>
            <div class="address-info">
              <i class="el-icon-map-location"></i>
              <span>{{ order.address }}</span>
            </div>
            <div v-if="order.rider" class="rider-info">
              <img :src="order.rider.avatar || '/src/assets/noImg.png'" @error="onImgError" class="rider-avatar" />
              <div class="rider-detail">
                <div class="rider-name">{{ order.rider.name }}</div>
                <div class="rider-phone">{{ order.rider.phone }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- 操作按钮组：不同状态显示不同按钮 -->
        <div class="action-bar">
          <!-- 待付款 -->
          <template v-if="order.status === 'pending'">
            <div class="countdown" v-if="countdown">
              剩余支付时间：<strong>{{ countdown }}</strong>
            </div>
            <el-button @click="onCancel" plain>取消订单</el-button>
            <el-button type="primary" @click="onPay">去付款</el-button>
          </template>

          <!-- 配送中 -->
          <template v-if="order.status === 'shipping'">
            <el-button @click="contactRider" plain>联系骑手</el-button>
            <el-button @click="onReorder" plain>再次购买</el-button>
            <el-button type="primary" @click="onConfirm">确认收货</el-button>
          </template>

          <!-- 已完成 -->
          <template v-if="order.status === 'completed'">
            <el-button @click="contactRider" plain>联系骑手</el-button>
            <el-button @click="onReorder" plain>再次购买</el-button>
            <el-button type="warning" v-if="!order.reviewed" @click="onReview">评价晒单</el-button>
          </template>

          <!-- 退款/售后 -->
          <template v-if="order.status === 'refund'">
            <el-button type="info" plain @click="onViewRefund">查看售后详情</el-button>
          </template>
        </div>

      </div>

      <div v-else class="empty">未能加载订单详情</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import orderApi from '@/api/user/order'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const id = route.params.id
const order = ref(null)
const countdown = ref('')

const totalPrice = computed(() =>
  (order.value?.items || []).reduce((s, it) => s + it.price * it.count, 0)
)

function getDeliveryStatus(status) {
  switch(status) {
    case 'pending': return '等待支付'
    case 'shipping': return '正在配送中'
    case 'completed': return '订单已完成'
    case 'refund': return '退款处理中'
    default: return '未知状态'
  }
}

// 按钮事件处理器 - 从列表页复用逻辑
async function onPay() {
  try {
    await orderApi.payOrder(order.value.id)
    ElMessage.success('支付成功')
    await fetch() // 刷新订单状态
  } catch(e) {
    ElMessage.error('支付失败，请重试')
  }
}

async function onCancel() {
  try {
    await ElMessageBox.confirm('确定要取消这个订单吗？', '取消订单', {
      type: 'warning'
    })
    await orderApi.cancelOrder(order.value.id)
    ElMessage.success('订单已取消')
    await fetch()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('取消订单失败')
    }
  }
}

async function onConfirm() {
  try {
    await ElMessageBox.confirm('确认已收到商品了吗？', '确认收货', {
      type: 'warning'
    })
    await orderApi.confirmOrder(order.value.id)
    ElMessage.success('收货成功')
    await fetch()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('确认收货失败')
    }
  }
}

async function onReorder() {
  try {
    await orderApi.reorder(order.value.id)
    ElMessage.success('已添加到购物车')
    router.push('/user/cart')
  } catch (e) {
    ElMessage.error('添加购物车失败')
  }
}

async function onReview() {
  router.push(`/user/review/${order.value.id}`)
}

async function onViewRefund() {
  try {
    const detail = await orderApi.refundDetail(order.value.id)
    // 这里可以打开退款详情弹窗
    ElMessageBox.alert(detail, '退款详情', {
      type: 'info'
    })
  } catch (e) {
    ElMessage.error('获取退款详情失败')
  }
}

async function contactRider() {
  try {
    if (!order.value.rider) {
      ElMessage.warning('暂无骑手信息')
      return
    }
    const info = await orderApi.contactRider(order.value.id)
    ElMessageBox.alert(`骑手电话：${info.phone}`, '联系骑手', {
      type: 'info'
    })
  } catch (e) {
    ElMessage.error('获取骑手信息失败')
  }
}

function openStore(id) {
  router.push(`/user/store/${id}`)
}

function onImgError(e) {
  try { e.target.src = '/src/assets/noImg.png' } catch (err) {}
}

// 更新倒计时
function updateCountdown() {
  if (!order.value?.payDeadline) return
  const now = Date.now()
  const deadline = new Date(order.value.payDeadline).getTime()
  const diff = deadline - now
  
  if (diff <= 0) {
    countdown.value = '已超时'
    fetch() // 刷新订单
    return
  }
  
  const minutes = Math.floor(diff / 60000)
  const seconds = Math.floor((diff % 60000) / 1000)
  countdown.value = `${minutes}:${seconds.toString().padStart(2, '0')}`
}

let timer = null

async function fetch() {
  try {
    const res = await orderApi.getOrderDetail(id)
    order.value = res?.data || res || null
    
    if (order.value?.status === 'pending') {
      updateCountdown()
      if (!timer) {
        timer = setInterval(updateCountdown, 1000)
      }
    } else if (timer) {
      clearInterval(timer)
      timer = null
    }
  } catch (e) {
    console.warn('API 不可用，使用本地数据')
    try {
      order.value = window.__RAW_ORDERS__?.find((o) => o.id === id) || null
    } catch {
      order.value = null
    }
  }
}

onMounted(() => {
  fetch()
})

onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})
</script>

<style scoped>
.order-detail-bg{
  width: 100%;
  min-height: 100vh;
  background: url('/src/assets/login/img_denglu_bj.jpg') center/cover no-repeat;
  background-attachment: fixed;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 60px 0;
}

/* 主内容区 */
.order-detail-page {
  width: 60%;
  max-width: 900px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(255, 193, 7, 0.35);
  padding: 28px;
  backdrop-filter: blur(6px);
  position: relative;
  z-index: 1;
  transition: transform 0.3s, box-shadow 0.3s;
}

.order-detail-page:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 28px rgba(255, 193, 7, 0.45);
}

/* 标题与返回 */
.header {
  display: flex;
  align-items: center;
  gap: 16px;
  border-bottom: 2px solid #ffe58f;
  padding-bottom: 12px;
  margin-bottom: 16px;
}

.header h2 {
  font-weight: 600;
  color: #333;
  margin: 0;
}

.back-btn {
  font-weight: bold;
  color: #ff9800;
  font-size: 16px;
}

.back-btn:hover {
  color: #ff9900;
}

.status-badge {
  margin-left: auto;
  padding: 4px 12px;
  border-radius: 12px;
  font-weight: 600;
  font-size: 14px;
}

.status-badge.pending { background: #fff7e6; color: #fa8c16; }
.status-badge.shipping { background: #e6f7ff; color: #1890ff; }
.status-badge.completed { background: #f6ffed; color: #52c41a; }
.status-badge.refund { background: #fff2e8; color: #fa541c; }

/* 卡片与区块 */
.detail-card {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.section {
  background: #fffef6;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(255, 193, 7, 0.1);
}

.section-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px dashed #ffe58f;
}

.section-header h3 {
  margin: 0;
  color: #444;
  font-weight: 600;
  font-size: 16px;
}

/* 店铺信息 */
.store-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  transition: 0.2s;
}

.store-info:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 12px rgba(255, 193, 7, 0.15);
}

.store-logo {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  object-fit: cover;
}

.store-detail {
  flex: 1;
}

.store-name {
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.order-time {
  font-size: 13px;
  color: #999;
}

.goto-icon {
  color: #ccc;
  font-size: 18px;
}

/* 商品清单改造 */
.goods {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.goods-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
}

.goods-img {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  object-fit: cover;
}

.goods-info {
  flex: 1;
}

.goods-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.goods-price {
  color: #ff4d4f;
  font-size: 14px;
}

.goods-count {
  font-size: 14px;
  color: #666;
  min-width: 60px;
  text-align: right;
}

.goods-total {
  width: 90px;
  text-align: right;
  color: #ff4d4f;
  font-weight: 600;
}

/* 配送信息 */
.delivery-info {
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.delivery-status {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.status-icon {
  font-size: 24px;
  color: #ff9800;
}

.status-text {
  font-size: 16px;
  color: #333;
  font-weight: 500;
}

.address-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px dashed #eee;
}

.rider-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.rider-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
}

.rider-name {
  font-weight: 500;
  margin-bottom: 2px;
}

.rider-phone {
  font-size: 13px;
  color: #666;
}

.total {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed #ffe58f;
}

.total-label {
  color: #666;
  font-size: 15px;
}

.total-price {
  color: #ff4d4f;
  font-size: 20px;
  font-weight: 600;
}

/* 操作栏 */
.action-bar {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #ffe58f;
}

.countdown {
  margin-right: auto;
  color: #fa8c16;
  font-size: 14px;
}

.countdown strong {
  font-family: monospace;
  font-size: 16px;
}

/* 空状态优化 */
.empty {
  text-align: center;
  padding: 60px 0;
  color: #bbb;
  font-size: 15px;
  background: #fffef6;
  border-radius: 12px;
  margin: 16px 0;
}

/* 响应式优化 */
@media (max-width: 768px) {
  .order-detail-page {
    width: 90%;
    padding: 16px;
  }
  
  .goods-img {
    width: 50px;
    height: 50px;
  }
  
  .action-bar {
    flex-wrap: wrap;
  }
  
  .countdown {
    width: 100%;
    margin-bottom: 12px;
    text-align: center;
  }
}
</style>
