<template>
  <div class="page-wrap agent-page">
    <div class="page-main agent-layout">
      <!-- ═══ 左侧：会话列表（WeChat 风格） ═══ -->
      <aside class="sidebar">
        <header class="sidebar-header">
          <h2>💬 消息</h2>
        </header>
        <div class="session-list">
          <template v-for="day in sidebarDays" :key="day.dateStr">
            <div class="date-divider">{{ day.dateLabel }}</div>
            <div
              class="session-item"
              :class="{ active: day.dateStr === activeDate }"
              @click="onSelectDay(day)"
            >
              <div class="session-avatar"><img :src="agentIcon" alt="AI" /></div>
              <div class="session-info">
                <div class="session-title">{{ day.sessions[0].last_message || '新会话' }}</div>
                <div class="session-preview">{{ day.dateStr === activeDate ? visibleCount : day.totalMessages }} 条消息</div>
              </div>
            </div>
          </template>
          <div v-if="sidebarDays.length === 0" class="empty-sessions">
            暂无历史记录
          </div>
        </div>
      </aside>

      <!-- ═══ 右侧：聊天区域 ═══ -->
      <main class="chat-main">
        <header class="agent-header">
          <button class="back-btn" @click="goBack">‹ 返回</button>
          <h2><img class="header-icon" :src="agentIcon" alt="" /> 外卖助手</h2>
          <button class="clear-btn" @click="confirmClear">🗑️ 清空</button>
        </header>

        <div class="chat-body" ref="messagesRef">
          <div v-if="messages.length === 0 && currentSessionId" class="welcome">
            <div class="welcome-icon"><img :src="agentIcon" alt="AI" /></div>
            <p>你好！我是你的外卖助手<br />查订单、推荐菜品、配送进度、售后问题都可以问我～</p>
          </div>
          <div v-if="!currentSessionId" class="welcome">
            <div class="welcome-icon">💬</div>
            <p>选择一个会话或新建一个对话</p>
          </div>

          <template v-for="(group, gi) in messageGroups" :key="gi">
            <!-- 日期间隔（类似微信的日期条） -->
            <div v-if="group.dateLabel" class="date-separator">{{ group.dateLabel }}</div>

            <div
              v-for="(msg, mi) in group.items"
              :key="gi + '-' + mi"
              class="msg-row"
              :class="msg.role"
            >
              <div v-if="msg.role === 'assistant' && msg.content" class="msg-avatar"><img :src="agentIcon" alt="AI" /></div>
              <div class="msg-content-wrap">
                <!-- 时间戳（类似微信：同 sender 5分钟内不重复） -->
                <div v-if="msg.showTime" class="msg-time">{{ formatTime(msg.timestamp) }}</div>
                <div class="msg-bubble-row">
                  <div v-if="shouldShowBubble(msg, mi, group.items)" class="msg-bubble" v-html="formatContent(msg.content)"></div>
                  <button
                    v-if="msg.role === 'assistant' && msg.content && ttsSupported"
                    class="speaker-btn"
                    :class="{ speaking: isSpeaking && speakingText === msg.content }"
                    @click="toggleSpeak(msg.content)"
                    :title="isSpeaking && speakingText === msg.content ? '停止朗读' : '朗读'"
                  >
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M11 5L6 9H2v6h4l5 4V5z"/>
                      <path v-if="isSpeaking && speakingText === msg.content" d="M16 9a5 5 0 0 1 0 6M19 6a8 8 0 0 1 0 12"/>
                    </svg>
                  </button>
                </div>
                <template v-if="msg.role === 'assistant' && msg.orderCards && msg.orderCards.length > 0">
                  <div class="order-cards-wrap">
                    <AgentOrderCard
                      v-for="(card, ci) in msg.orderCards.slice(0, 10)"
                      :key="'card-' + ci"
                      :data="card"
                      @send-message="onCardAction"
                    />
                    <div v-if="msg.orderCards.length > 10" class="order-cards-overflow" @click="onCardAction('查一下我的订单')">
                      <span>还有 {{ msg.orderCards.length - 10 }} 个订单 <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12h14M12 5l7 7-7 7"/></svg></span>
                    </div>
                  </div>
                </template>
              </div>
            </div>
          </template>

          <div v-if="showLoading" class="msg-row assistant">
            <div class="msg-avatar"><img :src="agentIcon" alt="AI" /></div>
            <div class="msg-bubble loading">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </div>
          </div>
          <div ref="bottomRef"></div>
        </div>

        <div v-if="messages.length === 0 && currentSessionId" class="suggestions">
          <button v-for="(s, i) in suggestions" :key="i" class="chip" @click="sendQuick(s.text)">
            {{ s.label }}
          </button>
        </div>

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
          <button
            v-if="speechSupported"
            class="mic-btn"
            :class="{ listening: isListening }"
            :disabled="isLoading"
            @click="toggleMic"
            title="语音输入"
          >
            <svg v-if="!isListening" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="2" width="6" height="11" rx="3"/>
              <path d="M5 10a7 7 0 0 0 14 0M12 17v4M9 21h6"/>
            </svg>
            <svg v-else class="pulse" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="2" width="6" height="11" rx="3"/>
              <path d="M5 10a7 7 0 0 0 14 0M12 17v4M9 21h6"/>
            </svg>
          </button>
          <button class="send-btn" :disabled="!inputMsg.trim() || isLoading" @click="onSend">
            <svg v-if="!isLoading" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/>
            </svg>
            <svg v-else class="spin" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10" stroke-dasharray="31.4 31.4" stroke-linecap="round"/>
            </svg>
          </button>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAgentChat, ChatMsg, SessionItem } from '@/composables/useAgentChat'
