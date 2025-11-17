<template>
  <div class="rider-stats">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">数据统计</h1>
      <div class="export-btn" @click="exportStats">
        <i class="css-icon download"></i>
      </div>
    </div>

    <!-- 时间选择器 -->
    <div class="time-selector">
      <div class="time-tabs">
        <div
          v-for="tab in timeTabs"
          :key="tab.value"
          class="time-tab"
          :class="{ active: activeTimeTab === tab.value }"
          @click="switchTimeTab(tab.value)"
        >
          {{ tab.label }}
        </div>
      </div>
      <div class="custom-date" v-if="activeTimeTab === 'custom'">
        <el-date-picker
          v-model="customDateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          @change="onCustomDateChange"
        />
      </div>
    </div>

    <!-- 核心指标卡片 -->
    <div class="metrics-cards">
      <div class="metric-card income">
        <div class="metric-icon">💰</div>
        <div class="metric-content">
          <div class="metric-value">¥{{ statsData.totalIncome.toFixed(2) }}</div>
          <div class="metric-label">总收入</div>
          <div class="metric-change" :class="getChangeClass(statsData.incomeChange)">
            {{ formatChange(statsData.incomeChange) }}
          </div>
        </div>
      </div>

      <div class="metric-card orders">
        <div class="metric-icon">📦</div>
        <div class="metric-content">
          <div class="metric-value">{{ statsData.totalOrders }}</div>
          <div class="metric-label">完成订单</div>
          <div class="metric-change" :class="getChangeClass(statsData.ordersChange)">
            {{ formatChange(statsData.ordersChange) }}
          </div>
        </div>
      </div>

      <div class="metric-card time">
        <div class="metric-icon">⏱️</div>
        <div class="metric-content">
          <div class="metric-value">{{ statsData.totalHours }}h</div>
          <div class="metric-label">工作时长</div>
          <div class="metric-change" :class="getChangeClass(statsData.hoursChange)">
            {{ formatChange(statsData.hoursChange) }}
          </div>
        </div>
      </div>

      <div class="metric-card rating">
        <div class="metric-icon">⭐</div>
        <div class="metric-content">
          <div class="metric-value">{{ statsData.avgRating }}</div>
          <div class="metric-label">平均评分</div>
          <div class="metric-change" :class="getChangeClass(statsData.ratingChange)">
            {{ formatChange(statsData.ratingChange) }}
          </div>
        </div>
      </div>
    </div>

    <!-- 收入趋势图表 -->
    <div class="chart-section">
      <div class="section-header">
        <h3>收入趋势</h3>
        <div class="chart-legend">
          <span class="legend-item">
            <span class="legend-color order-income"></span>
            订单收入
          </span>
          <span class="legend-item">
            <span class="legend-color bonus-income"></span>
            奖励收入
          </span>
        </div>
      </div>
      <div class="income-chart">
        <div class="chart-placeholder">
          <div class="chart-content">
            <div class="chart-bars">
              <div
                v-for="(item, index) in chartData"
                :key="index"
                class="chart-bar"
                :style="{ height: `${(item.total / maxChartValue) * 100}%` }"
              >
                <div class="bar-tooltip">
                  <div class="tooltip-content">
                    <div>{{ item.date }}</div>
                    <div>收入: ¥{{ item.total.toFixed(2) }}</div>
                    <div>订单: {{ item.orders }}单</div>
                  </div>
                </div>
              </div>
            </div>
            <div class="chart-labels">
              <span v-for="(item, index) in chartData" :key="index" class="chart-label">
                {{ formatChartLabel(item.date) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 详细统计 -->
    <div class="detail-stats">
      <div class="stats-grid">
        <!-- 订单统计 -->
        <div class="stats-card">
          <h4>订单统计</h4>
          <div class="stats-list">
            <div class="stat-item">
              <span class="stat-label">平均配送时间</span>
              <span class="stat-value">{{ statsData.avgDeliveryTime }}分钟</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">平均配送距离</span>
              <span class="stat-value">{{ statsData.avgDistance }}km</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">准时率</span>
              <span class="stat-value">{{ statsData.onTimeRate }}%</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">好评率</span>
              <span class="stat-value">{{ statsData.positiveRate }}%</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">接单率</span>
              <span class="stat-value">{{ statsData.acceptRate }}%</span>
            </div>
          </div>
        </div>

        <!-- 收入分析 -->
        <div class="stats-card">
          <h4>收入分析</h4>
          <div class="stats-list">
            <div class="stat-item">
              <span class="stat-label">订单收入</span>
              <span class="stat-value">¥{{ statsData.orderIncome.toFixed(2) }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">奖励收入</span>
              <span class="stat-value">¥{{ statsData.bonusIncome.toFixed(2) }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">平均单收入</span>
              <span class="stat-value">¥{{ statsData.avgOrderIncome.toFixed(2) }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">时收入</span>
              <span class="stat-value">¥{{ statsData.hourlyIncome.toFixed(2) }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">日均收入</span>
              <span class="stat-value">¥{{ statsData.dailyAvgIncome.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 工作效率 -->
      <div class="efficiency-section">
        <h4>工作效率</h4>
        <div class="efficiency-grid">
          <div class="efficiency-item">
            <div class="efficiency-label">在线时长</div>
            <div class="efficiency-value">{{ statsData.onlineHours }}小时</div>
            <div class="efficiency-progress">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: `${(statsData.onlineHours / 12) * 100}%` }"></div>
              </div>
              <span class="progress-text">{{ Math.round((statsData.onlineHours / 12) * 100) }}%</span>
            </div>
          </div>

          <div class="efficiency-item">
            <div class="efficiency-label">配送效率</div>
            <div class="efficiency-value">{{ statsData.deliveryEfficiency }}%</div>
            <div class="efficiency-progress">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: `${statsData.deliveryEfficiency}%` }"></div>
              </div>
              <span class="progress-text">{{ statsData.deliveryEfficiency }}%</span>
            </div>
          </div>

          <div class="efficiency-item">
            <div class="efficiency-label">客户满意度</div>
            <div class="efficiency-value">{{ statsData.customerSatisfaction }}%</div>
            <div class="efficiency-progress">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: `${statsData.customerSatisfaction}%` }"></div>
              </div>
              <span class="progress-text">{{ statsData.customerSatisfaction }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 排行榜 -->
      <div class="ranking-section">
        <h4>排行榜</h4>
        <div class="ranking-tabs">
          <div
            v-for="tab in rankingTabs"
            :key="tab.value"
            class="ranking-tab"
            :class="{ active: activeRankingTab === tab.value }"
            @click="switchRankingTab(tab.value)"
          >
            {{ tab.label }}
          </div>
        </div>

        <div class="ranking-list">
          <div
            v-for="(item, index) in rankingData"
            :key="item.id"
            class="ranking-item"
            :class="{ self: item.isSelf }"
          >
            <div class="rank-number" :class="getRankClass(index)">
              {{ index + 1 }}
            </div>
            <el-avatar :size="40" :src="item.avatar" />
            <div class="rider-info">
              <div class="rider-name">{{ item.name }}</div>
              <div class="rider-stats">{{ formatRankingValue(item) }}</div>
            </div>
            <div class="rank-value">{{ formatRankingDisplay(item) }}</div>
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
      <div class="nav-item" @click="$router.push('/rider/dashboard')">
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
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

// 状态管理
const loading = ref(false)
const activeTimeTab = ref('week')
const activeRankingTab = ref('income')
const customDateRange = ref([])

// 时间选择标签
const timeTabs = [
  { label: '今日', value: 'today' },
  { label: '本周', value: 'week' },
  { label: '本月', value: 'month' },
  { label: '自定义', value: 'custom' }
]

// 排行榜标签
const rankingTabs = [
  { label: '收入榜', value: 'income' },
  { label: '订单榜', value: 'orders' },
  { label: '效率榜', value: 'efficiency' },
  { label: '好评榜', value: 'rating' }
]

// 统计数据
const statsData = ref({
  totalIncome: 1280.50,
  incomeChange: 0.15,
  totalOrders: 68,
  ordersChange: 0.08,
  totalHours: 45,
  hoursChange: -0.05,
  avgRating: 4.8,
  ratingChange: 0.02,

  // 订单统计
  avgDeliveryTime: 18,
  avgDistance: 1.2,
  onTimeRate: 95,
  positiveRate: 98,
  acceptRate: 85,

  // 收入分析
  orderIncome: 1200.00,
  bonusIncome: 80.50,
  avgOrderIncome: 18.83,
  hourlyIncome: 28.46,
  dailyAvgIncome: 256.10,

  // 工作效率
  onlineHours: 6.5,
  deliveryEfficiency: 88,
  customerSatisfaction: 96
})

// 图表数据
const chartData = ref([
  { date: '11-11', orders: 8, total: 142.50 },
  { date: '11-12', orders: 10, total: 188.00 },
  { date: '11-13', orders: 7, total: 125.50 },
  { date: '11-14', orders: 12, total: 225.00 },
  { date: '11-15', orders: 9, total: 162.00 },
  { date: '11-16', orders: 11, total: 198.50 },
  { date: '11-17', orders: 11, total: 239.00 }
])

// 排行榜数据
const rankingData = ref([
  {
    id: 1,
    name: '王骑手',
    avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
    income: 3580.50,
    orders: 186,
    efficiency: 95,
    rating: 4.9
  },
  {
    id: 2,
    name: '李骑手',
    avatar: 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png',
    income: 3420.00,
    orders: 175,
    efficiency: 92,
    rating: 4.8,
    isSelf: true
  },
  {
    id: 3,
    name: '张骑手',
    avatar: 'https://cube.elemecdn.com/6/94/4d3ea53c4e4c9b5cc8b5c0b2e3e7dpng.png',
    income: 3280.75,
    orders: 168,
    efficiency: 90,
    rating: 4.7
  }
])

// 计算属性
const maxChartValue = computed(() => {
  return Math.max(...chartData.value.map(item => item.total))
})

// 方法定义
const getChangeClass = (change) => {
  if (change > 0) return 'positive'
  if (change < 0) return 'negative'
  return 'neutral'
}

const formatChange = (change) => {
  if (change > 0) return `↑ ${Math.abs(change * 100).toFixed(1)}%`
  if (change < 0) return `↓ ${Math.abs(change * 100).toFixed(1)}%`
  return '持平'
}

const formatChartLabel = (date) => {
  return date.slice(-2)
}

const getRankClass = (index) => {
  if (index === 0) return 'gold'
  if (index === 1) return 'silver'
  if (index === 2) return 'bronze'
  return ''
}

const formatRankingValue = (item) => {
  switch (activeRankingTab.value) {
    case 'income':
      return `${item.orders}单`
    case 'orders':
      return `¥${item.income.toFixed(2)}`
    case 'efficiency':
      return `${item.rating}分`
    case 'rating':
      return `${item.orders}单`
    default:
      return ''
  }
}

const formatRankingDisplay = (item) => {
  switch (activeRankingTab.value) {
    case 'income':
      return `¥${item.income.toFixed(2)}`
    case 'orders':
      return `${item.orders}单`
    case 'efficiency':
      return `${item.efficiency}%`
    case 'rating':
      return `${item.rating}分`
    default:
      return ''
  }
}

// 切换时间标签
const switchTimeTab = (tab) => {
  activeTimeTab.value = tab
  loadStatsData()
}

// 切换排行榜标签
const switchRankingTab = (tab) => {
  activeRankingTab.value = tab
  loadRankingData()
}

// 自定义日期变化
const onCustomDateChange = (dates) => {
  if (dates && dates.length === 2) {
    loadStatsData()
  }
}

// 加载统计数据
const loadStatsData = async () => {
  try {
    loading.value = true

    const params = {
      period: activeTimeTab.value,
      startDate: customDateRange.value[0] || '',
      endDate: customDateRange.value[1] || ''
    }

    // 调用API获取统计数据
    const response = await riderApi.getIncomeStats(params)

    if (response.code === 1) {
      // 更新统计数据
      statsData.value = {
        ...statsData.value,
        ...response.data
      }

      // 更新图表数据
      if (response.data.chartData) {
        chartData.value = response.data.chartData
      }
    }

  } catch (error) {
    console.error('加载统计数据失败:', error)
    // 使用Demo数据
  } finally {
    loading.value = false
  }
}

// 加载排行榜数据
const loadRankingData = async () => {
  try {
    // TODO: 调用排行榜API
    // const response = await riderApi.getRankingData(activeRankingTab.value)

    // Demo数据已在前面定义
  } catch (error) {
    console.error('加载排行榜数据失败:', error)
  }
}

// 导出统计
const exportStats = () => {
  ElMessage.info('数据导出功能开发中...')
}

onMounted(() => {
  loadStatsData()
  loadRankingData()
})
</script>

<style scoped>
/* CSS图标 */
.css-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  position: relative;
  font-size: inherit;
  color: inherit;
}

/* 返回图标 */
.css-icon.back::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-40%, -50%) rotate(-45deg);
  width: 10px;
  height: 10px;
  border-left: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
}

/* 下载图标 */
.css-icon.download::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.download::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 0;
  border-left: 6px solid transparent;
  border-right: 6px solid transparent;
  border-top: 8px solid currentColor;
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

.rider-stats {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 60px;
}

/* 顶部导航栏 */
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #FFD700;
  color: #333;
}

.back-btn, .export-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.3s ease;
}

.back-btn:hover, .export-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.back-btn .css-icon, .export-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 时间选择器 */
.time-selector {
  padding: 15px;
  background: white;
  border-bottom: 1px solid #f0f0f0;
}

.time-tabs {
  display: flex;
  background: #f5f5f5;
  border-radius: 20px;
  padding: 3px;
  margin-bottom: 10px;
}

.time-tab {
  flex: 1;
  padding: 8px 0;
  font-size: 14px;
  color: #666;
  text-align: center;
  cursor: pointer;
  border-radius: 17px;
  transition: all 0.3s ease;
}

.time-tab.active {
  background: #FFD700;
  color: white;
}

.custom-date {
  display: flex;
  justify-content: center;
}

/* 核心指标卡片 */
.metrics-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
  padding: 15px;
}

