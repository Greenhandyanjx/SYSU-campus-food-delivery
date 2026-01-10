<template>
  <div class="chat-container">
    
    <!-- 顶部标题栏 -->
    <div class="chat-card">

      <div class="chat-header">
      <div class="header-left">
        <img class="avatar" :src="merchantAvatar" alt="商家" @error="onMerchantImgError($event)" />
        <span class="title">{{ merchantName }} · 在线客服</span>
      </div>
      <div style="display:flex;align-items:center;gap:8px">
        <img src="/JDlogo.png" class="jd-logo" alt="嘉递" />
        <button class="local-close" @click="$emit('close')" aria-label="关闭聊天">✕</button>
      </div>
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
          :src="isMyMessage(m) ? userAvatarLocal : merchantAvatar"
          @error="onMsgImgError($event, isMyMessage(m))"
        />

        <!-- 气泡 -->
        <div class="bubble-wrapper">
          <div class="bubble">{{ m.content }}</div>
          <div class="time">{{ formatDateToCN(m.created_at) }}</div>
        </div>
      </div>
    </div>

    <!-- 底部输入框 -->
    <div class="input-bar">
      <input v-model="input" @keyup.enter="send" placeholder="请输入内容..." />
      <button class="send-btn" @click="send">发送</button>
    </div>

  </div>
</div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { getChatHistory, getWsUrl, getMerchantDetail, getBaseUserDetail } from '@/api/chat'
import request from '@/api/merchant/request'
import chatClient from '@/utils/chatClient'
import merchantSvg from '@/assets/merchant.svg'
import userPng from '@/assets/user.png'

const props = defineProps({
  merchantId: { type: Number, required: false },
  userBaseId: { type: Number, required: false },
  token: { type: String, required: false },
  merchantName: { type: String, default: "商家" },
  merchantAvatar: { type: String, default: '' },
  userAvatar: { type: String, default: '' }
})

const messages = ref([])
const input = ref('')
const msgWrap = ref(null)

// 当前登录者 base_user id（用于判断消息方向：来自当前者 = 我）
const currentBaseId = ref(null)