import { useSpeechRecognition } from '@/composables/useSpeechRecognition'
import { useSpeechSynthesis } from '@/composables/useSpeechSynthesis'
import AgentOrderCard from '@/components/Chat/AgentOrderCard.vue'
import agentIcon from '@/assets/icons/agent.svg'

const router = useRouter()
const {
  messages, isLoading, currentSessionId, sessionGroups,
  sendMessage, clearMessages, loadSessions, loadSession,
  loadDaySessions, newSession, setToken, proactiveLunch,
} = useAgentChat()

const inputMsg = ref('')

// ── 语音输入（麦克风） ──
const {
  isListening,
  supported: speechSupported,
  toggle: toggleMic,
} = useSpeechRecognition({
  lang: 'zh-CN',
  onResult: (text: string) => { inputMsg.value = text },
})

// ── 语音输出（朗读） ──
const {
  isSpeaking,
  supported: ttsSupported,
  currentText: speakingText,
  toggle: toggleSpeak,
} = useSpeechSynthesis()

const messagesRef = ref<HTMLDivElement | null>(null)
const bottomRef = ref<HTMLDivElement | null>(null)
const inputRef = ref<HTMLTextAreaElement | null>(null)

const suggestions = [
  { label: '📋 查订单', text: '帮我查一下我的订单' },
  { label: '🍽️ 推荐菜品', text: '今天有什么好吃的推荐？' },
  { label: '🚚 配送进度', text: '我的订单配送进度如何了？' },
  { label: '❓ 售后问题', text: '我有售后问题要处理' },
]

const showLoading = computed(() => {
  if (!isLoading.value) return false
  const last = messages.value[messages.value.length - 1]
  return last && last.role === 'assistant' && last.content === ''
})

interface SidebarDay {
  dateLabel: string
  dateStr: string
  sessions: SessionItem[]
  count: number
  totalMessages: number
}

const sidebarDays = computed<SidebarDay[]>(() => {
  const days: SidebarDay[] = []
  const keys = Object.keys(sessionGroups.value).sort().reverse()
  for (const dateStr of keys) {
    const sessions = sessionGroups.value[dateStr]
    if (!sessions || sessions.length === 0) continue
    days.push({
      dateLabel: formatDateLabel(dateStr),
      dateStr,
      sessions,
      count: sessions.length,
      totalMessages: sessions.reduce((sum, s) => sum + (s.message_count || 0), 0),
    })
  }
  return days
})

/** 当前可见问答对数（用户提问且有有效回复的数量） */
const visibleCount = computed(() => {
  let count = 0
  let pendingUser = false
  for (const m of messages.value) {
    if (m.role === 'user') {
      pendingUser = true
    } else if (m.role === 'assistant' && m.content && pendingUser) {
      count++
      pendingUser = false
    }
  }
  return count
})

