<template>
  <div class="merchant-chat-list">
    <div class="list-header">会话列表</div>
    <ul>
      <li v-for="c in chats" :key="c.user_base_id" @click="open(c)" :class="{ active: active === c.user_base_id }">
        <div class="left">
          <div class="name">{{ c.userName || ('用户 ' + c.user_base_id) }}</div>
          <div class="last">{{ c.last_message }}</div>
        </div>
        <div class="right">
          <span class="time">{{ formatTime(c.last_at) }}</span>
          <span v-if="getUnreadForUser(c.user_base_id)" class="badge">{{ getUnreadForUser(c.user_base_id) > 99 ? '99+' : getUnreadForUser(c.user_base_id) }}</span>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import request from '@/api/merchant/request'
import { getBaseUserDetail } from '@/api/chat'
import chatClient from '@/utils/chatClient'
import { useChatStore } from '@/stores/chatStore'

const chats = ref([])
const active = ref(null)
const isLoading = ref(false) // 添加加载标志，防止重复加载
const currentMerchantId = ref(null) // 缓存当前商家ID
const userNameCache = ref({}) // 缓存用户名称，避免重复请求
const chatStore = useChatStore()

function getUnreadForUser(uid) {
  try {
    if (!uid) return 0
    const v = chatStore.sessions[String(uid)]
    return (v && Number(v.unread)) ? Number(v.unread) : 0
  } catch (e) { return 0 }
}

function formatTime(s) {
  if (!s) return ''
  const dt = new Date(s)
  const now = new Date()
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const startOfYesterday = new Date(startOfToday.getTime() - 24 * 3600 * 1000)
  const pad = (n) => String(n).padStart(2, '0')
  const timePart = `${pad(dt.getHours())}:${pad(dt.getMinutes())}`
  if (dt >= startOfToday) return `今天 ${timePart}`
  if (dt >= startOfYesterday) return `昨天 ${timePart}`
  if (dt.getFullYear() === now.getFullYear()) return `${dt.getMonth() + 1}月${dt.getDate()}日 ${timePart}`
  return `${dt.getFullYear()}年${dt.getMonth() + 1}月${dt.getDate()}日 ${timePart}`
}

async function fetchUserName(userBaseId) {
  if (userNameCache.value[userBaseId]) return userNameCache.value[userBaseId]
  try {
    const r = await getBaseUserDetail(userBaseId)
    const u = r?.data?.data
    const name = u?.username || u?.nickname || null
    userNameCache.value[userBaseId] = name
    return name
  } catch (e) {
    console.warn(`fetchUserName failed for ${userBaseId}`, e)
    return null
  }
}

let lastLoadAt = 0
const MIN_LOAD_INTERVAL = 5000 // ms
async function load() {
  const now = Date.now()
  if (isLoading.value) return // 防止重复加载
  if (now - lastLoadAt < MIN_LOAD_INTERVAL) return // 节流：最小间隔
  isLoading.value = true
  lastLoadAt = now
  try {
    const res = await request.get('/merchant/chats')
    if (res.data && Number(res.data.code) === 1) {
      let list = res.data.data || []
      // 使用缓存批量获取用户名
      await Promise.all(list.map(async (c) => {
        c.userName = await fetchUserName(c.user_base_id)
      }))
      // 如果为空，提供mock数据仅在开发模式
      if (process.env.NODE_ENV === 'development' && (!list || list.length === 0)) {
        const now = Date.now()
        list = [
          { user_base_id: 501, last_message: '您好，有什么可以帮到您？', last_at: new Date(now-1800*1000).toISOString(), unread_count: 3, merchant_id: null, userName: '张三' },
          { user_base_id: 502, last_message: '请问订单可以合并吗？', last_at: new Date(now-3600*24*2).toISOString(), unread_count: 0, merchant_id: null, userName: '李四' },
        ]
      }
      chats.value = list
      // 初始化 Pinia 会话状态，key 使用 user_base_id 作为 peerId
      try {
        list.forEach(c => {
          if (c && c.user_base_id !== undefined && c.user_base_id !== null) {
            chatStore.upsertSession(String(c.user_base_id), { unread: c.unread_count || 0, meta: { userName: c.userName } })
          }
        })
      } catch (e) {}
    }
  } catch (e) {
    console.log('load chats failed', e)
    // 开发模式下fallback mock
    if (process.env.NODE_ENV === 'development') {
      const now = Date.now()
      chats.value = [
        { user_base_id: 601, last_message: '示例：请尽快处理订单', last_at: new Date(now-500000).toISOString(), unread_count: 1, merchant_id: null, userName: '示例用户' }
      ]
    }
  } finally {
    isLoading.value = false
  }
}

async function loadCurrentMerchantId() {
  if (currentMerchantId.value) return // 已缓存，直接返回
  try {
    const base = await getBaseUserDetail()
    const bid = base?.data?.data?.id
    if (bid) {
      const r = await request({ url: '/merchant/detail', method: 'get', params: { base_id: bid } })
      const md = r?.data?.data
      if (md) currentMerchantId.value = md.id
    }
  } catch (e) {
    console.warn('loadCurrentMerchantId failed', e)
  }
}

