<template>
	<div class="user-layout">
		<Navbar />
		<main class="content">
			<router-view />
		</main>
		<!-- 底部 tab（仿美团） -->
		<footer class="bottom-nav">
			<el-row type="flex" justify="space-around" align="middle">
				<el-col :span="6" class="tab" :class="{ active: route.path === '/user/home' }" @click="go('/user/home')">首页</el-col>
				<el-col :span="6" class="tab" :class="{ active: route.path === '/user/orderlist' }" @click="go('/user/orderlist')">订单</el-col>
				<el-col :span="6" class="tab" :class="{ active: route.path === '/user/cart' }" @click="go('/user/cart')">购物车</el-col>
				<el-col :span="6" class="tab" :class="{ active: route.path === '/user/my' }" @click="go('/user/my')">我的</el-col>
			</el-row>
		</footer>
	</div>
</template>

<script setup lang="ts">
import Navbar from './components/navbar.vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
function go(path: string) {
	router.push(path)
}
</script>

<style scoped>
.user-layout { display:flex; flex-direction:column;height:100vh }
.content { flex:1; overflow:auto; padding: 12px }
.bottom-nav { height:56px; border-top:1px solid #eee; background:#fff;width: 60%; margin:0 auto; border-radius: 12px 12px 0 0; box-shadow: 0 -2px 8px rgba(0,0,0,0.1); }
.bottom-nav .tab { text-align:center; line-height:56px; cursor:pointer; transition: all 180ms ease; color: #666 }
.bottom-nav .tab:hover { transform: translateY(-4px); background: linear-gradient(90deg, rgba(255,194,0,0.12), rgba(255,194,0,0.04));color: #d97706 }
.bottom-nav .tab.active { background: rgba(246, 189, 1, 0.705); color: #d97706; font-weight:700; border-radius: 8px }
.bottom-nav .tab { padding: 6px 8px }
</style>
