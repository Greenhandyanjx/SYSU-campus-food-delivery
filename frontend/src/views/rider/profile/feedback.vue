<template>
  <div class="feedback-page">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">意见反馈</h1>
      <div class="placeholder"></div>
    </div>

    <!-- 反馈类型选择 -->
    <div class="feedback-type-section">
      <div class="type-title">反馈类型</div>
      <div class="type-grid">
        <div
          v-for="type in feedbackTypes"
          :key="type.id"
          class="type-item"
          :class="{ active: selectedType === type.id }"
          @click="selectType(type.id)"
        >
          <div class="type-icon">
            <i :class="type.iconClass"></i>
          </div>
          <div class="type-name">{{ type.name }}</div>
        </div>
      </div>
    </div>

    <!-- 反馈表单 -->
    <div class="feedback-form-section">
      <div class="form-title">反馈内容</div>

      <!-- 问题标题 -->
      <div class="form-group">
        <div class="form-label">问题标题 *</div>
        <el-input
          v-model="feedbackForm.title"
          placeholder="请简要描述您的问题或建议"
          maxlength="50"
          show-word-limit
        />
      </div>

      <!-- 反馈内容 -->
      <div class="form-group">
        <div class="form-label">详细描述 *</div>
        <el-input
          v-model="feedbackForm.content"
          type="textarea"
          :rows="6"
          placeholder="请详细描述您遇到的问题或建议，我们会认真对待每一条反馈"
          maxlength="500"
          show-word-limit
        />
      </div>

      <!-- 联系方式 -->
      <div class="form-group">
        <div class="form-label">联系方式</div>
        <el-input
          v-model="feedbackForm.contact"
          placeholder="手机号或邮箱，方便我们联系您（选填）"
        />
      </div>

      <!-- 图片上传 -->
      <div class="form-group">
        <div class="form-label">相关图片</div>
        <div class="image-upload">
          <div v-for="(image, index) in feedbackForm.images" :key="index" class="image-item">
            <img :src="image" :alt="`上传图片${index + 1}`" />
            <div class="image-delete" @click="removeImage(index)">
              <i class="css-icon delete"></i>
            </div>
          </div>
          <div v-if="feedbackForm.images.length < 4" class="image-upload-btn" @click="uploadImage">
            <i class="css-icon camera"></i>
            <div class="upload-text">添加图片</div>
          </div>
        </div>
        <div class="upload-tip">最多上传4张图片，每张不超过5MB</div>
      </div>

      <!-- 位置信息 -->
      <div class="form-group">
        <div class="form-label">相关位置</div>
        <div class="location-info">
          <div v-if="feedbackForm.location.address" class="location-display">
            <i class="css-icon location"></i>
            <span>{{ feedbackForm.location.address }}</span>
            <div class="location-clear" @click="clearLocation">
              <i class="css-icon clear"></i>
            </div>
          </div>
          <div v-else class="location-add" @click="selectLocation">
            <i class="css-icon location-add"></i>
            <span>添加位置信息</span>
          </div>
        </div>
      </div>

      <!-- 提交时间 -->
      <div class="form-group">
        <div class="form-label">问题发生时间</div>
        <el-date-picker
          v-model="feedbackForm.issueTime"
          type="datetime"
          placeholder="选择问题发生时间（选填）"
          format="YYYY-MM-DD HH:mm"
          value-format="YYYY-MM-DD HH:mm"
        />
      </div>
    </div>

    <!-- 历史反馈 -->
    <div class="history-section">
      <div class="history-title">
        <span>历史反馈</span>
        <el-link type="primary" @click="viewAllHistory">查看全部</el-link>
      </div>

      <div class="history-list">
        <div v-for="item in recentFeedback" :key="item.id" class="history-item" @click="viewFeedbackDetail(item)">
          <div class="history-status">
            <div class="status-dot" :class="item.status"></div>
          </div>
          <div class="history-content">
            <div class="history-type">{{ getFeedbackTypeName(item.type) }}</div>
            <div class="history-title-text">{{ item.title }}</div>
            <div class="history-time">{{ formatTime(item.submitTime) }}</div>
          </div>
          <div class="history-arrow">
            <i class="css-icon arrow"></i>
          </div>
        </div>
      </div>

      <div v-if="recentFeedback.length === 0" class="empty-history">
        <div class="empty-icon">
          <i class="css-icon empty"></i>
        </div>
        <div class="empty-text">暂无反馈记录</div>
      </div>
    </div>

    <!-- 热门问题 -->
    <div class="hot-issues-section">
      <div class="hot-issues-title">热门问题</div>
      <div class="hot-issues-list">
        <div v-for="issue in hotIssues" :key="issue.id" class="hot-issue-item" @click="viewHotIssue(issue)">
          <div class="hot-issue-icon">🔥</div>
          <div class="hot-issue-content">
            <div class="hot-issue-title">{{ issue.title }}</div>
            <div class="hot-issue-count">{{ issue.feedbackCount }}人反馈</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 提交按钮 -->
    <div class="submit-section">
      <el-button type="primary" @click="submitFeedback" class="submit-btn" :loading="submitting">
        提交反馈
      </el-button>
    </div>

    <!-- 图片上传隐藏input -->
    <input
      ref="imageInput"
      type="file"
      accept="image/*"
      style="display: none"
      @change="handleImageSelect"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()
