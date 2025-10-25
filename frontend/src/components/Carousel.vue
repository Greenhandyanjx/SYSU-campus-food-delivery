<template>
  <div class="carousel-wrapper" ref="wrap" @mouseenter="pause" @mouseleave="resume">
    <div class="carousel-track" ref="track" :style="{ transform: `translateX(-${offset}px)` }">
      <div
        class="slide"
        v-for="(item, idx) in displayedSlides"
        :key="idx + '-' + (item && item.src)"
        :class="getSlideClass(idx)"
        @click="onSlideClick(idx)"
      >
        <img :src="item.src" :alt="item.title || 'banner'" draggable="false" />
        <div class="overlay" v-if="idx === currentPos">
          <h2>{{ item.title }}</h2>
          <p>{{ item.desc }}</p>
          <a :href="item.link" class="banner-btn">{{ item.buttonText }}</a>
        </div>
      </div>
    </div>
    <div class="dots">
      <span
        v-for="(img, idx) in images"
        :key="idx"
        class="dot"
        :class="{ active: idx === currentIndex }"
        @click="go(idx)"
      ></span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue'

const props = defineProps({
  images: { type: Array, default: () => [] },
  interval: { type: Number, default: 3500 },
  autoplay: { type: Boolean, default: true },
})

const currentIndex = ref(0)
const offset = ref(0)
const isTransitioning = ref(false)

// pos is the index in displayedSlides (0..n+1). starts at 1 (first real)
const pos = ref(1)
const currentPos = computed(() => pos.value)

const displayedSlides = computed(() => {
  const imgs = props.images || []
  const n = imgs.length
  if (n === 0) return []
  
  // 创建循环数组：[last, ...all, first]
  const arr = []
  const last = imgs[n - 1]
  const first = imgs[0]
  
  arr.push({ ...(last || {}), originalIndex: n - 1 })
  for (let i = 0; i < n; i++) arr.push({ ...(imgs[i] || {}), originalIndex: i })
  arr.push({ ...(first || {}), originalIndex: 0 })
  
  return arr
})

const wrap = ref(null)
const track = ref(null)
let timer = null

function getSlideClass(idx) {
  const posDiff = idx - currentPos.value
  if (posDiff === 0) return 'active'
  if (posDiff === -1) return 'left'
  if (posDiff === 1) return 'right'
  return 'hidden'
}

function go(i) {
  const n = (props.images || []).length
  if (n === 0 || isTransitioning.value) return
  isTransitioning.value = true
  const oldIndex = currentIndex.value
  const newIndex = (i + n) % n

  // detect wrapping
  const wrapForward = oldIndex === n - 1 && newIndex === 0
  const wrapBackward = oldIndex === 0 && newIndex === n - 1

  if (wrapForward) {
    // animate to duplicate first at pos = n+1
    pos.value = n + 1
    // keep currentIndex as newIndex for state
    currentIndex.value = newIndex
    updateOffset(true)
    // after animation, jump to real first
    setTimeout(() => {
      // jump without transition
      pos.value = 1
      updateOffset(false)
      isTransitioning.value = false
    }, 620)
  } else if (wrapBackward) {
    // animate to duplicate last at pos = 0
    pos.value = 0
    currentIndex.value = newIndex
    updateOffset(true)
    setTimeout(() => {
      pos.value = n
      updateOffset(false)
      isTransitioning.value = false
    }, 620)
  } else {
    // normal move
    currentIndex.value = newIndex
    pos.value = newIndex + 1
    updateOffset(true)
    setTimeout(() => {
      isTransitioning.value = false
    }, 620)
  }
}

function next() {
  go(currentIndex.value + 1)
}

function prev() {
  go(currentIndex.value - 1)
}

function pause() {
  if (timer) clearInterval(timer)
}

function resume() {
  if (props.autoplay) {
    pause()
    timer = setInterval(next, props.interval)
  }
}

function onSlideClick(idx) {
  const posDiff = idx - currentPos.value
  if (posDiff === -1) {
    prev()
  } else if (posDiff === 1) {
    next()
  }
}

