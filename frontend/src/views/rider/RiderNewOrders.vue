<template>
  <div>
    <div class="header">
      <div class="title">待接单</div>
      <div class="right">
        <el-button :loading="loading" type="primary" @click="load">刷新</el-button>
      </div>
    </div>

    <el-alert v-if="err" :title="err" type="error" show-icon class="mb" />

    <el-empty v-if="!loading && list.length === 0" description="暂无待接单订单" />

  <div class="grid">
    <RiderOrderCard
      v-for="o in list"
      :key="o.id"
      :order="o"
      mode="new"
      @accept="accept"
    />
  </div>

  </div>
</template>

<script setup lang="ts">
import RiderOrderCard from "@/components/rider/RiderOrderCard.vue";
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { riderApi, type RiderOrderItem } from "@/api/rider";

const list = ref<RiderOrderItem[]>([]);
const loading = ref(false);
const err = ref("");

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
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.title { font-size: 18px; font-weight: 700; }
.mb { margin-bottom: 12px; }

.grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}
.card { padding: 2px; }
.top { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom: 10px; }
.shop { font-size: 16px; font-weight: 800; color: var(--rider-text); }
.id { margin-top: 6px; font-size: 12px; color: var(--rider-sub); }

.info {
  background: #fafcff;
  border: 1px dashed #dbe7ff;
  border-radius: 12px;
  padding: 10px 12px;
}
.row { display:flex; gap: 10px; margin: 6px 0; }
.row span { width: 54px; color: var(--rider-sub); font-size: 12px; }
.row b { flex: 1; font-weight: 600; color: #303133; }

.meta {
  margin-top: 12px;
  display:flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
  color: #606266;
}

.actions { margin-top: 12px; display:flex; justify-content:flex-end; }

</style>
