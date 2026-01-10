<template>
	<div class="user-home-bg">
	<div class="user-home" style="width: 60%; margin:0 auto ;box-shadow:10px;" >
		<!-- 搜索和横幅 -->
		<header class="hero">
      <div class="search">
          <SearchSuggest v-model="query" @search="onSearch" @select="(s) => goToStore(s)" />
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
				<!-- <section class="activities">
					<div class="activity" v-for="(a, i) in activities" :key="i">
						<div class="act-icon" :style="{ backgroundImage: `linear-gradient(135deg, ${a.gradient[0]}, ${a.gradient[1]})` }">
							<img :src="a.icon" alt="" />
						</div>
						<div class="act-body">
							<div class="title">{{ a.title }}</div>
							<div class="sub">{{ a.sub }}</div>
						</div>
					</div>
				</section> -->

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
      <!-- 回到顶部按钮（固定在主内容右侧） -->
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
                <img
                    class="dish-img"
                    :src="d.image?d.image:noImg"
                    alt="dish"
                  />
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
  			        <div class="rating">⭐ {{ s.rating }} • 月售 {{ formatSales(s.sales) }}</div>
  			        <div class="price">起送 ¥{{ s.minOrder }} • 配送 ¥{{ s.deliveryFee }}</div>
  			      </div>
  			    </div>
  			  </div>
  			</div>
			</section>

  </div>
  <!-- 回到顶部按钮（固定在主内容右侧） -->
  <button v-show="showBack" class="back-to-top" @click="scrollToTop" aria-label="回到顶部">
  <img src="@/assets/icons/top-arrow.svg" alt="回到顶部" class="back-icon" />
  </button> 
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { addToCart, removeFromCart } from '@/api/user/store'
import { getCart } from '@/api/user/cart'
import Carousel from '@/components/Carousel.vue'
import SearchSuggest from '@/components/SearchSuggest.vue'
import { Search } from '@element-plus/icons-vue'
import CategoryItem from '@/components/CategoryItem.vue'
import banner1 from '@/assets/banners/images/banner1.png'
import banner2 from '@/assets/banners/images/banner2.png'
import banner3 from '@/assets/banners/images/banner3.png'
import banner4 from '@/assets/banners/images/banner4.png'
import banner5 from '@/assets/banners/images/banner5.png'
import axios from 'axios'
import { getDeliveryConfig } from '@/api/user/store'
import noImg from '@/assets/noImg.png'
import iconAll from '@/assets/icons/all.svg'
import iconSetmeal from '@/assets/icons/setmeal.svg'
import iconNoodle from '@/assets/icons/noodle.svg'
import iconBurger from '@/assets/icons/burger.svg'
import iconMilktea from '@/assets/icons/milktea.svg'
import iconBento from '@/assets/icons/bento.svg'
import iconBbq from '@/assets/icons/bbq.svg'
import iconFruit from '@/assets/icons/fruit.svg'
import iconDessert from '@/assets/icons/dessert.svg'
import iconStirfry from '@/assets/icons/stirfry.svg'
import iconRice from '@/assets/icons/rice.svg'
import iconDelivery from '@/assets/icons/delivery.svg'
import iconLunch from '@/assets/icons/lunch.svg'
import iconDiscount from '@/assets/icons/discount.svg'
import iconSalad from '@/assets/icons/salad.svg'
import iconAfternoon from '@/assets/icons/afternoon.svg'
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

const categories = ref([])
const formatSales = (sales) => {
  const ranges = ['10+', '100+', '1000+']
  return ranges[Math.floor(Math.random() * ranges.length)]
}

// 本地图标映射（以 id 为 key），用于将后端返回的分类 id 映射到前端图标
const iconMap: Record<number, string> = {
  0: iconAll,
  1: iconSetmeal,
  2: iconNoodle,
  3: iconBurger,
  4: iconMilktea,
  5: iconBento,
  6: iconBbq,
  7: iconFruit,
  8: iconDessert,
  9: iconStirfry,
  10: iconRice,
  11: iconDelivery,
  12: iconLunch,
  13: iconDiscount,
  14: iconSalad,
  15: iconAfternoon,
}

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


const stores = ref([])

// 辅助：从后端 categories 列表构建 id->label 映射
function buildCategoryMap(list: any[]) {
  const m: Record<number, string> = {}
  list.forEach((c: any) => {
    if (c && c.id != null) m[c.id] = c.name || c.Name || c.label || ''
  })
  return m
}

