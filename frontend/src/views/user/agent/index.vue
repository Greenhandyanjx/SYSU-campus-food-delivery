<template>
  <!--
    index.vue — 全屏外卖助手页面（/user/agent 路由）
    ======================================================
    与主页左下角 AgentPanel 共享同一份聊天状态（useAgentChat 单例），
    保证两个入口的聊天历史完全同步。

    此页面在 layout 的 .content 直接渲染。
    为了保证宽度 / 背景与主站一致，这里使用 .page-wrap + .page-main 结构，
    并保留底部安全区域防止导航栏遮挡输入框。
  -->
  <div class="page-wrap agent-page">
    <div class="page-main">
      <div class="agent-fullpage">
        <!-- 顶部栏 -->
        <header class="agent-header">
          <button class="back-btn" @click="goBack">‹ 返回</button>
          <h2>🤖 外卖助手</h2>
          <button class="clear-btn" @click="confirmClear">🗑️ 清空</button>
        </header>

        <!-- 聊天主体 -->
        <div class="chat-body" ref="messagesRef" @scroll="onScroll">
          <!-- 欢迎页 -->
          <div v-if="messages.length === 0" class="welcome">
            <div class="welcome-icon">🤖</div>
            <p>你好！我是你的外卖助手 🎉<br />查订单、推荐菜品、配送进度、售后问题都可以问我～</p>
          </div>

          <!-- 消息列表 -->
          <div
            v-for="(msg, idx) in messages"
            :key="idx"
            class="msg-row"
            :class="msg.role"
          >
            <!-- AI 头像：流式加载时最后一条 AI 消息的头像由下方 loading 动画展示，避免重复 -->
            <div v-if="msg.role === 'assistant' && !(showLoading && idx === messages.length - 1)" class="msg-avatar">🤖</div>
            <!-- 气泡：跳过空内容 AI 消息（由下方 loading 展示） -->
            <div
              v-if="shouldShowBubble(msg, idx)"
              class="msg-bubble"
              v-html="formatContent(msg.content)"
            ></div>
          </div>

          <!-- 三点加载动画 -->
          <div v-if="showLoading" class="msg-row assistant">
            <div class="msg-avatar">🤖</div>
            <div class="msg-bubble loading">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </div>
          </div>

          <div ref="bottomRef"></div>
        </div>

        <!-- 快捷建议 -->
        <div v-if="messages.length === 0" class="suggestions">
          <button v-for="(s, i) in suggestions" :key="i" class="chip" @click="sendQuick(s.text)">
            {{ s.label }}
          </button>
        </div>

        <!-- 输入区域 — padding-bottom 防底部导航栏遮挡 -->
        <div class="input-area">
          <textarea
            ref="inputRef"
            v-model="inputMsg"
            class="input-field"
            placeholder="输入你的问题..."
            rows="1"
            @keydown.enter.exact.prevent="onSend"
            @input="autoResize"
            :disabled="isLoading"
          ></textarea>
          <button class="send-btn" :disabled="!inputMsg.trim() || isLoading" @click="onSend">
            <svg v-if="!isLoading" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/>
            </svg>
            <svg v-else class="spin" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
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
 * index.vue — Vue 3 <script setup>
 * ================================
 * 使用 useAgentChat 单例与浮动面板共享同一份聊天状态。
 */
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useAgentChat } from '@/composables/useAgentChat'

const router = useRouter()
const { messages, isLoading, sendMessage, clearMessages } = useAgentChat()

// ── 本地状态 ──
const inputMsg = ref('')
const messagesRef = ref<HTMLDivElement | null>(null)
const bottomRef = ref<HTMLDivElement | null>(null)
const inputRef = ref<HTMLTextAreaElement | null>(null)

// ── 是否显示三点 loading 动画 ──
// 条件：当前正在加载，且最后一条消息是 AI 占位（content === ''）
const showLoading = computed(() => {
  if (!isLoading.value) return false
  const last = messages.value[messages.value.length - 1]
  return last && last.role === 'assistant' && last.content === ''
})

// ── 判断消息气泡是否应该渲染 ──
// 如果当前 AI 消息是最后一条且内容为空且正在加载中，跳过（用三点动画代替）
function shouldShowBubble(msg: { role: string; content: string }, idx: number): boolean {
  if (!msg.role) return false
  if (msg.content === '') {
    // 如果是最后一条 AI 占位消息且正在加载，隐藏气泡
    if (showLoading.value && idx === messages.value.length - 1) {
      return false
    }
    return false
  }
  return true
}

// ── 操作 ──
function goBack() { router.back() }

function confirmClear() {
  if (window.confirm('确定清空所有对话？')) {
    clearMessages()
  }
}

function sendQuick(text: string) {
  inputMsg.value = text
  onSend()
}

function onSend() {
  const t = inputMsg.value.trim()
  if (!t || isLoading.value) return
  inputMsg.value = ''
  resetHeight()
  scrollDown()
  sendMessage(t)
  setTimeout(() => scrollDown(), 150)
}

function formatContent(c: string): string {
  if (!c) return ''
  let html = c
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\n/g, '<br />')
  html = html.replace(
    /(https?:\/\/[^\s<]+)/g,
    '<a href="$1" target="_blank" rel="noopener">$1</a>',
  )
  return html
}

