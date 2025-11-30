<template>
  <div class="help-center">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">帮助中心</h1>
      <div class="search-btn" @click="showSearch = true">
        <i class="css-icon search"></i>
      </div>
    </div>

    <!-- 搜索框 -->
    <div v-if="showSearch" class="search-section">
      <div class="search-container">
        <el-input
          v-model="searchQuery"
          placeholder="搜索您的问题..."
          class="search-input"
          @keyup.enter="performSearch"
        >
          <template #suffix>
            <el-button @click="performSearch" type="primary">搜索</el-button>
          </template>
        </el-input>
      </div>
    </div>

    <!-- 快捷入口 -->
    <div class="quick-entry">
      <div class="entry-title">快捷帮助</div>
      <div class="entry-grid">
        <div v-for="entry in quickEntries" :key="entry.id" class="entry-item" @click="navigateToHelp(entry)">
          <div class="entry-icon">
            <i :class="entry.iconClass"></i>
          </div>
          <div class="entry-name">{{ entry.name }}</div>
        </div>
      </div>
    </div>

    <!-- 常见问题 -->
    <div class="faq-section">
      <div class="section-title">
        <span>常见问题</span>
        <el-link type="primary" @click="viewAllFAQ">查看全部</el-link>
      </div>

      <div class="faq-list">
        <div
          v-for="faq in displayFAQs"
          :key="faq.id"
          class="faq-item"
          @click="toggleFAQ(faq)"
        >
          <div class="faq-question">
            <span>{{ faq.question }}</span>
            <div class="faq-arrow" :class="{ expanded: faq.expanded }">
              <i class="css-icon arrow"></i>
            </div>
          </div>
          <div v-if="faq.expanded" class="faq-answer">
            {{ faq.answer }}
          </div>
        </div>
      </div>
    </div>

    <!-- 新手指南 -->
    <div class="guide-section">
      <div class="section-title">
        <span>新手指南</span>
        <el-link type="primary" @click="viewAllGuides">查看全部</el-link>
      </div>

      <div class="guide-list">
        <div v-for="guide in guides" :key="guide.id" class="guide-item" @click="openGuide(guide)">
          <div class="guide-icon">
            <i class="css-icon guide"></i>
          </div>
          <div class="guide-content">
            <div class="guide-title">{{ guide.title }}</div>
            <div class="guide-desc">{{ guide.description }}</div>
            <div class="guide-meta">
              <span class="guide-time">{{ guide.readTime }}分钟阅读</span>
              <span class="guide-views">{{ guide.viewCount }}次浏览</span>
            </div>
          </div>
          <div class="guide-arrow">
            <i class="css-icon arrow"></i>
          </div>
        </div>
      </div>
    </div>

    <!-- 视频教程 -->
    <div class="video-section">
      <div class="section-title">
        <span>视频教程</span>
        <el-link type="primary" @click="viewAllVideos">查看全部</el-link>
      </div>

      <div class="video-grid">
        <div v-for="video in videos" :key="video.id" class="video-item" @click="playVideo(video)">
          <div class="video-thumbnail">
            <img :src="video.thumbnail" :alt="video.title" />
            <div class="video-play-btn">
              <i class="css-icon play"></i>
            </div>
            <div class="video-duration">{{ video.duration }}</div>
          </div>
          <div class="video-info">
            <div class="video-title">{{ video.title }}</div>
            <div class="video-views">{{ video.viewCount }}次播放</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 联系方式 -->
    <div class="contact-section">
      <div class="section-title">联系我们</div>
      <div class="contact-list">
        <div class="contact-item" @click="callPhone('400-123-4567')">
          <div class="contact-icon">
            <i class="css-icon phone"></i>
          </div>
          <div class="contact-content">
            <div class="contact-title">客服热线</div>
            <div class="contact-desc">400-123-4567</div>
            <div class="contact-time">工作日 9:00-21:00</div>
          </div>
        </div>

        <div class="contact-item" @click="openOnlineChat">
          <div class="contact-icon">
            <i class="css-icon chat"></i>
          </div>
          <div class="contact-content">
            <div class="contact-title">在线客服</div>
            <div class="contact-desc">7x24小时在线</div>
            <div class="contact-status">
              <span class="status-dot online"></span>
              <span class="status-text">在线</span>
            </div>
          </div>
        </div>

        <div class="contact-item" @click="sendEmail">
          <div class="contact-icon">
            <i class="css-icon email"></i>
          </div>
          <div class="contact-content">
            <div class="contact-title">邮箱反馈</div>
            <div class="contact-desc">rider-support@example.com</div>
            <div class="contact-time">24小时内回复</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统公告 -->
    <div class="announcement-section">
      <div class="section-title">
        <span>系统公告</span>
        <el-link type="primary" @click="viewAllAnnouncements">查看全部</el-link>
      </div>

      <div class="announcement-list">
        <div v-for="announcement in announcements" :key="announcement.id" class="announcement-item" @click="openAnnouncement(announcement)">
          <div class="announcement-badge" :class="announcement.type">
            {{ getAnnouncementBadgeText(announcement.type) }}
          </div>
          <div class="announcement-content">
            <div class="announcement-title">{{ announcement.title }}</div>
            <div class="announcement-time">{{ formatTime(announcement.publishTime) }}</div>
          </div>
          <div class="announcement-arrow">
            <i class="css-icon arrow"></i>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- 视频播放对话框 -->
  <el-dialog
    v-model="videoDialog.visible"
    :title="videoDialog.title"
    width="90%"
    :before-close="closeVideoDialog"
  >
    <div class="video-player">
      <video
        ref="videoRef"
        :src="videoDialog.url"
        controls
        width="100%"
        height="400"
        @ended="onVideoEnded"
      />
    </div>
    <div class="video-info">
      <h3>{{ videoDialog.title }}</h3>
      <p>{{ videoDialog.description }}</p>
    </div>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 搜索相关
