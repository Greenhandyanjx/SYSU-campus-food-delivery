<template>
  <div class="me-page">
    <div class="header">
      <div class="title">
        <span class="title-icon">👤</span>
        个人中心
      </div>
      <div class="header-actions">
        <el-button :loading="loading" type="primary" @click="loadAll">
          <i class="iconfont icon-refresh"></i>刷新
        </el-button>
      </div>
    </div>

    <!-- 基本资料 + 在线 -->
    <el-card class="card profile-card" shadow="never">
      <div class="profile-header">
        <div class="avatar-wrapper">
          <div class="avatar">
            <span class="avatar-emoji">🛵</span>
          </div>
        </div>
        <div class="profile-info">
          <div class="name">{{ me?.name || username }}</div>
          <div class="sub">
            <span class="info-item">📱 {{ me?.phone || "-" }}</span>
            <span class="divider">|</span>
            <span class="info-item">✅ 已完成：{{ me?.completedOrders ?? 0 }}</span>
            <span class="divider">|</span>
            <span class="info-item">⭐ 评分：{{ me?.rating ?? 0 }}</span>
          </div>
        </div>
        <div class="profile-actions">
          <div class="online-switch">
            <span class="status-label">在线状态</span>
            <el-switch v-model="online" @change="saveOnline" size="large" />
          </div>
          <el-button type="danger" @click="logout" class="logout-btn">
            <i class="iconfont icon-logout"></i>退出登录
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 钱包概览 -->
    <div class="section-title">
      <span class="section-icon">💰</span>
      钱包概览
    </div>
    <el-row :gutter="16" class="wallet-row">
      <el-col :span="8">
        <div class="stat-card">
          <div class="card-icon">💵</div>
          <div class="card-content">
            <div class="k">可用余额</div>
            <div class="v">¥ {{ money(wallet?.balance) }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card">
          <div class="card-icon">🧊</div>
          <div class="card-content">
            <div class="k">冻结中</div>
            <div class="v">¥ {{ money(wallet?.frozenAmount) }}</div>
          </div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card">
          <div class="card-icon">💎</div>
          <div class="card-content">
            <div class="k">累计收入</div>
            <div class="v">¥ {{ money(wallet?.totalIncome) }}</div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 收入/提现 -->
    <el-card class="card" shadow="never" style="margin-top: 12px;">
      <div class="tabs-top">
        <div class="tabs-title">明细</div>
        <div class="tabs-actions">
          <el-button type="primary" plain @click="openWithdrawDialog">申请提现</el-button>
        </div>
      </div>

      <el-tabs v-model="activeTab" class="tabs">
        <el-tab-pane label="收入流水" name="income">
          <el-empty v-if="!loadingIncome && income.length === 0" description="暂无流水" />
          <el-table v-else :data="income" stripe border>
            <el-table-column prop="createdAt" label="时间" width="180">
              <template #default="{ row }">{{ fmt(row.createdAt) }}</template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="120" />
            <el-table-column prop="orderId" label="订单ID" width="100" />
            <el-table-column prop="remark" label="备注" />
            <el-table-column prop="amount" label="金额" width="120">
              <template #default="{ row }">
                <b>¥ {{ money(row.amount) }}</b>
              </template>
            </el-table-column>
          </el-table>

          <div class="pager">
            <el-button :loading="loadingIncome" @click="loadIncome">刷新流水</el-button>
          </div>
        </el-tab-pane>

        <el-tab-pane label="提现记录" name="withdraws">
          <el-empty v-if="!loadingWithdraws && withdraws.length === 0" description="暂无提现记录" />
          <el-table v-else :data="withdraws" stripe border>
            <el-table-column prop="appliedAt" label="申请时间" width="180">
              <template #default="{ row }">{{ fmt(row.appliedAt) }}</template>
            </el-table-column>
            <el-table-column prop="amount" label="金额" width="120">
              <template #default="{ row }">
                <b>¥ {{ money(row.amount) }}</b>
              </template>
            </el-table-column>
            <el-table-column prop="account" label="到账账号" />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="processedAt" label="处理时间" width="180">
              <template #default="{ row }">{{ row.processedAt ? fmt(row.processedAt) : "-" }}</template>
            </el-table-column>
          </el-table>

          <div class="pager">
            <el-button :loading="loadingWithdraws" @click="loadWithdraws">刷新记录</el-button>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 提现弹窗 -->
    <el-dialog v-model="withdrawDialog" title="申请提现" width="420px">
      <el-form :model="withdrawForm" label-width="88px">
        <el-form-item label="金额">
          <el-input v-model.number="withdrawForm.amount" type="number" placeholder="例如 20" />
          <div class="tip">可用余额：¥ {{ money(wallet?.balance) }}</div>
        </el-form-item>
        <el-form-item label="账号">
          <el-input v-model="withdrawForm.account" placeholder="如：支付宝/银行卡号" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="withdrawDialog = false">取消</el-button>
        <el-button :loading="withdrawing" type="primary" @click="submitWithdraw">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { useRouter } from "vue-router";
import {
  riderApi,
  type RiderMe,
  type RiderWallet,
  type RiderIncomeRecord,
  type RiderWithdraw,
} from "@/api/rider";

const router = useRouter();
const username = localStorage.getItem("username") || "rider";

const me = ref<RiderMe | null>(null);
const wallet = ref<RiderWallet | null>(null);

const online = ref(false);

const income = ref<RiderIncomeRecord[]>([]);
const withdraws = ref<RiderWithdraw[]>([]);

const activeTab = ref<"income" | "withdraws">("income");

const loading = ref(false);
const loadingIncome = ref(false);
const loadingWithdraws = ref(false);

const withdrawDialog = ref(false);
const withdrawing = ref(false);
const withdrawForm = ref({ amount: 0, account: "" });

const fmt = (s: string) => {
  if (!s) return "";
  const d = new Date(s);
  return isNaN(d.getTime()) ? s : d.toLocaleString();
};

const money = (n: any) => {
  const v = Number(n ?? 0);
  return isNaN(v) ? "0.00" : v.toFixed(2);
};

const loadMe = async () => {
  const res = await riderApi.getMe();
  me.value = res.data.data;
  online.value = !!me.value?.isOnline;
};

const loadWallet = async () => {
  const res = await riderApi.getWallet();
  wallet.value = res.data.data;
};

const loadIncome = async () => {
  loadingIncome.value = true;
  try {
    const res = await riderApi.getIncome({ page: 1, size: 50 });
    income.value = res.data.data || [];
  } finally {
    loadingIncome.value = false;
  }
};

const loadWithdraws = async () => {
  loadingWithdraws.value = true;
  try {
    const res = await riderApi.getWithdraws();
    withdraws.value = res.data.data || [];
  } finally {
    loadingWithdraws.value = false;
  }
};

const loadAll = async () => {
  loading.value = true;
  try {
    await Promise.all([loadMe(), loadWallet(), loadIncome(), loadWithdraws()]);
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || e?.message || "加载失败");
  } finally {
    loading.value = false;
  }
};

