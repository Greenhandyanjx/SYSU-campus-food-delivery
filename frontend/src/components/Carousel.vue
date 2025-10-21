<template>
  <div class="carousel-wrapper">
    <div
      class="carousel"
      :style="{ transform: `translateX(-${current * 100}%)` }"
    >
      <div
        class="slide"
        v-for="(img, idx) in images"
        :key="idx"
        :style="{ backgroundImage: `url(${img.src})` }"
      >
        <div class="overlay">
          <slot :index="idx"></slot>
        </div>
      </div>
    </div>

    <!-- 小圆点 -->
    <div class="dots">
      <span
        v-for="(img, idx) in images"
        :key="idx"
        class="dot"
        :class="{ active: idx === current }"
        @click="go(idx)"
      ></span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  images: { type: Array, default: () => [] },
  interval: { type: Number, default: 4000 },
})

const current = ref(0)
let timer = null

function go(i) {
  current.value = i
}

function next() {
  current.value = (current.value + 1) % props.images.length
}

onMounted(() => {
  timer = setInterval(next, props.interval)
})
onBeforeUnmount(() => clearInterval(timer))
</script>

<style scoped>
.carousel-wrapper {
  position: relative;
  overflow: hidden;
  border-radius: 10px;
}

.carousel {
  display: flex;
  transition: transform 0.6s ease;
  width: 100%;
  /* height: auto; */
}

.slide {
  min-width: 100%;
  aspect-ratio: 3 / 1;
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
  position: relative;
}

.overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(
    180deg,
    rgba(0, 0, 0, 0) 20%,
    rgba(0, 0, 0, 0.25) 100%
  );
  display: flex;
  align-items: flex-end;
  padding: 18px;
  box-sizing: border-box;
}

.dots {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  bottom: 8px;
  display: flex;
  gap: 8px;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  transition: background 0.3s;
}

.dot.active {
  background: #ff6b00;
}
</style>
