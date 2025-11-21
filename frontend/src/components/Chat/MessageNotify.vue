<template>
  <div />
</template>

<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import chatClient from '@/utils/chatClient'
import { ElNotification } from 'element-plus'
import { getBaseUserDetail } from '@/api/chat'
import request from '@/api/merchant/request'

let currentMerchantId = null
let currentBaseUserId = null

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
</script>

<style scoped>
/* empty, purely functional component */
</style>