.metric-card {
  display: flex;
  align-items: center;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.metric-icon {
  font-size: 32px;
  margin-right: 15px;
}

.metric-content {
  flex: 1;
}

.metric-value {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 4px;
}

.metric-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 4px;
}

.metric-change {
  font-size: 12px;
  font-weight: 500;
}

.metric-change.positive {
  color: #67C23A;
}

.metric-change.negative {
  color: #F56C6C;
}

.metric-change.neutral {
  color: #999;
}

/* 图表部分 */
.chart-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.chart-legend {
  display: flex;
  gap: 15px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #666;
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}

.legend-color.order-income {
  background: #409EFF;
}

.legend-color.bonus-income {
  background: #E6A23C;
}

.income-chart {
  height: 200px;
}

.chart-placeholder {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-content {
  width: 100%;
  height: 100%;
  position: relative;
}

.chart-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  height: 150px;
  padding: 0 10px;
}

.chart-bar {
  flex: 1;
  max-width: 40px;
  background: linear-gradient(to top, #409EFF, #67C23A);
  border-radius: 4px 4px 0 0;
  margin: 0 2px;
  position: relative;
  cursor: pointer;
}

.bar-tooltip {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  opacity: 0;
  visibility: hidden;
  transition: all 0.3s ease;
  z-index: 10;
}

.chart-bar:hover .bar-tooltip {
  opacity: 1;
  visibility: visible;
}

.tooltip-content {
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
  white-space: nowrap;
}

.chart-labels {
  display: flex;
  justify-content: space-between;
  padding: 10px;
  font-size: 12px;
  color: #666;
}

/* 详细统计 */
.detail-stats {
  padding: 0 15px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
  margin-bottom: 15px;
}

.stats-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stats-card h4 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #333;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.stats-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

.stat-value {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

/* 工作效率 */
.efficiency-section {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.efficiency-section h4 {
  margin: 0 0 20px 0;
  font-size: 16px;
  color: #333;
}

.efficiency-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.efficiency-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.efficiency-label {
  font-size: 14px;
  color: #666;
}

.efficiency-value {
  font-size: 18px;
  font-weight: bold;
  color: #FFD700;
}

.efficiency-progress {
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-bar {
  flex: 1;
  height: 8px;
  background: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #FFD700, #FFA500);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: #666;
  min-width: 35px;
}

/* 排行榜 */
.ranking-section {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.ranking-section h4 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #333;
}

.ranking-tabs {
  display: flex;
  background: #f5f5f5;
  border-radius: 8px;
  padding: 4px;
  margin-bottom: 20px;
}

.ranking-tab {
  flex: 1;
  padding: 8px 0;
  font-size: 14px;
  color: #666;
  text-align: center;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.ranking-tab.active {
  background: #FFD700;
  color: white;
}

.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  background: #f8f9fa;
  transition: all 0.3s ease;
}

.ranking-item.self {
  background: linear-gradient(135deg, #fff9e6, #fff7e6);
  border: 1px solid #FFD700;
}

.rank-number {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: bold;
  background: #e9ecef;
  color: #666;
}

.rank-number.gold {
  background: linear-gradient(135deg, #FFD700, #FFA500);
  color: white;
}

.rank-number.silver {
  background: linear-gradient(135deg, #C0C0C0, #808080);
  color: white;
}

.rank-number.bronze {
  background: linear-gradient(135deg, #CD7F32, #8B4513);
  color: white;
}

.rider-info {
  flex: 1;
}

.rider-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 2px;
}

.rider-stats {
  font-size: 12px;
  color: #666;
}

.rank-value {
  font-size: 16px;
  font-weight: bold;
  color: #FFD700;
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
  .metrics-cards {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .ranking-tabs {
    flex-wrap: wrap;
  }

  .ranking-tab {
    flex: 1;
    min-width: 80px;
  }
}
</style>