function updateOffset(withTransition = true) {
  nextTick(() => {
    const w = wrap.value
    const t = track.value
    if (!w || !t) return
    
    // 添加防抖，避免频繁调用
    if (isTransitioning.value) return
    
    const slides = Array.from(t.querySelectorAll('.slide'))
    if (slides.length === 0) return
    
    const effectiveIndex = currentPos.value
    const activeSlide = slides[effectiveIndex]
    
    if (!activeSlide) {
      console.warn(`No active slide found at index ${effectiveIndex}`)
      return
    }
    
    // 等待下一帧确保布局稳定
    requestAnimationFrame(() => {
      const wrapWidth = w.clientWidth
      const slideWidth = activeSlide.offsetWidth
      
      if (wrapWidth <= 0 || slideWidth <= 0) {
        console.warn('Invalid dimensions:', { wrapWidth, slideWidth })
        return
      }
      
      const targetOffset = activeSlide.offsetLeft - (wrapWidth - slideWidth) / 2
      const maxOffset = Math.max(0, t.scrollWidth - wrapWidth)
      const clampedOffset = Math.max(0, Math.min(targetOffset, maxOffset))
      
      if (t) {
        t.style.transition = withTransition 
          ? 'transform 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)'
          : 'none'
        offset.value = clampedOffset
      }
    })
  })
}

onMounted(() => {
  updateOffset(false)
  if (props.autoplay) timer = setInterval(next, props.interval)
  window.addEventListener('resize', () => updateOffset(false))
})

onBeforeUnmount(() => {
  pause()
  window.removeEventListener('resize', () => updateOffset(false))
})

watch(() => props.images.length, () => {
  if (currentIndex.value >= props.images.length) currentIndex.value = 0
  updateOffset(false)
})
</script>

<style scoped>
.carousel-wrapper {
    align-items: center;
  position: relative;
  width: 100%;
  height: 400px; /* 稍微增加高度以容纳更大的中间图片 */
  overflow: hidden;
  display: block;
}

.carousel-track {
  display: flex;
  align-items: center;
  height: 100%;
  will-change: transform;
  padding: 0; /* 减少两侧间距 */
}

.slide {
  flex: 0 0 auto;
  position: relative;
  transition: all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  cursor: pointer;
  user-select: none;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.slide img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

/* 左侧图片 */
.slide.left {
  width: 60%; /* 更小的宽度 */
  margin-right: -10%; /* 负边距让中间图片可以覆盖 */
  transform: scale(0.75);
  opacity: 0.6;
  z-index: 1;
  filter: brightness(0.7);
}

/* 中间图片 */
.slide.active {
  width: 80%; /* 更大的宽度 */
  transform: scale(1);
  opacity: 1;
  z-index: 3;
  filter: brightness(1);
  box-shadow: 0 15px 40px rgba(0, 0, 0, 0.4);
}

/* 右侧图片 */
.slide.right {
  width: 60%; /* 更小的宽度 */
  margin-left: -10%; /* 负边距让中间图片可以覆盖 */
  transform: scale(0.75);
  opacity: 0.6;
  z-index: 1;
  filter: brightness(0.7);
}

/* 隐藏的图片 */
.slide.hidden {
  display: none;
}

.overlay {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.8), transparent);
  color: #fff;
  padding: 30px 25px 25px;
  z-index: 10;
}

.overlay h2 {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 700;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.5);
}

.overlay p {
  margin: 0 0 15px;
  font-size: 14px;
  line-height: 1.4;
  opacity: 0.9;
  max-width: 80%;
}

.banner-btn {
  display: inline-block;
  background: linear-gradient(45deg, #ff6b00, #ff8c00);
  color: #fff;
  padding: 10px 20px;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(255, 107, 0, 0.3);
}

.banner-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(255, 107, 0, 0.4);
}

.dots {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  bottom: 20px;
  display: flex;
  gap: 10px;
  z-index: 10;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: all 0.3s ease;
}

.dot.active {
  background: #ff6b00;
  transform: scale(1.2);
}

.dot:hover {
  background: rgba(255, 255, 255, 0.8);
}

/* 隐藏原生滚动条 */
.carousel-wrapper::-webkit-scrollbar {
  display: none;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .carousel-wrapper {
    height: 300px;
  }
  
  .slide.active {
    width: 70%;
  }
  
  .slide.left,
  .slide.right {
    width: 20%;
  }
  
  .overlay h2 {
    font-size: 18px;
  }
  
  .overlay p {
    font-size: 12px;
    max-width: 90%;
  }
}
</style>