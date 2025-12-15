<template>
  <div>
    <div class="header">
      <div class="title">我的</div>
      <div class="header-actions">
        <el-button :loading="loading" type="primary" @click="loadAll">刷新</el-button>
      </div>
    </div>

    <!-- 基本资料 + 在线 -->
    <el-card class="card" shadow="never">
      <div class="row">
        <div class="left">
          <div class="name">{{ me?.name || username }}</div>
          <div class="sub">手机号：{{ me?.phone || "-" }}</div>
          <div class="sub">已完成：{{ me?.completedOrders ?? 0 }}，评分：{{ me?.rating ?? 0 }}</div>
        </div>

        <div class="right">
          <div class="online">
            <span class="label">在线</span>
            <el-switch v-model="online" @change="saveOnline" />
          </div>
          <el-button type="danger" plain @click="logout">退出登录</el-button>
        </div>
      </div>
    </el-card>

    <!-- 钱包概览 -->
    <div class="section-title">钱包</div>
    <el-row :gutter="12" class="wallet-row">
      <el-col :span="8">
        <div class="stat-card">
          <div class="k">余额</div>
          <div class="v">¥ {{ money(wallet?.balance) }}</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card">
          <div class="k">冻结中</div>
          <div class="v">¥ {{ money(wallet?.frozenAmount) }}</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card">
          <div class="k">累计收入</div>
          <div class="v">¥ {{ money(wallet?.totalIncome) }}</div>
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
.header { display:flex; justify-content:space-between; align-items:center; margin-bottom:12px; }
.title { font-size: 18px; font-weight: 800; }
.header-actions { display:flex; gap: 10px; }

.card { border-radius: 12px; border: 1px solid var(--rider-border); }

.row { display:flex; justify-content:space-between; align-items:flex-start; gap: 16px; }
.name { font-size: 20px; font-weight: 800; margin-bottom: 6px; }
.sub { color:#606266; margin-top: 6px; }

.online { display:flex; align-items:center; gap:10px; margin-bottom: 10px; justify-content:flex-end; }
.label { color:#606266; }
.right { display:flex; flex-direction:column; align-items:flex-end; gap:8px; }

.section-title {
  margin-top: 14px;
  margin-bottom: 10px;
  font-weight: 800;
  color: #303133;
}

.wallet-row { margin-top: 6px; }
.stat-card{
  background:#fff;
  border:1px solid var(--rider-border);
  border-radius: 12px;
  padding: 12px 14px;
}
.k{ font-size:12px; color:#909399; margin-bottom: 8px; }
.v{ font-size: 20px; font-weight: 900; color:#303133; }

.tabs-top{ display:flex; align-items:center; justify-content:space-between; margin-bottom: 6px; }
.tabs-title{ font-weight: 800; }
.tabs-actions{ display:flex; gap: 10px; }

.pager { margin-top: 12px; display:flex; justify-content:flex-end; }

.tip { margin-top: 6px; font-size: 12px; color:#909399; }
</style>
