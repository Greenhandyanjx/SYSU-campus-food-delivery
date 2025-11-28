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

    <!-- 骑手状态卡片 -->
    <div class="status-card">
      <div class="rider-status">
        <el-avatar :size="50" :src="riderInfo?.avatar || ''" />
        <div class="rider-info">
          <h3>{{ riderInfo?.name || '骑手' }}</h3>
          <p>{{ riderInfo?.completedOrders || 0 }}单 · {{ riderInfo?.rating || 5.0 }}分</p>
        </div>
        <el-switch
          v-model="isOnline"
          active-color="#FFD700"
          inactive-color="#DCDFE6"
          :active-text="isOnline ? '在线接单' : '已下线'"
          @change="toggleOnlineStatus"
        />
      </div>
      <div class="income-info">
        <div class="income-item income-highlight">
          <span class="income-label">今日收入</span>
          <span class="income-amount">¥{{ safeIncomeDisplay }}</span>
        </div>
        <div class="income-item">
          <span class="income-label">在线时长</span>
          <span class="income-value">{{ onlineHours }}h</span>
        </div>
        <div class="income-item">
          <span class="income-label">完成订单</span>
          <span class="income-value">{{ riderInfo.completedOrders }}单</span>
        </div>
        <div class="income-item">
          <span class="income-label">平均单价</span>
          <span class="income-value">¥{{ calculateAveragePrice() }}</span>
        </div>
      </div>
    </div>

    <!-- 订单状态Tab -->
    <div class="order-tabs-container">
      <el-tabs v-model="activeTab" type="border-card" class="order-tabs">
        <!-- 新订单 -->
        <el-tab-pane label="新订单" name="new">
          <div class="tab-header">
            <span class="tab-title">待接订单 ({{ newOrders.length }})</span>
            <el-button size="small" @click="refreshNewOrders" :loading="refreshing">
              <i class="css-icon refresh"></i>
              刷新
            </el-button>
          </div>

          <div class="order-list" v-if="newOrders.length > 0">
            <div v-for="order in newOrders" :key="order.id" class="order-card new-order">
              <div class="order-header">
                <el-tag type="warning">新订单</el-tag>
                <span class="order-time">{{ formatTime(order.createdAt) }}</span>
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
                    <span>距离: {{ order.distance }}km</span>
                    <span>预计收入: ¥{{ order.estimatedFee }}</span>
                    <span>预计时间: {{ order.estimatedTime }}分钟</span>
                  </div>
                  <el-button type="primary" size="large" @click="acceptOrder(order)" :loading="order.accepting">
                    立即抢单
                  </el-button>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="empty-state">
            <el-empty description="暂无新订单" />
          </div>
        </el-tab-pane>

        <!-- 待取货 -->
        <el-tab-pane label="待取货" name="pickup">
          <div class="tab-header">
            <span class="tab-title">待取货订单 ({{ pickupOrders.length }})</span>
            <el-button size="small" @click="refreshPickupOrders" :loading="refreshing">
              <i class="css-icon refresh"></i>
              刷新
            </el-button>
          </div>

          <div class="order-list" v-if="pickupOrders.length > 0">
            <div v-for="order in pickupOrders" :key="order.id" class="order-card pickup-order">
              <div class="order-header">
                <el-tag type="info">待取货</el-tag>
                <span class="order-timer">{{ formatCountdown(order.remainingTime) }}</span>
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

                <div class="pickup-info">
                  <div class="pickup-code">
                    <span>取餐码: </span>
                    <span class="code">{{ order.pickupCode }}</span>
                  </div>
                  <div class="shop-phone">
                    <span>商家电话: {{ order.shopPhone }}</span>
                  </div>
                </div>

                <div class="order-footer">
                  <div class="order-actions">
                    <el-button @click="callShop(order)">
                      <i class="css-icon phone"></i>
                      联系商家
                    </el-button>
                    <el-button type="success" @click="confirmPickup(order)" :loading="order.pickingUp">
                      确认取货
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="empty-state">
            <el-empty description="暂无待取货订单" />
          </div>
        </el-tab-pane>

        <!-- 配送中 -->
        <el-tab-pane label="配送中" name="delivering">
          <div class="tab-header">
            <span class="tab-title">配送中订单 ({{ deliveringOrders.length }})</span>
            <el-button size="small" @click="refreshDeliveringOrders" :loading="refreshing">
              <i class="css-icon refresh"></i>
              刷新
            </el-button>
          </div>

          <div class="order-list" v-if="deliveringOrders.length > 0">
            <div v-for="order in deliveringOrders" :key="order.id" class="order-card delivering-order">
              <div class="order-header">
                <el-tag type="danger">配送中</el-tag>
                <span class="order-timer">{{ formatCountdown(order.remainingTime) }}</span>
              </div>

              <div class="order-content">
                <div class="customer-info">
                  <el-avatar :size="40" :src="order.customerAvatar" />
                  <div class="customer-detail">
                    <div class="customer-name">{{ order.customer }}</div>
                    <div class="customer-phone">{{ order.customerPhone }}</div>
                  </div>
                </div>

                <div class="delivery-info">
                  <div class="delivery-address">
                    <i class="css-icon location"></i>
                    <span>{{ order.deliveryAddress }}</span>
                  </div>
                </div>

                <div class="order-footer">
                  <div class="order-actions">
                    <el-button @click="callCustomer(order)">
                      <i class="css-icon phone"></i>
                      联系顾客
                    </el-button>
                    <el-button type="primary" @click="startNavigation(order)">
                      <i class="css-icon map"></i>
                      开始导航
                    </el-button>
                    <el-button type="success" @click="completeDelivery(order)" :loading="order.completing">
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 基础状态
const currentTime = ref('')
const isOnline = ref(false)
const activeTab = ref('new')
const loading = ref(false)
const refreshing = ref(false)
// 移除全局loading状态，使用订单级别的loading状态
// const accepting = ref(false)
// const pickingUp = ref(false)
// const completing = ref(false)

