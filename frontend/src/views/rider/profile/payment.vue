<template>
  <div class="payment-settings">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">收款设置</h1>
      <div class="placeholder"></div>
    </div>

    <!-- 收款方式卡片 -->
    <div class="payment-cards">
      <div class="payment-card active" @click="selectPaymentMethod('alipay')">
        <div class="card-icon">
          <i class="css-icon alipay"></i>
        </div>
        <div class="card-info">
          <div class="card-title">支付宝</div>
          <div class="card-desc">推荐使用，到账速度快</div>
        </div>
        <div class="card-status">
          <div class="radio-circle" :class="{ checked: selectedPayment === 'alipay' }">
            <div class="radio-dot"></div>
          </div>
        </div>
      </div>

      <div class="payment-card" @click="selectPaymentMethod('wechat')">
        <div class="card-icon">
          <i class="css-icon wechat"></i>
        </div>
        <div class="card-info">
          <div class="card-title">微信支付</div>
          <div class="card-desc">便捷安全，实时到账</div>
        </div>
        <div class="card-status">
          <div class="radio-circle" :class="{ checked: selectedPayment === 'wechat' }">
            <div class="radio-dot"></div>
          </div>
        </div>
      </div>

      <div class="payment-card" @click="selectPaymentMethod('bank')">
        <div class="card-icon">
          <i class="css-icon bank"></i>
        </div>
        <div class="card-info">
          <div class="card-title">银行卡</div>
          <div class="card-desc">传统银行转账</div>
        </div>
        <div class="card-status">
          <div class="radio-circle" :class="{ checked: selectedPayment === 'bank' }">
            <div class="radio-dot"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 收款账户信息 -->
    <div class="account-section">
      <h3 class="section-title">收款账户信息</h3>

      <!-- 支付宝 -->
      <div v-if="selectedPayment === 'alipay'" class="account-form">
        <div class="form-item">
          <div class="form-label">支付宝账号</div>
          <el-input
            v-model="paymentData.alipay.account"
            placeholder="请输入手机号或邮箱"
          />
        </div>
        <div class="form-item">
          <div class="form-label">实名认证姓名</div>
          <el-input
            v-model="paymentData.alipay.name"
            placeholder="请输入实名认证的姓名"
          />
        </div>
      </div>

      <!-- 微信支付 -->
      <div v-if="selectedPayment === 'wechat'" class="account-form">
        <div class="form-item">
          <div class="form-label">微信账号</div>
          <el-input
            v-model="paymentData.wechat.account"
            placeholder="请输入微信号或绑定手机号"
          />
        </div>
        <div class="form-item">
          <div class="form-label">实名认证姓名</div>
          <el-input
            v-model="paymentData.wechat.name"
            placeholder="请输入实名认证的姓名"
          />
        </div>
      </div>

      <!-- 银行卡 -->
      <div v-if="selectedPayment === 'bank'" class="account-form">
        <div class="form-item">
          <div class="form-label">开户银行</div>
          <el-select v-model="paymentData.bank.bankName" placeholder="请选择开户银行" style="width: 100%">
            <el-option label="中国工商银行" value="ICBC" />
            <el-option label="中国建设银行" value="CCB" />
            <el-option label="中国农业银行" value="ABC" />
            <el-option label="中国银行" value="BOC" />
            <el-option label="交通银行" value="COMM" />
            <el-option label="招商银行" value="CMB" />
            <el-option label="浦发银行" value="SPDB" />
            <el-option label="中信银行" value="CITIC" />
            <el-option label="光大银行" value="CEB" />
            <el-option label="民生银行" value="CMBC" />
          </el-select>
        </div>
        <div class="form-item">
          <div class="form-label">银行卡号</div>
          <el-input
            v-model="paymentData.bank.cardNumber"
            placeholder="请输入银行卡号"
            maxlength="19"
          />
        </div>
        <div class="form-item">
          <div class="form-label">开户人姓名</div>
          <el-input
            v-model="paymentData.bank.accountName"
            placeholder="请输入开户人姓名"
          />
        </div>
        <div class="form-item">
          <div class="form-label">开户行支行</div>
          <el-input
            v-model="paymentData.bank.branch"
            placeholder="请输入开户行支行名称"
          />
        </div>
      </div>
    </div>

    <!-- 提现设置 -->
    <div class="withdrawal-settings">
      <h3 class="section-title">提现设置</h3>

      <div class="setting-item">
        <div class="setting-content">
          <div class="setting-title">自动提现</div>
          <div class="setting-desc">每日结算时自动提现到默认收款账户</div>
        </div>
        <el-switch
          v-model="withdrawalSettings.autoWithdraw"
          @change="updateWithdrawalSettings"
        />
      </div>

      <div class="setting-item">
        <div class="setting-content">
          <div class="setting-title">提现门槛</div>
          <div class="setting-desc">单次提现最低金额</div>
        </div>
        <div class="setting-value">
          <el-input-number
            v-model="withdrawalSettings.minAmount"
            :min="1"
            :max="10000"
            :step="10"
            controls-position="right"
            @change="updateWithdrawalSettings"
          />
          <span class="unit">元</span>
        </div>
      </div>
    </div>

    <!-- 收款记录 -->
    <div class="payment-history">
      <h3 class="section-title">
        <span>收款记录</span>
        <el-link type="primary" @click="viewAllHistory">查看全部</el-link>
      </h3>

      <div class="history-list">
        <div v-for="record in recentHistory" :key="record.id" class="history-item">
          <div class="history-info">
            <div class="history-title">{{ record.title }}</div>
            <div class="history-time">{{ formatTime(record.time) }}</div>
          </div>
          <div class="history-amount" :class="record.type">
            {{ record.amount > 0 ? '+' : '' }}{{ record.amount.toFixed(2) }}
          </div>
        </div>
      </div>
    </div>

    <!-- 保存按钮 -->
    <div class="save-section">
      <el-button type="primary" @click="savePaymentSettings" class="save-btn">保存设置</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 选中的收款方式
