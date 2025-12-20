<template>
  <div class="rider-profile">
    <!-- 顶部导航 -->
    <div class="header-nav">
      <div class="nav-left">
        <el-button type="text" @click="$router.back()">
          <i class="css-icon arrow-left"></i>
        </el-button>
      </div>
      <div class="nav-title">个人中心</div>
      <div class="nav-right">
        <el-button type="text" @click="showSettings">
          <i class="css-icon setting"></i>
        </el-button>
      </div>
    </div>

    <!-- 骑手信息卡片 -->
    <div class="profile-card">
      <div class="profile-header">
        <el-avatar :size="80" :src="riderInfo.avatar || '/src/assets/user.png'" />
        <div class="profile-info">
          <h2>{{ riderInfo.name }}</h2>
          <div class="rating-section">
            <el-rate v-model="riderInfo.rating" disabled />
            <span class="rating-text">{{ riderInfo.rating }}分</span>
          </div>
          <p class="completed-orders">已完成 {{ riderInfo.completedOrders }} 单</p>
        </div>
      </div>

      <div class="status-section">
        <div class="status-item">
          <span class="status-label">当前状态</span>
          <el-tag :type="isOnline ? 'success' : 'info'">
            {{ isOnline ? '在线接单' : '已下线' }}
          </el-tag>
        </div>
        <div class="status-item">
          <span class="status-label">工作时长</span>
          <span class="status-value">{{ workHours }}小时</span>
        </div>
      </div>
    </div>

    <!-- 数据统计 -->
    <div class="stats-section">
      <h3>数据统计</h3>
      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-icon income">
            <i class="css-icon wallet"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">¥{{ totalIncome.toFixed(2) }}</div>
            <div class="stat-label">总收入</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon orders">
            <i class="css-icon document"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ totalOrders }}</div>
            <div class="stat-label">完成订单</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon rating">
            <i class="css-icon star"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ riderInfo.rating }}</div>
            <div class="stat-label">评分</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-icon efficiency">
            <i class="css-icon data-analysis"></i>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ efficiency }}单/时</div>
            <div class="stat-label">效率</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 功能菜单 -->
    <div class="menu-section">
      <div class="menu-group">
        <div class="menu-item" @click="goToProfileEdit">
          <div class="menu-icon">
            <i class="css-icon user-edit"></i>
          </div>
          <span class="menu-title">个人资料</span>
          <i class="css-icon arrow-right"></i>
        </div>
        <div class="menu-item" @click="goToSecurity">
          <div class="menu-icon">
            <i class="css-icon lock"></i>
          </div>
          <span class="menu-title">账户安全</span>
          <i class="css-icon arrow-right"></i>
        </div>
        <div class="menu-item" @click="goToPayment">
          <div class="menu-icon">
            <i class="css-icon credit-card"></i>
          </div>
          <span class="menu-title">收款设置</span>
          <i class="css-icon arrow-right"></i>
        </div>
        <div class="menu-item" @click="goToWorkSettings">
          <div class="menu-icon">
            <i class="css-icon setting"></i>
          </div>
          <span class="menu-title">工作设置</span>
          <i class="css-icon arrow-right"></i>
        </div>
      </div>

      <div class="menu-group">
        <div class="menu-item" @click="goToNotification">
          <div class="menu-icon">
            <i class="css-icon notification"></i>
          </div>
          <span class="menu-title">消息通知</span>
          <span class="menu-badge" v-if="unreadCount > 0">{{ unreadCount }}</span>
          <i class="css-icon arrow-right"></i>
        </div>
        <div class="menu-item" @click="goToVerification">
          <div class="menu-icon">
            <i class="css-icon shield"></i>
          </div>
          <span class="menu-title">实名认证</span>
          <el-tag :type="verificationStatus === 'verified' ? 'success' : verificationStatus === 'pending' ? 'warning' : 'info'" size="small">
            {{ verificationStatusText }}
          </el-tag>
          <i class="css-icon arrow-right"></i>
        </div>
      </div>

      <div class="menu-group">
        <div class="menu-item" @click="goToHelp">
          <div class="menu-icon">
            <i class="css-icon service"></i>
          </div>
          <span class="menu-title">帮助中心</span>
          <i class="css-icon arrow-right"></i>
        </div>
        <div class="menu-item" @click="goToFeedback">
          <div class="menu-icon">
            <i class="css-icon message"></i>
          </div>
          <span class="menu-title">意见反馈</span>
          <i class="css-icon arrow-right"></i>
        </div>
      </div>

      <div class="menu-group logout">
        <div class="menu-item logout-item" @click="handleLogout">
          <div class="menu-icon">
            <i class="css-icon logout"></i>
          </div>
          <span class="menu-title">退出登录</span>
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
      <div class="nav-item active" @click="$router.push('/rider/profile')">
        <i class="css-icon user"></i>
        <span>我的</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

