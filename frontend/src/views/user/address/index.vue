<template>
  <div class="address-page">
    <div class="address-header">
      <h2 class="page-title" style="margin-right: 10px;">地址管理</h2>
      <div class="header-actions">
        <el-input
          v-model="searchQuery"
          placeholder="搜索地址或标签"
          prefix-icon="Search"
          size="small"
          class="search-bar"
          @keyup.enter="applySearch"
        />
        <el-button type="primary" round class="add-btn" @click="openAdd"> 新增地址</el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" stretch>
      <el-tab-pane label="我的收货地址" name="mine">
        <div class="addr-list">
          <el-card
            v-for="(a, i) in filteredAddresses"
            :key="i"
            class="addr-card"
            shadow="hover"
          >
            <div class="addr-row">
              <div>
                <div class="addr-name">
                  {{ a.name }}
                  <span :class="['addr-tag', tagColor(a.tag)]">{{ a.tag }}</span>
                  <span v-if="a.isDefault" class="default-tag">默认</span>
                </div>
                                <div class="addr-detail">{{ formatAddress(a) }}</div>
                <div class="addr-phone">{{ a.phone }}</div>
              </div>
              <div class="addr-actions">
                <el-button text size="small" @click="setDefault(i)">设为默认</el-button>
                <el-button text size="small" @click="editAddress(i)">编辑</el-button>
                <el-button text size="small" @click="removeAddress(i)">删除</el-button>
              </div>
            </div>
          </el-card>
          <div v-if="filteredAddresses.length === 0" class="empty">
            暂无匹配地址，点击“新增地址”添加。
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="附近地址" name="nearby">
        <div class="nearby-list">
          <el-card
            v-for="(a, i) in nearbyAddresses"
            :key="i"
            class="addr-card"
            shadow="hover"
          >
            <div class="addr-row">
              <div>
                <div class="addr-name">{{ a.name }}</div>
                  <div class="addr-detail">{{ a.detail }} <span class="addr-distance" v-if="a.distanceText">· {{ a.distanceText }}</span></div>
              </div>
              <div class="addr-actions">
                <el-button type="primary" size="small" round @click="useNearby(a)">选择</el-button>
              </div>
            </div>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 新增地址弹窗 -->
    <el-dialog v-model="showDialog" width="700px" class="dialog-box" @opened="initMap">
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
          <!-- 操作引导提示
          <div class="map-guide">
            <el-alert
              title="选择收货地址的方法"
              type="info"
              :closable="false"
              class="guide-alert"
            >
              <template #default>
                <ol class="guide-steps">
                  <li>方式一：在地图上<strong>点击</strong>或<strong>拖动图标</strong>到具体位置</li>
                  <li>方式二：在下方搜索框<strong>输入地址关键词</strong>后从建议列表选择</li>
                  <li>方式三：点击右下角<strong>定位图标</strong>快速定位到当前位置</li>
                </ol>
              </template>
            </el-alert>
          </div> -->

          <!-- 大地图容器 -->
          <div class="map-container-wrap">
            <div id="mapContainer" class="map-container"></div>
            <!-- 右下角定位按钮 -->
            <div class="map-controls">
              <el-button
                class="locate-btn-fixed"
                type="primary"
                size="small"
                circle
                title="定位到当前位置"
                @click="locateCurrent"
              >
                <img src="@/assets/icons/icon_locate.svg" class="locate-icon" alt="定位" />
              </el-button>
            </div>
          </div>

          <!-- 搜索区域 -->
          <div class="search-panel">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索地点、小区或街道名称"
              clearable
              class="map-search-input"
              @input="onKeywordInput"
            >
              <!-- <template #prefix>
                <el-icon ><Search /></el-icon>
              </template> -->
            </el-input>
            
            <!-- 候选项 -->
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
                <el-icon class="suggestion-icon"><Location /></el-icon>
              </div>
            </div>

            <!-- 下面显示最终选中的详细地址（可长文本换行） -->
            <div class="final-address" v-if="form.detail">
              <div class="label">已选择地址：</div>
              <div class="address-text">{{ form.detail }}</div>
            </div>
          </div>
        </div>
      </el-form-item>
      <div class="floating-form">
        <div class="form-item" :class="{ 'has-value': form.name }">
          <input type="text" v-model="form.name" id="name" required>
          <label for="name" :class="{ 'active': form.name }">收货人</label>
        </div>

        <div class="form-item" :class="{ 'has-value': form.phone }">
          <input type="tel" v-model="form.phone" id="phone" required>
          <label for="phone" :class="{ 'active': form.phone }">手机号码</label>
        </div>

        <div class="form-item" :class="{ 'has-value': form.detail }">
          <input type="text" v-model="form.detail" id="address" required readonly>
          <label for="address" :class="{ 'active': form.detail }">收货地址</label>
        </div>

        <div class="form-item" :class="{ 'has-value': form.tag }">
          <select v-model="form.tag" id="tag" required>
            <option value="" disabled selected></option>
            <option value="家">家</option>
            <option value="公司">公司</option>
            <option value="学校">学校</option>
          </select>
          <label for="tag" :class="{ 'active': form.tag }">地址标签</label>
        </div>
      </div>
      <template #footer>
        <el-button @click="closeDialog">取消</el-button>
        <el-button type="primary" @click="saveAddress">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listAddresses, addAddress, editAddress as apiEditAddress, setDefaultAddress, deleteAddress } from '@/api/common/address'
