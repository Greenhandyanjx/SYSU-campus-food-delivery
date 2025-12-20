<template>
  <div class="orders-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="title-section">
          <h1 class="page-title">
            <span class="title-icon">🚚</span>
            进行中
          </h1>
          <p class="page-subtitle">查看并管理您的配送订单</p>
        </div>
        <div class="header-actions">
          <el-button :loading="loading" type="primary" size="large" @click="load">
            <i class="iconfont icon-refresh"></i>
            刷新状态
          </el-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-container" v-if="!loading">
      <div class="stat-card" :class="{ active: tab === '3' }" @click="tab = '3'">
        <div class="stat-icon">🏪</div>
        <div class="stat-content">
          <div class="stat-value">{{ count3 }}</div>
          <div class="stat-label">待取餐</div>
          <div class="stat-desc">商家已准备好</div>
        </div>
        <div class="stat-arrow" v-if="count3 > 0">
          <i class="el-icon-arrow-right"></i>
        </div>
      </div>
      <div class="stat-card" :class="{ active: tab === '4' }" @click="tab = '4'">
        <div class="stat-icon">🛵</div>
        <div class="stat-content">
          <div class="stat-value">{{ count4 }}</div>
          <div class="stat-label">派送中</div>
          <div class="stat-desc">正在配送途中</div>
        </div>
        <div class="stat-arrow" v-if="count4 > 0">
          <i class="el-icon-arrow-right"></i>
        </div>
      </div>
    </div>

    <!-- 标签页 -->
    <el-tabs v-model="tab" class="status-tabs" v-if="!loading">
      <el-tab-pane name="3">
        <template #label>
          <span class="tab-label">
            <span class="tab-icon">🏪</span>
            待取餐
            <span class="tab-badge" v-if="count3 > 0">{{ count3 }}</span>
          </span>
        </template>
      </el-tab-pane>
      <el-tab-pane name="4">
        <template #label>
          <span class="tab-label">
            <span class="tab-icon">🛵</span>
            派送中
            <span class="tab-badge" v-if="count4 > 0">{{ count4 }}</span>
          </span>
        </template>
      </el-tab-pane>
    </el-tabs>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <!-- 订单列表 -->
    <div class="orders-grid" v-else-if="filtered.length > 0">
      <TransitionGroup name="order-list" tag="div">
        <RiderOrderCard
          v-for="o in filtered"
          :key="o.id"
          :order="o"
          mode="ongoing"
          class="order-item"
          @open-chat="handleOpenChat"
        >
          <template #actions>
            <el-button v-if="o.status === 3" type="primary" size="large" @click="pickup(o.id)">
              <i class="iconfont icon-pickup"></i>
              确认取货
            </el-button>
            <el-button v-else type="success" size="large" @click="deliver(o.id)">
              <i class="iconfont icon-deliver"></i>
              确认送达
            </el-button>
          </template>
        </RiderOrderCard>
      </TransitionGroup>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-empty :description="tab === '3' ? '暂无待取餐订单' : '暂无派送中订单'" :image-size="180">
        <el-button type="primary" @click="load">刷新页面</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { riderApi, type RiderOrderItem } from "@/api/rider";
import RiderOrderCard from "@/components/rider/RiderOrderCard.vue";

const list = ref<RiderOrderItem[]>([]);
const loading = ref(false);
const tab = ref<"3" | "4">("3");

const count3 = computed(() => list.value.filter(x => x.status === 3).length);
const count4 = computed(() => list.value.filter(x => x.status === 4).length);

const filtered = computed(() => {
  const st = Number(tab.value);
  return list.value.filter(x => x.status === st);
});

const load = async () => {
  loading.value = true;
  try {
    const res = await riderApi.getOngoing();
    list.value = res.data.data || [];
    // 自动切到有数据的 tab
    if (tab.value === "3" && count3.value === 0 && count4.value > 0) tab.value = "4";
    if (tab.value === "4" && count4.value === 0 && count3.value > 0) tab.value = "3";
  } finally {
    loading.value = false;
  }
};

const pickup = async (id: number) => {
  try {
    await riderApi.pickupOrder(id);
    ElMessage.success("已取货，进入派送中");
    await load();
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "取货失败");
  }
};

const deliver = async (id: number) => {
  try {
    await riderApi.deliverOrder(id);
    ElMessage.success("已送达");
    await load();
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "送达失败");
  }
};

// 处理打开聊天事件
const handleOpenChat = (data: { type: 'user' | 'merchant'; id: number; name: string }) => {
  // 发送全局事件，聊天组件会监听这个事件
  window.dispatchEvent(new CustomEvent('rider:openChat', { detail: data }));
};

onMounted(load);
</script>

