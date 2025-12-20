<template>
  <router-view />

  <MessageNotify />

  <teleport to="body">
    <div v-if="showGlobalChat" class="global-chat-overlay" @click.self="showGlobalChat = false">
      <div class="global-chat-modal">
        <template v-if="!isMerchant">
          <ChatWindow
            :key="String(globalMerchantId) + '-' + String(globalUserBaseId)"
            :merchantId="globalMerchantId"
            :userBaseId="globalUserBaseId"
            :token="token"
            :merchantName="globalMerchantName"
            :merchantAvatar="globalMerchantAvatar"
            @close="showGlobalChat = false"
          />
        </template>
        <template v-else>
          <MerchantChatWindow
            :merchantId="globalMerchantId"
            :userBaseId="globalUserBaseId"
            @close="showGlobalChat = false"
          />
        </template>
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import MessageNotify from '@/components/Chat/MessageNotify.vue'
import ChatWindow from '@/components/Chat/ChatWindow.vue'
import MerchantChatWindow from '@/components/Chat/MerchantChatWindow.vue'
import { getMerchantDetail } from '@/api/chat'

const showGlobalChat = ref(false)
const globalMerchantId = ref(null)
const globalUserBaseId = ref(null)
const globalMerchantName = ref('商家')
const globalMerchantAvatar = ref('/imgs/merchant.png')
const token = (typeof window !== 'undefined' && localStorage.getItem('token')) || ''
const isMerchant = ref(false)

async function openHandler(e) {
  const d = (e && e.detail) || {}
  let mid = null
  if (d && Object.prototype.hasOwnProperty.call(d, 'merchantId')) {
    mid = d.merchantId
  } else if (d && Object.prototype.hasOwnProperty.call(d, 'merchant_id')) {
    mid = d.merchant_id
  }

  const userBase = (d && (d.userBaseId || d.user_base_id)) || null

  console.log('[App] chat:open received', d, 'parsed merchantId=', mid, 'userBaseId=', userBase)

  if (mid !== null && typeof mid !== 'undefined') {
    globalMerchantId.value = Number(mid)
  } else {
    globalMerchantId.value = null
  }
  globalUserBaseId.value = userBase

  if (globalMerchantId.value !== null) {
    console.log(isMerchant.value ? '[App] running as merchant — showing merchant chat modal' : '[App] showing global chat modal for', globalMerchantId.value, globalUserBaseId.value)
    showGlobalChat.value = true
  } else {
    console.log('[App] not showing global chat: merchantId is null or undefined')
  }

  if (globalMerchantId.value !== null) {
    try {
      const r = await getMerchantDetail(globalMerchantId.value)
      if (r && r.data && r.data.data) {
        globalMerchantName.value = r.data.data.shop_name || globalMerchantName.value
        globalMerchantAvatar.value = r.data.data.logo || globalMerchantAvatar.value
      }
    } catch (err) {
      // ignore
    }
  }
}

onMounted(() => window.addEventListener('chat:open', openHandler))
onBeforeUnmount(() => window.removeEventListener('chat:open', openHandler))

isMerchant.value = (localStorage.getItem('isMerchant') === '1')
window.addEventListener('merchant:detected', () => { isMerchant.value = (localStorage.getItem('isMerchant') === '1') })
</script>

<style>
html, body, #app { height: 100%; margin: 0 }
.global-chat-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display:flex; align-items:center; justify-content:center; z-index:2147483647 }
.global-chat-modal { z-index:2147483648 }
</style>
