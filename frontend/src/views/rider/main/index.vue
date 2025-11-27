<template>
  <div class="rider-home">
    <!-- 顶部状态栏 -->
    <div class="status-bar">
      <div class="time-info">{{ currentTime }}</div>
      <div class="signal-info">
        <i class="css-icon signal"></i>
        <i class="css-icon wifi"></i>
        <i class="css-icon battery"></i>
      </div>
    </div>

    <!-- 骑手信息卡片 -->
    <div class="rider-info-card">
      <div class="rider-header">
        <div class="location-info">
          <i class="css-icon location"></i>
          <span>{{ currentLocation }}</span>
        </div>
        <div class="weather-info">
          <i class="css-icon sunny"></i>
          <span>{{ weather }}°C</span>
        </div>
      </div>

      <div class="rider-profile">
        <el-avatar :size="60" :src="riderInfo.avatar" />
        <div class="rider-details">
          <h2>{{ riderInfo.name }}</h2>
          <div class="rating-section">
            <el-rate v-model="riderInfo.rating" disabled />
            <span class="rating-text">{{ riderInfo.rating }}分 · {{ riderInfo.completedOrders }}单</span>
          </div>
        </div>
        <div class="online-control">
          <el-switch
            v-model="isOnline"
            active-color="#FFD700"
            inactive-color="#DCDFE6"
            :active-text="isOnline ? '在线接单' : '已下线'"
            @change="toggleOnlineStatus"
          />
        </div>
      </div>
    </div>

    <!-- 数据概览 -->
    <div class="data-overview">
      <div class="overview-card today">
        <div class="card-icon">💰</div>
        <div class="card-content">
          <div class="card-value">¥{{ todayIncome.toFixed(2) }}</div>
          <div class="card-label">今日收入</div>
        </div>
      </div>
      <div class="overview-card today-orders">
        <div class="card-icon">📦</div>
        <div class="card-content">
          <div class="card-value">{{ todayOrders }}</div>
          <div class="card-label">今日订单</div>
        </div>
      </div>
      <div class="overview-card delivering">
        <div class="card-icon">🚴</div>
        <div class="card-content">
          <div class="card-value">{{ deliveringOrders }}</div>
          <div class="card-label">配送中</div>
        </div>
      </div>
      <div class="overview-card waiting">
        <div class="card-icon">⏱️</div>
        <div class="card-content">
          <div class="card-value">{{ waitingPickup }}</div>
          <div class="card-label">待取货</div>
        </div>
      </div>
    </div>

    <!-- 订单管理入口 -->
    <div class="order-entry">
      <div class="entry-header">
        <h3>订单管理</h3>
        <el-button type="text" @click="goToOrders">查看全部</el-button>
      </div>
      <div class="order-stats">
        <div class="stat-item new" @click="goToNewOrders">
          <div class="stat-icon">
            <i class="css-icon notification"></i>
            <span class="stat-badge" v-if="pendingOrders > 0">{{ pendingOrders }}</span>
          </div>
          <span>新订单</span>
        </div>
        <div class="stat-item pickup" @click="goToPickupOrders">
          <div class="stat-icon">
            <i class="css-icon shop"></i>
          </div>
          <span>待取货</span>
        </div>
        <div class="stat-item delivering" @click="goToDeliveringOrders">
          <div class="stat-icon">
            <i class="css-icon bike"></i>
          </div>
          <span>配送中</span>
        </div>
        <div class="stat-item completed" @click="goToCompletedOrders">
          <div class="stat-icon">
            <i class="css-icon success"></i>
          </div>
          <span>已完成</span>
        </div>
      </div>
    </div>

    <!-- 快捷功能 -->
    <div class="quick-functions">
      <div class="function-item" @click="goToWallet">
        <div class="function-icon wallet">
          <i class="css-icon wallet"></i>
        </div>
        <span>我的钱包</span>
      </div>
      <div class="function-item" @click="goToStats">
        <div class="function-icon stats">
          <i class="css-icon data-analysis"></i>
        </div>
        <span>数据统计</span>
      </div>
      <div class="function-item" @click="goToSettings">
        <div class="function-icon settings">
          <i class="css-icon setting"></i>
        </div>
        <span>工作设置</span>
      </div>
      <div class="function-item" @click="goToHelp">
        <div class="function-icon help">
          <i class="css-icon service"></i>
        </div>
        <span>帮助中心</span>
      </div>
    </div>

    <!-- 底部导航栏 -->
    <div class="bottom-nav">
      <div class="nav-item active" @click="switchNav('home')">
        <i class="css-icon house"></i>
        <span>首页</span>
      </div>
      <div class="nav-item" @click="switchNav('orders')">
        <i class="css-icon list"></i>
        <span>订单</span>
      </div>
      <div class="nav-item" @click="switchNav('stats')">
        <i class="css-icon data-analysis"></i>
        <span>统计</span>
      </div>
      <div class="nav-item" @click="switchNav('profile')">
        <i class="css-icon user"></i>
        <span>我的</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 基础状态