// 骑手信息
const riderInfo = ref({
  name: '骑手',
  avatar: '',
  rating: 5.0,
  completedOrders: 0
})

// 工作数据
const todayIncome = ref(0)
const onlineHours = ref(0)

// 订单数据
const newOrders = ref([])
const pickupOrders = ref([])
const deliveringOrders = ref([])

// 更新时间
let timer = null
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('zh-CN', { hour12: false })
}

// 格式化时间
const formatTime = (timeString) => {
  if (!timeString) return ''
  const date = new Date(timeString)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 格式化倒计时
const formatCountdown = (milliseconds) => {
  if (!milliseconds) return '00:00'
  const minutes = Math.floor(milliseconds / (60 * 1000))
  const seconds = Math.floor((milliseconds % (60 * 1000)) / 1000)
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}

// 初始化数据
const initData = async () => {
  try {
    loading.value = true

    // 获取骑手信息
    try {
      const riderData = await riderApi.getRiderInfo()
      console.log('骑手信息API响应:', riderData)

      if (riderData && riderData.code === 1 && riderData.data) {
        riderInfo.value = riderData.data
        isOnline.value = riderData.data.isOnline
        console.log('骑手信息加载成功:', riderData.data)
      } else {
        console.warn('骑手信息API返回异常:', riderData)
        // 尝试使用demo数据
        const demoData = await riderApi.getRiderInfoWithDemo()
        if (demoData.code === 1 && demoData.data) {
          riderInfo.value = demoData.data
          isOnline.value = demoData.data.isOnline
          console.log('使用demo骑手信息:', demoData.data)
        }
      }
    } catch (error) {
      console.error('获取骑手信息失败:', error)

      // 检查是否是认证问题
      if (error.response?.status === 401) {
        ElMessage.error('登录已过期，请重新登录')
        localStorage.removeItem('token')
        setTimeout(() => {
          // 可以考虑跳转到登录页
          // window.location.href = '/login'
        }, 2000)
        return
      }

      console.warn('使用demo数据作为后备')
      const demoData = await riderApi.getRiderInfoWithDemo()
      if (demoData.code === 1 && demoData.data) {
        riderInfo.value = demoData.data
        isOnline.value = demoData.data.isOnline
      }
    }

    // 获取工作台数据
    try {
      const dashboardData = await riderApi.getDashboard()
      console.log('工作台API响应:', dashboardData)
      if (dashboardData && dashboardData.code === 1 && dashboardData.data) {
        // 使用安全数值处理
        todayIncome.value = safeNumber(dashboardData.data.todayIncome, 0)
        onlineHours.value = safeNumber(dashboardData.data.onlineHours, 0)
      }
    } catch (error) {
      console.warn('获取工作台数据失败，使用demo数据:', error)
      const demoData = await riderApi.getIncomeStatsWithDemo()
      if (demoData.code === 1 && demoData.data) {
        todayIncome.value = safeNumber(demoData.data.dailyIncome, 0)
        // 这里可以设置一个默认在线时长
        onlineHours.value = 2.5
      }
    }

    // 获取各状态订单
    await Promise.all([
      loadNewOrders(),
      loadPickupOrders(),
      loadDeliveringOrders()
    ])

  } catch (error) {
    console.error('初始化数据失败:', error)
    ElMessage.error('获取数据失败，请刷新重试')
  } finally {
    loading.value = false
  }
}

// 加载新订单
const loadNewOrders = async () => {
  try {
    // 使用真实API获取新订单数据
    const response = await riderApi.getNewOrders()
    console.log('新订单API响应:', response) // 调试日志
    if (response && response.code === 1 && Array.isArray(response.data)) {
      // 为每个订单初始化loading状态
      newOrders.value = response.data.map(order => ({
        ...order,
        accepting: false  // 为每个订单添加独立的loading状态
      }))
    } else {
      newOrders.value = []
      console.warn('获取的新订单数据格式不正确:', response)
    }
  } catch (error) {
    console.error('获取新订单失败，尝试使用demo数据:', error)
    try {
      const demoData = await riderApi.getNewOrdersWithDemo()
      if (demoData.code === 1 && Array.isArray(demoData.data)) {
        newOrders.value = demoData.data.map(order => ({
          ...order,
          accepting: false  // 为每个demo订单也添加loading状态
        }))
        console.log('使用demo订单数据:', newOrders.value.length)
      } else {
        newOrders.value = []
      }
    } catch (demoError) {
      console.error('demo数据也获取失败:', demoError)
      newOrders.value = []
      ElMessage.warning('无法获取新订单，请检查网络连接或刷新重试')
    }
  }
}

// 加载待取货订单
const loadPickupOrders = async () => {
  try {
    const response = await riderApi.getPickupOrders()
    console.log('待取货订单API响应:', response) // 调试日志
    if (response && response.code === 1) {
      // 为每个订单初始化loading状态
      pickupOrders.value = (response.data || []).map(order => ({
        ...order,
        pickingUp: false  // 为每个订单添加独立的loading状态
      }))
    } else {
      console.warn('获取待取货订单返回异常:', response?.msg)
      pickupOrders.value = []
    }
  } catch (error) {
    console.error('获取待取货订单失败，使用空数组:', error)
    pickupOrders.value = []
    // 待取货和配送中订单不显示demo数据，只显示空状态
  }
}

// 加载配送中订单
const loadDeliveringOrders = async () => {
  try {
    const response = await riderApi.getDeliveringOrders()
    console.log('配送中订单API响应:', response) // 调试日志
    if (response && response.code === 1) {
      // 为每个订单初始化loading状态
      deliveringOrders.value = (response.data || []).map(order => ({
        ...order,
        completing: false  // 为每个订单添加独立的loading状态
      }))
    } else {
      console.warn('获取配送中订单返回异常:', response?.msg)
      deliveringOrders.value = []
    }
  } catch (error) {
    console.error('获取配送中订单失败，使用空数组:', error)
    deliveringOrders.value = []
    // 待取货和配送中订单不显示demo数据，只显示空状态
  }
}

// 刷新新订单
const refreshNewOrders = async () => {
  refreshing.value = true
  await loadNewOrders()
  refreshing.value = false
}

// 刷新待取货订单
const refreshPickupOrders = async () => {
  refreshing.value = true
  await loadPickupOrders()
  refreshing.value = false
}

// 刷新配送中订单
const refreshDeliveringOrders = async () => {
  refreshing.value = true
  await loadDeliveringOrders()
  refreshing.value = false
}

// 切换在线状态
const toggleOnlineStatus = async (status) => {
  try {
    await riderApi.updateRiderStatus({ isOnline: status })
    ElMessage.success(status ? '已上线，开始接单' : '已下线，停止接单')

    // 如果上线，刷新订单
    if (status) {
      await loadNewOrders()
    }
  } catch (error) {
    ElMessage.error('状态更新失败，请重试')
    isOnline.value = !status
  }
}

// 接单
const acceptOrder = async (order) => {
  try {
    // 为当前订单设置loading状态
    order.accepting = true
    const response = await riderApi.acceptOrder(order.id)

    if (response.code === 1) {
      ElMessage.success(`接单成功！订单号：${order.id}`)

      // 从新订单中移除
      newOrders.value = newOrders.value.filter(o => o.id !== order.id)

      // 添加到待取货，并设置取餐码和loading状态
      pickupOrders.value.push({
        ...order,
        pickupCode: response.data?.pickupCode || 'A' + order.id,
        shopPhone: '138' + order.id.toString().padStart(8, '0'),
        remainingTime: 15 * 60 * 1000,
        pickingUp: false  // 初始化确认取货按钮的loading状态
      })

      // 切换到待取货tab
      activeTab.value = 'pickup'
    }
  } catch (error) {
    ElMessage.error('接单失败，请重试')
  } finally {
    // 清除当前订单的loading状态
    order.accepting = false
  }
}

// 确认取货
const confirmPickup = async (order) => {
  try {
    // 为当前订单设置loading状态
    order.pickingUp = true

    await riderApi.pickupOrder(order.id)

    ElMessage.success(`取货确认！订单号：${order.id}`)

    // 从待取货中移除
    pickupOrders.value = pickupOrders.value.filter(o => o.id !== order.id)

    // 添加到配送中
    deliveringOrders.value.push({
      ...order,
      customer: order.customer || '顾客',
      customerPhone: order.customerPhone || '13666666666',
      customerAvatar: order.customerAvatar || '',
      remainingTime: 30 * 60 * 1000,
      completing: false  // 初始化完成配送按钮的loading状态
    })

    // 切换到配送中tab
    activeTab.value = 'delivering'
  } catch (error) {
    ElMessage.error('取货确认失败，请重试')
  } finally {
    // 清除当前订单的loading状态
    order.pickingUp = false
  }
}

// 安全的数值计算函数
const safeNumber = (value, defaultValue = 0) => {
  const num = parseFloat(value)
  return isNaN(num) ? defaultValue : num
}

// 安全的收入显示
const safeIncomeDisplay = computed(() => {
  const income = safeNumber(todayIncome.value, 0)
  return isNaN(income) ? '0.00' : income.toFixed(2)
})

// 计算平均单价（带安全保护）
const calculateAveragePrice = () => {
  const income = safeNumber(todayIncome.value, 0)
  const orders = safeNumber(riderInfo.value.completedOrders, 0)

  if (orders > 0 && income >= 0) {
    return (income / orders).toFixed(2)
  } else {
    return '0.00'
  }
}

// 收入更新动画
const animateIncomeUpdate = (amount) => {
  const incomeElement = document.querySelector('.income-highlight')
  if (incomeElement) {
    incomeElement.classList.add('income-updated')
    setTimeout(() => {
      incomeElement.classList.remove('income-updated')
    }, 600)
  }

  // 显示收入增长提示
  const safeAmount = safeNumber(amount, 0)
  ElMessage({
    message: `收入 +¥${safeAmount.toFixed(2)}`,
    type: 'success',
    duration: 2000,
    showClose: false,
    customClass: 'income-toast'
  })
}

// 完成配送
const completeDelivery = async (order) => {
  try {
    console.log('开始完成配送，订单ID:', order.id)
    console.log('订单详情:', order)

    // 预检查订单状态，确保可以进行配送完成
    if (!order.id) {
      ElMessage.error('订单信息不完整，无法完成配送')
      return
    }

    // 检查骑手是否已登录
    const token = localStorage.getItem('token')
    if (!token) {
      ElMessage.error('请先登录骑手账号')
      return
    }

    // 为当前订单设置loading状态
    order.completing = true

    console.log('调用配送完成API...')
    console.log('使用的token:', token ? '已设置' : '未设置')

    const response = await riderApi.completeOrder(order.id)
    console.log('配送完成API响应:', response)

    if (response.code === 1) {
      // 从配送中移除
      deliveringOrders.value = deliveringOrders.value.filter(o => o.id !== order.id)

      // 更新收入数据 - 使用后端返回的实际费用
      const actualFee = safeNumber(response.data?.actualFee, 0) || safeNumber(order.estimatedFee, 0)

      // 立即更新本地显示（使用安全计算）
      todayIncome.value = safeNumber(todayIncome.value, 0) + actualFee
      riderInfo.value.completedOrders = safeNumber(riderInfo.value.completedOrders, 0) + 1

      // 显示成功消息，包含收入信息
      ElMessage({
        message: `配送完成！订单号：${order.id}，收入 ¥${actualFee.toFixed(2)}`,
        type: 'success',
        duration: 3000
      })

      // 触发收入更新动画
      animateIncomeUpdate(actualFee)

      // 延迟刷新数据以验证后端一致性
      setTimeout(async () => {
        const freshData = await riderApi.getDashboard()
        if (freshData.code === 1) {
          // 验证后端数据与本地更新是否一致
          const backendIncome = safeNumber(freshData.data.todayIncome, 0)
          const localIncome = safeNumber(todayIncome.value, 0)

          if (Math.abs(backendIncome - localIncome) > 0.01) {
            console.warn('收入数据不一致，后端:', backendIncome, '本地:', localIncome)
            todayIncome.value = backendIncome
          }
        }
      }, 1500)
    } else {
      console.error('配送完成失败，后端返回错误:', response)
      ElMessage.error(response.msg || '配送完成失败，请重试')
    }
  } catch (error) {
    console.error('配送完成异常:', error)
    console.error('错误详情:', {
      message: error.message,
      response: error.response,
      status: error.response?.status,
      data: error.response?.data
    })

    // 检查是否是认证问题
    if (error.response?.status === 401) {
      ElMessage.error('身份验证失败，请重新登录')
      // 清除无效token
      localStorage.removeItem('token')
      // 可以考虑跳转到登录页
      setTimeout(() => {
        // window.location.href = '/login'
      }, 2000)
    } else if (error.response?.status === 403) {
      ElMessage.error('无权限操作此订单，请确认您是该订单的骑手')
    } else if (error.response?.status === 404) {
      ElMessage.error('订单不存在或已完成')
    } else if (error.response?.status === 400) {
      const errorMsg = error.response?.data?.msg || '订单状态不正确'
      if (errorMsg.includes('状态不正确')) {
        ElMessage.error('订单状态不正确，请确认订单是否在配送中')
      } else {
        ElMessage.error(errorMsg)
      }
    } else if (error.code === 'NETWORK_ERROR' || !error.response) {
      ElMessage.error('网络连接失败，请检查网络后重试')
    } else {
      ElMessage.error('配送完成失败：' + (error.response?.data?.msg || error.message || '未知错误'))
    }
  } finally {
    // 清除当前订单的loading状态
    order.completing = false
  }
}

// 联系商家
const callShop = (order) => {
  ElMessage.info(`正在联系商家：${order.shopPhone}`)
  // 这里可以集成实际的电话拨打功能
  if (order.shopPhone) {
    window.location.href = `tel:${order.shopPhone}`
  }
}

// 联系顾客
const callCustomer = (order) => {
  ElMessage.info(`正在联系顾客：${order.customerPhone}`)
  // 这里可以集成实际的电话拨打功能
  if (order.customerPhone) {
    window.location.href = `tel:${order.customerPhone}`
  }
}

// 开始导航
const startNavigation = (order) => {
  ElMessage.info('正在启动导航...')
  // 这里可以集成地图导航功能
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
  initData()

  // 定期刷新订单数据
  const orderTimer = setInterval(() => {
    // 始终刷新订单数据，这样可以看到demo数据
    loadNewOrders()
    loadPickupOrders()
    loadDeliveringOrders()
  }, 15000) // 减少到15秒，提高刷新频率

  // 保存定时器以便清理
  timer = orderTimer
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
    if (timer.orderTimer) clearInterval(timer.orderTimer)
  }
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

.rider-dashboard {
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

/* 状态卡片 */
.status-card {
  background: white;
  margin: 10px;
  border-radius: 12px;
  padding: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.rider-status {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 15px;
}

.rider-info h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
}

.rider-info p {
  margin: 5px 0 0 0;
  color: #666;
  font-size: 14px;
}

.income-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 15px;
  border-top: 1px solid #f0f0f0;
}

.income-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.income-item.income-highlight {
  background: linear-gradient(135deg, #FFD700, #FFA500);
  color: white;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(255, 215, 0, 0.3);
}

.income-label {
  font-size: 14px;
  opacity: 0.8;
}

.income-item.income-highlight .income-label {
  opacity: 0.9;
}

.income-amount {
  font-size: 18px;
  font-weight: 700;
}

.income-value {
  font-size: 14px;
  font-weight: 500;
}

/* 收入变化动画 */
@keyframes incomeUpdate {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

.income-highlight.income-updated {
  animation: incomeUpdate 0.6s ease-in-out;
}

/* 订单Tab容器 */
.order-tabs-container {
  margin: 10px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.order-tabs :deep(.el-tabs__header) {
  border-radius: 12px 12px 0 0;
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

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid #f0f0f0;
}

.tab-title {
  font-weight: 600;
  color: #333;
}

/* 订单列表 */
.order-list {
  max-height: 500px;
  overflow-y: auto;
  padding: 10px;
}

.order-card {
  background: white;
  border-radius: 12px;
  padding: 15px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border: 1px solid #f0f0f0;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.order-time {
  color: #666;
  font-size: 14px;
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
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
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
}

.point-address {
  color: #666;
  font-size: 14px;
  line-height: 1.4;
}

.route-arrow {
  text-align: center;
  color: #FFD700;
  font-size: 20px;
  margin: 5px 0;
}

/* 取货信息 */
.pickup-info {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 10px;
  margin-bottom: 15px;
}

.pickup-code {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-bottom: 5px;
}

.code {
  font-family: 'Courier New', monospace;
  font-weight: bold;
  color: #409EFF;
  font-size: 16px;
}

.shop-phone {
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
  font-weight: 600;
  color: #333;
  margin-bottom: 2px;
}

.customer-phone {
  color: #666;
  font-size: 14px;
}

/* 配送信息 */
.delivery-info {
  margin-bottom: 15px;
}

.delivery-address {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #666;
  font-size: 14px;
}

/* 订单底部 */
.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.order-info span {
  color: #666;
  font-size: 13px;
}

.order-actions {
  display: flex;
  gap: 8px;
}

.order-actions :deep(.el-button) {
  flex: 1;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #999;
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
  .status-card {
    margin: 5px;
    padding: 10px;
  }

  .order-tabs-container {
    margin: 5px;
  }

  .order-list {
    padding: 5px;
  }

  .order-card {
    padding: 10px;
  }

  .order-actions {
    flex-direction: column;
    gap: 5px;
  }

  .order-actions :deep(.el-button) {
    flex: none;
  }
}
</style>