<style scoped lang="scss">
.orders-page {
  padding: 24px;
  background: var(--rider-bg);
  min-height: calc(100vh - 60px);
}

// 页面头部
.page-header {
  background: linear-gradient(135deg, var(--rider-primary) 0%, var(--rider-primary-dark) 100%);
  border-radius: var(--rider-radius);
  padding: 30px;
  margin-bottom: 24px;
  box-shadow: var(--rider-shadow);
  color: #fff;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
}

.title-section {
  .page-title {
    font-size: 28px;
    font-weight: 800;
    margin: 0 0 8px 0;
    display: flex;
    align-items: center;
    gap: 12px;

    .title-icon {
      font-size: 32px;
    }
  }

  .page-subtitle {
    font-size: 16px;
    opacity: 0.9;
    margin: 0;
  }
}

.header-actions {
  :deep(.el-button) {
    background: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.3);
    color: #fff;
    padding: 12px 24px;
    font-size: 15px;
    font-weight: 600;

    &:hover {
      background: rgba(255, 255, 255, 0.3);
      border-color: rgba(255, 255, 255, 0.5);
    }
  }
}

// 统计卡片
.stats-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: #fff;
  border: 2px solid var(--rider-border);
  border-radius: var(--rider-radius);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 20px;
  transition: all 0.3s ease;
  box-shadow: var(--rider-shadow);
  cursor: pointer;
  position: relative;
  overflow: hidden;

  &:hover {
    box-shadow: var(--rider-shadow-hover);
    transform: translateY(-2px);
  }

  &.active {
    border-color: var(--rider-primary);
    background: linear-gradient(135deg, rgba(255, 179, 2, 0.05) 0%, rgba(255, 179, 2, 0.02) 100%);

    .stat-icon {
      background: var(--rider-primary);
    }
  }

  .stat-icon {
    font-size: 36px;
    width: 70px;
    height: 70px;
    background: var(--rider-primary-light);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.3s ease;
  }

  .stat-content {
    flex: 1;

    .stat-value {
      font-size: 32px;
      font-weight: 800;
      color: var(--rider-text);
      margin-bottom: 4px;
    }

    .stat-label {
      font-size: 16px;
      color: var(--rider-text);
      font-weight: 600;
      margin-bottom: 4px;
    }

    .stat-desc {
      font-size: 13px;
      color: var(--rider-sub);
    }
  }

  .stat-arrow {
    font-size: 20px;
    color: var(--rider-primary);
    opacity: 0.8;
  }
}

// 标签页
.status-tabs {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 20px;
  margin-bottom: 24px;
  box-shadow: var(--rider-shadow);

  :deep(.el-tabs__header) {
    margin: 0;
  }

  :deep(.el-tabs__nav-wrap) {
    &::after {
      display: none;
    }
  }

  :deep(.el-tabs__item) {
    padding: 0 32px;
    font-size: 16px;
    font-weight: 600;
    color: var(--rider-sub);

    &.is-active {
      color: var(--rider-primary);
    }
  }

  :deep(.el-tabs__active-bar) {
    background-color: var(--rider-primary);
    height: 4px;
    border-radius: 2px;
  }
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;

  .tab-icon {
    font-size: 18px;
  }

  .tab-badge {
    background: var(--rider-primary);
    color: #fff;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 10px;
    font-weight: 600;
    min-width: 20px;
    text-align: center;
  }
}

// 加载状态
.loading-container {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 30px;
  box-shadow: var(--rider-shadow);
}

// 订单网格
.orders-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(600px, 1fr));
  gap: 20px;
  align-items: start;
}

// 空状态
.empty-state {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 60px;
  text-align: center;
  box-shadow: var(--rider-shadow);
}

// 过渡动画
.order-list-enter-active,
.order-list-leave-active {
  transition: all 0.5s ease;
}

.order-list-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.order-list-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

.order-list-move {
  transition: transform 0.5s ease;
}

// 响应式
@media (max-width: 768px) {
  .orders-page {
    padding: 16px;
  }

  .header-content {
    flex-direction: column;
    align-items: flex-start;
  }

  .page-header {
    padding: 20px;
  }

  .page-title {
    font-size: 24px !important;
  }

  .page-subtitle {
    font-size: 14px !important;
  }

  .stats-container {
    grid-template-columns: 1fr;
  }

  .status-tabs {
    padding: 16px;
  }

  .orders-grid {
    grid-template-columns: 1fr;
  }
}

// Icon font styles
.iconfont {
  font-family: "iconfont" !important;
  font-size: 14px;
  font-style: normal;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.icon-refresh:before { content: "🔄"; }
.icon-pickup:before { content: "📦"; }
.icon-deliver:before { content: "✅"; }
</style>
