<template>
  <div class="user-chat-list">
    <div class="list-header">会话</div>
    <ul>
      <li v-for="c in chats" :key="c.merchant_id" @click="open(c)" :class="{ active: active === c.merchant_id }">
        <img class="avatar" :src="c.merchantAvatar || '/imgs/merchant.png'" />
        <div class="meta">
          <div class="top">
            <div class="name">{{ c.merchantName || ('商家 ' + c.merchant_id) }}</div>
            <div class="time">{{ formatTime(c.last_at) }}</div>
          </div>
          <div class="bottom">
            <div class="last">{{ c.last_message }}</div>
            <div v-if="c.unread_count" class="badge">{{ c.unread_count }}</div>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import request from '@/api/merchant/request'
import { getMerchantDetail, getBaseUserDetail } from '@/api/chat'
import chatClient from '@/utils/chatClient'

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

  if (dt >= startOfToday) {
    return `今天 ${timePart}`
  }
  if (dt >= startOfYesterday) {
    return `昨天 ${timePart}`
  }
  if (dt.getFullYear() === now.getFullYear()) {
    return `${dt.getMonth() + 1}月${dt.getDate()}日 ${timePart}`
  }
  return `${dt.getFullYear()}年${dt.getMonth() + 1}月${dt.getDate()}日 ${timePart}`
}

let currentBaseId = null
async function load() {
  try {
    // 仅调用用户侧会话接口，移除本地 mock 回退
    const res = await request.get('/user/chats')

    if (res.data && res.data.code === 1) {
      // 兼容后端字段名（merchant_name / merchant_avatar）与前端旧字段
      const raw = res.data.data || []
      const list = raw.map((c) => ({
        merchant_id: c.merchant_id,
        merchantName: c.merchant_name || c.merchantName || null,
        merchantAvatar: c.merchant_avatar || c.merchantAvatar || null,
        last_message: c.last_message,
        last_at: c.last_at,
        unread_count: c.unread_count || 0,
        user_base_id: c.user_base_id || c.userBaseId || null,
      }))

      // 只有在后端未返回商家展示信息时，才去查询商家详情作为补充
      await Promise.all(list.map(async (c) => {
        if ((!c.merchantName || !c.merchantAvatar) && c.merchant_id) {
          try {
            const r = await getMerchantDetail(c.merchant_id)
            const md = r?.data?.data
            if (md) {
              c.merchantName = md.shop_name || md.shopName || c.merchantName
              c.merchantAvatar = md.logo || md.logoUrl || c.merchantAvatar
            }
          } catch (e) {
            // ignore enrichment errors
          }
        }
      }))

      chats.value = list
    } else {
      chats.value = []
    }
  } catch (e) {
    console.error('load user chats failed', e)
    // 请求失败时返回空列表（不再使用本地 mock）
    chats.value = []
  }
}

function open(c) {
  // debug: log item shape to help diagnose unexpected types
  try { console.log('[UserChatList.open] item:', c) } catch (e) {}

  // 读取 merchant id 时更加鲁棒：支持字段为 value / getter 函数 / 不同命名
  let mid = null
  try {
    if (c == null) mid = null
    else if (typeof c === 'object') {
      if (c.merchant_id !== undefined) {
        mid = (typeof c.merchant_id === 'function') ? c.merchant_id() : c.merchant_id
      } else if (c.merchantId !== undefined) {
        mid = (typeof c.merchantId === 'function') ? c.merchantId() : c.merchantId
      }
    }
  } catch (e) {
    console.warn('[UserChatList.open] read merchant id error', e)
  }

  active.value = mid
  // 标记为已读（通知后端把来自该商家的消息标记为 read）
  ;(async () => {
    try {
      if (mid) await request.post('/user/chats/mark_read', { merchant_id: mid })
    } catch (e) {
      console.warn('mark read failed', e)
    }
    // 立即刷新列表，保证未读数字实时消失
    await load()
    window.dispatchEvent(new CustomEvent('chat:open', { detail: { merchantId: mid || null, userBaseId: (c && (c.user_base_id || c.userBaseId)) || null } }))
  })()
}

onMounted(() => {
  // 获取当前登录者 id（用于判断哪些消息是发给我的）
  ;(async () => {
    try {
      const cur = await getBaseUserDetail()
      if (cur && cur.data && cur.data.data) currentBaseId = cur.data.data.id
    } catch (e) {}
    await load()
  })()
  // 定时刷新会话列表
  const timer = setInterval(load, 15000)

  // 注册全局 ws 通知：收到消息时如果是针对当前用户则刷新会话列表
  const handler = (msg) => {
    try {
      const uid = msg.user_base_id || msg.userBaseId
      if (!uid) return
      if (currentBaseId && Number(uid) === Number(currentBaseId)) {
        // 只要有新的消息，刷新会话列表以更新未读计数与排序
        load()
      }
    } catch (e) {}
  }
  chatClient.onMessage(handler)
  onBeforeUnmount(() => {
    clearInterval(timer)
    chatClient.offMessage(handler)
  })
})
</script>
<style scoped>
.user-chat-list {
  width: 100%;
  height: 100%;
  background: #fff;
  display: flex;
  flex-direction: column;
  border-radius: 12px;           /* 和外层卡片保持一致 */
  overflow: hidden;
}

/* 顶部标题栏 - 完全微信风格 */
.list-header {
  height: 50px;
  line-height: 50px;
  padding: 0 20px;
  font-size: 17px;
  font-weight: 600;
  color: #000;
  background: #f7f7f8;
  border-bottom: 0.5px solid #e5e5e5;
  flex-shrink: 0;
}

/* 列表区域 */
ul {
  list-style: none;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow-y: auto;
  /* 隐藏滚动条但仍可滚动（微信同款） */
  -ms-overflow-style: none;
  scrollbar-width: none;
}
ul::-webkit-scrollbar { display: none; }

/* 每一行 */
li {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  gap: 12px;
  cursor: pointer;
  position: relative;
  transition: background 0.2s;
}
li:hover,
li.active {
  background: #f0f0f0;           /* 微信选中/悬停色 */
}
li:not(:last-child)::after {
  content: '';
  position: absolute;
  left: 72px;
  right: 0;
  bottom: 0;
  height: 0.5px;
  background: #e5e5e5;
}

/* 头像 - 微信 52×52 圆角8px */
.avatar {
  width: 52px;
  height: 52px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}

/* 文字区域 */
.meta {
  flex: 1;
  min-width: 0;                  /* 重要：让文字能被截断 */
  display: flex;
  flex-direction: column;
  justify-content: center;
  height: 52px;
}
.top {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 6px;
}
.name {
  font-size: 17px;
  font-weight: 600;
  color: #000;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.time {
  font-size: 12px;
  color: #999;
}

/* 最后一条消息 + 未读 */
.bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.last {
  font-size: 14px;
  color: #888;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

/* 未读红点 - 完美还原微信 */
.badge {
  min-width: 20px;
  height: 20px;
  line-height: 20px;
  text-align: center;
  background: #f56c6c;
  color: #fff;
  font-size: 12px;
  border-radius: 10px;
  padding: 0 6px;
  flex-shrink: 0;
}
/* 未读数大于99时显示“99+” */
.badge:where([class*="9"], [class*="0"]) {
  padding: 0 4px;
}
</style>
