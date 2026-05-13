<template>
  <!--
    AgentPanel.vue — 外卖助手 AI 聊天面板（支持浮动 / 全屏两种模式）
    ===================================================================
    通过 `fullPage` prop 控制：
    - fullPage=false（默认）：浮动 FAB + 侧滑面板（类似美团/饿了么客服入口）
    - fullPage=true  ：全屏模式，填满父容器，无 FAB，无固定定位

    所有模式使用 useAgentChat 共享同一个聊天状态（单例模式）。
  -->

  <!-- ── 浮动模式：FAB 按钮 ── -->
  <Transition name="fab-fade">
    <button
      v-if="!fullPage && !panelVisible"
      class="agent-fab"
      @click="openPanel"
      :title="'打开外卖助手'"
      aria-label="打开外卖助手"
    >
      <svg class="fab-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2Z" fill="currentColor" opacity="0.9"/>
        <circle cx="9" cy="10" r="1.5" fill="white"/>
        <circle cx="15" cy="10" r="1.5" fill="white"/>
        <path d="M8 15C8 15 9.5 17 12 17C14.5 17 16 15 16 15" stroke="white" stroke-width="1.5" stroke-linecap="round" fill="none"/>
      </svg>
    </button>
  </Transition>

  <!-- ── 浮动模式：遮罩 + 面板 | 全屏模式：直接渲染 ── -->
  <div
    v-if="fullPage || panelVisible"
    class="agent-panel"
    :class="{
      'is-floating': !fullPage,
      'is-mobile': !fullPage && isMobile,
      'is-fullpage': fullPage,
    }"
  >
    <!-- 浮动模式的遮罩层（移动端点击外部关闭） -->
    <div
      v-if="!fullPage && isMobile"
      class="agent-overlay"
      @click="closePanel"
    ></div>

    <!-- 面板主体 -->
    <div class="agent-panel-inner">
      <!-- ── 浮动模式头部 ── -->
      <header v-if="!fullPage" class="agent-header">
        <div class="header-left">
          <div class="header-avatar">
            <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 2C6.48 2 2 6.48 2 12C2 17.52 6.48 22 12 22C17.52 22 22 17.52 22 12C22 6.48 17.52 2 12 2Z" fill="white" opacity="0.9"/>
              <circle cx="9" cy="10" r="1.5" fill="#FF6B35"/>
              <circle cx="15" cy="10" r="1.5" fill="#FF6B35"/>
              <path d="M8 15C8 15 9.5 17 12 17C14.5 17 16 15 16 15" stroke="#FF6B35" stroke-width="1.5" stroke-linecap="round" fill="none"/>
            </svg>
          </div>
          <div class="header-info">
            <span class="header-title">外卖助手</span>
            <span class="header-status" :class="{ online: isConnected }">
              {{ isConnected ? '在线' : '连接中...' }}
            </span>
          </div>
        </div>
        <button class="header-close" @click="closePanel" aria-label="关闭">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12"/>
          </svg>
        </button>
      </header>

      <!-- ── 聊天消息区域 ── -->
      <div class="chat-messages" ref="messagesRef" @scroll="onScroll">
        <!-- 欢迎消息 -->
        <div v-if="messages.length === 0" class="welcome-message">
          <div class="welcome-icon">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="2" y="4" width="20" height="16" rx="3" fill="#FFF3E0" stroke="#FF6B35" stroke-width="1.5"/>
              <circle cx="9" cy="11" r="1.5" fill="#FF6B35"/>
              <circle cx="15" cy="11" r="1.5" fill="#FF6B35"/>
              <path d="M8 15.5C8 15.5 9.5 17 12 17C14.5 17 16 15.5 16 15.5" stroke="#FF6B35" stroke-width="1.5" stroke-linecap="round" fill="none"/>
            </svg>
          </div>
          <h3 class="welcome-title">你好！我是你的外卖助手 🎉</h3>
          <p class="welcome-desc">
            我可以帮你查询订单、推荐菜品、查看配送进度，<br />
            也可以处理售后问题！
          </p>
        </div>

        <!-- 消息列表 -->
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="message-row"
          :class="msg.role"
        >
          <!-- AI 消息头像 -->
          <div v-if="msg.role === 'assistant'" class="msg-avatar">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
              <circle cx="12" cy="12" r="10" fill="#FF6B35"/>
              <circle cx="9" cy="10" r="1.5" fill="white"/>
              <circle cx="15" cy="10" r="1.5" fill="white"/>
              <path d="M8 14.5C8 14.5 9.5 16.5 12 16.5C14.5 16.5 16 14.5 16 14.5" stroke="white" stroke-width="1.5" stroke-linecap="round" fill="none"/>
            </svg>
          </div>

          <!-- 消息气泡：只有当消息有内容时才渲染 -->
          <!-- isLoading 时的空 AI 占位由下方的三点动画展示，避免空白气泡闪现 -->
          <div
            v-if="shouldShowBubble(msg, idx)"
            class="message-bubble"
            :class="{ 'is-streaming': msg.role === 'assistant' && idx === messages.length - 1 && isLoading }"
            v-html="formatMessage(msg.content)"
          ></div>
        </div>

        <!-- 加载中的三点动画（替代空白气泡） -->
        <div v-if="showLoadingDots" class="message-row assistant">
          <div class="msg-avatar">
            <svg viewBox="0 0 24 24" width="28" height="28" fill="none">
              <circle cx="12" cy="12" r="10" fill="#FF6B35"/>
              <circle cx="9" cy="10" r="1.5" fill="white"/>
              <circle cx="15" cy="10" r="1.5" fill="white"/>
              <path d="M8 14.5C8 14.5 9.5 16.5 12 16.5C14.5 16.5 16 14.5 16 14.5" stroke="white" stroke-width="1.5" stroke-linecap="round" fill="none"/>
            </svg>
          </div>
          <div class="message-bubble loading">
            <span class="dot"></span>
            <span class="dot"></span>
            <span class="dot"></span>
          </div>
        </div>

        <!-- 底部锚点用于自动滚动 -->
        <div ref="bottomRef"></div>
      </div>

      <!-- ── 快速建议芯片（仅无对话时显示） ── -->
      <div v-if="messages.length === 0" class="quick-suggestions">
        <button
          v-for="(suggestion, idx) in suggestions"
          :key="idx"
          class="suggestion-chip"
          @click="sendQuickMessage(suggestion.text)"
        >
          {{ suggestion.label }}
        </button>
      </div>

      <!-- ── 输入区域 ── -->
      <div class="input-area">
        <div class="input-wrapper">
          <textarea
            ref="inputRef"
            v-model="inputMessage"
            class="input-field"
            placeholder="输入你的问题..."
            rows="1"
            @keydown.enter.exact.prevent="handleSend"
            @input="autoResizeTextarea"
            :disabled="isLoading"
          ></textarea>
          <button
            class="send-btn"
            :disabled="!inputMessage.trim() || isLoading"
            @click="handleSend"
            aria-label="发送"
          >
            <!-- 加载中旋转图标，否则发送图标 -->
            <svg v-if="!isLoading" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/>
            </svg>
            <svg v-else class="spinner" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" stroke-dasharray="31.4 31.4" stroke-linecap="round"/>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * AgentPanel.vue — Vue 3 Composition API
 * =========================================
 * 使用 <script setup> 语法，依赖 @/composables/useAgentChat 共享状态。
 */
