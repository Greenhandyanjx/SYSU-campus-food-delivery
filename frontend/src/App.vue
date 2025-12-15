<template>
  <!-- 这里是路由占位符，路由匹配的页面会显示在这里 -->
  <router-view />

  <!-- 全局消息通知（无 UI） -->
  <MessageNotify />

  <!-- 全局聊天弹窗：使用 teleport 在 body 上展示居中模态窗口，行为与商家端一致 -->
  <teleport to="body">
    <div v-if="showGlobalChat" class="global-chat-overlay" @click.self="showGlobalChat = false">
      <div class="global-chat-modal">
        <ChatWindow
          :key="String(globalMerchantId) + '-' + String(globalUserBaseId)"
          :merchantId="globalMerchantId"
          :userBaseId="globalUserBaseId"
          :token="token"
          :merchantName="globalMerchantName"
          :merchantAvatar="globalMerchantAvatar"
          @close="showGlobalChat = false"
        />
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import MessageNotify from '@/components/Chat/MessageNotify.vue'
import ChatWindow from '@/components/Chat/ChatWindow.vue'
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
  // robustly accept merchant id in either camelCase or snake_case and treat 0 as valid
  let mid = null
  if (d && Object.prototype.hasOwnProperty.call(d, 'merchantId')) {
    mid = d.merchantId
  } else if (d && Object.prototype.hasOwnProperty.call(d, 'merchant_id')) {
    mid = d.merchant_id
  }

  // also accept userBaseId variants
  const userBase = (d && (d.userBaseId || d.user_base_id)) || null

  console.log('[App] chat:open received', d, 'parsed merchantId=', mid, 'userBaseId=', userBase)

  // only treat merchant as present when not null/undefined (0 is valid)
  if (mid !== null && typeof mid !== 'undefined') {
    globalMerchantId.value = Number(mid)
  } else {
    globalMerchantId.value = null
  }
  globalUserBaseId.value = userBase

  // 仅在非商家端展示全局聊天弹窗（商家使用页面内本地弹窗）
  if (globalMerchantId.value !== null) {
    if (!isMerchant.value) {
      console.log('[App] showing global chat modal for', globalMerchantId.value, globalUserBaseId.value)
      showGlobalChat.value = true
    } else {
      console.log('[App] running as merchant — skipping global chat modal')
    }
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
      // ignore enrichment error
    }
  }
}

onMounted(() => window.addEventListener('chat:open', openHandler))
onBeforeUnmount(() => window.removeEventListener('chat:open', openHandler))

// listen for MessageNotify detection of merchant role
isMerchant.value = (localStorage.getItem('isMerchant') === '1')
window.addEventListener('merchant:detected', () => { isMerchant.value = (localStorage.getItem('isMerchant') === '1') })
</script>

<style>
/* 全局样式可写在这里 */
html, body, #app {
  height: 100%;
  margin: 0;
}

/* 固定定位全局聊天窗口，右下角 */
.global-chat-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display:flex;
  align-items:center;
  justify-content:center;
  z-index: 2147483647;
}
.global-chat-modal { z-index:2147483648 }
</style>
