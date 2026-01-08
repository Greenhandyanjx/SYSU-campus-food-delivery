<template>
  <!-- 聊天容器 -->
  <div class="rider-chat-container" :class="{ 'chat-expanded': showChat }">
    <!-- 收起状态的浮动按钮 -->
    <div
      class="chat-float-button"
      @click="toggleChat"
      v-if="!showChat"
    >
      <el-badge :value="totalUnread" :hidden="totalUnread === 0">
        <i class="iconfont icon-message"></i>
      </el-badge>
    </div>

    <!-- 展开状态的聊天面板 -->
    <div class="chat-panel" v-if="showChat">
      <!-- 聊天列表 -->
      <div class="chat-sidebar" v-if="!currentChat">
        <div class="chat-header">
          <h3>消息列表</h3>
          <div class="header-actions">
            <el-badge :value="totalUnread" :hidden="totalUnread === 0" class="unread-badge">
              <i class="iconfont icon-message"></i>
            </el-badge>
            <div class="close-btn" @click="showChat = false">
              <i class="iconfont icon-close"></i>
            </div>
          </div>
        </div>

        <div class="chat-list">
          <div
            v-for="session in chatSessions"
            :key="session.merchant_id"
            class="chat-item"
            @click="openChat(session)"
          >
            <div class="chat-avatar">
              <img :src="getMerchantAvatar(session.merchant_id)" :alt="session.merchant_name">
            </div>
            <div class="chat-info">
              <div class="chat-name">{{ session.merchant_name || `商家 ${session.merchant_id}` }}</div>
              <div class="chat-last-message">{{ session.last_message }}</div>
            </div>
            <div class="chat-meta">
              <div class="chat-time">{{ formatTime(session.last_at) }}</div>
              <el-badge
                :value="session.unread_count"
                :hidden="session.unread_count === 0"
                class="unread-count"
              />
            </div>
          </div>
        </div>

        <div class="no-chats" v-if="chatSessions.length === 0">
          <i class="iconfont icon-empty"></i>
          <p>暂无聊天记录</p>
          <p class="hint">从订单页面点击"联系"按钮开始聊天</p>
        </div>
      </div>

      <!-- 聊天窗口 -->
      <div class="chat-window" v-if="currentChat">
        <div class="chat-header">
          <div class="chat-back" @click="closeCurrentChat">
            <i class="iconfont icon-arrow-left"></i>
          </div>
          <div class="chat-title">{{ currentChatTitle }}</div>
          <div class="chat-actions">
            <el-button
              size="small"
              type="text"
              @click="markAsRead"
              :disabled="currentChat.unread_count === 0"
            >
              标记已读
            </el-button>
            <div class="close-btn" @click="showChat = false">
              <i class="iconfont icon-close"></i>
            </div>
          </div>
        </div>

        <div class="chat-messages" ref="messagesContainer">
          <div
            v-for="message in messages"
            :key="message.id"
            class="message-item"
            :class="{ 'message-self': isFromSelf(message) }"
          >
            <div class="message-avatar">
              <img :src="getUserAvatar(message.from_base_id)" :alt="isFromSelf(message) ? '我' : '对方'">
            </div>
            <div class="message-content">
              <div class="message-bubble">{{ message.content }}</div>
              <div class="message-time">{{ formatMessageTime(message.created_at) }}</div>
            </div>
          </div>
        </div>

        <div class="chat-input">
          <el-input
            v-model="messageInput"
            type="textarea"
            :rows="2"
            placeholder="输入消息... (Ctrl+Enter 发送)"
            @keydown.enter.ctrl="sendMessage"
          />
          <el-button
            type="primary"
            @click="sendMessage"
            :disabled="!messageInput.trim()"
          >
            发送
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { riderApi, type ChatSession, type ChatMessage, type RiderMe } from '@/api/rider'
import chatClient from '@/utils/chatClient'

// 状态管理
const showChat = ref(false)
const chatSessions = ref<ChatSession[]>([])
const messages = ref<ChatMessage[]>([])
const currentChat = ref<ChatSession | null>(null)
const messageInput = ref('')
const messagesContainer = ref<HTMLElement>()
const currentUserBaseId = ref<number>(0)
const currentRider = ref<RiderMe | null>(null)
const currentChatTitle = ref('')
const currentChatType = ref<'merchant' | 'user'>('merchant')
const currentChatUserId = ref<number>(0) // 用于联系用户时的用户ID

// 计算总未读数
const totalUnread = computed(() => {
  return chatSessions.value.reduce((sum, session) => sum + session.unread_count, 0)
})

// 获取商家头像
const getMerchantAvatar = (merchantId: number) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${merchantId}`
}

// 获取用户头像
const getUserAvatar = (userBaseId: number) => {
  return `https://api.dicebear.com/7.x/avataaars/svg?seed=${userBaseId}`
}

// 格式化时间
const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return date.toLocaleDateString()
}

