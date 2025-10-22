<template>
	<div class="user-layout">
		<Navbar />
		<main class="content">
			<router-view />
		</main>
			<!-- 底部 tab（仿美团） -->
			<footer class="bottom-nav fixed-bottom">
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
.user-layout { display:flex; flex-direction:column; min-height:100vh }
.content { flex:1; overflow: visible; padding: 0px }
.page-wrap { display:flex; justify-content:center; position:relative }
.page-wrap::before,
.page-wrap::after {
	content: '';
	width: calc((100% - 1100px)/2);
	background-image: url('/src/assets/login/img_denglu_bj.jpg');
	background-size: cover;
	background-position: center;
}
.page-wrap::before { margin-right: 0 }
.page-wrap::after { margin-left: 0 }
.page-main { width: 1100px; max-width: calc(100% - 40px); background: #fff; border-radius: 8px; box-shadow: 0 12px 30px rgba(0,0,0,0.08); padding: 20px; margin-top: 12px }
.bottom-nav { height:56px;  background:transparent; width: 100%; }
.bottom-nav.fixed-bottom { position: fixed; left: 0; right: 0; bottom: 0; display:flex; justify-content:center; pointer-events: auto; z-index: 1200 }
.bottom-nav.fixed-bottom .el-row { width: 60%; background: rgba(255,255,255,0.95); border-radius: 12px 12px 0 0; box-shadow: 0 -2px 8px rgba(0,0,0,0.1); }
.bottom-nav .tab { text-align:center; line-height:56px; cursor:pointer; transition: all 180ms ease; color: #666 }
.bottom-nav .tab:hover { transform: translateY(-4px); background: linear-gradient(90deg, rgba(255,194,0,0.12), rgba(255,194,0,0.04));color: #d97706 }
.bottom-nav .tab.active { background: rgba(246, 189, 1, 0.705); color: #d97706; font-weight:700; border-radius: 8px }
.bottom-nav .tab { padding: 6px 8px }

/* ensure content doesn't create its own scrollbar inside page-main */
.page-main { overflow: visible }

</style>
