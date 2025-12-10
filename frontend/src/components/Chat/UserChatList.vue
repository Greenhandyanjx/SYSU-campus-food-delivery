<template>
  <div class="user-chat-list">
    <div class="list-header">会话</div>
    <ul>
      <li v-for="c in chats" :key="c.merchant_id" @click="open(c)" :class="{ active: active === c.merchant_id }">
        <img class="avatar" :src="c.merchantAvatar || '/imgs/merchant.png'" />
        <div class="meta">
          <div class="top">
            <div class="name">{{ c.merchantName || ('商家 ' + c.merchant_id) }}</div>
            <div class="time">{{ formatTime(c.last_at) }}</div>
          </div>
          <div class="bottom">
            <div class="last">{{ c.last_message }}</div>
            <div v-if="getUnread(c.merchant_id)" class="badge">{{ getUnread(c.merchant_id) > 99 ? '99+' : getUnread(c.merchant_id) }}</div>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import request from '@/api/merchant/request'
import { getMerchantDetail, getBaseUserDetail } from '@/api/chat'
import chatClient from '@/utils/chatClient'
import { useChatStore } from '@/stores/chatStore'

// 保留原有样式与模板，但逻辑沿用商家端会话外壳（以 merchant 为 peer）
const chats = ref([])
const active = ref(null)
const isLoading = ref(false)
const currentBaseId = ref(null) // 当前登录的用户 id
const chatStore = useChatStore()

