<template>
  <div class="rider-profile">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">个人中心</h1>
      <div class="settings-btn" @click="$router.push('/rider/settings')">
        <i class="css-icon settings"></i>
      </div>
    </div>

    <!-- 用户信息卡片 -->
    <div class="profile-header">
      <div class="profile-card">
        <div class="avatar-section">
          <el-avatar :size="80" :src="riderInfo.avatar" @click="changeAvatar" />
          <div class="camera-icon" @click="changeAvatar">
            <i class="css-icon camera"></i>
          </div>
        </div>
        <div class="info-section">
          <h2 class="rider-name">{{ riderInfo.name }}</h2>
          <div class="rider-id">骑手号：{{ riderInfo.id }}</div>
          <div class="rating-section">
            <el-rate v-model="riderInfo.rating" disabled />
            <span class="rating-text">{{ riderInfo.rating }}分</span>
          </div>
        </div>
      </div>

      <!-- 成就徽章 -->
      <div class="achievements">
        <div class="achievement-item">
          <div class="badge gold">🏆</div>
          <div class="badge-text">金牌骑手</div>
        </div>
        <div class="achievement-item">
          <div class="badge speed">⚡</div>
          <div class="badge-text">闪电配送</div>
        </div>
        <div class="achievement-item">
          <div class="badge star">⭐</div>
          <div class="badge-text">五星好评</div>
        </div>
      </div>
    </div>

    <!-- 数据概览 -->
    <div class="data-overview">
      <div class="overview-item">
        <div class="overview-value">{{ riderInfo.completedOrders }}</div>
        <div class="overview-label">累计订单</div>
      </div>
      <div class="overview-item">
        <div class="overview-value">{{ riderInfo.workDays }}</div>
        <div class="overview-label">工作天数</div>
      </div>
      <div class="overview-item">
        <div class="overview-value">{{ riderInfo.totalIncome.toFixed(2) }}</div>
        <div class="overview-label">总收入(元)</div>
      </div>
    </div>

    <!-- 功能菜单 -->
    <div class="menu-sections">
      <!-- 账户管理 -->
      <div class="menu-section">
        <h3 class="section-title">账户管理</h3>
        <div class="menu-list">
          <div class="menu-item" @click="editProfile">
            <div class="menu-icon">
              <i class="css-icon edit"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">个人资料</div>
              <div class="menu-desc">编辑基本信息</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>

          <div class="menu-item" @click="manageAccount">
            <div class="menu-icon">
              <i class="css-icon account"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">账户安全</div>
              <div class="menu-desc">密码、手机绑定</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>

          <div class="menu-item" @click="paymentSettings">
            <div class="menu-icon">
              <i class="css-icon payment"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">收款设置</div>
              <div class="menu-desc">银行卡、支付宝设置</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>
        </div>
      </div>

      <!-- 工作设置 -->
      <div class="menu-section">
        <h3 class="section-title">工作设置</h3>
        <div class="menu-list">
          <div class="menu-item" @click="workSettings">
            <div class="menu-icon">
              <i class="css-icon work"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">工作偏好</div>
              <div class="menu-desc">配送范围、工作时间</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>

          <div class="menu-item" @click="notificationSettings">
            <div class="menu-icon">
              <i class="css-icon notification"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">消息通知</div>
              <div class="menu-desc">新订单、系统通知</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>

          <div class="menu-item" @click="mapSettings">
            <div class="menu-icon">
              <i class="css-icon map"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">地图设置</div>
              <div class="menu-desc">导航偏好、语音设置</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>
        </div>
      </div>

      <!-- 帮助与反馈 -->
      <div class="menu-section">
        <h3 class="section-title">帮助与反馈</h3>
        <div class="menu-list">
          <div class="menu-item" @click="helpCenter">
            <div class="menu-icon">
              <i class="css-icon help"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">帮助中心</div>
              <div class="menu-desc">常见问题、使用指南</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>

          <div class="menu-item" @click="feedback">
            <div class="menu-icon">
              <i class="css-icon feedback"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">意见反馈</div>
              <div class="menu-desc">问题建议、功能需求</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>

          <div class="menu-item" @click="contactService">
            <div class="menu-icon">
              <i class="css-icon service"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">联系客服</div>
              <div class="menu-desc">400-123-4567</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>
        </div>
      </div>

      <!-- 关于 -->
      <div class="menu-section">
        <h3 class="section-title">关于</h3>
        <div class="menu-list">
          <div class="menu-item" @click="aboutUs">
            <div class="menu-icon">
              <i class="css-icon about"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">关于我们</div>
              <div class="menu-desc">版本 v1.0.0</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>

          <div class="menu-item" @click="privacyPolicy">
            <div class="menu-icon">
              <i class="css-icon privacy"></i>
            </div>
            <div class="menu-content">
              <div class="menu-title">隐私政策</div>
              <div class="menu-desc">用户协议、隐私条款</div>
            </div>
            <div class="menu-arrow">
              <i class="css-icon arrow"></i>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 退出登录按钮 -->
    <div class="logout-section">
      <el-button type="danger" @click="logout" class="logout-btn">
        退出登录
      </el-button>
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 骑手信息
const riderInfo = ref({
  id: 'R001',
  name: '李骑手',
  avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
  rating: 4.8,
  completedOrders: 1250,
  workDays: 186,
  totalIncome: 15680.50
})

