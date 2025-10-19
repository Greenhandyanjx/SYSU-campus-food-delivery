<template>
	<header class="meituan-navbar">
		<div class="left">
			<img src="/src/assets/logo.svg" alt="logo" class="logo" />
		</div>
		<div class="center">
			<el-input v-model="q" placeholder="搜索店铺/商品" clearable @keyup.enter="onSearch">
				<template #prefix>
					<i class="el-icon-search"></i>
				</template>
			</el-input>
		</div>
		<div class="right">
			<el-button type="text" icon="el-icon-location" @click="onLocation">附近</el-button>
			<el-button type="text" icon="el-icon-menu" @click="onMenu">分类</el-button>
		</div>
	</header>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const q = ref('')
const router = useRouter()

function onSearch() {
	if (!q.value) return
	router.push({ path: '/user/home', query: { q: q.value } })
}
function onLocation() { console.log('定位') }
function onMenu() { console.log('分类') }
</script>

<style scoped>
.meituan-navbar { display:flex; align-items:center; padding:8px 12px; background:#ffd400 }
.meituan-navbar .logo { height:36px }
.meituan-navbar .center { flex:1; padding:0 12px }
.meituan-navbar .right { display:flex; gap:8px }
.meituan-navbar .el-input__inner { border-radius:20px }
</style>
