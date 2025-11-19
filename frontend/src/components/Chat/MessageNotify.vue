<template>
  <div />
</template>

<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import chatClient from '@/utils/chatClient'
import { ElNotification } from 'element-plus'

function handleIncoming(msg) {
  // Expect msg to include merchant_id and content and from_base_id
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
}

onMounted(() => {
  chatClient.onMessage(handleIncoming)
  // ensure connection started
  chatClient.connect()
})

onBeforeUnmount(() => {
  chatClient.offMessage(handleIncoming)
})
</script>

<style scoped>
/* empty, purely functional component */
</style>
