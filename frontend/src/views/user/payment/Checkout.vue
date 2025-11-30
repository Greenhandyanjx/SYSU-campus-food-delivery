<template>
  <div class="checkout-wrapper">
    <div class="checkout-card">
    <!-- 地址区域 -->
    <div class="address-section">
      <div class="address-main">
        <div class="address-info">
          <div class="address-text">{{ selectedAddress ? formatFullAddress(selectedAddress) : '请选择收货地址' }}</div>
          <div class="address-sub" v-if="selectedAddress">
            {{ selectedAddress.name }} {{ selectedAddress.phone }}
          </div>
        </div>
          <div style="display:flex;gap:8px;align-items:center">
            <el-button type="text" @click="openAddressManager">选择/新增地址</el-button>
            <van-icon name="arrow" class="arrow" />
          </div>
      </div>
      <div class="delivery-time" v-if="selectedAddress">
        预计 {{ deliveryTime }} 送达
      </div>
    </div>

    <!-- 订单详情 -->
    <div class="order-section">
      <div class="shop-list">
        <div v-for="shop in shopList" :key="shop.storeId" class="shop-block">
          <div class="shop-name">
            <van-icon name="shop-o" />
            {{ shop.name }}
          </div>

          <!-- 菜品列表 -->
          <div class="dish-list">
            <div v-for="item in shop.items" :key="item.dishId" class="dish-item">
              <div class="dish-name">{{ item.name }}</div>
              <div class="dish-spec" v-if="item.spec">{{ item.spec }}</div>
              <div class="dish-quantity">x{{ item.qty }}</div>
              <div class="dish-price">¥{{ (item.price * item.qty).toFixed(2) }}</div>
            </div>
          </div>

          <!-- 打包费 & 配送费 -->
          <div class="fee-row" v-if="shop.packingFee > 0">
            <span>打包费</span>
            <span>¥{{ shop.packingFee.toFixed(2) }}</span>
          </div>
          <div class="fee-row">
            <span>配送费</span>
            <span>¥{{ shop.deliveryFee?.toFixed(2) || '0.00' }}</span>
          </div>

          <!-- 店铺小计 -->
          <div class="shop-total">
            <span>店铺合计</span>
            <span class="price">¥{{ shop.shopTotal.toFixed(2) }}</span>
          </div>
        </div>
      </div>

      <!-- 备注与餐具 -->
      <div class="extra-section">
        <van-field
          v-model="form.remark"
          rows="2"
          autosize
          type="textarea"
          placeholder="给商家留言（口味、偏好等）"
          class="remark-input"
        />
        <div class="tableware-row">
          <span>餐具份数</span>
          <van-stepper v-model="form.tableware" integer :min="0" />
        </div>
      </div>
    </div>

    <!-- 底部固定支付栏 -->
    <div class="bottom-bar">
      <div class="total-info">
        <div class="total-label">实付款</div>
        <div class="total-price">¥{{ totalAmount.toFixed(2) }}</div>
      </div>
      <div class="pay-btn" @click="onPay">去支付</div>
    </div>

    <!-- 支付二维码弹窗（保持不变） -->
    <div v-if="showPayModal" class="pay-modal-overlay" @click.self="closePayModal">
      <div class="pay-modal">
        <div class="pay-title">请使用微信/支付宝扫码支付</div>
        <img :src="payQrImg" alt="pay" style="width:200px;height:200px;margin:12px 0;" />
        <div class="pay-amount">¥{{ payAmount.toFixed(2) }}</div>
      </div>
    </div>
    </div>
  </div>

  <!-- 地址选择/新增模态 -->
  <el-dialog title="选择收货地址" v-model="showAddressModal" width="60%">
    <div class="addr-modal">
      <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
        <div style="font-weight:600">我的收货地址</div>
        <div>
          <el-button type="primary" size="small" @click="openAddInline">新增地址</el-button>
        </div>
      </div>

      <!-- 地址列表：使用与地址管理页相同的样式类 -->
      <div class="addr-list">
        <el-card v-for="a in addresses" :key="a.id" :class="['addr-card', {selected: selectedAddress && selectedAddress.id===a.id}]" @click="pickAddress(a)">
          <div class="addr-top"><strong>{{ a.name }}</strong> <span style="margin-left:8px">{{ a.phone }}</span> <span v-if="a.isDefault" class="default-tag">默认</span></div>
          <div class="addr-detail">{{ formatFullAddress(a) }}</div>
        </el-card>
      </div>

      <!-- 内联新增地址界面：与地址管理的新增表单字段保持一致（简洁版） -->
      <div v-if="showAddAddress" style="margin-top:12px">
        <el-card>
          <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px">
            <div style="font-weight:600">新增地址</div>
            <div>
              <el-button type="text" size="small" @click="closeAddInline">取消</el-button>
            </div>
          </div>
          <el-form :model="newAddress" label-position="top">
            <el-form-item label="收货人">
              <el-input v-model="newAddress.name" />
            </el-form-item>
            <el-form-item label="手机号">
              <el-input v-model="newAddress.phone" />
            </el-form-item>
            <el-form-item label="详细地址（街道/门牌）">
              <el-input v-model="newAddress.detail" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="addNewAddress">保存地址</el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </div>
  </el-dialog>
  <!-- 完整的新增地址对话（内置于结算页） -->
  <el-dialog v-model="showAddDialog" width="700px" class="dialog-box" @opened="initMapInline">
    <template #title>
      <div class="dialog-title">
        <span class="dialog-title-icon">📍</span>
        <div>
          <div class="dialog-title-main">新增地址</div>
          <div class="dialog-title-sub">选择位置或输入详细地址以便骑手准确配送</div>
        </div>
      </div>
    </template>

    <el-form-item class="map-form-item">
      <div class="map-panel">
        <div class="map-container-wrap">
          <div id="mapContainer" class="map-container"></div>
          <div class="map-controls">
            <el-button class="locate-btn-fixed" type="primary" size="small" circle title="定位到当前位置" @click="locateCurrent">
              <img src="@/assets/icons/icon_locate.svg" class="locate-icon" alt="定位" />
            </el-button>
          </div>
        </div>

        <div class="search-panel">
          <el-input v-model="searchKeyword" placeholder="搜索地点、小区或街道名称" clearable class="map-search-input" @input="onKeywordInput">
          </el-input>

          <div v-if="suggestions.length" class="suggestion-box">
            <div v-for="(item, idx) in suggestions" :key="idx" class="suggestion-item" @click="selectSuggestion(item)">
              <div class="suggestion-content">
                <div class="suggestion-name">{{ item.name }}</div>
                <div class="suggestion-address">{{ formatTipAddress(item) }}</div>
              </div>
            </div>
          </div>

          <div class="final-address" v-if="addressForm.detail">
            <div class="label">已选择地址：</div>
            <div class="address-text">{{ addressForm.detail }}</div>
          </div>
        </div>
      </div>
    </el-form-item>

    <div class="floating-form">
      <div class="form-item" :class="{ 'has-value': addressForm.name }">
        <input type="text" v-model="addressForm.name" id="name" required>
        <label for="name" :class="{ 'active': addressForm.name }">收货人</label>
      </div>

      <div class="form-item" :class="{ 'has-value': addressForm.phone }">
        <input type="tel" v-model="addressForm.phone" id="phone" required>
        <label for="phone" :class="{ 'active': addressForm.phone }">手机号码</label>
      </div>

      <div class="form-item" :class="{ 'has-value': addressForm.detail }">
        <input type="text" v-model="addressForm.detail" id="address" required readonly>
        <label for="address" :class="{ 'active': addressForm.detail }">收货地址</label>
      </div>

      <div class="form-item" :class="{ 'has-value': addressForm.tag }">
        <select v-model="addressForm.tag" id="tag" required>
          <option value="" disabled selected></option>
          <option value="家">家</option>
          <option value="公司">公司</option>
          <option value="学校">学校</option>
        </select>
        <label for="tag" :class="{ 'active': addressForm.tag }">地址标签</label>
      </div>
    </div>

    <template #footer>
      <el-button @click="closeAddDialog">取消</el-button>
      <el-button type="primary" @click="saveAddress">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import qrImg from '@/assets/qrcode.png'
