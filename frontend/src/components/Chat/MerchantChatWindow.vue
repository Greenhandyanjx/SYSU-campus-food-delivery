<template>
  <div class="merchant-chat">
    <div class="merchant-card">

    <div class="m-header">
      <div class="m-left">
        <img class="m-avatar" :src="merchantAvatar" />
        <div class="m-title">{{ chatUserName || ('用户 ' + (userBaseIdLocal || '')) }} · 会话</div>
      </div>
      <div style="display:flex;align-items:center;gap:8px">
        <img src="/JDlogo.png" class="jd-logo" alt="嘉递" />
        <!-- <button class="m-close" @click="$emit('close')">✕</button> -->
      </div>
    </div>

    <div class="m-messages" ref="wrap">
      <div v-for="m in messages" :key="m.id" class="m-row" :class="{ me: isMyMessage(m) }">
        <img class="m-msg-avatar" :src="isMyMessage(m) ? myAvatar : otherAvatar" />
        <div class="m-bubble-wrapper">
          <div class="m-bubble">{{ m.content }}</div>
          <div class="m-time">{{ formatDateToCN(m.created_at) }}</div>
        </div>
      </div>
    </div>

    <div class="m-input">
      <input v-model="input" @keyup.enter="send" placeholder="请输入内容..." />
      <button class="m-send" @click="send">发送</button>
    </div>
  </div>

  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import merchantSvg from '@/assets/merchant.svg'
import userPng from '@/assets/user.png'
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
const merchantAvatar = ref(merchantSvg)
const myAvatar = ref(merchantSvg)
const otherAvatar = ref(userPng)
const chatUserName = ref('')

const currentBaseId = ref(null)
const userBaseIdLocal = ref(props.userBaseId || null)

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
        // 当前登录者为商家时，myAvatar 应为商家头像；不要把 otherAvatar 设为商家头像
        myAvatar.value = merchantAvatar.value || myAvatar.value
      }
    } else if (currentBaseId.value) {
      // 直接请求 /merchant/detail?base_id=xxx
      const req = await fetch(`/api/merchant/detail?base_id=${currentBaseId.value}`)
      const jr = await req.json()
      if (jr && jr.data) {
        merchantName.value = jr.data.shop_name || merchantName.value
        merchantAvatar.value = jr.data.logo || merchantAvatar.value
        myAvatar.value = merchantAvatar.value || myAvatar.value
      }
    }
  } catch (e) {}

  // user id
  if (!userBaseIdLocal.value) {
    try {
      const u = await getBaseUserDetail()
      if (u && u.data && u.data.data) {
        userBaseIdLocal.value = u.data.data.id
        // myAvatar should be merchant avatar; otherAvatar should be user's avatar when available
        myAvatar.value = merchantAvatar.value || merchantSvg
        otherAvatar.value = userPng
      }
    } catch (e) {}
  }

  // 尝试获取会话用户的名称（用于在商家端显示用户名字）
  if (userBaseIdLocal.value) {
    try {
      const ru = await getBaseUserDetail(userBaseIdLocal.value)
      const uu = ru?.data?.data
      if (uu) chatUserName.value = uu.username || uu.nickname || ''
      // 用户头像优先取后端返回的 avatar 字段
      if (uu) otherAvatar.value = uu.avatar || uu.avatarUrl || otherAvatar.value || userPng
    } catch (e) { }
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
  // 通知其他组件：某个会话已被标记为已读
  try {
    window.dispatchEvent(new CustomEvent('merchant:chats:marked_read', { detail: { merchant_id: props.merchantId, user_base_id: userBaseIdLocal.value } }))
  } catch (e) {}
  chatClient.onMessage(handleIncoming)
  chatClient.connect()
})

// 当 props.merchantId 或 props.userBaseId 变化时重新加载
watch(() => props.userBaseId, async (nv) => {
  if (!nv) return
  messages.value = []
  userBaseIdLocal.value = nv
  await loadHistory()
  try { await request.post('/merchant/chats/mark_read', { merchant_id: props.merchantId, user_base_id: userBaseIdLocal.value }) } catch (e) {}
})

watch(() => props.merchantId, async (nv) => {
  if (!nv) return
  messages.value = []
  await ensure()
  await loadHistory()
})

onBeforeUnmount(() => {
  chatClient.offMessage(handleIncoming)
})
</script>

<style scoped>
/* 使商家端聊天窗口外观与用户端一致：去除外边框、固定大小、消息区独立滚动 */
.merchant-chat {
  width: 400px;
  height: 700px;
  min-width: 400px;
  min-height: 700px;
  display:flex;
  flex-direction:column;
  border-radius:12px;
  overflow:hidden;
  background: transparent; /* 由内部卡片负责白色背景 */
  border: none;
}
.merchant-card {
  width: 100%;
  height: 100%;
  display:flex;
  flex-direction:column;
  border-radius:12px;
  background:#fff;
  box-shadow: 0 10px 30px rgba(0,0,0,0.12);
  overflow:hidden;
}
.m-header { height:60px; background:#ffd600; display:flex; align-items:center; justify-content:space-between; padding:0 14px }
.m-left { display:flex; align-items:center }
.m-avatar { width:40px; height:40px; border-radius:50%; margin-right:10px }
.m-title { font-weight:700 }
.m-close { background:transparent; border:none; font-size:18px; cursor:pointer }
.jd-logo { width:28px; height:28px; object-fit:contain; border-radius:4px ;margin-right:30px;}
/* 消息区独立滚动 */
.m-messages { flex:1; overflow-y:auto; padding:14px; background:#f5f5f5; -webkit-overflow-scrolling: touch }
.m-row { display:flex; margin-bottom:14px; align-items:flex-end }
.m-row.me { flex-direction:row-reverse }
.m-msg-avatar { width:36px; height:36px; border-radius:50%; flex-shrink:0 ;margin-bottom: 20px;}
.m-bubble-wrapper { max-width:72%; position:relative }
.m-bubble { padding:10px 14px; border-radius:16px; background:#fff; border:1px solid #e8e8e8 ; word-break:break-word ;width: auto;}
.m-row.me .m-bubble { background:#ffe563 }
.m-time { font-size:11px; color:#999; margin-top:6px }
.m-input { height:64px; display:flex; padding:10px; gap:10px; background:#fff; align-items:center }
.m-input input { flex:1; height:44px; border-radius:22px; border:1px solid #ddd; padding:0 14px }
.m-send { width:72px; height:44px; border-radius:22px; background:#ffd600; border:none; cursor:pointer }
</style>