import { ref, nextTick, computed, onBeforeUnmount } from 'vue'
import { useAgentChat } from '@/composables/useAgentChat'

// ── Props ──
//
// fullPage: 是否以全屏模式渲染
//   false = 浮动 FAB + 侧滑面板（默认行为）
//   true  = 填充父容器，无 FAB，无固定定位
const props = withDefaults(
  defineProps<{
    fullPage?: boolean
  }>(),
  {
    fullPage: false,
  },
)

// ── 使用共享聊天状态（单例） ──
const {
  messages,
  isLoading,
  isConnected,
  sendMessage,
  clearMessages,
} = useAgentChat()

// ── 本地状态 ──

/** 面板是否可见（仅浮动模式使用） */
const panelVisible = ref(props.fullPage) // 全屏模式默认可见

/** 输入框消息文本 */
const inputMessage = ref('')

/** 是否显示三点 loading（最后一条 AI 消息内容为空时显示，替代空白气泡） */
const showLoadingDots = computed(() => {
  if (!isLoading.value) return false
  const last = messages.value[messages.value.length - 1]
  return last && last.role === 'assistant' && last.content === ''
})

// ── DOM 引用 ──

const messagesRef = ref<HTMLDivElement | null>(null)
const bottomRef = ref<HTMLDivElement | null>(null)
const inputRef = ref<HTMLTextAreaElement | null>(null)

