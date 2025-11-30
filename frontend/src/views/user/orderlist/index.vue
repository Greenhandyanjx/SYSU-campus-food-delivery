<template>
  <div class="orders-bg">
    <div class="orders-page">
      <!-- 顶部栏 -->
      <div class="topbar">
        <div class="left">
          <h2>我的订单</h2>
        </div>
        <div class="right">
          <template v-if="!searchMode">
            <i class="icon-bell" @click="openNotices" title="通知">🔔</i>
          </template>
          <template v-else>
            <el-input
              v-model="keyword"
              placeholder="搜索订单号 / 店铺 / 商品"
              clearable
              @clear="onClear"
              @keyup.enter="applySearch"
              size="small"
              class="search-input"
            />
          </template>
          <el-button plain size="small" class="search-toggle" @click="toggleSearch">
            {{ searchMode ? '取消' : '搜索订单' }}
          </el-button>
        </div>
      </div>

      <!-- 标签页 -->
      <div class="tabs">
        <el-button :type="activeTab===0? 'warning':''" plain @click="setTab(0)">全部</el-button>
        <el-button :type="activeTab===1? 'warning':''" plain @click="setTab(1)">待付款</el-button>
        <el-button :type="activeTab===2? 'warning':''" plain @click="setTab(2)">待接单</el-button>
        <el-button :type="activeTab===3? 'warning':''" plain @click="setTab(3)">待派送</el-button>
        <el-button :type="activeTab===4? 'warning':''" plain @click="setTab(4)">派送中</el-button>
        <el-button :type="activeTab===5? 'warning':''" plain @click="setTab(5)">已完成</el-button>
        <el-button :type="activeTab===6? 'warning':''" plain @click="setTab(6)">已取消</el-button>
      </div>

      <!-- 列表 -->
      <div class="list">
        <div v-for="(o, idx) in filteredOrders" :key="o.id" class="order-row">
          <OrderCard
            :order="mapOrder(o)"
            @pay="onPay"
            @cancel="onCancel"
            @confirm="onConfirm"
            @reorder="onReorder"
            @review="onReview"
            @view="onView"
            @view-refund="onViewRefund"
            @open-store="openStore"
            @auto-cancel="onAutoCancel"
          />
          <!-- 聊天入口已迁移到 OrderCard（ChatLauncher），此处不再需要单独按钮 -->
        </div>
        <div v-if="filteredOrders.length === 0" class="empty">暂无订单</div>
      </div>
    </div>
  </div>

  <!-- Chat modal moved into OrderCard via ChatLauncher -->
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import OrderCard from '@/components/OrderList/OrderCard.vue'
import orderApi from '@/api/user/order'
import storeApi from '@/api/user/store'
// Chat handled inside OrderCard via ChatLauncher component

const route = useRoute()
const router = useRouter()

const searchMode = ref(false)
const keyword = ref('')
const activeTab = ref(0)

// Chat state moved into OrderCard

// orders loaded from backend
const rawOrders = ref([])

onMounted(()=>{
  const oq = route.query.oq
  if (oq && typeof oq === 'string') keyword.value = oq
  loadOrders()
})

watch(()=>route.query.oq, (v)=>{ if (v && typeof v === 'string') keyword.value = v })

function setTab(t) { activeTab.value = t; loadOrders() }
async function loadOrders(page = 1, size = 20) {
  const params = { page, size }
  // activeTab uses numeric status codes (0 = all)
  if (activeTab.value && activeTab.value !== 0) params.status = activeTab.value
  try {
    const res = await orderApi.getOrderList(params)
    const payload = res && res.data && (res.data.data || res.data)
    const items = (payload && payload.items) ? payload.items : (res && res.data && res.data.items) || []
    rawOrders.value = items.map(mapBackendOrder)
  } catch (e) {
    console.error('加载订单失败', e)
    rawOrders.value = []
  }
}
function toggleSearch() { searchMode.value = !searchMode.value }
function applySearch() { loadOrders(1) }
function onClear() { keyword.value = '' }
function openNotices() { /* placeholder */ }

const filteredOrders = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  return rawOrders.value.filter(o => {
    if (activeTab.value !== 0 && o.status !== activeTab.value) return false
    if (!k) return true
    return (o.id && o.id.toLowerCase().includes(k)) || (o.storeName && o.storeName.toLowerCase().includes(k)) || (o.items && o.items.some(it=> (it.name||'').toLowerCase().includes(k)))
  })
})

function mapOrder(o) {
  return {
    ...o,
    statusText: o.statusText || mapStatusText(o.status)
  }
}