import { useRouter, useRoute } from 'vue-router'
import orderApi from '@/api/user/order'
import * as addressApi from '@/api/common/address'
import * as cartApi from '@/api/user/cart'
import { getDeliveryConfig } from '@/api/user/store'
import { showToast } from 'vant'

const router = useRouter()

const addresses = ref<any[]>([])
const selectedAddress = ref<any>(null)
const showAddressModal = ref(false)
const showAddAddress = ref(false)
const showAddDialog = ref(false)
const newAddress = ref({ name: '', phone: '', detail: '', province: '', city: '', district: '', street: '', isDefault: false })
// 完整对话使用的表单与搜索建议
const addressForm = ref({ name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 })
const searchKeyword = ref('')
const suggestions = ref<any[]>([])

// 地图/定位/搜索相关（在 initMapInline/setupMap/initAutoComplete 中初始化）
let map: any = null, marker: any = null, geocoder: any = null, autoComplete: any = null, placeSearch: any = null

const shopList = ref<any[]>([]) // 按店铺分组后的数据
const form = ref({ remark: '', tableware: 0 })

const showPayModal = ref(false)
const payQrImg = ref(qrImg)
const payAmount = ref(0)
const pendingOrders = ref<string[]>([])

// 格式化地址
function formatFullAddress(a: any) {
  return `${a.province || ''}${a.city || ''}${a.district || ''}${a.street || ''} ${a.detail || ''}`.trim()
}