// ── 响应式计算 ──

/** 是否为移动端（仅浮动模式使用） */
const isMobile = computed(() => {
  if (typeof window !== 'undefined') {
    return window.innerWidth <= 640
  }
  return false
})

// ── 快速建议列表 ──

const suggestions = [
  { label: '📋 查订单', text: '查一下我的订单' },
  { label: '🍜 推荐菜品', text: '推荐几个菜' },
  { label: '🚚 配送进度', text: '配送情况' },
  { label: '💬 客服帮助', text: '我需要客服帮助' },
]

// ── 面板控制（仅浮动模式） ──

function openPanel() {
  panelVisible.value = true
  nextTick(() => {
    inputRef.value?.focus()
    scrollToBottom()
  })
}

function closePanel() {
  panelVisible.value = false
}

// ── 消息渲染判定 ──

/**
 * 判断消息气泡是否应该渲染
 * 跳过空内容的 AI 消息（由三点 loading 代替），避免同时出现空白气泡 + 三点动画
 */
function shouldShowBubble(msg: { role: string; content: string }, idx: number): boolean {
  if (!msg.role) return false
  if (msg.content === '') {
    // 如果这是最后一 AI 条占位消息且正在加载，不渲染（由 showLoadingDots 替代）
    if (showLoadingDots.value && idx === messages.value.length - 1) {
      return false
    }
    return false
  }
  return true
}

// ── 消息发送 ──

function sendQuickMessage(text: string) {
  // 直接发送传入的文字，不经过 inputMessage 中转
  // 避免 handleSend 被 @keydown.enter 和 @click 同时触发导致的重复问题
  if (!text.trim() || isLoading.value) return
  scrollToBottom()
  sendMessage(text)
  setTimeout(() => scrollToBottom(), 150)
}

function handleSend() {
  const text = inputMessage.value.trim()
  if (!text || isLoading.value) return

  // 清空输入框并重置高度
  inputMessage.value = ''
  resetTextareaHeight()

  // 滚动到底部
  scrollToBottom()

  // 调用共享的 sendMessage 方法
  sendMessage(text)

  // 延迟滚动确保 AI 消息占位渲染后再滚一次
  setTimeout(() => scrollToBottom(), 100)
}

// ── 工具方法 ──

/**
 * 格式化消息内容（安全转义 HTML + Markdown 风格格式化）
 */
function formatMessage(content: string): string {
  if (!content) return ''

  // 转义 HTML 特殊字符
  let html = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  // 粗体 **text**
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')

  // 换行
  html = html.replace(/\n/g, '<br />')

  // 链接
  html = html.replace(
    /(https?:\/\/[^\s<]+)/g,
    '<a href="$1" target="_blank" rel="noopener noreferrer">$1</a>',
  )

  return html
}

/**
 * 自动滚动到聊天底部
 */
async function scrollToBottom() {
  await nextTick()
  if (bottomRef.value) {
    bottomRef.value.scrollIntoView({ behavior: 'smooth', block: 'end' })
  }
}

/**
 * 滚动事件处理（预留）
 */
function onScroll() {
  // 预留：可实现"滚动到顶部加载更多"等功能
}

/**
 * 自动调整输入框高度
 */
function autoResizeTextarea() {
  const el = inputRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 120) + 'px'
}

/**
 * 重置输入框高度
 */
function resetTextareaHeight() {
  const el = inputRef.value
  if (el) {
    el.style.height = 'auto'
  }
}

// ── 生命周期 ──

