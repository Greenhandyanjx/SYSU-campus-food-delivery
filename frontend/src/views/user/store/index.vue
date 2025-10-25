<template>
  <div class="store-page">
    <!-- 顶部店铺信息（半透明浮层，置于背景图之上） -->
    <div class="hero">
      <div class="hero-inner">
        <div class="logo">
          <img :src="store.logo || '/src/assets/noImg.png'" alt="logo" />
        </div>
        <div class="hero-meta">
          <h1 class="store-name">{{ store.name || '店铺名称' }}</h1>
          <div class="store-sub">
            <span class="rating">⭐ {{ store.rating || 4.8 }}</span>
            <span class="bullet">·</span>
            <span>{{ store.deliveryTime || '30 分钟内' }}</span>
            <span class="bullet">·</span>
            <span>{{ store.minOrder ? `起送 ¥${store.minOrder}` : '无起送' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 主体三栏：分类 / 菜品 / 背景信息区 -->
    <div class="main-grid">
      <!-- 左侧分类 -->
      <nav class="cate-col">
        <ul>
          <li
            v-for="(c, i) in categories"
            :key="c.id || i"
            :class="{ active: selectedCategory === c.id }"
            @click="selectCategory(c.id)"
          >
            <div class="cate-left">
              <img v-if="c.icon" :src="c.icon" class="cate-icon" />
              <span class="cate-label">{{ c.name }}</span>
            </div>
            <span class="badge" v-if="c.count > 0">{{ c.count }}</span>
          </li>
        </ul>
      </nav>

      <!-- 中间主内容（60% 宽） -->
      <section class="content-col">
        <div class="section-title">
          <h2>{{ currentCategoryName }}</h2>
          <div class="search-row">
            <el-input
              v-model="query"
              placeholder="搜索菜品 / 口味"
              clearable
              @keyup.enter="onSearch"
              size="small"
            >
              <template #append>
                <el-button size="small" type="warning" @click="onSearch">搜索</el-button>
              </template>
            </el-input>
          </div>
        </div>

        <div class="dishes">
          <div
            class="dish-card"
            v-for="(d, idx) in dishesFiltered"
            :key="d.id || idx"
          >
            <img class="thumb" :src="d.image || '/src/assets/noImg.png'" />
            <div class="dish-info">
              <div class="dish-top">
                <div class="name">{{ d.name }}</div>
                <div class="price">¥{{ formatPrice(d.price) }}</div>
              </div>
              <div class="desc">{{ d.desc }}</div>
              <div class="dish-bottom">
                <div class="tags">
                  <span class="tag" v-for="t in d.tags || []" :key="t">{{ t }}</span>
                </div>
                <div class="controls">
                  <el-button
                    size="mini"
                    circle
                    type="warning"
                    @click="dec(d)"
                    :disabled="(d.count || 0) <= 0"
                  >
                    -
                  </el-button>
                  <span class="count">{{ d.count || 0 }}</span>
                  <el-button size="mini" circle type="warning" @click="add(d)">
                    +
                  </el-button>
                </div>
              </div>
            </div>
          </div>
          <div v-if="dishesFiltered.length === 0" class="empty">暂无菜品</div>
        </div>
      </section>

      <!-- 右侧视觉区：背景图裁切 + 店铺小组件 -->
      <aside class="visual-col" :style="visualBgStyle">
        <div class="visual-overlay">
          <div class="store-card">
            <h3>店铺信息</h3>
            <p>{{ store.desc }}</p>
            <div class="meta-row">
              <div>营业时间：{{ store.openTime || '10:00 - 21:00' }}</div>
              <div>电话：{{ store.phone || '未填写' }}</div>
            </div>
            <el-button size="small" type="primary" @click="openShop">进入店铺</el-button>
          </div>

          <div class="cart-preview">
            <h4>已选商品</h4>
            <div v-if="cart.length === 0" class="empty-cart">购物车为空</div>
            <ul v-else>
              <li v-for="(c, i) in cart" :key="i">
                {{ c.name }} × {{ c.qty }}
                <span class="sub">¥{{ (c.qty * c.price).toFixed(2) }}</span>
              </li>
            </ul>
            <div class="cart-total">
              <div>合计：<strong>¥{{ cartTotal.toFixed(2) }}</strong></div>
              <el-button size="small" type="warning" @click="checkout" :disabled="cart.length===0">去结算</el-button>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
  <!-- 固定在右下角的购物栏 -->
<div class="floating-cart" @click.self="toggleCartPopup">
  <div class="cart-left" @click="toggleCartPopup">
    <div class="cart-icon">
      <i class="el-icon-shopping-cart-2"></i>
      <span v-if="cartCount > 0" class="badge">{{ cartCount }}</span>
    </div>
    <div class="cart-info">
      <div class="text-top" v-if="cartTotal < (store.minOrder || 0)">
        <span>{{ store.minOrder ? `¥${store.minOrder} 起送` : '无起送' }}</span>
        <span v-if="store.minOrder && cartTotal > 0" class="gap">还差 ¥{{ (store.minOrder - cartTotal).toFixed(2) }}</span>
      </div>
      <div class="text-top" v-else>
        <strong>共 ¥{{ cartTotal.toFixed(2) }}</strong>
      </div>
    </div>
  </div>

  <el-button
    v-if="cartTotal >= (store.minOrder || 0)"
    size="small"
    type="warning"
    class="checkout-btn"
    @click.stop="checkout"
  >
    去结算
  </el-button>
</div>

<!-- 弹出的购物车窗口 -->
<transition name="fade-slide">
  <div v-if="showCartPopup" class="cart-popup">
    <div class="cart-popup-header">
      <span>已选商品</span>
      <i class="el-icon-close" @click="closeCartPopup"></i>
    </div>
    <ul class="cart-popup-list">
      <li v-for="(c, i) in cart" :key="i">
        <div class="name">{{ c.name }}</div>
        <div class="controls">
          <el-button
            size="mini"
            circle
            type="warning"
            @click.stop="decFromPopup(c)"
            :disabled="c.qty <= 0"
          >-</el-button>
          <span class="count">{{ c.qty }}</span>
          <el-button size="mini" circle type="warning" @click.stop="addFromPopup(c)">+</el-button>
        </div>
        <div class="price">¥{{ (c.qty * c.price).toFixed(2) }}</div>
      </li>
    </ul>
    <div v-if="cart.length === 0" class="cart-popup-empty">购物车为空</div>
  </div>
</transition>

</template>

<script setup lang="ts">
import { reactive, ref, computed, onMounted,onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getStoreByName, getDishesByStore, addToCart, removeFromCart, getCart } from '@/api/user/store'

/* ------------------ 核心数据定义 ------------------ */

// 使用 ref 而不是 reactive，以便和后端数据动态替换兼容
const store = ref<any>({})
const categories = ref<any[]>([])
const dishes = ref<any[]>([])
const cart = ref<any[]>([])
const selectedCategory = ref('all')
const query = ref('')
const route = useRoute()
/* ------------------ Demo 数据备用 ------------------ */
const demoStore = {
  id: 101,
  name: '金色小厨 · 校园餐厅',
  logo: '/src/assets/demo/noImg.png',
  desc: '校园人气食堂 · 用心做好每一顿饭',
  rating: 4.9,
  deliveryTime: '25~35 分钟',
  minOrder: 15,
  openTime: '10:30 - 21:00',
  phone: '138-8888-6666',
  bg: '/src/assets/demo/store_bg.jpg'
}

const demoDishes = [
  // 🍛 主食类
  { id: 1, name: '宫保鸡丁盖饭', price: 22, desc: '微辣，花生香浓郁，经典川味', tags: ['招牌', '微辣'], image: '/src/assets/demo/dish_gongbao.jpg', category: '主食', sales: 289, count: 0 },
  { id: 2, name: '黑椒牛柳盖饭', price: 26, desc: '嫩滑牛柳配黑椒汁，香气四溢', tags: ['推荐'], image: '/src/assets/demo/dish_beef.jpg', category: '主食', sales: 175, count: 0 },
  { id: 3, name: '麻辣香锅', price: 28, desc: '多种蔬菜配肉片，麻辣过瘾', tags: ['麻辣'], image: '/src/assets/demo/dish_hotpot.jpg', category: '主食', sales: 199, count: 0 },
  { id: 4, name: '卤肉饭', price: 18, desc: '酱香浓郁，肥瘦相间，香气扑鼻', tags: [], image: '/src/assets/demo/dish_lurou.jpg', category: '主食', sales: 246, count: 0 },
  { id: 5, name: '鸡排咖喱饭', price: 24, desc: '香浓日式咖喱搭配炸鸡排', tags: ['人气'], image: '/src/assets/demo/dish_curry.jpg', category: '主食', sales: 312, count: 0 },

  // 🍜 面食类
  { id: 6, name: '红烧牛肉面', price: 20, desc: '汤浓味厚，筋道面条', tags: ['热销'], image: '/src/assets/demo/dish_beefnoodle.jpg', category: '面食', sales: 355, count: 0 },
  { id: 7, name: '重庆小面', price: 16, desc: '香辣鲜香，一口上瘾', tags: ['辣'], image: '/src/assets/demo/dish_chongqing.jpg', category: '面食', sales: 187, count: 0 },
  { id: 8, name: '炸酱面', price: 15, desc: '家常口味，酱香浓郁', tags: ['传统'], image: '/src/assets/demo/dish_zhajiang.jpg', category: '面食', sales: 99, count: 0 },
  { id: 9, name: '豚骨拉面', price: 27, desc: '日式浓汤拉面，叉烧入味', tags: ['精选'], image: '/src/assets/demo/dish_ramen.jpg', category: '面食', sales: 211, count: 0 },

  // 🍟 小吃类
  { id: 10, name: '炸鸡块', price: 14, desc: '金黄酥脆，外酥里嫩', tags: ['人气'], image: '/src/assets/demo/snack_chicken.jpg', category: '小吃', sales: 423, count: 0 },
  { id: 11, name: '薯条', price: 9, desc: '炸至金黄，搭配番茄酱最佳', tags: [], image: '/src/assets/demo/snack_fries.jpg', category: '小吃', sales: 278, count: 0 },
  { id: 12, name: '香辣鱿鱼须', price: 19, desc: '香辣劲爆，啤酒好搭档', tags: ['辣'], image: '/src/assets/demo/snack_squid.jpg', category: '小吃', sales: 143, count: 0 },
  { id: 13, name: '煎饺', price: 12, desc: '煎至金黄，皮薄馅多', tags: ['推荐'], image: '/src/assets/demo/snack_dumpling.jpg', category: '小吃', sales: 189, count: 0 },

  // 🍰 甜点类
  { id: 14, name: '提拉米苏', price: 18, desc: '意式风情，香滑浓郁', tags: ['甜品'], image: '/src/assets/demo/dessert_tiramisu.jpg', category: '甜点', sales: 92, count: 0 },
  { id: 15, name: '芝士蛋糕', price: 16, desc: '香甜柔滑，芝士浓香', tags: [], image: '/src/assets/demo/dessert_cheese.jpg', category: '甜点', sales: 78, count: 0 },
  { id: 16, name: '草莓布丁', price: 13, desc: '草莓果酱加布丁底，清爽可口', tags: ['新品'], image: '/src/assets/demo/dessert_pudding.jpg', category: '甜点', sales: 106, count: 0 },

  // 🧋 饮品类
  { id: 17, name: '珍珠奶茶', price: 12, desc: '香浓奶香配Q弹珍珠', tags: ['热销'], image: '/src/assets/demo/drink_milktea.jpg', category: '饮品', sales: 509, count: 0 },
  { id: 18, name: '柠檬蜂蜜水', price: 10, desc: '酸甜解腻，清爽好喝', tags: ['清新'], image: '/src/assets/demo/drink_lemon.jpg', category: '饮品', sales: 187, count: 0 },
  { id: 19, name: '芒果冰沙', price: 15, desc: '冰凉爽口，果香浓郁', tags: ['夏日限定'], image: '/src/assets/demo/drink_mango.jpg', category: '饮品', sales: 210, count: 0 },
  { id: 20, name: '西瓜汁', price: 11, desc: '新鲜现榨，夏日解暑', tags: ['新鲜'], image: '/src/assets/demo/drink_watermelon.jpg', category: '饮品', sales: 158, count: 0 },

  // 🍱 套餐类
  { id: 21, name: '学生特惠套餐A', price: 29, desc: '主食+饮品+小吃组合', tags: ['套餐'], image: '/src/assets/demo/set_a.jpg', category: '套餐', sales: 311, count: 0 },
  { id: 22, name: '情侣双人套餐', price: 55, desc: '两份主食+甜点+饮品', tags: ['双人'], image: '/src/assets/demo/set_b.jpg', category: '套餐', sales: 133, count: 0 }
]
/* ------------------ 页面加载与接口请求 ------------------ */
async function load() {
  const name = decodeURIComponent(String(route.params.name || ''))
  if (!name) {
    useDemoData()
    return
  }
  try {
    const res = await getStoreByName(name)
    const data = res && res.data ? res.data.data || res.data : res
    if (!data) throw new Error('无返回数据')
    store.value = data

    const id = data.id || data.storeId
    if (!id) throw new Error('无有效店铺ID')

    const r2 = await getDishesByStore(id)
    const dd = r2 && r2.data ? r2.data.data || r2.data : r2
    if (!dd || !Array.isArray(dd) || dd.length === 0) throw new Error('空菜品')
    dishes.value = dd
    generateCategories()
  } catch (e) {
    console.warn('加载失败，使用Demo数据:', e)
    useDemoData()
    cart.value = []
  }
}

/* ------------------ 辅助函数 ------------------ */
function useDemoData() {
  store.value = demoStore
  dishes.value = demoDishes
  generateCategories()
}

function generateCategories() {
  const cset = new Set(dishes.value.map(d => d.category))
  categories.value = [{ id: 'all', name: '全部', count: dishes.value.length }]
  cset.forEach(c => {
    categories.value.push({
      id: c,
      name: c,
      count: dishes.value.filter(d => d.category === c).length
    })
  })
}

// 刷新购物车
async function refreshCart() {
  try {
    const r = await getCart({ storeId: store.value.id })
    const data = r && r.data ? r.data.data || r.data : r
    cart.value = data && Array.isArray(data) ? data : []
    // 同步购物车数量到菜品
    for (const d of dishes.value) {
      const item = cart.value.find(c => c.dishId === d.id)
      d.count = item ? item.qty : 0
    }
  } catch (e) {
    cart.value = []
  }
}

/* ------------------ 用户交互逻辑 ------------------ */

// 添加到购物车(后端实现接口后使用)
// async function add(d: any) {
//   try {
//     await addToCart({
//       storeId: store.value.id,
//       dishId: d.id,
//       name: d.name,
//       price: d.price,
//       qty: 1
//     })
//     d.count = (d.count || 0) + 1
//     ElMessage.success('已加入购物车')
//     await refreshCart()
//   } catch (e: any) {
//     ElMessage.error('加入购物车失败: ' + (e.message || ''))
//   }
// }

// // 移除购物车项
// async function dec(d: any) {
//   if (!d.count || d.count <= 0) return
//   try {
//     await removeFromCart({
//       storeId: store.value.id,
//       dishId: d.id,
//       qty: 1
//     })
//     d.count--
//     ElMessage.success('已从购物车移除')
//     await refreshCart()
//   } catch (e: any) {
//     ElMessage.error('移除失败: ' + (e.message || ''))
//   }
// }
async function add(d: any) {
  d.count = (d.count || 0) + 1
  const exist = cart.value.find(c => c.id === d.id)
  if (exist) {
    exist.qty++
  } else {
    cart.value.push({ id: d.id, name: d.name, price: d.price, qty: 1 })
  }
}

async function dec(d: any) {
  if ((d.count || 0) <= 0) return
  d.count--
  const exist = cart.value.find(c => c.id === d.id)
  if (exist) {
    exist.qty--
    if (exist.qty <= 0)
      cart.value = cart.value.filter(c => c.id !== d.id)
  }
}
/* ------------------ 页面展示计算属性 ------------------ */

// 搜索与分类过滤
const dishesFiltered = computed(() => {
  const q = query.value.trim().toLowerCase()
  return dishes.value.filter(d => {
    const okCate = selectedCategory.value === 'all' || d.category === selectedCategory.value
    const okQuery =
      !q ||
      d.name.toLowerCase().includes(q) ||
      (d.desc && d.desc.toLowerCase().includes(q))
    return okCate && okQuery
  })
})

const currentCategoryName = computed(() => {
  const c = categories.value.find(x => x.id === selectedCategory.value)
  return c ? c.name : '菜品'
})

const cartTotal = computed(() => {
  return cart.value.reduce((sum, item) => sum + item.qty * item.price, 0)
})

/* ------------------ 其他UI事件 ------------------ */

function selectCategory(id: string) {
  selectedCategory.value = id
}

function formatPrice(p: number) {
  return p.toFixed(2).replace(/\.00$/, '')
}

function onSearch() {
  // 本地过滤即可，如需后端搜索可在此发请求
}

function openShop() {
  window.open(`/store/${store.value.name}`, '_blank')
}

function checkout() {
  if (cart.value.length === 0) {
    ElMessage.warning('购物车为空')
    return
  }
  ElMessage.success('跳转结算页')
  // 这里可跳转结算路由
}

/* ------------------ 背景与挂载 ------------------ */

const visualBgUrl = ref('/src/assets/login/img_denglu_bj.jpg')
const visualBgStyle = computed(() => ({
  backgroundImage: `url(${visualBgUrl.value})`
}))

onMounted(async () => {
  await load()
  await refreshCart()
})
const showCartPopup = ref(false)

function toggleCartPopup() {
  if (cart.value.length === 0) {
    ElMessage.info('购物车为空')
    return
  }
  showCartPopup.value = !showCartPopup.value
}

function closeCartPopup() {
  showCartPopup.value = false
}

// 点击页面空白处关闭购物车窗口
function handleOutsideClick(e: MouseEvent) {
  const popup = document.querySelector('.cart-popup')
  const bar = document.querySelector('.floating-cart')
  if (
    showCartPopup.value &&
    popup &&
    !popup.contains(e.target as Node) &&
    !bar?.contains(e.target as Node)
  ) {
    showCartPopup.value = false
  }
}
onMounted(() => document.addEventListener('click', handleOutsideClick))
onBeforeUnmount(() => document.removeEventListener('click', handleOutsideClick))

// 当前购物车数量
const cartCount = computed(() =>
  cart.value.reduce((sum, c) => sum + c.qty, 0)
)

// 从弹窗加减商品
function addFromPopup(c: any) {
  const d = dishes.value.find(x => x.id === c.dishId || x.id === c.id)
  if (d) add(d)
}
function decFromPopup(c: any) {
  const d = dishes.value.find(x => x.id === c.dishId || x.id === c.id)
  if (d) dec(d)
}
</script>


<style scoped>
/* 主题变量（黄色系） */
:root{
  --yellow-50: #fff8e1;
  --yellow-100: #fff1b8;
  --yellow-200: #ffd24d;
  --yellow-300: #ffc107; /* 主色 */
  --yellow-400: #ffb000;
  --text-dark: #222;
  --muted: #666;
  --card-shadow: 0 6px 18px rgba(34,34,34,0.08);
  --glass: rgba(255,255,255,0.6);
}

/* 页面基础 */
.store-page{
  font-family: -apple-system, "Segoe UI", Roboto, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Helvetica Neue", Arial;
  color: var(--text-dark);
  background: linear-gradient(180deg, #fff 0%, #fffaf0 60%);
  min-height: 100vh;
}

/* 顶部 hero：半透明框，位于视觉背景上方 */
.hero{
  height: 140px;
  background: linear-gradient(90deg, rgba(255,193,7,0.08), rgba(255,235,59,0.03));
  display:flex;
  align-items:center;
  justify-content:center;
  padding: 16px 24px;
  box-shadow: var(--card-shadow);
  margin-bottom: 18px;
}
.hero-inner{
  width: 1180px;
  display:flex;
  gap:16px;
  align-items:center;
}
.logo img{
  width:76px;height:76px;border-radius:10px;object-fit:cover;border:4px solid rgba(255,255,255,0.6);
  box-shadow: 0 4px 12px rgba(0,0,0,0.06);
}
.store-name{
  margin:0;font-size:20px;color: #2b2b2b;font-weight:700;
}
.store-sub{ color:var(--muted); margin-top:6px; font-size:13px; display:flex; gap:8px; align-items:center;}
.rating{ color:#b76e00; font-weight:600;}

/* 三栏布局 */
.main-grid{
  width: 1180px;
  margin: 0 auto 60px;
  display: grid;
  grid-template-columns: 20% 60% 20%;
  gap: 20px;
}

/* 左侧分类 */
.cate-col{
  position: relative;
  top: 0;
  align-self:start;
  background: linear-gradient(180deg, rgba(255,250,240,0.8), rgba(255,255,255,0.8));
  border-radius: 10px;
  padding: 12px;
  box-shadow: var(--card-shadow);
  height: calc(100vh - 220px);
  overflow:auto;
  position: sticky;
  top: 90px;
}
.cate-col ul{ list-style:none; padding:0; margin:0; }
.cate-col li{
  display:flex; justify-content:space-between; align-items:center;
  padding:10px 8px; margin-bottom:6px; border-radius:8px; cursor:pointer;
  transition: all .18s;
}
.cate-col li:hover{ transform: translateX(6px); background: rgba(255,243,205,0.6); }
.cate-col li.active{ background: linear-gradient(90deg,#fff3cd,#fff7e0); box-shadow: inset 0 1px 0 rgba(255,255,255,0.6); }
.cate-left{ display:flex; gap:8px; align-items:center; }
.cate-icon{ width:28px; height:28px; object-fit:cover; border-radius:6px; }
.cate-label{ font-weight:600; color:var(--text-dark); }
.badge{ background: #ffb74d; color:#fff; padding:4px 8px; border-radius:12px; font-size:12px; }

/* 中间内容 */
.content-col{
  background: rgba(255,255,255,0.9);
  padding: 16px;
  border-radius: 10px;
  box-shadow: var(--card-shadow);
  min-height: 400px;
}
.section-title{ display:flex; justify-content:space-between; align-items:center; margin-bottom:12px; gap:12px; }
.section-title h2{ margin:0; font-size:18px; font-weight:700; color:#333; }
.search-row{ width:320px; }

.dishes{ display:flex; flex-direction:column; gap:12px; }
.dish-card{
  display:flex; gap:12px; padding:12px; border-radius:10px; align-items:flex-start;
  transition: transform .18s, box-shadow .18s;
  background: linear-gradient(180deg, rgba(255,255,255,0.9), #fff);
  box-shadow: 0 6px 14px rgba(0,0,0,0.04);
}
.dish-card:hover{ transform: translateY(-6px); box-shadow: 0 14px 30px rgba(0,0,0,0.08); }
.thumb{ width:96px; height:72px; border-radius:8px; object-fit:cover; flex-shrink:0; }
.dish-info{ flex:1; display:flex; flex-direction:column; gap:6px; }
.dish-top{ display:flex; justify-content:space-between; align-items:center; }
.name{ font-weight:700; font-size:16px; }
.price{ color:#d97706; font-weight:700; }
.desc{ font-size:13px; color:var(--muted); }
.dish-bottom{ display:flex; justify-content:space-between; align-items:center; margin-top:6px; }
.tags .tag{ background: rgba(255,193,7,0.12); color:#b06b00; padding:4px 8px; border-radius:12px; font-size:12px; margin-right:6px; }
.controls{ display:flex; align-items:center; gap:8px; }
.count{ min-width:28px; text-align:center; font-weight:600; }

/* 右侧视觉区 */
.visual-col{
  border-radius: 12px;
  overflow: hidden;
  background-size: cover;
  background-position: center;
  min-height: 420px;
  position: relative;
  display:flex;
  align-items:center;
  justify-content:center;
}
.visual-overlay{
  width:100%; height:100%;
  backdrop-filter: blur(6px);
  background: linear-gradient(180deg, rgba(255,255,255,0.28), rgba(255,255,255,0.48));
  padding:18px;
  display:flex;
  flex-direction:column;
  justify-content:space-between;
}
.store-card{
  background: rgba(255,255,255,0.85);
  border-radius:10px; padding:12px; box-shadow: var(--card-shadow);
}
.store-card h3{ margin:0 0 6px 0; font-size:16px; }
.meta-row{ font-size:13px; color:var(--muted); margin-top:8px; display:flex; gap:12px; }

/* 购物车预览 */
.cart-preview{ margin-top:12px; background: rgba(255,255,255,0.9); border-radius:10px; padding:10px; box-shadow: var(--card-shadow); }
.cart-preview h4{ margin:0 0 8px 0; }
.cart-preview ul{ margin:0; padding:0; list-style:none; max-height:150px; overflow:auto; }
.cart-preview li{ display:flex; justify-content:space-between; padding:6px 0; border-bottom:1px dashed rgba(0,0,0,0.04); }
.cart-total{ display:flex; justify-content:space-between; align-items:center; margin-top:10px; }

/* empty 状态 */
.empty, .empty-cart{ text-align:center; padding:24px 0; color:var(--muted) }

/* 响应式（窄屏） */
@media (max-width: 1000px){
  .main-grid{
    grid-template-columns: 1fr;
    width: calc(100% - 40px);
    margin: 0 auto;
  }
  .cate-col{ position:relative; height:auto; top:0; display:flex; padding:8px; overflow:auto; }
  .visual-col{ display:none; } /* 移动端隐藏右侧视觉区，避免占用空间 */
}
/* 浮动购物栏 */
.floating-cart {
  position: fixed;
  right: 40px;
  bottom: 30px;
  background: #fff;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
  border-radius: 50px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 5px 10px;
  width: 17%;
  height: 50px;
  z-index: 1000;
  cursor: pointer;
  transition: all 0.25s;
}
.floating-cart:hover {
  transform: translateY(-2px);
}
.cart-left {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  flex: 1;
}
.cart-icon {
  position: relative;
  font-size: 22px;
  color: #ffb000;
}
.cart-icon .badge {
  position: absolute;
  top: -6px;
  right: -10px;
  background: #ff4d4f;
  color: white;
  border-radius: 50%;
  font-size: 12px;
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.cart-info .text-top {
  font-size: 13px;
  color: #444;
}
.cart-info .gap {
  margin-left: 6px;
  color: #ff9800;
}
.checkout-btn {
  border-radius: 20px;
  font-weight: 600;
}

/* 弹出购物车窗口 */
.cart-popup {
  position: fixed;
  right: 40px;
  bottom: 90px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.25);
  width: 320px;
  max-height: 420px;
  overflow: hidden;
  z-index: 1001;
  display: flex;
  flex-direction: column;
}
.cart-popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #ffeb99;
  padding: 8px 12px;
  font-weight: 600;
  color: #333;
}
.cart-popup-header i {
  cursor: pointer;
  font-size: 18px;
  color: #555;
}
.cart-popup-list {
  list-style: none;
  padding: 10px 12px;
  margin: 0;
  overflow-y: auto;
  flex: 1;
}
.cart-popup-list li {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 0;
  border-bottom: 1px dashed rgba(0, 0, 0, 0.05);
}
.cart-popup-list .name {
  flex: 1;
  font-size: 14px;
}
.cart-popup-list .controls {
  display: flex;
  align-items: center;
  gap: 6px;
}
.cart-popup-list .price {
  width: 60px;
  text-align: right;
  font-weight: 600;
  color: #d97706;
}
.cart-popup-empty {
  text-align: center;
  color: #888;
  padding: 40px 0;
}

/* 动画 */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.25s ease;
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(20px);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

/* 响应式隐藏 */
@media (max-width: 768px) {
  .floating-cart {
    right: 16px;
    bottom: 20px;
    width: 90%;
    border-radius: 12px;
    justify-content: space-between;
  }
  .cart-popup {
    right: 16px;
    width: calc(100% - 32px);
  }
}

</style>
