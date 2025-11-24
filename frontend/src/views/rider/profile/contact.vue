<template>
  <div class="contact-service">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">联系客服</h1>
      <div class="placeholder"></div>
    </div>

    <!-- 快捷联系 -->
    <div class="quick-contact">
      <div class="contact-title">快捷联系</div>
      <div class="contact-cards">
        <!-- 电话客服 -->
        <div class="contact-card" @click="callPhone('400-123-4567')">
          <div class="card-icon">
            <i class="css-icon phone"></i>
          </div>
          <div class="card-content">
            <div class="card-title">电话客服</div>
            <div class="card-desc">400-123-4567</div>
            <div class="card-time">工作日 9:00-21:00</div>
          </div>
          <div class="card-status">
            <span class="status-text">立即拨打</span>
          </div>
        </div>

        <!-- 在线客服 -->
        <div class="contact-card" @click="openOnlineChat">
          <div class="card-icon online">
            <i class="css-icon chat"></i>
          </div>
          <div class="card-content">
            <div class="card-title">在线客服</div>
            <div class="card-desc">7x24小时在线</div>
            <div class="card-status-info">
              <span class="status-dot online"></span>
              <span class="status-text">在线</span>
            </div>
          </div>
          <div class="card-status">
            <span class="status-text">立即对话</span>
          </div>
        </div>

        <!-- 微信客服 -->
        <div class="contact-card" @click="showWechatQR = true">
          <div class="card-icon wechat">
            <i class="css-icon wechat"></i>
          </div>
          <div class="card-content">
            <div class="card-title">微信客服</div>
            <div class="card-desc">扫码添加客服微信</div>
            <div class="card-time">工作日 9:00-18:00</div>
          </div>
          <div class="card-status">
            <span class="status-text">扫码添加</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 常见问题 -->
    <div class="common-issues">
      <div class="issues-title">常见问题快速解决</div>
      <div class="issues-list">
        <div v-for="issue in commonIssues" :key="issue.id" class="issue-item" @click="viewIssue(issue)">
          <div class="issue-icon">
            <i :class="issue.iconClass"></i>
          </div>
          <div class="issue-content">
            <div class="issue-title">{{ issue.title }}</div>
            <div class="issue-count">{{ issue.viewCount }}人查看</div>
          </div>
          <div class="issue-arrow">
            <i class="css-icon arrow"></i>
          </div>
        </div>
      </div>
    </div>

    <!-- 其他联系方式 -->
    <div class="other-contacts">
      <div class="contacts-title">其他联系方式</div>
      <div class="contacts-grid">
        <!-- 邮箱 -->
        <div class="contact-item" @click="sendEmail">
          <div class="contact-icon">
            <i class="css-icon email"></i>
          </div>
          <div class="contact-info">
            <div class="contact-name">邮箱反馈</div>
            <div class="contact-detail">rider-support@example.com</div>
            <div class="contact-desc">24小时内回复</div>
          </div>
        </div>

        <!-- 微博 -->
        <div class="contact-item" @click="openWeibo">
          <div class="contact-icon">
            <i class="css-icon weibo"></i>
          </div>
          <div class="contact-info">
            <div class="contact-name">官方微博</div>
            <div class="contact-detail">@校园配送骑手</div>
            <div class="contact-desc">@我们获得帮助</div>
          </div>
        </div>

        <!-- QQ群 -->
        <div class="contact-item" @click="copyQQGroup">
          <div class="contact-icon">
            <i class="css-icon qq"></i>
          </div>
          <div class="contact-info">
            <div class="contact-name">QQ交流群</div>
            <div class="contact-detail">123456789</div>
            <div class="contact-desc">点击复制群号</div>
          </div>
        </div>

        <!-- 意见反馈 -->
        <div class="contact-item" @click="goToFeedback">
          <div class="contact-icon">
            <i class="css-icon feedback"></i>
          </div>
          <div class="contact-info">
            <div class="contact-name">意见反馈</div>
            <div class="contact-detail">提交问题建议</div>
            <div class="contact-desc">帮助我们改进</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 客服工作时间 -->
    <div class="service-hours">
      <div class="hours-title">
        <i class="css-icon clock"></i>
        <span>客服工作时间</span>
      </div>
      <div class="hours-content">
        <div class="hours-item">
          <div class="hours-label">电话客服</div>
          <div class="hours-time">周一至周日 9:00-21:00</div>
        </div>
        <div class="hours-item">
          <div class="hours-label">在线客服</div>
          <div class="hours-time">7x24小时在线</div>
        </div>
        <div class="hours-item">
          <div class="hours-label">微信客服</div>
          <div class="hours-time">周一至周五 9:00-18:00</div>
        </div>
        <div class="hours-item">
          <div class="hours-label">邮箱反馈</div>
          <div class="hours-time">24小时内回复</div>
        </div>
      </div>
    </div>

    <!-- 紧急联系 -->
    <div class="emergency-contact">
      <div class="emergency-title">
        <i class="css-icon warning"></i>
        <span>紧急情况处理</span>
      </div>
      <div class="emergency-content">
        <div class="emergency-desc">
          如遇到安全事故、人身威胁等紧急情况，请立即拨打：
        </div>
        <div class="emergency-phone" @click="callPhone('110')">
          <i class="css-icon police"></i>
          <span>110 - 报警电话</span>
        </div>
        <div class="emergency-phone" @click="callPhone('120')">
          <i class="css-icon ambulance"></i>
          <span>120 - 急救电话</span>
        </div>
        <div class="emergency-tip">
          同时请尽快联系平台客服：400-123-4567
        </div>
      </div>
    </div>

    <!-- 微信二维码弹窗 -->
    <el-dialog
      v-model="showWechatQR"
      title="微信客服"
      width="90%"
      :before-close="closeWechatDialog"
    >
      <div class="wechat-qr-content">
        <div class="qr-code">
          <img src="https://via.placeholder.com/200x200?text=微信客服二维码" alt="微信客服二维码" />
        </div>
        <div class="qr-info">
          <div class="qr-title">扫码添加客服微信</div>
          <div class="qr-desc">工作时间：周一至周五 9:00-18:00</div>
          <div class="qr-tip">截图保存二维码，打开微信扫一扫添加</div>
          <el-button type="primary" @click="saveQRCode" class="save-qr-btn">保存二维码</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()

