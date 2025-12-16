<template>
  <div class="orders-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="title-section">
          <h1 class="page-title">
            <span class="title-icon">🔔</span>
            待接单
          </h1>
          <p class="page-subtitle">有新的订单等待您接单</p>
        </div>
        <div class="header-actions">
          <el-button :loading="loading" type="primary" size="large" @click="load">
            <i class="iconfont icon-refresh"></i>
            刷新订单
          </el-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-container" v-if="!loading">
      <div class="stat-card">
        <div class="stat-icon">⏰</div>
        <div class="stat-content">
          <div class="stat-value">{{ list.length }}</div>
          <div class="stat-label">待接单数量</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">💰</div>
        <div class="stat-content">
          <div class="stat-value">¥{{ totalFee }}</div>
          <div class="stat-label">预计总收益</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">📍</div>
        <div class="stat-content">
          <div class="stat-value">{{ uniqueAddresses }}</div>
          <div class="stat-label">配送地点数</div>
        </div>
      </div>
    </div>

    <!-- 错误提示 -->
    <el-alert v-if="err" :title="err" type="error" show-icon class="alert-container" />

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <!-- 订单列表 -->
    <div class="orders-grid" v-else-if="list.length > 0">
      <TransitionGroup name="order-list" tag="div">
        <RiderOrderCard
          v-for="o in list"
          :key="o.id"
          :order="o"
          mode="new"
          @accept="accept"
          class="order-item"
        />
      </TransitionGroup>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-empty description="暂无待接单订单" :image-size="180">
        <el-button type="primary" @click="load">刷新页面</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup lang="ts">
import RiderOrderCard from "@/components/rider/RiderOrderCard.vue";
import { onMounted, ref, computed } from "vue";
import { ElMessage } from "element-plus";
import { riderApi, type RiderOrderItem } from "@/api/rider";

const list = ref<RiderOrderItem[]>([]);
const loading = ref(false);
const err = ref("");

// 计算统计数据
const totalFee = computed(() => {
  return list.value.reduce((sum, order) => {
    const fee = Number(order.estimatedFee) || 0;
    return sum + fee;
  }, 0).toFixed(2);
});

const uniqueAddresses = computed(() => {
  const addresses = new Set();
  list.value.forEach(order => {
    if (order.deliveryAddress) {
      addresses.add(order.deliveryAddress);
    }
  });
  return addresses.size;
});

const fmt = (s: string) => {
  if (!s) return "";
  const d = new Date(s);
  return isNaN(d.getTime()) ? s : d.toLocaleString();
};

const load = async () => {
  loading.value = true;
  err.value = "";
  try {
    const res = await riderApi.getNewOrders();
    list.value = res.data.data || [];
  } catch (e: any) {
    err.value = e?.response?.data?.msg || e?.message || "请求失败";
  } finally {
    loading.value = false;
  }
};

const accept = async (id: number) => {
  try {
    await riderApi.acceptOrder(id);
    ElMessage.success("接单成功");
    await load();
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || "接单失败");
  }
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
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: #fff;
  border: 1px solid var(--rider-border);
  border-radius: var(--rider-radius);
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s ease;
  box-shadow: var(--rider-shadow);

  &:hover {
    box-shadow: var(--rider-shadow-hover);
    transform: translateY(-2px);
  }

  .stat-icon {
    font-size: 32px;
    width: 60px;
    height: 60px;
    background: var(--rider-primary-light);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .stat-content {
    .stat-value {
      font-size: 24px;
      font-weight: 700;
      color: var(--rider-text);
      margin-bottom: 4px;
    }

    .stat-label {
      font-size: 14px;
      color: var(--rider-sub);
      font-weight: 500;
    }
  }
}

// 提示容器
.alert-container {
  margin-bottom: 20px;
  border-radius: var(--rider-radius);
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
</style>