// 预计送达时间（当前时间 + 30 分钟）
const deliveryTime = computed(() => {
  const d = new Date()
  d.setMinutes(d.getMinutes() + 30)
  return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
})

// 加载地址
async function loadAddresses() {
  try {
    const res: any = await addressApi.listAddresses()
    const list = Array.isArray(res) ? res : (res.data?.addresses || res.data || [])
    addresses.value = list
    const def = list.find((x: any) => x.isDefault) || list[0]
    if (def) selectedAddress.value = def
  } catch (e) {
    console.error(e)
  }
}

// 加载购物车并按店铺分组
async function loadCart() {
  try {
    // 优先从 sessionStorage 获取来自购物车/店铺页传递的选中信息
    const raw = sessionStorage.getItem('checkout_payload')
    const cartData: any = await cartApi.getCart()

    // 规范化后端返回的购物车结构，尽可能找到一个数组形态的 shops
    let shops: any[] = []
    if (Array.isArray(cartData)) {
      shops = cartData
    } else if (Array.isArray(cartData.shops)) {
      shops = cartData.shops
    } else if (cartData.data && Array.isArray(cartData.data.shops)) {
      shops = cartData.data.shops
    } else if (Array.isArray(cartData.data)) {
      shops = cartData.data
    }

    if (raw) {
      try {
        const payload = JSON.parse(raw)
        const payloadShops = payload.shops || []
        // 基于后端返回的购物车数据，匹配 payload 中的 storeId/dishId 以获得完整的项
        const list: any[] = []
        for (const ps of payloadShops) {
          const storeId = ps.storeId || ps.merchantId || ps.store_id
          const origin = (shops || []).find((s: any) => String(s.storeId || s.merchant_id || s.id) === String(storeId))

          // 如果找不到 origin，也要基于 payload 构建条目（保证上一页选择的内容能呈现）
          const items = (ps.items || []).map((it: any) => {
            const found = origin ? (origin.items || []).find((oi: any) => String(oi.dishId || oi.dish_id || oi.id) === String(it.dishId || it.dish_id || it.id)) : null
            return {
              dishId: it.dishId || it.dish_id,
              name: found?.name || it.name || it.dishName || '商品',
              spec: found?.spec || found?.skuName || it.spec || '',
              qty: (typeof it.qty === 'number') ? it.qty : (found?.qty || 1),
              price: Number(found?.price ?? it.price ?? 0)
            }
          }).filter((x: any) => x.qty > 0)

          if (items.length === 0) continue

          const dishTotal = items.reduce((s: number, it: any) => s + Number(it.price || 0) * Number(it.qty || 0), 0)
          const packingFee = Number((origin && (origin.packingFee || origin.packing_fee)) || 0)
          const deliveryFee = Number((origin && (origin.deliveryFee || origin.delivery_fee)) || 0)
          const name = origin ? (origin.name || origin.storeName || '') : (ps.name || ps.storeName || `店铺 ${storeId}`)

          list.push({ storeId, name, items, packingFee, deliveryFee, shopTotal: dishTotal + packingFee + deliveryFee })
        }

        shopList.value = list
        // 使用完后清理临时数据
        sessionStorage.removeItem('checkout_payload')
        // 为每个店铺获取配送配置（覆盖 deliveryFee / minPrice 等）
        try {
          await Promise.all(shopList.value.map(async (s: any) => {
            const bid = s.storeId || s.merchantId || s.store_id || s.storeId
            if (!bid) return
            const r = await getDeliveryConfig(bid)
            const cfg = r && r.data ? r.data.data || r.data : r
            s.deliveryFee = Number(cfg?.delivery_fee ?? cfg?.deliveryFee ?? s.deliveryFee ?? 2)
            s.minPrice = Number(cfg?.min_price ?? cfg?.minPrice ?? s.minPrice ?? 15)
            s.deliveryRange = Number(cfg?.delivery_range ?? cfg?.deliveryRange ?? s.deliveryRange ?? 2000)
            // recalc shopTotal based on items + packing + delivery
            const itemsTotal = (s.items || []).reduce((sm: number, it: any) => sm + Number(it.price || 0) * Number(it.qty || 0), 0)
            const packing = Number(s.packingFee || s.packing_fee || 0)
            s.shopTotal = itemsTotal + packing + Number(s.deliveryFee || 0)
          }))
        } catch (e) { console.warn('fetch shop delivery configs failed', e) }
        return
      } catch (err) {
        console.warn('解析 checkout_payload 失败，回退到 getCart', err)
      }
    }

    // 默认回退：直接从后端购物车读取所有被选中的项
    shopList.value = (shops || [])
      .filter((s: any) => s.items?.some((it: any) => it.selected))
      .map((s: any) => {
        const items = (s.items || []).filter((it: any) => it.selected)
        const dishTotal = items.reduce((sum: number, it: any) => sum + Number(it.price || 0) * Number(it.qty || 0), 0)
        const packingFee = Number(s.packingFee || s.packing_fee || 0)
        const deliveryFee = Number(s.deliveryFee || s.delivery_fee || 0)
        const shopTotal = dishTotal + packingFee + deliveryFee

        return {
          storeId: s.storeId || s.merchant_id || s.id,
          name: s.name || s.storeName || s.merchant_name,
          packingFee,
          deliveryFee,
          shopTotal,
          items: items.map((it: any) => ({
            dishId: it.dishId || it.dish_id,
            name: it.name || it.dishName,
            spec: it.spec || it.skuName,
            qty: it.qty,
            price: Number(it.price || 0)
          }))
        }
      })
    // 为每个店铺补充配送配置并重新计算 shopTotal
    try {
      await Promise.all(shopList.value.map(async (s: any) => {
        const bid = s.storeId || s.merchant_id || s.id || s.storeId
        if (!bid) return
        const r = await getDeliveryConfig(bid)
        const cfg = r && r.data ? r.data.data || r.data : r
        s.deliveryFee = Number(cfg?.delivery_fee ?? cfg?.deliveryFee ?? s.deliveryFee ?? 2)
        s.minPrice = Number(cfg?.min_price ?? cfg?.minPrice ?? s.minPrice ?? 15)
        s.deliveryRange = Number(cfg?.delivery_range ?? cfg?.deliveryRange ?? s.deliveryRange ?? 2000)
        const itemsTotal = (s.items || []).reduce((sm: number, it: any) => sm + Number(it.price || 0) * Number(it.qty || 0), 0)
        const packing = Number(s.packingFee || s.packing_fee || 0)
        s.shopTotal = itemsTotal + packing + Number(s.deliveryFee || 0)
      }))
    } catch (e) { console.warn('fetch shop delivery configs failed', e) }
  } catch (e) {
    console.error(e)
  }
}

