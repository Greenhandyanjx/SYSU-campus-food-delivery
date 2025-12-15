<template>
  <div>
    <div class="header">
      <div>
        <div class="title">历史订单</div>
        <div class="sub">只展示已完成</div>
      </div>
      <el-button :loading="loading" type="primary" @click="load">刷新</el-button>
    </div>

    <el-empty v-if="!loading && list.length === 0" description="暂无历史订单" />

    <div class="grid" v-else>
      <RiderOrderCard
        v-for="o in list"
        :key="o.id"
        :order="o"
        mode="history"
      >
        <template #actions>
          <el-button disabled>已完成</el-button>
        </template>
      </RiderOrderCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { riderApi, type RiderOrderItem } from "@/api/rider";
import RiderOrderCard from "@/components/rider/RiderOrderCard.vue";

const list = ref<RiderOrderItem[]>([]);
const loading = ref(false);

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
.header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:12px; }
.title { font-size: 18px; font-weight: 800; }
.sub { margin-top: 4px; font-size: 12px; color:#909399; }
.grid { display:grid; grid-template-columns: 1fr; gap:12px; }
</style>
