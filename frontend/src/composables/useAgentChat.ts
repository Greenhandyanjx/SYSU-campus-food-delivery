/**
 * useAgentChat.ts — 共享 Agent 聊天状态（模块级单例）
 * =========================================================
 * 这个 composable 让 AgentPanel（浮动面板）和 Agent Page（全屏页面）
 * 共享同一个聊天状态，用户不会看到两个不同的聊天记录。
 *
 * 原理：
 *   messages、isLoading、isConnected、sessionId 的定义不在 composable
 *  函数内部，而是在模块顶层。所有组件调用 useAgentChat() 都返回
 *  同一组 ref 对象，因此状态天然共享。
 */
import { ref } from 'vue'
import {
  sendMessage as apiSendMessage,
  sendMessageStream,
  resetSession,
} from '@/api/agent'

// ── 消息类型定义 ──
export interface ChatMsg {
  role: 'user' | 'assistant'
  content: string
}

// ── 模块级单例状态（所有组件共享同一个实例） ──
const messages = ref<ChatMsg[]>([])
const isLoading = ref(false)
const isConnected = ref(false)
const sessionId = ref<string | null>(null)

// 标记是否已经初始化过（防重复操作）
let initialized = false

/**
 * 把 SSE 超时时间（毫秒）暴露为可配置常量，
 * 便于调试或根据后端响应速度调整。
 */
const STREAM_TIMEOUT_MS = 15000

/**
 * useAgentChat — 共享聊天状态 hook
 * ================================
 * 在任意组件中调用：
 *   const { messages, isLoading, sendMessage, clearMessages } = useAgentChat()
 *
 * @returns {{ messages, isLoading, isConnected, sendMessage, clearMessages }}
 */
export function useAgentChat() {
  // ── 首次调用时初始化 ──
  if (!initialized) {
    // sessionId 在 api/agent.ts 内部通过 getSessionId() 自动管理，
    // 首次调用 sendMessage / sendMessageStream 时生成。
    // 此处只需确保 initialized 标记为 true，避免重复初始化。
    initialized = true
  }

  // ── 流式超时定时器 ID ──
  let streamTimeout: ReturnType<typeof setTimeout> | null = null

  /**
   * 发送消息（优先尝试流式，15 秒无响应回退非流式）
   */
  async function sendMessage(text: string) {
    if (!text.trim() || isLoading.value) return

    // 添加用户消息
    messages.value.push({ role: 'user', content: text })
    isLoading.value = true
    isConnected.value = true

    // 预留 AI 消息的位置
    const aiIdx = messages.value.length
    messages.value.push({ role: 'assistant', content: '' })

    // 尝试流式调用
    try {
      let receivedAnyChunk = false

      // 启动 15 秒超时保护
      streamTimeout = setTimeout(() => {
        if (!receivedAnyChunk) {
          console.warn('[AgentChat] 流式 15 秒无响应，回退非流式')
          // 如果还没有任何数据块，取消当前流并走回退
          _fallbackNonStream(text, aiIdx)
        }
      }, STREAM_TIMEOUT_MS)

      sendMessageStream(text, {
        onMessage: (chunk: string) => {
          receivedAnyChunk = true
          // 清除超时（有数据了就不需要超时保护了）
          if (streamTimeout) {
            clearTimeout(streamTimeout)
            streamTimeout = null
          }
          // 逐块追加到 AI 回复
          messages.value[aiIdx].content += chunk
        },

        onDone: () => {
          if (streamTimeout) {
            clearTimeout(streamTimeout)
            streamTimeout = null
          }
          isLoading.value = false
        },

        onError: (error: string) => {
          if (streamTimeout) {
            clearTimeout(streamTimeout)
            streamTimeout = null
          }
          isLoading.value = false
          // 如果 AI 消息仍然是空的，回退非流式
          if (!messages.value[aiIdx].content) {
            _fallbackNonStream(text, aiIdx)
          }
        },
      })
    } catch {
      // fetch 初始化失败（如网络不通），直接走非流式
      if (streamTimeout) {
        clearTimeout(streamTimeout)
        streamTimeout = null
      }
      await _fallbackNonStream(text, aiIdx)
    }
  }

  /**
   * 非流式回退：当流式不可用或超时时使用
   */
  async function _fallbackNonStream(text: string, aiIdx: number) {
    try {
      const resp = await apiSendMessage(text)
      messages.value[aiIdx].content = resp.response
    } catch (err: any) {
      // 友好的错误提示
      const errMsg = err instanceof Error ? err.message : '请求失败'
      messages.value[aiIdx].content =
        `😅 连接服务失败，请稍后重试。\n\n错误详情：${errMsg}`
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 清空聊天记录并重置会话
   */
  function clearMessages() {
    messages.value = []
    resetSession()
  }

  return {
    messages,
    isLoading,
    isConnected,
    sendMessage,
    clearMessages,
  }
}
