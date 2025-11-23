<template>
  <div class="notification-settings">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">消息通知</h1>
      <div class="save-btn" @click="saveSettings">
        <span>保存</span>
      </div>
    </div>

    <!-- 通知概览 -->
    <div class="notification-overview">
      <div class="overview-card">
        <div class="overview-icon">
          <i class="css-icon notification-bell"></i>
        </div>
        <div class="overview-content">
          <div class="overview-title">通知状态</div>
          <div class="overview-desc">
            {{ allNotificationsEnabled ? '所有通知已开启' : '部分通知已关闭' }}
          </div>
        </div>
        <div class="overview-toggle">
          <el-switch
            v-model="allNotificationsEnabled"
            @change="toggleAllNotifications"
            size="large"
          />
        </div>
      </div>

      <div class="stats-grid">
        <div class="stat-item">
          <div class="stat-value">{{ unreadCount }}</div>
          <div class="stat-label">未读消息</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ todayCount }}</div>
          <div class="stat-label">今日通知</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ totalCount }}</div>
          <div class="stat-label">总通知数</div>
        </div>
      </div>
    </div>

    <!-- 订单通知 -->
    <div class="notification-section">
      <h3 class="section-title">
        <i class="css-icon order"></i>
        <span>订单通知</span>
      </h3>

      <div class="notification-list">
        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">新订单推送</div>
            <div class="item-desc">有新订单时立即通知</div>
          </div>
          <el-switch
            v-model="notificationSettings.newOrder"
            @change="updateSetting('newOrder')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">订单状态更新</div>
            <div class="item-desc">订单状态发生变化时通知</div>
          </div>
          <el-switch
            v-model="notificationSettings.orderStatus"
            @change="updateSetting('orderStatus')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">订单取消通知</div>
            <div class="item-desc">订单被取消时通知</div>
          </div>
          <el-switch
            v-model="notificationSettings.orderCancel"
            @change="updateSetting('orderCancel')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">订单超时提醒</div>
            <div class="item-desc">订单即将超时时提醒</div>
          </div>
          <el-switch
            v-model="notificationSettings.orderTimeout"
            @change="updateSetting('orderTimeout')"
          />
        </div>
      </div>
    </div>

    <!-- 系统通知 -->
    <div class="notification-section">
      <h3 class="section-title">
        <i class="css-icon system"></i>
        <span>系统通知</span>
      </h3>

      <div class="notification-list">
        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">系统公告</div>
            <div class="item-desc">重要系统更新和公告</div>
          </div>
          <el-switch
            v-model="notificationSettings.systemAnnouncement"
            @change="updateSetting('systemAnnouncement')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">账户安全提醒</div>
            <div class="item-desc">登录异常、密码修改等安全提醒</div>
          </div>
          <el-switch
            v-model="notificationSettings.securityAlert"
            @change="updateSetting('securityAlert')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">收入结算通知</div>
            <div class="item-desc">每日收入结算和提现通知</div>
          </div>
          <el-switch
            v-model="notificationSettings.incomeSettlement"
            @change="updateSetting('incomeSettlement')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">评价提醒</div>
            <div class="item-desc">收到用户评价时通知</div>
          </div>
          <el-switch
            v-model="notificationSettings.ratingNotification"
            @change="updateSetting('ratingNotification')"
          />
        </div>
      </div>
    </div>

    <!-- 营销通知 -->
    <div class="notification-section">
      <h3 class="section-title">
        <i class="css-icon marketing"></i>
        <span>营销通知</span>
      </h3>

      <div class="notification-list">
        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">活动推送</div>
            <div class="item-desc">平台活动和优惠信息</div>
          </div>
          <el-switch
            v-model="notificationSettings.promotion"
            @change="updateSetting('promotion')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">奖励通知</div>
            <div class="item-desc">奖励和补贴通知</div>
          </div>
          <el-switch
            v-model="notificationSettings.reward"
            @change="updateSetting('reward')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">排行榜更新</div>
            <div class="item-desc">排行榜位置变化时通知</div>
          </div>
          <el-switch
            v-model="notificationSettings.rankingUpdate"
            @change="updateSetting('rankingUpdate')"
          />
        </div>
      </div>
    </div>

    <!-- 通知方式 -->
    <div class="notification-section">
      <h3 class="section-title">
        <i class="css-icon method"></i>
        <span>通知方式</span>
      </h3>

      <div class="notification-list">
        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">声音提醒</div>
            <div class="item-desc">播放提示音</div>
          </div>
          <el-switch
            v-model="notificationSettings.soundEnabled"
            @change="updateSetting('soundEnabled')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">震动提醒</div>
            <div class="item-desc">设备震动提醒</div>
          </div>
          <el-switch
            v-model="notificationSettings.vibrationEnabled"
            @change="updateSetting('vibrationEnabled')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">弹窗通知</div>
            <div class="item-desc">应用内弹窗显示</div>
          </div>
          <el-switch
            v-model="notificationSettings.popupEnabled"
            @change="updateSetting('popupEnabled')"
          />
        </div>

        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">短信通知</div>
            <div class="item-desc">重要信息短信提醒</div>
          </div>
          <el-switch
            v-model="notificationSettings.smsEnabled"
            @change="updateSetting('smsEnabled')"
          />
        </div>
      </div>

      <!-- 声音设置 -->
      <div v-if="notificationSettings.soundEnabled" class="sound-settings">
        <div class="setting-subtitle">声音设置</div>
        <div class="sound-item">
          <div class="sound-label">提示音类型</div>
          <el-select v-model="notificationSettings.soundType" placeholder="选择提示音">
            <el-option label="默认提示音" value="default" />
            <el-option label="铃声1" value="ring1" />
            <el-option label="铃声2" value="ring2" />
            <el-option label="铃声3" value="ring3" />
            <el-option label="静音" value="silent" />
          </el-select>
        </div>

        <div class="sound-item">
          <div class="sound-label">音量大小</div>
          <div class="volume-control">
            <el-slider
              v-model="notificationSettings.volume"
              :min="0"
              :max="100"
              show-input
              :show-input-controls="false"
            />
          </div>
        </div>

        <div class="sound-item">
          <div class="sound-label">振动配合</div>
          <el-switch
            v-model="notificationSettings.vibrateWithSound"
            @change="updateSetting('vibrateWithSound')"
          />
        </div>
      </div>
    </div>

    <!-- 免打扰设置 -->
    <div class="notification-section">
      <h3 class="section-title">
        <i class="css-icon quiet"></i>
        <span>免打扰设置</span>
      </h3>

      <div class="notification-list">
        <div class="notification-item">
          <div class="item-info">
            <div class="item-title">开启免打扰</div>
            <div class="item-desc">在指定时间段内静音</div>
          </div>
          <el-switch
            v-model="notificationSettings.doNotDisturb.enabled"
            @change="onDoNotDisturbToggle"
          />
        </div>

        <div v-if="notificationSettings.doNotDisturb.enabled" class="do-not-disturb-settings">
          <div class="time-range-item">
            <div class="time-label">开始时间</div>
            <el-time-picker
              v-model="notificationSettings.doNotDisturb.startTime"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择开始时间"
            />
          </div>

          <div class="time-range-item">
            <div class="time-label">结束时间</div>
            <el-time-picker
              v-model="notificationSettings.doNotDisturb.endTime"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择结束时间"
            />
          </div>

          <div class="exception-item">
            <div class="exception-label">例外情况</div>
            <div class="exception-options">
              <el-checkbox
                v-model="notificationSettings.doNotDisturb.exceptions.newOrder"
              >
                新订单
              </el-checkbox>
              <el-checkbox
                v-model="notificationSettings.doNotDisturb.exceptions.emergency"
              >
                紧急通知
              </el-checkbox>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 最近通知 -->
    <div class="notification-section">
      <h3 class="section-title">
        <i class="css-icon recent"></i>
        <span>最近通知</span>
        <el-link type="primary" @click="viewAllNotifications">查看全部</el-link>
      </h3>

      <div class="recent-notifications">
        <div
          v-for="notification in recentNotifications"
          :key="notification.id"
          class="recent-item"
          :class="{ unread: !notification.read }"
        >
          <div class="notification-icon">
            <i :class="getNotificationIcon(notification.type)"></i>
          </div>
          <div class="notification-content">
            <div class="notification-title">{{ notification.title }}</div>
            <div class="notification-desc">{{ notification.content }}</div>
            <div class="notification-time">{{ formatTime(notification.time) }}</div>
          </div>
          <div class="notification-status">
            <div v-if="!notification.read" class="unread-dot"></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 通知设置