function open(c) {
  active.value = c.user_base_id
  try {
    const detail = { merchantId: c.merchant_id || null, merchant_id: c.merchant_id || null, userBaseId: c.user_base_id, user_base_id: c.user_base_id }
    console.log('[MerchantChatList.open] dispatching chat:open', detail)
    window.dispatchEvent(new CustomEvent('chat:open', { detail }))
  } catch (e) { console.warn('[MerchantChatList.open] dispatch failed', e) }
}

function updateChatFromMsg(msg) {
    try {
    const mid = msg.merchant_id || msg.merchantId
    const uid = msg.user_base_id || msg.userBaseId
    const from = msg.from_base_id || msg.fromBaseId
    if (!mid || !uid) return
    // 如果消息不是针对本商家则忽略
    if (currentMerchantId.value && Number(mid) !== Number(currentMerchantId.value)) return

    // 查找现有会话
    let idx = chats.value.findIndex(c => Number(c.user_base_id) === Number(uid))
    let c
    if (idx >= 0) {
      c = chats.value[idx]
    } else {
      // 创建新会话
      c = {
        user_base_id: uid,
        last_message: '',
        last_at: new Date().toISOString(),
        unread_count: 0,
        merchant_id: mid,
        userName: null
      }
      chats.value.unshift(c)
      idx = 0 // 新插入首位
    }

    c.last_message = msg.content || c.last_message
    c.last_at = msg.created_at || msg.last_at || new Date().toISOString()
    // 如果消息是来自用户（非本商家）则增加未读计数
    if (from && currentMerchantId.value && Number(from) !== Number(currentMerchantId.value)) {
      c.unread_count = (Number(c.unread_count) || 0) + 1
    }
    // 异步获取用户名如果缺失
    if (!c.userName) {
      fetchUserName(uid).then(name => { if (name) c.userName = name })
    }
    // 把消息写入 Pinia，会根据 isSelf 决定是否增加未读
    try {
      const isSelf = from && currentMerchantId.value && Number(from) === Number(currentMerchantId.value)
      chatStore.addMessage(String(uid), msg, !!isSelf)
    } catch (e) {}
    // 把该会话移动到列表顶部（如果不是新创建）
    if (idx > 0) {
      chats.value.splice(idx, 1)
      chats.value.unshift(c)
    }
  } catch (e) {
    console.warn('updateChatFromMsg failed', e)
  }
}

onMounted(async () => {
  // 先加载当前商家ID（只一次）
  await loadCurrentMerchantId()
  // 初始加载chats（只一次）
  await load()

  const wrappedHandler = (msg) => {
    updateChatFromMsg(msg)
  }

  chatClient.onMessage(wrappedHandler)
  try { chatClient.connect() } catch (e) {}

  // 监听已读事件，及时更新本地未读数
  const markReadHandler = (ev) => {
    try {
      const d = (ev && ev.detail) || {}
      const uid = d.user_base_id || d.userBaseId || d.userId || null
      if (!uid) return
      const idx = chats.value.findIndex(c => Number(c.user_base_id) === Number(uid))
      if (idx >= 0) chats.value[idx].unread_count = 0
      // 同步 Pinia 状态
      try { chatStore.markSessionRead(String(uid)) } catch (e) {}
    } catch (e) {
      console.warn('markReadHandler error', e)
    }
  }
  window.addEventListener('merchant:chats:marked_read', markReadHandler)

  onBeforeUnmount(() => {
    chatClient.offMessage(wrappedHandler)
    window.removeEventListener('merchant:chats:marked_read', markReadHandler)
  })
})
</script>

<style scoped>
.merchant-chat-list { 
  border: 1px solid #eee; 
  border-radius: 8px; 
  padding: 8px; 
  width: 320px; 
  background: #fff 
}
.merchant-chat-list .list-header { 
  font-weight: 700; 
  margin-bottom: 8px 
}
.merchant-chat-list ul { 
  list-style: none; 
  padding: 0; 
  margin: 0 
}
.merchant-chat-list li { 
  display: flex; 
  justify-content: space-between; 
  padding: 8px; 
  cursor: pointer; 
  border-radius: 6px 
}
.merchant-chat-list li.active { 
  background: #f6f8fa 
}
.merchant-chat-list .left { 
  max-width: 220px 
}
.merchant-chat-list .name { 
  font-weight: 600 
}
.merchant-chat-list .last { 
  color: #666; 
  font-size: 13px; 
  white-space: nowrap; 
  overflow: hidden; 
  text-overflow: ellipsis 
}
.merchant-chat-list .badge { 
  background: #f56c6c; 
  color: #fff; 
  padding: 2px 6px; 
  border-radius: 12px; 
  font-size: 12px 
}
.merchant-chat-list .time { 
  font-size: 12px; 
  color: #999; 
  margin-right: 8px 
}
</style>