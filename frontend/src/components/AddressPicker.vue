<template>
  <div v-if="innerVisible" class="ap-overlay">
    <div class="ap-modal">
      <div class="ap-header">选择地址 <button class="ap-close" @click="close">✕</button></div>
      <div class="picker-grid">
        <div class="map-wrap">
            <div ref="mapEl" class="map-canvas"></div>
            <button class="ap-locate-btn" @click="recenterOriginal" title="回到起始位置">
              <img src="\src\assets\icons\address.svg" style="width: 20px;height: 20px;color: blue;" alt="">
            </button>
        </div>

        <div class="controls">
          <el-form label-position="top">
            <el-form-item label="地址（格式化）">
              <el-input
                v-model="formatted"
                type="textarea"
                :autosize="{ minRows: 1, maxRows: 4 }"
                placeholder="点击地图或拖动标记获取地址"
                @input="onFormattedInput"
              />

              <div v-if="suggestions.length" class="suggestion-box ap-suggestion-box">
                <div
                  v-for="(s, i) in suggestions"
                  :key="i"
                  class="suggestion-item"
                  @click="selectSuggestion(s)"
                >
                  <div class="suggestion-content">
                    <div class="suggestion-name">{{ s.name || s.keyword || s.title }}</div>
                    <div class="suggestion-address">{{ formatTipAddress(s) }}</div>
                  </div>
                  <div class="suggestion-icon">›</div>
                </div>
              </div>
            </el-form-item>

            <el-form-item label="详址（门牌 / 楼层 / 房号）">
              <el-input
                v-model="detail"
                placeholder="填写门牌等详细信息"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="confirm">确认</el-button>
              <el-button @click="close">取消</el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import amapLoader from '@/utils/amap'

/* ===================== props / emits ===================== */

const props = defineProps<{
  modelValue?: {
    formatted?: string
    detail?: string
    lng?: number
    lat?: number
  }
  visible: boolean
}>()

const emit = defineEmits([
  'update:modelValue',
  'update:visible',
  'close'
])

/* ===================== dialog visible ===================== */

const innerVisible = ref(props.visible)

watch(
  () => props.visible,
  v => (innerVisible.value = v)
)

watch(
  innerVisible,
  async v => {
    console.log('AddressPicker innerVisible ->', v)
    emit('update:visible', v)
    if (v) {
      await nextTick()
      initMap()
    } else {
      onClosed()
    }
  }
)

/* ===================== address state ===================== */

const formatted = ref(props.modelValue?.formatted || '')
const detail = ref(props.modelValue?.detail || '')
const lng = ref(props.modelValue?.lng || 113.582)
const lat = ref(props.modelValue?.lat || 22.352)

/* ===================== AMap objects ===================== */

const mapEl = ref<HTMLDivElement | null>(null)

let map: any = null
let marker: any = null
let geocoder: any = null

let autoComplete: any = null
let placeSearch: any = null
let _keywordTimer: any = null
const suggestions = ref<any[]>([])
let initialLng: number | null = null
let initialLat: number | null = null

/* ===================== init map ===================== */

