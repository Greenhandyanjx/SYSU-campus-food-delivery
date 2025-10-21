<template>
	<div class="user-home" style="width: 60%; margin:0 auto ;box-shadow:10px;" >
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

			<!-- 轮播 banner -->
	<div class="banner-container">
    <Carousel :images="images" :interval="5000">
      <template #default="{ index }">
        <div class="banner-text">
          <h2>{{ images[index].title }}</h2>
          <p>{{ images[index].desc }}</p>
          <a :href="images[index].link" class="banner-btn">
            {{ images[index].buttonText }}
          </a>
        </div>
      </template>
    </Carousel>
  </div>

			<!-- 活动卡片 -->
				<section class="activities">
					<div class="activity" v-for="(a, i) in activities" :key="i">
						<div class="act-icon" :style="{ backgroundImage: `linear-gradient(135deg, ${a.gradient[0]}, ${a.gradient[1]})` }">
							<img :src="a.icon" alt="" />
						</div>
						<div class="act-body">
							<div class="title">{{ a.title }}</div>
							<div class="sub">{{ a.sub }}</div>
						</div>
					</div>
				</section>

			<!-- 推荐店铺（瀑布流） -->
			<section class="recommend">
				<h3>为你推荐</h3>
						<div class="masonry">
							<div class="store" v-for="(s, idx) in stores" :key="idx" @click="goToStore(s)">
								<div class="store-banner" :style="{ backgroundImage: `url(${s.img})` }"></div>
								<div class="store-body">
									<div class="row">
										<img class="logo" :src="s.logo" alt="logo" />
										<div class="info">
											<div class="name">{{ s.name }}</div>
											<div class="meta">{{ s.desc }}</div>
											<div class="tags">
												<span class="tag" v-for="(t,i) in s.tags" :key="i">{{ t }}</span>
											</div>
										</div>
									</div>
									<div class="row foot">
										<div class="rating">⭐ {{ s.rating }} • 月售 {{ s.sales }}</div>
										<div class="price">起送 ¥{{ s.minOrder }} • 配送 ¥{{ s.deliveryFee }}</div>
									</div>
								</div>
							</div>
						</div>
			</section>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Carousel from '@/components/Carousel.vue'
import banner1 from '@/assets/banners/banner1.svg'
import banner2 from '@/assets/banners/banner2.svg'
import banner3 from '@/assets/banners/banner3.svg'
import banner4 from '@/assets/banners/banner4.svg'
import banner5 from '@/assets/banners/banner5.svg'

const images = [
  {
    src: banner1,
    title: '立即下单',
    desc: '外卖下单立减 10 元',
    link: '/order',
    buttonText: '去下单'
  },
  {
    src: banner2,
    title: '限时优惠',
    desc: '今日饮品买一送一',
    link: '/drinks',
    buttonText: '查看优惠'
  },
  {
    src: banner3,
    title: '新品上线',
    desc: '尝鲜价仅 9.9 元起',
    link: '/new',
    buttonText: '去尝鲜'
  },
  {
    src: banner4,
    title: '午餐推荐',
    desc: '精选套餐，工作日立减 8 元',
    link: '/lunch',
    buttonText: '立即查看'
  },
  {
    src: banner5,
    title: '夜宵来袭',
    desc: '宵夜加码，买二送一',
    link: '/night',
    buttonText: '抢购夜宵'
  }
];
const router = useRouter()
const query = ref('')

const categories = ref([
  { label: '招牌套餐', icon: '/src/assets/icons/setmeal.svg' },
  { label: '现煮粉面', icon: '/src/assets/icons/noodle.svg' },
  { label: '汉堡炸鸡', icon: '/src/assets/icons/burger.svg' },
  { label: '奶茶咖啡', icon: '/src/assets/icons/milktea.svg' },
  { label: '日式便当', icon: '/src/assets/icons/bento.svg' },
  { label: '烧烤烤肉', icon: '/src/assets/icons/bbq.svg' },
  { label: '水果拼盘', icon: '/src/assets/icons/fruit.svg' },
  { label: '精致甜品', icon: '/src/assets/icons/dessert.svg' },
  { label: '家常快炒', icon: '/src/assets/icons/stirfry.svg' },
  { label: '粥粉面饭', icon: '/src/assets/icons/rice.svg' },
])


const stores = ref([
	{ name: '小张快餐', desc: '30分钟内送达', logo: '/src/assets/noImg.png', img: '/src/assets/noImg.png', sales: 1200, rating: 4.6, minOrder: 20, deliveryFee: 5, tags: ['快餐', '热销'] },
	{ name: '甜甜圈工坊', desc: '甜品畅销', logo: '/src/assets/noImg.png', img: '/src/assets/noImg.png', sales: 800, rating: 4.8, minOrder: 15, deliveryFee: 3, tags: ['甜点', '下午茶'] },
	{ name: '鲜榨饮品', desc: '现榨果汁', logo: '/src/assets/noImg.png', img: '/src/assets/noImg.png', sales: 430, rating: 4.4, minOrder: 10, deliveryFee: 2, tags: ['饮品', '健康'] },
])

// const banners = ref([
// 	{ src: '/src/assets/noImg.png', title: '限时满减', sub: '全场满30减10' },
// 	{ src: '/src/assets/logo.svg', title: '新店开张', sub: '新人立减5元' },
// 	{ src: '/src/assets/noImg.png', title: '暑期特惠', sub: '饮品买一送一' },
// ])

const activities = ref([
	{ title: '新客立减', sub: '满20减5', icon: '/src/assets/icons/activity1.svg', gradient: ['#ff9a9e', '#fad0c4'] },
	{ title: '满减活动', sub: '多买多省', icon: '/src/assets/icons/activity2.svg', gradient: ['#a18cd1', '#fbc2eb'] },
	{ title: '品牌专享', sub: '品质保障', icon: '/src/assets/icons/activity3.svg', gradient: ['#f6d365', '#fda085'] },
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
.store { break-inside: avoid; margin-bottom: 12px }
.masonry { column-count: 1; column-gap: 12px }
.masonry .store { display:inline-block; width:100% }
.activities { display:flex; gap:12px; margin:12px 0 }
.activity { display:flex; gap:8px; align-items:center; background:#fff; padding:8px; border-radius:8px; box-shadow: 0 1px 4px rgba(0,0,0,0.06) }
.activity img { width:48px; height:48px }
@media(max-width:1000px){ .masonry { column-count: 2 } }
@media(max-width:600px){ .masonry { column-count: 1 } .store { width:100% } }
.store-top { display:flex; gap:12px }
.logo { width:64px; height:64px; object-fit:cover }
.info .name { font-weight:600 }
.store-bottom { display:flex; justify-content:space-between; align-items:center; margin-top:8px }
@media(max-width:800px){ .store { width: calc(50% - 8px) } }
</style>
