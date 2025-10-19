<template>
	<div class="user-home">
		<!-- 搜索和横幅 -->
		<header class="hero">
			<div class="search">
				<el-input v-model="query" placeholder="搜索店铺/美食" clearable @keyup.enter="onSearch">
					<template #append>
						<el-button type="primary" @click="onSearch">搜索</el-button>
					</template>
				</el-input>
			</div>
		</header>

		<!-- 分类 -->
		<section class="categories">
			<div class="cat" v-for="(c, i) in categories" :key="i">
				<img :src="c.icon" alt="" />
				<div class="label">{{ c.label }}</div>
			</div>
		</section>

		<!-- 推荐店铺 -->
		<section class="recommend">
			<h3>为你推荐</h3>
			<div class="cards">
				<el-card class="store" v-for="(s, idx) in stores" :key="idx">
					<div class="store-top">
						<img class="logo" :src="s.logo" alt="logo" />
						<div class="info">
							<div class="name">{{ s.name }}</div>
							<div class="meta">{{ s.desc }}</div>
						</div>
					</div>
					<div class="store-bottom">
						<div class="score">销量：{{ s.sales }}</div>
						<el-button type="primary" size="small" @click="goToStore(s)">进店</el-button>
					</div>
				</el-card>
			</div>
		</section>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const query = ref('')

const categories = ref([
	{ label: '午餐', icon: '/src/assets/icons/meal.svg' },
	{ label: '甜点', icon: '/src/assets/icons/dessert.svg' },
	{ label: '饮品', icon: '/src/assets/icons/drink.svg' },
	{ label: '沙拉', icon: '/src/assets/icons/salad.svg' },
])

const stores = ref([
	{ name: '小张快餐', desc: '30分钟内送达', logo: '/src/assets/noImg.png', sales: 1200 },
	{ name: '甜甜圈工坊', desc: '甜品畅销', logo: '/src/assets/noImg.png', sales: 800 },
	{ name: '鲜榨饮品', desc: '现榨果汁', logo: '/src/assets/noImg.png', sales: 430 },
])

function onSearch() {
	// TODO: hook up search
	if (!query.value) return
	router.push({ path: '/user/home', query: { q: query.value } })
}

function goToStore(s: any) {
	// 进入店铺详情页（占位）
	router.push('/user/store/' + encodeURIComponent(s.name))
}
</script>

<style scoped>
.user-home { padding: 12px; }
.hero { margin-bottom: 12px; }
.search { max-width: 900px; margin: 0 auto; }
.categories { display:flex; gap:16px; padding: 12px 0; overflow:auto }
.cat { width:72px; text-align:center }
.cat img { width:48px; height:48px }
.recommend { margin-top: 20px }
.cards { display:flex; gap:12px; flex-wrap:wrap }
.store { width: calc(33.333% - 8px) }
.store-top { display:flex; gap:12px }
.logo { width:64px; height:64px; object-fit:cover }
.info .name { font-weight:600 }
.store-bottom { display:flex; justify-content:space-between; align-items:center; margin-top:8px }
@media(max-width:800px){ .store { width: calc(50% - 8px) } }
</style>
