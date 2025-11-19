<template>
  <div class="rider-orders">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">历史订单</h1>
      <div class="filter-btn" @click="showFilterDialog = true">
        <i class="css-icon filter"></i>
      </div>
    </div>

    <!-- 筛选条件 -->
    <div class="filter-section" v-if="activeFilters.length > 0">
      <div class="filter-tags">
        <el-tag
          v-for="filter in activeFilters"
          :key="filter.key"
          closable
          @close="removeFilter(filter.key)"
          class="filter-tag"
        >
          {{ filter.label }}
        </el-tag>
      </div>
      <el-button type="text" @click="clearAllFilters" class="clear-filters">清除全部</el-button>
    </div>

    <!-- 订单统计卡片 -->
    <div class="stats-cards">
      <div class="stat-card">
        <div class="stat-icon completed"></div>
        <div class="stat-content">
          <div class="stat-value">{{ orderStats.completed }}</div>
          <div class="stat-label">已完成</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon cancelled"></div>
        <div class="stat-content">
          <div class="stat-value">{{ orderStats.cancelled }}</div>
          <div class="stat-label">已取消</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon total-income"></div>
        <div class="stat-content">
          <div class="stat-value">¥{{ orderStats.totalIncome.toFixed(2) }}</div>
          <div class="stat-label">总收入</div>
        </div>
      </div>
    </div>

    <!-- 订单列表 -->
    <div class="order-list">
      <div v-if="loading" class="loading-container">
        <el-loading-directive />
      </div>

      <div v-else-if="orders.length === 0" class="empty-state">
        <el-empty description="暂无历史订单" />
      </div>

      <div v-else>
        <div
          v-for="order in orders"
          :key="order.id"
          class="order-card"
          @click="viewOrderDetail(order)"
        >
          <!-- 订单头部 -->
          <div class="order-header">
            <div class="order-info">
              <span class="order-no">订单号：{{ order.orderNo }}</span>
              <el-tag :type="getStatusType(order.status)" class="order-status">
                {{ getStatusText(order.status) }}
              </el-tag>
            </div>
            <div class="order-time">{{ formatDateTime(order.completedAt) }}</div>
          </div>

          <!-- 配送路线 -->
          <div class="delivery-route">
            <div class="route-point pickup">
              <div class="point-icon">
                <i class="css-icon shop"></i>
              </div>
              <div class="point-info">
                <div class="point-name">{{ order.restaurant }}</div>
                <div class="point-address">{{ order.pickupAddress }}</div>
              </div>
            </div>
            <div class="route-arrow">→</div>
            <div class="route-point delivery">
              <div class="point-icon">
                <i class="css-icon location"></i>
              </div>
              <div class="point-info">
                <div class="point-name">{{ order.customer }}</div>
                <div class="point-address">{{ order.deliveryAddress }}</div>
              </div>
            </div>
          </div>

          <!-- 订单信息 -->
          <div class="order-details">
            <div class="detail-item">
              <span class="detail-label">配送距离：</span>
              <span class="detail-value">{{ order.distance }}km</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">配送时长：</span>
              <span class="detail-value">{{ order.deliveryTime }}分钟</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">配送收入：</span>
              <span class="detail-value amount">¥{{ order.fee.toFixed(2) }}</span>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="order-actions">
            <el-button size="small" @click.stop="viewOrderDetail(order)">
              查看详情
            </el-button>
            <el-button v-if="order.status === 'completed'" size="small" type="primary" @click.stop="contactCustomer(order)">
              联系顾客
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载更多 -->
    <div v-if="hasMore && !loading" class="load-more">
      <el-button @click="loadMore" :loading="loadingMore">加载更多</el-button>
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
      <div class="nav-item active" @click="$router.push('/rider/orders')">
        <i class="css-icon list"></i>
        <span>订单</span>
      </div>
      <div class="nav-item" @click="$router.push('/rider/profile')">
        <i class="css-icon user"></i>
        <span>我的</span>
      </div>
    </div>

    <!-- 筛选弹窗 -->
    <el-dialog
      v-model="showFilterDialog"
      title="筛选条件"
      width="80%"
      :before-close="handleFilterClose"
    >
      <div class="filter-dialog-content">
        <div class="filter-group">
          <h4>订单状态</h4>
          <el-checkbox-group v-model="tempFilters.status">
            <el-checkbox label="completed">已完成</el-checkbox>
            <el-checkbox label="cancelled">已取消</el-checkbox>
            <el-checkbox label="timeout">超时订单</el-checkbox>
          </el-checkbox-group>
        </div>

        <div class="filter-group">
          <h4>时间范围</h4>
          <el-radio-group v-model="tempFilters.timeRange">
            <el-radio label="today">今天</el-radio>
            <el-radio label="week">本周</el-radio>
            <el-radio label="month">本月</el-radio>
            <el-radio label="custom">自定义</el-radio>
          </el-radio-group>

          <div v-if="tempFilters.timeRange === 'custom'" class="date-range">
            <el-date-picker
              v-model="tempFilters.dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
            />
          </div>
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="resetFilters">重置</el-button>
          <el-button type="primary" @click="applyFilters">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 订单详情弹窗 -->
    <el-dialog
      v-model="showOrderDetail"
      title="订单详情"
      width="90%"
      :before-close="handleDetailClose"
    >
      <div v-if="selectedOrder" class="order-detail-content">
        <div class="detail-section">
          <h4>基本信息</h4>
          <div class="detail-row">
            <span class="label">订单号：</span>
            <span class="value">{{ selectedOrder.orderNo }}</span>
          </div>
          <div class="detail-row">
            <span class="label">订单状态：</span>
            <el-tag :type="getStatusType(selectedOrder.status)">
              {{ getStatusText(selectedOrder.status) }}
            </el-tag>
          </div>
          <div class="detail-row">
            <span class="label">完成时间：</span>
            <span class="value">{{ formatDateTime(selectedOrder.completedAt) }}</span>
          </div>
        </div>

        <div class="detail-section">
          <h4>配送信息</h4>
          <div class="detail-row">
            <span class="label">取餐地址：</span>
            <span class="value">{{ selectedOrder.pickupAddress }}</span>
          </div>
          <div class="detail-row">
            <span class="label">送达地址：</span>
            <span class="value">{{ selectedOrder.deliveryAddress }}</span>
          </div>
          <div class="detail-row">
            <span class="label">配送距离：</span>
            <span class="value">{{ selectedOrder.distance }}km</span>
          </div>
          <div class="detail-row">
            <span class="label">配送时长：</span>
            <span class="value">{{ selectedOrder.deliveryTime }}分钟</span>
          </div>
        </div>

        <div class="detail-section">
          <h4>费用信息</h4>
          <div class="detail-row">
            <span class="label">配送收入：</span>
            <span class="value amount">¥{{ selectedOrder.fee.toFixed(2) }}</span>
          </div>
          <div class="detail-row" v-if="selectedOrder.bonus">
            <span class="label">奖励金额：</span>
            <span class="value bonus">+¥{{ selectedOrder.bonus.toFixed(2) }}</span>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