(window as any)._AMapSecurityConfig = {
  securityJsCode: '4bf89c4e16d60340e676f6cc39beff32'
}
const activeTab = ref('mine')
const route = useRoute()
const router = useRouter()
const showDialog = ref(false)
const searchQuery = ref('')

const myAddresses = ref<any[]>([])
const editingId = ref<number | null>(null)

const nearbyAddresses = ref<any[]>([
  { name: '教学楼南门', detail: '中山大学南门旁' },
  { name: '学生食堂', detail: '第一食堂附近' },
])

const form = ref({ name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 })

// helpers: split a raw detail string into address parts, and format for display
const splitDetailToParts = (detail: string) => {
  if (!detail) return { province: '', city: '', district: '', street: '' }
  // split by common delimiters: space, comma, chinese comma
  const parts = detail.split(/[\s,，]+/).filter(Boolean)
  const province = parts[0] || ''
  const city = parts[1] || ''
  const district = parts[2] || ''
  const street = parts.slice(3).join(' ') || ''
  return { province, city, district, street }
}

const formatAddress = (a: any) => {
  if (!a) return ''
  const pieces = [] as string[]
  if (a.province) pieces.push(a.province)
  if (a.city) pieces.push(a.city)
  if (a.district) pieces.push(a.district)
  if (a.street) pieces.push(a.street)
  if (a.detail) pieces.push(a.detail)
  return pieces.filter(Boolean).join(' ')
}

// Use AMap geocoder (if available) to parse a free-text address into structured fields.
// Falls back to splitDetailToParts when geocoder is not ready or fails.
const geocodeAddress = async (detail: string) => {
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
          // AMap geocode fields: province, city, district, township, street, formattedAddress, location
              // AMap geocoder returns detailed parts inside `addressComponent`
              const comp = g.addressComponent || {}
              const province = comp.province || g.province || ''
              const city = comp.city || g.city || ''
              const district = comp.district || g.district || ''
              // street may be in township/street/streetNumber
              const street = comp.township || comp.street || (comp.streetNumber && comp.streetNumber.street) || g.township || g.street || ''
          let lng = 0
          let lat = 0
          if (g.location) {
            // location may be in format 'lng,lat' or an object
            if (typeof g.location === 'string') {
              const parts = g.location.split(',')
              lng = parseFloat(parts[0]) || 0
              lat = parseFloat(parts[1]) || 0
            } else if (g.location.lng && g.location.lat) {
              lng = g.location.lng
              lat = g.location.lat
            }
          }
          const formatted = g.formattedAddress || g.formatted || detail
          resolve({ province, city, district, street, lng, lat, formatted })
          return
        }
        // fallback
        resolve({ ...splitDetailToParts(detail), lng: 0, lat: 0, formatted: detail })
      })
    } catch (e) {
      // any error -> fallback
      resolve({ ...splitDetailToParts(detail), lng: 0, lat: 0, formatted: detail })
    }
  })
}