const defaultAvatar = '/src/assets/user.png'

const router = useRouter()

// 状态数据
const loading = ref(false)
const isOnline = ref(false)
const unreadCount = ref(0)

// 骑手信息
const riderInfo = ref({
  name: '骑手',
  avatar: '',
  rating: 5.0,
  completedOrders: 0
})

// 工作数据
const workHours = ref(0)
const totalIncome = ref(0)
const totalOrders = ref(0)

// 认证状态
const verificationStatus = ref('unverified') // unverified, pending, verified

// 计算属性
const efficiency = computed(() => {
  if (workHours.value > 0) {
    return (totalOrders.value / workHours.value).toFixed(1)
  }
  return '0.0'
})

const verificationStatusText = computed(() => {
  const statusMap = {
    'unverified': '未认证',
    'pending': '审核中',
    'verified': '已认证'
  }
  return statusMap[verificationStatus.value] || '未认证'
})

// 初始化数据
const initData = async () => {
  try {
    loading.value = true

    // 获取骑手信息
    const riderData = await riderApi.getRiderInfo()
    if (riderData.code === 1 && riderData.data) {
      riderInfo.value = riderData.data
      isOnline.value = riderData.data.isOnline || false
    }

    // 获取统计数据
    const incomeData = await riderApi.getIncomeSummary()
    if (incomeData.code === 1 && incomeData.data) {
      totalIncome.value = incomeData.data.totalIncome || 0
      totalOrders.value = incomeData.data.completedOrders || 0
    }

    // 获取工作统计
    const workData = await riderApi.getWorkStats()
    if (workData.code === 1 && workData.data) {
      workHours.value = Math.round((workData.data.totalOrders * 0.3) / 60) // 估算在线时长
    }

    // 获取认证信息
    const verificationData = await riderApi.getVerification()
    if (verificationData.code === 1 && verificationData.data) {
      verificationStatus.value = verificationData.data.status || 'unverified'
    }

    // 获取未读通知数
    const notificationData = await riderApi.getNotifications({ read: false })
    if (notificationData.code === 1 && notificationData.data) {
      unreadCount.value = notificationData.data.unreadCount || 0
    }

  } catch (error) {
    console.error('初始化数据失败:', error)
    ElMessage.error('获取数据失败，请刷新重试')
  } finally {
    loading.value = false
  }
}

// 导航方法
const goToProfileEdit = () => {
  router.push('/rider/profile/edit')
}

const goToSecurity = () => {
  router.push('/rider/profile/security')
}

const goToPayment = () => {
  router.push('/rider/profile/payment')
}

const goToWorkSettings = () => {
  router.push('/rider/profile/work')
}

const goToNotification = () => {
  router.push('/rider/profile/notification')
}

const goToVerification = () => {
  router.push('/rider/profile/verification')
}

const goToHelp = () => {
  router.push('/rider/profile/help')
}

const goToFeedback = () => {
  router.push('/rider/profile/feedback')
}

