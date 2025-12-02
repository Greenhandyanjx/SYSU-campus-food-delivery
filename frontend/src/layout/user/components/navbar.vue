<template>
	<header ref="navRef" class="meituan-navbar">
		<div class="left">
			<el-button type="text" class="loc-btn" @click="onLocation">
        <img src="@\assets\icons\location.svg" alt="定位" />
        <div class="loc-info">
          <span class="loc-text">当前位置</span>
          <span class="city" :title="city">{{ city }}</span>
        </div>
			</el-button>
		</div>

    <div class="center">
      <!-- 在订单页显示订单搜索框 -->
      <div v-if="isOrderRoute" class="notice notice-search">
        <el-input
          v-model="orderQuery"
          placeholder="搜索订单号/店铺/商品"
          clearable
          class="search-input"
          @keyup.enter="onOrderSearch"
        >
          <template #suffix>
            <el-button class="search-btn" type="warning" round @click="onOrderSearch">
              <el-icon><Search /></el-icon>
            </el-button>
          </template>
        </el-input>
      </div>
      <!-- 非订单页仍维持原有的公告/搜索逻辑 -->
      <div v-else style="width: 100%;">
        <div v-if="isOverlapping" class="notice notice-search" style="margin: 0 auto;">
          <SearchSuggest v-model="query" @search="onSearch" @select="onSelectStore" />
        </div>
        <!-- 公告部分 -->
        <div v-else class="notice notice-promo">
          <i class="iconfont icon-fire"></i>
          <span>校园专享 · 午间特惠：满30减5，骑手极速达 🚴‍♀️</span>
        </div>
      </div>
    </div>

		<div class="right">
			<el-dropdown trigger="click" @command="handleCommand">
				<span class="el-dropdown-link user-link">
          <el-avatar :size="45" :src="avatar || defaultAvatar" />
            <span class="username">{{ username || '游客' }}</span>
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
import { ref, onMounted, onUnmounted, computed } from 'vue'
import * as myApi from '@/api/user/my'
import { useRouter, useRoute } from 'vue-router'
import SearchSuggest from '@/components/SearchSuggest.vue'
import { Search } from '@element-plus/icons-vue'

const query = ref('')
const orderQuery = ref('')

function onSelectStore(s: any) {
  if (!s || !s.name) return
  router.push('/user/store/' + encodeURIComponent(s.name))
}
const router = useRouter()
const route = useRoute()
const q = ref('')
const city = ref(localStorage.getItem('city') || '定位中...')
const username = ref(localStorage.getItem('username') || '')
const avatar = ref(localStorage.getItem('avatar') || '')
const defaultAvatar = '/src/assets/user.png'

// 仅当路由是订单列表或订单详情时，显示订单搜索框
const isOrderRoute = computed(() => {
  try { return route.path.startsWith('/user/orderlist') || route.path.startsWith('/user/order') } catch (e) { return false }
})

const isOverlapping = ref(false)
const navRef = ref<HTMLElement | null>(null)

function onSearch() {
  // 使用 navbar 的 query 输入进行搜索
  // 如果为空则直接跳到首页（显示全部）；否则将查询字符串作为 q 传给首页
  if (!query.value) {
    router.push({ path: '/user/home' })
    return
  }
  router.push({ path: '/user/home', query: { q: query.value } })
}

function onOrderSearch() {
  if (!orderQuery.value) {
    router.push({ path: '/user/orderlist' })
    return
  }
  router.push({ path: '/user/orderlist', query: { oq: orderQuery.value } })
}
// 点击导航栏左侧：先尝试刷新实时定位（异步），然后跳转到地址管理页
async function onLocation() {
  // 尝试刷新一次定位信息（不阻塞太久）
  try {
    await fetchAndSetCurrentAddress(3000)
  } catch (e) {
    // 忽略，仍然路由到地址页
  }
  router.push('/user/address')
}

// 加载高德脚本（如果未加载），返回当 script 加载完成
function ensureAMapLoaded(): Promise<void> {
  const amapKey = (import.meta.env.VITE_AMAP_KEY as string) || ''
  const url = `https://webapi.amap.com/maps?v=2.0&key=${amapKey}`
  return new Promise((resolve, reject) => {
    if ((window as any).AMap) return resolve()
    const s = document.createElement('script')
    s.src = url
    s.onload = () => resolve()
    s.onerror = () => reject(new Error('加载高德脚本失败'))
    document.head.appendChild(s)
  })
}

