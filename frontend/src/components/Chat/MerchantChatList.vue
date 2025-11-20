<template>
  <div class="merchant-chat-list">
    <div class="list-header">会话列表</div>
    <ul>
      <li v-for="c in chats" :key="c.user_base_id" @click="open(c)" :class="{ active: active === c.user_base_id }">
        <div class="left">
          <div class="name">用户 {{ c.user_base_id }}</div>
          <div class="last">{{ c.last_message }}</div>
        </div>
        <div class="right">
          <span class="time">{{ formatTime(c.last_at) }}</span>
          <span v-if="c.unread_count" class="badge">{{ c.unread_count }}</span>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import request from '@/api/merchant/request'

const chats = ref([])
const active = ref(null)

function formatTime(s) {
  if (!s) return ''
  const d = new Date(s)
  return `${d.getHours()}:${String(d.getMinutes()).padStart(2,'0')}`
}

async function load() {
  try {
    const res = await request.get('/merchant/chats')
    if (res.data && res.data.code === 1) {
      chats.value = res.data.data
    }
  } catch (e) {
    console.log('load chats failed', e)
  }
}

function open(c) {
  active.value = c.user_base_id
  window.dispatchEvent(new CustomEvent('chat:open', { detail: { merchantId: c.merchant_id || null, userBaseId: c.user_base_id } }))
}

onMounted(() => {
  load()
  // poll every 15s
  setInterval(load, 15000)
})
</script>

<style scoped>
.merchant-chat-list { border:1px solid #eee; border-radius:8px; padding:8px; width:320px; background:#fff }
.merchant-chat-list .list-header { font-weight:700; margin-bottom:8px }
.merchant-chat-list ul { list-style:none; padding:0; margin:0 }
.merchant-chat-list li { display:flex; justify-content:space-between; padding:8px; cursor:pointer; border-radius:6px }
.merchant-chat-list li.active { background:#f6f8fa }
.merchant-chat-list .left { max-width:220px }
.merchant-chat-list .name { font-weight:600 }
.merchant-chat-list .last { color:#666; font-size:13px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis }
.merchant-chat-list .badge { background:#f56c6c; color:#fff; padding:2px 6px; border-radius:12px; font-size:12px }
.merchant-chat-list .time { font-size:12px; color:#999; margin-right:8px }
</style>
