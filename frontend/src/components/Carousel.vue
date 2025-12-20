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
          <button type="button" class="banner-btn" @click.stop.prevent="handleBannerClick(item)">{{ item.buttonText }}</button>
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

const emit = defineEmits(['banner-click'])

function handleBannerClick(item) {
  // emit to parent with item (originally from images)
  emit('banner-click', item)
}

const currentIndex = ref(0)
const offset = ref(0)
const isTransitioning = ref(false)

// pos is the index in displayedSlides (0..n+1). starts at 1 (first real)
const pos = ref(2)
const currentPos = computed(() => pos.value)

const displayedSlides = computed(() => {
  const imgs = props.images || []
  const n = imgs.length
  if (n === 0) return []

  // 创建循环数组：[last, ...all, first]
  const arr = []
  const last1 = imgs[n - 1]
  const last2 = imgs[n - 2]
  const first1 = imgs[0]
  const first2 = imgs[1]
  
  arr.push({ ...(last2 || {}), originalIndex: n - 2 })
  arr.push({ ...(last1 || {}), originalIndex: n - 1 })
  for (let i = 0; i < n; i++) arr.push({ ...(imgs[i] || {}), originalIndex: i })
  arr.push({ ...(first1 || {}), originalIndex: 0 })
  arr.push({ ...(first2 || {}), originalIndex: 1 })
  
  return arr
})



const wrap = ref(null)
const track = ref(null)
let timer = null
let lastTransitionHandler = null

function getSlideClass(idx) {
  const posDiff = idx - currentPos.value
  if (posDiff === 0) return 'active'
  if (posDiff === -1) return 'left'
  if (posDiff === 1) return 'right'
  // 👇 新出现的那张，给个 entering 类用于渐入动画
  if (posDiff === 2 || posDiff === -2) return 'entering'
  return 'hidden'
}


function go(i) {
  const n = (props.images || []).length
  if (n === 0) return
  const oldIndex = currentIndex.value
  const newIndex = ((i % n) + n) % n
  if (newIndex === oldIndex) return

  const t = track.value
  // 如果 track 元素不存在，直接更新状态并退出（避免 isTransitioning 被误置）
  if (!t) {
    currentIndex.value = newIndex
    pos.value = newIndex + 2
    updateOffset(false)
    isTransitioning.value = false
    resume()
    return
  }
  // 准备 transitionend 处理器（保证每次只有一个监听器）
  if (t) {
    if (lastTransitionHandler && typeof lastTransitionHandler === 'function') {
      t.removeEventListener('transitionend', lastTransitionHandler)
      lastTransitionHandler = null
    }

    lastTransitionHandler = function handleTransitionEnd() {
      try { t.removeEventListener('transitionend', handleTransitionEnd) } catch (e) {}
      // 如果滑动到了虚拟首/尾，瞬移回真实位置
      if (pos.value === n + 2) {
        t.style.transition = 'none'
        pos.value = 2
        updateOffset(false)
      } else if (pos.value === 1) {
        t.style.transition = 'none'
        pos.value = n + 1
        updateOffset(false)
      }
      isTransitioning.value = false
      lastTransitionHandler = null
    }

    t.addEventListener('transitionend', lastTransitionHandler)
  }

  // 标记为动画中，避免并发修改状态
  isTransitioning.value = true

  const wrapForward = oldIndex === n - 1 && newIndex === 0
  const wrapBackward = oldIndex === 0 && newIndex === n - 1

  if (wrapForward) {
    pos.value = n + 2
    currentIndex.value = newIndex
    updateOffset(true)
  } else if (wrapBackward) {
    pos.value = 1
    currentIndex.value = newIndex
    updateOffset(true)
  } else {
    // 直接跳转到目标 slide（支持非左右邻居）
    currentIndex.value = newIndex
    pos.value = newIndex + 2
    updateOffset(true)
  }

  // 手动切换后重启自动播放计时器
  resume()
}