const activeDate = computed(() => {
  for (const [dateStr, sessions] of Object.entries(sessionGroups.value)) {
    if (sessions.some(s => s.session_id === currentSessionId.value)) return dateStr
  }
  return null
})

interface MessageGroup {
  dateLabel: string
  items: ChatMsg[]
}

const messageGroups = computed<MessageGroup[]>(() => {
  const groups: MessageGroup[] = []
  let currentDateStr = ''
  let currentGroup: MessageGroup | null = null

  // 显式追踪所有消息属性，确保 orderCards/content 变更时重新计算
  for (const m of messages.value) {
    void m.content
    void m.orderCards
  }

  for (let i = 0; i < messages.value.length; i++) {
    const msg = messages.value[i]
    const msgDate = msg.timestamp ? new Date(msg.timestamp * 1000) : new Date()
    const dateStr = msgDate.toDateString()

    if (dateStr !== currentDateStr) {
      currentDateStr = dateStr
      currentGroup = {
        dateLabel: formatChatDateLabel(msgDate),
        items: [],
      }
      groups.push(currentGroup)
    }

    // 在每个问答（用户消息）前面显示 HH:MM 时间戳
    const showTime = msg.role === 'user'

    currentGroup!.items.push({ ...msg, showTime })
  }

  return groups
})

function shouldShowBubble(msg: ChatMsg, idx: number, group: ChatMsg[]): boolean {
  if (!msg.role) return false
  if (msg.content === '') {
    if (showLoading.value && idx === messages.value.length - 1) return false
    return false
  }
  return true
}

function formatTime(ts?: number): string {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const h = d.getHours().toString().padStart(2, '0')
  const m = d.getMinutes().toString().padStart(2, '0')
  return `${h}:${m}`
}

function formatChatDateLabel(d: Date): string {
  const today = new Date()
  if (d.toDateString() === today.toDateString()) return '今天'
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return '昨天'
  if (d.getFullYear() === today.getFullYear()) {
    return `${d.getMonth() + 1}月${d.getDate()}日`
  }
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
}

function goBack() { router.back() }

function confirmClear() {
  if (window.confirm('确定清空当前对话？')) {
    clearMessages()
  }
}

async function onSelectSession(s: { session_id: string }) {
  await loadSession(s.session_id)
  await nextTick()
  scrollDown()
}

async function onSelectDay(day: SidebarDay) {
  await loadDaySessions(day.sessions)
  await nextTick()
  scrollDown()
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
  sendMessage(t)
  setTimeout(() => scrollDown(), 150)
}

function onCardAction(text: string) {
  if (!text.trim() || isLoading.value) return
  inputMsg.value = text
  onSend()
}

function formatContent(c: string): string {
  if (!c) return ''
  let html = c.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/\n/g, '<br />')
  html = html.replace(/(https?:\/\/[^\s<]+)/g, '<a href="$1" target="_blank" rel="noopener">$1</a>')
  return html
}

function formatDateLabel(date: string): string {
  const today = new Date()
  const d = new Date(date)
  if (isNaN(d.getTime())) return date
  if (d.toDateString() === today.toDateString()) return '今天'
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) return '昨天'
  return `${d.getMonth() + 1}月${d.getDate()}日`
}

function formatSessionPreview(text: string): string {
  if (!text) return '新会话'
  return text.length > 40 ? text.substring(0, 40) + '...' : text
}