// 总金额
const totalAmount = computed(() => {
  return shopList.value.reduce((sum, shop) => sum + shop.shopTotal, 0)
})

onMounted(async () => {
  await loadAddresses()
  await loadCart()
  // 如果在跳转前已由购物车页创建了 pending orders，则优先使用它们并从 sessionStorage 清除
  try {
    const pendingRaw = sessionStorage.getItem('pending_orders')
    if (pendingRaw) {
      const parsed = JSON.parse(pendingRaw)
      if (Array.isArray(parsed) && parsed.length > 0) {
        pendingOrders.value = parsed.map((x: any) => String(x))
        sessionStorage.removeItem('pending_orders')
      }
    }
  } catch (e) { console.warn('read pending_orders from session failed', e) }
  // 在用户进入结算页时：如果带有 orderId（从订单卡片/详情跳转），不创建新的 pending，而是直接加载该订单用于支付；
  // 否则按照购物车内容创建 pending 以便持久化未完成的结算尝试。
  try {
    const route = useRoute()
    const qid = route.query.orderId
    if (qid) {
      // 如果 URL 带有 orderId，立即将其作为待支付目标，避免后续回退到 createPayOrder
      pendingOrders.value = [String(qid)]
      // 支付已有订单 —— 不创建新的 pending，只将该订单 id 作为待支付目标并尝试加载详情
      try {
        const od: any = await orderApi.getOrderDetail(String(qid))
        const odata = od && od.data && (od.data.data || od.data)
        if (odata) {
          // 构建页面展示数据（兼容旧后端结构）
          shopList.value = [{
            storeId: odata.merchantId || odata.merchantid || odata.merchantID || 0,
            name: odata.storeName || odata.shopName || '',
            items: (odata.items || odata.orderDetailList || []).map((it: any) => ({
              dishId: it.id || it.skuId || null,
              name: it.name,
              spec: it.spec || it.sku || '',
              qty: it.qty || it.count || it.num || 1,
              price: Number(it.price || 0)
            })),
            packingFee: Number(odata.packAmount || odata.pack_amount || 0),
            deliveryFee: Number(odata.deliveryAmount || odata.delivery_amount || 0),
            shopTotal: Number(odata.amount || odata.total || 0)
          }]
          pendingOrders.value = [String(qid)]
          payAmount.value = Number(odata.amount || 0)
        }
      } catch (e) {
        console.warn('failed to fetch order detail for checkout', e)
      }
    } else {
      // 原购物车结算路径：为当前选中项创建 pending（持久化尝试）
      let payloadShops: any[] = []
      // send totalPrice as items total (exclude delivery); deliveryFee sent separately
      payloadShops = shopList.value.map((s: any) => {
        const itemsTotal = (s.items || []).reduce((sm: number, it: any) => sm + Number(it.price || 0) * Number(it.qty || 0), 0)
        return ({ merchantId: s.storeId || s.merchantId || s.id, totalPrice: itemsTotal, deliveryAmount: Number(s.deliveryFee || s.delivery_fee || 0) })
      })

      if (payloadShops && payloadShops.length > 0 && selectedAddress.value) {
        const payload = { shops: payloadShops, consigneeid: selectedAddress.value.id, totalPrice: totalAmount.value, remarks: form.value.remark }
        try {
          const cp: any = await cartApi.createPending(payload)
          if (cp && cp.data && cp.data.orders) {
            pendingOrders.value = (cp.data.orders || []).map((x: any) => String(x.orderId || x.OrderID || x.order_id))
          } else if (cp && cp.orders) {
            pendingOrders.value = (cp.orders || []).map((x: any) => String(x.orderId || x.OrderID || x.order_id))
          }
        } catch (e) {
          console.warn('create pending order failed', e)
        }
      }
    }
  } catch (e) {
    console.warn('checkout onMounted error', e)
  }
})