// 显示微信二维码
const showWechatQR = ref(false)

// 常见问题
const commonIssues = [
  {
    id: 1,
    title: '如何申请提现？',
    iconClass: 'css-icon issue-money',
    viewCount: 523,
    route: '/rider/help/withdraw'
  },
  {
    id: 2,
    title: '订单超时怎么处理？',
    iconClass: 'css-icon issue-timeout',
    viewCount: 412,
    route: '/rider/help/timeout'
  },
  {
    id: 3,
    title: '修改个人信息',
    iconClass: 'css-icon issue-profile',
    viewCount: 356,
    route: '/rider/help/profile'
  },
  {
    id: 4,
    title: '配送路线问题',
    iconClass: 'css-icon issue-route',
    viewCount: 298,
    route: '/rider/help/route'
  },
  {
    id: 5,
    title: '收入结算异常',
    iconClass: 'css-icon issue-income',
    viewCount: 245,
    route: '/rider/help/income'
  }
]

// 拨打电话
const callPhone = (phone) => {
  ElMessageBox.confirm(
    `确定要拨打客服电话 ${phone} 吗？`,
    '提示',
    {
      confirmButtonText: '确定拨打',
      cancelButtonText: '取消',
      type: 'info'
    }
  ).then(() => {
    ElMessage.info(`正在拨打客服电话：${phone}`)
    // 实际拨打电话
    // window.location.href = `tel:${phone}`
  }).catch(() => {
    // 用户取消
  })
}