async function initMap() {
  await nextTick()

  console.log('AddressPicker initMap, mapEl=', mapEl.value)

  if (!mapEl.value) {
    console.error('AddressPicker: map container not found')
    return
  }

  // 使用统一的 amapLoader 加载地图
  try {
    const AMap = await amapLoader.load({
      plugins: ['AMap.Geocoder', 'AMap.PlaceSearch', 'AMap.ToolBar', 'AMap.AutoComplete']
    });

    if (!AMap) {
      console.error('AddressPicker: AMap not available after load')
      return
    }

    // Initialize plugins
    AMap.plugin(['AMap.Geocoder', 'AMap.AutoComplete', 'AMap.PlaceSearch'], () => {
      geocoder = new AMap.Geocoder({ city: '全国' })
      try { autoComplete = new AMap.AutoComplete({ city: '全国' }) } catch (e) { autoComplete = null }
      try { placeSearch = new AMap.PlaceSearch({ city: '全国' }) } catch (e) { placeSearch = null }
    })

    map = new AMap.Map(mapEl.value, {
      zoom: 15,
      center: [lng.value, lat.value]
    })

    marker = new AMap.Marker({
      position: [lng.value, lat.value],
      draggable: true,
      map
    })

    map.on('click', (e: any) => updateLocation(e.lnglat))
    marker.on('dragend', (e: any) => updateLocation(e.lnglat))

    // 初始点位反向解析
    reverseGeocode(lng.value, lat.value)

    // store initial center so user can return to it
    initialLng = lng.value
    initialLat = lat.value

    // Ensure proper rendering
    setTimeout(() => map && map.resize(), 300)
  } catch (e) {
    console.error('AddressPicker: failed to load AMap script', e)
    return
  }

function recenterOriginal() {
  if (initialLng == null || initialLat == null) return
  lng.value = initialLng
  lat.value = initialLat
  if (marker) marker.setPosition([lng.value, lat.value])
  if (map) map.setCenter([lng.value, lat.value])
  reverseGeocode(lng.value, lat.value)
  setTimeout(() => map && map.resize(), 150)
}

// 输入联想（去抖）
function onFormattedInput(val: string) {
  if (!autoComplete) {
    suggestions.value = []
    return
  }
  if (_keywordTimer) clearTimeout(_keywordTimer)
  if (!val) {
    suggestions.value = []
    return
  }
  _keywordTimer = setTimeout(() => {
    try {
      autoComplete.search(val, (status: string, result: any) => {
        if (status === 'complete' && result?.tips) {
          suggestions.value = result.tips.filter((t: any) => !!t.name || !!t.location)
        } else {
          suggestions.value = []
        }
      })
    } catch (e) {
      suggestions.value = []
    }
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
  formatted.value = (item.name || item.keyword || item.title || '') + (item.district ? ' ' + item.district : '') + (item.address ? ' ' + item.address : '')
  suggestions.value = []
  const loc = item.location
  if (loc) {
    const lngv = loc.lng || loc[0]
    const latv = loc.lat || loc[1]
    if (lngv && latv) {
      lng.value = lngv
      lat.value = latv
      marker && marker.setPosition([lng.value, lat.value])
      map && map.setCenter([lng.value, lat.value])
      reverseGeocode(lng.value, lat.value)
      setTimeout(() => map && map.resize(), 200)
      return
    }
  }
  if (placeSearch && (item.name || item.keyword)) {
    placeSearch.search(item.name || item.keyword, (status: string, result: any) => {
      if (status === 'complete' && result?.poiList?.poifs?.length) {
        const p = result.poiList.poifs[0]
        if (p.location) {
          const parts = (p.location as string).split(',')
          const lngv = parseFloat(parts[0])
          const latv = parseFloat(parts[1])
          lng.value = lngv
          lat.value = latv
          marker && marker.setPosition([lng.value, lat.value])
          map && map.setCenter([lng.value, lat.value])
          reverseGeocode(lng.value, lat.value)
        }
      }
    })
  }
}

/* ===================== helpers ===================== */

function updateLocation(lnglat: any) {
  lng.value = lnglat.lng
  lat.value = lnglat.lat

  marker?.setPosition([lng.value, lat.value])
  reverseGeocode(lng.value, lat.value)
}

function reverseGeocode(lng: number, lat: number) {
  if (!geocoder) return

  geocoder.getAddress([lng, lat], (status: string, result: any) => {
    if (status === 'complete' && result?.regeocode) {
      formatted.value = result.regeocode.formattedAddress || ''
    }
  })
}

/* ===================== actions ===================== */

function confirm() {
  emit('update:modelValue', {
    formatted: formatted.value,
    detail: detail.value,
    lng: lng.value,
    lat: lat.value
  })
  innerVisible.value = false
}

function close() {
  innerVisible.value = false
}

function onClosed() {
  destroyMap()
  emit('close')
}

/* ===================== cleanup ===================== */

function destroyMap() {
  if (map) {
    map.destroy()
    map = null
    marker = null
    geocoder = null
  }
}
</script>

<style scoped>
.picker-grid {
  display: flex;
  gap: 16px;
}

.map-wrap {
  flex: 1;
}

.map-canvas {
  width: 100%;
  height: 420px;
  border-radius: 8px;
  background: #f0f0f0;
}

.controls {
  width: 320px;
}

.map-wrap { position: relative }

.ap-locate-btn {
  position: absolute;
  right: 12px;
  bottom: 12px;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: none;
  background: #fff;
  box-shadow: 0 6px 18px rgba(0,0,0,0.18);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 16px;
  z-index: 2200;
}

.ap-locate-btn:hover { transform: translateY(-2px) }

/* simple overlay/modal styles */
.ap-overlay {
  position: fixed;
  left: 0;
  top: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}
.ap-modal {
  position:fixed;
  opacity: 1;
  background: #fff;
  border-radius: 8px;
  width: 920px;
  max-width: calc(100% - 32px);
  padding: 16px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.2);
}
.ap-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  margin-bottom: 8px;
}
.ap-close {
  background: transparent;
  border: none;
  font-size: 16px;
  cursor: pointer;
}

/* suggestion dropdown inside picker */
.ap-suggestion-box {
  position: absolute;
  left: 0;
  right: 0;
  top: 46px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 6px 18px rgba(0,0,0,0.12);
  max-height: 260px;
  overflow: auto;
  z-index: 2100;
}
.ap-suggestion-box .suggestion-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
}
.ap-suggestion-box .suggestion-item:hover { background: #f5f7fa }
.ap-suggestion-box .suggestion-name { font-weight: 600; color: #222 }
.ap-suggestion-box .suggestion-address { font-size: 12px; color: #666; margin-top: 4px }
.ap-suggestion-box .suggestion-content { display: flex; flex-direction: column; min-width: 0 }
.ap-suggestion-box .suggestion-icon { color: #909399; margin-left: 12px }
</style>