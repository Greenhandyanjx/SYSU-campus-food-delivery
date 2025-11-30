<template>
  <div class="rider-stats">
    <!-- 顶部导航 -->
    <div class="header-nav">
      <div class="nav-left">
        <el-button type="text" @click="$router.back()">
          <i class="css-icon arrow-left"></i>
        </el-button>
      </div>
      <div class="nav-title">数据统计</div>
      <div class="nav-right">
        <el-button type="text" @click="refreshData" :loading="refreshing">
          <i class="css-icon refresh"></i>
        </el-button>
      </div>
    </div>

    <!-- 时间筛选 -->
    <div class="time-filter">
      <el-radio-group v-model="timeRange" @change="handleTimeChange">
        <el-radio-button label="today">今天</el-radio-button>
        <el-radio-button label="week">本周</el-radio-button>
        <el-radio-button label="month">本月</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 核心数据 -->
    <div class="core-stats">
      <div class="stat-card income">
        <div class="stat-icon">
          <i class="css-icon wallet"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">¥{{ coreData.totalIncome.toFixed(2) }}</div>
          <div class="stat-label">总收入</div>
          <div class="stat-change" :class="{ positive: coreData.incomeChange >= 0 }">
            {{ coreData.incomeChange >= 0 ? '+' : '' }}{{ coreData.incomeChange }}%
          </div>
        </div>
      </div>

      <div class="stat-card orders">
        <div class="stat-icon">
          <i class="css-icon document"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ coreData.totalOrders }}</div>
          <div class="stat-label">总订单</div>
          <div class="stat-change" :class="{ positive: coreData.ordersChange >= 0 }">
            {{ coreData.ordersChange >= 0 ? '+' : '' }}{{ coreData.ordersChange }}%
          </div>
        </div>
      </div>

      <div class="stat-card efficiency">
        <div class="stat-icon">
          <i class="css-icon data-analysis"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ coreData.efficiency.toFixed(1) }}单/时</div>
          <div class="stat-label">配送效率</div>
          <div class="stat-change" :class="{ positive: coreData.efficiencyChange >= 0 }">
            {{ coreData.efficiencyChange >= 0 ? '+' : '' }}{{ coreData.efficiencyChange }}%
          </div>
        </div>
      </div>

      <div class="stat-card rating">
        <div class="stat-icon">
          <i class="css-icon star"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ coreData.rating.toFixed(1) }}</div>
          <div class="stat-label">评分</div>
          <div class="stat-change" :class="{ positive: coreData.ratingChange >= 0 }">
            {{ coreData.ratingChange >= 0 ? '+' : '' }}{{ coreData.ratingChange.toFixed(1) }}
          </div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-section">
      <!-- 收入趋势图 -->
      <div class="chart-card">
        <div class="chart-header">
          <h3>收入趋势</h3>
          <el-radio-group v-model="chartType" size="small">
            <el-radio-button label="daily">日</el-radio-button>
            <el-radio-button label="weekly">周</el-radio-button>
            <el-radio-button label="monthly">月</el-radio-button>
          </el-radio-group>
        </div>
        <div class="chart-content">
          <div class="simple-chart">
            <div class="chart-bars">
              <div
                v-for="(item, index) in incomeData"
                :key="index"
                class="chart-bar"
                :style="{ height: `${(item.value / Math.max(...incomeData.map(d => d.value))) * 100}%` }"
              >
                <div class="bar-value">¥{{ item.value }}</div>
                <div class="bar-label">{{ item.label }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 订单分布图 -->
      <div class="chart-card">
        <div class="chart-header">
          <h3>订单分布</h3>
        </div>
        <div class="chart-content">
          <div class="distribution-chart">
            <div class="pie-segment completed" :style="{ width: `${orderDistribution.completed}%` }">
              <div class="segment-label">已完成</div>
              <div class="segment-value">{{ orderDistribution.completed }}%</div>
            </div>
            <div class="pie-segment cancelled" :style="{ width: `${orderDistribution.cancelled}%` }">
              <div class="segment-label">已取消</div>
              <div class="segment-value">{{ orderDistribution.cancelled }}%</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 详细统计 -->
    <div class="detailed-stats">
      <div class="stats-group">
        <h3>工作统计</h3>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-label">在线时长</div>
            <div class="stat-value">{{ workStats.onlineHours }}h</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">工作天数</div>
            <div class="stat-value">{{ workStats.workDays }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">平均配送时间</div>
            <div class="stat-value">{{ workStats.avgDeliveryTime }}分钟</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">平均配送距离</div>
            <div class="stat-value">{{ workStats.avgDistance }}km</div>
          </div>
        </div>
      </div>

      <div class="stats-group">
        <h3>评分统计</h3>
        <div class="rating-breakdown">
          <div class="rating-item">
            <span class="rating-label">5星</span>
            <div class="rating-bar">
              <div class="rating-fill" :style="{ width: `${ratingStats.fiveStar}%` }"></div>
            </div>
            <span class="rating-count">{{ ratingStats.fiveStarCount }}</span>
          </div>
          <div class="rating-item">
            <span class="rating-label">4星</span>
            <div class="rating-bar">
              <div class="rating-fill" :style="{ width: `${ratingStats.fourStar}%` }"></div>
            </div>
            <span class="rating-count">{{ ratingStats.fourStarCount }}</span>
          </div>
          <div class="rating-item">
            <span class="rating-label">3星</span>
            <div class="rating-bar">
              <div class="rating-fill" :style="{ width: `${ratingStats.threeStar}%` }"></div>
            </div>
            <span class="rating-count">{{ ratingStats.threeStarCount }}</span>
          </div>
          <div class="rating-item">
            <span class="rating-label">2星</span>
            <div class="rating-bar">
              <div class="rating-fill" :style="{ width: `${ratingStats.twoStar}%` }"></div>
            </div>
            <span class="rating-count">{{ ratingStats.twoStarCount }}</span>
          </div>
          <div class="rating-item">
            <span class="rating-label">1星</span>
            <div class="rating-bar">
              <div class="rating-fill" :style="{ width: `${ratingStats.oneStar}%` }"></div>
            </div>
            <span class="rating-count">{{ ratingStats.oneStarCount }}</span>
          </div>
        </div>
      </div>

      <div class="stats-group">
        <h3>效率分析</h3>
        <div class="efficiency-stats">
          <div class="efficiency-item">
            <div class="efficiency-icon">
              <i class="css-icon timer"></i>
            </div>
            <div class="efficiency-content">
              <div class="efficiency-title">准时率</div>
              <div class="efficiency-value">{{ efficiencyStats.onTimeRate }}%</div>
            </div>
          </div>
          <div class="efficiency-item">
            <div class="efficiency-icon">
              <i class="css-icon map"></i>
            </div>
            <div class="efficiency-content">
              <div class="efficiency-title">准时配送数</div>
              <div class="efficiency-value">{{ efficiencyStats.onTimeCount }}</div>
            </div>
          </div>
          <div class="efficiency-item">
            <div class="efficiency-icon">
              <i class="css-icon award"></i>
            </div>
            <div class="efficiency-content">
              <div class="efficiency-title">好评率</div>
              <div class="efficiency-value">{{ efficiencyStats.positiveRate }}%</div>
            </div>
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
      <div class="nav-item" @click="$router.push('/rider/orders')">
        <i class="css-icon list"></i>
        <span>订单</span>
      </div>
      <div class="nav-item active" @click="$router.push('/rider/stats')">
        <i class="css-icon data-analysis"></i>
        <span>统计</span>
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
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 状态数据
const loading = ref(false)
const refreshing = ref(false)
const timeRange = ref('today')
const chartType = ref('daily')

// 核心数据
const coreData = ref({
  totalIncome: 0,
  totalOrders: 0,
  efficiency: 0,
  rating: 0,
  incomeChange: 0,
  ordersChange: 0,
  efficiencyChange: 0,
  ratingChange: 0
})

// 收入数据
const incomeData = ref([])

// 订单分布
const orderDistribution = computed(() => {
  const total = (orderStats.value.completed || 0) + (orderStats.value.cancelled || 0)
  if (total === 0) {
    return { completed: 70, cancelled: 30 } // 默认值
  }
  return {
    completed: Math.round(((orderStats.value.completed || 0) / total) * 100),
    cancelled: Math.round(((orderStats.value.cancelled || 0) / total) * 100)
  }
})

// 工作统计
const workStats = ref({
  onlineHours: 0,
  workDays: 0,
  avgDeliveryTime: 0,
  avgDistance: 0
})

// 订单统计
const orderStats = ref({
  completed: 0,
  cancelled: 0
})

// 评分统计
const ratingStats = ref({
  fiveStar: 0,
  fourStar: 0,
  threeStar: 0,
  twoStar: 0,
  oneStar: 0,
  fiveStarCount: 0,
  fourStarCount: 0,
  threeStarCount: 0,
  twoStarCount: 0,
  oneStarCount: 0
})

// 效率统计
const efficiencyStats = ref({
  onTimeRate: 0,
  onTimeCount: 0,
  positiveRate: 0
})

// 初始化数据
const initData = async () => {
  try {
    loading.value = true

    // 获取月度统计
    const monthlyResponse = await riderApi.getMonthlyStats()
    if (monthlyResponse.code === 1 && monthlyResponse.data) {
      const data = monthlyResponse.data

      coreData.value.totalIncome = data.monthIncome || 0
      coreData.value.totalOrders = data.monthOrders || 0
      coreData.value.efficiency = data.efficiency || 0
      coreData.value.rating = data.rating || 0
      coreData.value.incomeChange = 15.2 // 模拟数据，后端暂未提供
      coreData.value.ordersChange = 8.5  // 模拟数据，后端暂未提供
      coreData.value.efficiencyChange = 5.3 // 模拟数据，后端暂未提供
      coreData.value.ratingChange = 0.2     // 模拟数据，后端暂未提供
    }

    // 获取工作统计
    const workResponse = await riderApi.getWorkStats({
      period: timeRange.value
    })
    if (workResponse.code === 1 && workResponse.data) {
      workStats.value = workResponse.data
    }

    // 获取收入统计
    const incomeResponse = await riderApi.getIncomeStats()
    if (incomeResponse.code === 1 && incomeResponse.data) {
      // 根据图表类型和时间范围处理数据
      const data = incomeResponse.data
      if (chartType.value === 'daily') {
        incomeData.value = [
          { label: '今天', value: data.dailyIncome || 0 },
          { label: '昨天', value: data.dailyIncome * 0.8 || 0 },
          { label: '前天', value: data.dailyIncome * 0.9 || 0 }
        ]
      } else if (chartType.value === 'weekly') {
        incomeData.value = [
          { label: '本周', value: data.weeklyIncome || 0 },
          { label: '上周', value: data.weeklyIncome * 0.85 || 0 },
          { label: '上上周', value: data.weeklyIncome * 0.9 || 0 }
        ]
      } else {
        incomeData.value = [
          { label: '本月', value: data.monthlyIncome || 0 },
          { label: '上月', value: data.monthlyIncome * 0.8 || 0 }
        ]
      }
    }

    // 获取订单统计 - 使用GetWorkStats API
    const orderResponse = await riderApi.getWorkStats({
      period: timeRange.value
    })
    if (orderResponse.code === 1 && orderResponse.data) {
      const data = orderResponse.data
      orderStats.value.completed = data.totalOrders || 0
      orderStats.value.cancelled = 0 // 后端当前没有取消订单统计，暂时设为0
    }

    // 获取评分统计
    const ratingResponse = await riderApi.getReviews({
      timeRange: timeRange.value
    })
    if (ratingResponse.code === 1 && ratingResponse.data) {
      ratingStats.value = ratingResponse.data
    }

    // 获取效率统计
    const efficiencyResponse = await riderApi.getWeeklyStats()
    if (efficiencyResponse.code === 1 && efficiencyResponse.data) {
      efficiencyStats.value = efficiencyResponse.data
    }

  } catch (error) {
    console.error('初始化统计数据失败:', error)
    // 使用Demo数据
    loadDemoData()
  } finally {
    loading.value = false
  }
}

// 加载Demo数据
const loadDemoData = () => {
  // 模拟核心数据
  coreData.value = {
    totalIncome: 1280.50,
    totalOrders: 125,
    efficiency: 2.8,
    rating: 4.8,
    incomeChange: 15.2,
    ordersChange: 8.5,
    efficiencyChange: 5.3,
    ratingChange: 0.2
  }

  // 模拟收入数据
  if (chartType.value === 'daily') {
    incomeData.value = [
      { label: '周一', value: 156 },
      { label: '周二', value: 189 },
      { label: '周三', value: 234 },
      { label: '周四', value: 178 },
      { label: '周五', value: 245 },
      { label: '周六', value: 167 },
      { label: '周日', value: 111 }
    ]
  } else if (chartType.value === 'weekly') {
    incomeData.value = [
      { label: '第1周', value: 856 },
      { label: '第2周', value: 1240 },
      { label: '第3周', value: 1089 },
      { label: '第4周', value: 1567 }
    ]
  } else {
    incomeData.value = [
      { label: '1月', value: 3200 },
      { label: '2月', value: 3890 },
      { label: '3月', value: 4156 },
      { label: '4月', value: 3789 }
    ]
  }

  // 模拟工作统计
  workStats.value = {
    onlineHours: 156,
    workDays: 25,
    avgDeliveryTime: 18,
    avgDistance: 1.2
  }

  // 模拟订单统计
  orderStats.value = {
    completed: 875,
    cancelled: 25,
    total: 900
  }

  // 模拟评分统计
  const totalRatings = 1200
  ratingStats.value = {
    fiveStar: 85,
    fourStar: 12,
    threeStar: 2,
    twoStar: 0.5,
    oneStar: 0.5,
    fiveStarCount: Math.round(totalRatings * 0.85),
    fourStarCount: Math.round(totalRatings * 0.12),
    threeStarCount: Math.round(totalRatings * 0.02),
    twoStarCount: Math.round(totalRatings * 0.005),
    oneStarCount: Math.round(totalRatings * 0.005)
  }

  // 模拟效率统计
  efficiencyStats.value = {
    onTimeRate: 95.2,
    onTimeCount: 834,
    positiveRate: 98.7
  }
}

// 刷新数据
const refreshData = async () => {
  try {
    refreshing.value = true
    await initData()
    ElMessage.success('刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败')
  } finally {
    refreshing.value = false
  }
}

// 时间范围变化
const handleTimeChange = () => {
  initData()
}

// 收入图表类型变化
const handleIncomeChartChange = () => {
  initData()
}

onMounted(() => {
  initData()
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

/* 箭头左 */
.css-icon.arrow-left::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  transform: translateY(-50%) rotate(-45deg);
  width: 10px;
  height: 10px;
  border-left: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
}

/* 刷新图标 */
.css-icon.refresh::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 50%;
  border-top-color: transparent;
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: translate(-50%, -50%) rotate(0deg); }
  to { transform: translate(-50%, -50%) rotate(360deg); }
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

/* 星星图标 */
.css-icon.star::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 12px;
  height: 12px;
  background: currentColor;
  clip-path: polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%);
}

