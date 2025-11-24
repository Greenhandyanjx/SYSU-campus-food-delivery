<template>
  <div class="rider-dashboard">
    <!-- 顶部状态栏 -->
    <div class="status-bar">
      <div class="time-info">{{ currentTime }}</div>
      <div class="signal-info">
        <i class="css-icon signal"></i>
        <i class="css-icon wifi"></i>
        <i class="css-icon battery"></i>
      </div>
    </div>

    <!-- 头部信息区 -->
    <div class="header-section">
      <div class="top-bar">
        <div class="location-info">
          <i class="css-icon location"></i>
          <span>定位: {{ currentLocation }}</span>
        </div>
        <div class="weather-info">
          <i class="css-icon sunny"></i>
          <span>{{ weather }}°C</span>
        </div>
      </div>

      <!-- 骑手工作状态 -->
      <div class="work-status">
        <div class="rider-avatar">
          <el-avatar :size="60" :src="riderInfo.avatar" />
          <div class="rider-info">
            <h2>{{ riderInfo.name }}</h2>
            <p class="rating">
              <el-rate v-model="riderInfo.rating" disabled />
              <span class="rating-text">{{ riderInfo.rating }}分 | {{ riderInfo.completedOrders }}单</span>
            </p>
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

      <!-- 收入统计卡片 -->
      <div class="income-cards">
        <div class="income-card">
          <div class="card-title">今日收入</div>
          <div class="card-value">¥{{ dailyIncome.toFixed(2) }}</div>
          <div class="card-subtitle">预计再赚 ¥{{ estimatedIncome }}</div>
        </div>
        <div class="income-card">
          <div class="card-title">本周收入</div>
          <div class="card-value">¥{{ weeklyIncome.toFixed(2) }}</div>
          <div class="card-subtitle">已完成 {{ weeklyOrders }} 单</div>
        </div>
      </div>
    </div>

    <!-- 快捷功能入口 -->
    <div class="quick-actions">
      <div class="action-item" @click="goToHistory">
        <i class="css-icon document"></i>
        <span>历史订单</span>
      </div>
      <div class="action-item" @click="goToWallet">
        <i class="css-icon wallet"></i>
        <span>我的钱包</span>
      </div>
      <div class="action-item" @click="goToRewards">
        <i class="css-icon trophy"></i>
        <span>奖励中心</span>
      </div>
      <div class="action-item" @click="goToHelp">
        <i class="css-icon service"></i>
        <span>帮助中心</span>
      </div>
    </div>

    <!-- 订单管理区域 -->
    <div class="order-management">
      <el-tabs v-model="activeOrderTab" type="border-card" class="order-tabs">
        <el-tab-pane label="待接单" name="new-orders">
          <div class="order-list" v-if="newOrders.length > 0">
            <div v-for="order in newOrders" :key="order.id" class="order-card">
              <div class="order-header">
                <el-tag type="warning">新订单</el-tag>
                <span class="order-time">{{ order.estimatedTime }}分钟内送达</span>
              </div>
              <div class="order-content">
                <div class="route-info">
                  <div class="route-point pickup">
                    <div class="point-icon">
                      <i class="css-icon shop"></i>
                    </div>
                    <div class="point-detail">
                      <div class="point-name">{{ order.restaurant }}</div>
                      <div class="point-address">{{ order.pickupAddress }}</div>
                    </div>
                  </div>
                  <div class="route-arrow">↓</div>
                  <div class="route-point delivery">
                    <div class="point-icon">
                      <i class="css-icon location"></i>
                    </div>
                    <div class="point-detail">
                      <div class="point-name">{{ order.customer }}</div>
                      <div class="point-address">{{ order.deliveryAddress }}</div>
                    </div>
                  </div>
                </div>
                <div class="order-footer">
                  <div class="order-info">
                    <span>距离: {{ order.distance }}km | 预估收入: ¥{{ order.estimatedFee }}</span>
                  </div>
                  <el-button type="primary" size="large" @click="acceptOrder(order)">立即抢单</el-button>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <el-empty description="暂无新订单" />
          </div>
        </el-tab-pane>

        <el-tab-pane label="待取货" name="pickup">
          <div class="order-list" v-if="pickupOrders.length > 0">
            <div v-for="order in pickupOrders" :key="order.id" class="order-card">
              <div class="order-header">
                <el-tag type="info">待取货</el-tag>
                <span class="order-timer">{{ formatTime(order.remainingTime) }}</span>
              </div>
              <div class="order-content">
                <div class="route-info">
                  <div class="route-point pickup">
                    <div class="point-icon">
                      <i class="css-icon shop"></i>
                    </div>
                    <div class="point-detail">
                      <div class="point-name">{{ order.restaurant }}</div>
                      <div class="point-address">{{ order.pickupAddress }}</div>
                    </div>
                  </div>
                </div>
                <div class="order-footer">
                  <div class="order-info">
                    <span>取餐码: {{ order.pickupCode }} | 联系商家: {{ order.shopPhone }}</span>
                  </div>
                  <el-button type="success" size="large" @click="confirmPickup(order)">确认取货</el-button>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <el-empty description="暂无待取货订单" />
          </div>
        </el-tab-pane>

        <el-tab-pane label="配送中" name="delivering">
          <div class="order-list" v-if="deliveringOrders.length > 0">
            <div v-for="order in deliveringOrders" :key="order.id" class="order-card">
              <div class="order-header">
                <el-tag type="danger">配送中</el-tag>
                <span class="order-timer">{{ formatTime(order.remainingTime) }}</span>
              </div>
              <div class="order-content">
                <div class="customer-info">
                  <el-avatar :size="40" :src="order.customerAvatar" />
                  <div class="customer-detail">
                    <div class="customer-name">{{ order.customer }}</div>
                    <div class="customer-phone">{{ order.customerPhone }}</div>
                  </div>
                </div>
                <div class="map-container">
                  <div class="map-placeholder">
                    <i class="css-icon map"></i>
                    <span>配送路线地图</span>
                  </div>
                </div>
                <div class="order-footer">
                  <div class="order-actions">
                    <el-button @click="callCustomer(order)">
                      <i class="css-icon phone"></i>
                      联系顾客
                    </el-button>
                    <el-button type="danger" @click="completeDelivery(order)">
                      <i class="css-icon success"></i>
                      完成配送
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <el-empty description="暂无配送中订单" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 底部导航栏 -->
    <div class="bottom-nav">
      <div class="nav-item" :class="{ active: activeNav === 'home' }" @click="switchNav('home')">
        <i class="css-icon house"></i>
        <span>首页</span>
      </div>
      <div class="nav-item" :class="{ active: activeNav === 'orders' }" @click="switchNav('orders')">
        <i class="css-icon list"></i>
        <span>订单</span>
      </div>
      <div class="nav-item" :class="{ active: activeNav === 'stats' }" @click="switchNav('stats')">
        <i class="css-icon data-analysis"></i>
        <span>统计</span>
      </div>
      <div class="nav-item" :class="{ active: activeNav === 'mine' }" @click="switchNav('mine')">
        <i class="css-icon user"></i>
        <span>我的</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import riderApi from '@/api/rider'