const imageInput = ref(null)

// 反馈类型
const feedbackTypes = [
  {
    id: 'bug',
    name: '程序异常',
    iconClass: 'css-icon bug'
  },
  {
    id: 'feature',
    name: '功能建议',
    iconClass: 'css-icon feature'
  },
  {
    id: 'ui',
    name: '界面问题',
    iconClass: 'css-icon ui'
  },
  {
    id: 'order',
    name: '订单问题',
    iconClass: 'css-icon order'
  },
  {
    id: 'payment',
    name: '支付问题',
    iconClass: 'css-icon payment'
  },
  {
    id: 'other',
    name: '其他问题',
    iconClass: 'css-icon other'
  }
]

// 反馈表单
const feedbackForm = reactive({
  type: '',
  title: '',
  content: '',
  contact: '',
  images: [],
  location: {
    address: '',
    latitude: null,
    longitude: null
  },
  issueTime: null
})

// 选中的反馈类型
const selectedType = ref('')

// 提交状态
const submitting = ref(false)

// 历史反馈
const recentFeedback = ref([
  {
    id: 1,
    type: 'bug',
    title: 'App经常闪退',
    status: 'resolved',
    submitTime: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000)
  },
  {
    id: 2,
    type: 'feature',
    title: '希望能添加自动接单功能',
    status: 'processing',
    submitTime: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000)
  }
])

// 热门问题
const hotIssues = ref([
  {
    id: 1,
    title: '配送路线规划不准确',
    feedbackCount: 156
  },
  {
    id: 2,
    title: '订单超时罚款问题',
    feedbackCount: 128
  },
  {
    id: 3,
    title: '收入结算显示异常',
    feedbackCount: 95
  }
])

// 选择反馈类型
const selectType = (typeId) => {
  selectedType.value = typeId
  feedbackForm.type = typeId
}

// 上传图片
const uploadImage = () => {
  imageInput.value?.click()
}

// 处理图片选择
const handleImageSelect = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // 检查文件大小
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过5MB')
    return
  }

  // 检查文件类型
  if (!file.type.startsWith('image/')) {
    ElMessage.error('请选择图片文件')
    return
  }

  // 读取文件并转换为URL
  const reader = new FileReader()
  reader.onload = (e) => {
    feedbackForm.images.push(e.target.result)
  }
  reader.readAsDataURL(file)

  // 清空input，允许重复选择同一文件
  event.target.value = ''
}

// 移除图片
const removeImage = (index) => {
  feedbackForm.images.splice(index, 1)
}

// 选择位置
const selectLocation = () => {
  ElMessage.info('位置选择功能开发中...')
  // 模拟选择位置
  feedbackForm.location = {
    address: '珠海市香洲区唐家湾大学路1号',
    latitude: 22.3080,
    longitude: 113.5400
  }
}

// 清除位置
const clearLocation = () => {
  feedbackForm.location = {
    address: '',
    latitude: null,
    longitude: null
  }
}

