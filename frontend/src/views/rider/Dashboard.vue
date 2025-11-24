<template>
  <div class="rider-workbench">
    <!-- 顶部状态栏 -->
    <div class="status-bar">
      <div class="time-info">{{ currentTime }}</div>
      <div class="rider-info">
        <el-avatar :size="30" :src="riderInfo.avatar" />
        <span class="rider-name">{{ riderInfo.name }}</span>
      </div>
    </div>

    <!-- 工作台内容 -->
    <div class="workbench-content">
      <!-- 工作状态控制 -->
      <div class="work-status-card">
        <div class="status-header">
          <h3>工作状态</h3>
          <el-switch
            v-model="isOnline"
            active-color="#FFD700"
            inactive-color="#DCDFE6"
            :active-text="isOnline ? '在线接单' : '已下线'"
            @change="toggleOnlineStatus"
          />
        </div>
        <div class="status-stats">
          <div class="stat-item">
            <div class="stat-value">{{ onlineHours }}</div>
            <div class="stat-label">在线时长(小时)</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ todayOrders }}</div>
            <div class="stat-label">今日订单</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ completedOrders }}</div>
            <div class="stat-label">已完成</div>
          </div>
        </div>
      </div>

      <!-- 收入概览 -->
      <div class="income-overview">
        <h3>收入概览</h3>
        <div class="income-cards">
          <div class="income-card">
            <div class="card-icon">💰</div>
            <div class="card-content">
              <div class="card-value">¥{{ todayIncome.toFixed(2) }}</div>
              <div class="card-label">今日收入</div>
            </div>
          </div>
          <div class="income-card">
            <div class="card-icon">📈</div>
            <div class="card-content">
              <div class="card-value">¥{{ weekIncome.toFixed(2) }}</div>
              <div class="card-label">本周收入</div>
            </div>
          </div>
          <div class="income-card">
            <div class="card-icon">🎯</div>
            <div class="card-content">
              <div class="card-value">{{ monthOrders }}</div>
              <div class="card-label">本月订单</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 订单统计图表 -->
      <div class="order-stats">
        <h3>订单统计</h3>
        <div class="stats-grid">
          <div class="stats-item">
            <div class="stats-label">待接单</div>
            <div class="stats-value pending">{{ pendingOrders }}</div>
          </div>
          <div class="stats-item">
            <div class="stats-label">待取货</div>
            <div class="stats-value pickup">{{ pickupOrders }}</div>
          </div>
          <div class="stats-item">
            <div class="stats-label">配送中</div>
            <div class="stats-value delivering">{{ deliveringOrders }}</div>
          </div>
          <div class="stats-item">
            <div class="stats-label">已完成</div>
            <div class="stats-value completed">{{ completedOrders }}</div>
          </div>
        </div>
      </div>

      <!-- 快捷操作 -->
      <div class="quick-actions">
        <h3>快捷操作</h3>
        <div class="action-grid">
          <div class="action-item" @click="refreshData">
            <div class="action-icon">🔄</div>
            <div class="action-label">刷新数据</div>
          </div>
          <div class="action-item" @click="viewOrders">
            <div class="action-icon">📋</div>
            <div class="action-label">查看订单</div>
          </div>
          <div class="action-item" @click="viewWallet">
            <div class="action-icon">💳</div>
            <div class="action-label">我的钱包</div>
          </div>
          <div class="action-item" @click="viewStats">
            <div class="action-icon">📊</div>
            <div class="action-label">数据统计</div>
          </div>
        </div>
      </div>

      <!-- 今日概况 -->
      <div class="today-summary">
        <h3>今日概况</h3>
        <div class="summary-list">
          <div class="summary-item">
            <span class="summary-label">平均配送时间</span>
            <span class="summary-value">{{ avgDeliveryTime }}分钟</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">平均配送距离</span>
            <span class="summary-value">{{ avgDistance }}km</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">好评率</span>
            <span class="summary-value">{{ positiveRate }}%</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">准时率</span>
            <span class="summary-value">{{ onTimeRate }}%</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部导航栏 -->
    <div class="bottom-nav">
      <div class="nav-item" @click="$router.push('/rider')">
        <i class="css-icon house"></i>
        <span>首页</span>
      </div>
      <div class="nav-item active" @click="$router.push('/rider/dashboard')">
        <i class="css-icon data-analysis"></i>
        <span>工作台</span>
      </div>
      <div class="nav-item" @click="$router.push('/rider/orders')">
        <i class="css-icon list"></i>
        <span>订单</span>
      </div>
      <div class="nav-item" @click="$router.push('/rider/profile')">
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

