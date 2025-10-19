<template>
	<div class="user-layout">
		<Navbar />
		<main class="content">
			<router-view />
		</main>
		<!-- 底部 tab（仿美团） -->
		<footer class="bottom-nav">
			<el-row type="flex" justify="space-around" align="middle">
				<el-col :span="6" class="tab" @click="go('/user/home')">首页</el-col>
				<el-col :span="6" class="tab" @click="go('/user/orderlist')">订单</el-col>
				<el-col :span="6" class="tab" @click="go('/user/cart')">购物车</el-col>
				<el-col :span="6" class="tab" @click="go('/user/my')">我的</el-col>
			</el-row>
		</footer>
	</div>
</template>

<script setup lang="ts">
import Navbar from './components/navbar.vue'
import { useRouter } from 'vue-router'

const router = useRouter()
function go(path: string) {
	router.push(path)
}
</script>

<style scoped>
.user-layout { display:flex; flex-direction:column; height:100vh }
.content { flex:1; overflow:auto; padding: 12px }
.bottom-nav { height:56px; border-top:1px solid #eee; background:#fff }
.bottom-nav .tab { text-align:center; line-height:56px; cursor:pointer }
</style>