function openAdd() {
  editingId.value = null
  form.value = { name: '', phone: '', detail: '', tag: '', isDefault: false, lng: 0, lat: 0 }
  showDialog.value = true
  nextTick(() => {
    setTimeout(initMap, 300)
  })
}
function closeDialog() { showDialog.value = false }

async function saveAddress() {
  if (!form.value.detail || !form.value.name) {
    ElMessage.warning('请填写完整的收货信息')
    return
  }
    // Prefer geocoding via AMap for better structured fields; fallback to simple split
    const geo = await geocodeAddress(form.value.detail)
    // If geocoder returned formatted value, update form lng/lat
    if (geo.lng) form.value.lng = geo.lng
    if (geo.lat) form.value.lat = geo.lat

    // Strip the high-level parts from the formatted address so `detail` only contains
    // the specific street/building/room info (we already store province/city/district/street separately).
    const strippedDetail = stripPrefixFromDetail(geo.formatted || form.value.detail, geo.province || '', geo.city || '', geo.district || '', geo.street || '')
    const payload: any = {
      name: form.value.name,
      phone: form.value.phone,
      province: geo.province || '',
      city: geo.city || '',
      district: geo.district || '',
      street: geo.street || '',
      detail: strippedDetail || (geo.formatted || form.value.detail),
      tag: form.value.tag,
      is_default: !!form.value.isDefault,
      lng: form.value.lng,
      lat: form.value.lat,
    }

  try {
    let res: any
    if (editingId.value) {
      res = await apiEditAddress(editingId.value, payload)
    } else {
      res = await addAddress(payload)
    }
    if (res && res.code === 1) {
      ElMessage.success('保存成功')
      showDialog.value = false
      await fetchAddresses()
      // 完成后返回地址管理主界面（不自动回跳）
    } else {
      ElMessage.error(res?.msg || '保存失败')
    }
  } catch (err) {
    console.error(err)
    ElMessage.error('保存地址时发生错误')
  }
}

// Remove leading province/city/district/street from a formatted address string.
// This helps keep `detail` focused on the specific street/room info instead of repeating
// the high-level administrative parts which are stored separately.
const stripPrefixFromDetail = (formatted: string, province: string, city: string, district: string, street: string) => {
  if (!formatted) return ''
  const parts = [province || '', city || '', district || '', street || '']
    .map(p => (p || '').trim())
    .filter(Boolean)
  if (parts.length === 0) return formatted.trim()

  // Build a regex that allows optional separators (space, comma, Chinese comma) between parts
  const escapeRegex = (s: string) => s.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&')
  const pattern = '^\\s*' + parts.map(p => escapeRegex(p)).join('[\\s,，]*') + '[\\s,，]*'
  try {
    const re = new RegExp(pattern)
    const stripped = formatted.replace(re, '')
    return (stripped || '').trim()
  } catch (e) {
    // If regex construction fails for any reason, just try a naive replace of joined parts
    const naive = parts.join('')
    return formatted.replace(naive, '').trim()
  }
}

function editAddress(i: number) {
  const a = myAddresses.value[i]
  if (!a) return
  editingId.value = a.id || null
  form.value = { ...a }
  showDialog.value = true
  nextTick(initMap)
}

async function removeAddress(i: number) {
  const a = myAddresses.value[i]
  if (!a) return
  const id = a.id
  if (!id) {
    myAddresses.value.splice(i, 1)
    return
  }
  try {
    const res: any = await deleteAddress(id)
    if (res && res.code === 1) {
      ElMessage.success('删除成功')
      await fetchAddresses()
    } else {
      ElMessage.error(res?.msg || '删除失败')
    }
  } catch (err) {
    console.error(err)
    ElMessage.error('删除地址失败')
  }
}

function useNearby(a: any) {
  // 尝试从 localStorage 获取用户默认联系人信息（若有）以便快速添加
  const defaultName = localStorage.getItem('userName') || ''
  const defaultPhone = localStorage.getItem('userPhone') || ''
  const nameToUse = defaultName || a.name || ''
  const phoneToUse = defaultPhone || ''
  myAddresses.value.push({ name: nameToUse, phone: phoneToUse, detail: a.detail, tag: '附近', isDefault: false })
  activeTab.value = 'mine'
}