// 获取反馈类型名称
const getFeedbackTypeName = (typeId) => {
  const type = feedbackTypes.find(t => t.id === typeId)
  return type ? type.name : '其他'
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

// 提交反馈
const submitFeedback = async () => {
  try {
    // 验证表单
    if (!feedbackForm.type) {
      ElMessage.error('请选择反馈类型')
      return
    }

    if (!feedbackForm.title.trim()) {
      ElMessage.error('请输入问题标题')
      return
    }

    if (!feedbackForm.content.trim()) {
      ElMessage.error('请输入详细描述')
      return
    }

    submitting.value = true

    // 构建提交数据
    const submitData = {
      type: feedbackForm.type,
      title: feedbackForm.title,
      content: feedbackForm.content,
      contact: feedbackForm.contact,
      images: feedbackForm.images,
      location: feedbackForm.location,
      issueTime: feedbackForm.issueTime,
      submitTime: new Date().toISOString()
    }

    // 模拟API调用
    // const response = await riderApi.submitFeedback(submitData)

    // 模拟提交成功
    await new Promise(resolve => setTimeout(resolve, 1500))

    ElMessage.success('反馈提交成功，我们会尽快处理')

    // 重置表单
    resetForm()

    // 返回上一页
    router.go(-1)
  } catch (error) {
    console.error('提交反馈失败:', error)
    ElMessage.error('提交失败，请重试')
  } finally {
    submitting.value = false
  }
}

// 重置表单
const resetForm = () => {
  feedbackForm.type = ''
  feedbackForm.title = ''
  feedbackForm.content = ''
  feedbackForm.contact = ''
  feedbackForm.images = []
  feedbackForm.location = {
    address: '',
    latitude: null,
    longitude: null
  }
  feedbackForm.issueTime = null
  selectedType.value = ''
}

// 查看反馈详情
const viewFeedbackDetail = (feedback) => {
  router.push(`/rider/feedback/detail/${feedback.id}`)
}

// 查看全部历史
const viewAllHistory = () => {
  router.push('/rider/feedback/history')
}

// 查看热门问题
const viewHotIssue = (issue) => {
  router.push(`/rider/feedback/hot-issue/${issue.id}`)
}

// 加载历史反馈
const loadRecentFeedback = async () => {
  try {
    // 模拟API调用
    // const response = await riderApi.getRecentFeedback()
    console.log('使用默认历史反馈数据')
  } catch (error) {
    console.error('加载历史反馈失败:', error)
  }
}

onMounted(() => {
  loadRecentFeedback()
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

/* Bug图标 */
.css-icon.bug::before {
  content: '🐛';
  font-size: 24px;
}

/* 功能图标 */
.css-icon.feature::before {
  content: '💡';
  font-size: 24px;
}

/* UI图标 */
.css-icon.ui::before {
  content: '🎨';
  font-size: 24px;
}

/* 订单图标 */
.css-icon.order::before {
  content: '📦';
  font-size: 24px;
}

/* 支付图标 */
.css-icon.payment::before {
  content: '💳';
  font-size: 24px;
}

/* 其他图标 */
.css-icon.other::before {
  content: '❓';
  font-size: 24px;
}

/* 删除图标 */
.css-icon.delete::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(45deg);
  width: 12px;
  height: 2px;
  background: white;
}

.css-icon.delete::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-45deg);
  width: 12px;
  height: 2px;
  background: white;
}

/* 相机图标 */
.css-icon.camera::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 24px;
  height: 18px;
  border: 2px solid #999;
  border-radius: 4px;
}

.css-icon.camera::after {
  content: '';
  position: absolute;
  bottom: -6px;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 8px;
  background: #999;
  border-radius: 0 0 4px 4px;
}

/* 位置图标 */
.css-icon.location::before {
  content: '📍';
  font-size: 16px;
}

.css-icon.location-add::before {
  content: '➕';
  font-size: 16px;
  color: #999;
}

/* 清除图标 */
.css-icon.clear::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(45deg);
  width: 8px;
  height: 8px;
  border-left: 1px solid #999;
  border-bottom: 1px solid #999;
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
  border-right: 2px solid #ccc;
  border-top: 2px solid #ccc;
}

/* 空状态图标 */
.css-icon.empty::before {
  content: '📭';
  font-size: 48px;
}