// 格式化消息时间
const formatMessageTime = (timeStr: string) => {
  const date = new Date(timeStr)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

// 判断是否是自己发送的消息
const isFromSelf = (message: ChatMessage) => {
  return message.from_base_id === currentUserBaseId.value
}

// 打开聊天
const openChat = async (session: ChatSession) => {
  currentChat.value = session
  currentChatTitle.value = session.merchant_name || `商家 ${session.merchant_id}`
  await loadChatHistory(session.merchant_id)
  await nextTick()
  scrollToBottom()
}

// 从订单直接打开聊天
const openChatFromOrder = (data: { type: 'user' | 'merchant'; id: number; name: string }) => {
  // 先显示聊天面板
  showChat.value = true
  currentChatType.value = data.type

  if (data.type === 'merchant') {
    // 创建或找到与商家的会话
    let session = chatSessions.value.find(s => s.merchant_id === data.id)
    if (!session) {
      // 如果没有现有会话，创建一个临时会话
      session = {
        merchant_id: data.id,
        merchant_name: data.name,
        last_message: '',
        last_at: new Date().toISOString(),
        unread_count: 0
      }
      chatSessions.value.unshift(session)
    }
    openChat(session)
  } else if (data.type === 'user') {
    // 创建与用户的临时会话
    currentChatUserId.value = data.id
    const tempSession = {
      merchant_id: 0, // 与用户聊天时merchant_id为0
      merchant_name: data.name,
      last_message: '',
      last_at: new Date().toISOString(),
      unread_count: 0,
      user_base_id: data.id
    }
    openChatWithUser(tempSession, data.name)
  }
}

// 打开与用户的聊天
const openChatWithUser = (session: any, userName: string) => {
  currentChat.value = session
  currentChatTitle.value = userName
  // 加载与该用户和商家相关的聊天历史
  loadUserChatHistory()
  nextTick(() => scrollToBottom())
}

// 加载用户聊天历史（骑手与顾客的聊天需要通过某个商家进行）
const loadUserChatHistory = async () => {
  // 这里需要一个商家ID作为中介，暂时使用第一个可用的商家
  // 在实际项目中，可能需要骑手选择通过哪个商家与顾客联系
  try {
    const merchants = chatSessions.value.filter(s => s.merchant_id > 0)
    if (merchants.length > 0) {
      const response = await riderApi.getChatHistory(merchants[0].merchant_id, currentChatUserId.value)
      messages.value = response.data.data || []
    }
  } catch (error) {
    console.error('加载用户聊天历史失败:', error)
  }
}

// 关闭当前聊天
const closeCurrentChat = () => {
  currentChat.value = null
  currentChatTitle.value = ''
}

// 切换聊天显示状态
const toggleChat = () => {
  showChat.value = !showChat.value
  if (showChat.value) {
    currentChat.value = null
    currentChatTitle.value = ''
  }
}

// 加载聊天历史
const loadChatHistory = async (merchantId: number) => {
  try {
    const response = await riderApi.getChatHistory(merchantId, currentUserBaseId.value)
    messages.value = response.data.data || []
  } catch (error) {
    console.error('加载聊天历史失败:', error)
  }
}

// 标记为已读
const markAsRead = async () => {
  if (!currentChat.value || currentChat.value.unread_count === 0) return

  try {
    await riderApi.markChatAsRead(currentChat.value.merchant_id)
    currentChat.value.unread_count = 0
    // 更新会话列表中的未读数
    const session = chatSessions.value.find(s => s.merchant_id === currentChat.value?.merchant_id)
    if (session) {
      session.unread_count = 0
    }
  } catch (error) {
    console.error('标记已读失败:', error)
  }
}

// 发送消息
const sendMessage = async () => {
  if (!messageInput.value.trim() || !currentChat.value) return

  const content = messageInput.value.trim()
  messageInput.value = ''

  // 骑手以用户身份发送消息给商家
  // 这样后端会正确处理消息转发
  const payload = {
    merchant_id: currentChat.value.merchant_id,
    user_base_id: currentUserBaseId.value, // 骑手作为"用户"发送
    content: `🏍️ 骑手: ${content}`, // 前缀标识骑手身份
    type: 'text'
  }

  console.log('骑手发送消息payload:', payload)

  // 通过WebSocket发送消息
  const success = chatClient.send(payload)

  if (success) {
    // 临时添加消息到本地（乐观更新）
    const tempMessage: ChatMessage = {
      id: Date.now(),
      from_base_id: currentUserBaseId.value,
      merchant_id: currentChat.value.merchant_id,
      user_base_id: currentUserBaseId.value,
      content,
      type: 'text',
      status: 'sent',
      created_at: new Date().toISOString()
    }
    messages.value.push(tempMessage)
    await nextTick()
    scrollToBottom()
  } else {
    // 发送失败，恢复输入内容
    messageInput.value = content
    console.error('消息发送失败')
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}


// 加载聊天会话列表
const loadChatSessions = async () => {
  try {
    const response = await riderApi.getChatSessions()
    chatSessions.value = response.data.data || []
  } catch (error) {
    console.error('加载聊天会话失败:', error)
  }
}

// WebSocket消息处理
const handleWebSocketMessage = (data: any) => {
  console.log('收到WebSocket消息:', data)

  // 如果是当前聊天的消息，添加到消息列表
  if (currentChat.value && data.merchant_id === currentChat.value.merchant_id) {
    const message: ChatMessage = {
      id: data.id || Date.now(),
      from_base_id: data.from_base_id,
      merchant_id: data.merchant_id,
      user_base_id: data.user_base_id,
      content: data.content,
      type: data.type || 'text',
      status: data.status || 'delivered',
      created_at: data.created_at || new Date().toISOString()
    }
    messages.value.push(message)
    nextTick(() => scrollToBottom())
  }

  // 更新会话列表
  loadChatSessions()
}

// 获取当前骑手信息
const getCurrentRider = async () => {
  try {
    const response = await riderApi.getMe()
    currentRider.value = response.data.data
    // 假设骑手的base_user_id就是骑手ID，实际可能需要根据API调整
    currentUserBaseId.value = currentRider.value?.id || 0
  } catch (error) {
    console.error('获取骑手信息失败:', error)
  }
}

// 处理来自订单组件的打开聊天请求
const handleOpenChatEvent = (event: CustomEvent) => {
  openChatFromOrder(event.detail)
}

onMounted(async () => {
  await getCurrentRider()
  await loadChatSessions()

  // 监听WebSocket消息
  chatClient.onMessage(handleWebSocketMessage)

  // 确保WebSocket连接
  if (!chatClient || !chatClient.connect) {
    chatClient.connect()
  }

  // 监听来自订单组件的打开聊天事件
  window.addEventListener('rider:openChat', handleOpenChatEvent as EventListener)
})

onUnmounted(() => {
  // 移除WebSocket消息监听
  chatClient.offMessage(handleWebSocketMessage)
  // 移除打开聊天事件监听
  window.removeEventListener('rider:openChat', handleOpenChatEvent as EventListener)
})
</script>

<style scoped lang="scss">
.rider-chat-container {
  position: fixed;
  bottom: 80px;
  right: 20px;
  z-index: 1000;
  transition: all 0.3s ease;
}

.chat-panel {
  width: 360px;
  height: 500px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.chat-sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.chat-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #409eff 0%, #67c23a 100%);
  color: white;

  h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .unread-badge :deep(.el-badge__content) {
    background-color: #ff4757;
  }

  .close-btn {
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: background-color 0.2s;

    &:hover {
      background-color: rgba(255, 255, 255, 0.2);
    }
  }
}

.chat-list {
  flex: 1;
  overflow-y: auto;
}

.chat-item {
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 12px;
  transition: background-color 0.2s;

  &:hover {
    background-color: #f8f9fa;
  }
}

.chat-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.chat-info {
  flex: 1;
  min-width: 0;
}

.chat-name {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
}

.chat-last-message {
  font-size: 12px;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.chat-time {
  font-size: 11px;
  color: #999;
}

.unread-count :deep(.el-badge__content) {
  background-color: #ff4757;
  font-size: 10px;
  padding: 0 4px;
  height: 16px;
  line-height: 16px;
}

.no-chats {
  padding: 40px 20px;
  text-align: center;
  color: #999;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;

  i {
    font-size: 48px;
    opacity: 0.5;
  }

  p {
    margin: 0;
    font-size: 14px;
  }

  .hint {
    font-size: 12px;
    color: #ccc;
    line-height: 1.4;
  }
}

.chat-window {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.chat-window .chat-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(135deg, #409eff 0%, #67c23a 100%);
  color: white;
}

.chat-back {
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: background-color 0.2s;

  &:hover {
    background-color: rgba(255, 255, 255, 0.2);
  }
}

.chat-title {
  flex: 1;
  font-weight: 600;
  font-size: 16px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message-item {
  display: flex;
  gap: 8px;
  align-items: flex-start;

  &.message-self {
    flex-direction: row-reverse;

    .message-bubble {
      background-color: #409eff;
      color: white;
    }
  }
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}

.message-content {
  max-width: 70%;
}

.message-bubble {
  background-color: #f0f0f0;
  padding: 8px 12px;
  border-radius: 12px;
  word-break: break-word;
}

.message-time {
  font-size: 11px;
  color: #999;
  margin-top: 4px;
  text-align: right;
}

.chat-input {
  padding: 16px;
  border-top: 1px solid #eee;
  display: flex;
  gap: 12px;
  align-items: flex-end;

  .el-textarea {
    flex: 1;
  }
}

.chat-float-button {
  position: fixed;
  bottom: 100px;
  right: 20px;
  width: 56px;
  height: 56px;
  background-color: #409eff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.4);
  transition: all 0.3s ease;
  z-index: 999;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(64, 158, 255, 0.5);
  }

  i {
    color: white;
    font-size: 24px;
  }

  :deep(.el-badge__content) {
    background-color: #ff4757;
  }
}

.iconfont {
  font-family: "iconfont" !important;
  font-style: normal;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.icon-message:before { content: "💬"; }
.icon-arrow-left:before { content: "◀"; }
.icon-close:before { content: "✕"; }
.icon-empty:before { content: "📭"; }
</style>