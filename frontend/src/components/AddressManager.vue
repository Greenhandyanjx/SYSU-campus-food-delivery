<template>
  <!-- Dialog 1: Address List -->
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="选择收货地址"
    width="600px"
    :close-on-click-modal="false"
    @opened="loadAddresses"
  >
    <div class="addr-modal">
      <div class="addr-modal-header">
        <span style="font-weight:600">我的收货地址</span>
        <el-button type="primary" size="small" @click="openCreateDialog">+ 新增地址</el-button>
      </div>
      <div v-if="loading" style="text-align:center;padding:24px;color:#999">加载中...</div>
      <div v-else-if="addresses.length === 0" style="text-align:center;padding:24px;color:#999">
        暂无收货地址，请先在「我的地址」中添加
      </div>
      <div v-else class="addr-scroll-list">
        <div class="addr-list">
          <el-card
            v-for="a in addresses"
            :key="a.id"
            :class="['addr-card', { selected: selectedId === a.id }]"
            shadow="hover"
            @click="selectedId = a.id"
          >
            <div class="addr-top">
              <strong>{{ a.name }}</strong>
              <span style="margin-left:8px">{{ a.phone }}</span>
              <span v-if="a.isDefault" class="default-tag">默认</span>
            </div>
            <div class="addr-detail">{{ formatAddr(a) }}</div>
          </el-card>
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="closeDialog">取消</el-button>
      <el-button type="primary" :disabled="!selectedId" @click="confirmAddress">确认</el-button>
    </template>
  </el-dialog>

  <!-- Dialog 2: Map-based Address Creation -->
  <el-dialog
    :model-value="showCreate"
    @update:model-value="showCreate = $event"
    width="700px"
    class="dialog-box"
    append-to-body
    :close-on-click-modal="false"
    @opened="initMapInline"
    @closed="destroyMap"
  >
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
            <el-button
              class="locate-btn-fixed"
              type="primary"
              size="small"
              circle
              title="定位到当前位置"
              @click="locateCurrent"
            >
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="3"/>
                <path d="M12 2v4M12 18v4M2 12h4M18 12h4"/>
              </svg>
            </el-button>
          </div>
        </div>

        <div class="search-panel">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索地点、小区或街道名称"
            clearable
            class="map-search-input"
            @input="onKeywordInput"
          />
          <div v-if="suggestions.length" class="suggestion-box">
            <div
              v-for="(item, idx) in suggestions"
              :key="idx"
              class="suggestion-item"
              @click="selectSuggestion(item)"
            >
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

    <div class="floating-form address-manager-floating-form">
      <div class="form-item" :class="{ 'has-value': addressForm.name }">
        <input type="text" v-model="addressForm.name" id="am-name" required />
        <label for="am-name" :class="{ active: addressForm.name }">收货人</label>
      </div>
      <div class="form-item" :class="{ 'has-value': addressForm.phone }">
        <input type="tel" v-model="addressForm.phone" id="am-phone" required />
        <label for="am-phone" :class="{ active: addressForm.phone }">手机号码</label>
      </div>
      <div class="form-item" :class="{ 'has-value': addressForm.detail }">
        <input type="text" v-model="addressForm.detail" id="am-address" required readonly />
        <label for="am-address" :class="{ active: addressForm.detail }">收货地址</label>
      </div>
      <div class="form-item" :class="{ 'has-value': addressForm.tag }">
        <select v-model="addressForm.tag" id="am-tag" required>
          <option value="" disabled selected></option>
          <option value="家">家</option>
          <option value="公司">公司</option>
          <option value="学校">学校</option>
        </select>
        <label for="am-tag" :class="{ active: addressForm.tag }">地址标签</label>
      </div>
    </div>

    <template #footer>
      <el-button @click="showCreate = false">取消</el-button>
      <el-button type="primary" @click="saveAddress">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { listAddresses, addAddress } from '@/api/common/address'

defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: [address: any]
}>()

// ── Address List State ──
const addresses = ref<any[]>([])
const selectedId = ref<number>(0)
const loading = ref(false)

// ── Create Dialog State ──
const showCreate = ref(false)
const addressForm = ref({ name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 })
const searchKeyword = ref('')
const suggestions = ref<any[]>([])

// ── AMap Instances ──
let map: any = null
let marker: any = null
let geocoder: any = null
let autoComplete: any = null
let placeSearch: any = null
let _keywordTimer: any = null

async function loadAddresses() {
  loading.value = true
  try {
    // 先检查 token 是否存在
    const storedToken = localStorage.getItem('token')
    if (!storedToken) {
      ElMessage.warning('请先登录')
      loading.value = false
      return
    }

    const res: any = await listAddresses()

    if (res && res.code === 1) {
      addresses.value = Array.isArray(res.data) ? res.data : []
    } else if (res && (res.code === 401 || res.code === '401')) {
      ElMessage.warning('登录已过期，请刷新页面重新登录')
    } else {
      addresses.value = []
      if (res && res.msg) ElMessage.error(res.msg)
    }
  } catch (err) {
    console.error('[AddressManager] loadAddresses error:', err)
    ElMessage.error('加载地址失败')
    addresses.value = []
  } finally {
    loading.value = false
  }
}

