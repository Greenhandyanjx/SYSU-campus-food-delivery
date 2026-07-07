<template>
  <div class="wallet-bg" :style="{ backgroundImage: `url(${bgImg})` }">
    <div class="wallet-page">
      <!-- 余额卡片 -->
      <div class="balance-card">
        <div class="balance-header">
          <div class="balance-label">我的余额</div>
          <div class="balance-amount">¥{{ balance.toFixed(2) }}</div>
        </div>
        <div class="recharge-btn-wrap">
          <el-button type="warning" round class="recharge-btn" @click="showRecharge = true">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px">
              <path d="M12 5v14M5 12h14"/>
            </svg>
            充值
          </el-button>
        </div>
      </div>

      <!-- 交易流水 -->
      <div class="transactions-card">
        <div class="section-title">交易记录</div>
        <div v-if="transactions.length === 0" class="empty-txns">
          暂无交易记录
        </div>
        <div v-else class="txn-list">
          <div v-for="txn in transactions" :key="txn.id" class="txn-item">
            <div class="txn-left">
              <div class="txn-icon" :class="txn.type">
                <svg v-if="txn.type === 'recharge'" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 1v22M17 5l-5-4-5 4M5 19l5 4 5-4"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M12 23V1M7 19l5 4 5-4M19 5l-5-4-5 4"/>
                </svg>
              </div>
              <div class="txn-info">
                <div class="txn-desc">{{ txn.description || txn.type === 'recharge' ? '充值' : '支付' }}</div>
                <div class="txn-time">{{ formatTime(txn.created_at) }}</div>
              </div>
            </div>
            <div class="txn-right">
              <div class="txn-amount" :class="txn.type">
                {{ txn.type === 'recharge' ? '+' : '-' }}¥{{ Math.abs(txn.amount).toFixed(2) }}
              </div>
              <div class="txn-balance">余额 ¥{{ txn.balance_after.toFixed(2) }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 充值弹窗 -->
      <el-dialog v-model="showRecharge" title="余额充值" width="400px" :close-on-click-modal="false">
        <div class="recharge-modal" v-if="!showRechargeQr">
          <div class="recharge-presets">
            <div
              v-for="amt in presets"
              :key="amt"
              class="preset-btn"
              :class="{ active: rechargeAmount === amt }"
              @click="rechargeAmount = amt"
            >¥{{ amt }}</div>
          </div>
          <div class="custom-amount">
            <span class="custom-label">自定义金额</span>
            <el-input-number v-model="rechargeAmount" :min="1" :max="99999" :step="10" controls-position="right" />
          </div>
          <el-button type="warning" :disabled="!rechargeAmount || rechargeAmount <= 0" class="confirm-recharge-btn" @click="doRecharge">
            确认充值 ¥{{ (rechargeAmount || 0).toFixed(2) }}
          </el-button>
        </div>

        <!-- 假二维码 -->
        <div v-if="showRechargeQr" class="qr-section">
          <div class="qr-title">请使用微信/支付宝扫码支付</div>
          <img :src="qrImg" alt="QR Code" class="qr-img" />
          <div class="qr-amount">¥{{ (rechargeAmount || 0).toFixed(2) }}</div>
          <div class="qr-tip">支付成功后自动到账</div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as myApi from '@/api/user/my'
import request from '@/api/merchant/request'
import qrImg from '@/assets/qrcode.png'
import bgImg from '@/assets/login/img_denglu_bj.jpg'
import { ElMessage } from 'element-plus'

const balance = ref(0)
const transactions = ref<any[]>([])
const showRecharge = ref(false)
const showRechargeQr = ref(false)
const rechargeAmount = ref(50)
const presets = [10, 20, 50, 100, 200, 500]

async function loadWallet() {
  try {
    const res = await request.get('/user/wallet')
    if (res?.data?.code === 1 && res.data.data) {
      balance.value = res.data.data.balance || 0
      transactions.value = res.data.data.transactions || []
    } else if (res?.data?.data) {
      balance.value = res.data.data.balance || 0
      transactions.value = res.data.data.transactions || []
    }
  } catch (e) {
    console.error('Failed to load wallet', e)
  }
}

async function doRecharge() {
  if (!rechargeAmount.value || rechargeAmount.value <= 0) return
  showRechargeQr.value = true

  // 后台发起充值请求（不阻塞，confirm 独立工作）
  request.post('/user/wallet/recharge', { amount: rechargeAmount.value }).catch(() => {})

  // 3秒后自动确认充值（模拟扫码支付成功）
  setTimeout(async () => {
    try {
      const res = await request.post('/user/wallet/recharge/confirm', { amount: rechargeAmount.value })
      // request.ts 返回完整 AxiosResponse，data 为后端 JSON 体 {code, data, msg}
      const body = res?.data || res
      if (body?.code === 1) {
        const newBalance = body?.data?.balance || balance.value + rechargeAmount.value
        balance.value = newBalance
        ElMessage.success(`充值成功！余额 ¥${(newBalance || 0).toFixed(2)}`)
      } else {
        ElMessage.error(body?.msg || '充值失败')
      }
    } catch (e: any) {
      console.error('充值确认失败:', e?.response?.data || e?.message || e)
      ElMessage.error('充值失败，请重试')
    }
    showRechargeQr.value = false
    showRecharge.value = false
    rechargeAmount.value = 50
    loadWallet() // 刷新
  }, 2500)
}

function formatTime(ts: string | number) {
  if (!ts) return ''
  const d = new Date(ts)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

onMounted(() => {
  loadWallet()
})
</script>

<style scoped>
.wallet-bg {
  width: 100%;
  min-height: 100vh;
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: 60px 0;
}

.wallet-page {
  width: 60%;
  max-width: 760px;
}

/* 余额卡片 */
.balance-card {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  border-radius: 16px;
  padding: 32px;
  color: #fff;
  box-shadow: 0 8px 24px rgba(245, 158, 11, 0.35);
  margin-bottom: 20px;
  position: relative;
  overflow: hidden;
}

.balance-card::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -20%;
  width: 200px;
  height: 200px;
  border-radius: 50%;
  background: rgba(255,255,255,0.1);
}

.balance-label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.balance-amount {
  font-size: 40px;
  font-weight: 700;
  letter-spacing: -1px;
}

.recharge-btn-wrap {
  margin-top: 16px;
}

.recharge-btn {
  background: rgba(255,255,255,0.2) !important;
  border: 1px solid rgba(255,255,255,0.4) !important;
  color: #fff !important;
  font-weight: 600;
  padding: 10px 28px;
  font-size: 15px;
  backdrop-filter: blur(4px);
}

.recharge-btn:hover {
  background: rgba(255,255,255,0.3) !important;
}

/* 交易流水卡片 */
.transactions-card {
  background: rgba(255, 248, 225, 0.96);
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(255, 193, 7, 0.15);
  padding: 24px;
  backdrop-filter: blur(6px);
}

.section-title {
  font-size: 17px;
  font-weight: 700;
  color: #4a2c00;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(255,193,7,0.2);
}

.empty-txns {
  text-align: center;
  color: #c0a06b;
  padding: 40px 0;
  font-size: 14px;
}

.txn-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 0;
  border-bottom: 1px solid rgba(0,0,0,0.04);
}

