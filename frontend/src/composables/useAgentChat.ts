/**
 * useAgentChat.ts — 共享 Agent 聊天状态（模块级单例）
 * =========================================================
 * 支持多会话管理，按用户隔离历史记录。
 *
 * 功能：
 *  - 会话列表（按日期分组）
 *  - 切换/新建会话
 *  - 流式/非流式聊天
 *  - 订单卡片解析
 */
import { ref } from 'vue'
import {
  sendMessage as apiSendMessage,
  sendMessageStream,
} from '@/api/agent'

// ── 订单卡片数据类型 ──
export interface OrderDish {
  name: string
  qty: number
  price: number
}

export interface OrderCardData {
  orderId: number | string
  status: number
  merchant: string
  amount: number
  deliveryFee?: number
  dishes: OrderDish[]
  consignee?: string
  address?: string
  phone?: string
  orderTime?: string
  logo?: string
}

// ── 会话项 ──
export interface SessionItem {
  session_id: string
  session_key: string
  date: string
  last_message: string
  updated_at: string
  message_count?: number
}

// ── 消息类型定义 ──
export interface ChatMsg {
  role: 'user' | 'assistant'
  content: string
  timestamp?: number      // Unix 时间戳（秒），用于展示消息时间
  orderCards?: OrderCardData[]
  showTime?: boolean
}

// ── 模块级单例状态 ──
const messages = ref<ChatMsg[]>([])
const isLoading = ref(false)
const isConnected = ref(false)
const currentSessionId = ref<string | null>(null)
const sessionGroups = ref<Record<string, SessionItem[]>>({})
const token = ref('')

const AGENT_API_BASE = import.meta.env.VITE_AGENT_API_URL || 'http://127.0.0.1:8001'

/**
 * 从消息内容中解析 ORDER_CARD 标记，提取订单卡片数据
 */
function parseOrderCards(content: string): { cleanContent: string; cards: OrderCardData[] } {
  const cards: OrderCardData[] = []
  const regex = /(?:<!--ORDER_CARD-->|\[ORDER_CARD_START\])([\s\S]*?)(?:<!--END-->|\[ORDER_CARD_END\])/g
  const cleanContent = content.replace(regex, (_match, jsonStr) => {
    try {
      const data = JSON.parse(jsonStr.trim()) as OrderCardData
      cards.push(data)
    } catch {
      // ignore parse errors
    }
    return ''
  })
  return { cleanContent: cleanContent.trim(), cards }
}

const STREAM_TIMEOUT_MS = 15000

/**
 * useAgentChat — 共享聊天状态 hook
 */