// import {
//   Signal,
//   Wifi,
//   Battery,
//   Location,
//   Sunny,
//   Document,
//   Wallet,
//   Trophy,
//   Service,
//   Shop,
//   House,
//   List,
//   DataAnalysis,
//   User,
//   Map,
//   Phone,
//   SuccessFilled
// } from '@element-plus/icons-vue'

// 路由
const router = useRouter()

// 基础状态
const currentTime = ref('')
const currentLocation = ref('中山大学珠海校区')
const weather = ref(25)
const isOnline = ref(true)
const activeOrderTab = ref('new-orders')
const activeNav = ref('home')
const loading = ref(false)

// 骑手信息
const riderInfo = ref({
  name: '李骑手',
  avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
  rating: 4.8,
  completedOrders: 1250
})

// 收入信息
const dailyIncome = ref(185.5)
const weeklyIncome = ref(1280.0)
const estimatedIncome = ref(45.0)
const weeklyOrders = ref(68)

// 订单数据
const newOrders = ref([
  {
    id: 1,
    restaurant: '麦当劳',
    pickupAddress: '珠海市香洲区唐家湾大学路1号',
    customer: '张同学',
    deliveryAddress: '珠海市香洲区中山大学珠海校区榕园',
    distance: 1.2,
    estimatedFee: 6.5,
    estimatedTime: 20
  },
  {
    id: 2,
    restaurant: '肯德基',
    pickupAddress: '珠海市香洲区唐家湾大学路101号',
    customer: '王老师',
    deliveryAddress: '珠海市香洲区中山大学珠海校区荔园',
    distance: 0.8,
    estimatedFee: 5.0,
    estimatedTime: 15
  }
])

const pickupOrders = ref([
  {
    id: 3,
    restaurant: '星巴克',
    pickupAddress: '珠海市香洲区唐家湾大学路201号',
    pickupCode: 'A123',
    shopPhone: '13788888888',
    remainingTime: 10 * 60 * 1000
  }
])

