import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useChatStore = defineStore('chat', () => {
  // sessions: { [peerId: string]: { messages: Array, unread: number, meta: object } }
  const sessions = ref({} as Record<string, { messages: any[]; unread: number; meta: any }>)

  function ensureSession(peerId: string) {
    if (!peerId) return
    if (!sessions.value[peerId]) {
      sessions.value[peerId] = { messages: [], unread: 0, meta: {} }
    }
  }

  function addMessage(peerId: string, msg: any, isSelf = false) {
    if (!peerId) return
    ensureSession(peerId)
    try {
      sessions.value[peerId].messages.push(msg)
      // 如果不是自己发的且消息未标记为已读，则增加未读
      if (!isSelf && !(msg && msg.isRead)) {
        sessions.value[peerId].unread = (sessions.value[peerId].unread || 0) + 1
      }
    } catch (e) {
      console.warn('chatStore.addMessage error', e)
    }
  }

  function markSessionRead(peerId: string) {
    if (!peerId) return
    if (!sessions.value[peerId]) return
    try {
      sessions.value[peerId].unread = 0
      sessions.value[peerId].messages.forEach(m => { if (m) m.isRead = true })
    } catch (e) {
      console.warn('chatStore.markSessionRead error', e)
    }
  }

  function upsertSession(peerId: string, payload: { unread?: number; meta?: any } = {}) {
    if (!peerId) return
    ensureSession(peerId)
    if (payload.unread !== undefined) sessions.value[peerId].unread = Number(payload.unread) || 0
    if (payload.meta) sessions.value[peerId].meta = { ...sessions.value[peerId].meta, ...payload.meta }
  }

  return {
    sessions,
    addMessage,
    markSessionRead,
    upsertSession,
  }
})
