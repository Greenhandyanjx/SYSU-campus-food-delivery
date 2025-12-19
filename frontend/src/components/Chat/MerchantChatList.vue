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
          <span class="time">{{ formatDateToCN(c.last_at) }}</span>
          <span v-if="getUnreadForUser(c.user_base_id)" class="badge">{{ getUnreadForUser(c.user_base_id) > 99 ? '99+' : getUnreadForUser(c.user_base_id) }}</span>
        </div>
      </li>
    </ul>
    <!-- 订单通知弹窗（由 WebSocket 推送触发） -->
    <div v-if="orderNotify" class="order-notify">
      <div class="order-notify-card">
        <div class="order-header">订单提醒 · #{{ orderNotify.orderId }}</div>
        <div class="order-body">
          <div class="row">{{ orderNotify.pickupPoint || '取餐点未知' }} ｜ {{ orderNotify.amount ? '￥' + orderNotify.amount : '' }} ｜ 共 {{ orderNotify.itemCount || 0 }} 件</div>
          <div class="row">商品：{{ orderNotify.itemsText || '（详情稍后加载）' }}</div>
          <div class="row">状态：{{ orderNotify.status || '待接单' }}</div>
        </div>
        <div class="order-actions">
          <button class="btn accept" @click="acceptOrder(orderNotify)">接单</button>
          <button class="btn view" @click="viewOrderDetail(orderNotify)">查看详情</button>
        </div>
      </div>
    </div>
  </div>

  <!-- 本地商家聊天弹窗（替代全局 chat:open，商家使用页面内弹窗） -->
  <teleport to="body">
    <div v-if="showChat" class="chat-overlay" @click.self="closeLocalChat">
      <div class="chat-modal">
          <button class="close-btn" @click="closeLocalChat">✕</button>
          <div class="chat-body">
            <MerchantChatWindow
              :merchantId="merchantIdLocal"
              :merchantName="merchantNameLocal"
              :merchantAvatar="merchantAvatarLocal"
              :userBaseId="active"
              @close="closeLocalChat"
            />
        </div>
      </div>
    </div>
  </teleport>
</template>
<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import ChatWindow from '@/components/Chat/ChatWindow.vue'
import MerchantChatWindow from '@/components/Chat/MerchantChatWindow.vue'
import { useRouter } from 'vue-router'
import noImg from '@/assets/noImg.png'
import request from '@/api/merchant/request'
import { orderAccept } from '@/api/merchant/order'
import { getBaseUserDetail } from '@/api/chat'
import chatClient from '@/utils/chatClient'
import { useChatStore } from '@/stores/chatStore'
import { emitOrderChanged } from '@/utils/orderEvents'

const chats = ref([])
const active = ref(null)
const showChat = ref(false)
const merchantIdLocal = ref(null)
const merchantNameLocal = ref('')
const merchantAvatarLocal = ref('')
const isLoading = ref(false) // 添加加载标志，防止重复加载
const currentMerchantId = ref(null) // 缓存当前商家ID
const orderNotify = ref(null) // { orderId, amount, pickupPoint, itemsText, itemCount, status, raw }
const userNameCache = ref({}) // 缓存用户名称，避免重复请求
const chatStore = useChatStore()
const router = useRouter()

function getUnreadForUser(uid) {
  try {
    if (!uid) return 0
    const v = chatStore.sessions[String(uid)]
    return (v && Number(v.unread)) ? Number(v.unread) : 0
  } catch (e) { return 0 }
}