const currentTime = ref('')
const currentLocation = ref('定位中...')
const weather = ref(25)
const isOnline = ref(false)
const loading = ref(false)

// 骑手信息
const riderInfo = ref({
  name: '骑手',
  avatar: '',
  rating: 5.0,
  completedOrders: 0
})

// 数据概览
const todayIncome = ref(0)
const todayOrders = ref(0)
const deliveringOrders = ref(0)
const waitingPickup = ref(0)
const pendingOrders = ref(0)

// 更新时间
let timer = null
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('zh-CN', { hour12: false })
}

// 初始化数据
const initData = async () => {
  try {
    loading.value = true

    // 获取骑手信息
    const riderData = await riderApi.getRiderInfo()
    if (riderData.code === 1 && riderData.data) {
      riderInfo.value = riderData.data
      isOnline.value = riderData.data.isOnline
    }

    // 获取工作台数据
    const dashboardData = await riderApi.getDashboard()
    if (dashboardData.code === 1 && dashboardData.data) {
      todayIncome.value = dashboardData.data.todayIncome || 0
      todayOrders.value = dashboardData.data.todayOrders || 0
      deliveringOrders.value = dashboardData.data.delivering || 0
      waitingPickup.value = dashboardData.data.waitPickup || 0
    }

    // 获取新订单数量（使用真实API）
    try {
      const response = await riderApi.getNewOrders()
      if (response.code === 1 && Array.isArray(response.data)) {
        pendingOrders.value = response.data.length
      } else {
        pendingOrders.value = 0
        console.warn('新订单数据格式不正确:', response)
      }
    } catch (error) {
      console.error('获取新订单数量失败:', error)
      pendingOrders.value = 0
    }

  } catch (error) {
    console.error('初始化数据失败:', error)
    ElMessage.error('获取数据失败，请刷新重试')
  } finally {
    loading.value = false
  }
}

// 切换在线状态
const toggleOnlineStatus = async (status) => {
  try {
    await riderApi.updateRiderStatus({ isOnline: status })
    ElMessage.success(status ? '已上线，开始接单' : '已下线，停止接单')

    // 更新状态后刷新数据
    await initData()
  } catch (error) {
    ElMessage.error('状态更新失败，请重试')
    // 回滚状态
    isOnline.value = !status
  }
}

// 导航跳转
const switchNav = (nav) => {
  const routes = {
    home: '/rider',
    orders: '/rider/orders',
    stats: '/rider/stats',
    profile: '/rider/profile'
  }

  if (routes[nav]) {
    router.push(routes[nav])
  }
}

// 订单相关跳转
const goToOrders = () => {
  router.push('/rider/orders')
}

const goToNewOrders = () => {
  router.push('/rider/dashboard')
}

const goToPickupOrders = () => {
  router.push('/rider/dashboard')
}

const goToDeliveringOrders = () => {
  router.push('/rider/dashboard')
}

const goToCompletedOrders = () => {
  router.push('/rider/orders')
}

// 功能跳转
const goToWallet = () => {
  router.push('/rider/wallet')
}

const goToStats = () => {
  router.push('/rider/stats')
}

const goToSettings = () => {
  router.push('/rider/profile/work')
}