const showSearch = ref(false)
const searchQuery = ref('')

// 快捷入口
const quickEntries = [
  {
    id: 1,
    name: '接单流程',
    iconClass: 'css-icon order-process',
    route: '/rider/help/order-process'
  },
  {
    id: 2,
    name: '配送规范',
    iconClass: 'css-icon delivery-standard',
    route: '/rider/help/delivery-standard'
  },
  {
    id: 3,
    name: '收入提现',
    iconClass: 'css-icon income-withdraw',
    route: '/rider/help/income-withdraw'
  },
  {
    id: 4,
    name: '账号问题',
    iconClass: 'css-icon account-issue',
    route: '/rider/help/account-issue'
  }
]

// 常见问题
const faqs = ref([
  {
    id: 1,
    question: '如何开始接单？',
    answer: '首先确保您已完成实名认证并设置好收款方式。然后在工作台点击"上线接单"，系统就会为您推送附近的订单。',
    expanded: false
  },
  {
    id: 2,
    question: '配送费如何计算？',
    answer: '配送费由基础配送费、距离费、时段费、重量费等组成。系统会根据订单距离、配送时段、商品重量等因素自动计算。',
    expanded: false
  },
  {
    id: 3,
    question: '如何提现收入？',
    answer: '在钱包页面点击"提现"，选择收款账户并输入提现金额。提现申请提交后，一般1-3个工作日到账。',
    expanded: false
  },
  {
    id: 4,
    question: '订单超时怎么办？',
    answer: '如遇特殊情况可能导致超时，请及时在订单页面点击"异常报告"并说明原因，平台会根据实际情况进行处理。',
    expanded: false
  }
])

// 新手指南
const guides = [
  {
    id: 1,
    title: '骑手快速入门指南',
    description: '从注册到开始赚钱的完整流程',
    readTime: 5,
    viewCount: 1280,
    content: '/rider/help/guide/quick-start'
  },
  {
    id: 2,
    title: '配送路线优化技巧',
    description: '如何规划最优路线提高效率',
    readTime: 8,
    viewCount: 856,
    content: '/rider/help/guide/route-optimization'
  },
  {
    id: 3,
    title: '服务规范与礼仪',
    description: '提供优质服务获得好评',
    readTime: 6,
    viewCount: 643,
    content: '/rider/help/guide/service-standard'
  }
]

