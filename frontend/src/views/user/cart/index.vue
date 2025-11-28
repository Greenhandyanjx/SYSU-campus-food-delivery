<template>
  <div class="cart-bg">
    <div class="cart-page">
    <div class="cart-top">
      <div class="categories-bar">
        <button :class="['cat-btn', { active: activeCategory === '全部' }]" @click="setCategory('全部')">全部</button>
        <button v-for="(c, idx) in categories" :key="idx" :class="['cat-btn', { active: activeCategory === c }]" @click="setCategory(c)">{{ c }}</button>
      </div>
      <div class="manage-area">
        <el-button size="small" type="primary" plain @click="toggleManage">{{ manageMode ? '退出管理' : '管理' }}</el-button>
      </div>
    </div>

    <div class="cart-list">
      <div v-for="(shop, sIdx) in visibleShops" :key="shop.storeId || shop.id || shop.merchant_id" class="shop-card">
        <div class="shop-header">
                <el-checkbox v-model="shop.selected" @change="onToggleShop(shop)" />
                <div class="shop-name" @click="goStore(shop)">
                  <img class="shop-logo" :src="shop.logo || '/src/assets/noImg.png'" @error="onImgError" />
                  {{ shop.name || shop.merchant_name }}
                </div>
              </div>

        <div class="shop-items">
          <div v-for="(it, iIdx) in shop.items.filter(it => showItemByCategory(it))" :key="it.dish_id" class="item-row">
            <div class="item-left">
              <el-checkbox v-model="it.selected" @change="onToggleItem(shop, it)" />
            </div>
            <div class="item-thumb">
              <img :src="it.img || '/src/assets/noImg.png'" @error="onImgError" alt="dish" />
            </div>
            <div class="item-mid" @click="goStore(shop)">
              <div class="item-name">{{ it.name }}</div>
              <div class="item-price">¥{{ it.price.toFixed(2) }}</div>
            </div>
            <div class="item-right">
              <el-button size="small" circle @click="decQty(shop, it)">-</el-button>
              <span class="qty">{{ it.qty }}</span>
              <el-button size="small" circle @click="incQty(shop, it)">+</el-button>
            </div>
          </div>
        </div>
        <!-- 取消的商品卡片（只读）：当当前数量少于原始数量且未处于管理模式时显示 -->
        <div v-if="showCanceled(shop)" class="canceled-card">
          <div class="canceled-header">已取消商品</div>
          <div class="canceled-items">
            <div v-for="ci in canceledItems(shop)" :key="ci.dish_id" class="canceled-row">
              <div class="canceled-name">{{ ci.name }}</div>
              <div class="canceled-count">已取消: {{ ci.canceledQty }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
    </div>
  </div>

    <!-- 底部固定栏 -->
  <div class="cart-bottom">
    <div class="cart-bottom-inner">
      <div class="left">
        <el-checkbox v-model="selectAllChecked" @change="onSelectAllChange" />
        <span class="all-text">全选</span>
      </div>
      <div class="center">合计: <span class="total">¥{{ totalPrice.toFixed(2) }}</span></div>
      <div class="right">
        <el-button type="danger" v-if="manageMode" @click="onDeleteSelected">删除</el-button>
        <el-button type="primary" v-else @click="onCheckout">去结算</el-button>
      </div>
    </div>
  </div>
  <!-- 支付二维码弹窗（放在主 template 内） -->
  <div v-if="showPayModal" class="pay-modal-overlay" @click.self="closePayModal">
    <div class="pay-modal">
      <h3>请使用微信/支付宝扫码付款</h3>
      <div class="qr-grid" style="display:flex;flex-wrap:wrap;gap:12px;justify-content:center;margin-top:12px;">
        <div style="text-align:center;">
          <img src="/src/assets/qrcode.png" alt="pay-qr" style="width:200px;height:200px;border:1px solid #eee;border-radius:6px;" />
          <div style="margin-top:8px;font-size:14px;color:#333;font-weight:600">应付金额：¥{{ payAmount.toFixed(2) }}</div>
        </div>
      </div>
          <div style="margin-top:12px; display:flex; gap:12px; justify-content:center; flex-wrap:wrap;">
            <div v-for="(o, idx) in payOrders" :key="idx" style="min-width:140px;text-align:center;font-size:12px;">
              <div>商家 {{ o.merchantId }}</div>
              <div v-if="o.paid" style="color:green;font-weight:600;margin-top:4px">已支付</div>
              <div v-else style="color:#999;margin-top:4px">未支付</div>
            </div>
          </div>
      <div style="margin-top:12px;display:flex;gap:8px;justify-content:center;">
        <el-button type="primary" @click="closePayModal">关闭</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { CATEGORIES } from '@/constants/categories'
import { useRouter } from 'vue-router'
import * as cartApi from '@/api/user/cart'
import { getDishesByStore } from '@/api/user/store'
import { ElMessageBox, ElMessage } from 'element-plus'

const router = useRouter()

const cartData = ref<any>({ shops: [] })
const manageMode = ref(false)
const activeCategory = ref('全部')

function buildCategoriesFromCart(data: any) {
  // Build categories from canonical list CATEGORIES (exclude id 0 which is "全部")
  console.log('buildCategoriesFromCart', data.shops)
  console.log('CATEGORIES', CATEGORIES)
  const present: string[] = []
  const seen = new Set<string>()
  let unmatched = false
  ;(data.shops || []).forEach((s: any) => {
    (s.items || []).forEach((it: any) => {
      const catName = it.category || it.categoryName || it.cat || it.name||''
      const catId = it.id||it.categoryId || it.category_id || it.catId ||null
      let matched = false
      for (const c of CATEGORIES) {
        if (c.id === 0) continue
        // match by id if available
        if (catId != null && String(c.id) === String(catId)) {
          if (!seen.has(c.label)) { present.push(c.label); seen.add(c.label) }
          matched = true
          break
        }
        // match by textual label/key/filter
        if (catName && (String(c.label) === String(catName) || String(c.key) === String(catName) || String(c.filter) === String(catName))) {
          if (!seen.has(c.label)) { present.push(c.label); seen.add(c.label) }
          matched = true
          break
        }
      }
      if (!matched) unmatched = true
    })
  })
  if (unmatched && !seen.has('其它')) present.push('其它')
  return present
}

const categories = ref<string[]>([])

async function load() {
  const d = await cartApi.getCart()
  // 支持后端返回两种形式：直接 { shops: [...] }（demo）或统一包裹形式 { code, msg, data: { shops: [...] } }
  let shops: any[] = []
  if (d) {
    if (Array.isArray(d.shops)) shops = d.shops
    else if (d.data && Array.isArray(d.data.shops)) shops = d.data.shops
    else if (Array.isArray(d)) shops = d
  }

  // Normalize shape and coerce numeric fields to numbers to avoid template errors
  cartData.value = { shops: (shops || []).map((s: any) => {
    const items = (s.items || []).map((it: any) => {
      const qty = Number(it.qty || it.qty === 0 ? it.qty : (it.qty === undefined ? 0 : it.qty)) || 0
      const price = Number(it.price) || 0
      return {
        // preserve backend keys (dish_id, dishId, id...), but ensure numeric types and selection flag
        ...it,
        qty,
        price,
        selected: !!it.selected,
        originalQty: it.originalQty != null ? Number(it.originalQty) : qty
      }
    })
    // determine shop selected based on selectable items
    const selectable = items.filter(isSelectableItem)
    // normalize shop fields to canonical keys used across the frontend
    const storeId = s.storeId || s.merchant_id || s.id || s.merchantId || s.base_id || s.baseId
    const name = s.name || s.merchant_name || s.storeName || s.store_name || s.shop_name || s.store || ''
    const logo = s.logo || s.storeLogo || s.store_logo || s.logo_url || s.img || ''
    return {
      ...s,
      storeId,
      id: s.id || storeId,
      base_id: s.base_id || s.baseId || storeId,
      name,
      logo,
      items,
      selected: selectable.length > 0 ? selectable.every((it: any) => !!it.selected) : false
    }
  }) }

  // Try to enrich each shop's items with dish metadata (category/categoryId/img/name)
  try {
    // For each shop, fetch its dishes and merge by id
    await Promise.all((cartData.value.shops || []).map(async (s: any) => {
      const sid = s.storeId || s.id || s.base_id || s.baseId || s.merchant_id || s.merchantId
      if (!sid) return
      try {
        const res: any = await getDishesByStore(sid)
        let dishes: any[] = []
        if (!res) return
        if (Array.isArray(res)) dishes = res
        else if (res.data && Array.isArray(res.data)) dishes = res.data
        else if (res.data && res.data.data && Array.isArray(res.data.data)) dishes = res.data.data

        const map = new Map()
        dishes.forEach((d: any) => {
          const id = d.id || d.dishId || d.DishId
          if (id != null) map.set(String(id), d)
        });

        // merge
        (s.items || []).forEach((it: any) => {
          const key = String(it.dishId || it.dish_id || it.id || '')
          const d = map.get(key)
          if (d) {
            // keep existing fields but add category info and image if missing
            if (d.categoryId != null) it.categoryId = d.categoryId
            else if (d.category != null) it.categoryId = d.category
            if (!it.name && (d.name || d.dishName)) it.name = d.name || d.dishName
            if (!it.img && (d.image || d.imageUrl || d.img)) it.img = d.image || d.imageUrl || d.img
            if (!it.categoryName && (d.categoryName || d.categoryLabel || d.category)) it.categoryName = d.categoryName || d.categoryLabel || d.category
            // if dish has an explicit category id but shop item lacks it, set it
            if (!it.categoryId && (d.categoryId || d.category)) it.categoryId = d.categoryId || d.category
          }
        })
      } catch (e) {
        // ignore per-shop failure
        console.warn('getDishesByStore failed for', sid, e)
      }
    }))
  } catch (e) {
    console.warn('enrich cart items failed', e)
  }

  categories.value = buildCategoriesFromCart(cartData.value)
}

onMounted(() => { load() })

const visibleShops = computed(() => {
  if (activeCategory.value === '全部') return cartData.value.shops || []
  return (cartData.value.shops || []).map((s: any) => ({ ...s, items: s.items.filter((it: any) => showItemByCategory(it)) })).filter((s: any) => s.items.length > 0)
})

function isSelectableItem(it: any) {
  return (it.qty || 0) > 0
}

function setCategory(c: string) {
  activeCategory.value = c
}

function showItemByCategory(item: any) {
  if (activeCategory.value === '全部') return true
  // If item has textual category, compare directly
  if (item.category && String(item.category) === String(activeCategory.value)) return true
  if (item.categoryName && String(item.categoryName) === String(activeCategory.value)) return true
  // Otherwise try numeric category id match against CATEGORIES
  const catObj = CATEGORIES.find((c: any) => String(c.label) === String(activeCategory.value))
  if (catObj) {
    const cid = catObj.id
    if (String(item.categoryId) === String(cid) || String(item.category_id) === String(cid)) return true
  }
  return false
}

async function onToggleShop(shop: any) {
  // optimistically update local state
  shop.items.forEach((it: any) => { if (isSelectableItem(it)) it.selected = !!shop.selected })
  // ensure shop.selected is consistent
  shop.selected = shop.items.filter(isSelectableItem).length === 0 ? false : shop.items.filter(isSelectableItem).every((x: any) => !!x.selected)
  // fire API without blocking UI
  try { await cartApi.toggleShopSelection({  storeId: shop.storeId || shop.merchant_id , selected: !!shop.selected }) } catch (e) {}
}

async function onToggleItem(shop: any, it: any) {
  // optimistic update already done by v-model; ensure shop selected sync
  shop.selected = shop.items.filter(isSelectableItem).length === 0 ? false : shop.items.filter(isSelectableItem).every((x: any) => !!x.selected)
  try { await cartApi.toggleItemSelection({ storeId: shop.storeId || shop.merchant_id , dishId: it.dishId||it.dish_id, selected: !!it.selected }) } catch (e) {}
}

async function incQty(shop: any, it: any) {
  // 不能超过 originalQty（原始已加入购物车数量）
  const orig = it.originalQty || it.qty || 0
  if ((it.qty || 0) >= orig) {
    ElMessage({ type: 'warning', message: '数量不能超过原始加入购物车的数量' })
    return
  }
  it.qty = (it.qty || 0) + 1
  try { await cartApi.updateQty({ storeId: shop.storeId, dishId: it.dishId, qty: it.qty }) } catch (e) {}
}

async function decQty(shop: any, it: any) {
  if ((it.qty || 0) <= 0) return
  it.qty = it.qty - 1
  // if qty becomes 0, ensure it's deselected to avoid select-all ambiguity
  if ((it.qty || 0) === 0) it.selected = false
  try { await cartApi.updateQty({ storeId: shop.storeId, dishId: it.dishId, qty: it.qty }) } catch (e) {}
}

const selectAllChecked = computed({
  get() {
    const shops = cartData.value.shops || []
    if (!shops.length) return false
    // Only consider shops that have at least one selectable item
    const relevant = shops.filter((s: any) => s.items && s.items.some(isSelectableItem))
    if (!relevant.length) return false
    return relevant.every((s: any) => s.selected)
  },
  set(v: boolean) {
    // set all selectable items to v
    cartData.value.shops.forEach((s: any) => {
      const hasSelectable = s.items && s.items.some(isSelectableItem)
      s.selected = hasSelectable ? v : false
      s.items.forEach((it: any) => { if (isSelectableItem(it)) it.selected = v })
    })
  }
})

async function onSelectAllChange(v: any) {
  // optimistic local update already applied by computed setter; call API to persist
  try { await cartApi.selectAll(!!selectAllChecked.value) } catch (e) {}
}

function onImgError(e: any) { try { e.target && (e.target.src = '/src/assets/noImg.png') } catch (err) {} }

const totalPrice = computed(() => {
  let total = 0
  ;(cartData.value.shops || []).forEach((s: any) => s.items.forEach((it: any) => { if (it.selected) total += (it.price || 0) * (it.qty || 0) }))
  return total
})

async function onCheckout() {
  const anySelected = (cartData.value.shops || []).some((s: any) => s.items.some((it: any) => it.selected))
  if (!anySelected) { ElMessage({ type: 'warning', message: '请选择要结算的商品' }); return }
  // 构建被选中的商家与菜品信息并保存在 sessionStorage，支付页优先使用该数据
  const selectedShops = (cartData.value.shops || []).filter((s: any) => s.items.some((it: any) => it.selected))
  const shopsPayload = selectedShops.map((s: any) => ({
    storeId: s.storeId || s.merchant_id || s.id,
    name: s.name || s.merchant_name || s.storeName || '',
    items: (s.items || []).filter((it: any) => it.selected).map((it: any) => ({
      dishId: it.dishId || it.dish_id || it.id,
      name: it.name || it.dishName || '',
      price: Number(it.price || 0),
      qty: it.qty
    }))
  }))
  sessionStorage.setItem('checkout_payload', JSON.stringify({ shops: shopsPayload }))
  router.push('/user/payment/confirm')
}

// 支付 modal 管理
const showPayModal = ref(false)
const payOrders = ref<any[]>([])
const payUrl = ref<string>('')
const payAmount = ref<number>(0)
let payPollTimer: any = null

function openPayModal(orders: any) {
  // orders 可以是单个对象或数组；payUrl 已由 onCheckout 设置
  payOrders.value = Array.isArray(orders) ? orders : [orders]
  showPayModal.value = true
  // 每 2 秒轮询每个订单状态，直到全部支付完成
  payPollTimer = setInterval(async () => {
    try {
      for (const o of payOrders.value) {
        const id = o.orderId || o.id
        if (!id) continue
        const res = await fetch('/api/order/status?orderId=' + encodeURIComponent(id), { credentials: 'include' })
        if (!res.ok) continue
        const body = await res.json()
        const status = body?.data?.status || null
        const payStatus = body?.data?.pay_status || null
        if (status === 2 || payStatus === 'paid') {
          o.paid = true
        }
      }
      // 如果全部 paid，则关闭并跳转
      if (payOrders.value.every((x: any) => x.paid)) {
        clearInterval(payPollTimer)
        showPayModal.value = false
        ElMessage({ type: 'success', message: '支付成功' })
        window.location.href = '/#/user/payment/success'
      }
    } catch (e) {
      // ignore
    }
  }, 2000)
}

function closePayModal() {
  showPayModal.value = false
  payOrders.value = []
  if (payPollTimer) {
    clearInterval(payPollTimer)
    payPollTimer = null
  }
}

async function onDeleteSelected() {
  const anySelected = (cartData.value.shops || []).some((s: any) => s.items.some((it: any) => it.selected))
  if (!anySelected) { ElMessage({ type: 'warning', message: '请选择要删除的商品' }); return }
  try {
    await ElMessageBox.confirm('确认删除选中商品吗？', '删除', { type: 'warning' })
    await cartApi.deleteSelected()
    ElMessage({ type: 'success', message: '删除成功' })
    await load()
  } catch (e) {
    // ignore
  }
}

function toggleManage() { manageMode.value = !manageMode.value }

function goStore(shop: any) {
  console.log('goStore', shop)
  const id = shop.storeId || shop.id || shop.base_id || shop.baseId || shop.merchant_id || shop.merchantId
  if (id) {
    router.push('/user/store/' + encodeURIComponent(String(id)))
  } else {
    router.push('/user/store/' + encodeURIComponent(shop.name || ''))
  }
}

function canceledItems(shop: any) {
  return (shop.items || []).map((it: any) => ({ dishId: it.dishId, name: it.name, canceledQty: Math.max(0, (it.originalQty || 0) - (it.qty || 0)) })).filter((x: any) => x.canceledQty > 0)
}

function showCanceled(shop: any) {
  return !manageMode.value && canceledItems(shop).length > 0
}

</script>

<style scoped>
/* === 页面整体背景 === */
/* 使用左右对称的校园背景图填充两边 */
body {
  background:
    url('@/assets/login/img_denglu_bj.jpg') left top no-repeat,
    url('@/assets/login/img_denglu_bj.jpg') right top no-repeat,
    linear-gradient(180deg, #fffef5, #fff9cc);
  background-size: 28%, 28%, cover;
  background-attachment: fixed;
  background-repeat: no-repeat;
  background-position: left top, right top, center;
}

/* === 页面主内容（居中 60%） === */
.cart-page {
  position: relative;
  padding: 20px 24px 100px; /* 底部留出结算栏空间 */
  width: 60%;
  margin: 0 auto;
  min-height: calc(100vh - 80px);
  background: rgba(255, 255, 255, 0.9);
  border-radius: 14px;
  box-shadow: 0 6px 26px rgba(255, 204, 0, 0.15);
  backdrop-filter: blur(10px);
  z-index: 1;
}

/* === 顶部分类栏 === */
.cart-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  background: linear-gradient(90deg, #fffbe6, #fff9d6);
  border-radius: 12px;
  padding: 10px 16px;
  box-shadow: 0 2px 8px rgba(255, 204, 0, 0.2);
}

.categories-bar {
  display: flex;
  gap: 8px;
  align-items: center;
}

.cat-btn {
  background: #fff9e6;
  border: 1px solid #ffe58f;
  padding: 6px 12px;
  border-radius: 18px;
  cursor: pointer;
  font-weight: 500;
  color: #5b3b00;
  transition: all 0.25s ease;
}
.cat-btn:hover {
  background: #ffe58f;
}
.cat-btn.active {
  background: linear-gradient(90deg, #ffd666, #ffcc00);
  color: #5b3b00;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(255, 193, 7, 0.3);
}

/* === 店铺卡片 === */
.shop-card {
  background: #fffef9;
  padding: 14px;
  border-radius: 10px;
  margin-bottom: 14px;
  box-shadow: 0 3px 10px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}
.shop-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 14px rgba(255, 204, 0, 0.25);
}

.shop-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 6px;
  border-bottom: 1px dashed rgba(0, 0, 0, 0.06);
}
.shop-name {
  font-weight: 700;
  cursor: pointer;
  color: #b8860b;
  transition: color 0.2s ease;
}
.shop-name:hover {
  color: #ff9900;
}

.shop-logo {
  width: 36px;
  height: 36px;
  object-fit: cover;
  border-radius: 8px;
  margin-right: 8px;
  vertical-align: middle;
}

/* === 商品行 === */
.shop-items {
  margin-top: 8px;
}
.item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px dashed rgba(0, 0, 0, 0.05);
}
.item-left {
  width: 40px;
}
.item-thumb { width: 72px }
.item-thumb img { width: 72px; height: 72px; object-fit: cover; border-radius: 6px }
.item-mid {
  flex: 1;
  cursor: pointer;
}
.item-name {
  font-weight: 500;
  color: #333;
}
.item-price {
  color: #e53935;
  font-weight: 600;
}
.item-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.qty {
  min-width: 24px;
  text-align: center;
  color: #333;
  font-weight: 500;
}