function mapBackendOrder(o) {
  // try common field mappings, be defensive about backend shape
  const items = []
  const sourceItems = o.items || o.order_dishes || o.orderDishes || []
  if (Array.isArray(sourceItems)) {
    for (const it of sourceItems) {
      items.push({
        id: it.id || it.dish_id || it.dishId || null,
        name: it.name || it.dish_name || it.title || '',
        price: it.price || it.unit_price || it.amount || 0,
        count: it.qty || it.count || it.quantity || 1,
        img: it.image || it.img || it.picture || ''
      })
    }
  }

  const statusNum = Number(o.status || o.order_status || 0)
  function formatFriendlyTime(iso) {
    if (!iso) return ''
    const d = new Date(iso)
    if (isNaN(d.getTime())) return iso
    const M = d.getMonth() + 1
    const D = d.getDate()
    const hh = String(d.getHours()).padStart(2, '0')
    const mm = String(d.getMinutes()).padStart(2, '0')
    return `${M}月${D}日 ${hh}:${mm}`
  }

  const rawTime = o.created_at || o.time || o.createdAt || ''
  return {
    id: o.id || o.order_no || o.orderNo || '',
    storeId: o.store_id || o.storeId || o.merchant_id || o.merchantId || '',
    storeName: o.store_name || o.storeName || o.shop_name || '',
    storeLogo: o.store_logo || o.logo || '/src/assets/noImg.png',
    status: statusNum,
    statusText: o.status_text || o.statusText || mapStatusText(statusNum),
    time: formatFriendlyTime(rawTime),
    payDeadline: o.pay_deadline || o.payDeadline || null,
    // delivery fee: support multiple backend field names
    delivery_fee: o.delivery_fee ?? o.deliveryFee ?? o.deliveryAmount ?? o.delivery ?? o.fee ?? 0,
    // keep legacy aliases for safety
    deliveryFee: o.delivery_fee ?? o.deliveryFee ?? o.deliveryAmount ?? o.delivery ?? o.fee ?? 0,
    items
  }
}

function mapStatusText(status) {
  const s = Number(status)
  switch (s) {
  case 1:
    return '待付款'
  case 2:
    return '待接单'
  case 3:
    return '待派送'
  case 4:
    return '派送中'
  case 5:
    return '已完成'
  case 6:
    return '已取消'
  default:
    return ''
  }
}

// actions
function onPay(order) {
  try {
    // 直接复用已有订单：写入 pending_orders，结算页会优先使用该项并直接支付
    const oid = order && (order.id || order.ID || order.orderId)
    if (oid) {
      try { sessionStorage.setItem('pending_orders', JSON.stringify([String(oid)])) } catch (e) {}
    }
    router.push('/user/payment/confirm')
  } catch (e) {
    console.warn('prepare checkout payload from order failed', e)
    // fallback: navigate to confirm page without payload
    router.push('/user/payment/confirm')
  }
}
async function onCancel(order) {
  try { await orderApi.cancelOrder(order.id) } catch (e) {}
  // set numeric cancelled status
  order.status = 6
  alert('已取消: ' + order.id)
}
function onConfirm(order) { order.status='completed'; alert('确认收货: ' + order.id) }
async function onReorder(order) {
  try { await orderApi.reorder(order) } catch (e) {
    for (const it of order.items) {
      try { await storeApi.addToCart({ storeId: order.storeId, dishId: it.id || null, qty: it.count || 1 }) } catch(e){}
    }
  }
  alert('已加入购物车，前往购物车结算')
  router.push('/user/cart')
}
function onReview(order) { alert('去评价: ' + order.id) }
function onViewRefund(order) { alert('查看退款详情: ' + order.id) }
function openStore(id) { 
  // prefer path-based navigation using numeric id to avoid relying on 'name' param
  if (id === undefined || id === null) {
    // fallback: try to navigate by raw id otherwise do nothing
    console.warn('openStore called with undefined id')
    return
  }
  router.push({ path: `/user/store/${id}` })
}
async function onAutoCancel(order) {
  try { await orderApi.cancelOrder(order.id) } catch (e) {}
  order.status = 6
  alert('支付超时，订单已取消：' + order.id)
}
function onView(order) { router.push({ path: `/user/order/${order.id}` }) }
// openChat moved to ChatLauncher inside OrderCard

// async function openChat(order) {
//   // 优先使用订单内的 merchantId 字段（后端真实字段）
//     // helper: try to coerce/resolve a merchant id to a number
//     async function resolveMerchantId(candidate) {
//       if (!candidate && candidate !== 0) return null
//       // numeric already?
//       const n = Number(candidate)
//       if (Number.isFinite(n) && String(n) !== 'NaN') return n
//       // fallback: if candidate looks like a store code, try storeApi.getStoreById
//       try {
//         const s = await storeApi.getStoreById(candidate)
//         const storeData = s && s.data && (s.data.data || s.data)
//         if (storeData && (storeData.merchantid || storeData.merchantId)) {
//           return Number(storeData.merchantid || storeData.merchantId)
//         }
//       } catch (e) {}
//       return null
//     }