// 视频教程
const videos = [
  {
    id: 1,
    title: '新骑手必看：接单流程演示',
    thumbnail: 'https://via.placeholder.com/200x120?text=视频1',
    duration: '3:45',
    viewCount: 2340,
    url: 'https://example.com/video1.mp4'
  },
  {
    id: 2,
    title: '配送安全注意事项',
    thumbnail: 'https://via.placeholder.com/200x120?text=视频2',
    duration: '5:20',
    viewCount: 1876,
    url: 'https://example.com/video2.mp4'
  },
  {
    id: 3,
    title: 'App使用技巧',
    thumbnail: 'https://via.placeholder.com/200x120?text=视频3',
    duration: '4:15',
    viewCount: 1543,
    url: 'https://example.com/video3.mp4'
  },
  {
    id: 4,
    title: '异常处理指南',
    thumbnail: 'https://via.placeholder.com/200x120?text=视频4',
    duration: '6:30',
    viewCount: 987,
    url: 'https://example.com/video4.mp4'
  }
]

// 系统公告
const announcements = [
  {
    id: 1,
    title: '春节期间配送服务安排通知',
    type: 'notice',
    publishTime: new Date('2024-01-15'),
    content: '春节期间配送服务调整...'
  },
  {
    id: 2,
    title: '平台系统升级维护公告',
    type: 'maintenance',
    publishTime: new Date('2024-01-10'),
    content: '系统将于本周六凌晨进行升级...'
  },
  {
    id: 3,
    title: '新的奖励政策发布',
    type: 'update',
    publishTime: new Date('2024-01-08'),
    content: '平台推出新的奖励政策...'
  }
]

// 视频对话框
const videoDialog = reactive({
  visible: false,
  title: '',
  url: '',
  description: ''
})

const videoRef = ref(null)

// 显示的FAQ数量
const displayFAQs = computed(() => {
  return faqs.value.slice(0, 4)
})

// 加载帮助数据
const loadHelpData = async () => {
  try {
    // 模拟API调用
    // const response = await riderApi.getHelpData()

    // 使用默认数据
    console.log('使用默认帮助数据')
  } catch (error) {
    console.error('加载帮助数据失败:', error)
    ElMessage.warning('加载数据失败，显示默认内容')
  }
}

// 搜索
const performSearch = () => {
  if (!searchQuery.value.trim()) {
    ElMessage.warning('请输入搜索关键词')
    return
  }
  ElMessage.info(`搜索 "${searchQuery.value}" 的相关内容...`)
  // 实际搜索逻辑
}

// 导航到帮助页面
const navigateToHelp = (entry) => {
  router.push(entry.route)
}

// 切换FAQ展开状态
const toggleFAQ = (faq) => {
  faq.expanded = !faq.expanded
}

// 查看全部FAQ
const viewAllFAQ = () => {
  router.push('/rider/help/faq')
}

// 打开指南
const openGuide = (guide) => {
  router.push(guide.content)
}

// 查看全部指南
const viewAllGuides = () => {
  router.push('/rider/help/guides')
}

// 播放视频
const playVideo = (video) => {
  videoDialog.visible = true
  videoDialog.title = video.title
  videoDialog.url = video.url
  videoDialog.description = `这是一个关于${video.title}的详细教程视频。`
}

// 关闭视频对话框
const closeVideoDialog = () => {
  if (videoRef.value) {
    videoRef.value.pause()
  }
  videoDialog.visible = false
  videoDialog.title = ''
  videoDialog.url = ''
  videoDialog.description = ''
}

// 视频播放结束
const onVideoEnded = () => {
  ElMessage.success('视频播放完成')
}

// 查看全部视频
const viewAllVideos = () => {
  router.push('/rider/help/videos')
}