const saveOnline = async () => {
  try {
    // 你现在后端是 PATCH /api/rider/online，就按这个来
    // 如果你后端也支持 POST，这里改成 post 也行
    const res = await (await import("axios")).default.request({
      method: "PATCH",
      url: "/api/rider/online",
      data: { isOnline: online.value },
      headers: { Authorization: localStorage.getItem("token") || "" },
    });

    const okk = String(res.data?.code) === "1" || res.data?.code === 1;
    if (!okk) throw new Error(res.data?.msg || "update failed");

    if (me.value) me.value.isOnline = online.value;
    ElMessage.success("状态已更新");
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || e?.message || "更新失败");
    online.value = !online.value; // 回滚
  }
};

const openWithdrawDialog = () => {
  withdrawForm.value = { amount: 0, account: "" };
  withdrawDialog.value = true;
};

const submitWithdraw = async () => {
  const amount = Number(withdrawForm.value.amount || 0);
  const account = (withdrawForm.value.account || "").trim();

  if (amount <= 0) {
    ElMessage.warning("请输入正确的金额");
    return;
  }
  if (!account) {
    ElMessage.warning("请输入到账账号");
    return;
  }

  withdrawing.value = true;
  try {
    const res = await riderApi.applyWithdraw({ amount, account });
    const okk = String(res.data.code) === "1" || res.data.code === 1;
    if (!okk) throw new Error(res.data.msg || "提现失败");

    ElMessage.success("已提交提现申请");
    withdrawDialog.value = false;
    await Promise.all([loadWallet(), loadWithdraws()]);
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || e?.message || "提现失败");
  } finally {
    withdrawing.value = false;
  }
};

