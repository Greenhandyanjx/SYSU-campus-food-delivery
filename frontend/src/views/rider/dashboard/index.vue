<template>
  <div class="wrap">
    <div class="title">
      <div class="h1">工作台</div>
      <div class="sub">骑手个人与状态概览</div>
    </div>

    <el-row :gutter="16" class="row">
      <el-col :span="6">
        <div class="card">
          <div class="card-icon">👤</div>
          <div class="card-content">
            <div class="k">当前账号</div>
            <div class="v">{{ me?.name || username }}</div>
          </div>
        </div>
      </el-col>

      <el-col :span="6">
        <div class="card">
          <div class="card-icon" :class="{ 'online': me?.isOnline }">{{ me?.isOnline ? '🟢' : '🔴' }}</div>
          <div class="card-content">
            <div class="k">在线状态</div>
            <div class="v">
              <el-tag :type="me?.isOnline ? 'success' : 'info'" effect="light">
                {{ me?.isOnline ? "在线" : "离线" }}
              </el-tag>
            </div>
          </div>
        </div>
      </el-col>

      <el-col :span="6">
        <div class="card">
          <div class="card-icon">✅</div>
          <div class="card-content">
            <div class="k">已完成</div>
            <div class="v">{{ me?.completedOrders ?? 0 }}</div>
          </div>
        </div>
      </el-col>

      <el-col :span="6">
        <div class="card">
          <div class="card-icon">⭐</div>
          <div class="card-content">
            <div class="k">评分</div>
            <div class="v">{{ me?.rating ?? 0 }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="row">
      <el-col :span="8">
        <div class="card action-card">
          <div class="card-icon">🔔</div>
          <div class="card-content">
            <div class="k">待接单</div>
            <div class="v">{{ stat.newCount }}</div>
            <el-button class="btn" type="primary" @click="go('/rider/new')">
              <i class="iconfont icon-notification"></i>去接单
            </el-button>
          </div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="card action-card">
          <div class="card-icon">🚚</div>
          <div class="card-content">
            <div class="k">进行中</div>
            <div class="v">{{ stat.ongoingCount }}</div>
            <el-button class="btn" type="primary" @click="go('/rider/ongoing')">
              <i class="iconfont icon-truck"></i>查看
            </el-button>
          </div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="card action-card">
          <div class="card-icon">📋</div>
          <div class="card-content">
            <div class="k">历史订单</div>
            <div class="v">{{ stat.historyCount }}</div>
            <el-button class="btn" type="primary" @click="go('/rider/history')">
              <i class="iconfont icon-history"></i>查看
            </el-button>
          </div>
        </div>
      </el-col>
    </el-row>

    <div class="actions">
      <el-button :loading="loading" type="primary" @click="refreshAll" class="refresh-btn">
        <i class="iconfont icon-refresh" :class="{ 'loading-animation': loading }"></i>
        刷新数据
      </el-button>
    </div>

    <!-- 快捷操作浮窗 -->
    <div class="quick-actions" v-if="!loading">
      <div class="quick-action-item" @click="go('/rider/new')">
        <span class="quick-icon">🔔</span>
        <span class="quick-label">待接单</span>
        <span class="quick-badge" v-if="stat.newCount > 0">{{ stat.newCount }}</span>
      </div>
      <div class="quick-action-item" @click="go('/rider/ongoing')">
        <span class="quick-icon">🚚</span>
        <span class="quick-label">进行中</span>
      </div>
      <div class="quick-action-item" @click="go('/rider/history')">
        <span class="quick-icon">📋</span>
        <span class="quick-label">历史订单</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { riderApi, type RiderMe } from "@/api/rider";

const router = useRouter();
const username = localStorage.getItem("username") || "rider";

const me = ref<RiderMe | null>(null);
const loading = ref(false);

const stat = ref({
  newCount: 0,
  ongoingCount: 0,
  historyCount: 0,
});

const go = (path: string) => router.push(path);

const refreshAll = async () => {
  loading.value = true;
  try {
    const [meRes, newRes, ongoingRes, hisRes] = await Promise.all([
      riderApi.getMe(),
      riderApi.getNewOrders(),
      riderApi.getOngoing(),
      riderApi.getHistory(),
    ]);

    me.value = meRes.data.data;
    stat.value.newCount = (newRes.data.data || []).length;
    stat.value.ongoingCount = (ongoingRes.data.data || []).length;
    stat.value.historyCount = (hisRes.data.data || []).length;
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  refreshAll();
});
</script>

<style scoped lang="scss">
.wrap {
  width: 100%;
  padding: 20px;
  background: var(--rider-bg);
  min-height: calc(100vh - 60px);
}

.title {
  margin-bottom: 24px;
  .h1 {
    font-size: 24px;
    font-weight: 700;
    color: var(--rider-text);
    display: flex;
    align-items: center;
    gap: 10px;

    &::before {
      content: "🏠";
      font-size: 28px;
    }
  }
  .sub {
    font-size: 14px;
    color: var(--rider-sub);
    margin-top: 6px;
  }
}

.row {
  margin-top: 16px;
}

.card {
  background: #fff;
  border: 1px solid var(--rider-border);
  border-radius: var(--rider-radius);
  padding: 20px;
  min-height: 100px;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  transition: all 0.3s ease;
  box-shadow: var(--rider-shadow);

  &:hover {
    box-shadow: var(--rider-shadow-hover);
    transform: translateY(-2px);
  }
}

.card-icon {
  font-size: 32px;
  flex-shrink: 0;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--rider-primary-light);
  border-radius: 50%;
  transition: all 0.3s ease;

  &.online {
    animation: pulse 2s infinite;
  }
}

