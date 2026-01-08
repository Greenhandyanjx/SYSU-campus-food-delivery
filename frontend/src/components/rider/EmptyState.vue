<template>
  <div class="empty-state">
    <div class="empty-content">
      <div class="empty-icon">
        <component :is="iconComponent" />
      </div>
      <h3 class="empty-title">{{ title }}</h3>
      <p class="empty-description">{{ description }}</p>
      <div class="empty-actions" v-if="$slots.actions">
        <slot name="actions" />
      </div>
    </div>
    <div class="empty-decoration" v-if="showDecoration">
      <div class="decoration-item" v-for="i in 6" :key="i"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  type: 'orders' | 'network' | 'search' | 'default';
  title?: string;
  description?: string;
  showDecoration?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showDecoration: true,
});

const iconComponent = computed(() => {
  const icons = {
    orders: '📦',
    network: '📡',
    search: '🔍',
    default: '📋',
  };
  return icons[props.type] || icons.default;
});
</script>

<style scoped lang="scss">
.empty-state {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 60px 40px;
  text-align: center;
  box-shadow: var(--rider-shadow);
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;

  &:hover {
    box-shadow: var(--rider-shadow-hover);
  }
}

.empty-content {
  position: relative;
  z-index: 2;
}

.empty-icon {
  font-size: 80px;
  margin-bottom: 24px;
  display: inline-block;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.empty-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--rider-text);
  margin: 0 0 12px 0;
}

.empty-description {
  font-size: 15px;
  color: var(--rider-sub);
  margin: 0 0 32px 0;
  line-height: 1.6;
}

.empty-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
}

.empty-decoration {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 1;
}

.decoration-item {
  position: absolute;
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: radial-gradient(circle, var(--rider-primary-light) 0%, transparent 70%);
  opacity: 0.1;
  animation: float-decoration 20s ease-in-out infinite;

  &:nth-child(1) {
    top: -50px;
    left: -50px;
    animation-delay: 0s;
  }
  &:nth-child(2) {
    top: -30px;
    right: -40px;
    width: 80px;
    height: 80px;
    animation-delay: 3s;
  }
  &:nth-child(3) {
    bottom: -60px;
    left: 20%;
    width: 120px;
    height: 120px;
    animation-delay: 6s;
  }
  &:nth-child(4) {
    bottom: -40px;
    right: 10%;
    width: 60px;
    height: 60px;
    animation-delay: 9s;
  }
  &:nth-child(5) {
    top: 30%;
    left: -30px;
    width: 90px;
    height: 90px;
    animation-delay: 12s;
  }
  &:nth-child(6) {
    top: 40%;
    right: -20px;
    width: 70px;
    height: 70px;
    animation-delay: 15s;
  }
}

@keyframes float-decoration {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(30px, -30px) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.9);
  }
}

// 响应式
@media (max-width: 768px) {
  .empty-state {
    padding: 40px 20px;
  }

  .empty-icon {
    font-size: 60px;
  }

  .empty-title {
    font-size: 18px;
  }

  .empty-description {
    font-size: 14px;
  }

  .decoration-item {
    width: 60px !important;
    height: 60px !important;
  }
}
</style>