<template>
  <div class="wrap">
    <div class="title">
      <div class="h1">工作台</div>
      <div class="sub">骑手个人与状态概览</div>
    </div>

    <el-row :gutter="16" class="row">
      <el-col :span="6">
        <div class="card">
          <div class="k">当前账号</div>
          <div class="v">{{ me?.name || username }}</div>
        </div>
      </el-col>

      <el-col :span="6">
        <div class="card">
          <div class="k">在线状态</div>
          <div class="v">
            <el-tag :type="me?.isOnline ? 'success' : 'info'">
              {{ me?.isOnline ? "在线" : "离线" }}
            </el-tag>
          </div>
        </div>
      </el-col>

      <el-col :span="6">
        <div class="card">
          <div class="k">已完成</div>
          <div class="v">{{ me?.completedOrders ?? 0 }}</div>
        </div>
      </el-col>

      <el-col :span="6">
        <div class="card">
          <div class="k">评分</div>
          <div class="v">{{ me?.rating ?? 0 }}</div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="row">
      <el-col :span="8">
        <div class="card">
          <div class="k">待接单</div>
          <div class="v">{{ stat.newCount }}</div>
          <el-button class="btn" type="primary" plain @click="go('/rider/new')">去接单</el-button>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="card">
          <div class="k">进行中</div>
          <div class="v">{{ stat.ongoingCount }}</div>
          <el-button class="btn" type="primary" plain @click="go('/rider/ongoing')">查看</el-button>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="card">
          <div class="k">历史订单</div>
          <div class="v">{{ stat.historyCount }}</div>
          <el-button class="btn" type="primary" plain @click="go('/rider/history')">查看</el-button>
        </div>
      </el-col>
    </el-row>

    <div class="actions">
      <el-button :loading="loading" type="primary" @click="refreshAll">刷新数据</el-button>
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
.wrap { width: 100%; }
.title { margin-bottom: 14px; }
.h1 { font-size: 20px; font-weight: 700; }
.sub { font-size: 12px; color: #909399; margin-top: 4px; }
.row { margin-top: 12px; }

.card {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 12px;
  padding: 14px;
  min-height: 92px;
}
.k { font-size: 12px; color: #909399; margin-bottom: 8px; }
.v { font-size: 20px; font-weight: 700; }
.btn { margin-top: 10px; }

.actions { margin-top: 16px; }
</style>