/* 计时器图标 */
.css-icon.timer::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 12px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.timer::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 6px;
  height: 2px;
  background: currentColor;
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

/* 奖章图标 */
.css-icon.award::before {
  content: '★';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  color: currentColor;
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
  background: linear-gradient(to bottom, #FFFDE7, #FFFFFF);
  min-height: 100vh;
  padding-bottom: 60px;
  font-family: 'PingFang SC', 'Helvetica Neue', sans-serif;
}

/* 顶部导航 */
.header-nav {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #FFD700;
  color: #333;
}

.nav-title {
  font-size: 18px;
  font-weight: 600;
}

/* 时间筛选 */
.time-filter {
  padding: 15px;
  background: white;
  margin: 10px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  text-align: center;
}

/* 核心数据 */
.core-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  padding: 0 15px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 15px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: white;
  flex-shrink: 0;
}

.stat-card.income .stat-icon {
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
}

.stat-card.orders .stat-icon {
  background: linear-gradient(135deg, #409EFF 0%, #1890ff 100%);
}

.stat-card.efficiency .stat-icon {
  background: linear-gradient(135deg, #67C23A 0%, #52c41a 100%);
}

.stat-card.rating .stat-icon {
  background: linear-gradient(135deg, #E6A23C 0%, #ffa502 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 6px;
}

.stat-change {
  font-size: 12px;
  color: #f56c6c;
}

.stat-change.positive {
  color: #67c23a;
}

/* 图表区域 */
.charts-section {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
  padding: 0 15px;
}

.chart-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.chart-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.chart-content {
  height: 200px;
}

/* 柱状图 */
.simple-chart {
  width: 100%;
  height: 100%;
}

.chart-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  height: 100%;
  gap: 5px;
}

.chart-bar {
  flex: 1;
  max-width: 60px;
  background: linear-gradient(to top, #409EFF, #1890ff);
  border-radius: 4px 4px 0 0;
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  align-items: center;
  padding-top: 8px;
}

.bar-value {
  font-size: 12px;
  color: #333;
  font-weight: 500;
}

.bar-label {
  font-size: 11px;
  color: #666;
  margin-top: 4px;
}

/* 饼图 */
.distribution-chart {
  width: 100%;
  height: 60px;
  border-radius: 30px;
  overflow: hidden;
  display: flex;
  position: relative;
}

.pie-segment {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  font-size: 12px;
  color: white;
  font-weight: 500;
}

.pie-segment.completed {
  background: linear-gradient(135deg, #67C23A 0%, #52c41a 100%);
}

.pie-segment.cancelled {
  background: linear-gradient(135deg, #F56C6C 0%, #ff4757 100%);
}

.segment-label {
  position: absolute;
  left: 8px;
  font-size: 10px;
}

.segment-value {
  font-size: 11px;
  margin-left: 4px;
}

/* 详细统计 */
.detailed-stats {
  padding: 0 15px;
}

.stats-group {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stats-group h3 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #333;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}

.stat-item {
  text-align: center;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-item .stat-label {
  font-size: 12px;
  color: #666;
  margin-bottom: 8px;
  display: block;
}

.stat-item .stat-value {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

/* 评分统计 */
.rating-breakdown {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rating-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.rating-label {
  width: 30px;
  font-size: 13px;
  color: #666;
}

.rating-bar {
  flex: 1;
  height: 8px;
  background: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
}

.rating-fill {
  height: 100%;
  background: linear-gradient(90deg, #FFD700, #FFA500);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.rating-count {
  width: 30px;
  font-size: 13px;
  color: #666;
  text-align: right;
}

/* 效率统计 */
.efficiency-stats {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.efficiency-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
}

.efficiency-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #409EFF;
  color: white;
  font-size: 18px;
}

.efficiency-content {
  flex: 1;
}

.efficiency-title {
  font-size: 14px;
  color: #666;
  margin-bottom: 4px;
}

.efficiency-value {
  font-size: 16px;
  font-weight: 600;
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
  .core-stats {
    grid-template-columns: 1fr;
    gap: 8px;
    padding: 0 10px;
  }

  .stat-card {
    padding: 15px;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
    font-size: 20px;
  }

  .stat-value {
    font-size: 20px;
  }

  .charts-section {
    grid-template-columns: 1fr;
    gap: 8px;
    padding: 0 10px;
  }

  .chart-card {
    padding: 15px;
  }

  .detailed-stats {
    padding: 0 10px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .rating-item {
    flex-direction: column;
    gap: 5px;
    text-align: center;
  }

  .efficiency-stats {
    gap: 8px;
  }

  .efficiency-item {
    padding: 8px;
  }

  .efficiency-icon {
    width: 32px;
    height: 32px;
    font-size: 16px;
  }
}
</style>