@keyframes pulse {
  0% {
    box-shadow: 0 0 0 0 rgba(76, 175, 80, 0.7);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(76, 175, 80, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(76, 175, 80, 0);
  }
}

.card-content {
  flex: 1;
}

.k {
  font-size: 14px;
  color: var(--rider-sub);
  margin-bottom: 8px;
  font-weight: 600;
}

.v {
  font-size: 24px;
  font-weight: 700;
  color: var(--rider-text);
}

.action-card {
  .card-content {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }
}

.btn {
  margin-top: 12px;
  border-radius: 20px;
  padding: 8px 20px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;

  i {
    font-size: 14px;
  }
}

.actions {
  margin-top: 24px;
  text-align: center;

  .el-button {
    padding: 12px 32px;
    border-radius: 24px;
    font-weight: 600;
    font-size: 15px;
  }
}

:deep(.el-button--primary) {
  background: var(--rider-primary);
  border-color: var(--rider-primary);

  &:hover {
    background: var(--rider-primary-dark);
    border-color: var(--rider-primary-dark);
  }
}

:deep(.el-tag--light) {
  border-radius: 16px;
  font-weight: 600;
  padding: 4px 12px;
}

// Icon font styles
.iconfont {
  font-family: "iconfont" !important;
  font-size: 14px;
  font-style: normal;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

// 快捷操作浮窗
.quick-actions {
  position: fixed;
  right: 24px;
  bottom: 24px;
  background: #fff;
  border-radius: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  padding: 16px;
  display: flex;
  gap: 16px;
  z-index: 100;
  animation: slideUp 0.5s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.quick-action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;

  &:hover {
    background: var(--rider-primary-light);
    transform: translateY(-2px);
  }

  &:active {
    transform: scale(0.95);
  }
}

.quick-icon {
  font-size: 24px;
  display: block;
}

.quick-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--rider-text);
}

.quick-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  background: var(--rider-primary);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
  animation: bounce-in 0.5s ease-out;
}

.icon-notification:before { content: "🔔"; }
.icon-truck:before { content: "🚚"; }
.icon-history:before { content: "📋"; }
.icon-refresh:before { content: "🔄"; }

// 响应式
@media (max-width: 768px) {
  .quick-actions {
    left: 50%;
    right: auto;
    transform: translateX(-50%);
    bottom: 16px;
    gap: 12px;
    padding: 12px;
  }

  .quick-action-item {
    padding: 8px;
  }

  .quick-icon {
    font-size: 20px;
  }

  .quick-label {
    font-size: 11px;
  }
}
</style>