async function checkProactiveRecommendation() {
  // 时间闸：仅 10:00-11:00 触发
  const now = new Date()
  const hour = now.getHours()
  if (hour < 10 || hour >= 11) return

  // 每日一次：localStorage 标记
  const todayKey = `proactive_lunch_${now.toDateString()}`
  try {
    if (localStorage.getItem(todayKey)) return
  } catch { return }

  // 非阻塞调用，不延迟页面渲染
  setTimeout(async () => {
    try {
      const msg = await proactiveLunch()
      if (msg && messages.value.length > 0) {
        messages.value.unshift({
          role: 'assistant',
          content: msg,
          timestamp: Date.now() / 1000,
        })
        localStorage.setItem(todayKey, '1')
        await nextTick()
        scrollDown()
      }
    } catch { /* 静默失败 */ }
  }, 500)  // 延迟 500ms，确保会话历史已渲染
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

/** 格式化为 YYYY-MM-DD（用本地时间，与服务端 datetime.now() 对齐） */
function formatLocalDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

onMounted(async () => {
  setToken()
  await loadSessions()

  // 导航回来：已有会话且消息还在，保持不动（只刷新侧边栏）
  if (currentSessionId.value && messages.value.length > 0) {
    return
  }

  // 首次加载：检测今天是否有会话
  const todayStr = formatLocalDate(new Date())
  const todaySessions = sessionGroups.value[todayStr]
  if (todaySessions && todaySessions.length > 0) {
    await loadDaySessions(todaySessions)
  } else {
    newSession()
  }

  await nextTick()
  inputRef.value?.focus()
  scrollDown()

  // 非阻塞：主动午餐推荐（时间闸在函数内部）
  checkProactiveRecommendation()
})
</script>

<style scoped>
/* ── 左右布局容器 ── */
.agent-layout {
  display: flex !important;
  flex-direction: row !important;
  padding: 0 !important;
  width: 85% !important;
  max-width: 1100px !important;
  height: calc(100vh - 64px - 56px - 24px);
  height: calc(100dvh - 64px - 56px - 24px);
  overflow: hidden;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
}

/* ═══ 左侧栏 ═══ */
.sidebar {
  width: 300px;
  min-width: 260px;
  border-right: 1px solid #f0f0f0;
  display: flex;
  flex-direction: column;
  background: #fafafa;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
  flex-shrink: 0;
}
.sidebar-header h2 {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
}

.new-chat-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.2s;
}
.new-chat-btn:hover { background: #f0f0f0; }

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.date-divider {
  padding: 6px 16px 4px;
  font-size: 12px;
  color: #999;
  font-weight: 600;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.15s;
  border-left: 3px solid transparent;
}
.session-item:hover { background: #f0f0f0; }
.session-item.active {
  background: #fff3e8;
  border-left-color: #FF6B35;
}

.session-avatar {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 10px;
}
.session-avatar img {
  width: 32px;
  height: 32px;
  object-fit: contain;
}

.session-info {
  flex: 1;
  min-width: 0;
}
.session-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.session-preview {
  font-size: 12px;
  color: #999;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 2px;
}

.empty-sessions {
  text-align: center;
  color: #bbb;
  padding: 32px 16px;
  font-size: 14px;
}

/* ═══ 右侧聊天区 ═══ */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.agent-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: linear-gradient(135deg, #FF6B35, #FF8C42);
  color: white;
  flex-shrink: 0;
}
.agent-header h2 {
  font-size: 16px;
  font-weight: 700;
  margin: 0;
}
.back-btn {
  background: rgba(255,255,255,0.15);
  border: none;
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 14px;
  cursor: pointer;
}
.back-btn:hover { background: rgba(255,255,255,0.3); }
.clear-btn {
  background: rgba(255,255,255,0.15);
  border: 1px solid rgba(255,255,255,0.3);
  color: white;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
}
.clear-btn:hover { background: rgba(255,255,255,0.3); }

/* ── 聊天消息 ── */
.chat-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #F8F9FA;
}
.chat-body::-webkit-scrollbar { width: 5px; }
.chat-body::-webkit-scrollbar-thumb { background: #ddd; border-radius: 3px; }

.welcome {
  text-align: center;
  padding: 48px 16px 16px;
  color: #888;
  font-size: 14px;
  line-height: 1.8;
}
.welcome-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.welcome-icon img {
  width: 56px;
  height: 56px;
}

.header-icon {
  width: 24px;
  height: 24px;
  vertical-align: middle;
  margin-right: 4px;
  margin-top: -2px;
}

.msg-row {
  display: flex;
  gap: 8px;
  max-width: 85%;
}
.msg-row.user { flex-direction: row-reverse; align-self: flex-end; }
.msg-row.assistant { align-self: flex-start; }

.msg-avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 8px;
}
.msg-avatar img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

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
  box-shadow: 0 2px 6px rgba(0,0,0,0.06);
  border-bottom-left-radius: 4px;
}
.msg-bubble strong { color: #FF6B35; }
.msg-bubble a { color: #FF6B35; text-decoration: underline; }

/* ── 日期分隔条（类似微信） ── */
.date-separator {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin: 8px 0;
  position: relative;
}
.date-separator::before,
.date-separator::after {
  content: '';
  position: absolute;
  top: 50%;
  width: 20%;
  height: 1px;
  background: #e0e0e0;
}
.date-separator::before { left: 8%; }
.date-separator::after { right: 8%; }

/* ── 消息时间戳 ── */
.msg-time {
  text-align: center;
  font-size: 11px;
  color: #b0b0b0;
  margin-bottom: 4px;
}

/* ── 消息内容外层 ── */
.msg-content-wrap {
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.order-cards-wrap { width: 100%; }
.order-cards-overflow {
  margin-top: 8px;
  padding: 10px 14px;
  background: #FFF8F4;
  border: 1px dashed #FFB088;
  border-radius: 12px;
  text-align: center;
  font-size: 13px;
  color: #FF6B35;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}
.order-cards-overflow:hover {
  background: #FFF0E8;
  border-color: #FF6B35;
}
.order-cards-overflow svg {
  vertical-align: middle;
}

.msg-bubble.loading {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 14px 18px;
  min-width: 56px;
}
.dot {
  width: 8px; height: 8px;
  background: #FF6B35;
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
}
.dot:nth-child(1) { animation-delay: -0.32s; }
.dot:nth-child(2) { animation-delay: -0.16s; }
.dot:nth-child(3) { animation-delay: 0s; }
@keyframes bounce {
  0%,80%,100% { transform: scale(0.6); opacity: 0.4; }
  40% { transform: scale(1); opacity: 1; }
}

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
.chip:hover { background: #FFF3E8; border-color: #FF6B35; }

.input-area {
  flex-shrink: 0;
  padding: 12px 16px;
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
.input-field:focus { border-color: #FF6B35; }
.input-field:disabled { background: #f5f5f5; }

.send-btn {
  margin-left: 10px;
  width: 42px; height: 42px;
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
.send-btn:hover:not(:disabled) { transform: scale(1.08); }
.send-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.spin { animation: rotate 1s linear infinite; }
@keyframes rotate { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

/* ── 语音按钮 ── */
.mic-btn {
  margin-left: 8px;
  width: 42px; height: 42px;
  border-radius: 50%;
  border: 1px solid #e0e0e0;
  background: white;
  color: #666;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s;
}
.mic-btn:hover:not(:disabled) { border-color: #FF6B35; color: #FF6B35; }
.mic-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.mic-btn.listening {
  background: #FF6B35;
  border-color: #FF6B35;
  color: white;
  animation: mic-pulse 1.2s ease-in-out infinite;
}
@keyframes mic-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(255, 107, 53, 0.4); }
  50% { box-shadow: 0 0 0 8px rgba(255, 107, 53, 0); }
}
.pulse { animation: pulse-icon 0.8s ease-in-out infinite alternate; }
@keyframes pulse-icon {
  from { opacity: 0.6; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1.05); }
}

.speaker-btn {
  flex-shrink: 0;
  width: 28px; height: 28px;
  border-radius: 50%;
  border: 1px solid #e0e0e0;
  background: white;
  color: #999;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 6px;
  transition: all 0.2s;
  opacity: 0;
}
.msg-row.assistant:hover .speaker-btn { opacity: 1; }
.speaker-btn:hover { border-color: #FF6B35; color: #FF6B35; }
.speaker-btn.speaking {
  opacity: 1;
  border-color: #FF6B35;
  color: #FF6B35;
  background: #FFF3E8;
}

.msg-bubble-row {
  display: flex;
  align-items: flex-start;
  gap: 4px;
}

/* ── 移动端适配 ── */
@media (max-width: 768px) {
  .agent-layout { width: 100% !important; flex-direction: column !important; height: auto; border-radius: 0; }
  .sidebar { width: 100%; max-height: 200px; border-right: none; border-bottom: 1px solid #f0f0f0; }
  .chat-main { height: calc(100vh - 64px - 56px - 200px); }
}
</style>
