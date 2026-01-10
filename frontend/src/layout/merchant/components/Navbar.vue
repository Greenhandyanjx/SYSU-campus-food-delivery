<template>
  <header class="navbar">
    <!-- 左侧标题 -->
    <div class="navbar-left">
      <span class="title">中珠校园外卖 - 商家端</span>
    </div>

    <!-- 中间菜单 -->
    <div class="navbar-center">
      <el-menu
        mode="horizontal"
        :ellipsis="false"
        :default-active="activePath"
        background-color="transparent"
        text-color="#fff"
        active-text-color="#ffd04b"
        @select="handleSelect"
      >
        <el-menu-item index="/merchant/dashboard">工作台</el-menu-item>
        <el-menu-item index="/merchant/orders">订单管理</el-menu-item>
        <el-menu-item index="/merchant/menu">菜品管理</el-menu-item>
        <el-menu-item index="/merchant/statistics">数据统计</el-menu-item>
        <el-menu-item index="/merchant/meal">套餐管理</el-menu-item>
      </el-menu>
    </div>

    <!-- 右侧操作区 -->
    <div class="navbar-right">
      <!-- background bootstrapper keeps websocket alive -->
      <ChatBootstrap />

      <el-popover placement="bottom-end" width="360" trigger="click">
        <template #reference>
          <el-button type="text" style="display:flex;align-items:center;gap:6px;color:#fff">
            <el-badge :value="totalUnread" class="unread-badge" v-if="totalUnread > 0">
              <span style="font-size:18px">🔔</span>
            </el-badge>
            <span v-else style="font-size:18px">🔔</span>
          </el-button>
        </template>

        <div style="padding:6px 8px; max-height:60vh; overflow:auto; min-width:320px;">
          <div v-if="sessionList.length === 0" style="padding:12px;color:#666">暂无会话</div>
          <ul style="list-style:none;padding:0;margin:0">
            <li v-for="s in sessionList" :key="s.peerId" @click="openSession(s)" style="display:flex;justify-content:space-between;padding:8px;cursor:pointer;border-bottom:1px solid #f1f1f1">
              <div style="max-width:220px;overflow:hidden;white-space:nowrap;text-overflow:ellipsis">
                <div style="font-weight:600">{{ s.name }}</div>
                <div style="color:#666;font-size:13px">{{ s.lastMessage }}</div>
              </div>
              <div style="text-align:right;min-width:72px">
                <div style="font-size:12px;color:#999">{{ s.lastAtDisplay }}</div>
                <div v-if="s.unread" class="unread-dot" title="未读消息"></div>
              </div>
            </li>
          </ul>
        </div>
      </el-popover>

      <el-dropdown trigger="click" @command="handleCommand">
        <span class="el-dropdown-link" style="display:flex;align-items:center;gap:8px;cursor:pointer;background-color: #409EFF;">
          <el-avatar :size="35" :src="logoSrc" />
          <span class="username">{{ username || '用户' }}</span>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">个人信息</el-dropdown-item>
            <el-dropdown-item command="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue'
import ChatBootstrap from '@/components/Chat/ChatBootstrap.vue'
import merchantSvg from '@/assets/merchant.svg'
import { useRouter, useRoute } from 'vue-router'
import { useChatStore } from '@/stores/chatStore'
import { getMerchantProfile } from '@/api/merchant/profile'
import request from '@/api/merchant/request'

const router = useRouter()
const route = useRoute()

const activePath = ref(route.path)

watch(
  () => route.path,
  (newPath) => {
    activePath.value = newPath
  }
)

const handleSelect = (path: string) => {
  router.push(path)
}

const username = ref(localStorage.getItem('username') || '')
const chatStore = useChatStore()
const totalUnread = computed(() => {
  const s = chatStore.sessions || {}
  let c = 0
  for (const k in s) {
    const n = Number(s[k]?.unread || s[k]?.unread_count || 0)
    if (!isNaN(n)) c += n
  }
  return c
})

const sessionList = computed(() => {
  const arr = []
  const s = chatStore.sessions || {}
  for (const k in s) {
    const item = s[k]
    const msgs = (item && item.messages) || []
    const last = msgs.length ? msgs[msgs.length - 1] : null
    const lastMsg = last ? (last.content || '') : (item.meta && item.meta.last_message) || ''
    const lastAt = last ? (last.created_at || last.last_at || null) : (item.meta && item.meta.last_at) || null
    arr.push({ peerId: k, unread: Number(item.unread || 0), name: (item.meta && (item.meta.userName || item.meta.merchantName)) || ('用户 ' + k), lastMessage: lastMsg, lastAt: lastAt })
  }
  // sort by lastAt desc
  arr.sort((a, b) => {
    const ta = a.lastAt ? new Date(a.lastAt).getTime() : 0
    const tb = b.lastAt ? new Date(b.lastAt).getTime() : 0
    return tb - ta
  })
  // format display
  return arr.map(x => ({ ...x, lastAtDisplay: x.lastAt ? new Date(x.lastAt).toLocaleString() : '' }))
})