const logout = () => {
  localStorage.removeItem("token");
  localStorage.removeItem("role");
  localStorage.removeItem("username");
  router.push("/login");
};

onMounted(loadAll);
</script>

<style scoped lang="scss">
.me-page {
  padding: 20px;
  background: var(--rider-bg);
  min-height: calc(100vh - 60px);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.title {
  font-size: 24px;
  font-weight: 800;
  color: var(--rider-text);
  display: flex;
  align-items: center;
  gap: 12px;

  .title-icon {
    font-size: 28px;
  }
}

.header-actions {
  display: flex;
  gap: 12px;
}

.card {
  border-radius: var(--rider-radius);
  border: 1px solid var(--rider-border);
  box-shadow: var(--rider-shadow);
  transition: all 0.3s ease;

  &:hover {
    box-shadow: var(--rider-shadow-hover);
  }
}

.profile-card {
  margin-bottom: 24px;

  :deep(.el-card__body) {
    padding: 30px;
  }
}

.profile-header {
  display: flex;
  align-items: flex-start;
  gap: 24px;
}

.avatar-wrapper {
  flex-shrink: 0;
}

.avatar {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, var(--rider-primary) 0%, var(--rider-primary-dark) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(255, 179, 2, 0.3);

  .avatar-emoji {
    font-size: 40px;
  }
}

.profile-info {
  flex: 1;
}

.name {
  font-size: 24px;
  font-weight: 800;
  color: var(--rider-text);
  margin-bottom: 12px;
}

.sub {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: var(--rider-sub);
  font-size: 14px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.divider {
  color: var(--rider-border);
  margin: 0 4px;
}

.profile-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 16px;
}

.online-switch {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.status-label {
  font-size: 13px;
  color: var(--rider-sub);
  font-weight: 600;
}

.logout-btn {
  border-radius: 20px;
  padding: 8px 20px;
}

.section-title {
  margin-top: 32px;
  margin-bottom: 16px;
  font-weight: 800;
  font-size: 18px;
  color: var(--rider-text);
  display: flex;
  align-items: center;
  gap: 10px;

  .section-icon {
    font-size: 22px;
  }
}

.wallet-row {
  margin-top: 16px;
}

.stat-card {
  background: #fff;
  border: 1px solid var(--rider-border);
  border-radius: var(--rider-radius);
  padding: 20px;
  display: flex;
  align-items: flex-start;
  gap: 16px;
  transition: all 0.3s ease;
  box-shadow: var(--rider-shadow);

  &:hover {
    box-shadow: var(--rider-shadow-hover);
    transform: translateY(-2px);
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

    &:first-child {
      color: var(--rider-primary);
    }
  }
}

.tabs-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.tabs-title {
  font-weight: 700;
  font-size: 16px;
  color: var(--rider-text);
}

.tabs-actions {
  display: flex;
  gap: 12px;
}

.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.tip {
  margin-top: 8px;
  font-size: 12px;
  color: var(--rider-sub);
}

:deep(.el-button--primary) {
  background: var(--rider-primary);
  border-color: var(--rider-primary);
  border-radius: 20px;
  padding: 8px 20px;
  font-weight: 600;

  &:hover {
    background: var(--rider-primary-dark);
    border-color: var(--rider-primary-dark);
  }
}

:deep(.el-tabs__header) {
  margin: 0;
}

:deep(.el-tabs__nav-wrap::after) {
  display: none;
}

:deep(.el-tabs__item) {
  font-weight: 600;
  font-size: 15px;
  padding: 0 20px;
}

:deep(.el-tabs__active-bar) {
  background-color: var(--rider-primary);
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
.icon-logout:before { content: "🚪"; }
</style>
