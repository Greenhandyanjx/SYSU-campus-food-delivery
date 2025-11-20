<template>
  <div class="merchant-chat">
    <div class="m-header">
      <div class="m-left">
        <img class="m-avatar" :src="merchantAvatar" />
        <div class="m-title">{{ merchantName }} · 在线客服</div>
      </div>
      <button class="m-close" @click="$emit('close')">✕</button>
    </div>

    <div class="m-messages" ref="wrap">
      <div v-for="m in messages" :key="m.id" class="m-row" :class="{ me: isMyMessage(m) }">
        <img class="m-msg-avatar" :src="isMyMessage(m) ? myAvatar : otherAvatar" />
        <div class="m-bubble-wrapper">
          <div class="m-bubble">{{ m.content }}</div>
          <div class="m-time">{{ formatTime(m.created_at) }}</div>
        </div>
      </div>
    </div>

    <div class="m-input">
      <input v-model="input" @keyup.enter="send" placeholder="请输入内容..." />
      <button class="m-send" @click="send">发送</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import chatClient from '@/utils/chatClient'
import { getChatHistory, getMerchantDetail, getBaseUserDetail } from '@/api/chat'
import request from '@/api/merchant/request'

const props = defineProps({
  merchantId: { type: Number, required: false },
  userBaseId: { type: Number, required: false },
})

const messages = ref([])
const input = ref('')
const wrap = ref(null)

const merchantName = ref('商家')
const merchantAvatar = ref('/imgs/merchant.png')
const myAvatar = ref('/imgs/user.png')
const otherAvatar = ref('/imgs/merchant.png')

const currentBaseId = ref(null)
const userBaseIdLocal = ref(props.userBaseId || null)

function formatTime(s) {
  const d = new Date(s)
  return `${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function loadHistory() {
  if (!userBaseIdLocal.value) return
  const r = await getChatHistory(props.merchantId, userBaseIdLocal.value)
  if (r && r.data && r.data.data) {
    messages.value = r.data.data.reverse()
    await nextTick(scrollBottom)
  }
}

function scrollBottom() {
  if (wrap.value) wrap.value.scrollTop = wrap.value.scrollHeight
}

function isMyMessage(msg) {
  return Number(msg.from_base_id) === Number(currentBaseId.value)
}

function handleIncoming(msg) {
  const mid = msg.merchant_id || msg.merchantId
  const uid = msg.user_base_id || msg.userBaseId
  if (!mid || Number(mid) !== Number(props.merchantId)) return
  // if userBaseIdLocal is set, only append messages for that conversation
  if (userBaseIdLocal.value && Number(uid) !== Number(userBaseIdLocal.value)) return
  messages.value.push(msg)
  nextTick(scrollBottom)
}

async function ensure() {
  // current base
  try {
    const cur = await getBaseUserDetail()
    if (cur?.data?.data) currentBaseId.value = cur.data.data.id
  } catch (e) {}

  // merchant info: 如果没有传入 merchantId，则尝试通过当前 base user 去查商家
  try {
    if (props.merchantId) {
      const r = await getMerchantDetail(props.merchantId)
      if (r && r.data && r.data.data) {
        merchantName.value = r.data.data.shop_name || merchantName.value
        merchantAvatar.value = r.data.data.logo || merchantAvatar.value
        otherAvatar.value = merchantAvatar.value
      }
    } else if (currentBaseId.value) {
      // 直接请求 /merchant/detail?base_id=xxx
      const req = await fetch(`/api/merchant/detail?base_id=${currentBaseId.value}`)
      const jr = await req.json()
      if (jr && jr.data) {
        merchantName.value = jr.data.shop_name || merchantName.value
        merchantAvatar.value = jr.data.logo || merchantAvatar.value
        otherAvatar.value = merchantAvatar.value
      }
    }
  } catch (e) {}

  // user id
  if (!userBaseIdLocal.value) {
    try {
      const u = await getBaseUserDetail()
      if (u && u.data && u.data.data) {
        userBaseIdLocal.value = u.data.data.id
        myAvatar.value = '/imgs/user.png'
      }
    } catch (e) {}
  }
}

function send() {
  if (!input.value) return
  const payload = { merchant_id: Number(props.merchantId), user_base_id: userBaseIdLocal.value, content: input.value, type: 'text' }
  const ok = chatClient.send(payload)
  if (!ok) console.warn('[MerchantChatWindow] chatClient not open')
  // optimistic
  messages.value.push({ from_base_id: currentBaseId.value || userBaseIdLocal.value, user_base_id: userBaseIdLocal.value, merchant_id: payload.merchant_id, content: payload.content, type: payload.type, created_at: new Date().toISOString() })
  input.value = ''
  nextTick(scrollBottom)
}

onMounted(async () => {
  await ensure()
  await loadHistory()
  // 标记为已读（仅标记用户发来的消息）
  try {
    await request.post('/merchant/chats/mark_read', { merchant_id: props.merchantId, user_base_id: userBaseIdLocal.value })
  } catch (e) {}
  chatClient.onMessage(handleIncoming)
  chatClient.connect()
})

onBeforeUnmount(() => {
  chatClient.offMessage(handleIncoming)
})
</script>

<style scoped>
.merchant-chat { width: 360px; height: 420px; display:flex; flex-direction:column; border-radius:10px; overflow:hidden; background:#fff; border:1px solid #eee }
.m-header { height:56px; background:#ffd600; display:flex; align-items:center; justify-content:space-between; padding:0 12px }
.m-left { display:flex; align-items:center }
.m-avatar { width:36px; height:36px; border-radius:50%; margin-right:8px }
.m-title { font-weight:600 }
.m-close { background:transparent; border:none; font-size:16px; cursor:pointer }
.m-messages { flex:1; overflow-y:auto; padding:12px; background:#f5f5f5 }
.m-row { display:flex; margin-bottom:12px; align-items:flex-end }
.m-row.me { flex-direction:row-reverse }
.m-msg-avatar { width:32px; height:32px; border-radius:50%; flex-shrink:0 }
.m-bubble-wrapper { max-width:72%; position:relative }
.m-bubble { padding:8px 12px; border-radius:14px; background:#fff; border:1px solid #e8e8e8 }
.m-row.me .m-bubble { background:#ffe563 }
.m-time { font-size:11px; color:#999; margin-top:4px }
.m-input { height:56px; display:flex; padding:8px; gap:8px; background:#fff; align-items:center }
.m-input input { flex:1; height:40px; border-radius:20px; border:1px solid #ddd; padding:0 12px }
.m-send { width:64px; height:40px; border-radius:20px; background:#ffd600; border:none; cursor:pointer }
</style>