/* === 已取消商品块 === */
.canceled-card {
  margin-top: 8px;
  background: #fff7f0;
  border-radius: 8px;
  padding: 8px;
  border: 1px dashed rgba(0, 0, 0, 0.04);
}
.canceled-header {
  font-weight: 600;
  color: #c23516;
  margin-bottom: 6px;
}
.canceled-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
}
.canceled-name {
  color: #666;
}
.canceled-count {
  color: #999;
}

/* === 底部固定结算栏 === */
/* 🚀 脱离 .cart-page，始终固定在屏幕底部 🚀 */
.cart-bg {
  width: 100%;
  min-height: 100vh;
  background: url('/src/assets/login/img_denglu_bj.jpg') center/cover no-repeat;
  background-attachment: fixed;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 60px 0;
}

.cart-page {
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

.cart-page:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 28px rgba(255, 193, 7, 0.45);
}

.cart-bottom {
  position: fixed;
  bottom: 55px;
  left: 50%;
  transform: translateX(-50%);
  width: 60%;
  background: linear-gradient(90deg, #fffbe6, #fff8c8);
  padding: 12px 30px;
  border-top: 1px solid #ffe58f;
  box-shadow: 0 -4px 16px rgba(255, 193, 7, 0.2);
  align-items: center;
  justify-content: center;
  z-index: 99;
}

.cart-bottom-inner {
  width: 100%;
  max-width: 900px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.cart-bottom .left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.cart-bottom .center {
  color: #333;
  font-weight: 600;
  font-size: 15px;
}
.cart-bottom .center .total {
  color: #e53935;
  font-weight: 700;
}
.cart-bottom .right {
  display: flex;
  align-items: center;
}

/* === 响应式 === */
@media(max-width:900px){
  .cart-page { width: 92%; padding: 12px; }
  .cart-bottom-inner { width: 92%; }
}

/* 支付弹窗样式 */
.pay-modal-overlay{
  position:fixed;left:0;top:0;right:0;bottom:0;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.45);z-index:1200;
}
.pay-modal{background:#fff;padding:18px;border-radius:8px;box-shadow:0 10px 30px rgba(0,0,0,0.2);text-align:center}
</style>