onBeforeUnmount(() => {
  // 注意：不在这里调用 resetSession()，因为切换到全屏页面时
  // 浮动面板卸载但状态由 useAgentChat 单例保持。
  // 只有在组件明确替换场景（如路由跳转离开）时才需要清除。
})
</script>

<style scoped>
/* ============================================================
   AgentPanel 样式
   ============================================================ */

/* ── FAB 按钮（仅浮动模式） ── */

.agent-fab {
  position: fixed;
  bottom: 80px;
  right: 20px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 16px rgba(255, 107, 53, 0.4);
  z-index: 1300;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.agent-fab:hover {
  transform: scale(1.1) translateY(-2px);
  box-shadow: 0 6px 20px rgba(255, 107, 53, 0.5);
}

.agent-fab:active {
  transform: scale(0.95);
}

.fab-icon {
  width: 28px;
  height: 28px;
}

/* FAB 进入/离开动画 */
.fab-fade-enter-active,
.fab-fade-leave-active {
  transition: all 0.3s ease;
}
.fab-fade-enter-from,
.fab-fade-leave-to {
  opacity: 0;
  transform: scale(0.8);
}

/* ── 面板容器 ── */

/* 浮动模式：固定在右下角 */
.agent-panel.is-floating {
  position: fixed;
  bottom: 0;
  right: 0;
  width: 400px;
  height: 100vh;
  max-height: 100vh;
  z-index: 1400;
  display: flex;
  flex-direction: column;
  pointer-events: none;
}

.agent-panel.is-floating.is-mobile {
  width: 100%;
}

/* 全屏模式：填满父容器 */
.agent-panel.is-fullpage {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  pointer-events: auto;
}

/* 遮罩层（浮动 + 移动端） */
.agent-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: -1;
  pointer-events: auto;
}

/* 面板内部容器 */
.agent-panel-inner {
  width: 100%;
  height: 100%;
  background: white;
  display: flex;
  flex-direction: column;
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.12);
  pointer-events: auto;
  position: relative;
}

/* 全屏模式下去掉圆角和阴影 */
.agent-panel.is-fullpage .agent-panel-inner {
  box-shadow: none;
  border-radius: 0;
}

/* 浮动面板滑入/滑出动画 */
.panel-slide-enter-active,
.panel-slide-leave-active {
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.panel-slide-enter-from,
.panel-slide-leave-to {
  transform: translateX(100%);
}

/* ── 头部（仅浮动模式） ── */

.agent-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-avatar {
  width: 36px;
  height: 36px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(4px);
}

.header-avatar svg {
  width: 24px;
  height: 24px;
}

.header-info {
  display: flex;
  flex-direction: column;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  line-height: 1.3;
}

.header-status {
  font-size: 11px;
  opacity: 0.8;
}

.header-status.online {
  opacity: 1;
}

.header-status.online::before {
  content: '';
  display: inline-block;
  width: 6px;
  height: 6px;
  background: #4ade80;
  border-radius: 50%;
  margin-right: 4px;
  vertical-align: middle;
  animation: pulse-dot 2s infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.header-close {
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s;
}

.header-close:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* ── 聊天消息区域 ── */

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #F8F9FA;
  display: flex;
  flex-direction: column;
  gap: 16px;
  scroll-behavior: smooth;
}

/* 自定义滚动条 */
.chat-messages::-webkit-scrollbar {
  width: 6px;
}
.chat-messages::-webkit-scrollbar-track {
  background: transparent;
}
.chat-messages::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 3px;
}
.chat-messages::-webkit-scrollbar-thumb:hover {
  background: #bbb;
}

/* ── 欢迎消息 ── */

.welcome-message {
  text-align: center;
  padding: 40px 20px 20px;
}

.welcome-icon {
  margin-bottom: 16px;
}

.welcome-title {
  font-size: 18px;
  color: #333;
  margin: 0 0 8px;
  font-weight: 600;
}

.welcome-desc {
  font-size: 14px;
  color: #888;
  line-height: 1.6;
  margin: 0;
}

/* ── 消息行 ── */

.message-row {
  display: flex;
  gap: 8px;
  max-width: 85%;
}