// 加载骑手信息
const loadRiderInfo = async () => {
  try {
    const response = await riderApi.getRiderInfoWithDemo()

    if (response.code === 1 && response.data) {
      riderInfo.value = {
        ...riderInfo.value,
        ...response.data
      }
    }
  } catch (error) {
    console.error('加载骑手信息失败:', error)
    // 使用Demo数据
  }
}

// 菜单点击事件处理
const changeAvatar = () => {
  ElMessage.info('更换头像功能开发中...')
}

const editProfile = () => {
  router.push('/rider/profile/edit')
}

const manageAccount = () => {
  router.push('/rider/profile/security')
}

const paymentSettings = () => {
  router.push('/rider/profile/payment')
}

const workSettings = () => {
  router.push('/rider/profile/work')
}

const notificationSettings = () => {
  router.push('/rider/profile/notification')
}

const mapSettings = () => {
  router.push('/rider/profile/map')
}

const helpCenter = () => {
  ElMessage.info('帮助中心功能开发中...')
}

const feedback = () => {
  ElMessage.info('意见反馈功能开发中...')
}

const contactService = () => {
  ElMessage.info('正在拨打客服电话：400-123-4567')
}

const aboutUs = () => {
  ElMessage.info('关于我们功能开发中...')
}

const privacyPolicy = () => {
  ElMessage.info('隐私政策功能开发中...')
}

// 退出登录
const logout = async () => {
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
    localStorage.removeItem('riderInfo')

    ElMessage.success('退出登录成功')

    // 跳转到登录页面
    router.push('/login')
  } catch (error) {
    // 用户取消操作
  }
}

onMounted(() => {
  loadRiderInfo()
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

/* 设置图标 */
.css-icon.settings::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.settings::after {
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

/* 相机图标 */
.css-icon.camera::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 4px;
}

.css-icon.camera::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 8px;
  height: 6px;
  background: currentColor;
  border-radius: 0 0 4px 4px;
}

/* 编辑图标 */
.css-icon.edit::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-45deg);
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.edit::after {
  content: '';
  position: absolute;
  bottom: 2px;
  left: 2px;
  width: 8px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
  transform: rotate(45deg);
}