function getUnread(mid) {
  try {
    if (mid === null || typeof mid === 'undefined') return 0
    const v = chatStore.sessions[String(mid)]
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

const userNameCache = ref({})
async function fetchMerchantInfo(mid) {
  if (mid === null || typeof mid === 'undefined') return null
  if (userNameCache.value[mid]) return userNameCache.value[mid]
  try {
    const r = await getMerchantDetail(mid)
    const md = r?.data?.data
    const info = { name: md?.shop_name || md?.shopName || null, avatar: md?.logo || md?.logoUrl || null }
    userNameCache.value[mid] = info
    return info
  } catch (e) {
    return null
  }
}

let lastLoadAt = 0
const MIN_LOAD_INTERVAL = 5000
async function load() {
  const now = Date.now()
  if (isLoading.value) return
  if (now - lastLoadAt < MIN_LOAD_INTERVAL) return
  isLoading.value = true
  lastLoadAt = now
  try {
    const res = await request.get('/user/chats')
    if (res.data && Number(res.data.code) === 1) {
      let list = (res.data.data || []).map(c => ({
        merchant_id: c.merchant_id,
        merchantName: c.merchant_name || c.merchantName || null,
        merchantAvatar: c.merchant_avatar || c.merchantAvatar || null,
        last_message: c.last_message,
        last_at: c.last_at,
        unread_count: c.unread_count || 0,
        user_base_id: c.user_base_id || c.userBaseId || null
      }))

      // 异步补充商家展示信息（当后端未返回时）
      await Promise.all(list.map(async (c) => {
        if ((!c.merchantName || !c.merchantAvatar) && c.merchant_id !== null && typeof c.merchant_id !== 'undefined') {
          try {
            const info = await fetchMerchantInfo(c.merchant_id)
            if (info) {
              c.merchantName = c.merchantName || info.name
              c.merchantAvatar = c.merchantAvatar || info.avatar
            }
          } catch (e) { }
        }
      }))

      chats.value = list

      // 初始化 Pinia 会话状态（key 使用 merchant_id）
      try {
        list.forEach(c => {
          if (c && c.merchant_id !== undefined && c.merchant_id !== null) {
            chatStore.upsertSession(String(c.merchant_id), { unread: c.unread_count || 0, meta: { merchantName: c.merchantName, merchantAvatar: c.merchantAvatar } })
          }
        })
      } catch (e) {}
    } else {
      chats.value = []
    }
  } catch (e) {
    console.warn('load user chats failed', e)
    chats.value = []
  } finally {
    isLoading.value = false
  }
}

async function open(c) {
  // 使用订单页内的发起聊天流程：优先使用 numeric merchant_id，补充商家和用户信息后再打开聊天窗口
  let midCandidate = (c && (c.merchant_id || c.merchantId || c.storeId || c.store_id))
  let mid = null
  if (midCandidate === 0 || midCandidate) {
    const n = Number(midCandidate)
    if (!Number.isNaN(n)) mid = n
  }

  // 尝试从会话项里直接获取 user_base_id（如果后端返回）
  let baseUserId = (c && (c.user_base_id || c.userBaseId)) || null

  if (mid === null) {
    // 回退：无法解析 merchant id，提示并返回
    console.warn('[UserChatList.open] cannot resolve merchant id from item', c)
    alert('无法定位商家 ID，无法发起聊天')
    return
  }

  active.value = mid

  // 标记本地 session 已读（乐观）
  try { chatStore.markSessionRead(String(mid)) } catch (e) {}

  // 后端标记已读并刷新列表
  try {
    await request.post('/user/chats/mark_read', { merchant_id: mid })
  } catch (e) { console.warn('mark read failed', e) }

  // 获取当前登录用户 id（用于 chat window payload）
  try {
    const cur = await getBaseUserDetail()
    if (cur && cur.data && cur.data.data) baseUserId = cur.data.data.id
  } catch (e) { console.warn('getBaseUserDetail failed', e) }

  // 补充商家信息以便 ChatWindow 里能直接展示名称/头像
  let merchantInfo = null
  try {
    const m = await getMerchantDetail(mid)
    merchantInfo = m && m.data && m.data.data
  } catch (e) { }

  const detail = {
    merchantId: mid,
    merchant_id: mid,
    userBaseId: baseUserId,
    user_base_id: baseUserId,
    merchantName: merchantInfo?.shop_name || merchantInfo?.shopName || null,
    merchantAvatar: merchantInfo?.logo || merchantInfo?.logoUrl || null
  }

  try {
    console.log('[UserChatList.open] dispatching chat:open', detail)
    window.dispatchEvent(new CustomEvent('chat:open', { detail }))
  } catch (e) { console.warn('[UserChatList.open] dispatch failed', e) }

  try { window.dispatchEvent(new CustomEvent('user:chats:marked_read', { detail: { merchant_id: mid } })) } catch (e) {}

  // 最后刷新本地列表（异步）
  try { await load() } catch (e) { console.warn('reload after open failed', e) }
}

function updateChatFromMsg(msg) {
  try {
    const mid = msg.merchant_id || msg.merchantId
    const uid = msg.user_base_id || msg.userBaseId
    const from = msg.from_base_id || msg.fromBaseId
    if (typeof uid === 'undefined' || uid === null) return
    // 仅处理发给当前登录用户的消息
    if (currentBaseId.value && Number(uid) !== Number(currentBaseId.value)) return

    let idx = chats.value.findIndex(c => Number(c.merchant_id) === Number(mid))
    let c
    if (idx >= 0) c = chats.value[idx]
    else {
      c = { merchant_id: mid, last_message: '', last_at: new Date().toISOString(), unread_count: 0, user_base_id: uid, merchantName: null, merchantAvatar: null }
      chats.value.unshift(c)
      idx = 0
    }

    c.last_message = msg.content || c.last_message
    c.last_at = msg.created_at || msg.last_at || new Date().toISOString()
    // 如果消息来自对方（非当前用户）则计为未读
    if (from && currentBaseId.value && Number(from) !== Number(currentBaseId.value)) {
      c.unread_count = (Number(c.unread_count) || 0) + 1
    }

    // 补充商家名
    if (!c.merchantName && c.merchant_id) fetchMerchantInfo(c.merchant_id).then(info => { if (info) { c.merchantName = info.name; c.merchantAvatar = info.avatar } })

    try {
      const isSelf = from && currentBaseId.value && Number(from) === Number(currentBaseId.value)
      chatStore.addMessage(String(mid), msg, !!isSelf)
    } catch (e) {}

    if (idx > 0) {
      chats.value.splice(idx, 1)
      chats.value.unshift(c)
    }
  } catch (e) {
    console.warn('updateChatFromMsg failed', e)
  }
}

onMounted(() => {
  ;(async () => {
    try {
      const cur = await getBaseUserDetail()
      if (cur && cur.data && cur.data.data) currentBaseId.value = cur.data.data.id
    } catch (e) {}
    await load()
  })()

  const wrappedHandler = (msg) => {
    try {
      // 只处理发给当前用户的消息
      const uid = msg.user_base_id || msg.userBaseId
      if (!uid) return
      if (currentBaseId.value && Number(uid) === Number(currentBaseId.value)) {
        // 防抖合并短时间内多次通知
        if (!window.__userChatListRefreshTimer) {
          window.__userChatListRefreshTimer = setTimeout(() => { window.__userChatListRefreshTimer = null; load() }, 600)
        }
        updateChatFromMsg(msg)
      }
    } catch (e) {}
  }

  chatClient.onMessage(wrappedHandler)
  try { chatClient.connect() } catch (e) {}

  const userMarkReadHandler = (ev) => {
    try {
      const d = (ev && ev.detail) || {}
      const mid = d.merchant_id || d.merchantId || null
      if (mid !== null && typeof mid !== 'undefined') load()
    } catch (e) { console.warn('userMarkReadHandler error', e) }
  }
  window.addEventListener('user:chats:marked_read', userMarkReadHandler)

  onBeforeUnmount(() => {
    chatClient.offMessage(wrappedHandler)
    try { if (window.__userChatListRefreshTimer) { clearTimeout(window.__userChatListRefreshTimer); window.__userChatListRefreshTimer = null } } catch(e) {}
    window.removeEventListener('user:chats:marked_read', userMarkReadHandler)
  })
})
</script>
<style scoped>
.user-chat-list {
  width: 100%;
  height: 100%;
  background: #fff;
  display: flex;
  flex-direction: column;
  border-radius: 12px;           /* 和外层卡片保持一致 */
  overflow: hidden;
}

/* 顶部标题栏 - 完全微信风格 */
.list-header {
  height: 50px;
  line-height: 50px;
  padding: 0 20px;
  font-size: 17px;
  font-weight: 600;
  color: #000;
  background: #f7f7f8;
  border-bottom: 0.5px solid #e5e5e5;
  flex-shrink: 0;
}

/* 列表区域 */
ul {
  list-style: none;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow-y: auto;
  /* 隐藏滚动条但仍可滚动（微信同款） */
  -ms-overflow-style: none;
  scrollbar-width: none;
}
ul::-webkit-scrollbar { display: none; }

/* 每一行 */
li {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  gap: 12px;
  cursor: pointer;
  position: relative;
  transition: background 0.2s;
}
li:hover,
li.active {
  background: #f0f0f0;           /* 微信选中/悬停色 */
}
li:not(:last-child)::after {
  content: '';
  position: absolute;
  left: 72px;
  right: 0;
  bottom: 0;
  height: 0.5px;
  background: #e5e5e5;
}

/* 头像 - 微信 52×52 圆角8px */
.avatar {
  width: 52px;
  height: 52px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}

/* 文字区域 */
.meta {
  flex: 1;
  min-width: 0;                  /* 重要：让文字能被截断 */
  display: flex;
  flex-direction: column;
  justify-content: center;
  height: 52px;
}
.top {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 6px;
}
.name {
  font-size: 17px;
  font-weight: 600;
  color: #000;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.time {
  font-size: 12px;
  color: #999;
}

/* 最后一条消息 + 未读 */
.bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.last {
  font-size: 14px;
  color: #888;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

/* 未读红点 - 完美还原微信 */
.badge {
  min-width: 20px;
  height: 20px;
  line-height: 20px;
  text-align: center;
  background: #f56c6c;
  color: #fff;
  font-size: 12px;
  border-radius: 10px;
  padding: 0 6px;
  flex-shrink: 0;
}
/* 未读数大于99时显示“99+” */
.badge:where([class*="9"], [class*="0"]) {
  padding: 0 4px;
}
</style>
