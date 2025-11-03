<template>
	<div class="user-home-bg">
	<div class="user-home" style="width: 60%; margin:0 auto ;box-shadow:10px;" >
		<!-- 搜索和横幅 -->
		<header class="hero">
		  <div class="search">
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
		</header>
    <!-- 分类 -->
    <section class="categories">
      <CategoryItem
        v-for="(c, i) in categories"
        :key="i"
        :label="c.label"
        :icon="c.icon"
        :active="activeCategory === c.label"
        @select="() => onCategoryClick(c)"
      />
    </section>

			<!-- 轮播 banner -->
<div class="banner-container">
    <Carousel :images="images" :interval="4000" @banner-click="onCategoryClick(images)" />
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
            <section class="recommend" id="meals">
  			<h3>为你推荐</h3>
        <div class="masonry">
        <div
          class="store"
          v-for="(s, idx) in filteredStores"
          :key="idx"
          @click="goToStore(s)"
        >
  			    <!-- <div class="store-banner" :style="{ backgroundImage: `url(${s.img})` }"></div> -->
  			    <div class="store-body">
  			      <div class="row top">
  			        <!-- 左侧：店铺logo和信息 -->
  			        <div class="left-info">
 			 			          <img class="logo" :src="s.logo" alt="logo" />
  			          <div class="info">
  			            <div class="name">{{ s.name }}</div>
  			            <div class="meta">{{ s.desc }}</div>
  			            <div class="tags">
  			 			             <span class="tag" v-for="(t, i) in s.tags" :key="i">{{ t }}</span>
  			            </div>
  			          </div>
  			        </div>

  			        <!-- 右侧：推荐菜 -->
  			  <div class="right-dishes">
                    <div class="dish" v-for="(d, i) in (s.dishes || []).slice(0,3)" :key="i" @click.stop>
              <div class="dish-img-box">
                <img class="dish-img" src="/src/assets/noImg.png" alt="dish" />
              </div>
              <div class="dish-info">
                <div class="dish-name">{{ d.name }}</div>
                <div class="dish-bottom">
                  <span class="dish-price">¥{{ d.price }}</span>
                  <div class="dish-btns">
                    <el-button
                      size="small"
                      circle
                      type="warning"
                      @click.stop="decDish(s, d)"
                      :disabled="d.count === 0"
                    >
                      <el-icon><Minus /></el-icon>
                    </el-button>
                    <span class="dish-count">{{ d.count || 0 }}</span>
                    <el-button
                      size="small"
                      circle
                      type="warning"
                      @click.stop="addDish(s, d)"
                    >
                      <el-icon><Plus /></el-icon>
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

							<!-- 底部信息 -->
							<div class="row foot">
  			        <div class="rating">⭐ {{ s.rating }} • 月售 {{ s.sales }}</div>
  			        <div class="price">起送 ¥{{ s.minOrder }} • 配送 ¥{{ s.deliveryFee }}</div>
  			      </div>
  			    </div>
  			  </div>
  			</div>
			</section>

	</div>
	</div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { addToCart, removeFromCart } from '@/api/user/store'
import { Search } from '@element-plus/icons-vue'
import Carousel from '@/components/Carousel.vue'
import CategoryItem from '@/components/CategoryItem.vue'
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
    lable:'低价满减',
    link: '/order',
    buttonText: '去下单'
  },
  {
    src: banner2,
    title: '午餐推荐',
    desc: '精选套餐，工作日立减 8 元',
    lable:'午餐推荐',
    link: '/lunch',
    buttonText: '立即查看'
  },
  {
    src: banner3,
    title: '新品上线',
    desc: '尝鲜价仅 9.9 元起',
    lable:'沙拉轻食',
    link: '/new',
    buttonText: '去尝鲜'
  },
  {
    src: banner4,
    title: '限时优惠',
    desc: '今日饮品买一送一',
    lable:'极速配送',
    link: '/drinks',
    buttonText: '查看优惠'
  },
  {
    src: banner5,
    title: '极速配送',
    desc: '下单立减配送费',
    lable:'极速配送',
    link: '/delivery',
    buttonText: '享受优惠'
  }
];
const router = useRouter()
const route = useRoute()
const query = ref('')

