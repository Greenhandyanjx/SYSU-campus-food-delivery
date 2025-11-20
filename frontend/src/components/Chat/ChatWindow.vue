<template>
  <div class="chat-container">
    
    <!-- 顶部标题栏 -->
    <div class="chat-header">
      <div class="header-left">
        <img class="avatar" :src="merchantAvatar" alt="商家" />
        <span class="title">{{ merchantName }} · 在线客服</span>
      </div>
      <button class="local-close" @click="$emit('close')" aria-label="关闭聊天">✕</button>
    </div>

    <!-- 消息区域 -->
    <div class="messages" ref="msgWrap">
      <div
        v-for="m in messages"
        :key="m.id"
        class="message-row"
        :class="{ 'me': isMyMessage(m) }"
      >
        <!-- 头像 -->
        <img 
          class="msg-avatar" 
          :src="isMyMessage(m) ? merchantAvatar : userAvatarLocal"
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

// 当前登录者 base_user id（用于判断消息方向：来自当前者 = 我）
const currentBaseId = ref(null)

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
  // 先尝试推断当前登录者 base id（用于判断消息的“我”）
  try {
    const cur = await getBaseUserDetail()
    if (cur && cur.data && cur.data.data) {
      currentBaseId.value = cur.data.data.id
    }
  } catch (e) {}

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
  // user info: 若没有传入 props.userBaseId，则默认把当前登录者视为 user（普通用户打开聊天）
  if (!props.userBaseId) {
    if (currentBaseId.value) {
      userBaseIdLocal.value = currentBaseId.value
    } else {
      try {
        const u = await getBaseUserDetail()
        if (u && u.data && u.data.data) {
          userBaseIdLocal.value = u.data.data.id
          userAvatarLocal.value = userAvatarLocal.value || '/imgs/user.png'
          userNameLocal.value = u.data.data.username || userNameLocal.value
        }
      } catch (e) {}
    }
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
  merchant_id: Number(props.merchantId),
  user_base_id: userBaseIdLocal.value,  // ⭐⭐ 必须加
  content: input.value,
  type: 'text'
}


  console.log('[ChatWindow] send payload', payload)

  // 只发送一次，并且只发送对象，让 chatClient 来 stringify
  const ok = chatClient.send(payload)

  if (!ok) {
    console.warn('[ChatWindow] chatClient failed, fallback to local ws')
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(payload))
    }
  }

  // 本地立即显示一条消息
  messages.value.push({
    from_base_id: currentBaseId.value || userBaseIdLocal.value,
    user_base_id: userBaseIdLocal.value,
    merchant_id: payload.merchant_id,
    content: payload.content,
    type: payload.type,
    created_at: new Date().toISOString()
  })

  nextTick(scrollBottom)
  input.value = ''
}



function scrollBottom() {
  if (msgWrap.value) {
    msgWrap.value.scrollTop = msgWrap.value.scrollHeight
  }
}
const isMerchantSide = ref(false)

async function detectRole() {
  try {
    const cur = await getBaseUserDetail()
    if (cur?.data?.data) {
      currentBaseId.value = cur.data.data.id
      if (Number(currentBaseId.value) === Number(props.merchantId)) {
        isMerchantSide.value = true
      }
    }
  } catch (e) {}
}

const isMyMessage = (msg) => {
  if (isMerchantSide.value) {
    return Number(msg.from_base_id) === Number(props.merchantId)
  }
  return Number(msg.from_base_id) === Number(currentBaseId.value)
}

onMounted(async () => {
  await detectRole()
  await ensureNames()
  await loadHistory()
  connectWs()
  chatClient.onMessage(handleGlobalMessage)
})

onBeforeUnmount(() => {
  ws?.close()
  chatClient.offMessage(handleGlobalMessage)
})
</script>

<style scoped>
/* ====================== 外层布局 ====================== */
.chat-container {
  width: 100%;
  max-width: 450px;
  height: 100%; /* fill the parent modal height */
  border: 1px solid #e5e5e5;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  background: #fff;
  box-shadow: 0 10px 30px rgba(0,0,0,0.12);
  overflow: hidden;                 /* 重点：防止子元素溢出遮挡 */
}