// 获取浏览器定位并通过高德逆地理解析成可读地址，超时参数 ms（可选）
function fetchAndSetCurrentAddress(timeoutMs = 5000): Promise<void> {
  return new Promise(async (resolve) => {
    if (!navigator.geolocation) {
      city.value = localStorage.getItem('city') || '无法定位'
      return resolve()
    }

    let done = false
    const timer = setTimeout(() => {
      if (!done) {
        done = true
        resolve()
      }
    }, timeoutMs)

    navigator.geolocation.getCurrentPosition(async pos => {
      if (done) return
      try {
        await ensureAMapLoaded()
        const AMap = (window as any).AMap
        if (!AMap) throw new Error('AMap 未初始化')

        const lng = pos.coords.longitude
        const lat = pos.coords.latitude
        console.log('成功获取定位', { lng, lat })

        // 🔹 先尝试获取当前 POI 名称（类似“中山大学南校区”）
        AMap.plugin(['AMap.Geocoder', 'AMap.PlaceSearch'], () => {
          const geocoder = new AMap.Geocoder({ city: '全国' })
          const placeSearch = new AMap.PlaceSearch({ city: '全国' })

          // 搜索附近 100 米的 POI
          placeSearch.searchNearBy('', [lng, lat], 100, (status: string, result: any) => {
            let placeName = ''
            if (status === 'complete' && result?.poiList?.pois?.length) {
              // 取第一个最近的 POI 名称
              const nearest = result.poiList.pois[0]
              placeName = nearest.name || ''
              console.log('附近最近地点:', placeName)
            }

            // 如果没有找到 POI，则回退到逆地理地址
            geocoder.getAddress([lng, lat], (geoStatus: string, geoResult: any) => {
              if (geoStatus === 'complete' && geoResult?.regeocode) {
                const comp = geoResult.regeocode.addressComponent
                const detailParts: string[] = []
                if (comp.district) detailParts.push(comp.district)
                if (comp.township) detailParts.push(comp.township)
                if (comp.street) detailParts.push(comp.street)
                if (comp.streetNumber) detailParts.push(comp.streetNumber)
                if (comp.neighborhood?.name) detailParts.push(comp.neighborhood.name)

                const fallback = detailParts.join('') || geoResult.regeocode.formattedAddress || '未知地址'

                // 最终取：附近地点名 > 逆地理地址
                const finalAddr = placeName || fallback

                city.value = finalAddr
                localStorage.setItem('city', finalAddr)
                console.log('当前地址:', finalAddr)
              } else {
                city.value = placeName || localStorage.getItem('city') || '定位失败'
              }

              clearTimeout(timer)
              done = true
              resolve()
            })
          })
        })
      } catch (err) {
        console.warn('定位解析异常', err)
        clearTimeout(timer)
        done = true
        city.value = localStorage.getItem('city') || '定位失败'
        resolve()
      }
    }, err => {
      console.warn('获取定位失败', err)
      clearTimeout(timer)
      done = true
      city.value = localStorage.getItem('city') || '定位失败'
      resolve()
    }, { enableHighAccuracy: true, timeout: timeoutMs, maximumAge: 0 })
  })
}



function handleCommand(command: string) {
  if (command === 'logout') {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/user/my')
  }
}

function checkOverlap() {
  try {
    const searchEl = document.querySelector('.user-home .search') as HTMLElement | null
    const navEl = navRef.value
    if (!searchEl || !navEl) { isOverlapping.value = false; return }
    const searchRect = searchEl.getBoundingClientRect()
    const navRect = navEl.getBoundingClientRect()
    isOverlapping.value = searchRect.bottom < navRect.bottom
  } catch (e) {
    isOverlapping.value = false
  }
}

let rafId: number | null = null
function onScroll() {
  if (rafId != null) cancelAnimationFrame(rafId)
  rafId = requestAnimationFrame(() => { checkOverlap(); rafId = null })
}

onMounted(() => {
  checkOverlap();
  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('resize', onScroll);
  // 页面加载时获取一次实时地址展示
  fetchAndSetCurrentAddress().catch(() => {})
  // 尝试从后端拉取用户资料（头像/用户名）以显示在导航栏
  myApi.getProfile().then((p: any) => {
    if (!p) return
    username.value = p.nickname || p.username || localStorage.getItem('username') || username.value
    avatar.value = p.avatar_url || p.avatar || avatar.value
    if (avatar.value) localStorage.setItem('avatar', avatar.value)
  }).catch(() => {})
})
onUnmounted(() => { window.removeEventListener('scroll', onScroll); window.removeEventListener('resize', onScroll); if (rafId != null) cancelAnimationFrame(rafId) })
</script>

<style scoped>
/* === 顶部导航栏整体 === */
.meituan-navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px; /* 固定高度 */
  padding: 0 28px;
  background: #ffcc00; /* 主体黄 (#FFC300稍偏暖) */
  color: #1b1b1b;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06);
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  position: sticky;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1100;
  font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
}

/* === 左侧区 === */
.meituan-navbar .left {
  width: 20%;
  border-width: 0cap;
  display: flex;
  align-items: center;
  gap: 8px;
}

.loc-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #1b1b1b;
  font-weight: 600;
  cursor: pointer;
  transition: 0.2s;
}
.loc-btn img{
  width: 20px;
  height: 20px;
}
.loc-btn:hover {
  opacity: 0.8;
}
.city {
  color: rgba(27, 27, 27, 0.85);
  margin-left: 6px;
  display: block;
  max-width: 220px;
  white-space: normal;
  word-break: break-word;
}

