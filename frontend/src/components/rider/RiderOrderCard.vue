<template>
  <el-card class="order-card" shadow="never">
    <!-- 顶部：店名 + 订单号 + 状态 -->
    <div class="top">
      <div class="left">
        <div class="shop">{{ order.restaurant || "（无店名）" }}</div>
        <div class="sub">
          <span>订单 #{{ order.id }}</span>
          <span class="dot">·</span>
          <span>下单 {{ fmt(order.createdAt) }}</span>
        </div>
      </div>

      <div class="right">
        <el-tag :type="tagType" effect="light">{{ tagText }}</el-tag>
      </div>
    </div>

    <!-- 中部：信息块（密度更合理） -->
    <div class="body">
      <div class="block">
        <div class="k">取餐点</div>
        <div class="v">{{ order.pickupAddress || "-" }}</div>
      </div>
      <div class="block">
        <div class="k">收货人</div>
        <div class="v">{{ order.customer || "-" }}</div>
      </div>
      <div class="block span2">
        <div class="k">送达地址</div>
        <div class="v">{{ order.deliveryAddress || "-" }}</div>
      </div>
    </div>

    <!-- 底部：费用/预计/时间线 + 操作 -->
    <div class="bottom">
      <div class="meta">
        <div class="meta-item">
          <div class="k">配送费</div>
          <div class="v">¥{{ money(order.estimatedFee) }}</div>
        </div>
        <div class="meta-item">
          <div class="k">预计</div>
          <div class="v">{{ order.estimatedTime ?? "-" }} 分钟</div>
        </div>

        <!-- 时间线：有就显示，没有就不占位 -->
        <div class="meta-item" v-if="order.acceptedAt">
          <div class="k">接单</div>
          <div class="v">{{ fmt(order.acceptedAt) }}</div>
        </div>
        <div class="meta-item" v-if="order.pickupAt">
          <div class="k">取货</div>
          <div class="v">{{ fmt(order.pickupAt) }}</div>
        </div>
        <div class="meta-item" v-if="order.finishAt">
          <div class="k">完成</div>
          <div class="v">{{ fmt(order.finishAt) }}</div>
        </div>
      </div>

      <div class="actions">
        <slot name="actions" />
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { RiderOrderItem } from "@/api/rider";

const props = defineProps<{
  order: RiderOrderItem;
  mode?: "new" | "ongoing" | "history";
}>();

const fmt = (s?: string | null) => {
  if (!s) return "-";
  const d = new Date(s);
  return isNaN(d.getTime()) ? String(s) : d.toLocaleString();
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
</script>

<style scoped lang="scss">
.order-card {
  border-radius: 14px;
  border: 1px solid #ebeef5;
}

.top {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}
.shop {
  font-size: 16px;
  font-weight: 800;
  color: #303133;
}
.sub {
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 8px;
}
.dot { opacity: .6; }

.body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 14px;
  background: #fafcff;
  border: 1px dashed #dbe7ff;
  border-radius: 12px;
  padding: 12px;
}
.block .k {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
}
.block .v {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  line-height: 1.4;
  word-break: break-all;
}
.span2 { grid-column: span 2; }

.bottom {
  margin-top: 14px;
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 12px;
  flex-wrap: wrap;
}

.meta {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
}
.meta-item .k {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}
.meta-item .v {
  font-weight: 800;
  color: #303133;
}

.actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}
</style>