/* ====================== 顶部栏 ====================== */
.chat-header {
  height: 60px;
  flex-shrink: 0;                   /* 禁止被压缩 */
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background: #ffd600;
  color: #000;
}

.chat-header .header-left { display:flex; align-items:center }

.local-close {
  border: none;
  background: transparent;
  font-size: 18px;
  cursor: pointer;
  padding: 6px 8px;
  border-radius: 6px;
}
.local-close:hover { background: rgba(0,0,0,0.06) }

.chat-header .avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid #fff;
}

.chat-header .title {
  margin-left: 12px;
  font-size: 16px;
  font-weight: bold;
}

/* ====================== 消息区域 ====================== */
.messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 12px;
  background: #f5f5f5;
  /* 解决 iPhone 底部安全区被遮挡 */
  padding-bottom: env(safe-area-inset-bottom, 20px);
}

/* 滚动条美化（可选） */
.messages::-webkit-scrollbar {
  width: 4px;
}
.messages::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.2);
  border-radius: 2px;
}

.message-row {
  display: flex;
  margin-bottom: 16px;
  align-items: flex-end;             /* 关键：让气泡底部对齐 */
}

.message-row.me {
  flex-direction: row-reverse;
}

/* 头像 */
.msg-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* 气泡容器 */
.bubble-wrapper {
  max-width: 72%;
  position: relative;
}

/* 气泡主体 */
.bubble {
  padding: 10px 14px;
  border-radius: 18px;
  font-size: 15px;
  line-height: 1.45;
  word-break: break-word;
  position: relative;
  display: inline-block;
}

/* 左边（商家）气泡 - 白色 + 尖角 */
.message-row:not(.me) .bubble {
  background: #ffffff;
  border: 1px solid #e8e8e8;
  margin-left: 8px;
}

/* 右边（自己）气泡 - 美团黄 + 尖角 */
.message-row.me .bubble {
  background: #ffe563;
  margin-right: 8px;
}

/* 气泡小尖角（纯 CSS 实现） */
.message-row:not(.me) .bubble::before {
  content: "";
  position: absolute;
  left: -7px;
  bottom: 8px;
  border: 8px solid transparent;
  border-right-color: #ffffff;
}
.message-row:not(.me) .bubble::after {
  content: "";
  position: absolute;
  left: -9px;
  bottom: 8px;
  border: 8px solid transparent;
  border-right-color: #e8e8e8;
}

.message-row.me .bubble::before {
  content: "";
  position: absolute;
  right: -7px;
  bottom: 8px;
  border: 8px solid transparent;
  border-left-color: #ffe563;
}

/* 时间 */
.time {
  font-size: 11px;
  color: #999;
  margin-top: 4px;
  text-align: center;
  padding: 0 8px;
}

/* 我发的消息时间在左边 */
.message-row.me .time {
  text-align: right;
}

/* ====================== 输入栏（关键修复）====================== */
.input-bar {
  flex-shrink: 0;                   /* 禁止被压缩 */
  height: 64px;
  padding: 10px 12px;
  background: #fff;
  border-top: 1px solid #eee;
  display: flex;
  align-items: center;
  /* 解决 iPhone 刘海屏底部被遮挡 */
  padding-bottom: env(safe-area-inset-bottom, 10px);
}

.input-bar input {
  flex: 1;
  height: 44px;
  border: 1px solid #ddd;
  border-radius: 22px;
  padding: 0 16px;
  font-size: 15px;
  outline: none;
  background: #f9f9f9;
}

.input-bar input:focus {
  border-color: #ffd600;
  background: #fff;
}

.send-btn {
  margin-left: 10px;
  width: 72px;
  height: 44px;
  background: #ffd600;
  border: none;
  border-radius: 22px;
  font-size: 15px;
  font-weight: bold;
  color: #000;
  cursor: pointer;
  transition: all 0.2s;
}

.send-btn:hover {
  background: #ffeb3b;
}

.send-btn:active {
  transform: scale(0.95);
}
</style>
