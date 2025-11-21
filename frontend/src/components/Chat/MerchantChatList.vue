<template>
  <div class="merchant-chat-list">
    <div class="list-header">会话列表</div>
    <ul>
      <li v-for="c in chats" :key="c.user_base_id" @click="open(c)" :class="{ active: active === c.user_base_id }">
        <div class="left">
          <div class="name">{{ c.userName || ('用户 ' + c.user_base_id) }}</div>
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
import { getBaseUserDetail } from '@/api/chat'

const chats = ref([])
const active = ref(null)

function formatTime(s) {
  if (!s) return ''
  const dt = new Date(s)
  const now = new Date()
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const startOfYesterday = new Date(startOfToday.getTime() - 24 * 3600 * 1000)
  const pad = (n) => String(n).padStart(2, '0')
  const timePart = `${pad(dt.getHours())}:${pad(dt.getMinutes())}`
  if (dt >= startOfToday) return `今天 ${timePart}`
  if (dt >= startOfYesterday) return `昨天 ${timePart}`
  if (dt.getFullYear() === now.getFullYear()) return `${dt.getMonth() + 1}月${dt.getDate()}日 ${timePart}`
  return `${dt.getFullYear()}年${dt.getMonth() + 1}月${dt.getDate()}日 ${timePart}`
}

async function load() {
  try {
    const res = await request.get('/merchant/chats')
    if (res.data && res.data.code === 1) {
      const list = res.data.data || []
      // enrich each chat with user display name (try to fetch base user detail)
      await Promise.all(list.map(async (c) => {
        try {
          const r = await getBaseUserDetail(c.user_base_id)
          const u = r?.data?.data
          c.userName = u?.username || u?.nickname || null
        } catch (e) { c.userName = null }
      }))
      // if empty, provide demo mock conversations for development
      if (!list || list.length === 0) {
        const now = Date.now()
        chats.value = [
          { user_base_id: 501, last_message: '您好，有什么可以帮到您？', last_at: new Date(now-1800*1000).toISOString(), unread_count: 3, merchant_id: null, userName: '张三' },
          { user_base_id: 502, last_message: '请问订单可以合并吗？', last_at: new Date(now-3600*24*2).toISOString(), unread_count: 0, merchant_id: null, userName: '李四' },
        ]
      } else {
        chats.value = list
      }
    }
  } catch (e) {
    console.log('load chats failed', e)
    // fallback mock data for development
    const now = Date.now()
    chats.value = [
      { user_base_id: 601, last_message: '示例：请尽快处理订单', last_at: new Date(now-500000).toISOString(), unread_count: 1, merchant_id: null, userName: '示例用户' }
    ]
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