function formatDateToCN(s) {
  if (!s) return ''
  const dt = new Date(s)
  if (isNaN(dt.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  const yyyy = dt.getFullYear()
  const mm = pad(dt.getMonth() + 1)
  const dd = pad(dt.getDate())
  const HH = pad(dt.getHours())
  const MM = pad(dt.getMinutes())
  return `${yyyy}年${mm}月${dd}日 ${HH}:${MM}`
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

async function open(c) {
  active.value = c.user_base_id
  try {
    // 打开本地弹窗，而不是触发全局 chat:open（全局由用户端使用）
    merchantIdLocal.value = c.merchant_id || null
    merchantNameLocal.value = c.userName || ''
    merchantAvatarLocal.value = ''
    // 尝试补全商家信息（若当前页面能拿到）
    try {
      if (!merchantNameLocal.value && merchantIdLocal.value) {
        const md = await import('@/api/chat').then(m => m.getMerchantDetail(merchantIdLocal.value))
        const data = md && md.data && (md.data.data || md.data)
        if (data) {
          merchantNameLocal.value = data.shop_name || merchantNameLocal.value
          merchantAvatarLocal.value = data.logo || merchantAvatarLocal.value
        }
      }
    } catch (e) {}
    showChat.value = true
  } catch (e) { console.warn('[MerchantChatList.open] open failed', e) }
}

function closeLocalChat() {
  showChat.value = false
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
    // 保证去重
    try { dedupeChats() } catch(e) {}
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
    // 如果是订单推送（只包含最必要字段）则展示订单弹窗
    try {
      const orderId = msg.orderId || msg.order_id || msg.orderIdStr || null
      const status = msg.status || msg.order_status || msg.event || null
      if (orderId) {
        // 只在订单进入待接单时提醒商家（后端可发送精简消息）
        const shortStatus = String(status || '').toLowerCase()
        const shouldNotify = shortStatus.includes('待接单') || shortStatus.includes('waiting') || shortStatus.includes('pending') || !status
        if (shouldNotify) {
          const amount = msg.amount || msg.total_amount || msg.price || null
          const pickupPoint = msg.pickupPoint || msg.pickup_point || msg.pickup || (msg.pickup_info && msg.pickup_info.name) || ''
          const items = msg.items || msg.goods || msg.dishes || []
          let itemsText = ''
          let itemCount = 0
          try {
            if (Array.isArray(items)) {
              itemCount = items.reduce((s, it) => s + (it.count || it.qty || it.quantity || 1), 0)
              itemsText = items.slice(0,3).map(it => `${it.name || it.title || it.dish_name} ×${it.count || it.qty || it.quantity || 1}`).join('，')
            }
          } catch(e) {}
          orderNotify.value = { orderId, amount, pickupPoint, itemsText, itemCount, status, raw: msg }
        }
      }
    } catch (e) { console.warn('order notify parse failed', e) }
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

  // 监听来自全局通知的订单提醒
  const orderNotifyHandler = (ev) => {
    try {
      const d = (ev && ev.detail) || {}
      if (!d) return
      orderNotify.value = d
    } catch (e) { console.warn('orderNotifyHandler error', e) }
  }
  window.addEventListener('merchant:order:notify', orderNotifyHandler)

  onBeforeUnmount(() => {
    chatClient.offMessage(wrappedHandler)
    window.removeEventListener('merchant:chats:marked_read', markReadHandler)
    window.removeEventListener('merchant:order:notify', orderNotifyHandler)
  })
})

// 去重 helper：确保同一 user_base_id 只有一条会话
function dedupeChats() {
  const seen = new Set()
  const out = []
  for (const c of chats.value) {
    const key = String(c.user_base_id || c.userBaseId || c.id || '')
    if (!key) continue
    if (!seen.has(key)) {
      seen.add(key)
      out.push(c)
    }
  }
  chats.value = out
}

// 在 load 与消息更新后确保去重
const originalLoad = load
load = async function() {
  await originalLoad()
  dedupeChats()
}

// order notification handlers
async function acceptOrder(o) {
  if (!o || !o.orderId) return
  try {
    // 使用与订单管理一致的接单 API（POST /merchant/order/accept），统一参数为 { id }
    await orderAccept({ id: Number(o.orderId) })
    // 接单成功后关闭通知并可跳转详情
    orderNotify.value = null
    try { alert('接单成功: ' + o.orderId) } catch(e){}
    try { emitOrderChanged({ orderId: o.orderId }) } catch (e) {}
  } catch (e) {
    console.warn('acceptOrder failed', e)
    try { alert('接单失败，请重试') } catch(e){}
  }
}

function viewOrderDetail(o) {
  if (!o || !o.orderId) return
  try {
    const ev = new CustomEvent('merchant:open_order_detail', { detail: { orderId: o.orderId } })
    window.dispatchEvent(ev)
  } catch (e) {
    // fallback to route if dispatching event fails
    try { router.push({ path: `/merchant/order/${o.orderId}` }) } catch (e) { window.open(`/merchant/order/${o.orderId}`, '_blank') }
  }
}

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
<style scoped>
.order-notify {
  position: fixed;
  right: 20px;
  bottom: 20px;
  z-index: 2147483647;
}
.order-notify-card {
  width: 320px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  padding: 12px;
}
.order-header { font-weight: 700; margin-bottom: 8px }
.order-body .row { margin-bottom: 6px; color: #333; font-size: 13px }
.order-actions { display:flex; gap:8px; justify-content:flex-end }
.btn { padding: 8px 12px; border-radius:6px; border: none; cursor:pointer }
.btn.accept { background: #4caf50; color: #fff }
.btn.view { background:#fff; border:1px solid #ddd }
</style>
<style scoped>
/* 居中模态样式（用于本地 ChatWindow 弹窗） */
.chat-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2147483647;
}
.chat-modal { width: 400px; max-width: 92%; background: #fff; border-radius: 10px; overflow: hidden; display:flex; flex-direction:column; z-index:2147483648; position:relative }
.chat-body { width:100%; height: 700px; }
.chat-modal .close-btn {
  position: absolute;
  right: 8px;
  top: 8px;
  border: none;
  background: transparent;
  font-size: 16px;
  cursor: pointer;
}
</style>