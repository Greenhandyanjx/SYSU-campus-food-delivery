/**
 * Agent API 模块
 * =========================================================
 * 封装与 JD-Agent API 后端的所有通信。
 *
 * 支持两种模式：
 *   1. sendMessage()       — 非流式调用，返回完整 JSON 响应
 *   2. sendMessageStream() — 流式调用，返回 EventSource 风格的接口
 *
 * JD-Agent API 地址通过环境变量 VITE_AGENT_API_URL 配置，
 * 默认值为 http://127.0.0.1:8000
 */

const AGENT_API_BASE = import.meta.env.VITE_AGENT_API_URL || 'http://127.0.0.1:8000'
const CHAT_ENDPOINT = `${AGENT_API_BASE}/api/v1/chat`

// 会话 ID（首次调用时自动生成）
let _sessionId: string | null = null

/**
 * 获取或生成会话 ID
 * 生命周期与当前页面一致，刷新后重新生成
 */
function getSessionId(): string {
  if (!_sessionId) {
    // 生成一个简短的随机会话 ID
    _sessionId = Math.random().toString(36).substring(2, 8)
  }
  return _sessionId
}

/**
 * 从 localStorage 获取 JWT token，并构造 Authorization 请求头
 */
function getAuthHeaders(): Record<string, string> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  try {
    const token = localStorage.getItem('token')
    if (token) {
      headers['Authorization'] = token.startsWith('Bearer ') ? token : `Bearer ${token}`
    }
  } catch {
    // localStorage 不可用时（SSR等），不传 auth
  }
  return headers
}

/**
 * ChatMessage 接口
 * 表示聊天中的一条消息
 */
export interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
}

/**
 * ChatResponse 接口
 * 非流式聊天的响应格式
 */
export interface ChatResponse {
  response: string
  session_id: string
  tools_used: string[]
}

/**
 * StreamCallback 接口
 * 流式回调函数类型
 */
export interface StreamCallbacks {
  /** 每收到一段文本时触发 */
  onMessage: (text: string) => void
  /** 流结束时触发 */
  onDone: () => void
  /** 发生错误时触发 */
  onError: (error: string) => void
}

/**
 * 非流式发送消息
 * =================
 * 发送用户消息到 Agent API，返回完整的回复文本。
 *
 * @param message  - 用户消息内容
 * @param stream   - 是否流式（默认 false）
 * @returns         - 完整的 ChatResponse 响应
 */
export async function sendMessage(message: string): Promise<ChatResponse> {
  const body = {
    message,
    session_id: getSessionId(),
    stream: false,
  }

  const response = await fetch(CHAT_ENDPOINT, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(body),
  })

  if (!response.ok) {
    let errorMsg = `请求失败 (HTTP ${response.status})`
    try {
      const errData = await response.json()
      errorMsg = errData.detail || errorMsg
    } catch {
      // ignore parse errors
    }
    throw new Error(errorMsg)
  }

  const data: ChatResponse = await response.json()
  return data
}

/**
 * 流式发送消息
 * =================
 * 使用 fetch + ReadableStream 实现 SSE 流式解析。
 *
 * 调用方式：
 * ```typescript
 * sendMessageStream('你好', {
 *   onMessage: (text) => console.log('收到:', text),
 *   onDone: () => console.log('完成'),
 *   onError: (err) => console.error('错误:', err),
 * })
 * ```
 *
 * 如果浏览器不支持 ReadableStream，会降级到非流式模式。
 *
 * @param message   - 用户消息内容
 * @param callbacks - 流式回调
 * @returns         - AbortController，可用于取消请求
 */
export function sendMessageStream(
  message: string,
  callbacks: StreamCallbacks,
): AbortController {
  const controller = new AbortController()

  const body = {
    message,
    session_id: getSessionId(),
    stream: true,
  }

  // 启动异步请求（不 await，让调用方通过回调接收结果）
  _doStreamFetch(body, callbacks, controller.signal)

  return controller
}

/**
 * 内部：执行流式 fetch 并解析 SSE 数据
 */
async function _doStreamFetch(
  body: Record<string, unknown>,
  callbacks: StreamCallbacks,
  signal: AbortSignal,
): Promise<void> {
  const { onMessage, onDone, onError } = callbacks

  try {
    const response = await fetch(CHAT_ENDPOINT, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(body),
      signal,
    })

    if (!response.ok) {
      let errorMsg = `请求失败 (HTTP ${response.status})`
      try {
        const errData = await response.json()
        errorMsg = errData.detail || errorMsg
      } catch {
        // ignore
      }
      onError(errorMsg)
      return
    }

    // 检查响应体是否可读
    const contentType = response.headers.get('content-type') || ''
    if (!response.body) {
      // 降级：浏览器不支持 ReadableStream，走非流式
      const data = await response.json()
      const text = data.response || JSON.stringify(data)
      onMessage(text)
      onDone()
      return
    }

    // 使用 ReadableStream 解析 SSE 数据
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      // 按行分割 buffer
      const lines = buffer.split('\n')
      // 保留最后一个可能不完整的行
      buffer = lines.pop() || ''

      for (const line of lines) {
        const trimmed = line.trim()

        // SSE 格式: data: {"text": "..."}
        if (trimmed.startsWith('data: ')) {
          const jsonStr = trimmed.slice(6).trim()

          // 结束标记
          if (jsonStr === '[DONE]') {
            onDone()
            continue
          }

          try {
            const parsed = JSON.parse(jsonStr)
            if (parsed.text) {
              onMessage(parsed.text)
            } else if (parsed.error) {
              onError(parsed.error)
            }
          } catch {
            // 非 JSON 格式的 data，直接作为文本
            if (jsonStr) {
              onMessage(jsonStr)
            }
          }
        }
      }
    }

    // 处理 buffer 中残留的数据
    if (buffer.trim()) {
      const trimmed = buffer.trim()
      if (trimmed.startsWith('data: ')) {
        const jsonStr = trimmed.slice(6).trim()
        if (jsonStr !== '[DONE]') {
          try {
            const parsed = JSON.parse(jsonStr)
            if (parsed.text) onMessage(parsed.text)
          } catch {
            if (jsonStr) onMessage(jsonStr)
          }
        }
      }
    }

    onDone()
  } catch (err: unknown) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      // 用户主动取消，不触发 onError
      return
    }
    const errorMsg = err instanceof Error ? err.message : '未知错误'
    onError(errorMsg)
  }
}

/**
 * 重置会话 ID
 * 当用户希望开始新对话时调用
 */
export function resetSession(): void {
  _sessionId = null
}