// 状态管理
const loading = ref(false)
const loadingMore = ref(false)
const orders = ref([])
const page = ref(1)
const pageSize = ref(10)
const hasMore = ref(true)
const total = ref(0)

// 筛选相关
const showFilterDialog = ref(false)
const showOrderDetail = ref(false)
const selectedOrder = ref(null)

const filters = ref({
  status: [],
  timeRange: 'week',
  dateRange: [],
  startDate: '',
  endDate: ''
})

const tempFilters = ref({
  status: [],
  timeRange: 'week',
  dateRange: []
})

// 订单统计
const orderStats = ref({
  completed: 128,
  cancelled: 3,
  totalIncome: 856.50
})

// 计算属性
const activeFilters = computed(() => {
  const active = []
  if (filters.value.status.length > 0) {
    active.push({ key: 'status', label: `状态: ${filters.value.status.join(', ')}` })
  }
  if (filters.value.timeRange !== 'week') {
    active.push({ key: 'timeRange', label: `时间: ${getTimeRangeText(filters.value.timeRange)}` })
  }
  return active
})

// 方法定义
const getStatusType = (status) => {
  const typeMap = {
    completed: 'success',
    cancelled: 'danger',
    timeout: 'warning'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    completed: '已完成',
    cancelled: '已取消',
    timeout: '超时订单'
  }
  return textMap[status] || '未知状态'
}

