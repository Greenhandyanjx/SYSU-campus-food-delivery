<template>
  <div class="chat-container">
    
    <!-- 顶部标题栏 -->
    <div class="chat-header">
      <img class="avatar" :src="merchantAvatar" alt="商家" />
      <span class="title">{{ merchantName }} · 在线客服</span>
    </div>

    <!-- 消息区域 -->
    <div class="messages" ref="msgWrap">
      <div
        v-for="m in messages"
        :key="m.id"
        class="message-row"
        :class="{ 'me': m.from_base_id === userBaseIdLocal }"
      >
        <!-- 头像 -->
        <img 
          class="msg-avatar" 
          :src="m.from_base_id === userBaseIdLocal ? userAvatarLocal : merchantAvatar" 
        />

        <!-- 气泡 -->
        <div class="bubble-wrapper">
          <div class="bubble">{{ m.content }}</div>
          <div class="time">{{ formatTime(m.created_at) }}</div>
        </div>
      </div>
    </div>

    <!-- 底部输入框 -->
    <div class="input-bar">
      <input v-model="input" @keyup.enter="send" placeholder="请输入内容..." />
      <button class="send-btn" @click="send">发送</button>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { getChatHistory, getWsUrl, getMerchantDetail, getBaseUserDetail } from '@/api/chat'
import chatClient from '@/utils/chatClient'

const props = defineProps({
  merchantId: { type: Number, required: true },
  userBaseId: { type: Number, required: false },
  token: { type: String, required: true },
  merchantName: { type: String, default: "商家" },
  merchantAvatar: { type: String, default: "/imgs/merchant.png" },
  userAvatar: { type: String, default: "/imgs/user.png" }
})

const messages = ref([])
const input = ref('')
let ws = null
const msgWrap = ref(null)

// reactive local display fields (initialized from props)
const merchantName = ref(props.merchantName || '商家')
const merchantAvatar = ref(props.merchantAvatar || '/imgs/merchant.png')
const userAvatarLocal = ref(props.userAvatar || '/imgs/user.png')
const userNameLocal = ref('我')
const userBaseIdLocal = ref(props.userBaseId || null)

function formatTime(s) {
  const d = new Date(s)
  return `${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

async function loadHistory() {
  // load history using the resolved userBaseId (from ensureNames)
  if (!userBaseIdLocal.value || !props.merchantId) return
  const res = await getChatHistory(props.merchantId, userBaseIdLocal.value)
  if (res && res.data?.data) {
    messages.value = res.data.data.reverse()
    await nextTick()
    scrollBottom()
  }
}

async function ensureNames() {
  // merchant info
  if ((!props.merchantName || props.merchantName === '商家') && props.merchantId) {
    try {
      const r = await getMerchantDetail(props.merchantId)
      if (r && r.data && r.data.data) {
        merchantName.value = r.data.data.shop_name || r.data.data.shopName || merchantName.value
        merchantAvatar.value = r.data.data.logo || r.data.data.logoUrl || merchantAvatar.value
      }
    } catch (e) {}
  }

  // user info
  if (!props.userBaseId) {
    try {
      const u = await getBaseUserDetail()
      if (u && u.data && u.data.data) {
        userBaseIdLocal.value = u.data.data.id
        userAvatarLocal.value = userAvatarLocal.value || '/imgs/user.png'
        userNameLocal.value = u.data.data.username || userNameLocal.value
      }
    } catch (e) {}
  } else {
    try {
      const u = await getBaseUserDetail(props.userBaseId)
      if (u && u.data && u.data.data) {
        userBaseIdLocal.value = u.data.data.id
        userNameLocal.value = u.data.data.username || userNameLocal.value
      }
    } catch (e) {}
  }
}

function connectWs() {
  // backend expects raw token without the "Bearer " prefix in the query param
  const pureToken = (props.token || '').replace(/^Bearer\s+/i, '')
  const url = getWsUrl() + `?token=${encodeURIComponent(pureToken)}`
  ws = new WebSocket(url)

  ws.onmessage = (ev) => {
    try {
      const data = JSON.parse(ev.data)
      messages.value.push(data)
      nextTick(scrollBottom)
    } catch (err) {
      // ignore malformed messages
    }
  }
}

function handleGlobalMessage(msg) {
  // append only if message belongs to this chat (merchant/user pair)
  const mid = msg.merchant_id || msg.merchantId
  const uid = msg.user_base_id || msg.userBaseId
  if (!mid || Number(mid) !== Number(props.merchantId)) return
  // push and scroll
  messages.value.push(msg)
  nextTick(scrollBottom)
}

function send() {
  if (!input.value) return
  const payload = {
    merchant_id: props.merchantId,
    content: input.value,
    type: 'text',
    from_base_id: userBaseIdLocal.value,
    created_at: new Date().toISOString()
  }

  messages.value.push(payload)   // <-- 立即渲染
  nextTick(scrollBottom)

  const sent = chatClient.send(payload)
  if (!sent && ws && ws.readyState === WebSocket.OPEN) {
    try { ws.send(JSON.stringify(payload)) } catch (e) {}
  }

  input.value = ''
}


function scrollBottom() {
  if (msgWrap.value) {
    msgWrap.value.scrollTop = msgWrap.value.scrollHeight
  }
}

onMounted(async () => {
  // ensure we know current user id and merchant display info first
  await ensureNames()
  await loadHistory()
  connectWs()
  // subscribe to global client so messages delivered while dialog closed still appear
  chatClient.onMessage(handleGlobalMessage)
})
onBeforeUnmount(() => {
  ws?.close()
  chatClient.offMessage(handleGlobalMessage)
})
</script>

<style scoped>
/* ======================
   外层布局
====================== */
.chat-container {
  width: 100%;
  max-width: 450px;
  height: 620px;
  border: 1px solid #e5e5e5;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  background: #fff;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
  overflow: hidden;
}

/* ======================
   顶部栏
====================== */
.chat-header {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 16px;
  background: #ffd600;
  border-bottom: 1px solid #f0f0f0;
}

.chat-header .avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
}

.chat-header .title {
  margin-left: 12px;
  font-size: 16px;
  font-weight: bold;
}

/* ======================
   消息列表
====================== */
.messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: #fafafa;
}

.message-row {
  display: flex;
  margin-bottom: 14px;
}

.message-row.me {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  margin: 0 10px;
}

.bubble-wrapper {
  max-width: 70%;
}

.bubble {
  padding: 10px 14px;
  border-radius: 14px;
  font-size: 14px;
  line-height: 1.4;
  word-break: break-word;
  display: inline-block;
}

/* 商家气泡（灰） */
.message-row .bubble {
  background: #ffffff;
  border: 1px solid #e0e0e0;
}

/* 用户气泡（美团黄） */
.message-row.me .bubble {
  background: #ffe980;
  border: none;
}

/* 时间 */
.time {
  font-size: 11px;
  color: #a8a8a8;
  margin-top: 4px;
}

/* ======================
   底部输入框
====================== */
.input-bar {
  height: 64px;
  display: flex;
  padding: 10px 12px;
  background: #fff;
  border-top: 1px solid #eee;
}

.input-bar input {
  flex: 1;
  border: 1px solid #ddd;
  border-radius: 18px;
  padding: 8px 12px;
  outline: none;
}

.send-btn {
  background: #ffd600;
  border: none;
  padding: 0 16px;
  margin-left: 10px;
  border-radius: 18px;
  font-weight: bold;
  cursor: pointer;
}

.send-btn:active {
  background: #f3c900;
}
</style>
