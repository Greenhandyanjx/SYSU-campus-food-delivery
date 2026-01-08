<template>
  <div class="orders-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="title-section">
          <h1 class="page-title">
            <span class="title-icon">📋</span>
            历史订单
          </h1>
          <p class="page-subtitle">查看您已完成的所有订单</p>
        </div>
        <div class="header-actions">
          <el-button :loading="loading" type="primary" size="large" @click="load">
            <i class="iconfont icon-refresh"></i>
            刷新记录
          </el-button>
        </div>
      </div>
    </div>

    <!-- 统计卡片区域 -->
    <div class="stats-container" v-if="!loading && list.length > 0">
      <div class="stat-card">
        <div class="stat-icon">📦</div>
        <div class="stat-content">
          <div class="stat-value">{{ list.length }}</div>
          <div class="stat-label">总订单数</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">💰</div>
        <div class="stat-content">
          <div class="stat-value">¥{{ totalEarnings }}</div>
          <div class="stat-label">累计收益</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">⭐</div>
        <div class="stat-content">
          <div class="stat-value">{{ averageRating }}</div>
          <div class="stat-label">平均评分</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">⏱️</div>
        <div class="stat-content">
          <div class="stat-value">{{ averageTime }}分钟</div>
          <div class="stat-label">平均配送时间</div>
        </div>
      </div>
    </div>

    <!-- 筛选和排序 -->
    <div class="filter-bar" v-if="!loading && list.length > 0">
      <div class="filter-left">
        <el-select v-model="filterMonth" placeholder="选择月份" clearable>
          <el-option label="全部" value="" />
          <el-option v-for="month in availableMonths" :key="month" :label="month" :value="month" />
        </el-select>
      </div>
      <div class="filter-right">
        <el-radio-group v-model="sortBy" size="large">
          <el-radio-button label="date">按时间</el-radio-button>
          <el-radio-button label="fee">按金额</el-radio-button>
          <el-radio-button label="rating">按评分</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>

    <!-- 订单列表 -->
    <div class="orders-grid" v-else-if="filteredOrders.length > 0">
      <TransitionGroup name="order-list" tag="div">
        <RiderOrderCard
          v-for="o in filteredOrders"
          :key="o.id"
          :order="o"
          mode="history"
          class="order-item"
        >
          <template #actions>
            <el-tag type="success" effect="light" size="large">
              <i class="iconfont icon-check"></i>
              已完成
            </el-tag>
          </template>
        </RiderOrderCard>
      </TransitionGroup>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <el-empty description="暂无历史订单" :image-size="180">
        <template #description>
          <p>您还没有完成任何订单</p>
          <p class="empty-tip">快去接单吧！</p>
        </template>
        <el-button type="primary" @click="$router.push('/rider/new')">去接单</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { useRouter } from "vue-router";
import { riderApi, type RiderOrderItem } from "@/api/rider";
import RiderOrderCard from "@/components/rider/RiderOrderCard.vue";

const router = useRouter();
const list = ref<RiderOrderItem[]>([]);
const loading = ref(false);
const filterMonth = ref("");
const sortBy = ref("date");

// 计算统计数据
const totalEarnings = computed(() => {
  return list.value.reduce((sum, order) => {
    const fee = Number(order.estimatedFee) || 0;
    return sum + fee;
  }, 0).toFixed(2);
});

const averageRating = computed(() => {
  const validRatings = list.value.filter(o => o.rating && o.rating > 0);
  if (validRatings.length === 0) return "0.0";
  const sum = validRatings.reduce((sum, o) => sum + (o.rating || 0), 0);
  return (sum / validRatings.length).toFixed(1);
});

const averageTime = computed(() => {
  const times: number[] = [];
  list.value.forEach(order => {
    if (order.acceptedAt && order.finishAt) {
      const start = new Date(order.acceptedAt).getTime();
      const end = new Date(order.finishAt).getTime();
      const minutes = Math.round((end - start) / 60000);
      if (minutes > 0 && minutes < 300) { // 过滤掉异常数据
        times.push(minutes);
      }
    }
  });
  if (times.length === 0) return 0;
  return Math.round(times.reduce((a, b) => a + b, 0) / times.length);
});

const availableMonths = computed(() => {
  const months = new Set<string>();
  list.value.forEach(order => {
    if (order.finishAt) {
      const date = new Date(order.finishAt);
      const month = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
      months.add(month);
    }
  });
  return Array.from(months).sort().reverse();
});

const filteredOrders = computed(() => {
  let filtered = [...list.value];

  // 月份筛选
  if (filterMonth.value) {
    filtered = filtered.filter(order => {
      if (!order.finishAt) return false;
      const date = new Date(order.finishAt);
      const month = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
      return month === filterMonth.value;
    });
  }

  // 排序
  filtered.sort((a, b) => {
    if (sortBy.value === "date") {
      return new Date(b.finishAt || 0).getTime() - new Date(a.finishAt || 0).getTime();
    } else if (sortBy.value === "fee") {
      return Number(b.estimatedFee || 0) - Number(a.estimatedFee || 0);
    } else if (sortBy.value === "rating") {
      return (b.rating || 0) - (a.rating || 0);
    }
    return 0;
  });

  return filtered;
});

const load = async () => {
  loading.value = true;
  try {
    const res = await riderApi.getHistory();
    list.value = res.data.data || [];
  } finally {
    loading.value = false;
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
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
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

// 筛选栏
.filter-bar {
  background: #fff;
  border-radius: var(--rider-radius);
  padding: 16px 24px;
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: var(--rider-shadow);
  gap: 16px;

  .filter-left {
    :deep(.el-select) {
      width: 160px;
    }
  }

  .filter-right {
    :deep(.el-radio-group) {
      .el-radio-button__inner {
        border-radius: 20px;
        padding: 10px 20px;
        font-weight: 600;
      }
    }
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

  .empty-tip {
    color: var(--rider-sub);
    font-size: 14px;
    margin-top: 8px;
  }
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
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .stat-card {
    padding: 16px;
  }

  .stat-icon {
    width: 50px !important;
    height: 50px !important;
    font-size: 28px !important;
  }

  .stat-value {
    font-size: 20px !important;
  }

  .filter-bar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .filter-left {
    :deep(.el-select) {
      width: 100%;
    }
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
.icon-check:before { content: "✅"; }
</style>