const router = useRouter()

// 状态数据
const currentTime = ref('')
const isOnline = ref(true)
const loading = ref(false)

// 骑手信息
const riderInfo = ref({
  name: '李骑手',
  avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
  completedOrders: 1250
})

// 工作统计
const onlineHours = ref(6.5)
const todayOrders = ref(12)
const completedOrders = ref(8)

// 收入数据
const todayIncome = ref(185.5)
const weekIncome = ref(1280.0)
const monthOrders = ref(186)

// 订单统计
const pendingOrders = ref(3)
const pickupOrders = ref(2)
const deliveringOrders = ref(1)

// 绩效指标
const avgDeliveryTime = ref(18)
const avgDistance = ref(1.2)
const positiveRate = ref(98)
const onTimeRate = ref(95)

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
    const riderData = await riderApi.getRiderInfoWithDemo()
    if (riderData.code === 1 && riderData.data) {
      riderInfo.value = riderData.data
      isOnline.value = riderData.data.isOnline
      completedOrders.value = riderData.data.completedOrders || 0
    }

    // 获取收入统计
    const incomeData = await riderApi.getIncomeStatsWithDemo()
    if (incomeData.code === 1 && incomeData.data) {
      todayIncome.value = incomeData.data.dailyIncome || 0
      weekIncome.value = incomeData.data.weeklyIncome || 0
    }

    // 获取订单统计
    // TODO: 实现获取各状态订单数量的接口

  } catch (error) {
    console.error('初始化工作台数据失败:', error)
    ElMessage.error('获取数据失败，请刷新重试')
  } finally {
    loading.value = false
  }
}

// 切换在线状态
const toggleOnlineStatus = async (status) => {
  try {
    await riderApi.updateOnlineStatus(status)
    ElMessage.success(status ? '已上线，开始接单' : '已下线，停止接单')
  } catch (error) {
    ElMessage.error('状态更新失败，请重试')
    isOnline.value = !status
  }
}

// 刷新数据
const refreshData = () => {
  ElMessage.info('正在刷新数据...')
  initData()
}

// 查看订单
const viewOrders = () => {
  router.push('/rider/orders')
}

// 查看钱包
const viewWallet = () => {
  router.push('/rider/wallet')
}

// 查看统计
const viewStats = () => {
  router.push('/rider/stats')
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
  initData()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
/* 全局样式，确保CSS图标能正常工作 */
.css-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  position: relative;
  font-size: inherit;
  color: inherit;
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

.rider-workbench {
  background: linear-gradient(to bottom, #FFFDE7, #FFFFFF);
  min-height: 100vh;
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

.rider-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rider-name {
  font-weight: 500;
}

/* 工作台内容 */
.workbench-content {
  padding: 15px;
}

/* 工作状态卡片 */
.work-status-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.status-header h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
}

.status-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 15px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #FFD700;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

/* 收入概览 */
.income-overview {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.income-overview h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: #333;
}

.income-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 15px;
}

.income-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 15px;
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
  border-radius: 10px;
  color: white;
}

.card-icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.card-value {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 4px;
}

.card-label {
  font-size: 12px;
  opacity: 0.9;
}

/* 订单统计 */
.order-stats {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.order-stats h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: #333;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
}

.stats-item {
  text-align: center;
  padding: 10px;
  border-radius: 8px;
  background: #f8f9fa;
}

.stats-label {
  font-size: 12px;
  color: #666;
  margin-bottom: 5px;
}

.stats-value {
  font-size: 20px;
  font-weight: bold;
}

.stats-value.pending {
  color: #E6A23C;
}

.stats-value.pickup {
  color: #409EFF;
}

.stats-value.delivering {
  color: #F56C6C;
}

.stats-value.completed {
  color: #67C23A;
}

/* 快捷操作 */
.quick-actions {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.quick-actions h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: #333;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 15px;
  border-radius: 8px;
  background: #f8f9fa;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-item:hover {
  background: #e9ecef;
  transform: translateY(-2px);
}

.action-icon {
  font-size: 24px;
}

.action-label {
  font-size: 12px;
  color: #666;
}

/* 今日概况 */
.today-summary {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.today-summary h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: #333;
}

.summary-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 6px;
}

.summary-label {
  font-size: 14px;
  color: #666;
}

.summary-value {
  font-size: 16px;
  font-weight: bold;
  color: #333;
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
  .workbench-content {
    padding: 10px;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .action-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .summary-list {
    grid-template-columns: 1fr;
  }
}
</style>