// 本地后备分类（回退用）
const fallbackCategories = [
  { id: 0, label: '全部', icon: iconMap[0], key: 'all', filter: null },
  { id: 1, label: '招牌套餐', icon: iconMap[1], key: 'setmeal' },
  { id: 2, label: '现煮粉面', icon: iconMap[2], key: 'noodle' },
  { id: 3, label: '汉堡炸鸡', icon: iconMap[3], key: 'burger' },
  { id: 4, label: '奶茶咖啡', icon: iconMap[4], key: 'milktea' },
  { id: 5, label: '日式便当', icon: iconMap[5], key: 'bento' },
  { id: 6, label: '烧烤烤肉', icon: iconMap[6], key: 'bbq' },
  { id: 7, label: '水果拼盘', icon: iconMap[7], key: 'fruit' },
  { id: 8, label: '精致甜品', icon: iconMap[8], key: 'dessert' },
  { id: 9, label: '家常快炒', icon: iconMap[9], key: 'stirfry' },
  { id: 10, label: '粥粉面饭', icon: iconMap[10], key: 'rice' },
  { id: 11, label: '极速配送', icon: iconMap[11], key: 'fast_delivery' },
  { id: 12, label: '午餐推荐', icon: iconMap[12], key: 'lunch' },
  { id: 13, label: '低价满减', icon: iconMap[13], key: 'low_price' },
  { id: 14, label: '沙拉轻食', icon: iconMap[14], key: 'salad' },
  { id: 15, label: '精致下午茶', icon: iconMap[15], key: 'afternoon' },
]

async function loadData() {
  let backendCats: any[] = []
  try {
    // 1) 尝试获取后端分类（type=1 表示菜品分类）
    const catRes = await axios.get('/api/merchant/category/list', { params: { type: 1 } })
    if (catRes.data && (catRes.data.code === 1 || catRes.data.code === '1')) {
      backendCats = catRes.data.data || []
    }
  } catch (e: any) {
    console.warn('category API failed, will use fallback categories', e && e.message)
  }

  // 如果后端没有返回完整的 15 个分类，使用本地回退列表
  if (!backendCats || backendCats.length < 15) {
    categories.value = fallbackCategories
    backendCats = fallbackCategories.filter((c: any) => c.id && c.id !== 0).map((c: any) => ({ id: c.id, name: c.label }))
  } else {
    const cats: any[] = [{ id: 0, label: '全部', icon: iconMap[0], key: 'all', filter: null }]
    backendCats.forEach((c: any) => {
      const id = Number(c.id)
      cats.push({ id, label: c.name || c.Name, icon: iconMap[id] || iconMap[0], key: c.name || `cat_${id}`, filter: null })
    })
    categories.value = cats
  }

  // 2) 获取后端店铺列表（若失败则保持空数组）
  try {
    const storesRes = await axios.get('/api/user/stores')
    if (storesRes.data && (storesRes.data.code === 1 || storesRes.data.code === '1')) {
      const catMap = buildCategoryMap(backendCats)
      stores.value = (storesRes.data.data || []).map((s: any) => {
        const sd = s.dishes || []
        const dishes = sd.map((d: any) => ({
          id: d.id,
          name: d.name,
          price: d.price,
          count: 0,
          image: d.image || noImg,
          description: d.description || '',
          categories: Array.isArray(d.categoryId) ? d.categoryId.map((cid: number) => catMap[cid] || cid) : [(catMap[d.categoryId] || d.categoryId)],
        }))
        return {
          id: s.id,
          name: s.name || s.shop_name,
          desc: s.desc || s.shop_location,
          img: s.logo || noImg,
          logo: s.logo || noImg,
          tags: s.tags || [],
          rating: s.rating ?? s.avg_score ?? s.avgScore ?? (s.merchant && (s.merchant.avg_score ?? s.merchant.AvgScore ?? s.merchant.avgScore)) ?? 4.8,
          sales: s.sales || 0,
          minOrder: s.minOrder || 0,
          deliveryFee: s.deliveryFee || 0,
          dishes,
        }
      })
      // 补充每个商家的配送配置（避免首页显示为 0）
      try {
        await Promise.all(stores.value.map(async (st: any) => {
          const bid = st.id || st.base_id || st.baseId
          if (!bid) return
          const r = await getDeliveryConfig(bid)
          const cfg = r && r.data ? r.data.data || r.data : r
          st.minOrder = cfg?.min_price ?? cfg?.minPrice ?? st.minOrder ?? 15
          st.deliveryFee = cfg?.delivery_fee ?? cfg?.deliveryFee ?? st.deliveryFee ?? 2
        }))
      } catch (e) { console.warn('fetch delivery configs for stores failed', e) }
      // 同步用户购物车到首页，用于展示菜品已加入数量
      try {
        const cartData = await getCart()
        const shops = cartData && (cartData.shops || cartData.data || cartData) || []
        // 重置计数
        stores.value.forEach((st: any) => st.dishes.forEach((d: any) => (d.count = 0)))
        for (const sh of (shops || [])) {
          const sid = sh.storeId || sh.store_id || sh.merchantId || sh.merchant_id || sh.id || sh.base_id
          const storeMatch = stores.value.find((st: any) => String(st.id) === String(sid) || String(st.base_id) === String(sid) || String(st.id) === String(sh.storeId))
          if (!storeMatch) continue
          const items = sh.items || []
          for (const it of items) {
            const did = it.dishId || it.dish_id || it.id
            const dish = storeMatch.dishes.find((x: any) => String(x.id) === String(did))
            if (dish) dish.count = Number(it.qty || it.Qty || it.qty || 0)
          }
        }
      } catch (e) { console.warn('sync cart to homepage failed', e) }
    } else {
      // 如果返回结构不正确，保持空 stores（不会抛）
      stores.value = []
    }
  } catch (e: any) {
    console.warn('stores API failed, frontend will show no stores until DB has data', e && e.message)
    stores.value = []
  }
}

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
    // First call addToCart so backend/cart is up-to-date, then navigate to store page.
    await addToCart({ storeId: store.id, dishId: dish.id, name: dish.name, price: dish.price, qty: 1 })
    // optimistic update local UI count
    dish.count = (dish.count || 0) + 1
    ElMessage.success('已加入购物车')
    // navigate using store id/base_id when available
    const sid = store.id || store.base_id || store.baseId || store.shop_id || store.shopId
    if (sid) await router.push('/user/store/' + encodeURIComponent(String(sid)))
    else await router.push('/user/store/' + encodeURIComponent(store.name))
  } catch (e: any) {
    ElMessage.error('加入购物车失败：' + (e.message || ''))
  }
}