async function setDefault(i: number) {
  const a = myAddresses.value[i]
  if (!a || !a.id) return
  try {
    const res: any = await setDefaultAddress(a.id)
    if (res && res.code === 1) {
      ElMessage.success('已设为默认地址')
      await fetchAddresses()
    } else {
      ElMessage.error(res?.msg || '操作失败')
    }
  } catch (err) {
    console.error(err)
    ElMessage.error('设置默认地址失败')
  }
}

async function fetchAddresses() {
  try {
    const res: any = await listAddresses()
    if (res && res.code === 1) {
      myAddresses.value = Array.isArray(res.data) ? res.data : []
    } else {
      myAddresses.value = []
      // 仅在有错误信息时显示
      if (res && res.msg) ElMessage.error(res.msg)
    }
  } catch (err) {
    console.error(err)
    ElMessage.error('加载地址失败')
    myAddresses.value = []
  }
}

const filteredAddresses = computed(() =>
  myAddresses.value.filter(a => {
    const q = (searchQuery.value || '').trim()
    if (!q) return true
    const name = (a.name || '') as string
    const tag = (a.tag || '') as string
    const addrText = formatAddress(a)
    return (
      name.includes(q) ||
      addrText.includes(q) ||
      tag.includes(q)
    )
  })
)

function applySearch() {
  // 入口由 Enter 键触发：将结果面板切换到我的地址，computed 会自动更新列表
  activeTab.value = 'mine'
}

function tagColor(tag: string) {
  return {
    '家': 'tag-home',
    '公司': 'tag-work',
    '学校': 'tag-school',
    '附近': 'tag-near',
  }[tag] || 'tag-default'
}
// 高德地图相关逻辑
let map: any, marker: any, geocoder: any
function initMap() {
  // 从 Vite 环境变量读取高德 key 与安全码
  // 请在项目根创建本地 .env 或 .env.local 并添加 VITE_AMAP_KEY 与 VITE_AMAP_SECURITY_CODE
  const amapKey = (import.meta.env.VITE_AMAP_KEY as string) || ''
  const amapSec = (import.meta.env.VITE_AMAP_SECURITY_CODE as string) || ''

  if (amapSec) {
    ;(window as any)._AMapSecurityConfig = {
      securityJsCode: amapSec,
    }
  } else {
    // 若未配置安全码，仅打印提示；高德控制台可通过 referer 限制来保护 key
    console.warn('VITE_AMAP_SECURITY_CODE 未配置，建议在本地 .env 中设置安全码以增强安全性')
  }

  const AMapScriptUrl = `https://webapi.amap.com/maps?v=2.0&key=${amapKey}`
  const old = document.getElementById('mapContainer')
  if (!old) return
  old.innerHTML = '' // 清空旧内容

  // 确保容器可见后初始化
  setTimeout(() => {
    if (!(window as any).AMap) {
      const script = document.createElement('script')
      script.src = AMapScriptUrl
      script.onload = setupMap
      document.head.appendChild(script)
    } else {
      setupMap()
    }
  }, 300)
}


function setupMap() {
  const AMap = (window as any).AMap
  console.log('✅ setupMap 初始化成功')

  // ✅ 保证 Geocoder 插件加载
  AMap.plugin('AMap.Geocoder', () => {
    geocoder = new AMap.Geocoder({
      city: '全国', // 可选：限制查询范围
    })
    console.log('🟢 Geocoder 已加载')
  })

  map = new AMap.Map('mapContainer', {
    zoom: 15,
    center: [113.582, 22.352],
  })

  marker = new AMap.Marker({
    position: [113.582, 22.352],
    draggable: true,
    map,
  })

  map.on('click', (e: any) => updateLocation(e.lnglat))
  marker.on('dragend', (e: any) => updateLocation(e.lnglat))

  // ✅ 地图显示修正
  setTimeout(() => map.resize(), 500)
  initAutoComplete()
}
let autoComplete: any, placeSearch: any
const searchKeyword = ref('')      // 输入关键字（单独控制）
const suggestions = ref<any[]>([])