const getTimeRangeText = (range) => {
  const textMap = {
    today: '今天',
    week: '本周',
    month: '本月',
    custom: '自定义'
  }
  return textMap[range] || range
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  const date = new Date(dateTime)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 加载订单列表
const loadOrders = async (isLoadMore = false) => {
  try {
    if (isLoadMore) {
      loadingMore.value = true
    } else {
      loading.value = true
      page.value = 1
    }

    const params = {
      page: page.value,
      size: pageSize.value,
      status: filters.value.status.join(','),
      date: filters.value.startDate || ''
    }

    // 调用API获取订单列表
    const response = await riderApi.getOrderHistory(params)

    if (response.code === 1) {
      const newOrders = response.data.items || []

      if (isLoadMore) {
        orders.value = [...orders.value, ...newOrders]
      } else {
        orders.value = newOrders
      }

      total.value = response.data.total || 0
      hasMore.value = orders.value.length < total.value

      if (isLoadMore) {
        page.value++
      }
    }

  } catch (error) {
    console.error('加载订单列表失败:', error)
    ElMessage.error('加载订单失败，请重试')

    // Demo数据
    if (!isLoadMore) {
      orders.value = getDemoOrders()
      total.value = orders.value.length
      hasMore.value = false
    }
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

// Demo数据
const getDemoOrders = () => {
  return [
    {
      id: 'order001',
      orderNo: 'RD20241117001',
      status: 'completed',
      restaurant: '麦当劳',
      pickupAddress: '珠海市香洲区唐家湾大学路1号',
      customer: '张同学',
      deliveryAddress: '珠海市香洲区中山大学珠海校区榕园',
      distance: 1.2,
      deliveryTime: 18,
      fee: 6.5,
      bonus: 1.0,
      completedAt: '2024-11-17T14:30:00'
    },
    {
      id: 'order002',
      orderNo: 'RD20241117002',
      status: 'completed',
      restaurant: '肯德基',
      pickupAddress: '珠海市香洲区唐家湾大学路101号',
      customer: '王老师',
      deliveryAddress: '珠海市香洲区中山大学珠海校区荔园',
      distance: 0.8,
      deliveryTime: 15,
      fee: 5.0,
      completedAt: '2024-11-17T13:45:00'
    },
    {
      id: 'order003',
      orderNo: 'RD20241117003',
      status: 'cancelled',
      restaurant: '星巴克',
      pickupAddress: '珠海市香洲区唐家湾大学路201号',
      customer: '李同学',
      deliveryAddress: '珠海市香洲区中山大学珠海校区翰林',
      distance: 1.5,
      deliveryTime: 0,
      fee: 7.0,
      completedAt: '2024-11-17T12:20:00'
    }
  ]
}

// 加载更多
const loadMore = () => {
  if (!hasMore.value || loadingMore.value) return
  page.value++
  loadOrders(true)
}

// 查看订单详情
const viewOrderDetail = (order) => {
  selectedOrder.value = order
  showOrderDetail.value = true
}

// 联系顾客
const contactCustomer = (order) => {
  ElMessage.info(`正在联系顾客：${order.customer}`)
  // 这里可以集成实际的电话拨打功能
}

// 筛选相关方法
const applyFilters = () => {
  filters.value = { ...tempFilters.value }
  if (filters.value.dateRange && filters.value.dateRange.length === 2) {
    filters.value.startDate = filters.value.dateRange[0]
    filters.value.endDate = filters.value.dateRange[1]
  }
  showFilterDialog.value = false
  loadOrders()
}

const resetFilters = () => {
  tempFilters.value = {
    status: [],
    timeRange: 'week',
    dateRange: []
  }
}

const clearAllFilters = () => {
  filters.value = {
    status: [],
    timeRange: 'week',
    dateRange: [],
    startDate: '',
    endDate: ''
  }
  loadOrders()
}

const removeFilter = (key) => {
  if (key === 'status') {
    filters.value.status = []
    tempFilters.value.status = []
  } else if (key === 'timeRange') {
    filters.value.timeRange = 'week'
    tempFilters.value.timeRange = 'week'
    filters.value.dateRange = []
    filters.value.startDate = ''
    filters.value.endDate = ''
    tempFilters.value.dateRange = []
  }
  loadOrders()
}

const handleFilterClose = () => {
  tempFilters.value = { ...filters.value }
  showFilterDialog.value = false
}

const handleDetailClose = () => {
  selectedOrder.value = null
  showOrderDetail.value = false
}

onMounted(() => {
  loadOrders()
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

/* 筛选图标 */
.css-icon.filter::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 0 0 12px 12px;
  border-top: none;
}

.css-icon.filter::after {
  content: '';
  position: absolute;
  top: 6px;
  left: 2px;
  width: 12px;
  height: 2px;
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

.rider-orders {
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

.back-btn, .filter-btn {
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

.back-btn:hover, .filter-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.back-btn .css-icon, .filter-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 筛选条件 */
.filter-section {
  display: flex;
  align-items: center;
  padding: 10px 15px;
  background: white;
  border-bottom: 1px solid #f0f0f0;
}

.filter-tags {
  display: flex;
  gap: 8px;
  flex: 1;
}

.filter-tag {
  font-size: 12px;
}

.clear-filters {
  font-size: 12px;
  color: #999;
  padding: 0;
  margin-left: 10px;
}

/* 统计卡片 */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  padding: 15px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.stat-icon.completed {
  background: #67C23A;
  color: white;
}

.stat-icon.cancelled {
  background: #F56C6C;
  color: white;
}

.stat-icon.total-income {
  background: #FFD700;
  color: white;
}

.stat-value {
  font-size: 18px;
  font-weight: bold;
  color: #333;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

/* 订单列表 */
.order-list {
  padding: 0 15px;
}

.loading-container {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.empty-state {
  padding: 40px 0;
}

.order-card {
  background: white;
  border-radius: 12px;
  padding: 15px;
  margin-bottom: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s ease;
}

.order-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* 订单头部 */
.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.order-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.order-no {
  font-size: 14px;
  color: #666;
}

.order-time {
  font-size: 12px;
  color: #999;
}

/* 配送路线 */
.delivery-route {
  margin-bottom: 15px;
}

.route-point {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 8px;
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

.route-point.delivery .point-icon {
  background: #F56C6C;
}

.point-name {
  font-weight: 500;
  margin-bottom: 2px;
}

.point-address {
  font-size: 12px;
  color: #666;
  line-height: 1.4;
}

.route-arrow {
  text-align: center;
  color: #FFD700;
  font-size: 16px;
  margin: 4px 0;
  padding-left: 21px;
}

/* 订单信息 */
.order-details {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-bottom: 15px;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
}

.detail-item {
  text-align: center;
}

.detail-label {
  font-size: 12px;
  color: #666;
  display: block;
  margin-bottom: 4px;
}

.detail-value {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.detail-value.amount {
  color: #67C23A;
}

/* 操作按钮 */
.order-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

/* 加载更多 */
.load-more {
  text-align: center;
  padding: 20px 0;
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

/* 筛选弹窗 */
.filter-dialog-content {
  padding: 10px 0;
}

.filter-group {
  margin-bottom: 20px;
}

.filter-group h4 {
  margin: 0 0 10px 0;
  font-size: 16px;
  color: #333;
}

.date-range {
  margin-top: 10px;
}

/* 订单详情弹窗 */
.order-detail-content {
  padding: 10px 0;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section h4 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #333;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.detail-row {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.detail-row .label {
  width: 100px;
  font-size: 14px;
  color: #666;
  flex-shrink: 0;
}

.detail-row .value {
  font-size: 14px;
  color: #333;
  flex: 1;
}

.detail-row .value.amount {
  color: #67C23A;
  font-weight: 500;
}

.detail-row .value.bonus {
  color: #E6A23C;
  font-weight: 500;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .stats-cards {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .order-details {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .detail-item {
    text-align: left;
  }
}
</style>