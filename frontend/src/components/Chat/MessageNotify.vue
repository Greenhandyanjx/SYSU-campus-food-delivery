<template>
  <div />

  <!-- 全局订单弹窗（商家端） -->
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
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import chatClient from '@/utils/chatClient'
import { ElNotification } from 'element-plus'
import { getBaseUserDetail } from '@/api/chat'
import request from '@/api/merchant/request'
import { useChatStore } from '@/stores/chatStore'
import { orderAccept } from '@/api/merchant/order'

let currentMerchantId = null
let currentBaseUserId = null
const chatStore = useChatStore()
const orderNotify = ref(null)

async function handleIncoming(msg) {
  // Expect msg to include merchant_id and content and from_base_id
  console.log('[MessageNotify] incoming', msg)
  const from = msg.from_base_id || msg.fromBaseId
  // don't notify for messages that originate from this client
  if (from && currentBaseUserId && Number(from) === Number(currentBaseUserId)) {
    console.debug('[MessageNotify] ignoring self message')
    return
  }

  let title = '新消息'
  try {
    // If current client is a merchant, sender is likely a user -> fetch user name
    if (currentMerchantId) {
      const uid = msg.from_base_id || msg.user_base_id || msg.userBaseId
      if (uid) {
        const r = await getBaseUserDetail(uid)
        const u = r?.data?.data
        title = `来自 ${u?.username || u?.nickname || ('用户 ' + uid)}`
      } else if (msg.user_base_id) {
        title = `来自 用户 ${msg.user_base_id}`
      }
    } else {
      // current client is a regular user: sender likely merchant -> fetch merchant name
      const mid = msg.merchant_id || msg.merchantId
      if (mid) {
        try {
          const mr = await request({ url: '/merchant/detail', method: 'get', params: { id: mid } })
          const md = mr?.data?.data
          title = `来自 ${md?.shop_name || md?.shopName || ('商家 ' + mid)}`
        } catch (e) {
          title = `来自 商家 ${mid}`
        }
      }
    }
  } catch (e) {
    // fallback: use merchant id or user id
    if (msg.merchant_id) title = `来自 商家 ${msg.merchant_id}`
    else if (msg.user_base_id) title = `来自 用户 ${msg.user_base_id}`
  }

  const body = msg.content || '[非文本消息]'
  // 如果消息看起来像订单提醒，优先在商家端派发订单通知事件（不展示普通消息通知）
  const orderId = msg.orderId || msg.order_id || msg.orderIdStr || null
  const status = msg.status || msg.order_status || msg.event || null
  const looksLikeOrder = !!orderId
  if (looksLikeOrder) {
    // helper: attempt to ensure we know whether this client is a merchant
    const ensureMerchantDetected = async () => {
      if (currentMerchantId) return true
      if (localStorage.getItem('isMerchant') === '1') return true
      try {
        const base = await getBaseUserDetail()
        const bid = base?.data?.data?.id
        if (!bid) return false
        const r = await request({ url: '/merchant/detail', method: 'get', params: { base_id: bid } })
        const md = r?.data?.data
        if (md && md.id) {
          currentMerchantId = md.id
          try { localStorage.setItem('isMerchant', '1'); window.dispatchEvent(new Event('merchant:detected')) } catch (e) {}
          return true
        }
      } catch (e) {
        // ignore
      }
      return false
    }

    const isMerchantClient = await ensureMerchantDetected()
    const shouldNotifyOrder = isMerchantClient && (String(status || '').toLowerCase().includes('待接单') || String(status || '').toLowerCase().includes('waiting') || String(status || '').toLowerCase().includes('pending') || !status)
    if (shouldNotifyOrder) {
      try {
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
        const detail = { orderId, amount, pickupPoint, itemsText, itemCount, status, raw: msg }
        try { window.dispatchEvent(new CustomEvent('merchant:order:notify', { detail })) } catch(e) {}
        // 同时在此组件显示全局弹窗（以防原先负责弹窗的组件未挂载）
        try { orderNotify.value = detail } catch(e) {}
        return
      } catch(e) { console.warn('order notify dispatch failed', e) }
    }
    // else fallthrough to normal message notification
  }

  // 普通消息通知：先确保 Pinia 中存在该会话（方便 Navbar/会话列表显示）
  try {
    const uid = msg.user_base_id || msg.userBaseId || null
    const mid = msg.merchant_id || msg.merchantId || null
    if (currentMerchantId && uid) {
      chatStore.upsertSession(String(uid), { unread: (chatStore.sessions[String(uid)]?.unread || 0), meta: { userName: null, last_message: msg.content, last_at: msg.created_at } })
    } else if (!currentMerchantId && mid) {
      chatStore.upsertSession(String(mid), { unread: (chatStore.sessions[String(mid)]?.unread || 0), meta: { merchantName: null, last_message: msg.content, last_at: msg.created_at } })
    }

  } catch(e) { console.warn('MessageNotify upsertSession failed', e) }

  ElNotification({
    title,
    message: body,
    duration: 5000,
    onClick() {
      // dispatch global event to open chat UI
      const detail = {
        merchantId: msg.merchant_id || msg.merchantId || null,
        userBaseId: msg.user_base_id || msg.userBaseId || null,
      }
      window.dispatchEvent(new CustomEvent('chat:open', { detail }))

      // 点击通知时同时尝试标记会话为已读并同步后端 + 更新 Pinia
      ;(async () => {
        try {
          if (currentMerchantId) {
            const uid = msg.user_base_id || msg.userBaseId || null
            if (uid) {
              await request.post('/merchant/chats/mark_read', { merchant_id: Number(currentMerchantId), user_base_id: Number(uid) })
              try { window.dispatchEvent(new CustomEvent('merchant:chats:marked_read', { detail: { merchant_id: Number(currentMerchantId), user_base_id: Number(uid) } })) } catch(e) {}
              try { chatStore.markSessionRead(String(uid)) } catch(e) {}
            }
          } else {
            const mid = msg.merchant_id || msg.merchantId || null
            if (mid) {
              await request.post('/user/chats/mark_read', { merchant_id: Number(mid) })
              try { window.dispatchEvent(new CustomEvent('user:chats:marked_read', { detail: { merchant_id: Number(mid) } })) } catch(e) {}
              try { chatStore.markSessionRead(String(mid)) } catch(e) {}
            }
          }
        } catch (e) {
          console.warn('mark_read from notification click failed', e)
        }
      })()
    }
  })

  // 商家端不应自动打开会话，用户需点击通知或商家在会话列表中打开对应会话
}