// 计算两个经纬度点之间的距离（米），使用 Haversine 公式
function distanceMeters(lon1: number, lat1: number, lon2: number, lat2: number) {
  const toRad = (deg: number) => (deg * Math.PI) / 180
  const R = 6371000 // 地球半径（米）
  const dLat = toRad(lat2 - lat1)
  const dLon = toRad(lon2 - lon1)
  const a = Math.sin(dLat / 2) * Math.sin(dLat / 2) + Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) * Math.sin(dLon / 2) * Math.sin(dLon / 2)
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
  return R * c
}

function formatDistance(meters: number) {
  if (meters < 1000) return Math.round(meters) + 'm'
  return (meters / 1000).toFixed(1) + 'km'
}

// 初始化 AutoComplete + PlaceSearch（在 setupMap() 完成后调用 initAutoComplete()）
function initAutoComplete() {
  const AMap = (window as any).AMap
  if (!AMap) return
  AMap.plugin(['AMap.AutoComplete', 'AMap.PlaceSearch'], () => {
    // AutoComplete 用于获取 tips (提示)
    autoComplete = new AMap.AutoComplete({
      city: '全国',
      // 不直接绑定 DOM input id：我们用 programmatic search
    })
    // PlaceSearch 用于进一步查询详情（可选，不必须）
    placeSearch = new AMap.PlaceSearch({
      city: '全国',
      // map, // 不自动渲染到地图，除非需要
    })
    console.log('🟢 AutoComplete & PlaceSearch 已初始化')
  })
}

