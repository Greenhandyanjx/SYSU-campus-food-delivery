<template>
  <div class="skeleton-container">
    <div class="skeleton-card" v-for="i in count" :key="i">
      <div class="skeleton-header">
        <div class="skeleton-avatar shimmer"></div>
        <div class="skeleton-header-info">
          <div class="skeleton-title shimmer"></div>
          <div class="skeleton-subtitle shimmer"></div>
        </div>
        <div class="skeleton-tag shimmer"></div>
      </div>
      <div class="skeleton-body">
        <div class="skeleton-section">
          <div class="skeleton-section-title shimmer"></div>
          <div class="skeleton-item">
            <div class="skeleton-label shimmer"></div>
            <div class="skeleton-value shimmer"></div>
          </div>
          <div class="skeleton-item">
            <div class="skeleton-label shimmer"></div>
            <div class="skeleton-value shimmer"></div>
          </div>
        </div>
      </div>
      <div class="skeleton-footer">
        <div class="skeleton-price">
          <div class="skeleton-price-label shimmer"></div>
          <div class="skeleton-price-value shimmer"></div>
        </div>
        <div class="skeleton-button shimmer"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  count?: number;
}

withDefaults(defineProps<Props>(), {
  count: 3,
});
</script>

<style scoped lang="scss">
.skeleton-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(600px, 1fr));
  gap: 20px;
}

.skeleton-card {
  background: #fff;
  border: 1px solid var(--rider-border);
  border-radius: var(--rider-radius);
  padding: 24px;
  box-shadow: var(--rider-shadow);
}

.skeleton-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 20px;
}

.skeleton-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
}

.skeleton-header-info {
  flex: 1;
}

.skeleton-title {
  height: 20px;
  width: 60%;
  border-radius: 4px;
  margin-bottom: 8px;
}

.skeleton-subtitle {
  height: 16px;
  width: 40%;
  border-radius: 4px;
}

.skeleton-tag {
  height: 28px;
  width: 80px;
  border-radius: 14px;
}

.skeleton-body {
  background: var(--rider-primary-light);
  border-radius: var(--rider-radius);
  padding: 16px;
  margin-bottom: 16px;
}

.skeleton-section {
  margin-bottom: 16px;

  &:last-child {
    margin-bottom: 0;
  }
}

.skeleton-section-title {
  height: 16px;
  width: 80px;
  border-radius: 4px;
  margin-bottom: 12px;
}

.skeleton-item {
  display: flex;
  gap: 12px;
  margin-bottom: 8px;

  &:last-child {
    margin-bottom: 0;
  }
}

.skeleton-label {
  height: 14px;
  width: 50px;
  border-radius: 4px;
  flex-shrink: 0;
}

.skeleton-value {
  height: 14px;
  width: 70%;
  border-radius: 4px;
  flex: 1;
}

.skeleton-footer {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
}

.skeleton-price {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.skeleton-price-label {
  height: 14px;
  width: 60px;
  border-radius: 4px;
}

.skeleton-price-value {
  height: 28px;
  width: 100px;
  border-radius: 4px;
}

.skeleton-button {
  height: 36px;
  width: 120px;
  border-radius: 18px;
}

// 闪烁动画
.shimmer {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% {
    background-position: -200% 0;
  }
  100% {
    background-position: 200% 0;
  }
}

// 响应式
@media (max-width: 768px) {
  .skeleton-container {
    grid-template-columns: 1fr;
  }

  .skeleton-card {
    padding: 16px;
  }

  .skeleton-avatar {
    width: 50px;
    height: 50px;
  }

  .skeleton-footer {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .skeleton-button {
    width: 100%;
  }
}
</style>