// 打开在线客服
const openOnlineChat = () => {
  ElMessage.info('正在连接在线客服...')
  // 实际打开在线客服逻辑
  setTimeout(() => {
    ElMessage.success('已连接在线客服')
    // 可以跳转到聊天页面或打开聊天窗口
    // router.push('/rider/chat')
  }, 1500)
}

// 发送邮件
const sendEmail = () => {
  ElMessage.info('正在打开邮件客户端...')
  // 实际发送邮件逻辑
  // window.location.href = 'mailto:rider-support@example.com'
}

// 打开微博
const openWeibo = () => {
  ElMessage.info('正在跳转到官方微博...')
  // 实际打开微博逻辑
  // window.open('https://weibo.com/校园配送骑手')
}

// 复制QQ群号
const copyQQGroup = () => {
  const qqNumber = '123456789'

  // 创建临时文本区域
  const textArea = document.createElement('textarea')
  textArea.value = qqNumber
  textArea.style.position = 'fixed'
  textArea.style.left = '-999999px'
  textArea.style.top = '-999999px'
  document.body.appendChild(textArea)
  textArea.focus()
  textArea.select()

  try {
    const successful = document.execCommand('copy')
    if (successful) {
      ElMessage.success('QQ群号已复制')
    } else {
      throw new Error('复制失败')
    }
  } catch (err) {
    ElMessage.error('复制失败，请手动复制：' + qqNumber)
  }

  document.body.removeChild(textArea)
}

// 跳转到意见反馈
const goToFeedback = () => {
  router.push('/rider/profile/feedback')
}

// 查看问题详情
const viewIssue = (issue) => {
  router.push(issue.route)
}

// 关闭微信对话框
const closeWechatDialog = () => {
  showWechatQR.value = false
}

// 保存二维码
const saveQRCode = () => {
  ElMessage.info('正在保存二维码...')
  // 实际保存二维码逻辑
  setTimeout(() => {
    ElMessage.success('二维码已保存到相册')
  }, 1000)
}

// 加载联系信息
const loadContactInfo = async () => {
  try {
    // 模拟API调用
    // const response = await riderApi.getContactInfo()
    console.log('使用默认联系信息')
  } catch (error) {
    console.error('加载联系信息失败:', error)
    ElMessage.warning('加载信息失败，显示默认内容')
  }
}

