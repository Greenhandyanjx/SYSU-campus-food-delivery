<template>
  <div class="agent-store-card">
    <!-- 卡片头部：logo + 商家名称 + 评分 -->
    <div class="card-header">
      <div class="merchant-info">
        <img v-if="data.logo" :src="data.logo" class="merchant-logo" alt="logo"
          @error="onLogoError" />
        <span v-else class="merchant-icon">🏪</span>
        <span class="merchant-name">{{ data.name || '商家' }}</span>
      </div>
      <div class="store-rating" v-if="data.rating > 0">
        <span class="rating-stars">{{ renderStars(data.rating) }}</span>
        <span class="rating-score">{{ data.rating }}</span>
      </div>
    </div>

    <!-- 销售信息 -->
    <div class="store-meta" v-if="data.sales > 0 || data.deliveryFee !== undefined || data.minOrder > 0">
      <span v-if="data.sales > 0" class="meta-item">📊 月售 {{ data.sales }}</span>
      <span v-if="data.deliveryFee !== undefined" class="meta-item">
        配送费 ¥{{ data.deliveryFee }}
      </span>
      <span v-if="data.minOrder > 0" class="meta-item">
        起送 ¥{{ data.minOrder }}
      </span>
    </div>

    <!-- 描述/位置 -->
    <div class="store-desc" v-if="data.desc">
      📍 {{ data.desc }}
    </div>

    <!-- 示例菜品列表（最多 3 个） -->
    <div class="dish-list" v-if="data.dishes && data.dishes.length > 0">
      <div class="dish-item" v-for="(dish, i) in data.dishes.slice(0, 3)" :key="i">
        <img v-if="dish.image" :src="dish.image" class="dish-img" alt="" @error="onDishImgError" />
        <div class="dish-info">
          <span class="dish-name">{{ dish.name }}</span>
          <span class="dish-price">¥{{ dish.price }}</span>
        </div>
      </div>
      <div class="dish-more" v-if="data.dishes.length > 3">
        ...还有 {{ data.dishes.length - 3 }} 个菜品
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <button class="action-btn primary" @click="viewMenu">
        🍽️ 查看菜单
      </button>
      <button class="action-btn outline" @click="sendToAgent('去 ' + data.name + ' 看看有什么好吃的')">
        🔍 去逛逛
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

export interface StoreDish {
  name: string
  price: number
  image?: string
}

export interface StoreCardData {
  id: number | string
  name: string
  rating: number
  sales: number
  desc?: string
  logo?: string
  deliveryFee?: number
  minOrder?: number
  dishes?: StoreDish[]
  tags?: string[]
}

const props = defineProps<{
  data: StoreCardData
}>()

const emit = defineEmits<{
  sendMessage: [text: string]
}>()

const router = useRouter()

function renderStars(rating: number): string {
  const full = Math.round(rating / 2)
  return '⭐'.repeat(Math.max(1, Math.min(full, 5)))
}

function onLogoError(e: Event) {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
}

function onDishImgError(e: Event) {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
}

function viewMenu() {
  router.push(`/user/store/${props.data.id}`)
}

function sendToAgent(text: string) {
  emit('sendMessage', text)
}
</script>

<style scoped>
.agent-store-card {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #e8e8e8;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  padding: 12px;
  margin-top: 8px;
  margin-bottom: 4px;
  font-size: 13px;
  line-height: 1.5;
  width: 100%;
  min-width: 280px;
}

/* 头部 */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.merchant-info {
  display: flex;
  align-items: center;
  gap: 6px;
}

.merchant-icon {
  font-size: 18px;
}

.merchant-logo {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  object-fit: cover;
  flex-shrink: 0;
}

.merchant-name {
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

.store-rating {
  display: flex;
  align-items: center;
  gap: 4px;
}

.rating-stars {
  font-size: 12px;
}

.rating-score {
  font-size: 13px;
  font-weight: 600;
  color: #FF6B35;
}

/* 销售信息 */
.store-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 12px;
  color: #888;
}

.meta-item {
  background: #f5f5f5;
  padding: 1px 8px;
  border-radius: 10px;
}

/* 描述 */
.store-desc {
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #f0f0f0;
}

/* 菜品列表 */
.dish-list {
  margin-bottom: 8px;
}

.dish-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.dish-img {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  object-fit: cover;
  flex-shrink: 0;
  background: #f5f5f5;
}

.dish-info {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-width: 0;
}

.dish-name {
  color: #555;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.dish-price {
  color: #FF6B35;
  font-weight: 600;
  margin-left: 8px;
}

.dish-more {
  color: #aaa;
  font-size: 12px;
  padding: 2px 0;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.action-btn {
  padding: 6px 12px;
  border-radius: 16px;
  border: 1px solid #e0e0e0;
  background: white;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  color: #555;
}

.action-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.action-btn:active {
  transform: translateY(0);
}

.action-btn.primary {
  border-color: #FF6B35;
  color: #FF6B35;
  background: #FFF3E8;
}

.action-btn.primary:hover {
  background: #FF6B35;
  color: white;
}

.action-btn.outline {
  border-color: #ddd;
  color: #666;
  background: #fafafa;
}

.action-btn.outline:hover {
  border-color: #bbb;
  background: #f0f0f0;
}
</style>