function openAddressManager() {
  // 在页面内显示地址选择/新增模态，而不是跳转
  showAddressModal.value = true
}

async function pickAddress(a: any) {
  selectedAddress.value = a
  showAddressModal.value = false
  // 如果已有 pending orders，更新它们的 consigneeid
  try {
    if (pendingOrders.value && pendingOrders.value.length > 0) {
      for (const oid of pendingOrders.value) {
        await orderApi.updateOrderAddress(String(oid), { consigneeid: a.id }).catch(() => {})
      }
    }
  } catch (e) { console.warn('update pending order address failed', e) }
}

async function addNewAddress() {
  // 简单校验
  if (!newAddress.value.name || !newAddress.value.phone || !newAddress.value.detail) {
    showToast('请填写姓名、手机号和详细地址')
    return
  }
  try {
    const payload = {
      name: newAddress.value.name,
      phone: newAddress.value.phone,
      detail: newAddress.value.detail,
      province: newAddress.value.province || '',
      city: newAddress.value.city || '',
      district: newAddress.value.district || '',
      street: newAddress.value.street || '',
      isDefault: newAddress.value.isDefault ? 1 : 0
    }
    const res: any = await addressApi.addAddress(payload)
    // 刷新地址列表并选择新地址（后端返回格式兼容性较多）
    await loadAddresses()
    const maybe = addresses.value.find((x: any) => x.phone === newAddress.value.phone && x.name === newAddress.value.name)
    if (maybe) selectedAddress.value = maybe
    // 清空表单
    newAddress.value = { name: '', phone: '', detail: '', province: '', city: '', district: '', street: '', isDefault: false }
    // 如果是内联新增则关闭内联面板
    showAddAddress.value = false
    showToast('地址已添加')
  } catch (e) {
    console.error(e)
    showToast('新增地址失败')
  }
}

function openAddInline() {
  // 在结算页内打开完整的新增地址对话
  showAddDialog.value = true
}

function closeAddInline() {
  showAddAddress.value = false
}

function closeAddDialog() {
  showAddDialog.value = false
}

// 初始化内联完整地图对话（包装 initMap 的显示时机）
function initMapInline() {
  // 等待容器可见后初始化地图
  nextTick(() => setTimeout(initMap, 200))
}

function initMap() {
  const amapKey = (import.meta.env.VITE_AMAP_KEY as string) || ''
  const old = document.getElementById('mapContainer')
  if (!old) return
  old.innerHTML = ''

  const AMap = (window as any).AMap
  if (!AMap) {
    // 动态加载脚本
    const script = document.createElement('script')
    script.id = 'amap-script'
    script.src = `https://webapi.amap.com/maps?v=2.0&key=${amapKey}`
    script.onload = () => setupMap()
    document.head.appendChild(script)
    return
  }
  setupMap()
}