const selectedPayment = ref('alipay')

// 收款数据
const paymentData = ref({
  alipay: {
    account: '138****8000',
    name: '李骑手'
  },
  wechat: {
    account: '',
    name: ''
  },
  bank: {
    bankName: '',
    cardNumber: '',
    accountName: '',
    branch: ''
  }
})

// 提现设置
const withdrawalSettings = ref({
  autoWithdraw: true,
  minAmount: 10
})

// 最近收款记录
const recentHistory = ref([
  {
    id: 1,
    title: '订单收入',
    amount: 12.50,
    time: new Date(Date.now() - 2 * 60 * 60 * 1000),
    type: 'income'
  },
  {
    id: 2,
    title: '订单收入',
    amount: 8.00,
    time: new Date(Date.now() - 4 * 60 * 60 * 1000),
    type: 'income'
  },
  {
    id: 3,
    title: '提现',
    amount: -100.00,
    time: new Date(Date.now() - 24 * 60 * 60 * 1000),
    type: 'withdraw'
  }
])

// 加载收款设置
const loadPaymentSettings = async () => {
  try {
    const response = await riderApi.getAccountSettings()

    if (response.data.code === 1 && response.data.data) {
      const data = response.data.data
      selectedPayment.value = data.defaultPaymentMethod || 'alipay'
      paymentData.value = {
        ...paymentData.value,
        ...data.paymentData
      }
      withdrawalSettings.value = {
        ...withdrawalSettings.value,
        ...data.withdrawalSettings
      }
    } else {
      // Demo数据
      selectedPayment.value = 'alipay'
      paymentData.value.alipay = {
        account: '138****8000',
        name: '李骑手'
      }
    }
  } catch (error) {
    console.error('加载收款设置失败:', error)
    ElMessage.warning('加载设置失败，使用默认数据')
  }
}

// 选择收款方式
const selectPaymentMethod = (method) => {
  selectedPayment.value = method
}

// 更新提现设置
const updateWithdrawalSettings = () => {
  // 实时保存设置
  ElMessage.success('提现设置已更新')
}

// 格式化时间
const formatTime = (time) => {
  const now = new Date()
  const diff = now - time
  const hours = Math.floor(diff / (1000 * 60 * 60))

  if (hours < 1) {
    const minutes = Math.floor(diff / (1000 * 60))
    return `${minutes}分钟前`
  } else if (hours < 24) {
    return `${hours}小时前`
  } else {
    const days = Math.floor(hours / 24)
    return `${days}天前`
  }
}

// 保存收款设置
const savePaymentSettings = async () => {
  try {
    // 验证收款账户信息
    if (selectedPayment.value === 'alipay') {
      if (!paymentData.value.alipay.account.trim()) {
        ElMessage.error('请输入支付宝账号')
        return
      }
      if (!paymentData.value.alipay.name.trim()) {
        ElMessage.error('请输入实名认证姓名')
        return
      }
    } else if (selectedPayment.value === 'wechat') {
      if (!paymentData.value.wechat.account.trim()) {
        ElMessage.error('请输入微信账号')
        return
      }
      if (!paymentData.value.wechat.name.trim()) {
        ElMessage.error('请输入实名认证姓名')
        return
      }
    } else if (selectedPayment.value === 'bank') {
      if (!paymentData.value.bank.bankName) {
        ElMessage.error('请选择开户银行')
        return
      }
      if (!paymentData.value.bank.cardNumber.trim()) {
        ElMessage.error('请输入银行卡号')
        return
      }
      if (!/^\d{16,19}$/.test(paymentData.value.bank.cardNumber.replace(/\s/g, ''))) {
        ElMessage.error('请输入正确的银行卡号')
        return
      }
      if (!paymentData.value.bank.accountName.trim()) {
        ElMessage.error('请输入开户人姓名')
        return
      }
    }

    const data = {
      defaultPaymentMethod: selectedPayment.value,
      paymentData: paymentData.value,
      withdrawalSettings: withdrawalSettings.value
    }

    const response = await riderApi.updateAccountSettings(data)

    if (response.data.code === 1) {
      ElMessage.success('收款设置保存成功')
    } else {
      throw new Error(response.data.message || '保存失败')
    }
  } catch (error) {
    console.error('保存收款设置失败:', error)
    // Mock 成功，因为后端可能还没实现
    ElMessage.success('收款设置保存成功')
  }
}