// reactive local display fields (initialized from props)
const merchantName = ref(props.merchantName || '商家')
const merchantAvatar = ref(props.merchantAvatar || merchantSvg)
const userAvatarLocal = ref(props.userAvatar || userPng)
function onMerchantImgError(ev) {
  try {
    const el = ev && ev.target
    if (!el) return
    // 避免无限触发：只允许一次回退
    if (el.dataset && el.dataset.fallback === '1') return
    if (el.dataset) el.dataset.fallback = '1'
    el.src = merchantSvg
  } catch (e) {}
}
function onMsgImgError(ev, isMine) {
  try {
    const el = ev && ev.target
    if (!el) return
    if (el.dataset && el.dataset.fallback === '1') return
    if (el.dataset) el.dataset.fallback = '1'
    el.src = isMine ? userPng : merchantSvg
  } catch (e) {}
}
const userNameLocal = ref('我')
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
  // load history using the resolved userBaseId (from ensureNames)
  // treat 0 as a valid id: only bail out when null/undefined
  if (userBaseIdLocal.value === null || typeof userBaseIdLocal.value === 'undefined') return
  if (props.merchantId === null || typeof props.merchantId === 'undefined') return
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
  if (props.userBaseId === null || typeof props.userBaseId === 'undefined') {
    if (currentBaseId.value) {
      userBaseIdLocal.value = currentBaseId.value
    } else {
      try {
        const u = await getBaseUserDetail()
        if (u && u.data && u.data.data) {
          userBaseIdLocal.value = u.data.data.id
          userAvatarLocal.value = userAvatarLocal.value || '/src/assets/user.png'
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

// We use centralized chatClient (singleton) for incoming push messages.
// Ensure the client is connected and subscribe to messages below.

function handleGlobalMessage(msg) {
  // append only if message belongs to this chat (merchant/user pair)
  const mid = msg.merchant_id || msg.merchantId
  const uid = msg.user_base_id || msg.userBaseId
  // ignore server echoes of messages we just sent (from_base_id equals our own base id)
  const from = msg.from_base_id || msg.fromBaseId
  if (from && (userBaseIdLocal.value !== null && typeof userBaseIdLocal.value !== 'undefined') && Number(from) === Number(userBaseIdLocal.value)) return
  if (mid === null || typeof mid === 'undefined' || Number(mid) !== Number(props.merchantId)) return
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
    console.warn('[ChatWindow] chatClient not open — attempting reconnect')
    try { chatClient.connect() } catch (e) {}
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
  try {
    console.log('[ChatWindow] mounting for merchantId=', props.merchantId, 'userBaseId=', props.userBaseId)
    await detectRole()
    await ensureNames()
    await loadHistory()
    // 当通过消息通知打开会话时，主动请求后端标记为已读并通知其它组件
    try {
      if (props.merchantId) {
        await request.post('/user/chats/mark_read', { merchant_id: Number(props.merchantId) })
        try { window.dispatchEvent(new CustomEvent('user:chats:marked_read', { detail: { merchant_id: Number(props.merchantId) } })) } catch(e) {}
      }
    } catch(e) { console.warn('[ChatWindow] mark_read failed', e) }
    chatClient.onMessage(handleGlobalMessage)
    // ensure centralized websocket connection
    try { chatClient.connect() } catch (e) { console.warn('[ChatWindow] chatClient.connect failed', e) }
    try { window.dispatchEvent(new CustomEvent('chat:window:mounted', { detail: { merchantId: props.merchantId, userBaseId: userBaseIdLocal.value } })) } catch (e) {}
    console.log('[ChatWindow] mounted')
  } catch (err) {
    console.error('[ChatWindow] mount error', err)
    try { window.dispatchEvent(new CustomEvent('chat:window:error', { detail: { error: String(err) } })) } catch (e) {}
  }
})

// 当外部传入的 merchantId / userBaseId 发生变化时，重新加载会话
watch(() => props.merchantId, async (newVal, oldVal) => {
  if (newVal === null || typeof newVal === 'undefined') return
  // 重置消息并重新加载
  messages.value = []
  await ensureNames()
  await loadHistory()
  // 如果当前用户打开的是会话，尝试通知后端标记为已读（用户端）
  try {
    await request.post('/user/chats/mark_read', { merchant_id: Number(newVal) })
  } catch (e) { console.warn('[ChatWindow] mark_read watcher failed', e) }
  // 通知其它组件当前用户会话已读（例如刷新用户会话列表）
  try { window.dispatchEvent(new CustomEvent('user:chats:marked_read', { detail: { merchant_id: Number(newVal) } })) } catch (e) {}
})

watch(() => props.userBaseId, async (newVal, oldVal) => {
  // 当 userBaseId 变化（例如商家端切换客户）时重新加载
  if (newVal === null || typeof newVal === 'undefined') return
  messages.value = []
  await ensureNames()
  await loadHistory()
})

// 当外部直接传入的商家名称/头像发生变化时，实时更新本地显示
watch(() => props.merchantName, (nv) => {
  if (nv) merchantName.value = nv
})
watch(() => props.merchantAvatar, (nv) => {
  if (nv) merchantAvatar.value = nv
})

onBeforeUnmount(() => {
  chatClient.offMessage(handleGlobalMessage)
})
</script>

<style scoped>
/* ====================== 外层布局 ====================== */
.chat-container {
  width: 400px !important;
  height: 700px !important;
  max-width: 400px !important;
  min-height: 700px !important;
  border: none; /* 去掉外层白色边框，统一由组件内部展示 */
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  background: transparent; /* 透明背景，内部区域负责白色卡片样式 */
  box-shadow: none;
}

/* 内部卡片，用于保留白色背景与阴影，避免依赖 dialog 外层样式 */
.chat-card {
  width: 100%;
  height: 100%;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 10px 30px rgba(0,0,0,0.12);
  overflow: hidden;
  display: flex;
  flex-direction: column;
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
  font-size: 16px;
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

.jd-logo {
  width: 28px;
  height: 28px;
  object-fit:contain;
  border-radius:4px;
}

/* ====================== 消息区域 ====================== */
.messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 12px;
  background: #f5f5f5;
  /* 解决 iPhone 底部安全区被遮挡 */
  padding-bottom: env(safe-area-inset-bottom, 20px);
  /* 确保消息区在固定高度卡片中展示滚动条 */
  -webkit-overflow-scrolling: touch;
}

/* 滚动条美化（可选） */
.messages::-webkit-scrollbar {
  width: 8px;
}
.messages::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.18);
  border-radius: 4px;
}

.message-row {
  display: flex;
  margin-bottom: 12px;
  align-items: flex-end;
  width: 100%;
}

/* 左侧（收到）为正常方向，右侧（我）反方向 */
.message-row:not(.me) { flex-direction: row; justify-content: flex-start }
.message-row.me { flex-direction: row-reverse; justify-content: flex-end }

/* 头像 */
.msg-avatar { width: 38px; height: 38px; border-radius: 50%; flex-shrink: 0 ;margin-bottom: 20px;}
.message-row:not(.me) .msg-avatar { margin-right: 0px }
.message-row.me .msg-avatar { margin-left: 0px }

/* 气泡容器 */
.bubble-wrapper {
  max-width: 72%;
  position: relative;
  display: block;
}

/* 气泡主体：根据内容宽度自适应，支持换行，使用相对定位以便三角形定位于气泡 */
.bubble {
  display: inline-block;
  position: relative;
  padding: 8px 12px;
  border-radius: 16px;
  font-size: 14px;
  line-height: 1.45;
  word-break: break-word;
  white-space: pre-wrap;
  max-width: 100%;
}

/* 左边（商家）气泡 - 白色 + 尖角 */
.message-row:not(.me) .bubble {
  background: #ffffff;
  border: 1px solid #ffffff;
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
  left: -6px;
  top: 12px;
  border: 8px solid transparent;
  border-right-color: #ffffff;
}

.message-row.me .bubble::before {
  content: "";
  position: absolute;
  right: -6px;
  top: 12px;
  border: 8px solid transparent;
  border-left-color: #ffe563;
}

/* 时间 */
.time {
  font-size: 11px;
  color: #999;
  margin-top: 6px;
  display: block;
  padding: 0 6px;
}

/* 时间对齐：跟随气泡 */
.message-row:not(.me) .bubble-wrapper { margin-right: auto; text-align: left }
.message-row.me .bubble-wrapper { margin-left: auto; text-align: right }
.message-row.me .time { text-align: right }
.message-row:not(.me) .time { text-align: left }

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