const deliveringOrders = ref([
  {
    id: 4,
    customer: '陈教授',
    customerPhone: '13666666666',
    customerAvatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
    deliveryAddress: '珠海市香洲区中山大学珠海校区榕园',
    remainingTime: 25 * 60 * 1000
  }
])

// 初始化数据
const initRiderData = async () => {
  try {
    loading.value = true

    // 获取骑手信息
    const riderData = await riderApi.getRiderInfoWithDemo()
    if (riderData.code === 1 && riderData.data) {
      riderInfo.value = riderData.data
      isOnline.value = riderData.data.isOnline
    }

    // 获取收入统计
    const incomeData = await riderApi.getIncomeStatsWithDemo()
    if (incomeData.code === 1 && incomeData.data) {
      dailyIncome.value = incomeData.data.dailyIncome || 0
      weeklyIncome.value = incomeData.data.weeklyIncome || 0
      estimatedIncome.value = incomeData.data.estimatedIncome || 0
      weeklyOrders.value = incomeData.data.completedOrders || 0
    }

    // 获取新订单
    const ordersData = await riderApi.getNewOrdersWithDemo()
    if (ordersData.code === 1 && ordersData.data) {
      newOrders.value = ordersData.data.map(order => ({
        id: order.id,
        restaurant: order.restaurant,
        pickupAddress: order.pickupAddress,
        customer: order.customer,
        deliveryAddress: order.deliveryAddress,
        distance: order.distance,
        estimatedFee: order.estimatedFee,
        estimatedTime: order.estimatedTime,
        createdAt: order.createdAt
      }))
    }

    // 获取待取货订单
    try {
      const pickupData = await riderApi.getPickupOrders()
      if (pickupData.data?.code === 1 && pickupData.data?.data) {
        pickupOrders.value = pickupData.data.data.map(order => ({
          id: order.id,
          restaurant: order.restaurant,
          pickupAddress: order.pickupAddress,
          pickupCode: order.pickupCode,
          shopPhone: order.shopPhone,
          remainingTime: order.remainingTime || 15 * 60 * 1000
        }))
      }
    } catch (e) {
      console.warn('获取待取货订单失败，使用demo数据')
    }

    // 获取配送中订单
    try {
      const deliveringData = await riderApi.getDeliveringOrders()
      if (deliveringData.data?.code === 1 && deliveringData.data?.data) {
        deliveringOrders.value = deliveringData.data.data.map(order => ({
          id: order.id,
          customer: order.customer,
          customerPhone: order.customerPhone,
          customerAvatar: order.customerAvatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
          deliveryAddress: order.deliveryAddress,
          remainingTime: order.remainingTime || 30 * 60 * 1000
        }))
      }
    } catch (e) {
      console.warn('获取配送中订单失败，使用demo数据')
    }

  } catch (error) {
    console.error('初始化骑手数据失败:', error)
    ElMessage.error('获取数据失败，请刷新重试')
  } finally {
    loading.value = false
  }
}

// 更新时间
let timer = null
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('zh-CN', { hour12: false })
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
  initRiderData()

  // 每30秒刷新订单数据
  const orderTimer = setInterval(() => {
    if (isOnline.value) {
      refreshOrders()
    }
  }, 30000)

  // 保存定时器以便清理
  timer.orderTimer = orderTimer
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    if (timer.orderTimer) clearInterval(timer.orderTimer)
  }
})

// 方法
const toggleOnlineStatus = async (status) => {
  try {
    await riderApi.updateOnlineStatus(status)
    ElMessage.success(status ? '已上线，开始接单' : '已下线，停止接单')
  } catch (error) {
    ElMessage.error('状态更新失败，请重试')
    // 回滚状态
    isOnline.value = !status
  }
}

const formatTime = (milliseconds) => {
  const minutes = Math.floor(milliseconds / (60 * 1000))
  const seconds = Math.floor((milliseconds % (60 * 1000)) / 1000)
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
}

const acceptOrder = async (order) => {
  try {
    loading.value = true
    await riderApi.acceptOrder(order.id)
    ElMessage.success(`抢单成功！订单号：${order.id}`)

    // 从新订单中移除，添加到待取货
    newOrders.value = newOrders.value.filter(o => o.id !== order.id)
    pickupOrders.value.push({
      ...order,
      pickupCode: 'A' + order.id,
      shopPhone: '138' + order.id.toString().padStart(8, '0'),
      remainingTime: 15 * 60 * 1000
    })
  } catch (error) {
    ElMessage.error('接单失败，请重试')
  } finally {
    loading.value = false
  }
}