onMounted(() => {
  loadContactInfo()
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

/* 微信图标 */
.css-icon.wechat::before {
  content: '💬';
  font-size: 20px;
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

.css-icon.email::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 4px;
  height: 4px;
  background: currentColor;
  border-radius: 50%;
}

/* 微博图标 */
.css-icon.weibo::before {
  content: '📢';
  font-size: 20px;
}

/* QQ图标 */
.css-icon.qq::before {
  content: '🐧';
  font-size: 20px;
}

/* 反馈图标 */
.css-icon.feedback::before {
  content: '💭';
  font-size: 20px;
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

/* 时钟图标 */
.css-icon.clock::before {
  content: '🕒';
  font-size: 20px;
}

/* 警告图标 */
.css-icon.warning::before {
  content: '⚠️';
  font-size: 20px;
}

/* 警察图标 */
.css-icon.police::before {
  content: '👮';
  font-size: 20px;
}

/* 救护车图标 */
.css-icon.ambulance::before {
  content: '🚑';
  font-size: 20px;
}

/* 问题图标 */
.css-icon.issue-money::before {
  content: '💰';
  font-size: 20px;
}

.css-icon.issue-timeout::before {
  content: '⏰';
  font-size: 20px;
}

.css-icon.issue-profile::before {
  content: '👤';
  font-size: 20px;
}

.css-icon.issue-route::before {
  content: '🗺️';
  font-size: 20px;
}

.css-icon.issue-income::before {
  content: '📊';
  font-size: 20px;
}

.contact-service {
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

.back-btn, .placeholder {
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

.back-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 快捷联系 */
.quick-contact {
  margin: 70px 15px 15px;
}

.contact-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.contact-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.contact-card {
  display: flex;
  align-items: center;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s ease;
}

.contact-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.card-icon {
  width: 50px;
  height: 50px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.card-icon.online {
  background: #f0f9ff;
}

.card-icon.wechat {
  background: #f0fdf4;
}

.card-icon .css-icon {
  font-size: 24px;
  color: #FFD700;
}

.card-content {
  flex: 1;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.card-desc {
  font-size: 14px;
  color: #666;
  margin-bottom: 4px;
}

.card-time {
  font-size: 12px;
  color: #999;
}

.card-status-info {
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
  font-size: 12px;
  color: #67C23A;
}

.card-status .status-text {
  color: #FFD700;
  font-weight: 500;
}

/* 常见问题 */
.common-issues {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.issues-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.issues-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.issue-item {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.3s ease;
}

.issue-item:last-child {
  border-bottom: none;
}

.issue-item:hover {
  background: #f8f9fa;
  margin: 0 -20px;
  padding-left: 20px;
  padding-right: 20px;
}

.issue-icon {
  width: 40px;
  height: 40px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.issue-icon .css-icon {
  font-size: 20px;
}

.issue-content {
  flex: 1;
}

.issue-title {
  font-size: 14px;
  color: #333;
  margin-bottom: 4px;
}

.issue-count {
  font-size: 12px;
  color: #999;
}

.issue-arrow {
  display: flex;
  align-items: center;
  color: #ccc;
}

/* 其他联系方式 */
.other-contacts {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.contacts-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.contacts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
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
  margin-right: 12px;
}

.contact-icon .css-icon {
  font-size: 20px;
  color: #FFD700;
}

.contact-info {
  flex: 1;
}

.contact-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.contact-detail {
  font-size: 12px;
  color: #666;
  margin-bottom: 2px;
}

.contact-desc {
  font-size: 11px;
  color: #999;
}

/* 客服工作时间 */
.service-hours {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.hours-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.hours-title .css-icon {
  font-size: 20px;
  color: #FFD700;
}

.hours-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hours-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.hours-item:last-child {
  border-bottom: none;
}

.hours-label {
  font-size: 14px;
  color: #333;
}

.hours-time {
  font-size: 14px;
  color: #666;
}

/* 紧急联系 */
.emergency-contact {
  margin: 15px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.emergency-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #dc2626;
  margin-bottom: 15px;
}

.emergency-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.emergency-desc {
  font-size: 14px;
  color: #7f1d1d;
  line-height: 1.5;
}

.emergency-phone {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: white;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.emergency-phone:hover {
  background: #fef3c7;
}

.emergency-phone .css-icon {
  font-size: 20px;
}

.emergency-phone span {
  font-size: 16px;
  font-weight: 500;
  color: #dc2626;
}

.emergency-tip {
  font-size: 12px;
  color: #991b1b;
  text-align: center;
  padding: 8px;
  background: rgba(220, 38, 38, 0.1);
  border-radius: 6px;
}

/* 微信二维码弹窗 */
.wechat-qr-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}

.qr-code {
  margin-bottom: 20px;
}

.qr-code img {
  width: 200px;
  height: 200px;
  border-radius: 8px;
}

.qr-info {
  text-align: center;
}

.qr-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 8px;
}

.qr-desc {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.qr-tip {
  font-size: 12px;
  color: #999;
  margin-bottom: 15px;
}

.save-qr-btn {
  width: 200px;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .quick-contact,
  .common-issues,
  .other-contacts,
  .service-hours,
  .emergency-contact {
    margin: 10px;
    padding: 15px;
  }

  .contact-card {
    padding: 15px;
  }

  .card-icon {
    width: 40px;
    height: 40px;
    margin-right: 12px;
  }

  .card-icon .css-icon {
    font-size: 20px;
  }

  .contacts-grid {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .contact-item {
    padding: 12px;
  }

  .contact-icon {
    width: 36px;
    height: 36px;
  }

  .contact-icon .css-icon {
    font-size: 18px;
  }
}
</style>