/* 账户图标 */
.css-icon.account::before {
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

.css-icon.account::after {
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

/* 支付图标 */
.css-icon.payment::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.payment::after {
  content: '';
  position: absolute;
  bottom: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 1px;
  background: currentColor;
  border-radius: 1px;
}

/* 箭头图标 */
.css-icon.arrow::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(45deg);
  width: 8px;
  height: 8px;
  border-right: 2px solid currentColor;
  border-top: 2px solid currentColor;
}

/* 工作图标 */
.css-icon.work::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.work::after {
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

/* 通知图标 */
.css-icon.notification::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 50% 50% 50% 0;
  transform: translateX(-50%) rotate(-45deg);
}

.css-icon.notification::after {
  content: '';
  position: absolute;
  bottom: 0;
  right: 0;
  width: 6px;
  height: 6px;
  background: #F56C6C;
  border: 1px solid white;
  border-radius: 50%;
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

/* 帮助图标 */
.css-icon.help::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.help::after {
  content: '?';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: bold;
}

/* 反馈图标 */
.css-icon.feedback::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 8px 8px 0 0;
}

.css-icon.feedback::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 10px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
}

/* 客服图标 */
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

/* 关于图标 */
.css-icon.about::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.about::after {
  content: 'i';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: bold;
  font-style: italic;
}

/* 隐私图标 */
.css-icon.privacy::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 18px;
  border: 2px solid currentColor;
  border-radius: 8px 8px 0 0;
}

.css-icon.privacy::after {
  content: '';
  position: absolute;
  top: 4px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 2px;
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

.rider-profile {
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

.back-btn, .settings-btn {
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

.back-btn:hover, .settings-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.back-btn .css-icon, .settings-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 用户信息卡片 */
.profile-header {
  padding: 20px 15px;
  background: white;
}

.profile-card {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.avatar-section {
  position: relative;
  margin-right: 20px;
}

.camera-icon {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 24px;
  height: 24px;
  background: #FFD700;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border: 2px solid white;
}

.camera-icon .css-icon {
  font-size: 12px;
  color: #333;
}

.info-section {
  flex: 1;
}

.rider-name {
  margin: 0 0 5px 0;
  font-size: 20px;
  color: #333;
  font-weight: 500;
}

.rider-id {
  font-size: 14px;
  color: #999;
  margin-bottom: 8px;
}

.rating-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rating-text {
  font-size: 14px;
  color: #666;
}

/* 成就徽章 */
.achievements {
  display: flex;
  justify-content: space-around;
}

.achievement-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
}

.badge {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.badge.gold {
  background: linear-gradient(135deg, #FFD700, #FFA500);
}

.badge.speed {
  background: linear-gradient(135deg, #409EFF, #67C23A);
}

.badge.star {
  background: linear-gradient(135deg, #E6A23C, #F56C6C);
}

.badge-text {
  font-size: 12px;
  color: #666;
}

/* 数据概览 */
.data-overview {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  padding: 15px;
  background: white;
  border-top: 1px solid #f0f0f0;
}

.overview-item {
  text-align: center;
  padding: 15px 0;
}

.overview-value {
  font-size: 20px;
  font-weight: bold;
  color: #333;
  margin-bottom: 5px;
}

.overview-label {
  font-size: 12px;
  color: #666;
}

/* 功能菜单 */
.menu-sections {
  padding: 15px;
}

.menu-section {
  background: white;
  border-radius: 12px;
  margin-bottom: 15px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-title {
  margin: 0;
  padding: 15px 20px 10px;
  font-size: 16px;
  color: #333;
  font-weight: 500;
}

.menu-list {
  padding: 0 20px;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-item:hover {
  background: #f8f9fa;
  margin: 0 -20px;
  padding-left: 20px;
  padding-right: 20px;
}

.menu-icon {
  width: 40px;
  height: 40px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.menu-icon .css-icon {
  font-size: 20px;
  color: #FFD700;
}

.menu-content {
  flex: 1;
}

.menu-title {
  font-size: 16px;
  color: #333;
  margin-bottom: 2px;
}

.menu-desc {
  font-size: 12px;
  color: #999;
}

.menu-arrow {
  width: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.menu-arrow .css-icon {
  font-size: 16px;
  color: #ccc;
}

/* 退出登录 */
.logout-section {
  padding: 20px 15px;
}

.logout-btn {
  width: 100%;
  height: 50px;
  font-size: 16px;
  border-radius: 25px;
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
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .menu-item:hover {
    margin: 0;
    padding-left: 0;
    padding-right: 0;
  }
}
</style>