const confirmPickup = async (order) => {
  try {
    loading.value = true
    await riderApi.confirmPickup(order.id)
    ElMessage.success(`取货确认！订单号：${order.id}`)

    // 从待取货中移除，添加到配送中
    pickupOrders.value = pickupOrders.value.filter(o => o.id !== order.id)
    deliveringOrders.value.push({
      ...order,
      customer: order.customer || '张同学',
      customerPhone: order.customerPhone || '13666666666',
      customerAvatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
      remainingTime: 30 * 60 * 1000
    })
  } catch (error) {
    ElMessage.error('取货确认失败，请重试')
  } finally {
    loading.value = false
  }
}

const completeDelivery = async (order) => {
  try {
    loading.value = true
    const result = await riderApi.completeDelivery(order.id)
    ElMessage.success(`配送完成！订单号：${order.id}`)

    // 从配送中移除
    deliveringOrders.value = deliveringOrders.value.filter(o => o.id !== order.id)

    // 更新收入和订单数
    const actualFee = result?.data?.actualFee || order.estimatedFee
    dailyIncome.value += actualFee
    weeklyOrders.value += 1
    riderInfo.value.completedOrders += 1
  } catch (error) {
    ElMessage.error('配送完成失败，请重试')
  } finally {
    loading.value = false
  }
}

const callCustomer = (order) => {
  ElMessage.info(`正在联系顾客：${order.customerPhone}`)
  // 这里可以集成实际的电话拨打功能
}

const switchNav = (nav) => {
  activeNav.value = nav

  // 根据导航项跳转到不同页面
  const routes = {
    home: '/rider',
    orders: '/rider/orders',
    stats: '/rider/stats',
    mine: '/rider/profile'
  }

  if (nav === 'home') {
    ElMessage.info('当前已在首页')
  } else if (routes[nav]) {
    router.push(routes[nav])
  } else {
    ElMessage.info(`切换到${nav === 'orders' ? '订单' : nav === 'stats' ? '统计' : '我的'}`)
  }
}

const goToHistory = () => {
  router.push('/rider/orders')
}

const goToWallet = () => {
  router.push('/rider/wallet')
}

const goToRewards = () => {
  ElMessage.info('奖励中心功能开发中...')
}

const goToHelp = () => {
  ElMessage.info('帮助中心功能开发中...')
}

// 刷新订单数据
const refreshOrders = async () => {
  try {
    // 刷新新订单
    const ordersData = await riderApi.getNewOrdersWithDemo()
    if (ordersData.code === 1 && ordersData.data) {
      newOrders.value = ordersData.data.map(order => ({
        id: order.id,
        restaurant: order.restaurant,
        pickupAddress: order.pickupAddress,
        customer: order.customer,
        deliveryAddress: order.deliveryAddress,
        distance: order.distance,
        estimatedFee: order.estimatedFee,
        estimatedTime: order.estimatedTime,
        createdAt: order.createdAt
      }))
    }

    // 刷新待取货订单
    try {
      const pickupData = await riderApi.getPickupOrders()
      if (pickupData.data?.code === 1 && pickupData.data?.data) {
        pickupOrders.value = pickupData.data.data.map(order => ({
          id: order.id,
          restaurant: order.restaurant,
          pickupAddress: order.pickupAddress,
          pickupCode: order.pickupCode,
          shopPhone: order.shopPhone,
          remainingTime: order.remainingTime || 15 * 60 * 1000
        }))
      }
    } catch (e) {
      console.warn('刷新待取货订单失败')
    }

    // 刷新配送中订单
    try {
      const deliveringData = await riderApi.getDeliveringOrders()
      if (deliveringData.data?.code === 1 && deliveringData.data?.data) {
        deliveringOrders.value = deliveringData.data.data.map(order => ({
          id: order.id,
          customer: order.customer,
          customerPhone: order.customerPhone,
          customerAvatar: order.customerAvatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
          deliveryAddress: order.deliveryAddress,
          remainingTime: order.remainingTime || 30 * 60 * 1000
        }))
      }
    } catch (e) {
      console.warn('刷新配送中订单失败')
    }
  } catch (error) {
    console.error('刷新订单数据失败:', error)
  }
}
</script>

<style>
/* 全局样式，确保CSS图标能正常工作 */
.css-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  position: relative;
  font-size: inherit;
  color: inherit;
}
</style>
<style scoped>
/* CSS图标样式 */

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

/* 文档图标 */
.css-icon.document::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 12px;
  height: 16px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.document::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 6px;
  height: 1px;
  background: currentColor;
  box-shadow: 0 2px 0 currentColor, 0 4px 0 currentColor;
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