// 拨打电话
const callPhone = (phone) => {
  ElMessage.info(`正在拨打客服电话：${phone}`)
  // 实际拨打电话逻辑
  // window.location.href = `tel:${phone}`
}

// 打开在线客服
const openOnlineChat = () => {
  ElMessage.info('正在连接在线客服...')
  // 实际打开在线客服逻辑
}

// 发送邮件
const sendEmail = () => {
  ElMessage.info('正在打开邮件客户端...')
  // 实际发送邮件逻辑
  // window.location.href = 'mailto:rider-support@example.com'
}

// 打开公告
const openAnnouncement = (announcement) => {
  router.push(`/rider/help/announcement/${announcement.id}`)
}

// 查看全部公告
const viewAllAnnouncements = () => {
  router.push('/rider/help/announcements')
}

// 获取公告徽章文本
const getAnnouncementBadgeText = (type) => {
  const badgeMap = {
    notice: '通知',
    maintenance: '维护',
    update: '更新'
  }
  return badgeMap[type] || '公告'
}

// 格式化时间
const formatTime = (time) => {
  const now = new Date()
  const diff = now - time
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) {
    return '今天'
  } else if (days === 1) {
    return '昨天'
  } else if (days < 7) {
    return `${days}天前`
  } else {
    return time.toLocaleDateString()
  }
}

onMounted(() => {
  loadHelpData()
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

/* 搜索图标 */
.css-icon.search::before {
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

.css-icon.search::after {
  content: '';
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 6px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
  transform: rotate(45deg);
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

/* 接单流程图标 */
.css-icon.order-process::before {
  content: '📦';
  font-size: 24px;
}

/* 配送规范图标 */
.css-icon.delivery-standard::before {
  content: '🛵';
  font-size: 24px;
}

/* 收入提现图标 */
.css-icon.income-withdraw::before {
  content: '💰';
  font-size: 24px;
}

/* 账号问题图标 */
.css-icon.account-issue::before {
  content: '👤';
  font-size: 24px;
}

/* 指南图标 */
.css-icon.guide::before {
  content: '📖';
  font-size: 20px;
}

/* 播放图标 */
.css-icon.play::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-30%, -50%);
  width: 0;
  height: 0;
  border-top: 8px solid transparent;
  border-bottom: 8px solid transparent;
  border-left: 12px solid white;
}

/* 电话图标 */
.css-icon.phone::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 20px;
  border: 2px solid currentColor;
  border-radius: 4px;
}

/* 聊天图标 */
.css-icon.chat::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 8px 8px 0 0;
}