function next() {
  go(currentIndex.value + 1)
}

function prev() {
  go(currentIndex.value - 1)
}

function pause() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
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
    // if (isTransitioning.value) return
    
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
// 允许略超一点点范围，避免出现闪现
      const clampedOffset = Math.max(0, Math.min(targetOffset + 1, maxOffset + 1))
      
      if (t) {
        t.style.transition = withTransition 
          ? 'transform 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94)'
          : 'none'
        offset.value = clampedOffset
        t.style.transform = `translateX(-${offset.value}px)`
      }
    })
  })
}

// 使用命名的 resize 处理函数，保证 add/remove 使用相同引用
const onResize = () => updateOffset(false)

onMounted(() => {
  updateOffset(false)
  if (props.autoplay) timer = setInterval(next, props.interval)
  window.addEventListener('resize', onResize)

  // Bind image load listeners so we recalc offset after images finish loading.
  // This prevents incorrect initial measurements on first visit.
  bindImageLoadListeners()

  // Also recalc once when the whole window load fires (images cached or slow loads)
  window.addEventListener('load', onWindowLoad)

  // Small timeout fallback: ensure offset recalculated after micro delay
  // in case images load after initial paint.
  _retryTimeout = setTimeout(() => updateOffset(false), 220)
})

onBeforeUnmount(() => {
  pause()
  window.removeEventListener('resize', onResize)
  window.removeEventListener('load', onWindowLoad)
  // 清理可能残留的 transitionend 监听
  const t = track.value
  if (t && lastTransitionHandler) {
    try { t.removeEventListener('transitionend', lastTransitionHandler) } catch (e) {}
    lastTransitionHandler = null
  }

  // 清理 image load listeners and any retry timers
  removeImageLoadListeners()
  if (_retryTimeout) {
    clearTimeout(_retryTimeout)
    _retryTimeout = null
  }
})

// --- 图片加载监听支持：确保在图片加载完成后重新计算 offset ---
let _imageLoadHandlers = null
let _retryTimeout = null

function bindImageLoadListeners() {
  const t = track.value
  if (!t) return
  const imgs = Array.from(t.querySelectorAll('img'))
  if (!imgs || imgs.length === 0) return

  _imageLoadHandlers = []
  imgs.forEach((img) => {
    // 如果图片已经加载完成（cache），仍然触发一次 update
    const handler = () => {
      try { img.removeEventListener('load', handler) } catch (e) {}
      updateOffset(false)
    }
    img.addEventListener('load', handler)
    _imageLoadHandlers.push({ img, handler })
    if (img.complete) {
      // defer to next tick to avoid layout thrash
      requestAnimationFrame(() => updateOffset(false))
    }
  })
}

function removeImageLoadListeners() {
  if (!_imageLoadHandlers) return
  _imageLoadHandlers.forEach(({ img, handler }) => {
    try { img.removeEventListener('load', handler) } catch (e) {}
  })
  _imageLoadHandlers = null
}

function onWindowLoad() {
  updateOffset(false)
}

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
  height: 450px;
  overflow: hidden;
  display: block;
}

.carousel-track {
  display: flex;
  align-items: center;
  height: 100%;
  will-change: transform;
  padding: 0;
}

.slide {
  flex: 0 0 auto;
  position: relative;
  transition: all 0.6s cubic-bezier(0.25, 0.46, 0.45, 0.94);
  cursor: pointer;
  user-select: none;
  border-radius: 32px;
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
.slide.entering {
  z-index: 1;
  opacity: 0;
  scale: 0.5;
  animation: enterZoom 0.6s forwards;
}

/* 新出现图的放大渐入关键帧 */
@keyframes enterZoom {
  0% {
    opacity: 0;
    transform: scale(0.5);
  }
  100% {
    opacity: 0.9;
    transform: scale(0.95);
  }
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
  border-width: 0px;
  text-decoration: none;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.3s ease;
  box-shadow: 0 3px 3px rgba(255, 107, 0, 0.3);
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