.feedback-page {
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

/* 反馈类型选择 */
.feedback-type-section {
  margin: 70px 15px 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.type-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.type-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 15px;
}

.type-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 15px 10px;
  border: 2px solid #f0f0f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.type-item:hover {
  border-color: #FFD700;
}

.type-item.active {
  border-color: #FFD700;
  background: #fffbf0;
}

.type-icon {
  width: 50px;
  height: 50px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}

.type-icon .css-icon {
  font-size: 24px;
}

.type-name {
  font-size: 12px;
  color: #333;
  text-align: center;
}

/* 反馈表单 */
.feedback-form-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-label {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 8px;
}

/* 图片上传 */
.image-upload {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.image-item {
  position: relative;
  width: 80px;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
}

.image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-delete {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 20px;
  height: 20px;
  background: #F56C6C;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.image-delete .css-icon {
  font-size: 12px;
}

.image-upload-btn {
  width: 80px;
  height: 80px;
  border: 2px dashed #ddd;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.image-upload-btn:hover {
  border-color: #FFD700;
}

.image-upload-btn .css-icon {
  margin-bottom: 4px;
}

.upload-text {
  font-size: 10px;
  color: #999;
}

.upload-tip {
  font-size: 12px;
  color: #999;
  margin-top: 8px;
}

/* 位置信息 */
.location-info {
  margin-top: 8px;
}

.location-display,
.location-add {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.location-display:hover,
.location-add:hover {
  background: #e9ecef;
}

.location-display {
  justify-content: space-between;
}

.location-add span {
  color: #999;
}

.location-clear {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

/* 历史反馈 */
.history-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.history-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 15px;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: background 0.3s ease;
}

.history-item:last-child {
  border-bottom: none;
}

.history-item:hover {
  background: #f8f9fa;
  margin: 0 -20px;
  padding-left: 20px;
  padding-right: 20px;
}

.history-status {
  margin-right: 12px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.resolved {
  background: #67C23A;
}

.status-dot.processing {
  background: #E6A23C;
}

.status-dot.pending {
  background: #909399;
}

.history-content {
  flex: 1;
}

.history-type {
  font-size: 12px;
  color: #999;
  margin-bottom: 4px;
}

.history-title-text {
  font-size: 14px;
  color: #333;
  margin-bottom: 4px;
}

.history-time {
  font-size: 11px;
  color: #999;
}

.history-arrow {
  display: flex;
  align-items: center;
  color: #ccc;
}

.empty-history {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 0;
}

.empty-icon {
  margin-bottom: 10px;
}

.empty-text {
  font-size: 14px;
  color: #999;
}

/* 热门问题 */
.hot-issues-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.hot-issues-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 15px;
}

.hot-issues-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hot-issue-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.hot-issue-item:hover {
  background: #e9ecef;
}

.hot-issue-icon {
  margin-right: 12px;
  font-size: 20px;
}

.hot-issue-content {
  flex: 1;
}

.hot-issue-title {
  font-size: 14px;
  color: #333;
  margin-bottom: 4px;
}

.hot-issue-count {
  font-size: 11px;
  color: #999;
}

/* 提交按钮 */
.submit-section {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  padding: 15px;
  border-top: 1px solid #f0f0f0;
  z-index: 100;
}

.submit-btn {
  width: 100%;
  height: 50px;
  background: #FFD700;
  color: #333;
  border: none;
  border-radius: 25px;
  font-size: 16px;
  font-weight: 500;
}

/* 自定义组件样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 12px 15px;
}

:deep(.el-textarea__inner) {
  border-radius: 8px;
  padding: 12px 15px;
}

:deep(.el-date-editor) {
  width: 100%;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .feedback-type-section,
  .feedback-form-section,
  .history-section,
  .hot-issues-section {
    margin: 10px;
    padding: 15px;
  }

  .type-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .type-item {
    padding: 12px 8px;
  }

  .type-icon {
    width: 40px;
    height: 40px;
  }

  .type-icon .css-icon {
    font-size: 20px;
  }

  .image-upload {
    gap: 8px;
  }

  .image-item,
  .image-upload-btn {
    width: 70px;
    height: 70px;
  }
}
</style>