function setupMap() {
  const AMap = (window as any).AMap
  if (!AMap) return
  try {
    AMap.plugin('AMap.Geocoder', () => {
      geocoder = new AMap.Geocoder()
    })
  } catch (e) {
    console.warn('AMap geocoder failed', e)
  }

  map = new AMap.Map('mapContainer', { center: [113.582, 22.352], zoom: 15 })
  marker = new AMap.Marker({ draggable: true, map })
  map.on('click', (e: any) => updateLocation(e.lnglat))
  marker.on('dragend', (e: any) => updateLocation(e.lnglat))
  setTimeout(() => map.resize(), 500)
  initAutoComplete()
}

function initAutoComplete() {
  const AMap = (window as any).AMap
  if (!AMap) return
  try {
    AMap.plugin(['AMap.AutoComplete', 'AMap.PlaceSearch'], () => {
      autoComplete = new AMap.AutoComplete({})
      placeSearch = new AMap.PlaceSearch({})
    })
  } catch (e) {
    console.warn('initAutoComplete error', e)
  }
}

let _keywordTimer: any = null
function onKeywordInput(val: string) {
  if (_keywordTimer) clearTimeout(_keywordTimer)
  if (!val) {
    suggestions.value = []
    return
  }
  _keywordTimer = setTimeout(() => {
    const AMap = (window as any).AMap
    if (!AMap || !autoComplete) return
    autoComplete.search(val, (status: any, result: any) => {
      if (status === 'complete' && result && result.tips) {
        suggestions.value = result.tips || []
      }
    })
  }, 250)
}

function formatTipAddress(tip: any) {
  const parts: string[] = []
  if (tip.district) parts.push(tip.district)
  if (tip.address) parts.push(tip.address)
  if (parts.length === 0 && tip.name) parts.push(tip.name)
  return parts.join(' ')
}

function selectSuggestion(item: any) {
  const name = item.name || ''
  const district = item.district || ''
  const address = item.address || ''
  let final = name
  if (district) final += ' ' + district
  if (address) final += ' ' + address

  addressForm.value.detail = final.trim()
  searchKeyword.value = addressForm.value.detail
  suggestions.value = []
  // 如果带 location，则直接更新
  if (item.location) {
    const parts = (item.location + '').split(',')
    const lng = Number(parts[0])
    const lat = Number(parts[1])
    updateLocation({ lng, lat })
  }
}

function updateLocation(lnglat: any) {
  if (!map || !marker) return
  addressForm.value.lng = lnglat.lng
  addressForm.value.lat = lnglat.lat
  marker.setPosition([lnglat.lng, lnglat.lat])
  if (geocoder) {
    geocoder.getAddress([lnglat.lng, lnglat.lat], (status: string, result: any) => {
      if (status === 'complete' && result && result.regeocode) {
        const formatted = result.regeocode.formattedAddress || ''
        addressForm.value.detail = formatted
      }
    })
  }
}

function locateCurrent() {
  if (!navigator.geolocation) {
    showToast('浏览器不支持定位')
    return
  }
  navigator.geolocation.getCurrentPosition(pos => {
    const lng = pos.coords.longitude
    const lat = pos.coords.latitude
    map && map.setCenter([lng, lat])
    updateLocation({ lng, lat })
  }, err => {
    console.warn('定位失败', err)
    showToast('定位失败')
  }, { enableHighAccuracy: true, timeout: 5000 })
}

// 保存地址（来自完整对话）
async function saveAddress() {
  if (!addressForm.value.detail || !addressForm.value.name) {
    showToast('请填写完整的收货信息')
    return
  }
  // 尝试使用地理信息（已在 updateLocation 中取得）
  const payload: any = {
    name: addressForm.value.name,
    phone: addressForm.value.phone,
    province: '',
    city: '',
    district: '',
    street: '',
    detail: addressForm.value.detail,
    tag: addressForm.value.tag,
    is_default: !!addressForm.value.isDefault,
    lng: addressForm.value.lng,
    lat: addressForm.value.lat,
  }
  try {
    const res: any = await addressApi.addAddress(payload)
    await loadAddresses()
    const maybe = addresses.value.find((x: any) => x.phone === addressForm.value.phone && x.name === addressForm.value.name)
    if (maybe) selectedAddress.value = maybe
    // 清理并关闭
    addressForm.value = { name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 }
    showAddDialog.value = false
    showToast('地址已添加')
  } catch (e) {
    console.error(e)
    showToast('新增地址失败')
  }
}

