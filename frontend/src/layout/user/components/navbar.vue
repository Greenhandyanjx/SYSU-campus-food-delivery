<template>
	<header ref="navRef" class="meituan-navbar">
		<div class="left">
			<el-button type="text" class="loc-btn" @click="onLocation">
				<i class="el-icon-location"></i>
				<span class="loc-text">当前定位</span>
				<span class="city">{{ city }}</span>
			</el-button>
		</div>

		<div class="center">
			<div v-if="isOverlapping" class="notice notice-search">
		    <el-input
		      v-model="query"
		      placeholder="搜索店铺 / 美食"
		      clearable
		      class="search-input"
		    >
		      <template #suffix>
		        <el-button class="search-btn" type="warning" round @click="onSearch">
		          <el-icon><Search /></el-icon>
		        </el-button>
		      </template>
		    </el-input>
			</div>
			<div v-else class="notice">
				<i class="iconfont icon-bell"></i>
				今日满30减5元，骑手配送更快！
			</div>
		</div>

		<div class="right">
			<el-dropdown trigger="click" @command="handleCommand">
				<span class="el-dropdown-link user-link">
					<el-avatar size="32" icon="User"/>
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
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
const query = ref('')
const router = useRouter()
const q = ref('')
const city = ref(localStorage.getItem('city') || '校园')
const username = ref(localStorage.getItem('username') || '')
const avatar = ref('/src/assets/login/mini-logo.png')

const isOverlapping = ref(false)
const navRef = ref<HTMLElement | null>(null)

function onSearch() {
	if (!q.value) return
	router.push({ path: '/user/home', query: { q: q.value } })
}
function onLocation() { console.log('定位') }

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

onMounted(() => { checkOverlap(); window.addEventListener('scroll', onScroll, { passive: true }); window.addEventListener('resize', onScroll) })
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
.loc-btn:hover {
  opacity: 0.8;
}
.city {
  color: rgba(27, 27, 27, 0.75);
  margin-left: 6px;
}

/* === 中间区（搜索或公告） === */
.meituan-navbar .center {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
}

.notice {
  position: relative;
  z-index: 1110;
  background: rgba(255, 255, 255, 0.5);
  padding: 6px 16px;
  /* border-radius: 8px; */
  font-size: 14px;
  color: #1a1a1a;
}

/* === 搜索框容器 === */
.notice-search {
  max-width: 1000px;
  width: 80%;
  background-color: #fffef4;
  border-radius: 32px !important;
  border: 2px solid #ffb400;
  box-shadow: 0 3px 8px rgba(250, 173, 20, 0.25);
  padding: 6px 20px;
  transition: 0.25s;
  display: flex;
  align-items: center;
}
.notice-search:hover,
.notice-search:focus-within {
  box-shadow: 0 0 0 3px rgba(255, 213, 79, 0.35);
}

/* === 输入框内部 === */
.notice-search :deep(.el-input__wrapper) {
  background: transparent;
  border: none;
  box-shadow: none;
}
.notice-search :deep(.el-input__inner) {
  font-size: 15px;
  color: #704f00;
  padding-right: 36px;
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

