<template>
  <div class="my-bg">
    <div class="my-page">
    <!-- 顶部个人信息卡片 -->
    <div class="profile-card modern">
        <div class="profile-cover"></div>
        <div class="profile-content">
          <div class="left">
            <el-avatar size="92" class="avatar" :src="''">
              <!-- fallback icon -->
            </el-avatar>
            <div class="info">
              <div class="name">{{ username }}</div>
              <div class="meta">会员：<span class="vip">普通用户</span> · 积分 <strong>{{ points }}</strong></div>
              <div class="quick-links">
                <div class="badge"><img src="/src/assets/icons/vip.svg" alt="vip" @error="onImgError"/> 会员中心</div>
                <div class="badge"><img src="/src/assets/icons/points.svg" alt="points" @error="onImgError"/> 积分</div>
              </div>
            </div>
          </div>
          <div class="right">
            <div class="summary">
              <div class="s-item" @click="go('orders')">
                <img src="/src/assets/icons/orders.svg" alt="orders" @error="onImgError" />
                <div class="s-text">全部订单<span class="count">{{ orderCount }}</span></div>
              </div>
              <div class="s-item" @click="go('coupons')">
                <img src="/src/assets/icons/coupons.svg" alt="coupons" @error="onImgError" />
                <div class="s-text">我的优惠券<span class="count">{{ couponCount }}</span></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 功能网格 -->
      <div class="feature-grid">
        <div class="f-item" @click="go('orders')">
          <img src="/src/assets/icons/orders.svg" alt="orders" @error="onImgError" />
          <div>我的订单</div>
        </div>
        <div class="f-item" @click="go('wallet')">
          <img src="/src/assets/icons/wallet.svg" alt="wallet" @error="onImgError" />
          <div>钱包</div>
        </div>
        <div class="f-item" @click="go('coupons')">
          <img src="/src/assets/icons/coupons.svg" alt="coupons" @error="onImgError" />
          <div>优惠券</div>
        </div>
        <div class="f-item" @click="go('address')">
          <img src="/src/assets/icons/address.svg" alt="address" @error="onImgError" />
          <div>我的地址</div>
        </div>
        <div class="f-item" @click="go('support')" style="position:relative">
          <img src="/src/assets/icons/support.svg" alt="support" @error="onImgError" />
          <div>客服与帮助</div>
          <span v-if="unreadSupport" class="support-badge">{{ unreadSupport }}</span>
        </div>
        <div class="f-item" @click="go('settings')">
          <img src="/src/assets/icons/settings.svg" alt="settings" @error="onImgError" />
          <div>设置</div>
        </div>
        <div class="f-item" @click="go('vip')">
          <img src="/src/assets/icons/vip.svg" alt="vip" @error="onImgError" />
          <div>会员中心</div>
        </div>
        <div class="f-item danger" @click="go('logout')">
          <img src="/src/assets/icons/logout.svg" alt="logout" @error="onImgError" />
          <div>退出登录</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import * as myApi from '@/api/user/my'
import request from '@/api/merchant/request'
import chatClient from '@/utils/chatClient'

const router = useRouter()
const username = ref(localStorage.getItem('username') || '游客')
const points = ref(0)
const orderCount = ref(0)
const couponCount = ref(0)
const unreadSupport = ref(0)

onMounted(async () => {
  const p = await myApi.getProfile()
  username.value = p.username || username.value
  points.value = p.points || 0
  orderCount.value = p.orderCount || 0
  couponCount.value = p.couponCount || 0
  // 获取用户未读会话总数
  try {
    const r = await request.get('/user/chats')
    if (r?.data?.code === 1) {
      const arr = r.data.data || []
      unreadSupport.value = arr.reduce((s, it) => s + (it.unread_count || 0), 0)
    }
  } catch (e) {}
  // 订阅 ws，实时更新未读总数
  const handler = (msg) => {
    try {
      // 若收到针对当前用户的消息，重新计算未读总数
      const uid = msg.user_base_id || msg.userBaseId
      if (!uid) return
      // 触发重新加载
      request.get('/user/chats').then(rr => {
        if (rr?.data?.code === 1) {
          const arr = rr.data.data || []
          unreadSupport.value = arr.reduce((s, it) => s + (it.unread_count || 0), 0)
        }
      }).catch(()=>{})
    } catch(e) {}
  }
  chatClient.onMessage(handler)
  // 移除监听在组件卸载时
  window.addEventListener('beforeunload', () => chatClient.offMessage(handler))
})

async function go(key: string) {
  // Map logical keys to routes
  const map: Record<string, string> = {
    orders: '/user/orderlist',
    wallet: '/user/wallet',
    coupons: '/user/coupons',
    address: '/user/address',
    support: '/user/support',
    settings: '/user/settings',
    vip: '/user/vip',
    // fallback
  }

  if (key === 'logout') {
    // perform logout via API (demo-friendly) then clear local state and route to login
    try {
      await myApi.logout()
    } catch (e) {}
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    router.push({ path: '/login' })
    return
  }

  const path = map[key]
  if (path) {
    router.push({ path })
  } else {
    // unknown key — push to home as safe fallback
    router.push({ path: '/user/home' })
  }
}