const goToHelp = () => {
  router.push('/rider/profile/help')
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
  initData()

  // 获取用户位置
  if (navigator.geolocation) {
    navigator.geolocation.getCurrentPosition(
      (position) => {
        currentLocation.value = '已定位'
      },
      (error) => {
        console.warn('获取位置失败:', error)
        currentLocation.value = '定位失败'
      }
    )
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
/* CSS图标样式 */
.css-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  position: relative;
  font-size: inherit;
  color: inherit;
}

/* 信号图标 */
.css-icon.signal::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 4px;
  height: 4px;
  background: currentColor;
  border-radius: 50%;
  box-shadow:
    -8px -8px 0 0 currentColor,
    -16px -16px 0 0 currentColor,
    -24px -24px 0 0 currentColor;
}

/* WiFi图标 */
.css-icon.wifi::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 20px;
  height: 10px;
  border: 2px solid currentColor;
  border-top: none;
  border-radius: 0 0 10px 10px;
  background: transparent;
}

.css-icon.wifi::after {
  content: '';
  position: absolute;
  bottom: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 10px;
  height: 5px;
  border: 2px solid currentColor;
  border-top: none;
  border-radius: 0 0 5px 5px;
  background: transparent;
}

/* 电池图标 */
.css-icon.battery::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  transform: translateY(-50%);
  width: 20px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.battery::after {
  content: '';
  position: absolute;
  top: 50%;
  right: -4px;
  transform: translateY(-50%);
  width: 2px;
  height: 6px;
  background: currentColor;
  border-radius: 0 1px 1px 0;
}

/* 定位图标 */
.css-icon.location::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 10px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 50% 50% 50% 0;
  transform: translateX(-50%) rotate(-45deg);
}

.css-icon.location::after {
  content: '';
  position: absolute;
  top: 8px;
  left: 50%;
  transform: translateX(-50%) translateY(-50%);
  width: 4px;
  height: 4px;
  background: currentColor;
  border-radius: 50%;
}

/* 太阳图标 */
.css-icon.sunny::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 10px;
  height: 10px;
  background: currentColor;
  border-radius: 50%;
  box-shadow:
    -15px 0 0 -2px currentColor,
    15px 0 0 -2px currentColor,
    0 -15px 0 -2px currentColor,
    0 15px 0 -2px currentColor,
    -10px -10px 0 -2px currentColor,
    10px -10px 0 -2px currentColor,
    -10px 10px 0 -2px currentColor,
    10px 10px 0 -2px currentColor;
}

/* 房子图标 */
.css-icon.house::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 10px;
  border: 2px solid currentColor;
  border-top: none;
}

.css-icon.house::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 0;
  border-left: 10px solid transparent;
  border-right: 10px solid transparent;
  border-bottom: 8px solid currentColor;
}

/* 列表图标 */
.css-icon.list::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 0 4px 0 currentColor, 0 8px 0 currentColor;
}

/* 数据分析图标 */
.css-icon.data-analysis::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: 2px;
  width: 3px;
  height: 6px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 4px 0 0 currentColor, 8px 0 0 currentColor, 12px 0 0 currentColor;
}

.css-icon.data-analysis::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 2px;
  width: 3px;
  height: 10px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 8px 0 0 currentColor;
}

/* 用户图标 */
.css-icon.user::before {
  content: '';
  position: absolute;
  top: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
}

.css-icon.user::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 10px;
  height: 8px;
  background: currentColor;
  border-radius: 5px 5px 0 0;
}

/* 钱包图标 */
.css-icon.wallet::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.wallet::after {
  content: '';
  position: absolute;
  top: 6px;
  left: 8px;
  width: 6px;
  height: 1px;
  background: currentColor;
  border-radius: 1px;
}

/* 设置图标 */
.css-icon.setting::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.setting::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
}

/* 服务图标 */
.css-icon.service::before {
  content: '';
  position: absolute;
  top: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 0 4px 0 currentColor, 0 8px 0 currentColor;
}

.css-icon.service::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 2px;
  height: 10px;
  background: currentColor;
  border-radius: 1px;
}

/* 商店图标 */
.css-icon.shop::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 8px;
  border: 2px solid currentColor;
  border-bottom: none;
  border-radius: 8px 8px 0 0;
}

.css-icon.shop::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 8px;
  border: 2px solid currentColor;
  border-top: none;
}