function formatAddr(a: any) {
  const parts = [a.province || '', a.city || '', a.district || '', a.street || '', a.detail || ''].filter(Boolean)
  return parts.join(' ') || '暂无详细地址'
}

function confirmAddress() {
  const addr = addresses.value.find(a => a.id === selectedId.value)
  if (addr) {
    emit('confirm', addr)
    emit('update:visible', false)
  }
}

function closeDialog() {
  emit('update:visible', false)
}

function openCreateDialog() {
  addressForm.value = { name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 }
  searchKeyword.value = ''
  suggestions.value = []
  showCreate.value = true
}

// ── AMap Functions ──

function initMapInline() {
  setTimeout(initMap, 200)
}

function initMap() {
  const amapKey = (import.meta.env.VITE_AMAP_KEY as string) || ''
  const old = document.getElementById('mapContainer')
  if (!old) return
  old.innerHTML = ''

  const AMap = (window as any).AMap
  if (!AMap) {
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

  map = new AMap.Map('mapContainer', { center: [113.58399174, 22.34937810], zoom: 15 })
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
    ElMessage.warning('浏览器不支持定位')
    return
  }
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      const lng = pos.coords.longitude
      const lat = pos.coords.latitude
      map && map.setCenter([lng, lat])
      updateLocation({ lng, lat })
    },
    (err) => {
      console.warn('定位失败', err)
      ElMessage.warning('定位失败')
    },
    { enableHighAccuracy: true, timeout: 5000 }
  )
}

// ── Geolocation Helpers ──

function splitDetailToParts(detail: string) {
  if (!detail) return { province: '', city: '', district: '', street: '' }
  const txt = detail.trim()
  const re = /^(.*?(?:省|自治区|特别行政区|市))?\s*(.*?(?:市|自治州|地区|盟))?\s*(.*?(?:区|县|市|旗))?\s*(.*)$/
  const m = txt.match(re)
  if (m) {
    const province = (m[1] || '').trim()
    const city = (m[2] || '').trim()
    const district = (m[3] || '').trim()
    const street = (m[4] || '').trim()
    if (province || city || district) return { province, city, district, street }
  }
  const tokens = detail.split(/[,，\s]+/).filter(Boolean).map(s => s.trim())
  return { province: tokens[0] || '', city: tokens[1] || '', district: tokens[2] || '', street: tokens.slice(3).join(' ') || '' }
}

function stripPrefixFromDetail(formatted: string, province: string, city: string, district: string, street: string) {
  if (!formatted) return ''
  const parts = [province || '', city || '', district || '', street || ''].filter(Boolean)
  if (parts.length === 0) return formatted.trim()
  const escapeRegex = (s: string) => s.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&')
  const pattern = '^\\s*' + parts.map(p => escapeRegex(p)).join('[\\s,，]*') + '[\\s,，]*'
  try {
    const re = new RegExp(pattern)
    const stripped = formatted.replace(re, '')
    return (stripped || '').trim()
  } catch (e) {
    return formatted.trim()
  }
}

async function geocodeAddress(detail: string) {
  if (!detail) return { province: '', city: '', district: '', street: '', lng: 0, lat: 0, formatted: '' }
  const AMap = (window as any).AMap
  if (!AMap || !geocoder) {
    return { ...splitDetailToParts(detail), lng: 0, lat: 0, formatted: detail }
  }
  return await new Promise<any>((resolve) => {
    try {
      geocoder.getLocation(detail, (status: string, result: any) => {
        if (status === 'complete' && result && result.geocodes && result.geocodes.length) {
          const g = result.geocodes[0]
          const comp = g.addressComponent || {}
          const province = comp.province || g.province || ''
          const city = comp.city || g.city || ''
          const district = comp.district || g.district || ''
          const street = comp.township || comp.street || (comp.streetNumber && comp.streetNumber.street) || g.township || g.street || ''
          let lng = 0; let lat = 0
          if (g.location) {
            if (typeof g.location === 'string') {
              const parts = g.location.split(',')
              lng = parseFloat(parts[0]) || 0; lat = parseFloat(parts[1]) || 0
            } else if (g.location.lng && g.location.lat) {
              lng = g.location.lng; lat = g.location.lat
            }
          }
          resolve({ province, city, district, street, lng, lat, formatted: g.formattedAddress || g.formatted || detail })
          return
        }
        resolve({ ...splitDetailToParts(detail), lng: 0, lat: 0, formatted: detail })
      })
    } catch (e) {
      resolve({ ...splitDetailToParts(detail), lng: 0, lat: 0, formatted: detail })
    }
  })
}