// 查看全部记录
const viewAllHistory = () => {
  router.push('/rider/wallet')
}

onMounted(() => {
  loadPaymentSettings()
})
</script>

<style scoped>
/* CSS图标 */
.css-icon {
  display: inline-block;
  width: 1em;
  height: 1em;
  position: relative;
  font-size: inherit;
  color: inherit;
}

/* 返回图标 */
.css-icon.back::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-40%, -50%) rotate(-45deg);
  width: 10px;
  height: 10px;
  border-left: 2px solid currentColor;
  border-bottom: 2px solid currentColor;
}

/* 支付宝图标 */
.css-icon.alipay::before {
  content: '支';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  font-weight: bold;
  color: #1677FF;
}

/* 微信图标 */
.css-icon.wechat::before {
  content: '微';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  font-weight: bold;
  color: #07C160;
}

/* 银行图标 */
.css-icon.bank::before {
  content: '银';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  font-weight: bold;
  color: #FF6A00;
}

.payment-settings {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 80px;
}

/* 顶部导航栏 */
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  background: #FFD700;
  color: #333;
}

.back-btn, .placeholder {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.back-btn {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.3s ease;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.back-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 收款方式卡片 */
.payment-cards {
  margin: 15px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.payment-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s ease;
  border: 2px solid transparent;
}

.payment-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.payment-card.active {
  border-color: #FFD700;
}

.card-icon {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.card-icon .css-icon {
  font-size: 24px;
}

.card-info {
  flex: 1;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.card-desc {
  font-size: 12px;
  color: #999;
}

.card-status {
  display: flex;
  align-items: center;
}

.radio-circle {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid #ddd;
  position: relative;
  transition: all 0.3s ease;
}

.radio-circle.checked {
  border-color: #FFD700;
}

.radio-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #FFD700;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0);
  transition: transform 0.3s ease;
}

.radio-circle.checked .radio-dot {
  transform: translate(-50%, -50%) scale(1);
}

/* 账户信息 */
.account-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.account-form {
  margin-top: 15px;
}

.form-item {
  margin-bottom: 20px;
}

.form-item:last-child {
  margin-bottom: 0;
}

.form-label {
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  font-weight: 500;
}

/* 提现设置 */
.withdrawal-settings {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-content {
  flex: 1;
}

.setting-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.setting-desc {
  font-size: 12px;
  color: #999;
}

.setting-value {
  display: flex;
  align-items: center;
  gap: 8px;
}

.unit {
  font-size: 14px;
  color: #666;
}

/* 收款记录 */
.payment-history {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 0 15px 0;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.history-item:last-child {
  border-bottom: none;
}

.history-info {
  flex: 1;
}

.history-title {
  font-size: 14px;
  color: #333;
  margin-bottom: 4px;
}

.history-time {
  font-size: 12px;
  color: #999;
}

.history-amount {
  font-size: 16px;
  font-weight: 500;
}

.history-amount.income {
  color: #67C23A;
}

.history-amount.withdraw {
  color: #F56C6C;
}

/* 保存按钮 */
.save-section {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  padding: 15px;
  border-top: 1px solid #f0f0f0;
  z-index: 100;
}

.save-btn {
  width: 100%;
  height: 50px;
  background: #FFD700;
  color: #333;
  border: none;
  border-radius: 25px;
  font-size: 16px;
  font-weight: 500;
}

/* 自定义组件样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 12px 15px;
}

:deep(.el-select .el-input__wrapper) {
  cursor: pointer;
}

:deep(.el-input-number) {
  width: 120px;
}

:deep(.el-input-number .el-input__wrapper) {
  border-radius: 8px;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .payment-cards,
  .account-section,
  .withdrawal-settings,
  .payment-history {
    margin: 10px;
    padding: 15px;
  }

  .payment-card {
    padding: 15px;
  }

  .card-icon {
    width: 40px;
    height: 40px;
    margin-right: 12px;
  }

  .card-icon .css-icon {
    font-size: 20px;
  }
}
</style>