.txn-item:last-child { border-bottom: none; }

.txn-left { display: flex; align-items: center; gap: 12px; }

.txn-icon {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.txn-icon.recharge {
  background: #e8f5e9;
  color: #2e7d32;
}

.txn-icon.payment {
  background: #fce4ec;
  color: #c62828;
}

.txn-info {}

.txn-desc {
  font-size: 14px;
  font-weight: 600;
  color: #3d2800;
}

.txn-time {
  font-size: 12px;
  color: #a08050;
  margin-top: 2px;
}

.txn-right { text-align: right; }

.txn-amount {
  font-size: 16px;
  font-weight: 700;
}

.txn-amount.recharge { color: #2e7d32; }
.txn-amount.payment { color: #c62828; }

.txn-balance {
  font-size: 11px;
  color: #a08050;
  margin-top: 2px;
}

/* 充值弹窗 */
.recharge-modal {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.recharge-presets {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.preset-btn {
  padding: 14px;
  text-align: center;
  border-radius: 12px;
  border: 2px solid #f0e0c0;
  font-size: 18px;
  font-weight: 700;
  color: #6b3f00;
  cursor: pointer;
  transition: all 0.2s;
  background: #fffef6;
}

.preset-btn:hover {
  border-color: #f59e0b;
  background: #fff9e6;
}

.preset-btn.active {
  border-color: #f59e0b;
  background: #fef3c7;
  color: #b45309;
}

.custom-amount {
  display: flex;
  align-items: center;
  gap: 16px;
}

.custom-label {
  font-size: 14px;
  color: #8c6b00;
  white-space: nowrap;
}

.confirm-recharge-btn {
  width: 100%;
  padding: 14px !important;
  font-size: 16px !important;
  font-weight: 700 !important;
}

/* QR 码 */
.qr-section {
  text-align: center;
  padding: 20px 0;
}

.qr-title {
  font-size: 16px;
  font-weight: 600;
  color: #3d2800;
  margin-bottom: 16px;
}

.qr-img {
  width: 200px;
  height: 200px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
}

.qr-amount {
  font-size: 24px;
  font-weight: 700;
  color: #f59e0b;
  margin-top: 12px;
}

.qr-tip {
  font-size: 13px;
  color: #a08050;
  margin-top: 8px;
}

@media(max-width:900px) {
  .wallet-page { width: 92%; }
  .balance-amount { font-size: 32px; }
  .recharge-presets { grid-template-columns: repeat(3, 1fr); }
}
</style>