const showSettings = () => {
  ElMessage.info('设置功能开发中...')
}

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要退出登录吗？',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 清除本地存储的用户信息
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')

    ElMessage.success('退出登录成功')

    // 跳转到登录页面
    router.push('/login')
  } catch (error) {
    // 用户取消操作
  }
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

/* 箭头右 */
.css-icon.arrow-right::before {
  content: '';
  position: absolute;
  top: 50%;
  right: 0;
  transform: translateY(-50%) rotate(45deg);
  width: 10px;
  height: 10px;
  border-right: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
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

/* 用户编辑图标 */
.css-icon.user-edit::before {
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

.css-icon.user-edit::after {
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

/* 锁图标 */
.css-icon.lock::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 12px;
  height: 8px;
  border: 2px solid currentColor;
  border-radius: 2px 2px 0 0;
}

.css-icon.lock::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
}

/* 信用卡图标 */
.css-icon.credit-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.credit-card::after {
  content: '';
  position: absolute;
  bottom: 2px;
  left: 2px;
  width: 12px;
  height: 1px;
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

/* 盾牌图标 */
.css-icon.shield::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 50% 50% 0 0;
}

.css-icon.shield::after {
  content: '';
  position: absolute;
  top: 6px;
  left: 50%;
  transform: translateX(-50%);
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

/* 消息图标 */
.css-icon.message::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 14px;
  height: 10px;
  border: 2px solid currentColor;
  border-radius: 7px;
}

.css-icon.message::after {
  content: '';
  position: absolute;
  bottom: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
}

/* 退出图标 */
.css-icon.logout::before {
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

.css-icon.logout::after {
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

.rider-profile {
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

/* 骑手信息卡片 */
.profile-card {
  background: white;
  margin: 15px;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.profile-info h2 {
  margin: 0;
  font-size: 22px;
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

.completed-orders {
  margin: 5px 0 0 0;
  color: #666;
  font-size: 14px;
}

.status-section {
  display: flex;
  justify-content: space-around;
  padding-top: 15px;
  border-top: 1px solid #f0f0f0;
}

.status-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
}

.status-label {
  color: #666;
  font-size: 14px;
}

.status-value {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

/* 数据统计 */
.stats-section {
  background: white;
  margin: 15px;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stats-section h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: #333;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 10px;
}

.stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: white;
}

.stat-icon.income {
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
}

.stat-icon.orders {
  background: linear-gradient(135deg, #409EFF 0%, #1890ff 100%);
}

.stat-icon.rating {
  background: linear-gradient(135deg, #67C23A 0%, #52c41a 100%);
}

.stat-icon.efficiency {
  background: linear-gradient(135deg, #E6A23C 0%, #ffa502 100%);
}

.stat-value {
  font-size: 20px;
  font-weight: bold;
  color: #333;
  margin-bottom: 2px;
}

.stat-label {
  font-size: 14px;
  color: #666;
}

/* 功能菜单 */
.menu-section {
  margin: 15px;
}

.menu-group {
  background: white;
  border-radius: 12px;
  margin-bottom: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-item:hover {
  background: #f8f9fa;
}

.menu-item.logout-item {
  justify-content: center;
  color: #f56c6c;
}

.menu-item.logout-item:hover {
  background: #fef0f0;
}

.menu-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
  color: white;
  font-size: 16px;
  margin-right: 12px;
  flex-shrink: 0;
}

.menu-title {
  flex: 1;
  font-size: 16px;
  color: #333;
}

.menu-badge {
  background: #ff4757;
  color: white;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
  margin-right: 8px;
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
  .profile-card {
    margin: 10px;
    padding: 15px;
  }

  .stats-section {
    margin: 10px;
    padding: 15px;
  }

  .menu-section {
    margin: 10px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .stat-item {
    padding: 12px;
  }

  .menu-item {
    padding: 12px 15px;
  }
}
</style>