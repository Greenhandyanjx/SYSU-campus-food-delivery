<template>
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
      <div v-for="(shop, sIdx) in visibleShops" :key="shop.storeId" class="shop-card">
        <div class="shop-header">
          <el-checkbox v-model="shop.selected" @change="onToggleShop(shop)" />
          <div class="shop-name" @click="goStore(shop)">{{ shop.name }}</div>
        </div>

        <div class="shop-items">
          <div v-for="(it, iIdx) in shop.items.filter(it => showItemByCategory(it))" :key="it.dishId" class="item-row">
            <div class="item-left">
              <el-checkbox v-model="it.selected" @change="onToggleItem(shop, it)" />
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
            <div v-for="ci in canceledItems(shop)" :key="ci.dishId" class="canceled-row">
              <div class="canceled-name">{{ ci.name }}</div>
              <div class="canceled-count">已取消: {{ ci.canceledQty }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部固定栏 -->
    <div class="cart-bottom">
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
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import * as cartApi from '@/api/user/cart'
import { ElMessageBox, ElMessage } from 'element-plus'

const router = useRouter()

const cartData = ref<any>({ shops: [] })
const manageMode = ref(false)
const activeCategory = ref('全部')

function buildCategoriesFromCart(data: any) {
  const set = new Set<string>()
  data.shops.forEach((s: any) => s.items.forEach((it: any) => set.add(it.category || '其它')))
  return Array.from(set)
}

const categories = ref<string[]>([])

async function load() {
  const d = await cartApi.getCart()
  // Normalize shape
  cartData.value = { shops: (d.shops || d.shops || []) }
  // ensure selected/qty fields
  cartData.value.shops.forEach((s: any) => {
    s.selected = !!s.selected
    s.items = (s.items || []).map((it: any) => ({ selected: !!it.selected, qty: it.qty || 1, originalQty: it.originalQty || it.qty || 1, ...it }))
  })
  categories.value = buildCategoriesFromCart(cartData.value)
}

onMounted(() => { load() })

const visibleShops = computed(() => {
  if (activeCategory.value === '全部') return cartData.value.shops || []
  return (cartData.value.shops || []).map((s: any) => ({ ...s, items: s.items.filter((it: any) => it.category === activeCategory.value) })).filter((s: any) => s.items.length > 0)
})

function setCategory(c: string) {
  activeCategory.value = c
}

function showItemByCategory(item: any) {
  return activeCategory.value === '全部' || item.category === activeCategory.value
}

async function onToggleShop(shop: any) {
  await cartApi.toggleShopSelection({ storeId: shop.storeId, selected: !!shop.selected })
  // sync local
  shop.items.forEach((it: any) => (it.selected = !!shop.selected))
}

async function onToggleItem(shop: any, it: any) {
  await cartApi.toggleItemSelection({ storeId: shop.storeId, dishId: it.dishId, selected: !!it.selected })
  // sync shop selected
  shop.selected = shop.items.every((x: any) => !!x.selected)
}

async function incQty(shop: any, it: any) {
  // 不能超过 originalQty（原始已加入购物车数量）
  const orig = it.originalQty || it.qty || 0
  if ((it.qty || 0) >= orig) {
    ElMessage({ type: 'warning', message: '数量不能超过原始加入购物车的数量' })
    return
  }
  it.qty = (it.qty || 0) + 1
  await cartApi.updateQty({ storeId: shop.storeId, dishId: it.dishId, qty: it.qty })
}

async function decQty(shop: any, it: any) {
  if ((it.qty || 0) <= 0) return
  it.qty = it.qty - 1
  await cartApi.updateQty({ storeId: shop.storeId, dishId: it.dishId, qty: it.qty })
}

const selectAllChecked = computed({
  get() {
    const shops = cartData.value.shops || []
    if (!shops.length) return false
    return shops.every((s: any) => s.selected)
  },
  set(v: boolean) {
    cartData.value.shops.forEach((s: any) => {
      s.selected = v
      s.items.forEach((it: any) => (it.selected = v))
    })
  }
})

async function onSelectAllChange(v: any) {
  await cartApi.selectAll(!!selectAllChecked.value)
}

const totalPrice = computed(() => {
  let total = 0
  ;(cartData.value.shops || []).forEach((s: any) => s.items.forEach((it: any) => { if (it.selected) total += (it.price || 0) * (it.qty || 0) }))
  return total
})

async function onCheckout() {
  const anySelected = (cartData.value.shops || []).some((s: any) => s.items.some((it: any) => it.selected))
  if (!anySelected) { ElMessage({ type: 'warning', message: '请选择要结算的商品' }); return }
  try {
    const res = await ElMessageBox.confirm('确认要结算已选商品吗？', '结算', { type: 'warning' })
    // proceed
    const r = await cartApi.checkout()
    ElMessage({ type: 'success', message: '结算成功：' + (r.orderId || 'demo') })
    // reload
    await load()
  } catch (e) {
    // cancel or fail
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

function goStore(shop: any) { router.push('/user/store/' + encodeURIComponent(shop.name)) }

function canceledItems(shop: any) {
  return (shop.items || []).map((it: any) => ({ dishId: it.dishId, name: it.name, canceledQty: Math.max(0, (it.originalQty || 0) - (it.qty || 0)) })).filter((x: any) => x.canceledQty > 0)
}

function showCanceled(shop: any) {
  return !manageMode.value && canceledItems(shop).length > 0
}

</script>

<style scoped>
.cart-page { padding: 12px; width: 60%; margin: 0 auto; }
.cart-top { display:flex; justify-content:space-between; align-items:center; margin-bottom:12px }
.categories-bar { display:flex; gap:8px; align-items:center }
.cat-btn { background:#fff9e6; border:1px solid #ffe58f; padding:6px 10px; border-radius:16px; cursor:pointer }
.cat-btn.active { background:#ffd666; color:#5b3b00; font-weight:600 }
.manage-area { display:flex; align-items:center }
.shop-card { background:#fffef7; padding:12px; border-radius:8px; margin-bottom:10px; box-shadow:0 1px 4px rgba(0,0,0,0.04) }
.shop-header { display:flex; align-items:center; gap:12px }
.shop-name { font-weight:700; cursor:pointer; color:#b8860b }
.shop-items { margin-top:8px }
.item-row { display:flex; align-items:center; justify-content:space-between; padding:8px 0; border-bottom:1px dashed rgba(0,0,0,0.04) }
.item-left { width:40px }
.item-mid { flex:1; cursor:pointer }
.item-name { font-weight:500 }
.item-price { color:#e53935 }
.item-right { display:flex; align-items:center; gap:8px }
.qty { min-width:24px; text-align:center }
.cart-bottom { position:fixed; left:50%; transform:translateX(-50%); bottom:60px; background:#fff; width:60%; max-width:900px; padding:10px 16px; border-radius:8px; display:flex; align-items:center; justify-content:space-between; box-shadow:0 8px 24px rgba(0,0,0,0.08) }
.cart-bottom .left { display:flex; align-items:center; gap:8px }
.cart-bottom .center { color:#333 }
.cart-bottom .right { display:flex; align-items:center }

.canceled-card { margin-top:8px; background:#fff7f0; border-radius:8px; padding:8px; border:1px dashed rgba(0,0,0,0.04) }
.canceled-header { font-weight:600; color:#c23516; margin-bottom:6px }
.canceled-row { display:flex; justify-content:space-between; padding:6px 0 }
.canceled-name { color:#666 }
.canceled-count { color:#999 }

/* small screens */
@media(max-width:900px){ .cart-page { width:92% } .cart-bottom { width:92% } }
</style>