.message-row.user {
  flex-direction: row-reverse;
  align-self: flex-end;
}

.message-row.assistant {
  align-self: flex-start;
}

/* 消息头像 */
.msg-avatar {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}

.msg-avatar svg {
  width: 28px;
  height: 28px;
}

/* ── 消息气泡 ── */

.message-bubble {
  padding: 10px 14px;
  border-radius: 16px;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
  position: relative;
}

/* 用户消息气泡（橙色渐变） */
.message-row.user .message-bubble {
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  border-bottom-right-radius: 4px;
}

/* AI 消息气泡（白色 + 阴影） */
.message-row.assistant .message-bubble {
  background: white;
  color: #333;
  border: 1px solid #eee;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border-bottom-left-radius: 4px;
}

/* 正在流式接收时的脉冲边框 */
.message-bubble.is-streaming {
  border-color: #FF6B35;
  box-shadow: 0 2px 8px rgba(255, 107, 53, 0.15);
}

/* 气泡内部链接 */
.message-bubble a {
  color: inherit;
  text-decoration: underline;
  opacity: 0.9;
}

.message-row.assistant .message-bubble strong {
  color: #FF6B35;
}

/* ── 加载指示器（三跳点） ── */

.message-bubble.loading {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 14px 18px;
}

/* 三个独立的跳点 span */
.dot {
  width: 8px;
  height: 8px;
  background: #FF6B35;
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
}

.dot:nth-child(1) {
  animation-delay: -0.32s;
}

.dot:nth-child(2) {
  animation-delay: -0.16s;
}

.dot:nth-child(3) {
  animation-delay: 0s;
}

@keyframes bounce {
  0%, 80%, 100% {
    transform: scale(0.6);
    opacity: 0.4;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

/* ── 快速建议芯片 ── */

.quick-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0 20px 12px;
  flex-shrink: 0;
}

.suggestion-chip {
  padding: 6px 14px;
  border: 1.5px solid #FFD4B8;
  background: white;
  border-radius: 20px;
  font-size: 13px;
  color: #FF6B35;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.suggestion-chip:hover {
  background: #FFF3E8;
  border-color: #FF6B35;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(255, 107, 53, 0.15);
}

/* ── 输入区域 ── */

.input-area {
  padding: 12px 20px 16px;
  border-top: 1px solid #eee;
  background: white;
  flex-shrink: 0;
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  background: #F5F5F5;
  border-radius: 24px;
  padding: 4px 4px 4px 16px;
  transition: box-shadow 0.2s;
}

.input-wrapper:focus-within {
  box-shadow: 0 0 0 2px rgba(255, 107, 53, 0.2);
  background: white;
}

.input-field {
  flex: 1;
  border: none;
  background: transparent;
  padding: 8px 0;
  font-size: 14px;
  color: #333;
  outline: none;
  resize: none;
  max-height: 120px;
  line-height: 1.5;
  font-family: inherit;
}

.input-field:disabled {
  color: #999;
  cursor: not-allowed;
}

.input-field::placeholder {
  color: #aaa;
}

.send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.send-btn:hover:not(:disabled) {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(255, 107, 53, 0.3);
}

.send-btn:active:not(:disabled) {
  transform: scale(0.95);
}

.send-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* 发送按钮旋转动画（加载中） */
.spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ── 响应式适配 ── */

@media (min-width: 641px) {
  .agent-panel.is-floating {
    right: 20px;
    bottom: 80px;
    width: 400px;
    height: 600px;
    border-radius: 16px;
    overflow: hidden;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  }

  .agent-panel.is-floating .agent-panel-inner {
    border-radius: 16px;
  }

  .agent-panel.is-floating .agent-header {
    border-radius: 0;
  }

  .agent-fab {
    right: 20px;
    bottom: 80px;
  }
}

@media (max-width: 640px) {
  .agent-fab {
    bottom: 72px;
    right: 16px;
    width: 52px;
    height: 52px;
  }

  .quick-suggestions {
    padding: 0 12px 8px;
    justify-content: center;
  }

  .input-area {
    padding: 8px 12px 12px;
  }
}
</style>
