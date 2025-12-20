<template><div style="display:none"><slot/></div></template>

<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'
import chatClient from '@/utils/chatClient'
import { getBaseUserDetail } from '@/api/chat'
import request from '@/api/merchant/request'
import { useChatStore } from '@/stores/chatStore'

const chatStore = useChatStore()
let baseUserId = null
let isMerchant = false
let merchantId = null

async function detectRole() {
  try {
    const cur = await getBaseUserDetail()
    if (cur && cur.data && cur.data.data) {
      baseUserId = cur.data.data.id
    }
  } catch (e) {}
  try {
    if (baseUserId) {
      const r = await request({ url: '/merchant/detail', method: 'get', params: { base_id: baseUserId } })
      if (r && r.data && r.data.data) {
        merchantId = r.data.data.id
        isMerchant = true
        try { localStorage.setItem('isMerchant', '1'); localStorage.setItem('merchantId', String(merchantId)) } catch(e) {}
      }
    }
  } catch (e) {}
}

function handleIncoming(msg) {
  try {
    // determine peer id depending on role
    const uid = msg.user_base_id || msg.userBaseId || null
    const mid = msg.merchant_id || msg.merchantId || null
    const from = msg.from_base_id || msg.fromBaseId || null
    const peer = (isMerchant ? uid : mid) || uid || mid || null
    if (!peer) return
    const peerStr = String(peer)
    const isSelf = from && baseUserId && Number(from) === Number(baseUserId)
    try { chatStore.addMessage(peerStr, msg, !!isSelf) } catch (e) { console.warn('ChatBootstrap addMessage failed', e) }

    // dispatch global event so mounted UI components can update lists
    try { window.dispatchEvent(new CustomEvent('chat:incoming', { detail: msg })) } catch (e) {}
  } catch (e) {
    console.warn('ChatBootstrap handler error', e)
  }
}

onMounted(async () => {
  await detectRole()
  chatClient.onMessage(handleIncoming)
  try { chatClient.connect() } catch (e) { }
})

onBeforeUnmount(() => {
  chatClient.offMessage(handleIncoming)
})
</script>
