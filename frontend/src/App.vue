<template>
  <!-- 这里是路由占位符，路由匹配的页面会显示在这里 -->
  <router-view />

  <!-- 全局消息通知（无 UI） -->
  <MessageNotify />

  <!-- 全局聊天弹窗：任何页面都可以通过 window.dispatchEvent(new CustomEvent('chat:open', {detail:{merchantId, userBaseId}})) 打开 -->
  <el-dialog v-model="showGlobalChat" title="聊天" :width="'380px'">
    <ChatWindow v-if="showGlobalChat" :merchantId="globalMerchantId" :userBaseId="globalUserBaseId" :token="token" :merchantName="globalMerchantName" :merchantAvatar="globalMerchantAvatar" />
  </el-dialog>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import MessageNotify from '@/components/Chat/MessageNotify.vue'
import ChatWindow from '@/components/Chat/ChatWindow.vue'
import { getMerchantDetail } from '@/api/chat'

const showGlobalChat = ref(false)
const globalMerchantId = ref(null)
const globalUserBaseId = ref(null)
const globalMerchantName = ref('商家')
const globalMerchantAvatar = ref('/imgs/merchant.png')
const token = (typeof window !== 'undefined' && localStorage.getItem('token')) || ''

async function openHandler(e) {
  const d = (e && e.detail) || {}
  if (!d.merchantId) return
  globalMerchantId.value = Number(d.merchantId)
  globalUserBaseId.value = d.userBaseId || null
  try {
    const r = await getMerchantDetail(globalMerchantId.value)
    if (r && r.data && r.data.data) {
      globalMerchantName.value = r.data.data.shop_name || globalMerchantName.value
      globalMerchantAvatar.value = r.data.data.logo || globalMerchantAvatar.value
    }
  } catch (err) {}
  showGlobalChat.value = true
}

onMounted(() => window.addEventListener('chat:open', openHandler))
onBeforeUnmount(() => window.removeEventListener('chat:open', openHandler))
</script>

<style>
/* 全局样式可写在这里 */
html, body, #app {
  height: 100%;
  margin: 0;
}
</style>