function openSession(s: any) {
  try {
    const merchantId = localStorage.getItem('merchantId') || null
    const detail = { merchantId: merchantId ? Number(merchantId) : null, userBaseId: Number(s.peerId) }
    window.dispatchEvent(new CustomEvent('chat:open', { detail }))
    // mark session as read locally so UI updates (remove unread dot)
    try { chatStore.markSessionRead(String(s.peerId)) } catch (e) {}
    // also notify backend and other listeners that this session was marked read
    (async () => {
      try {
        const mid = merchantId ? Number(merchantId) : null
        if (mid !== null) {
          await request.post('/merchant/chats/mark_read', { merchant_id: mid, user_base_id: Number(s.peerId) })
          try { window.dispatchEvent(new CustomEvent('merchant:chats:marked_read', { detail: { merchant_id: mid, user_base_id: Number(s.peerId) } })) } catch(e) {}
        }
      } catch (e) {
        console.warn('mark_read request failed', e)
      }
    })()
  } catch (e) { console.warn('openSession failed', e) }
}

// load chats from backend on mount so Navbar shows recent sessions after refresh
async function loadChats() {
  try {
    const res = await request.get('/merchant/chats')
    if (res && res.data && Number(res.data.code) === 1) {
      const list = res.data.data || []
      try {
        list.forEach((c: any) => {
          if (c && (c.user_base_id !== undefined && c.user_base_id !== null)) {
            chatStore.upsertSession(String(c.user_base_id), { unread: c.unread_count || 0, meta: { userName: c.userName || (c.user_name || null), last_message: c.last_message, last_at: c.last_at } })
          }
        })
      } catch (e) {}
    }
  } catch (e) {
    // ignore load errors; Navbar can function without backend
    console.warn('Navbar.loadChats failed', e)
  }
}

const logoSrc = ref(merchantSvg)

onMounted(async () => {
  try {
    const r: any = await getMerchantProfile()
      if (r && r.data && r.data.data) {
      const d = r.data.data
      logoSrc.value = d.logo || d.logoUrl || merchantSvg
    }
    // ensure we have recent sessions in chatStore so navbar shows them after refresh
    try { await loadChats() } catch(e) {}
  } catch (e) {}
})

const handleCommand = (command: string) => {
  if (command === 'logout') {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/merchant/profile')
  }
}

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('username')
  router.push('/login')
}
</script>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 30px;
  background-color: #409eff;
  color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.navbar-left .title {
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}
.navbar-center {
  flex: 1;
  display: flex;
  justify-content: center;
  overflow: visible !important;
  white-space: nowrap;
  min-width: 0; 
}


.navbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: 30px;
}

.navbar-right .el-button {
  font-size: 15px;
  font-weight: 500;
  color: #fff !important;
  background-color: transparent !important;
  border: none;
  border-radius: 0;
  height: 56px;
  padding: 0 20px;
  transition: background-color 0.3s ease;
}
.navbar-right .el-button:hover {
  background-color: rgba(255, 255, 255, 0.15) !important;
  color: #fff !important;
}

.unread-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #f56c6c;
  margin-top: 6px;
}

/* 修复 el-avatar 默认白色背景/图片未铺满导致的左侧白色轮廓问题 */
::v-deep(.el-avatar) {
  background-color: transparent !important;
  border-radius: 50% !important;
  overflow: hidden !important;
}
::v-deep(.el-avatar__inner) {
  background-color: transparent !important;
  display: block !important;
  width: 100% !important;
  height: 100% !important;
}
::v-deep(.el-avatar__inner img) {
  width: 100% !important;
  height: 100% !important;
  object-fit: cover !important;
  display: block !important;
}


::v-deep(.el-menu.el-menu--horizontal) {
  flex-shrink: 0 !important;
  border-bottom: none;
  background-color: transparent !important;
  overflow: visible !important;
}

::v-deep(.el-menu-item) {
  font-size: 15px;
  font-weight: 500;
  padding: 0 22px !important;
  overflow: visible !important;
  text-overflow: unset !important;
}

/* 鼠标悬停颜色 */
::v-deep(.el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.2) !important;
}
</style>