export function useAgentChat() {
  let streamTimeout: ReturnType<typeof setTimeout> | null = null

  /**
   * 设置 JWT token（从 localStorage 获取）
   */
  function setToken(t?: string) {
    token.value = t || ''
    try {
      const stored = localStorage.getItem('token')
      if (stored) token.value = stored
    } catch {
      // ignore
    }
  }

  /**
   * 获取 auth headers
   */
  function authHeaders(): Record<string, string> {
    const headers: Record<string, string> = {}
    if (token.value) {
      headers['Authorization'] = token.value.startsWith('Bearer ') ? token.value : `Bearer ${token.value}`
    }
    return headers
  }

  /**
   * 加载会话列表（按天分组）
   */
  async function loadSessions() {
    setToken()
    if (!token.value) return
    try {
      const resp = await fetch(`${AGENT_API_BASE}/api/v1/sessions`, {
        headers: authHeaders(),
      })
      if (!resp.ok) return
      const data = await resp.json()
      sessionGroups.value = data.groups || {}
    } catch (e) {
      console.warn('[AgentChat] 加载会话列表失败', e)
    }
  }

  /**
   * 加载指定会话的历史消息
   */
  async function loadSession(sessionId: string) {
    currentSessionId.value = sessionId
    messages.value = []

    const sessionKey = `u:${getUsernameFromToken()}:${sessionId}`
    try {
      const resp = await fetch(`${AGENT_API_BASE}/api/v1/history?session_key=${encodeURIComponent(sessionKey)}`)
      if (!resp.ok) return
      const data = await resp.json()
      if (data.history && Array.isArray(data.history)) {
        for (const msg of data.history) {
          if (msg.role === 'user' || msg.role === 'assistant') {
            const { cleanContent, cards } = parseOrderCards(msg.content || '')
            let ts: number | undefined = undefined
            if (msg.timestamp != null) {
              ts = typeof msg.timestamp === 'number' ? msg.timestamp : new Date(msg.timestamp).getTime() / 1000
            }
            const entry: ChatMsg = { role: msg.role, content: cleanContent, timestamp: ts }
            if (cards.length > 0) entry.orderCards = cards
            messages.value.push(entry)
          }
        }
      }
    } catch (e) {
      console.warn('[AgentChat] 加载会话历史失败', e)
    }
  }

  /**
   * 创建新会话
   */
  function newSession() {
    const id = Math.random().toString(36).substring(2, 10)
    currentSessionId.value = id
    messages.value = []
    return id
  }

  /**
   * 加载一天内所有会话的消息并合并显示
   */
  async function loadDaySessions(sessions: SessionItem[]) {
    if (!sessions.length) {
      messages.value = []
      currentSessionId.value = null
      return
    }

    // 并行拉取所有会话的历史
    const results = await Promise.allSettled(
      sessions.map(s => _loadSingleSessionHistory(s.session_key || `u:${getUsernameFromToken()}:${s.session_id}`))
    )

    // 按时间戳合并所有消息
    const allMessages: ChatMsg[] = []
    for (const result of results) {
      if (result.status === 'fulfilled' && result.value.length > 0) {
        allMessages.push(...result.value)
      }
    }
    allMessages.sort((a, b) => (a.timestamp || 0) - (b.timestamp || 0))

    messages.value = allMessages
    // 使用最后更新的会话 ID 作为当前会话（新消息写入该会话）
    currentSessionId.value = sessions.reduce((latest, s) =>
      s.updated_at > latest.updated_at ? s : latest
    , sessions[0]).session_id
  }

  /**
   * 加载单个会话的历史消息（内部工具函数）
   */
  async function _loadSingleSessionHistory(sessionKey: string): Promise<ChatMsg[]> {
    try {
      const resp = await fetch(`${AGENT_API_BASE}/api/v1/history?session_key=${encodeURIComponent(sessionKey)}`)
      if (!resp.ok) return []
      const data = await resp.json()
      const msgs: ChatMsg[] = []
      if (data.history && Array.isArray(data.history)) {
        for (const msg of data.history) {
          if (msg.role === 'user' || msg.role === 'assistant') {
            const { cleanContent, cards } = parseOrderCards(msg.content || '')
            let ts: number | undefined
            if (msg.timestamp != null) {
              ts = typeof msg.timestamp === 'number' ? msg.timestamp : new Date(msg.timestamp).getTime() / 1000
            }
            const entry: ChatMsg = { role: msg.role, content: cleanContent, timestamp: ts }
            if (cards.length > 0) entry.orderCards = cards
            msgs.push(entry)
          }
        }
      }
      return msgs
    } catch {
      return []
    }
  }

  /**
   * 从 JWT 提取用户名
   */
  function getUsernameFromToken(): string {
    const t = token.value
    if (!t) return 'anonymous'
    try {
      const payload = t.split('.')[1]
      const padding = 4 - payload.length % 4
      const b64 = padding !== 4 ? payload + '='.repeat(padding) : payload
      const decoded = JSON.parse(atob(b64))
      return decoded.username || 'anonymous'
    } catch {
      return 'anonymous'
    }
  }

  /**
   * 发送消息
   */
  async function sendMessage(text: string) {
    if (!text.trim() || isLoading.value) return
    setToken()

    // 确保有当前会话
    let sid = currentSessionId.value
    if (!sid) {
      sid = newSession()
    }

    // 添加用户消息
    messages.value.push({ role: 'user', content: text, timestamp: Date.now() / 1000 })
    isLoading.value = true
    isConnected.value = true

    const aiIdx = messages.value.length
    messages.value.push({ role: 'assistant', content: '', timestamp: Date.now() / 1000 })

    try {
      let receivedAnyChunk = false

      streamTimeout = setTimeout(() => {
        if (!receivedAnyChunk) {
          console.warn('[AgentChat] 流式超时，回退非流式')
          _fallbackNonStream(text, aiIdx)
        }
      }, STREAM_TIMEOUT_MS)

      sendMessageStream(text, {
        onMessage: (chunk: string) => {
          receivedAnyChunk = true
          if (streamTimeout) {
            clearTimeout(streamTimeout)
            streamTimeout = null
          }
          messages.value[aiIdx].content += chunk
        },

        onDone: () => {
          if (streamTimeout) {
            clearTimeout(streamTimeout)
            streamTimeout = null
          }
          isLoading.value = false
          if (messages.value[aiIdx]) {
            const { cleanContent, cards } = parseOrderCards(messages.value[aiIdx].content)
            messages.value[aiIdx].content = cleanContent
            if (cards.length > 0) {
              messages.value[aiIdx].orderCards = cards
            }
          }
          loadSessions()
        },

        onError: (error: string) => {
          if (streamTimeout) {
            clearTimeout(streamTimeout)
            streamTimeout = null
          }
          isLoading.value = false
          if (!messages.value[aiIdx]?.content) {
            _fallbackNonStream(text, aiIdx, sid)
          }
        },
      }, sid)
    } catch {
      if (streamTimeout) {
        clearTimeout(streamTimeout)
        streamTimeout = null
      }
      await _fallbackNonStream(text, aiIdx, sid)
    }
  }

  async function _fallbackNonStream(text: string, aiIdx: number, sid?: string) {
    try {
      const resp = await apiSendMessage(text, sid)
      const { cleanContent, cards } = parseOrderCards(resp.response)
      messages.value[aiIdx].content = cleanContent
      if (cards.length > 0) {
        messages.value[aiIdx].orderCards = cards
      }
      loadSessions()
    } catch (err: any) {
      const errMsg = err instanceof Error ? err.message : '请求失败'
      messages.value[aiIdx].content = `😅 连接服务失败，请稍后重试。\n\n错误详情：${errMsg}`
    } finally {
      isLoading.value = false
    }
  }

  function clearMessages() {
    messages.value = []
  }

  return {
    messages,
    isLoading,
    isConnected,
    currentSessionId,
    sessionGroups,
    sendMessage,
    clearMessages,
    loadSessions,
    loadSession,
    loadDaySessions,
    newSession,
    setToken,
  }
}