// 简洁表单保存（兼容保留）
async function saveAddressInline() {
  if (!addressForm.value.name || !addressForm.value.phone || !addressForm.value.detail) {
    showToast('请填写姓名、手机号和详细地址')
    return
  }
  try {
    const payload = {
      name: addressForm.value.name,
      phone: addressForm.value.phone,
      detail: addressForm.value.detail,
      province: '', city: '', district: '', street: '', isDefault: addressForm.value.isDefault ? 1 : 0
    }
    await addressApi.addAddress(payload)
    await loadAddresses()
    const maybe = addresses.value.find((x: any) => x.phone === addressForm.value.phone && x.name === addressForm.value.name)
    if (maybe) selectedAddress.value = maybe
    addressForm.value = { name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 }
    showAddAddress.value = false
    showToast('地址已添加')
  } catch (e) {
    console.error(e)
    showToast('新增地址失败')
  }
}

async function onPay() {
  if (!selectedAddress.value) {
    showToast('请先选择收货地址')
    return
  }

  if (shopList.value.length === 0) {
    showToast('购物车为空')
    return
  }

  try {
    // 若存在 pendingOrders（例如来自已有订单或已创建的 pending），直接标记这些订单为已支付
    if (pendingOrders.value && pendingOrders.value.length > 0) {
      payAmount.value = totalAmount.value || payAmount.value || 0
      payQrImg.value = qrImg
      showPayModal.value = true

      setTimeout(async () => {
        showPayModal.value = false
        for (const oid of pendingOrders.value) {
          try { await orderApi.payOrder(String(oid)) } catch (e) { console.warn('payOrder failed', e) }
        }
        // 清理购物车中已结算的项
        try { await cartApi.deleteSelected() } catch (e) {}
        await loadCart()
        router.push('/user/payment/success')
      }, 3000)
      return
    }

    // 否则走购物车结算流程（createPayOrder）
    const payloadShops = shopList.value.map(s => ({
      storeId: s.storeId,
      items: s.items.map((it: any) => ({ dishId: it.dishId, qty: it.qty }))
    }))

    const payload = {
      shops: payloadShops,
      consigneeAddressId: selectedAddress.value.id,
      remark: form.value.remark,
      tableware: form.value.tableware
    }

    const resp: any = await cartApi.checkout(payload)
    payAmount.value = totalAmount.value
    payQrImg.value = qrImg
    showPayModal.value = true

    setTimeout(async () => {
      showPayModal.value = false
      try {
        const orders = (resp && resp.data && (resp.data.orders || resp.data.orders)) || resp.orders || []
        for (const o of orders) {
          const oid = o.orderId || o.OrderID || o.orderID || o.id || o
          if (oid) {
            try { await orderApi.payOrder(String(oid)) } catch (e) { console.warn('payOrder failed', e) }
          }
        }
      } catch (e) { console.warn('mark orders paid failed', e) }

      await cartApi.deleteSelected()
      await loadCart()
      router.push('/user/payment/success')
    }, 3000)
  } catch (e: any) {
    showToast('下单失败：' + (e.message || ''))
  }
}

function closePayModal() {
  showPayModal.value = false
}
</script>

<style scoped>
.checkout-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 60px; /* 留出底部栏 */
  font-family: -apple-system, BlinkMacSystemFont, "Helvetica Neue", Helvetica, Arial, sans-serif;
}

/* 地址区域 */
.address-section {
  background: #fff;
  padding: 16px;
  margin-bottom: 8px;
}
.address-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.address-info {
  flex: 1;
}
.address-text {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}
.address-sub {
  font-size: 13px;
  color: #999;
}
.arrow {
  color: #ccc;
  font-size: 18px;
}
.delivery-time {
  margin-top: 12px;
  font-size: 14px;
  color: #666;
}

/* 订单详情 */
.order-section {
  background: #fff;
  margin-bottom: 8px;
}
.shop-block {
  padding: 12px 16px;
  border-bottom: 8px solid #f5f5f5;
}
.shop-name {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin-bottom: 12px;
}
.shop-name .van-icon {
  margin-right: 6px;
  color: #ffb400;
}
.dish-list {
  margin-bottom: 12px;
}
.dish-item {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
}
.dish-name {
  flex: 1;
  margin-right: 8px;
}
.dish-spec {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}
.dish-quantity {
  color: #999;
  margin-right: 12px;
}
.dish-price {
  font-weight: 600;
  min-width: 70px;
  text-align: right;
}
.fee-row {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}
.shop-total {
  display: flex;
  justify-content: space-between;
  font-size: 15px;
  font-weight: 600;
  color: #333;
  padding-top: 8px;
}
.shop-total .price {
  color: #ff6600;
}

/* 备注 & 餐具 */
.extra-section {
  padding: 16px;
  background: #fff;
}
.remark-input {
  margin-bottom: 16px;
}
.tableware-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 15px;
}