async function scrollDown() {
  await nextTick()
  bottomRef.value?.scrollIntoView({ behavior: 'smooth', block: 'end' })
}

function autoResize() {
  const el = inputRef.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = Math.min(el.scrollHeight, 120) + 'px'
}

function resetHeight() {
  if (inputRef.value) inputRef.value.style.height = 'auto'
}

function onScroll() {
  // 预留
}

onMounted(() => {
  nextTick(() => {
    inputRef.value?.focus()
    scrollDown()
  })
})
</script>

<style scoped>
/* ================================================================
   全屏 Agent 页面样式
   使用 .page-wrap 和 .page-main 结构与首页一致，
   背景和宽度自动继承 layout 的 :before / :after 装饰列。
   ================================================================ */

/* ── 页面容器 ── */
.agent-page.page-wrap {
  display: flex;
  justify-content: center;
  position: relative;
  min-height: calc(100vh - 64px - 56px);
}

.agent-page .page-main {
  width: 60%;
  max-width: calc(100% - 40px);
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
  padding: 20px;
  margin-top: 12px;
  display: flex;
  flex-direction: column;
}

/* ── 全屏聊天容器 ── */
.agent-fullpage {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px - 56px - 72px - 24px);  /* navbar(64) + bottomNav(56) + page-main padding(40+bottomPadding) + mt(12) */
  height: calc(100dvh - 64px - 56px - 72px - 24px);
  overflow: hidden;
  border-radius: 12px;
  background: #fff;
}

/* ── 顶部栏 ── */
.agent-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  border-radius: 12px 12px 0 0;
  flex-shrink: 0;
}

.agent-header h2 {
  font-size: 16px;
  font-weight: 700;
  margin: 0;
}

.back-btn {
  background: rgba(255, 255, 255, 0.15);
  border: none;
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 14px;
  cursor: pointer;
}
.back-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.clear-btn {
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
}
.clear-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* ── 聊天消息区域 ── */
.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #F8F9FA;
}

.chat-body::-webkit-scrollbar {
  width: 5px;
}
.chat-body::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 3px;
}

/* 欢迎消息 */
.welcome {
  text-align: center;
  padding: 48px 16px 16px;
  color: #888;
  font-size: 14px;
  line-height: 1.8;
}
.welcome-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

/* ── 消息行 ── */
.msg-row {
  display: flex;
  gap: 8px;
  max-width: 85%;
}
.msg-row.user {
  flex-direction: row-reverse;
  align-self: flex-end;
}
.msg-row.assistant {
  align-self: flex-start;
}

.msg-avatar {
  font-size: 24px;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  line-height: 28px;
  text-align: center;
}

/* ── 消息气泡 ── */
.msg-bubble {
  padding: 10px 14px;
  border-radius: 14px;
  font-size: 14px;
  line-height: 1.65;
  word-break: break-word;
}

.msg-row.user .msg-bubble {
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  border-bottom-right-radius: 4px;
}

.msg-row.assistant .msg-bubble {
  background: white;
  color: #333;
  border: 1px solid #eee;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  border-bottom-left-radius: 4px;
}

.msg-bubble strong {
  color: #FF6B35;
}

.msg-bubble a {
  color: #FF6B35;
  text-decoration: underline;
}

/* ── 三点加载动画 ── */
.msg-bubble.loading {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 14px 18px;
  min-width: 56px;
}

.dot {
  width: 8px;
  height: 8px;
  background: #FF6B35;
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
}
.dot:nth-child(1) { animation-delay: -0.32s; }
.dot:nth-child(2) { animation-delay: -0.16s; }
.dot:nth-child(3) { animation-delay: 0s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}

/* ── 快捷建议 ── */
.suggestions {
  display: flex;
  gap: 8px;
  padding: 0 16px 12px;
  flex-wrap: wrap;
  justify-content: center;
  flex-shrink: 0;
}

.chip {
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid #ffc8a8;
  background: white;
  color: #FF6B35;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}
.chip:hover {
  background: #FFF3E8;
  border-color: #FF6B35;
}

/* ── 输入区域 ── */
.input-area {
  flex-shrink: 0;
  padding: 12px 16px;
  /* 底部留出足够的 padding 防止被导航栏（56px）遮挡 */
  padding-bottom: calc(16px + env(safe-area-inset-bottom, 0px));
  background: white;
  border-top: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
}

.input-field {
  flex: 1;
  border: 1px solid #e0e0e0;
  border-radius: 22px;
  padding: 10px 16px;
  font-size: 14px;
  outline: none;
  resize: none;
  line-height: 1.4;
  max-height: 120px;
  font-family: inherit;
  transition: border-color 0.2s;
}
.input-field:focus {
  border-color: #FF6B35;
}
.input-field:disabled {
  background: #f5f5f5;
}

.send-btn {
  margin-left: 10px;
  width: 42px;
  height: 42px;
  border-radius: 50%;
  border: none;
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}
.send-btn:hover:not(:disabled) {
  transform: scale(1.08);
}
.send-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.spin {
  animation: rotate 1s linear infinite;
}
@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ── 移动端适配 ── */
@media (max-width: 640px) {
  .agent-page .page-main {
    max-width: 100%;
    padding: 12px;
    border-radius: 0;
  }
  .agent-fullpage {
    height: calc(100vh - 64px - 56px);
    height: calc(100dvh - 64px - 56px);
    border-radius: 0;
  }
}
</style>