const decDish = async (store: any, dish: any) => {
  if (!dish.count || dish.count <= 0) return
  try {
    // call remove API first, update local count, then navigate
    await removeFromCart({ storeId: store.id, dishId: dish.id, qty: 1 })
    dish.count = Math.max(0, (dish.count || 0) - 1)
    ElMessage.success('已从购物车移除')
    const sid = store.id || store.base_id || store.baseId || store.shop_id || store.shopId
    if (sid) await router.push('/user/store/' + encodeURIComponent(String(sid)))
    else await router.push('/user/store/' + encodeURIComponent(store.name))
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
    .then(() => scrollToMeals(true))
    .catch(() => scrollToMeals(true))
}

function goToStore(s: any) {
  // Prefer numeric id/base_id when available to avoid name-based lookups
  const sid = s.id || s.base_id || s.baseId || s.shop_id || s.shopId
  if (sid) router.push('/user/store/' + encodeURIComponent(String(sid)))
  else router.push('/user/store/' + encodeURIComponent(s.name || s.shop_name || ''))
}

const scrollToMeals = (smooth = true) => {
  nextTick(() => {
    const el = document.getElementById('meals')
    if (el) el.scrollIntoView({ behavior: smooth ? 'smooth' : 'auto', block: 'start' })
  })
}



const showBack = ref(false)
function scrollHandler() {
  showBack.value = window.scrollY > 240
}
function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}
onMounted(() => {
  window.addEventListener('scroll', scrollHandler, { passive: true })
  loadData()
})
onBeforeUnmount(() => window.removeEventListener('scroll', scrollHandler))

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
  position: relative; /* 让下拉可以绝对定位 */
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

/* 回到顶部按钮样式 */
.back-to-top {
  position: fixed;
  right: calc(20% - 36px);
  bottom: 120px;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  border: none;
  background-color: #fff; /* 改成白色背景 */
  color: #333;
  font-size: 22px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.3s ease;
  z-index: 1200;
  position: fixed;
}

/* hover 效果：轻微上浮 + 黄色描边 */
.back-to-top:hover {
  transform: translateY(-4px);
  border: 2px solid #ffd100;
}

/* 提示气泡（默认隐藏） */
.back-to-top::before {
  content: "回到顶部";
  position: absolute;
  right: 110%;
  top: 50%;
  transform: translateY(-50%);
  white-space: nowrap;
  background: rgba(51, 51, 51, 0.9);
  color: #fff;
  font-size: 13px;
  padding: 6px 10px;
  border-radius: 6px;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.25s ease, transform 0.25s ease;
}

/* 提示框出现时 */
.back-to-top:hover::before {
  opacity: 1;
  transform: translateY(-50%) translateX(-4px);
}

/* 提示气泡的小三角形 */
.back-to-top::after {
  content: "";
  position: absolute;
  right: calc(110% - 6px);
  top: 50%;
  transform: translateY(-50%);
  border-width: 6px;
  border-style: solid;
  border-color: transparent transparent transparent rgba(51, 51, 51, 0.9);
  opacity: 0;
  transition: opacity 0.25s ease;
}

/* 悬停时显示小三角 */
.back-to-top:hover::after {
  opacity: 1;
}
.back-icon {
  width: 40px;
  height: 40px;
  pointer-events: none;
}

/* 搜索建议下拉样式 */
.search-suggestions {
  position: absolute;
  left: 0;
  top: calc(100% + 8px);
  width: 100%;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  z-index: 1300;
  max-height: 320px;
  overflow: auto;
  border: 1px solid rgba(0,0,0,0.06);
}
.search-suggestions ul { list-style: none; margin: 0; padding: 8px 0; }
.sugg-item { padding: 10px 14px; cursor: pointer; display:flex; flex-direction:column; gap:4px }
.sugg-item:hover { background: #fff9e6 }
.sugg-name strong { background: rgba(255, 235, 59, 0.5); padding: 0 2px; }
.sugg-desc { font-size: 12px; color: #888 }

/* 响应式：在小屏幕上贴右边 */
@media (max-width: 900px) {
  .back-to-top {
    right: 16px;
  }
}

</style>