/* 底部支付栏 */
.total-info {
  flex: 1;
  padding-left: 16px;
}
.total-label {
  font-size: 13px;
  color: #999;
}
.total-price {
  font-size: 18px;
  font-weight: 700;
  color: #ff6600;
}
.pay-btn {
  width: 120px;
  height: 100%;
  background: #ffb400;
  color: #fff;
  font-size: 17px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 支付弹窗 */
.pay-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}
.pay-modal {
  background: #fff;
  padding: 24px;
  border-radius: 12px;
  text-align: center;
  width: 280px;
}
.pay-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
}
.pay-amount {
  font-size: 20px;
  font-weight: 700;
  color: #ff6600;
}

/* 地址卡片样式（从 address/index.vue 迁移，用于结算页内的地址列表） */
.addr-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.addr-card {
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  transition: all 0.3s ease;
  padding: 12px;
}

.addr-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.addr-top { margin-bottom: 8px }
.addr-detail { color: #666; font-size: 13px; }
.addr-phone { color: #999; margin-top: 6px }
.default-tag { display:inline-block; padding:3px 8px; border-radius:6px; font-size:12px; background:#fff3e0; color:#ff9800; margin-left:8px }

/* Card wrapper: 中心卡片，宽度 60% */
.checkout-wrapper {
  display: flex;
  justify-content: center;
  padding: 40px 0 120px; /* 底部留出支付栏位置 */
  background: transparent;
}
.checkout-card {
  width: 60%;
  max-width: 980px;
  background: rgba(255, 248, 225, 0.98);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(255, 193, 7, 0.15);
  padding: 20px;
}

/* 底部支付栏：居中并与卡片对齐 */
.bottom-bar {
  position: fixed;
  bottom: 60px;
  left: 50%;
  transform: translateX(-50%);
  width: 60%;
  max-width: 980px;
  height: 64px;
  background: #fff;
  display: flex;
  align-items: center;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  border-radius: 12px;
  z-index: 120;
}
.bottom-bar .total-info { padding-left: 24px }
.bottom-bar .pay-btn { width: 160px;border-radius: 12px;cursor: pointer;hover { background: #fcca7e } }

/* 地图 / 地址表单样式（完整迁移，避免挤在一起） */
.map-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  margin-top: 10px;
}
.map-container-wrap {
  position: relative;
  width: 100%;
  height: 400px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
.map-container { width:100%; height:100%; background: linear-gradient(180deg, #f8fafb, #eef2f6); }
.map-controls { position: absolute; right: 16px; bottom: 16px; z-index:110 }
.locate-btn-fixed { background: #fff !important; box-shadow: 0 2px 8px rgba(0,0,0,0.15); border:none !important }
.locate-btn-fixed:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.2) }
.locate-icon { width:18px; height:18px }
.search-panel { position: relative; width:100% }
.map-search-input .el-input__inner { height:42px; padding-left:34px !important; box-sizing:border-box }
.suggestion-box { position:absolute; top:100%; left:0; right:0; margin-top:4px; background:#fff; border-radius:8px; box-shadow:0 4px 16px rgba(0,0,0,0.1); max-height:300px; overflow-y:auto; z-index:2000 }
.suggestion-item { display:flex; align-items:center; justify-content:space-between; padding:12px 16px; cursor:pointer; transition:all .2s; border-bottom:1px solid rgba(0,0,0,0.06) }
.suggestion-item:hover { background:#f5f7fa }
.suggestion-name { font-size:14px; font-weight:600; color:#2c3e50; margin-bottom:4px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis }
.suggestion-address { font-size:12px; color:#666; white-space:nowrap; overflow:hidden; text-overflow:ellipsis }
.final-address { margin-top:12px; padding:12px; background:#f8fafc; border:1px solid #e4e7ed; border-radius:8px }
.final-address .label { font-size:13px; color:#909399; margin-bottom:6px }
.final-address .address-text { color:#2c3e50; line-height:1.5; word-break:break-all; white-space:pre-wrap }

.floating-form { display:flex; flex-direction:column; gap:24px; margin-top:20px }
.form-item { position:relative; width:100% }
.form-item input, .form-item select { width:100%; padding:16px; font-size:15px; border:1px solid #dcdfe6; border-radius:8px; background:white }
.form-item input[readonly] { background-color:white; cursor:default }
.form-item label { position:absolute; left:16px; top:50%; transform:translateY(-50%); font-size:15px; color:#909399; pointer-events:none; transition:0.2s; background:white; padding:0 4px }
.form-item.has-value label, .form-item label.active { top:0; font-size:12px; color:#409EFF; transform:translateY(-50%) }


</style>