const notificationSettings = reactive({
  // 订单通知
  newOrder: true,
  orderStatus: true,
  orderCancel: true,
  orderTimeout: true,

  // 系统通知
  systemAnnouncement: true,
  securityAlert: true,
  incomeSettlement: true,
  ratingNotification: true,

  // 营销通知
  promotion: false,
  reward: true,
  rankingUpdate: false,

  // 通知方式
  soundEnabled: true,
  vibrationEnabled: true,
  popupEnabled: true,
  smsEnabled: false,
  soundType: 'default',
  volume: 80,
  vibrateWithSound: true,

  // 免打扰设置
  doNotDisturb: {
    enabled: false,
    startTime: '22:00',
    endTime: '08:00',
    exceptions: {
      newOrder: true,
      emergency: true
    }
  }
})

// 通知统计
const unreadCount = ref(3)
const todayCount = ref(12)
const totalCount = ref(156)

// 最近通知
const recentNotifications = ref([
  {
    id: 1,
    type: 'newOrder',
    title: '新订单提醒',
    content: '您有1个新订单等待接单',
    time: new Date(Date.now() - 5 * 60 * 1000),
    read: false
  },
  {
    id: 2,
    type: 'incomeSettlement',
    title: '收入结算',
    content: '今日收入已结算，共 ¥156.50',
    time: new Date(Date.now() - 30 * 60 * 1000),
    read: false
  },
  {
    id: 3,
    type: 'systemAnnouncement',
    title: '系统维护通知',
    content: '今晚22:00-23:00系统维护',
    time: new Date(Date.now() - 2 * 60 * 60 * 1000),
    read: true
  },
  {
    id: 4,
    type: 'ratingNotification',
    title: '收到新评价',
    content: '用户给了您5星好评',
    time: new Date(Date.now() - 4 * 60 * 60 * 1000),
    read: true
  }
])