//     // 优先使用订单内的 merchantId 字段（后端真实字段）
//     if (order.merchantId) {
//       const resolved = await resolveMerchantId(order.merchantId)
//       if (resolved == null) {
//         // cannot resolve merchant numeric id — abort with notice
//         alert('无法解析商家 ID，请联系管理员')
//         return
//       }
//       chatMerchantId.value = resolved
//     // 尝试把店铺/商家名传给聊天窗口
//       try {
//         const r = await getMerchantDetail(chatMerchantId.value)
//         if (r && r.data && r.data.data) {
//           chatMerchantName.value = r.data.data.shop_name || r.data.data.shopName || ''
//           chatMerchantAvatar.value = r.data.data.logo || r.data.data.logoUrl || ''
//         }
//       } catch (e) {}
//     // 获取当前用户信息以便传入 ChatWindow
//       try {
//         const u = await getBaseUserDetail()
//         if (u && u.data && u.data.data) {
//           chatUserId.value = u.data.data.id
//           chatUserName.value = u.data.data.username
//         }
//       } catch (e) {}
//     showChat.value = true
//     return
//   }

//   // 如果订单没有 merchantId，尝试通过订单详情或 storeId 查询
//     if (order.id) {
//       try {
//         const res = await orderApi.getOrderDetail(order.id)
//         const data = res && res.data && res.data.data
//         if (data && data.merchantid) {
//           const resolved = await resolveMerchantId(data.merchantid)
//           if (resolved == null) {
//             alert('无法解析商家 ID，请联系管理员')
//             return
//           }
//           chatMerchantId.value = resolved
//           try {
//             const r = await getMerchantDetail(chatMerchantId.value)
//             if (r && r.data && r.data.data) {
//               chatMerchantName.value = r.data.data.shop_name || r.data.data.shopName || ''
//               chatMerchantAvatar.value = r.data.data.logo || r.data.data.logoUrl || ''
//             }
//           } catch (e) {}
//         }
//       } catch (e) {}
//   }

//   // 回退：尝试使用 storeId -> 通过 storeApi.getStoreById（如后端实现）
//     if (order.storeId) {
//       try {
//         const s = await storeApi.getStoreById(order.storeId)
//         const storeData = s && s.data && (s.data.data || s.data)
//         if (storeData) {
//           // 如果后端返回 merchantId，使用它
//           if (storeData.merchantid || storeData.merchantId) {
//             const resolved = await resolveMerchantId(storeData.merchantid || storeData.merchantId)
//             if (resolved == null) {
//               alert('无法解析商家 ID，请联系管理员')
//               return
//             }
//             chatMerchantId.value = resolved
//           }
//           chatMerchantName.value = storeData.shop_name || storeData.name || ''
//           chatMerchantAvatar.value = storeData.logo || storeData.logoUrl || ''
//           try {
//             const u = await getBaseUserDetail()
//             if (u && u.data && u.data.data) {
//               chatUserId.value = u.data.data.id
//               chatUserName.value = u.data.data.username
//             }
//           } catch (e) {}
//           showChat.value = true
//           return
//         }
//       } catch (e) {}
//     }

//     // 最后回退：无法解析到 numeric merchant id，提示并返回
//     alert('无法定位商家 ID，无法发起聊天')
//     return
//   }

  // expose mock data in dev
  if (typeof window !== 'undefined') window.__RAW_ORDERS__ = rawOrders.value

</script>

<style scoped>
/* 背景层 */
.orders-bg {
  width: 100%;
  min-height: 100vh;
  background: url('/src/assets/login/img_denglu_bj.jpg') center/cover no-repeat;
  background-attachment: fixed; /* ✅ 背景不随滚动 */
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 60px 0;
}

/* 主体内容容器 */
.orders-page {
  width: 60%;
  background: rgba(255, 248, 225, 0.96);
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(255, 193, 7, 0.35);
  padding: 28px;
  backdrop-filter: blur(6px);
  transition: 0.3s;
}

.orders-page:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 28px rgba(255, 193, 7, 0.45);
}

/* 顶部栏 */
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0 16px;
  border-bottom: 2px solid #ffe58f;
}

.topbar .left h2 {
  margin: 0;
  color: #b8860b;
  font-weight: bold;
}

.topbar .right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-bell {
  font-size: 20px;
  cursor: pointer;
  transition: 0.2s;
}
.icon-bell:hover {
  transform: scale(1.2);
  color: #faad14;
}

/* 搜索框美化 */
.search-input :deep(.el-input__wrapper) {
  background: #fffdf5;
  border: 1.5px solid #faad14;
  border-radius: 20px;
  box-shadow: 0 2px 6px rgba(250, 173, 20, 0.2);
}

.search-toggle {
  border-radius: 20px;
  color: #ad8b00;
  border-color: #faad14;
  background: #fffdf5;
}
.search-toggle:hover {
  background: #faad14;
  color: white;
}

/* 标签按钮 */
.tabs {
  display: flex;
  gap: 10px;
  margin: 16px 0;
  flex-wrap: wrap;
}

.tabs .el-button {
  border-radius: 20px;
  font-weight: 500;
  transition: 0.2s;
}

.tabs .el-button--warning {
  background: #faad14;
  color: #fff;
}

/* 列表区域 */
.list {
  min-height: 300px;
  margin-top: 10px;
}

.empty {
  padding: 60px;
  text-align: center;
  color: #bfbfbf;
  font-size: 15px;
}

/* 按钮状态统一黄色主题 */
.el-button--plain[aria-pressed="true"],
.el-button--plain.is-active {
  color: #ff9800;
}
</style>