const categories = ref([
  { label: '全部', icon: '/src/assets/icons/all.svg', key: 'all', filter: null },
  { label: '招牌套餐', icon: '/src/assets/icons/setmeal.svg', key: 'setmeal', filter: 'setmeal' },
  { label: '现煮粉面', icon: '/src/assets/icons/noodle.svg', key: 'noodle', filter: 'noodle' },
  { label: '汉堡炸鸡', icon: '/src/assets/icons/burger.svg', key: 'burger', filter: 'burger' },
  { label: '奶茶咖啡', icon: '/src/assets/icons/milktea.svg', key: 'milktea', filter: 'milktea' },
  { label: '日式便当', icon: '/src/assets/icons/bento.svg', key: 'bento', filter: 'bento' },
  { label: '烧烤烤肉', icon: '/src/assets/icons/bbq.svg', key: 'bbq', filter: 'bbq' },
  { label: '水果拼盘', icon: '/src/assets/icons/fruit.svg', key: 'fruit', filter: 'fruit' },
  { label: '精致甜品', icon: '/src/assets/icons/dessert.svg', key: 'dessert', filter: 'dessert' },
  { label: '家常快炒', icon: '/src/assets/icons/stirfry.svg', key: 'stirfry', filter: 'stirfry' },
  { label: '粥粉面饭', icon: '/src/assets/icons/rice.svg', key: 'rice', filter: 'rice' },
  //五个额外分类
  { label: '极速配送', icon: '/src/assets/icons/delivery.svg', key: 'fast_delivery', filter: 'fast' },
  { label: '午餐推荐', icon: '/src/assets/icons/lunch.svg', key: 'lunch', filter: 'lunch' },
  { label: '低价满减', icon: '/src/assets/icons/discount.svg', key: 'low_price', filter: 'discount' },
  { label: '沙拉轻食', icon: '/src/assets/icons/salad.svg', key: 'salad', filter: 'salad' },
  { label: '精致下午茶', icon: '/src/assets/icons/afternoon.svg', key: 'afternoon', filter: 'afternoon' },
])

// 计算当前激活的分类：优先使用 route.query.cat，其次如果 q 与某个分类 label 相同也视作激活
const activeCategory = computed(() => {
  const q = (route.query.q || '').toString()
  const cat = (route.query.cat || '').toString()
  if (cat) return cat || '全部'
  if (q) {
    const matched = categories.value.find((c: any) => c.label === q)
    if (matched) return matched.label
  }
  // 默认激活 '全部'
  return '全部'
})


const stores = ref([
  {
    name: '黄焖鸡米饭',
    desc: '经典家常菜，味道鲜香',
    img: '/src/assets/noImg.png',
    logo: '/src/assets/noImg.png',
    tags: ['家常菜', '下饭王'],
    rating: 4.8,
    sales: 320,
    minOrder: 15,
    deliveryFee: 2,
    dishes: [
      { name: '黄焖鸡套餐', price: 18, count: 0, categories: ['招牌套餐'], tags: ['招牌','热销'] },
      { name: '香菇滑鸡饭', price: 16, count: 0, categories: ['家常快炒'], tags: ['家常'] },
      { name: '青椒土豆丝', price: 10, count: 0, categories: ['家常快炒'], tags: [] }
    ]
  },
  {
    name: '茶百道',
    desc: '奶香浓郁，果茶清爽',
    img: '/src/assets/noImg.png',
    logo: '/src/assets/noImg.png',
    tags: ['奶茶', '饮品', '水果茶'],
    rating: 4.9,
    sales: 520,
    minOrder: 12,
    deliveryFee: 1,
    dishes: [
      { name: '乌龙奶茶', price: 12, count: 0, categories: ['奶茶咖啡'], tags: ['清新'] },
      { name: '杨枝甘露', price: 15, count: 0, categories: ['奶茶咖啡'], tags: ['网红'] },
      { name: '百香果绿茶', price: 11, count: 0, categories: ['奶茶咖啡','精致下午茶'], tags: ['水果'] },
      { name: '芝士奶盖', price: 13, count: 0, categories: ['精致下午茶'], tags: ['奶盖'] }
    ]
  }
])