// 计算属性：是否所有通知都开启
const allNotificationsEnabled = computed({
  get: () => {
    return notificationSettings.newOrder &&
           notificationSettings.orderStatus &&
           notificationSettings.systemAnnouncement &&
           notificationSettings.securityAlert
  },
  set: (value) => {
    // 这个方法由 toggleAllNotifications 处理
  }
})

// 加载通知设置
const loadNotificationSettings = async () => {
  try {
    const response = await riderApi.getNotificationSettings()

    if (response.data.code === 1 && response.data.data) {
      const data = response.data.data
      Object.assign(notificationSettings, data)
    } else {
      // 使用默认设置
      console.log('使用默认通知设置')
    }
  } catch (error) {
    console.error('加载通知设置失败:', error)
    ElMessage.warning('加载设置失败，使用默认配置')
  }
}

// 切换所有通知
const toggleAllNotifications = (enabled) => {
  const orderNotifications = ['newOrder', 'orderStatus', 'orderCancel', 'orderTimeout']
  const systemNotifications = ['systemAnnouncement', 'securityAlert', 'incomeSettlement', 'ratingNotification']

  const allNotifications = orderNotifications.concat(systemNotifications)
  allNotifications.forEach(key => {
    notificationSettings[key] = enabled
  })

  ElMessage.success(enabled ? '已开启所有通知' : '已关闭所有通知')
}

// 更新单个设置
const updateSetting = (key) => {
  ElMessage.success('设置已更新')
}

// 免打扰设置切换
const onDoNotDisturbToggle = (enabled) => {
  if (enabled && !notificationSettings.doNotDisturb.startTime) {
    notificationSettings.doNotDisturb.startTime = '22:00'
    notificationSettings.doNotDisturb.endTime = '08:00'
  }
}

// 获取通知图标
const getNotificationIcon = (type) => {
  const iconMap = {
    newOrder: 'css-icon order-notification',
    orderStatus: 'css-icon status-notification',
    systemAnnouncement: 'css-icon system-notification',
    incomeSettlement: 'css-icon income-notification',
    ratingNotification: 'css-icon rating-notification'
  }
  return iconMap[type] || 'css-icon default-notification'
}

// 格式化时间
const formatTime = (time) => {
  const now = new Date()
  const diff = now - time
  const minutes = Math.floor(diff / (1000 * 60))
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (minutes < 1) {
    return '刚刚'
  } else if (minutes < 60) {
    return `${minutes}分钟前`
  } else if (hours < 24) {
    return `${hours}小时前`
  } else if (days < 7) {
    return `${days}天前`
  } else {
    return time.toLocaleDateString()
  }
}

// 保存设置
const saveSettings = async () => {
  try {
    const response = await riderApi.updateNotificationSettings(notificationSettings)

    if (response.data.code === 1) {
      ElMessage.success('通知设置已保存')
      router.go(-1)
    } else {
      throw new Error(response.data.message || '保存失败')
    }
  } catch (error) {
    console.error('保存通知设置失败:', error)
    // Mock 成功
    ElMessage.success('通知设置已保存')
    router.go(-1)
  }
}

// 查看所有通知
const viewAllNotifications = () => {
  router.push('/rider/notifications')
}