/* 通知图标 */
.css-icon.notification::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.notification::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 6px;
  height: 6px;
  background: #ff4757;
  border-radius: 50%;
  border: 1px solid white;
}

/* 骑行图标 */
.css-icon.bike::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 16px;
  height: 10px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.bike::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
}

/* 成功图标 */
.css-icon.success::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-45deg);
  width: 10px;
  height: 6px;
  border-left: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
  border-radius: 0 0 0 2px;
}

.rider-home {
  background: linear-gradient(to bottom, #FFFDE7, #FFFFFF);
  min-height: 100vh;
  padding-bottom: 60px;
  font-family: 'PingFang SC', 'Helvetica Neue', sans-serif;
}

/* 顶部状态栏 */
.status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  background: #FFD700;
  color: #333;
  font-size: 14px;
}

.signal-info {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* 骑手信息卡片 */
.rider-info-card {
  background: white;
  margin: 10px;
  border-radius: 12px;
  padding: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.rider-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.location-info, .weather-info {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #666;
  font-size: 14px;
}

.rider-profile {
  display: flex;
  align-items: center;
  gap: 15px;
}

.rider-details {
  flex: 1;
}

.rider-details h2 {
  margin: 0;
  font-size: 20px;
  color: #333;
}

.rating-section {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 5px;
}

.rating-text {
  color: #666;
  font-size: 14px;
}

/* 数据概览 */
.data-overview {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin: 10px;
}

.overview-card {
  background: white;
  border-radius: 10px;
  padding: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.overview-card.today {
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
  color: white;
}

.overview-card.today-orders {
  background: linear-gradient(135deg, #409EFF 0%, #1890ff 100%);
  color: white;
}

.overview-card.delivering {
  background: linear-gradient(135deg, #F56C6C 0%, #ff4757 100%);
  color: white;
}

.overview-card.waiting {
  background: linear-gradient(135deg, #E6A23C 0%, #ffa502 100%);
  color: white;
}

.card-icon {
  font-size: 24px;
  opacity: 0.9;
}

.card-value {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 2px;
}

.card-label {
  font-size: 12px;
  opacity: 0.8;
}

/* 订单管理入口 */
.order-entry {
  background: white;
  margin: 10px;
  border-radius: 12px;
  padding: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.entry-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.entry-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.order-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  padding: 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.stat-item.new {
  background: rgba(255, 215, 0, 0.1);
}

.stat-item.pickup {
  background: rgba(64, 158, 255, 0.1);
}

.stat-item.delivering {
  background: rgba(245, 108, 108, 0.1);
}

.stat-item.completed {
  background: rgba(103, 194, 58, 0.1);
}

.stat-item:hover {
  transform: translateY(-2px);
}

.stat-icon {
  position: relative;
  font-size: 20px;
  color: #666;
}

.stat-badge {
  position: absolute;
  top: -5px;
  right: -5px;
  background: #ff4757;
  color: white;
  font-size: 10px;
  padding: 2px 4px;
  border-radius: 10px;
  min-width: 16px;
  text-align: center;
}

.stat-item span {
  font-size: 12px;
  color: #666;
}

/* 快捷功能 */
.quick-functions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
  margin: 10px;
}

.function-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 15px 10px;
  background: white;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.function-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.function-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.function-icon.wallet {
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
  color: white;
}

.function-icon.stats {
  background: linear-gradient(135deg, #409EFF 0%, #1890ff 100%);
  color: white;
}

.function-icon.settings {
  background: linear-gradient(135deg, #909399 0%, #606266 100%);
  color: white;
}

.function-icon.help {
  background: linear-gradient(135deg, #67C23A 0%, #52c41a 100%);
  color: white;
}

.function-item span {
  font-size: 12px;
  color: #666;
}

/* 底部导航 */
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  border-top: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-around;
  padding: 5px 0;
  z-index: 100;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 5px 15px;
  cursor: pointer;
  color: #999;
  transition: all 0.3s ease;
}

.nav-item.active {
  color: #FFD700;
}

.nav-item .css-icon {
  font-size: 20px;
}

.nav-item span {
  font-size: 12px;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .data-overview {
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
    margin: 8px;
  }

  .order-stats {
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }

  .quick-functions {
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
    margin: 8px;
  }
}
</style>