<template>
  <div>
    <div class="header">
      <div>
        <div class="title">进行中</div>
        <div class="sub">按阶段查看：待取餐 / 派送中</div>
      </div>
      <el-button :loading="loading" type="primary" @click="load">刷新</el-button>
    </div>

    <div class="stats">
      <el-card class="stat" shadow="never">
        <div class="k">待取餐</div>
        <div class="v">{{ count3 }}</div>
      </el-card>
      <el-card class="stat" shadow="never">
        <div class="k">派送中</div>
        <div class="v">{{ count4 }}</div>
      </el-card>
    </div>

    <el-tabs v-model="tab" class="tabs">
      <el-tab-pane label="待取餐" name="3" />
      <el-tab-pane label="派送中" name="4" />
    </el-tabs>

    <el-empty v-if="!loading && filtered.length === 0" description="暂无订单" />

    <div class="grid" v-else>
      <RiderOrderCard
        v-for="o in filtered"
        :key="o.id"
        :order="o"
        mode="ongoing"
      >
        <template #actions>
          <el-button v-if="o.status === 3" type="primary" @click="pickup(o.id)">取货</el-button>
          <el-button v-else type="success" @click="deliver(o.id)">送达</el-button>
        </template>
      </RiderOrderCard>
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

onMounted(load);
</script>

<style scoped lang="scss">
.header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:12px; }
.title { font-size: 18px; font-weight: 800; }
.sub { margin-top: 4px; font-size: 12px; color:#909399; }

.stats {
  display: grid;
  grid-template-columns: 220px 220px;
  gap: 12px; 
  margin-bottom: 8px;
}
.stat { border-radius: 12px; border: 1px solid #ebeef5; }
.k { color:#909399; font-size: 12px; }
.v { font-size: 22px; font-weight: 900; margin-top: 6px; }

.tabs { margin-bottom: 8px; }

.grid { display: grid; grid-template-columns: 1fr; gap: 12px; }
</style>