onMounted(() => {
  chatClient.onMessage(handleIncoming)
  // ensure connection started
  console.log('[MessageNotify] connecting chatClient')
  chatClient.connect()

  // 试图确定当前用户是否为商家：获取 base user -> merchant by base_id
  getBaseUserDetail().then(res => {
    const base = res?.data?.data
    if (!base) return
    currentBaseUserId = base.id
    // 直接调用后端 /merchant/detail?base_id=xxx
    return request({ url: '/merchant/detail', method: 'get', params: { base_id: base.id } })
  }).catch(() => null).then(r => {
    if (!r) return
    if (r && r.data && r.data.data) {
      currentMerchantId = r.data.data.id
      console.log('[MessageNotify] detected merchant id =', currentMerchantId)
      try {
        localStorage.setItem('isMerchant', '1')
        window.dispatchEvent(new Event('merchant:detected'))
      } catch (e) {}
    }
  }).catch(() => {})
})

onBeforeUnmount(() => {
  chatClient.offMessage(handleIncoming)
})

async function acceptOrder(o) {
  if (!o || !o.orderId) return
  try {
    await orderAccept({ id: Number(o.orderId) || o.orderId })
    orderNotify.value = null
    try { alert('接单成功: ' + o.orderId) } catch (e) {}
    try { window.dispatchEvent(new CustomEvent('merchant:order:accepted', { detail: { orderId: o.orderId } })) } catch (e) {}
  } catch (e) {
    console.warn('acceptOrder failed', e)
    try { alert('接单失败，请重试') } catch (e) {}
  }
}

function viewOrderDetail(o) {
  if (!o || !o.orderId) return
  try {
    window.dispatchEvent(new CustomEvent('merchant:open_order_detail', { detail: { orderId: o.orderId } }))
  } catch (e) {
    try { window.open(`/merchant/order/${o.orderId}`, '_blank') } catch (e) {}
  }
}
</script>

<style scoped>
/* empty, purely functional component */
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