// 当用户输入关键字时调用（带简单去抖）
let _keywordTimer: any = null
function onKeywordInput(val: string) {
  if (_keywordTimer) clearTimeout(_keywordTimer)
  if (!val || !autoComplete) {
    suggestions.value = []
    return
  }
  _keywordTimer = setTimeout(() => {
    autoComplete.search(val, (status: string, result: any) => {
      if (status === 'complete' && result?.tips) {
        suggestions.value = result.tips.filter((t: any) => !!t.location)
      } else {
        suggestions.value = []
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

  form.value.detail = final.trim()
  searchKeyword.value = form.value.detail  // 更新输入框为最终地址

  if (item.location) {
    const lng = item.location.lng
    const lat = item.location.lat
    form.value.lng = lng
    form.value.lat = lat
    if (map) {
      map.setCenter([lng, lat])
      marker && marker.setPosition([lng, lat])
      setTimeout(() => map && map.resize(), 300)
    }
  } else {
    if (placeSearch && name) {
      placeSearch.search(name, (status: string, result: any) => {
        if (status === 'complete' && result?.poiList?.poifs?.length) {
          const p = result.poiList.poifs[0]
          if (p.location) {
            const lng = p.location.lng
            const lat = p.location.lat
            form.value.lng = lng
            form.value.lat = lat
            map.setCenter([lng, lat])
            marker && marker.setPosition([lng, lat])
          }
        }
      })
    }
  }
  suggestions.value = []
}
function updateLocation(lnglat: any) {
  if (!geocoder) {
    console.warn('⚠️ Geocoder 未初始化')
    return
  }

  // 更新经纬度 + Marker
  form.value.lng = lnglat.lng
  form.value.lat = lnglat.lat
  marker.setPosition([lnglat.lng, lnglat.lat])

  // 使用 Geocoder 获取地址
  geocoder.getAddress([lnglat.lng, lnglat.lat], (status: string, result: any) => {
    if (status === 'complete' && result.regeocode) {
      const addr = result.regeocode.formattedAddress
      console.log('逆地理解析成功：', addr)
      form.value.detail = addr // ✅ 自动填充输入框
      // 点击地图后清空搜索框与建议列表，让最终选择更明确
      searchKeyword.value = ''
      suggestions.value = []
    } else {
      console.warn('逆地理解析失败', status, result)
    }
  })
}


// --------- 当前定位逻辑 ----------
function locateCurrent() {
  if (!navigator.geolocation) {
    ElMessage.error('当前浏览器不支持定位功能')
    return
  }

  navigator.geolocation.getCurrentPosition(
    pos => {
      const { latitude, longitude } = pos.coords
      const lnglat = { lng: longitude, lat: latitude }
      map.setCenter([longitude, latitude])
      updateLocation(lnglat)
      ElMessage.success('定位成功，已更新到当前位置')
    },
    err => {
      switch (err.code) {
        case err.PERMISSION_DENIED:
          ElMessage.error('定位权限被拒绝')
          break
        case err.POSITION_UNAVAILABLE:
          ElMessage.error('位置信息不可用')
          break
        case err.TIMEOUT:
          ElMessage.error('定位超时')
          break
        default:
          ElMessage.error('定位失败，请重试')
      }
    },
    { enableHighAccuracy: true, timeout: 5000, maximumAge: 0 }
  )
}

// 加载高德脚本（如未加载），并查询当前位置周边 POI 填充 nearbyAddresses
async function loadNearbyAddresses() {
  console.log('📍 开始执行 loadNearbyAddresses')

  if (!navigator.geolocation) {
    ElMessage.error('浏览器不支持定位，无法获取附近地址')
    return
  }

  navigator.geolocation.getCurrentPosition(async pos => {
    console.log('✅ 成功获取定位', pos.coords)
    const lng = pos.coords.longitude
    const lat = pos.coords.latitude
    
    const amapKey = import.meta.env.VITE_AMAP_KEY || ''
    const AMapScriptUrl = `https://webapi.amap.com/maps?v=2.0&key=${amapKey}`
    console.log('🧭 加载地图脚本:', AMapScriptUrl)

    try {
      if (!(window as any).AMap) {
        await new Promise<void>((resolve, reject) => {
          const s = document.createElement('script')
          s.src = AMapScriptUrl
          s.onload = () => resolve()
          s.onerror = () => reject(new Error('加载高德地图脚本失败'))
          document.head.appendChild(s)
        })
      }
    } catch (e) {
      console.error('❌ 加载地图失败:', e)
      return
    }

    const AMap = (window as any).AMap
    if (!AMap) {
      ElMessage.error('未能初始化高德地图')
      return
    }
    console.log('✅ AMap 初始化成功')

    await new Promise((resolve) => {
      const check = setInterval(() => {
        if (AMap.plugin) {
          clearInterval(check)
          resolve(true)
        }
      }, 100)
    })
    console.log('✅ AMap.plugin 可用')

    try {
      AMap.plugin('AMap.PlaceSearch', () => {
        console.log('✅ PlaceSearch 插件加载成功')
        const ps = new AMap.PlaceSearch({ city: '全国' })
        ps.searchNearBy('', [lng, lat], 2000, (status: string, result: any) => {
          console.log('📍 AMap 返回结果:', status, result)
          if (status === 'complete' && result?.poiList?.pois?.length) {
            const list = result.poiList.pois.map((p: any) => {
              const loc = p.location || p._location || { lng: 0, lat: 0 }
              const dist = distanceMeters(lng, lat, loc.lng, loc.lat)
              return {
                name: p.name || '',
                detail: p.address || (p.district ? `${p.district} ${p.name}` : p.name),
                location: loc,
                distance: dist,
                distanceText: formatDistance(dist),
              }
            })
            // 按距离升序排序
            list.sort((a: any, b: any) => (a.distance || 0) - (b.distance || 0))
            nearbyAddresses.value = list
          } else {
            nearbyAddresses.value = []
          }
        })
      })
    } catch (err: any) {
      console.warn('查询附近地址失败', err)
      nearbyAddresses.value = []
    }
  }, err => {
    console.warn('❌ 定位失败', err)
    ElMessage.error('获取定位失败')
  }, { enableHighAccuracy: true, timeout: 5000 })
}

// 当用户切换标签到 nearby 时，自动加载附近地址
onMounted(() => {
  fetchAddresses()
  // 已移除 route-based 自动打开逻辑（结算页直接内嵌对话）
})

watch(activeTab, (v) => {
  console.log('当前切换 tab：', v)
  if (v === 'nearby') {
    nearbyAddresses.value = []
    loadNearbyAddresses()
  }
})
</script>

<style scoped>
/* 页面标题与新增按钮美化 */
.page-title {
  font-size: 20px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 10px;
}
.title-badge {
  font-size: 12px;
  color: #fff;
  background: linear-gradient(90deg,#ffd54f,#ffb400);
  padding: 4px 8px;
  border-radius: 999px;
  font-weight: 600;
}
.add-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
}
.add-btn .plus {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #fff3e0;
  color: #ff9800;
  font-weight: 700;
}

.dialog-title {
  display: flex;
  align-items: center;
  gap: 12px;
}
.dialog-title-icon { font-size: 18px }
.dialog-title-main { font-weight: 700; font-size: 16px }
.dialog-title-sub { font-size: 12px; color: #909399 }

/* 浮动标签表单样式 */
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
  transition: all 0.3s;
  outline: none;
  color: #2c3e50;
}

.form-item input[readonly] {
  background-color: white;
  cursor: default;
}

.form-item select {
  appearance: none;
  padding-right: 30px;
  cursor: pointer;
  background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23666' d='M6 8L1 3h10z'/%3E%3C/svg%3E") no-repeat right 12px center;
}

.form-item label {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 15px;
  color: #909399;
  pointer-events: none;
  transition: 0.2s ease all;
  background: white;
  padding: 0 4px;
}

.form-item input:focus,
.form-item select:focus {
  border-color: #409EFF;
  box-shadow: 0 0 0 2px rgba(64,158,255,0.2);
}

.form-item input:focus ~ label,
.form-item select:focus ~ label,
.form-item.has-value label,
.form-item label.active {
  top: 0;
  font-size: 12px;
  color: #409EFF;
  transform: translateY(-50%);
}

.form-item input:focus::placeholder {
  color: transparent;
}

/* 地址管理头部样式优化 */
.address-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  justify-content: flex-end;
}

.search-bar {
  width: 280px;
  margin-right: auto; /* 将搜索框推到左边 */
}

/* 地址卡片样式优化 */
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
}

.addr-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 标签样式优化 */
.addr-tag, .default-tag {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
}
.address-page {
  max-width: 700px;
  margin: 0 auto;
  background: #fff;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 3px 12px rgba(0, 0, 0, 0.05);
  font-family: 'Arial', sans-serif;
}
.addr-list, .nearby-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.addr-card {
  border-radius: 10px;
  transition: all 0.2s;
}
.addr-card:hover {
  transform: translateY(-2px);
}

.addr-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.addr-name {
  font-weight: 600;
  font-size: 15px;
}

.addr-detail {
  color: #666;
  margin-top: 6px;
}
.addr-distance {
  color: #999;
  font-size: 12px;
  margin-left: 6px;
}

.addr-phone {
  color: #888;
  font-size: 13px;
  margin-top: 4px;
}

.addr-actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: flex-end;
}

.addr-tag {
  margin-left: 8px;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}
.tag-home { background: #fffbe6; color: #ffb400; }
.tag-work { background: #e6f7ff; color: #409eff; }
.tag-school { background: #f0f9eb; color: #67c23a; }
.tag-near { background: #fdf6ec; color: #e6a23c; }
.default-tag {
  margin-left: 6px;
  padding: 2px 6px;
  background: #ffd54f;
  color: #222;
  border-radius: 6px;
  font-size: 12px;
}

.empty {
  color: #999;
  padding: 24px;
  background: #fafafa;
  text-align: center;
  border-radius: 8px;
}
.dialog-box {
  border-radius: 12px;
}
/* 地图面板与控件样式 */
.map-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  margin-top: 10px;
}

.guide-alert {
  border-radius: 8px;
  margin-bottom: 10px;
}

.guide-steps {
  margin: 8px 0 4px 20px;
  padding: 0;
  color: #666;
  font-size: 13px;
  line-height: 1.6;
}

.guide-steps strong {
  color: #333;
  font-weight: 600;
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

.locate-icon {
  width: 18px;
  height: 18px;
  display: inline-block;
}

.search-panel {
  position: relative;
  width: 100%;
}

.map-search-input {
  .el-input__inner {
    height: 42px;              /* 自定义输入框高度 */
    padding-left: 34px !important; /* 给 prefix 腾出空间 */
    box-sizing: border-box;
  }
  .prefix-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #999;
    margin-left: 4px;
  }

  .el-input__prefix {
    left: 8px !important; /* 图标位置 */
    display: flex;
    align-items: center;
  }
  width: 100%;
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
  transition: all 0.2s ease;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.suggestion-item:hover {
  background: #f5f7fa;
}

.suggestion-content {
  flex: 1;
  min-width: 0;
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

.suggestion-icon {
  color: #909399;
  font-size: 16px;
  margin-left: 12px;
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
</style>
