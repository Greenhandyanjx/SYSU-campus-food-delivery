// Simple singleton WebSocket client for app-wide chat notifications
import { getWsUrl } from '@/api/chat'

type MessageHandler = (msg: any) => void

class ChatClient {
  private ws: WebSocket | null = null
  private handlers: Set<MessageHandler> = new Set()
  private reconnectTimer: any = null

  connect() {
    const raw = (typeof window !== 'undefined' && localStorage.getItem('token')) || ''
    const pure = raw.replace(/^Bearer\s+/i, '')
    const url = getWsUrl() + `?token=${encodeURIComponent(pure)}`
    try {
      this.ws = new WebSocket(url)
    } catch (e) {
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      console.debug('chatClient connected')
      if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null }
    }

    this.ws.onmessage = (ev) => {
      try {
        const data = JSON.parse(ev.data)
        this.handlers.forEach(h => h(data))
      } catch (e) {
        // ignore
      }
    }

    this.ws.onclose = () => {
      console.debug('chatClient closed, will reconnect')
      this.scheduleReconnect()
    }

    this.ws.onerror = () => {
      // close will trigger reconnect
    }
  }

  scheduleReconnect() {
    if (this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, 3000)
  }

  send(payload: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(payload))
      return true
    }
    return false
  }

  onMessage(fn: MessageHandler) {
    this.handlers.add(fn)
  }

  offMessage(fn: MessageHandler) {
    this.handlers.delete(fn)
  }

  close() {
    this.ws?.close()
    this.ws = null
    if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null }
  }
}

const instance = new ChatClient()
export default instance
