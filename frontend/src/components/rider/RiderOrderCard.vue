<template>
  <el-card class="order-card" shadow="never">
    <!-- 顶部：店名 + 订单号 + 状态 -->
    <div class="top">
      <div class="left">
        <div class="shop">
          <span class="shop-icon">🏪</span>
          {{ order.restaurant || "（无店名）" }}
        </div>
        <div class="sub">
          <span class="order-id">📋 订单 #{{ order.id }}</span>
          <span class="dot">·</span>
          <span class="time">🕐 下单 {{ fmt(order.createdAt) }}</span>
        </div>
      </div>

      <div class="right">
        <el-tag :type="tagType" effect="light" class="status-tag">
          <span class="status-icon">{{ statusIcon }}</span>
          {{ tagText }}
        </el-tag>
      </div>
    </div>

    <!-- 中部：信息块（密度更合理） -->
    <div class="body">
      <div class="info-section">
        <div class="section-header">
          <span class="section-icon">📍</span>
          <span class="section-title">取餐信息</span>
        </div>
        <div class="info-item">
          <span class="info-label">商家</span>
          <span class="info-value">{{ order.restaurant || "-" }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">地址</span>
          <span class="info-value">{{ order.pickupAddress || "-" }}</span>
        </div>
      </div>

      <div class="divider"></div>

      <div class="info-section">
        <div class="section-header">
          <span class="section-icon">🏠</span>
          <span class="section-title">配送信息</span>
        </div>
        <div class="info-item">
          <span class="info-label">收货人</span>
          <span class="info-value">{{ order.customer || "-" }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">地址</span>
          <span class="info-value">{{ order.deliveryAddress || "-" }}</span>
        </div>
      </div>
    </div>

    <!-- 底部：费用/预计/时间线 + 操作 -->
    <div class="bottom">
      <div class="bottom-left">
        <div class="price-section">
          <span class="price-label">配送费</span>
          <span class="price-value">¥{{ money(order.estimatedFee) }}</span>
        </div>
        <div class="timeline-section">
          <div class="timeline-item" v-if="order.acceptedAt">
            <span class="timeline-icon">✅</span>
            <div class="timeline-content">
              <span class="timeline-time">{{ fmtShort(order.acceptedAt) }}</span>
              <span class="timeline-label">已接单</span>
            </div>
          </div>
          <div class="timeline-item" v-if="order.pickupAt">
            <span class="timeline-icon">📦</span>
            <div class="timeline-content">
              <span class="timeline-time">{{ fmtShort(order.pickupAt) }}</span>
              <span class="timeline-label">已取货</span>
            </div>
          </div>
          <div class="timeline-item" v-if="order.deliverAt">
            <span class="timeline-icon">🚚</span>
            <div class="timeline-content">
              <span class="timeline-time">{{ fmtShort(order.deliverAt) }}</span>
              <span class="timeline-label">配送中</span>
            </div>
          </div>
          <div class="timeline-item" v-if="order.finishAt">
            <span class="timeline-icon">🎯</span>
            <div class="timeline-content">
              <span class="timeline-time">{{ fmtShort(order.finishAt) }}</span>
              <span class="timeline-label">已完成</span>
            </div>
          </div>
        </div>
      </div>

      <div class="bottom-right">
        <div class="estimated-time" v-if="order.estimatedTime">
          <i class="iconfont icon-clock"></i>
          <span>预计 {{ order.estimatedTime }} 分钟</span>
        </div>
        <div class="actions">
          <slot name="actions" />
          <!-- 联系商家按钮 -->
          <el-button
            v-if="props.mode === 'ongoing'"
            size="small"
            type="info"
            plain
            @click="openChat"
          >
            <i class="iconfont icon-message"></i>
            联系商家
          </el-button>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { RiderOrderItem } from "@/api/rider";

const emit = defineEmits<{
  openChat: [data: { type: 'user' | 'merchant'; id: number; name: string }]
}>()

const props = defineProps<{
  order: RiderOrderItem;
  mode?: "new" | "ongoing" | "history";
}>();

const fmt = (s?: string | null) => {
  if (!s) return "-";
  const d = new Date(s);
  return isNaN(d.getTime()) ? String(s) : d.toLocaleString();
};

const fmtShort = (s?: string | null) => {
  if (!s) return "-";
  const d = new Date(s);
  if (isNaN(d.getTime())) return String(s);

  const today = new Date();
  const isToday = d.toDateString() === today.toDateString();

  if (isToday) {
    return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  } else {
    return d.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
  }
};

const money = (n: any) => {
  const x = Number(n);
  return Number.isFinite(x) ? x.toFixed(2).replace(/\.00$/, "") : (n ?? "-");
};

const tagText = computed(() => {
  const st = props.order.status;
  if (props.mode === "new") return "待接单";
  if (props.mode === "history") return "已完成";
  // ongoing: 3/4
  if (st === 3) return "待取餐";
  if (st === 4) return "派送中";
  return `状态 ${st}`;
});

const tagType = computed(() => {
  const st = props.order.status;
  if (props.mode === "new") return "warning";
  if (props.mode === "history") return "success";
  if (st === 3) return "info";
  if (st === 4) return "primary";
  return "default";
});

const statusIcon = computed(() => {
  const st = props.order.status;
  if (props.mode === "new") return "⏰";
  if (props.mode === "history") return "✅";
  if (st === 3) return "🏪";
  if (st === 4) return "🚚";
  return "📋";
});

// 打开与商家的聊天
const openChat = () => {
  emit('openChat', {
    type: 'merchant',
    id: props.order.merchantId,
    name: props.order.restaurant || '商家'
  })
}

// 打开高德地图导航
const openNavigation = (destination: string, address: string) => {
  // 使用高德地图URL Scheme打开导航
  // 格式：https://uri.amap.com/navigation?to=目的地名称,经度,纬度,地址
  // 这里没有经纬度，所以只用目的地名称和地址
  const navUrl = `https://uri.amap.com/navigation?to=${encodeURIComponent(destination)},,${encodeURIComponent(address)}&mode=car&coordinate=gaode`;

  // 在新窗口打开导航
  window.open(navUrl, '_blank');
}
</script>

<style scoped lang="scss">
.order-card {
  border-radius: var(--rider-radius);
  border: 1px solid var(--rider-border);
  transition: all 0.3s ease;
  box-shadow: var(--rider-shadow);

  &:hover {
    box-shadow: var(--rider-shadow-hover);
    transform: translateY(-2px);
  }

  :deep(.el-card__body) {
    padding: 20px;
  }
}

.top {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--rider-border);
}

.shop {
  font-size: 18px;
  font-weight: 800;
  color: var(--rider-text);
  display: flex;
  align-items: center;
  gap: 8px;
}

.shop-icon {
  font-size: 20px;
}

.sub {
  margin-top: 8px;
  font-size: 13px;
  color: var(--rider-sub);
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.order-id, .time {
  display: flex;
  align-items: center;
  gap: 4px;
}

.dot {
  opacity: 0.6;
  color: var(--rider-primary);
}

.status-tag {
  border-radius: 20px;
  padding: 6px 16px;
  font-weight: 600;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-icon {
  font-size: 16px;
}

.body {
  background: linear-gradient(135deg, var(--rider-primary-light) 0%, #fff 100%);
  border: 1px solid rgba(255, 179, 2, 0.1);
  border-radius: var(--rider-radius);
  padding: 20px;
  margin-bottom: 16px;
}

.info-section {
  margin-bottom: 16px;

  &:last-child {
    margin-bottom: 0;
  }
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;

  .section-icon {
    font-size: 18px;
  }

  .section-title {
    font-size: 14px;
    font-weight: 700;
    color: var(--rider-text);
  }
}

.info-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 8px;

  &:last-child {
    margin-bottom: 0;
  }

  .info-label {
    font-size: 13px;
    color: var(--rider-sub);
    font-weight: 600;
    min-width: 50px;
    flex-shrink: 0;
  }

  .info-value {
    font-size: 14px;
    color: var(--rider-text);
    font-weight: 500;
    line-height: 1.5;
    flex: 1;
  }
}

.divider {
  height: 1px;
  background: rgba(255, 179, 2, 0.2);
  margin: 16px 0;
}

.bottom {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 24px;
}

.bottom-left {
  flex: 1;
}

.price-section {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;

  .price-label {
    font-size: 14px;
    color: var(--rider-sub);
    font-weight: 600;
  }

  .price-value {
    font-size: 28px;
    font-weight: 800;
    color: var(--rider-primary);
  }
}

.timeline-section {
  display: flex;
  gap: 24px;
}

.timeline-item {
  display: flex;
  align-items: center;
  gap: 8px;

  .timeline-icon {
    font-size: 18px;
    flex-shrink: 0;
  }

  .timeline-content {
    display: flex;
    flex-direction: column;
    gap: 2px;

    .timeline-time {
      font-size: 13px;
      font-weight: 600;
      color: var(--rider-text);
    }

    .timeline-label {
      font-size: 12px;
      color: var(--rider-sub);
    }
  }
}

.bottom-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
}

.estimated-time {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--rider-sub);
  font-weight: 600;
  background: var(--rider-primary-light);
  padding: 6px 12px;
  border-radius: 20px;

  .iconfont {
    font-size: 14px;
  }
}

.actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

:deep(.el-button) {
  border-radius: 20px;
  padding: 8px 20px;
  font-weight: 600;
}

:deep(.el-button--primary) {
  background: var(--rider-primary);
  border-color: var(--rider-primary);

  &:hover {
    background: var(--rider-primary-dark);
    border-color: var(--rider-primary-dark);
  }
}

:deep(.el-button--success) {
  background: #67C23A;
  border-color: #67C23A;

  &:hover {
    background: #85CE61;
    border-color: #85CE61;
  }
}

:deep(.el-button--warning) {
  background: #E6A23C;
  border-color: #E6A23C;

  &:hover {
    background: #EEBE77;
    border-color: #EEBE77;
  }
}

:deep(.el-tag--light) {
  border-radius: 16px;
  font-weight: 600;
}

// 响应式设计
@media (max-width: 768px) {
  .order-card {
    :deep(.el-card__body) {
      padding: 16px;
    }
  }

  .shop {
    font-size: 16px !important;
  }

  .shop-icon {
    font-size: 18px !important;
  }

  .body {
    padding: 16px;
  }

  .section-header {
    margin-bottom: 8px;
  }

  .info-item {
    gap: 8px;
  }

  .timeline-section {
    flex-wrap: wrap;
    gap: 16px;
  }

  .bottom {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }

  .bottom-right {
    align-items: flex-start;
  }

  .price-value {
    font-size: 24px !important;
  }

  .actions {
    justify-content: stretch;

    :deep(.el-button) {
      flex: 1;
    }
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

.icon-clock:before { content: "⏱️"; }
.icon-message:before { content: "💬"; }
</style>