/* 邮箱图标 */
.css-icon.email::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 16px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.help-center {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 20px;
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

.back-btn, .search-btn {
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

.search-btn {
  cursor: pointer;
}

.back-btn .css-icon,
.search-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 搜索框 */
.search-section {
  position: fixed;
  top: 70px;
  left: 0;
  right: 0;
  background: white;
  padding: 15px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  z-index: 99;
}

.search-container {
  max-width: 600px;
  margin: 0 auto;
}

/* 主要内容区域 */
.help-center > *:not(.header-bar):not(.search-section) {
  margin-top: 140px;
}

/* 快捷入口 */
.quick-entry {
  background: white;
  margin: 15px;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.entry-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.entry-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 15px;
}

.entry-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  cursor: pointer;
  transition: transform 0.3s ease;
}

.entry-item:hover {
  transform: translateY(-2px);
}

.entry-icon {
  width: 50px;
  height: 50px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}

.entry-icon .css-icon {
  font-size: 24px;
  color: #FFD700;
}

.entry-name {
  font-size: 12px;
  color: #333;
}

/* 区块样式 */
.faq-section,
.guide-section,
.video-section,
.contact-section,
.announcement-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 15px;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

/* 常见问题 */
.faq-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.faq-item {
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
}

.faq-question {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #f8f9fa;
  font-size: 14px;
  color: #333;
  transition: background 0.3s ease;
}

.faq-question:hover {
  background: #e9ecef;
}

.faq-arrow {
  transition: transform 0.3s ease;
}

.faq-arrow.expanded {
  transform: rotate(90deg);
}

.faq-answer {
  padding: 15px;
  font-size: 13px;
  color: #666;
  line-height: 1.6;
  border-top: 1px solid #f0f0f0;
}

/* 新手指南 */
.guide-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.guide-item {
  display: flex;
  align-items: center;
  padding: 15px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.guide-item:hover {
  border-color: #FFD700;
  background: #fffbf0;
}

.guide-icon {
  width: 40px;
  height: 40px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.guide-icon .css-icon {
  font-size: 20px;
  color: #FFD700;
}

.guide-content {
  flex: 1;
}

.guide-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.guide-desc {
  font-size: 12px;
  color: #999;
  margin-bottom: 6px;
}

.guide-meta {
  display: flex;
  gap: 15px;
}

.guide-time,
.guide-views {
  font-size: 11px;
  color: #999;
}

.guide-arrow {
  display: flex;
  align-items: center;
  color: #ccc;
}

/* 视频教程 */
.video-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}

.video-item {
  cursor: pointer;
  transition: transform 0.3s ease;
}

.video-item:hover {
  transform: translateY(-2px);
}

.video-thumbnail {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 8px;
}

.video-thumbnail img {
  width: 100%;
  height: 120px;
  object-fit: cover;
}

.video-play-btn {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 40px;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.video-title {
  font-size: 13px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
  line-height: 1.4;
}

.video-views {
  font-size: 11px;
  color: #999;
}

/* 联系方式 */
.contact-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.contact-item {
  display: flex;
  align-items: center;
  padding: 15px;
  border: 1px solid #f0f0f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.contact-item:hover {
  border-color: #FFD700;
  background: #fffbf0;
}

.contact-icon {
  width: 40px;
  height: 40px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.contact-icon .css-icon {
  font-size: 20px;
  color: #FFD700;
}

.contact-content {
  flex: 1;
}

.contact-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.contact-desc {
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.contact-time {
  font-size: 11px;
  color: #999;
}

.contact-status {
  display: flex;
  align-items: center;
  gap: 4px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-dot.online {
  background: #67C23A;
}

.status-text {
  font-size: 11px;
  color: #67C23A;
}

/* 系统公告 */
.announcement-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.announcement-item {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.3s ease;
}

.announcement-item:last-child {
  border-bottom: none;
}

.announcement-item:hover {
  background: #f8f9fa;
  margin: 0 -20px;
  padding-left: 20px;
  padding-right: 20px;
}

.announcement-badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  margin-right: 12px;
  min-width: 40px;
  text-align: center;
}

.announcement-badge.notice {
  background: #e1f3ff;
  color: #409eff;
}

.announcement-badge.maintenance {
  background: #fff3e0;
  color: #e6a23c;
}

.announcement-badge.update {
  background: #f0f9ff;
  color: #67c23a;
}

.announcement-content {
  flex: 1;
}

.announcement-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.announcement-time {
  font-size: 11px;
  color: #999;
}

.announcement-arrow {
  display: flex;
  align-items: center;
  color: #ccc;
}

/* 视频播放对话框 */
.video-player {
  margin-bottom: 20px;
}

.video-info h3 {
  margin: 0 0 10px 0;
  font-size: 16px;
  color: #333;
}

.video-info p {
  margin: 0;
  font-size: 14px;
  color: #666;
  line-height: 1.6;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .quick-entry,
  .faq-section,
  .guide-section,
  .video-section,
  .contact-section,
  .announcement-section {
    margin: 10px;
    padding: 15px;
  }

  .entry-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .entry-icon {
    width: 40px;
    height: 40px;
  }

  .entry-icon .css-icon {
    font-size: 20px;
  }

  .video-grid {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .video-thumbnail img {
    height: 100px;
  }

  .guide-meta {
    flex-direction: column;
    gap: 4px;
  }
}
</style>