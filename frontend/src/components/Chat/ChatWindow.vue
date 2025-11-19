<template>
  <div class="chat-window">
    <div class="messages" ref="msgWrap">
      <div v-for="m in messages" :key="m.id" class="message">
        <div class="meta">{{ m.from_base_id }}</div>
        <div class="content">{{ m.content }}</div>
        <div class="time">{{ formatTime(m.created_at) }}</div>
      </div>
    </div>
    <div class="input-area">
      <input v-model="input" @keyup.enter="send" placeholder="输入消息，回车发送" />
      <button @click="send">发送</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { getChatHistory, getWsUrl } from '@/api/chat'

const props = defineProps({
  merchantId: { type: Number, required: true },
  userBaseId: { type: Number, required: true },
  token: { type: String, required: true }
})

const messages = ref([])
const input = ref('')
let ws = null
const msgWrap = ref(null)

function formatTime(s) {
  return new Date(s).toLocaleString()
}

async function loadHistory() {
  try {
    const res = await getChatHistory(props.merchantId, props.userBaseId)
    if (res && res.data && res.data.data) {
      messages.value = res.data.data.reverse() // 按时间正序显示
      await nextTick()
      scrollBottom()
    }
  } catch (e) {
    console.error(e)
  }
}

function connectWs() {
  const url = getWsUrl() + `?token=${encodeURIComponent(props.token)}`
  ws = new WebSocket(url)
  ws.onopen = () => console.log('ws open')
  ws.onmessage = (ev) => {
    try {
      const data = JSON.parse(ev.data)
      messages.value.push(data)
      nextTick(scrollBottom)
    } catch (e) {
      console.error(e)
    }
  }
  ws.onclose = () => console.log('ws closed')
}

function send() {
  if (!input.value) return
  const payload = {
    merchant_id: props.merchantId,
    content: input.value,
    type: 'text'
  }
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(payload))
    input.value = ''
  }
}

function scrollBottom() {
  if (msgWrap.value) {
    msgWrap.value.scrollTop = msgWrap.value.scrollHeight
  }
}

onMounted(() => {
  loadHistory()
  connectWs()
})

onBeforeUnmount(() => {
  if (ws) ws.close()
})
</script>

<style scoped>
.chat-window { border: 1px solid #ddd; width: 320px; display:flex; flex-direction:column }
.messages { flex:1; padding:8px; overflow:auto; height:360px }
.message { margin-bottom:8px }
.meta { font-size:12px; color:#888 }
.content { background:#f5f5f5; padding:6px; border-radius:4px }
.time { font-size:11px; color:#aaa }
.input-area { display:flex; padding:8px }
.input-area input { flex:1; padding:6px }
.input-area button { margin-left:8px }
</style>