// 根据路由 query 进行过滤
const filteredStores = computed(() => {
  const qv = (route.query.q || '').toString().trim().toLowerCase()
  const cat = (route.query.cat || '').toString()
  if (!qv && (!cat || cat === '全部')) return stores.value
  return stores.value.filter((s: any) => {
    if (cat) {
      if (cat === '全部') return true
      return s.dishes.some((d: any) => {
        // 兼容旧字段 d.category 或 新结构 d.categories (数组)
        if (Array.isArray(d.categories)) return d.categories.includes(cat)
        return d.category === cat
      })
    }
    return s.name.toLowerCase().includes(qv) || s.dishes.some((d: any) => d.name.toLowerCase().includes(qv))
  })
})

const addDish = async (store: any, dish: any) => {
  try {
    // first navigate to store page
    await router.push('/user/store/' + encodeURIComponent(store.name))
    // then call addToCart API
    await addToCart({ storeId: store.id, dishId: dish.id, name: dish.name, price: dish.price, qty: 1 })
    dish.count = (dish.count || 0) + 1
    ElMessage.success('已加入购物车')
  } catch (e: any) {
    ElMessage.error('加入购物车失败：' + (e.message || ''))
  }
}

const decDish = async (store: any, dish: any) => {
  if (!dish.count || dish.count <= 0) return
  try {
    await router.push('/user/store/' + encodeURIComponent(store.name))
    await removeFromCart({ storeId: store.id, dishId: dish.id, qty: 1 })
    dish.count--
    ElMessage.success('已从购物车移除')
  } catch (e: any) {
    ElMessage.error('移除失败：' + (e.message || ''))
  }
}

// const goToStore = (store) => {
//   console.log('进入店铺：', store.name)
// }

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

const scrollToMeals = (smooth = true) => {
  nextTick(() => {
    const el = document.getElementById('meals')
    if (el) el.scrollIntoView({ behavior: smooth ? 'smooth' : 'auto', block: 'start' })
  })
}

// function onBannerClick(item: any) {
//   if (item && item.link) {
//     // 保持现有路由行为（可能会跳转）
//     router.push(item.link).catch(() => {})
//   }
//   scrollToMeals(true)
// }

function onCategoryClick(c: any) {
  const qv = (query.value || '').toString().trim()
  const newQuery: any = {}
  if (qv) newQuery.q = qv
  if (c.label && c.label !== '全部') newQuery.cat = c.label

  router.push({ path: '/user/home', query: newQuery }).then(() => scrollToMeals(true)).catch(() => scrollToMeals(true))
}
</script>

<style scoped>
.user-home-bg {
  width: 100%;
  min-height: 100vh;
  background-color: #fffbe6; /* 主色调：柔和黄 */
  background-image: url('/src/assets/login/img_denglu_bj.jpg');
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  /* Keep a small gap so search is visible initially but can scroll under sticky navbar */
  padding-top: 16px;
  padding-bottom: 60px;
  background-attachment: fixed;
}

/* 中间内容卡片 */
.user-home {
  width: 60%;
  background: rgba(255, 248, 225, 0.95); /* 半透明浅黄 */
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(255, 193, 7, 0.35);
  padding: 24px;
  backdrop-filter: blur(6px);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

/* 鼠标悬浮时略微浮起 */
.user-home:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 28px rgba(255, 193, 7, 0.45);
}
.hero {
	max-width: 1000px; margin: 0 auto;
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 20px 0;
}
.search {
  width: 80%;
}

/* 调整 el-input 外观 */
.search-input :deep(.el-input__wrapper) {
  background-color: #fffef4;
  border-radius: 30px;
  border: 2px solid #faad14;
  box-shadow: 0 2px 6px rgba(250, 173, 20, 0.25);
  padding-right: 0px; /* 给按钮留出空间 */
  height: 46px;
  transition: 0.2s;
}

