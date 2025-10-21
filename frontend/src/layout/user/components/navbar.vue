<template>
	<header class="meituan-navbar">
		<div class="left">
			<el-button type="text" class="loc-btn" @click="onLocation">
				<i class="el-icon-location"></i>
				<span class="loc-text">当前定位</span>
				<span class="city">{{ city }}</span>
			</el-button>
		</div>

		<div class="center">
			<div class="notice">
    		<i class="iconfont icon-bell"></i>
    		今日满30减5元，骑手配送更快！
  		</div>
		</div>

		<div class="right">
			<!-- <el-button type="text" icon="el-icon-menu" @click="onMenu" class="icon-btn" /> -->

			<el-dropdown trigger="click" @command="handleCommand">
				<span class="el-dropdown-link user-link">
					<el-avatar size="32" icon="User" />
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const q = ref('')
const city = ref(localStorage.getItem('city') || '校园')
const username = ref(localStorage.getItem('username') || '')
const avatar = ref('/src/assets/login/mini-logo.png')

function onSearch() {
	if (!q.value) return
	router.push({ path: '/user/home', query: { q: q.value } })
}
function onLocation() {
	// TODO: hook into real location picker
	console.log('定位')
}
function onMenu() { router.push('/user/home') }

function handleCommand(command: string) {
	if (command === 'logout') {
		localStorage.removeItem('token')
		localStorage.removeItem('username')
		router.push('/login')
	} else if (command === 'profile') {
		router.push('/user/my')
	}
}
</script>

<style scoped>
.meituan-navbar {
	display: flex;
	align-items: center;
	padding: 10px 20px;
	background: #FFC200; /* user top bar yellow */
	color: #1b1b1b;
	box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
	border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}
.meituan-navbar .left {
	width: 200px;
}
.loc-btn {
	display: flex;
	align-items: center;
	gap: 6px;
	color: #1b1b1b;
}
.loc-text { font-weight: 600 }
.city { color: rgba(27,27,27,0.75); margin-left: 6px }
.meituan-navbar .center { flex: 1; display:flex; justify-content:center }
.search-wrap { width: 60%; min-width: 240px }
.meituan-navbar .right { width: 260px; display:flex; justify-content:flex-end; gap:12px; align-items:center }
.icon-btn { color: rgba(27,27,27,0.9) }
.user-link { display:flex; align-items:center; gap:8px; cursor:pointer }
.username { font-weight:600 }

/* improve button visuals in navbar */
.meituan-navbar .el-button {
	background: rgba(255,255,255,0.9);
	border-radius: 6px;
	padding: 6px 10px;
	border: 1px solid rgba(0,0,0,0.06);
}
.meituan-navbar .el-button:hover { transform: translateY(-1px); transition: transform 180ms ease }

</style>