function onImgError(e: any) {
  try { e.target && (e.target.src = '/src/assets/noImg.png') } catch (e) {}
}
</script>

<style scoped>
/* === 页面整体背景（改为单图 center/cover，与订单页/主页一致） === */
.my-bg {
  width: 100%;
  min-height: 100vh;
  background: url('/src/assets/login/img_denglu_bj.jpg') center/cover no-repeat;
  background-attachment: fixed;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 60px 0;
}

/* 主卡片样式：与订单页/主页保持一致的悬停与阴影 */
.my-page {
  width: 60%;
  background: rgba(255, 248, 225, 0.96);
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(255, 193, 7, 0.35);
  padding: 28px;
  backdrop-filter: blur(6px);
  transition: 0.3s;
  position: relative;
  z-index: 2;
}

.my-page:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 28px rgba(255, 193, 7, 0.45);
}

/* === 顶部个人资料卡 === */

.profile-card.modern {
  position: relative;
  /* border-radius:   ; */
  overflow: hidden;
  margin-bottom: 20px;
  box-shadow: 0 10px 30px rgba(255, 180, 20, 0.15);
  background: linear-gradient(180deg, rgba(255,250,235,0.95), rgba(255,245,225,0.95));
}

.profile-cover {
  height: 110px;
  background: linear-gradient(135deg, #ffd666, #f7b500);
  filter: saturate(1.05);
}

.profile-content { display: flex; gap: 18px; padding: 18px; align-items: center; }

.profile-card .left { display:flex; gap:14px; align-items:center }
.profile-card .avatar { box-shadow: 0 6px 20px rgba(255, 170, 0, 0.18); border: 3px solid #fff; background: #fff7e6 }
.profile-card .info { text-align: left }
.profile-card .info .name { font-weight:700; color:#4a2c00; font-size:18px }
.profile-card .info .meta { color:#8c6b00; margin-top:6px; font-size:13px }
.profile-card .info .quick-links { display:flex; gap:8px; margin-top:10px }
.profile-card .info .badge { display:flex; gap:6px; align-items:center; padding:6px 8px; background:#fffef4; border-radius:8px; box-shadow:0 2px 6px rgba(250,173,20,0.06); font-size:12px }
.profile-card .info .badge img { width:18px; height:18px }

.profile-card .right { margin-left:auto }
.profile-card .summary { display:flex; gap:12px }
.profile-card .s-item { display:flex; gap:8px; align-items:center; background:#fff; padding:8px 12px; border-radius:10px; cursor:pointer }
.profile-card .s-item img { width:26px; height:26px }
.profile-card .s-text { font-size:13px; color:#6b3f00 }
.profile-card .count { display:block; font-weight:700; color:#ff6b6b; margin-top:4px }

.feature-grid { display:grid; grid-template-columns: repeat(4, 1fr); gap:12px; margin-top:18px }
.f-item { background:#fffdf6; border-radius:12px; padding:14px 12px; display:flex; flex-direction:column; align-items:center; gap:8px; cursor:pointer; transition:transform .12s, box-shadow .12s }
.f-item img { width:34px; height:34px }
.f-item:hover { transform:translateY(-6px); box-shadow:0 10px 24px rgba(250,173,20,0.12) }
.f-item.danger { background:#fff6f4 }
.f-item.danger:hover { box-shadow:0 10px 24px rgba(255,120,100,0.12) }
.support-badge {
  position: absolute;
  top: 6px;
  right: 8px;
  background: #f56c6c;
  color: #fff;
  min-width: 20px;
  height: 20px;
  line-height: 20px;
  font-size: 12px;
  border-radius: 10px;
  text-align: center;
  padding: 0 6px;
}

@media(max-width:900px){ .feature-grid { grid-template-columns: repeat(2, 1fr) } }

.info .name {
  font-size: 20px;
  font-weight: 700;
  color: #4a2c00;
}

.info .sub {
  font-size: 14px;
  color: #8c6b00;
}

/* === 菜单卡片 === */
.options {
  background: rgba(255, 255, 255, 0.95);
  border: none;
  border-radius: 12px;
  box-shadow: 0 4px 14px rgba(255, 204, 0, 0.2);
}

.option-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px;
  border-radius: 10px;
  background: #fffdf5;
  cursor: pointer;
  transition: all 0.25s ease;
  color: #5b3b00;
  font-weight: 500;
}

.option-item:hover {
  background: #fff6d6;
  box-shadow: 0 2px 10px rgba(255, 204, 0, 0.25);
  transform: translateY(-1px);
}

.option-item i {
  font-size: 18px;
  color: #ff9900;
}

/* 响应式 */
@media(max-width:900px){
  .my-page { width: 92%; padding: 12px; }
  .option-item { padding: 12px; font-size: 14px; }
  .info .name { font-size: 18px; }
}
</style>
