<template>
  <div class="rider-orders">
    <!-- 顶部导航 -->
    <div class="header-nav">
      <div class="nav-left">
        <el-button type="text" @click="$router.back()">
          <i class="css-icon arrow-left"></i>
        </el-button>
      </div>
      <div class="nav-title">历史订单</div>
      <div class="nav-right">
        <el-button type="text" @click="refreshOrders" :loading="refreshing">
          <i class="css-icon refresh"></i>
        </el-button>
      </div>
    </div>

    <!-- 订单统计 -->
    <div class="order-stats">
      <div class="stat-item completed">
        <div class="stat-icon">
          <i class="css-icon success"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ orderStats.completed }}</div>
          <div class="stat-label">已完成</div>
        </div>
      </div>
      <div class="stat-item cancelled">
        <div class="stat-icon">
          <i class="css-icon close"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ orderStats.cancelled }}</div>
          <div class="stat-label">已取消</div>
        </div>
      </div>
      <div class="stat-item income">
        <div class="stat-icon">
          <i class="css-icon wallet"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">¥{{ orderStats.totalIncome.toFixed(2) }}</div>
          <div class="stat-label">总收入</div>
        </div>
      </div>
      <div class="stat-item efficiency">
        <div class="stat-icon">
          <i class="css-icon data-analysis"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ orderStats.efficiency }}单/时</div>
          <div class="stat-label">效率</div>
        </div>
      </div>
    </div>

    <!-- 订单筛选 -->
    <div class="filter-section">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange" class="order-tabs">
        <el-tab-pane label="全部" name="all">
          <div class="filter-bar">
            <el-select v-model="dateFilter" placeholder="时间筛选" size="small" @change="handleDateChange">
              <el-option label="今天" value="today" />
              <el-option label="昨天" value="yesterday" />
              <el-option label="本周" value="week" />
              <el-option label="本月" value="month" />
            </el-select>
            <el-select v-model="statusFilter" placeholder="状态筛选" size="small" @change="handleStatusChange">
              <el-option label="全部状态" value="" />
              <el-option label="已完成" value="completed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </div>
        </el-tab-pane>

        <el-tab-pane label="已完成" name="completed">
          <div class="tab-content"></div>
        </el-tab-pane>

        <el-tab-pane label="已取消" name="cancelled">
          <div class="tab-content"></div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 订单列表 -->
    <div class="order-list" v-loading="loading">
      <div v-if="filteredOrders.length === 0" class="empty-state">
        <el-empty description="暂无订单数据" />
      </div>

      <div v-else>
        <div
          v-for="order in filteredOrders"
          :key="order.id"
          class="order-item"
          @click="goToOrderDetail(order.id)"
        >
          <div class="order-header">
            <div class="order-info">
              <span class="order-id">订单号：{{ order.orderNo }}</span>
              <span class="order-time">{{ formatTime(order.createdAt) }}</span>
            </div>
            <el-tag
              :type="getStatusType(order.status)"
              size="small"
            >
              {{ getStatusText(order.status) }}
            </el-tag>
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
                  <div class="point-name">{{ order.customer || '顾客' }}</div>
                  <div class="point-address">{{ order.deliveryAddress }}</div>
                </div>
              </div>
            </div>

            <div class="order-details">
              <div class="detail-item">
                <span class="detail-label">距离：</span>
                <span class="detail-value">{{ order.distance }}km</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">时长：</span>
                <span class="detail-value">{{ order.deliveryTime }}分钟</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">收入：</span>
                <span class="detail-value amount">¥{{ order.fee.toFixed(2) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载更多 -->
    <div class="load-more" v-if="hasMore && !loading">
      <el-button type="text" @click="loadMore" :loading="loadingMore">
        加载更多
      </el-button>
    </div>

    <!-- 底部导航栏 -->
    <div class="bottom-nav">
      <div class="nav-item" @click="$router.push('/rider')">
        <i class="css-icon house"></i>
        <span>首页</span>
      </div>
      <div class="nav-item active" @click="$router.push('/rider/orders')">
        <i class="css-icon list"></i>
        <span>订单</span>
      </div>
      <div class="nav-item" @click="$router.push('/rider/stats')">
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
const loadingMore = ref(false)
const activeTab = ref('all')
const dateFilter = ref('')
const statusFilter = ref('')
const hasMore = ref(true)
const currentPage = ref(1)

// 订单数据
const orders = ref([])

// 订单统计
const orderStats = ref({
  completed: 0,
  cancelled: 0,
  totalIncome: 0,
  efficiency: 0
})

// 计算属性
const filteredOrders = computed(() => {
  let filtered = orders.value

  // 按Tab筛选
  if (activeTab.value !== 'all') {
    filtered = filtered.filter(order => {
      if (activeTab.value === 'completed') {
        return order.status === 'completed'
      } else if (activeTab.value === 'cancelled') {
        return order.status === 'cancelled'
      }
      return true
    })
  }

  // 按日期筛选
  if (dateFilter.value) {
    const now = new Date()
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

    filtered = filtered.filter(order => {
      const orderDate = new Date(order.createdAt)

      switch (dateFilter.value) {
        case 'today':
          return orderDate >= today
        case 'yesterday':
          const yesterday = new Date(today)
          yesterday.setDate(yesterday.getDate() - 1)
          return orderDate >= yesterday && orderDate < today
        case 'week':
          const weekAgo = new Date(today)
          weekAgo.setDate(weekAgo.getDate() - 7)
          return orderDate >= weekAgo
        case 'month':
          const monthAgo = new Date(today)
          monthAgo.setMonth(monthAgo.getMonth() - 1)
          return orderDate >= monthAgo
        default:
          return true
      }
    })
  }

  // 按状态筛选 - 修复状态值匹配
  if (statusFilter.value) {
    filtered = filtered.filter(order => {
      // 将数字状态转换为文本状态进行匹配
      const orderStatus = order.status === 'completed' ? 'completed' :
                      order.status === 'cancelled' ? 'cancelled' :
                      order.status
      return orderStatus === statusFilter.value
    })
  }

  return filtered
})

// 格式化时间
const formatTime = (timeString) => {
  if (!timeString) return ''
  const date = new Date(timeString)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 获取状态类型
const getStatusType = (status) => {
  const typeMap = {
    'completed': 'success',
    'cancelled': 'danger'
  }
  return typeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const textMap = {
    'completed': '已完成',
    'cancelled': '已取消'
  }
  return textMap[status] || '未知状态'
}

// 初始化数据
const initData = async () => {
  try {
    loading.value = true
    currentPage.value = 1

    // 获取订单历史
    await loadOrders()

    // 获取统计数据
    try {
      const statsResult = await riderApi.getWorkStats()
      if (statsResult.code === 1 && statsResult.data) {
        orderStats.value.completed = statsResult.data.completedOrders || 0
        orderStats.value.cancelled = statsResult.data.cancelledOrders || 0
        orderStats.value.totalIncome = statsResult.data.totalIncome || 0
        orderStats.value.efficiency = statsResult.data.efficiency || 0
      }
    } catch (statsError) {
      console.warn('统计数据加载失败，使用默认值:', statsError)
      // 设置默认值，不影响用户使用
      orderStats.value.completed = 0
      orderStats.value.cancelled = 0
      orderStats.value.totalIncome = 0
      orderStats.value.efficiency = 0
    }

  } catch (error) {
    console.error('初始化订单数据失败:', error)
    console.error('错误详情:', {
      message: error.message,
      response: error.response,
      status: error.response?.status,
      data: error.response?.data
    })

    // 根据错误类型显示不同信息
    if (error.response?.status === 401) {
      ElMessage.error('身份验证失败，请重新登录')
    } else if (error.response?.status === 403) {
      ElMessage.error('无权限访问历史订单')
    } else if (error.response?.status === 404) {
      ElMessage.error('历史订单接口不存在')
    } else if (error.response?.status === 500) {
      ElMessage.error('服务器内部错误，请稍后重试')
    } else if (error.code === 'NETWORK_ERROR' || !error.response) {
      ElMessage.error('网络连接失败，请检查网络后重试')
    } else {
      ElMessage.error('获取订单数据失败：' + (error.response?.data?.msg || error.message || '未知错误'))
    }
  } finally {
    loading.value = false
  }
}

// 加载订单
const loadOrders = async (page = 1) => {
  try {
    // 使用API获取订单历史
    const result = await riderApi.getOrderHistory({
      page,
      pageSize: 20,
      dateFilter: dateFilter.value,
      statusFilter: statusFilter.value
    })

    if (result.code === 1) {
      const response = result.data

      if (page === 1) {
        orders.value = response.items || []
      } else {
        orders.value = [...orders.value, ...(response.items || [])]
      }

      hasMore.value = response.items && response.items.length === 20

    } else {
      console.error('获取订单数据失败：', result.msg || '未知错误')
      ElMessage.error('获取订单数据失败')
    }
  } catch (error) {
    console.error('加载订单失败:', error)

    // 如果API失败，尝试使用演示数据
    if (error.response?.status >= 500 || !error.response) {
      console.log('使用演示数据')
      const demoResult = await riderApi.getOrderHistoryWithDemo()
      if (demoResult.code === 1 && demoResult.data.items) {
        const demoItems = demoResult.data.items.map(item => ({
          id: item.id,
          orderNo: `#${item.orderId}`,
          orderId: item.orderId,
          restaurant: item.restaurant,
          customer: item.customer,
          deliveryAddress: item.customer ? `${item.customer}的配送地址` : '配送地址',
          pickupAddress: '取餐地址',
          distance: (Math.random() * 3 + 0.5).toFixed(1),
          deliveryTime: Math.floor(Math.random() * 30 + 10),
          fee: item.amount * 0.8,
          amount: item.amount,
          status: item.status === 4 ? 'completed' : 'cancelled',
          createdAt: item.time,
          remark: item.remark
        }))

        if (page === 1) {
          orders.value = demoItems
        } else {
          orders.value = [...orders.value, ...demoItems]
        }

        hasMore.value = false

        if (page === 1) {
          ElMessage({
            message: '网络连接异常，当前显示演示数据',
            type: 'warning',
            duration: 4000
          })
        }
      }
    } else {
      ElMessage.error('加载订单失败：' + (error.response?.data?.msg || error.message))
    }
  } finally {
    loading.value = false
  }
}

// 刷新订单
const refreshOrders = async () => {
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

// 加载更多
const loadMore = async () => {
  try {
    loadingMore.value = true
    currentPage.value++
    await loadOrders(currentPage.value)
  } catch (error) {
    currentPage.value--
    ElMessage.error('加载更多失败')
  } finally {
    loadingMore.value = false
  }
}

// Tab切换
const handleTabChange = () => {
  // Tab切换时不需要重新加载数据，只需要过滤
}

// 日期筛选变化
const handleDateChange = () => {
  initData()
}

// 状态筛选变化
const handleStatusChange = () => {
  // 状态筛选变化时不需要重新加载数据，只需要过滤
}

// 跳转到订单详情
const goToOrderDetail = (orderId) => {
  router.push(`/rider/orders/${orderId}`)
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

/* 关闭图标 */
.css-icon.close::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(45deg);
  width: 12px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
}

.css-icon.close::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-45deg);
  width: 12px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
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

.rider-orders {
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

/* 订单统计 */
.order-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  padding: 15px;
  background: white;
  margin: 10px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 10px;
}

.stat-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: white;
  flex-shrink: 0;
}

.stat-item.completed .stat-icon {
  background: linear-gradient(135deg, #67C23A 0%, #52c41a 100%);
}

.stat-item.cancelled .stat-icon {
  background: linear-gradient(135deg, #F56C6C 0%, #ff4757 100%);
}

.stat-item.income .stat-icon {
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
}

.stat-item.efficiency .stat-icon {
  background: linear-gradient(135deg, #409EFF 0%, #1890ff 100%);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 18px;
  font-weight: bold;
  color: #333;
  margin-bottom: 2px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

/* 筛选区域 */
.filter-section {
  background: white;
  margin: 10px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.order-tabs :deep(.el-tabs__nav) {
  background: #f8f9fa;
}

.tab-content {
  padding: 15px;
}

.filter-bar {
  display: flex;
  gap: 10px;
}

/* 订单列表 */
.order-list {
  padding: 10px;
}

.order-item {
  background: white;
  border-radius: 12px;
  padding: 15px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s ease;
}

.order-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.order-id {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.order-time {
  font-size: 12px;
  color: #666;
}

/* 订单内容 */
.order-content {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.route-info {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

.route-point {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  flex: 1;
}

.point-icon {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
  font-size: 14px;
}

.route-point.pickup .point-icon {
  background: #409EFF;
}

.route-point.delivery .point-icon {
  background: #67C23A;
}

.point-detail {
  flex: 1;
}

.point-name {
  font-weight: 600;
  color: #333;
  margin-bottom: 2px;
  font-size: 14px;
}

.point-address {
  color: #666;
  font-size: 13px;
  line-height: 1.4;
}

.route-arrow {
  color: #FFD700;
  font-size: 16px;
  margin: 5px 0;
  text-align: center;
}

/* 订单详情 */
.order-details {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
}

.detail-item {
  text-align: center;
}

.detail-label {
  font-size: 12px;
  color: #666;
  display: block;
  margin-bottom: 2px;
}

.detail-value {
  font-size: 13px;
  color: #333;
  font-weight: 500;
}

.detail-value.amount {
  color: #67C23A;
  font-weight: bold;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #999;
}

/* 加载更多 */
.load-more {
  text-align: center;
  padding: 20px;
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
  .order-stats {
    margin: 5px;
    padding: 10px;
    gap: 8px;
  }

  .filter-section {
    margin: 5px;
  }

  .tab-content {
    padding: 10px;
  }

  .order-list {
    padding: 5px;
  }

  .order-item {
    padding: 12px;
  }

  .order-details {
    grid-template-columns: 1fr;
    gap: 5px;
  }

  .route-info {
    flex-direction: column;
    gap: 5px;
  }

  .route-point {
    flex-direction: row;
  }
}
</style>