.loc-info { display: flex; flex-direction: column; align-items: flex-start; }
.loc-text { font-size: 12px; color: rgba(27,27,27,0.6); }

/* === 中间区（搜索或公告） === */
.meituan-navbar .center {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

/* === 公告样式 notice-promo === */
.notice-promo {
  position: relative;
  z-index: 1110;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 20px;
  border-radius: 999px;
  background: linear-gradient(90deg, #fff8e1, #ffe7b3, #ffd580);
  color: #7a3600;
  font-weight: 600;
  font-size: 20px;
  box-shadow: 0 2px 6px rgba(255, 193, 7, 0.25);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  backdrop-filter: blur(3px);
}

.notice-promo:hover {
  transform: translateY(-1px);
  box-shadow: 0 3px 10px rgba(255, 183, 77, 0.5);
}

/* 图标动画 */
.notice-promo i {
  margin-right: 6px;
  color: #ff7e29;
  animation: flamePulse 2s infinite ease-in-out;
}

@keyframes flamePulse {
  0%, 100% { transform: scale(1); opacity: 0.9; }
  50% { transform: scale(1.2); opacity: 1; }
}

/* 文字发光渐变 */
.notice-promo span {
  background: linear-gradient(90deg, #ff9800, #ff6b00);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: shineText 3s infinite ease-in-out;
}

@keyframes shineText {
  0%, 100% { opacity: 0.9; }
  50% { opacity: 1; filter: drop-shadow(0 0 4px rgba(255, 153, 0, 0.6)); }
}


/* === 搜索框容器 === */
.notice-search {
  max-width: 1000px;
  width: 80%;
  background-color: #fffef4 !important;
  border-radius: 5px !important;
  /* border: 2px solid #fffef4; */
  box-shadow: 0 2px 6px rgba(250, 173, 20, 0.25);
  padding: 8px;
  transition: 0.25s;
  display: flex;
  align-items: center;
}
.notice-search:hover,
.notice-search:focus-within {
  box-shadow: 0 0 0 3px rgba(255, 213, 79, 0.3);
}

/* === 输入框内部 === */
.notice-search :deep(.el-input__wrapper) {
  background-color: #fffef4;
  border-radius: 30px;
  border: 2px solid #faad14;
  box-shadow: 0 2px 6px rgba(250, 173, 20, 0.25);
  padding-right: 0px;
  height: 46px;
}
.notice-search :deep(.el-input__inner) {
  font-size: 15px;
  color: #8c6d1f;
}
.notice-search :deep(.el-input__suffix) {
  position: relative;
  width: 0;
}

/* === 搜索按钮 === */
.notice-search  .search-btn{
  position: absolute;
  /* right: -10px;
  top: -3px; */
  right: 4px;
  top: 2px; 
  height: 38px;
  width: 38px;
  border-radius: 50%;
  background-color: #ffb400;
  color: #fff;
  border: none;
  box-shadow: 0 2px 4px rgba(250, 173, 20, 0.35);
  cursor: pointer;
  transition: 0.25s;
}
.navbar-suggestions {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  width: 100%;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  z-index: 1400;
  max-height: 300px;
  overflow: auto;
  border: 1px solid rgba(0,0,0,0.06);
}
.navbar-suggestions ul{ margin:0; padding:8px 0; list-style:none }
.nav-sugg-item{ padding:8px 12px; cursor:pointer; display:flex; flex-direction:column; gap:6px }
.nav-sugg-item + .nav-sugg-item{ border-top: 1px solid rgba(0,0,0,0.06) }
.nav-sugg-item.active{ background: #fff9e6 }
.nav-sugg-name strong{ background: rgba(255,235,59,0.5); padding:0 2px }
.nav-sugg-desc{ font-size:12px; color:#888 }
.notice-search .search-btn:hover {
  background-color: #ffd34e;
  color: #744d00;
  transform: scale(1.05);
}

/* === 右侧按钮区 === */
.meituan-navbar .right {
  width: 20%;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 14px;
}

.icon-btn {
  color: rgba(27, 27, 27, 0.9);
  cursor: pointer;
  transition: 0.2s;
}
.icon-btn:hover {
  transform: translateY(-1px);
  opacity: 0.8;
}

.user-link {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.username {
  font-weight: 600;
  color: #1b1b1b;
}

/* === 右侧 element 按钮优化 === */
.meituan-navbar :deep(.right.el-button) {
  background: #fff;
  color: #1b1b1b;
  border-radius: 8px;
  padding: 6px 12px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  transition: 0.2s;
}
.meituan-navbar :deep(.el-button:hover) {
  background: #ffefb3;
  transform: translateY(-1px);
}

/* === 响应式微调 === */
@media (max-width: 900px) {
  .notice-search {
    width: 80%;
  }
  .meituan-navbar .right {
    display: none;
  }
}
</style>