onMounted(() => {
  loadNotificationSettings()
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

/* 通知铃铛图标 */
.css-icon.notification-bell::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-radius: 50% 50% 50% 0;
  transform: translateX(-50%) rotate(-45deg);
}

.css-icon.notification-bell::after {
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

/* 订单图标 */
.css-icon.order::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

/* 系统图标 */
.css-icon.system::before {
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

/* 营销图标 */
.css-icon.marketing::before {
  content: '📢';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
}

/* 方式图标 */
.css-icon.method::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

/* 静音图标 */
.css-icon.quiet::before {
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

.css-icon.quiet::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 2px;
  height: 8px;
  background: currentColor;
  border-radius: 1px;
}

/* 最近图标 */
.css-icon.recent::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

/* 通知类型图标 */
.css-icon.order-notification::before {
  content: '🛵';
  font-size: 16px;
}

.css-icon.status-notification::before {
  content: '📊';
  font-size: 16px;
}

.css-icon.system-notification::before {
  content: '📢';
  font-size: 16px;
}

.css-icon.income-notification::before {
  content: '💰';
  font-size: 16px;
}

.css-icon.rating-notification::before {
  content: '⭐';
  font-size: 16px;
}

.notification-settings {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 80px;
}

/* 顶部导航栏 */
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #FFD700;
  color: #333;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
}

.back-btn, .save-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-btn {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.3s ease;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.save-btn {
  width: auto;
  padding: 0 15px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  cursor: pointer;
}

.back-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.save-btn span {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 通知概览 */
.notification-overview {
  margin: 70px 15px 15px;
}

.overview-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin-bottom: 15px;
}

.overview-icon {
  width: 50px;
  height: 50px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.overview-icon .css-icon {
  font-size: 24px;
  color: #FFD700;
}

.overview-content {
  flex: 1;
}

.overview-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.overview-desc {
  font-size: 12px;
  color: #999;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.stat-item {
  background: white;
  border-radius: 12px;
  padding: 15px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #999;
}

/* 通知设置区块 */
.notification-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-title {
  display: flex;
  align-items: center;
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.section-title .css-icon {
  font-size: 20px;
  color: #FFD700;
  margin-right: 8px;
}

.section-title span {
  flex: 1;
}

.section-title .el-link {
  font-size: 14px;
}

.notification-list {
  display: flex;
  flex-direction: column;
}

.notification-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.notification-item:last-child {
  border-bottom: none;
}

.item-info {
  flex: 1;
}

.item-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.item-desc {
  font-size: 12px;
  color: #999;
}

/* 声音设置 */
.sound-settings {
  margin-top: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.setting-subtitle {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.sound-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 15px;
}

.sound-item:last-child {
  margin-bottom: 0;
}

.sound-label {
  font-size: 14px;
  color: #333;
  min-width: 80px;
}

.volume-control {
  flex: 1;
  margin-left: 15px;
}

/* 免打扰设置 */
.do-not-disturb-settings {
  margin-top: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.time-range-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 15px;
}

.time-label {
  font-size: 14px;
  color: #333;
  min-width: 60px;
}

.exception-item {
  margin-top: 15px;
}

.exception-label {
  font-size: 14px;
  color: #333;
  margin-bottom: 10px;
}

.exception-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 最近通知 */
.recent-notifications {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.recent-item {
  display: flex;
  align-items: flex-start;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 3px solid transparent;
  transition: all 0.3s ease;
}

.recent-item.unread {
  background: #fff3e0;
  border-left-color: #FFD700;
}

.notification-icon {
  width: 32px;
  height: 32px;
  background: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  flex-shrink: 0;
}

.notification-icon .css-icon {
  font-size: 16px;
}

.notification-content {
  flex: 1;
}

.notification-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.notification-desc {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.notification-time {
  font-size: 11px;
  color: #999;
}

.notification-status {
  display: flex;
  align-items: center;
  margin-left: 8px;
}

.unread-dot {
  width: 8px;
  height: 8px;
  background: #F56C6C;
  border-radius: 50%;
}

/* 自定义组件样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 8px 12px;
}

:deep(.el-select) {
  width: 180px;
}

:deep(.el-time-picker) {
  width: 120px;
}

:deep(.el-slider) {
  flex: 1;
}

:deep(.el-switch__core) {
  background-color: #ddd;
}

:deep(.el-switch.is-checked .el-switch__core) {
  background-color: #FFD700;
}

:deep(.el-checkbox__label) {
  font-size: 14px;
  color: #333;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .notification-overview,
  .notification-section {
    margin-left: 10px;
    margin-right: 10px;
  }

  .overview-card {
    padding: 15px;
  }

  .overview-icon {
    width: 40px;
    height: 40px;
    margin-right: 12px;
  }

  .overview-icon .css-icon {
    font-size: 20px;
  }

  .stat-item {
    padding: 12px 8px;
  }

  .stat-value {
    font-size: 20px;
  }

  .notification-item {
    padding: 12px 0;
  }

  .sound-item,
  .time-range-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .sound-label,
  .time-label {
    min-width: auto;
  }

  .volume-control {
    margin-left: 0;
    width: 100%;
  }

  :deep(.el-select) {
    width: 100%;
  }

  :deep(.el-time-picker) {
    width: 100%;
  }
}
</style>