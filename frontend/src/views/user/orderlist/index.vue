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
        <el-button :type="activeTab==='all'? 'warning':''" plain @click="setTab('all')">全部</el-button>
        <el-button :type="activeTab==='pending'? 'warning':''" plain @click="setTab('pending')">待付款</el-button>
        <el-button :type="activeTab==='shipping'? 'warning':''" plain @click="setTab('shipping')">待收货</el-button>
        <el-button :type="activeTab==='completed'? 'warning':''" plain @click="setTab('completed')">已完成</el-button>
        <el-button :type="activeTab==='refund'? 'warning':''" plain @click="setTab('refund')">退款/售后</el-button>
      </div>

      <!-- 列表 -->
      <div class="list">
        <OrderCard
          v-for="(o, idx) in filteredOrders"
          :key="o.id"
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
        <div v-if="filteredOrders.length === 0" class="empty">暂无订单</div>
      </div>
    </div>
  </div>
</template>


<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import OrderCard from '@/components/OrderList/OrderCard.vue'
import { useRouter } from 'vue-router'
import orderApi from '@/api/user/order'
import storeApi from '@/api/user/store'

// mock orders data
const rawOrders = ref([
  {
    id: 'ORD20251027001',
    storeId: 'S001',
    storeName: '川味小馆',
    storeLogo: '/src/assets/noImg.png',
    status: 'pending', // pending, shipping, completed, refund
    statusText: '待付款',
    time: '2025-10-27 11:20',
    payDeadline: new Date(Date.now() + 1000 * 60 * 15).toISOString(),
    items: [ { name: '宫保鸡丁', price: 28, count:1, img: '' }, { name: '米饭', price: 3, count:2, img: '' } ]
  },
  {
    id: 'ORD20251027002',
    storeId: 'S002',
    storeName: '鲜甜水果',
    storeLogo: '/src/assets/noImg.png',
    status: 'shipping',
    statusText: '配送中',
    time: '2025-10-26 18:05',
    items: [ { name: '水果拼盘', price: 56, count:1, img: '' } ]
  },
  {
    id: 'ORD20251026003',
    storeId: 'S003',
    storeName: '芝士工坊',
    storeLogo: '/src/assets/noImg.png',
    status: 'completed',
    statusText: '已完成',
    time: '2025-10-25 12:10',
    items: [ { name: '披萨(大)', price: 88, count:1, img: '' } ]
  },
  {
    id: 'ORD20251024004',
    storeId: 'S004',
    storeName: '退单示例店',
    storeLogo: '/src/assets/noImg.png',
    status: 'refund',
    statusText: '退款/售后',
    time: '2025-10-24 09:30',
    items: [ { name: '示例商品', price: 10, count:2, img: '' } ]
  }
])

const activeTab = ref('all')
const searchMode = ref(false)
const keyword = ref('')

const route = useRoute()

onMounted(()=>{
  // 如果路由带 oq 查询参数，作为初始关键字
  const oq = route.query.oq
  if (oq && typeof oq === 'string') keyword.value = oq
})

watch(()=>route.query.oq, (v)=>{ if (v && typeof v === 'string') keyword.value = v })

function setTab(t) { activeTab.value = t }
function toggleSearch() { searchMode.value = !searchMode.value }
function applySearch() { /* filter applied by computed */ }
function onClear() { keyword.value = '' }
function openNotices() { /* placeholder */ }

const filteredOrders = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  return rawOrders.value.filter(o => {
    if (activeTab.value !== 'all' && o.status !== activeTab.value) return false
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

function mapStatusText(status) {
  switch(status) {
    case 'pending': return '待付款'
    case 'shipping': return '待收货'
    case 'completed': return '已完成'
    case 'refund': return '退款/售后'
    default: return ''
  }
}

const router = useRouter()

// actions
function onPay(order) {
  // 跳转到支付页（占位），传递 orderId
  router.push({ path: '/user/pay', query: { orderId: order.id } })
}
async function onCancel(order) {
  try {
    await orderApi.cancelOrder(order.id)
  } catch (e) {
    // 后端不存在时直接修改本地 mock
  }
  order.status = 'cancelled'
  alert('已取消: ' + order.id)
}
function onConfirm(order) { order.status='completed'; alert('确认收货: ' + order.id) }
async function onReorder(order) {
  try {
    // 尝试通过后端 reorder 接口
    await orderApi.reorder(order)
  } catch (e) {
    // 如果后端不存在，使用 addToCart 逐条加入
    for (const it of order.items) {
      try { await storeApi.addToCart({ storeId: order.storeId, dishId: it.id || null, qty: it.count || 1 }) } catch(e){}
    }
  }
  alert('已加入购物车，前往购物车结算')
  router.push('/user/cart')
}
function onReview(order) { alert('去评价: ' + order.id) }
function onViewRefund(order) { alert('查看退款详情: ' + order.id) }
function openStore(id) { router.push({ name: 'userStore', params: { name: id } }) }
function onAutoCancel(order) { order.status = 'cancelled'; alert('支付超时，订单已取消：' + order.id) }

function onView(order) {
  // 跳转到订单详情页
  router.push({ path: `/user/order/${order.id}` })
}

// 把 mock 数据暴露以便 detail 页面回退使用（development only）
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