/* 奖杯图标 */
.css-icon.trophy::before {
  content: '';
  position: absolute;
  top: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 8px;
  background: currentColor;
  border-radius: 7px 7px 0 0;
}

.css-icon.trophy::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 4px;
  background: currentColor;
  border-radius: 0 0 2px 2px;
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

.css-icon.list::after {
  content: '';
  position: absolute;
  top: 0;
  right: 2px;
  width: 10px;
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

/* 地图图标 */
.css-icon.map::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.map::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 3px;
  height: 3px;
  background: currentColor;
  border-radius: 50%;
  box-shadow: 6px 3px 0 currentColor, 3px 6px 0 currentColor;
}

/* 电话图标 */
.css-icon.phone::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-45deg);
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 20% 20% 20% 20%;
}

.css-icon.phone::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 6px;
  height: 6px;
  border-left: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
  border-radius: 0 0 0 2px;
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
.rider-dashboard {
  background: linear-gradient(to bottom, #FFFDE7, #FFFFFF);
  min-height: 100vh;
  font-family: 'PingFang SC', 'Helvetica Neue', sans-serif;
  padding-bottom: 60px;
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

/* 头部信息区 */
.header-section {
  background: white;
  margin: 10px;
  border-radius: 12px;
  padding: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.top-bar {
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

/* 工作状态 */
.work-status {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.rider-avatar {
  display: flex;
  align-items: center;
  gap: 15px;
}

.rider-info h2 {
  margin: 0;
  font-size: 20px;
  color: #333;
}

.rider-info .rating {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 5px 0 0 0;
}

.rating-text {
  color: #666;
  font-size: 14px;
}

/* 收入卡片 */
.income-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
  margin-top: 15px;
}

.income-card {
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
  padding: 15px;
  border-radius: 10px;
  color: white;
  text-align: center;
}

.card-title {
  font-size: 12px;
  opacity: 0.9;
  margin-bottom: 5px;
}

.card-value {
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 5px;
}

.card-subtitle {
  font-size: 12px;
  opacity: 0.8;
}

/* 快捷功能 */
.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
  margin: 15px 10px;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  padding: 15px;
  background: white;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.action-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.action-item .css-icon {
  font-size: 24px;
  color: #FFD700;
}

.action-item span {
  font-size: 12px;
  color: #666;
}

/* 订单管理 */
.order-management {
  margin: 15px 10px;
}

.order-tabs :deep(.el-tabs__header) {
  border-radius: 10px 10px 0 0;
  overflow: hidden;
}

.order-tabs :deep(.el-tabs__nav) {
  background: #FFD700;
}

.order-tabs :deep(.el-tabs__item) {
  color: white;
  border-right: 1px solid rgba(255, 255, 255, 0.3);
}

.order-tabs :deep(.el-tabs__item.is-active) {
  background: white;
  color: #FFD700;
}

.order-list {
  max-height: 400px;
  overflow-y: auto;
}

.order-card {
  background: white;
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 10px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.order-time {
  color: #666;
  font-size: 12px;
}

.order-timer {
  color: #f56c6c;
  font-weight: bold;
  font-size: 14px;
}

/* 路线信息 */
.route-info {
  margin-bottom: 15px;
}

.route-point {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 10px;
}

.point-icon {
  width: 32px;
  height: 32px;
  background: #FFD700;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.point-detail {
  flex: 1;
}

.point-name {
  font-weight: bold;
  margin-bottom: 2px;
}

.point-address {
  color: #666;
  font-size: 12px;
  line-height: 1.4;
}

.route-arrow {
  text-align: center;
  color: #FFD700;
  font-size: 20px;
  margin: 5px 0;
}

.route-point.delivery .point-icon {
  background: #f56c6c;
}

/* 订单底部 */
.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
}

.order-info {
  color: #666;
  font-size: 14px;
}

/* 客户信息 */
.customer-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
}

.customer-detail {
  flex: 1;
}

.customer-name {
  font-weight: bold;
  margin-bottom: 2px;
}

.customer-phone {
  color: #666;
  font-size: 12px;
}

/* 地图容器 */
.map-container {
  height: 120px;
  background: #f5f5f5;
  border-radius: 8px;
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 14px;
}

/* 订单操作 */
.order-actions {
  display: flex;
  gap: 10px;
}

.order-actions :deep(.el-button) {
  flex: 1;
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

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 40px 0;
  color: #999;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .header-section {
    margin: 5px;
    padding: 10px;
  }

  .quick-actions {
    margin: 10px 5px;
  }

  .order-management {
    margin: 10px 5px;
  }
}
</style>