async function saveAddress() {
  if (!addressForm.value.detail || !addressForm.value.name) {
    ElMessage.warning('请填写完整的收货信息')
    return
  }
  const geo = await geocodeAddress(addressForm.value.detail || '')
  const detailStripped = stripPrefixFromDetail(geo.formatted || addressForm.value.detail || '', geo.province || '', geo.city || '', geo.district || '', geo.street || '')
  const payload: any = {
    name: addressForm.value.name,
    phone: addressForm.value.phone,
    province: geo.province || '',
    city: geo.city || '',
    district: geo.district || '',
    street: geo.street || '',
    detail: detailStripped || addressForm.value.detail,
    tag: addressForm.value.tag,
    is_default: !!addressForm.value.isDefault,
    lng: geo.lng || addressForm.value.lng,
    lat: geo.lat || addressForm.value.lat,
  }
  try {
    const res: any = await addAddress(payload)
    await loadAddresses()
    const maybe = addresses.value.find((x: any) => x.phone === addressForm.value.phone && x.name === addressForm.value.name)
    if (maybe) selectedId.value = maybe.id
    addressForm.value = { name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 }
    showCreate.value = false
    ElMessage.success('地址已添加')
  } catch (e) {
    console.error(e)
    ElMessage.error('新增地址失败')
  }
}

function destroyMap() {
  map = null
  marker = null
  geocoder = null
  autoComplete = null
  placeSearch = null
}
</script>

<style scoped>
/* ── Address List ── */
.addr-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.addr-scroll-list {
  max-height: 400px;
  overflow-y: auto;
}
.addr-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}
.addr-card {
  border: 1px solid #e4e7ed;
  border-radius: 12px;
  transition: all 0.3s ease;
  padding: 12px;
  cursor: pointer;
}
.addr-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
.addr-card.selected {
  border-color: #FF6B35;
  background: #FFF3E8;
}
.addr-top {
  margin-bottom: 8px;
}
.addr-detail {
  color: #666;
  font-size: 13px;
}
.default-tag {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 6px;
  font-size: 12px;
  background: #fff3e0;
  color: #ff9800;
  margin-left: 8px;
}

/* ── Dialog Title ── */
.dialog-title {
  display: flex;
  align-items: center;
  gap: 12px;
}
.dialog-title-icon {
  font-size: 28px;
}
.dialog-title-main {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}
.dialog-title-sub {
  font-size: 13px;
  color: #999;
  margin-top: 4px;
}

/* ── Map ── */
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
.map-container {
  width: 100%;
  height: 100%;
  background: linear-gradient(180deg, #f8fafb, #eef2f6);
}
.map-controls {
  position: absolute;
  right: 16px;
  bottom: 16px;
  z-index: 110;
}
.locate-btn-fixed {
  background: #fff !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  border: none !important;
}
.locate-btn-fixed:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
.search-panel {
  position: relative;
  width: 100%;
}
.map-search-input :deep(.el-input__inner) {
  height: 42px;
  padding-left: 34px !important;
  box-sizing: border-box;
}
.suggestion-box {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 4px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  max-height: 300px;
  overflow-y: auto;
  z-index: 2000;
}
.suggestion-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.2s;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}
.suggestion-item:hover {
  background: #f5f7fa;
}
.suggestion-name {
  font-size: 14px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.suggestion-address {
  font-size: 12px;
  color: #666;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.final-address {
  margin-top: 12px;
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}
.final-address .label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 6px;
}
.final-address .address-text {
  color: #2c3e50;
  line-height: 1.5;
  word-break: break-all;
  white-space: pre-wrap;
}
.floating-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
  margin-top: 20px;
}
.form-item {
  position: relative;
  width: 100%;
}
.form-item input,
.form-item select {
  width: 100%;
  padding: 16px;
  font-size: 15px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: white;
  box-sizing: border-box;
}
.form-item input[readonly] {
  background-color: white;
  cursor: default;
}
.form-item label {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 15px;
  color: #909399;
  pointer-events: none;
  transition: 0.2s;
  background: white;
  padding: 0 4px;
}
.form-item.has-value label,
.form-item label.active {
  top: 0;
  font-size: 12px;
  color: #409EFF;
  transform: translateY(-50%);
}
</style>

<!-- 非 scoped 样式：第二个 el-dialog 使用 append-to-body teleport 到 body，scoped CSS 不作用于 teleport 内的元素 -->
<style>
.address-manager-floating-form .form-item {
  position: relative;
  width: 100%;
}
.address-manager-floating-form .form-item input,
.address-manager-floating-form .form-item select {
  width: 100%;
  padding: 16px;
  font-size: 15px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  background: white;
  box-sizing: border-box;
}
.address-manager-floating-form .form-item input[readonly] {
  background-color: white;
  cursor: default;
}
.address-manager-floating-form .form-item label {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 15px;
  color: #909399;
  pointer-events: none;
  transition: 0.2s;
  background: white;
  padding: 0 4px;
}
.address-manager-floating-form .form-item.has-value label,
.address-manager-floating-form .form-item label.active {
  top: 0;
  font-size: 12px;
  color: #409EFF;
  transform: translateY(-50%);
}
</style>