/* 聚焦效果 */
.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 3px rgba(255, 213, 79, 0.3);
}
/* 内部文字 */
.search-input :deep(.el-input__inner) {
  font-size: 15px;
  color: #8c6d1f;
}
/* 搜索按钮样式（嵌入输入框内部） */
.search-btn {
  position: absolute;
  right: 5px;
  top: 1px;
  height: 38px;
  width: 38px;
  border-radius: 50%;
  background-color: #faad14;
  color: white;
  border: none;
  box-shadow: 0 2px 4px rgba(250, 173, 20, 0.4);
  cursor: pointer;
  transition: 0.2s;
}
.search-btn:hover {
  background-color: #ffd666;
  color: #ad8b00;
}
/* 让 suffix 区域不撑开输入框 */
.search-input :deep(.el-input__suffix) {
  position: relative;
  width: 0;
}
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
.recommend {
  padding: 20px;
  background: #fffbe6;
}
.hero .search {
  margin-top: 8px;
  background: #fffef4;
  padding: 8px;
  border-radius: 10px;
  box-shadow: 0 2px 6px rgba(255, 193, 7, 0.25);
}

/* 分类部分的底色和圆角 */
.categories {
  display: flex;
  justify-content: space-around;
  flex-wrap: wrap;
  background: #fffdf2;
  border-radius: 12px;
  padding: 12px;
  box-shadow: inset 0 1px 4px rgba(255, 193, 7, 0.2);
}

/* 推荐模块标题 */
.recommend h3 {
  color: #d48806;
  font-weight: bold;
  border-left: 5px solid #faad14;
  padding-left: 10px;
  margin-bottom: 14px;
}
.recommend h3 {
  color: #d48806;
  font-weight: bold;
  margin-bottom: 16px;
}

.masonry {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.store {
  background: #fff8e1;
  border-radius: 16px;
  box-shadow: 0 2px 6px rgba(0,0,0,0.08);
  transition: all 0.3s ease;
  overflow: hidden;
}

.store:hover {
  box-shadow: 0 4px 12px rgba(255,193,7,0.4);
  transform: translateY(-2px);
}

.store-banner {
  height: 160px;
  background-size: cover;
  background-position: center;
}

.store-body {
  padding: 16px;
}

.row.top {
  display: flex;
  justify-content: space-between;
  gap: 20px;
}

/* 左侧店铺信息 */
.left-info {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  width: 45%;
}

.logo {
  width: 70px;
  height: 70px;
  border-radius: 12px;
  object-fit: cover;
}

.info .name {
  font-weight: bold;
  font-size: 16px;
  color: #b8860b;
}

.info .meta {
  font-size: 13px;
  color: #666;
  margin-top: 4px;
}

.tags .tag {
  background-color: #ffe58f;
  color: #ad6800;
  border-radius: 8px;
  padding: 2px 8px;
  margin-right: 4px;
  font-size: 12px;
}

/* 右侧推荐菜品 */
.right-dishes {
  display: flex;
  gap: 12px;
  width: 60%;
  flex-wrap: wrap;
}

.dish {
  background: #fffdf0;
  border-radius: 8px;
  padding: 6px;	
  width: 27%;
  text-align: center;
  transition: background 0.2s;
}

.dish:hover {
  background: #fff4cc;
}

.dish-img {
  width: 100%;
  aspect-ratio: 1 / 1;
  object-fit: cover;
  border-radius: 6px;
}

.dish-name {
  font-size: 13px;
  font-weight: 500;
  color: #ad8b00;
  margin-top: 6px;
}

.dish-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 4px;
}

.dish-price {
  color: #E53935;
  font-weight: bold;
  font-size: 13px;
}

.dish-btns {
  display: flex;
  align-items: center;
  gap: 4px;
}

.icon-plus,
.icon-minus {
  font-size: 18px;
  color: #d4b106;
  cursor: pointer;
}

.dish-count {
  width: 18%;
  text-align: center;
  font-size: 13px;
  color: #874d00;
}

/* 底部信息 */
.row.foot {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  border-top: 1px solid #fae09a;
  padding-top: 8px;
  color: #8c6d1f;
  font-size: 13px;
}
.banner-container{
  margin: 20px 0;
}
</style>
