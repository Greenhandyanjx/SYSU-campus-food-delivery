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

function handleIncoming(msg) {
  // Expect msg to include merchant_id and content and from_base_id
  console.log('[MessageNotify] incoming', msg)
  const title = msg.merchant_id ? `来自商家 ${msg.merchant_id}` : '新消息'
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
    // 直接调用后端 /merchant/detail?base_id=xxx
    return request({ url: '/merchant/detail', method: 'get', params: { base_id: base.id } })
  }).catch(() => null).then(r => {
    if (!r) return
    if (r && r.data && r.data.data) {
      currentMerchantId = r.data.data.id
      console.log('[MessageNotify] detected merchant id =', currentMerchantId)
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
