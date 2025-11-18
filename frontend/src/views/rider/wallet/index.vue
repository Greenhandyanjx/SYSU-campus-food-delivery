<template>
  <div class="rider-wallet">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">我的钱包</h1>
      <div class="withdraw-btn" @click="showWithdrawDialog = true">
        提现
      </div>
    </div>

    <!-- 钱包概览卡片 -->
    <div class="wallet-overview">
      <div class="balance-card">
        <div class="balance-header">
          <div class="balance-info">
            <div class="balance-label">账户余额</div>
            <div class="balance-value">¥{{ walletInfo.balance.toFixed(2) }}</div>
          </div>
          <div class="balance-icon">💰</div>
        </div>

        <div class="balance-details">
          <div class="detail-item">
            <span class="detail-label">冻结金额：</span>
            <span class="detail-value">¥{{ walletInfo.frozenAmount.toFixed(2) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">可提现：</span>
            <span class="detail-value available">¥{{ (walletInfo.balance - walletInfo.frozenAmount).toFixed(2) }}</span>
          </div>
        </div>

        <div class="balance-stats">
          <div class="stat-item">
            <div class="stat-value">{{ walletInfo.totalIncome.toFixed(2) }}</div>
            <div class="stat-label">总收入</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ walletInfo.withdrawnAmount.toFixed(2) }}</div>
            <div class="stat-label">已提现</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 收入统计 -->
    <div class="income-section">
      <div class="section-header">
        <h3>收入统计</h3>
        <div class="time-tabs">
          <div
            v-for="tab in timeTabs"
            :key="tab.value"
            class="time-tab"
            :class="{ active: activeTimeTab === tab.value }"
            @click="switchTimeTab(tab.value)"
          >
            {{ tab.label }}
          </div>
        </div>
      </div>

      <div class="income-chart">
        <div class="chart-placeholder">
          <div class="chart-icon">📊</div>
          <div class="chart-text">收入趋势图表</div>
          <div class="chart-period">{{ getCurrentTimeText() }}</div>
        </div>
      </div>

      <div class="income-summary">
        <div class="summary-item">
          <span class="summary-label">订单收入：</span>
          <span class="summary-value">¥{{ incomeSummary.orderIncome.toFixed(2) }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">奖励收入：</span>
          <span class="summary-value bonus">¥{{ incomeSummary.bonusIncome.toFixed(2) }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">总收入：</span>
          <span class="summary-value total">¥{{ incomeSummary.totalIncome.toFixed(2) }}</span>
        </div>
      </div>
    </div>

    <!-- 提现记录 -->
    <div class="withdraw-section">
      <div class="section-header">
        <h3>提现记录</h3>
        <el-button type="text" @click="viewAllWithdraws" v-if="withdrawHistory.length > 0">
          查看全部
        </el-button>
      </div>

      <div v-if="withdrawHistory.length === 0" class="empty-withdraw">
        <el-empty description="暂无提现记录" />
      </div>

      <div v-else>
        <div
          v-for="withdraw in withdrawHistory"
          :key="withdraw.id"
          class="withdraw-item"
        >
          <div class="withdraw-info">
            <div class="withdraw-amount">-¥{{ withdraw.amount.toFixed(2) }}</div>
            <div class="withdraw-status" :class="withdraw.status">
              {{ getWithdrawStatusText(withdraw.status) }}
            </div>
          </div>
          <div class="withdraw-details">
            <div class="withdraw-time">{{ formatDateTime(withdraw.appliedAt) }}</div>
            <div class="withdraw-account">{{ maskAccount(withdraw.account) }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 收入明细 -->
    <div class="income-details-section">
      <div class="section-header">
        <h3>收入明细</h3>
        <el-button type="text" @click="viewAllIncome">
          查看全部
        </el-button>
      </div>

      <div v-if="incomeHistory.length === 0" class="empty-income">
        <el-empty description="暂无收入记录" />
      </div>

      <div v-else>
        <div
          v-for="income in incomeHistory"
          :key="income.id"
          class="income-item"
        >
          <div class="income-icon" :class="income.type">
            {{ getIncomeIcon(income.type) }}
          </div>
          <div class="income-info">
            <div class="income-title">{{ getIncomeTitle(income.type) }}</div>
            <div class="income-time">{{ formatDateTime(income.time) }}</div>
            <div v-if="income.remark" class="income-remark">{{ income.remark }}</div>
          </div>
          <div class="income-amount">
            <span class="amount" :class="income.type">+¥{{ income.amount.toFixed(2) }}</span>
          </div>
        </div>
      </div>

      <div v-if="hasMoreIncome" class="load-more">
        <el-button @click="loadMoreIncome" :loading="loadingMore">加载更多</el-button>
      </div>
    </div>

    <!-- 底部导航栏 -->
    <div class="bottom-nav">
      <div class="nav-item" @click="$router.push('/rider')">
        <i class="css-icon house"></i>
        <span>首页</span>
      </div>
      <div class="nav-item" @click="$router.push('/rider/dashboard')">
        <i class="css-icon data-analysis"></i>
        <span>工作台</span>
      </div>
      <div class="nav-item" @click="$router.push('/rider/orders')">
        <i class="css-icon list"></i>
        <span>订单</span>
      </div>
      <div class="nav-item active" @click="$router.push('/rider/profile')">
        <i class="css-icon user"></i>
        <span>我的</span>
      </div>
    </div>

    <!-- 提现弹窗 -->
    <el-dialog
      v-model="showWithdrawDialog"
      title="申请提现"
      width="85%"
      :before-close="handleWithdrawClose"
    >
      <div class="withdraw-dialog-content">
        <div class="withdraw-form">
          <div class="form-item">
            <label class="form-label">提现金额</label>
            <div class="amount-input-group">
              <span class="currency">¥</span>
              <input
                v-model="withdrawForm.amount"
                type="number"
                class="amount-input"
                placeholder="请输入提现金额"
                @input="validateWithdrawAmount"
              />
            </div>
            <div class="amount-tips">
              <span>可提现余额：¥{{ (walletInfo.balance - walletInfo.frozenAmount).toFixed(2) }}</span>
            </div>
          </div>

          <div class="form-item">
            <label class="form-label">提现账户</label>
            <el-select v-model="withdrawForm.account" placeholder="请选择提现账户" style="width: 100%">
              <el-option
                v-for="account in withdrawAccounts"
                :key="account.id"
                :label="account.label"
                :value="account.value"
              />
            </el-select>
            <div class="account-actions">
              <el-button type="text" @click="showAddAccountDialog = true">
                + 添加账户
              </el-button>
            </div>
          </div>

          <div class="withdraw-notice">
            <h4>提现说明</h4>
            <ul>
              <li>提现申请将在1-3个工作日内处理</li>
              <li>单次最低提现金额为¥10</li>
              <li>提现手续费为提现金额的0.6%</li>
              <li>周末及节假日顺延处理</li>
            </ul>
          </div>
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showWithdrawDialog = false">取消</el-button>
          <el-button type="primary" @click="submitWithdraw" :disabled="!isValidWithdrawAmount" :loading="withdrawing">
            确认提现
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 添加账户弹窗 -->
    <el-dialog
      v-model="showAddAccountDialog"
      title="添加提现账户"
      width="85%"
    >
      <div class="add-account-form">
        <div class="form-item">
          <label class="form-label">账户类型</label>
          <el-radio-group v-model="newAccount.type">
            <el-radio label="alipay">支付宝</el-radio>
            <el-radio label="wechat">微信</el-radio>
            <el-radio label="bank">银行卡</el-radio>
          </el-radio-group>
        </div>

        <div class="form-item">
          <label class="form-label">账户姓名</label>
          <input
            v-model="newAccount.name"
            type="text"
            class="form-input"
            placeholder="请输入账户姓名"
          />
        </div>

        <div class="form-item">
          <label class="form-label">账户号码</label>
          <input
            v-model="newAccount.number"
            type="text"
            class="form-input"
            placeholder="请输入账户号码"
          />
        </div>

        <div v-if="newAccount.type === 'bank'" class="form-item">
          <label class="form-label">开户银行</label>
          <input
            v-model="newAccount.bank"
            type="text"
            class="form-input"
            placeholder="请输入开户银行"
          />
        </div>
      </div>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddAccountDialog = false">取消</el-button>
          <el-button type="primary" @click="addAccount">添加</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

// 状态管理
const loading = ref(false)
const loadingMore = ref(false)
const withdrawing = ref(false)
const showWithdrawDialog = ref(false)
const showAddAccountDialog = ref(false)

// 钱包信息
const walletInfo = ref({
  balance: 2580.50,
  frozenAmount: 120.00,
  totalIncome: 15680.00,
  withdrawnAmount: 13000.00
})

// 时间统计
const timeTabs = [
  { label: '今日', value: 'today' },
  { label: '本周', value: 'week' },
  { label: '本月', value: 'month' },
  { label: '全部', value: 'all' }
]

const activeTimeTab = ref('week')

// 收入统计
const incomeSummary = ref({
  orderIncome: 856.50,
  bonusIncome: 45.00,
  totalIncome: 901.50
})

// 提现记录
const withdrawHistory = ref([
  {
    id: 'w001',
    amount: 500.00,
    status: 'success',
    appliedAt: '2024-11-15T10:30:00',
    account: '支付宝(138****8000)'
  },
  {
    id: 'w002',
    amount: 300.00,
    status: 'processing',
    appliedAt: '2024-11-16T14:20:00',
    account: '微信(136****6666)'
  }
])

// 收入明细
const incomeHistory = ref([
  {
    id: 'i001',
    type: 'order',
    amount: 6.50,
    time: '2024-11-17T14:30:00',
    remark: '订单RD20241117001配送费'
  },
  {
    id: 'i002',
    type: 'bonus',
    amount: 2.00,
    time: '2024-11-17T13:45:00',
    remark: '准时配送奖励'
  },
  {
    id: 'i003',
    type: 'order',
    amount: 5.00,
    time: '2024-11-17T12:20:00',
    remark: '订单RD20241117003配送费'
  }
])

const hasMoreIncome = ref(true)
const incomePage = ref(1)

// 提现表单
const withdrawForm = ref({
  amount: '',
  account: ''
})

// 提现账户
const withdrawAccounts = ref([
  { id: 'acc001', label: '支付宝(138****8000)', value: 'alipay_138****8000' },
  { id: 'acc002', label: '微信(136****6666)', value: 'wechat_136****6666' },
  { id: 'acc003', label: '建设银行(****1234)', value: 'bank_****1234' }
])

// 新账户
const newAccount = ref({
  type: 'alipay',
  name: '',
  number: '',
  bank: ''
})

// 计算属性
const isValidWithdrawAmount = computed(() => {
  const amount = parseFloat(withdrawForm.value.amount)
  const availableAmount = walletInfo.value.balance - walletInfo.value.frozenAmount
  return amount >= 10 && amount <= availableAmount && withdrawForm.value.account
})

// 方法定义
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  const date = new Date(dateTime)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const maskAccount = (account) => {
  return account.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
}

const getWithdrawStatusText = (status) => {
  const statusMap = {
    success: '提现成功',
    processing: '处理中',
    failed: '提现失败',
    pending: '待处理'
  }
  return statusMap[status] || '未知状态'
}

const getIncomeIcon = (type) => {
  const iconMap = {
    order: '📦',
    bonus: '🎁',
    reward: '🏆',
    penalty: '⚠️'
  }
  return iconMap[type] || '💰'
}

const getIncomeTitle = (type) => {
  const titleMap = {
    order: '订单收入',
    bonus: '奖励收入',
    reward: '额外奖励',
    penalty: '违约扣款'
  }
  return titleMap[type] || '其他收入'
}

const getCurrentTimeText = () => {
  const tab = timeTabs.find(t => t.value === activeTimeTab.value)
  return tab ? tab.label : '本周'
}

// 加载钱包信息
const loadWalletInfo = async () => {
  try {
    loading.value = true
    const response = await riderApi.getWalletInfo()

    if (response.code === 1) {
      walletInfo.value = response.data
    }
  } catch (error) {
    console.error('加载钱包信息失败:', error)
    // 使用Demo数据
  } finally {
    loading.value = false
  }
}

// 切换时间统计
const switchTimeTab = (tab) => {
  activeTimeTab.value = tab
  // TODO: 重新加载对应的收入统计数据
  loadIncomeStats()
}

// 加载收入统计
const loadIncomeStats = async () => {
  try {
    const params = { period: activeTimeTab.value }
    const response = await riderApi.getIncomeStats(params)

    if (response.code === 1) {
      incomeSummary.value = {
        orderIncome: response.data.orderIncome || 0,
        bonusIncome: response.data.bonusIncome || 0,
        totalIncome: response.data.totalIncome || 0
      }
    }
  } catch (error) {
    console.error('加载收入统计失败:', error)
    // 使用Demo数据
  }
}

// 加载提现记录
const loadWithdrawHistory = async () => {
  try {
    const response = await riderApi.getWithdrawHistory()

    if (response.code === 1) {
      withdrawHistory.value = response.data || []
    }
  } catch (error) {
    console.error('加载提现记录失败:', error)
    // 使用Demo数据
  }
}

// 加载收入明细
const loadIncomeHistory = async (isLoadMore = false) => {
  try {
    if (isLoadMore) {
      loadingMore.value = true
    }

    const params = {
      page: incomePage.value,
      size: 10
    }

    const response = await riderApi.getIncomeHistory(params)

    if (response.code === 1) {
      const newIncome = response.data.items || []

      if (isLoadMore) {
        incomeHistory.value = [...incomeHistory.value, ...newIncome]
      } else {
        incomeHistory.value = newIncome
      }

      hasMoreIncome.value = newIncome.length === params.size
    }
  } catch (error) {
    console.error('加载收入明细失败:', error)
    // 使用Demo数据
  } finally {
    loadingMore.value = false
  }
}

// 加载更多收入明细
const loadMoreIncome = () => {
  if (!hasMoreIncome.value || loadingMore.value) return
  incomePage.value++
  loadIncomeHistory(true)
}

// 验证提现金额
const validateWithdrawAmount = () => {
  const amount = parseFloat(withdrawForm.value.amount)
  const availableAmount = walletInfo.value.balance - walletInfo.frozenAmount

  if (amount > availableAmount) {
    withdrawForm.value.amount = availableAmount.toString()
  }
}

// 提交提现申请
const submitWithdraw = async () => {
  try {
    await ElMessageBox.confirm('确认提交提现申请？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    withdrawing.value = true
    const withdrawData = {
      amount: parseFloat(withdrawForm.value.amount),
      account: withdrawForm.value.account
    }

    const response = await riderApi.withdraw(withdrawData)

    if (response.code === 1) {
      ElMessage.success('提现申请提交成功')
      showWithdrawDialog.value = false
      withdrawForm.value = { amount: '', account: '' }

      // 重新加载数据
      loadWalletInfo()
      loadWithdrawHistory()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('提现申请失败:', error)
      ElMessage.error('提现申请失败，请重试')
    }
  } finally {
    withdrawing.value = false
  }
}

// 添加提现账户
const addAccount = () => {
  if (!newAccount.value.name || !newAccount.value.number) {
    ElMessage.warning('请填写完整的账户信息')
    return
  }

  const accountLabel = `${newAccount.value.type === 'alipay' ? '支付宝' :
                     newAccount.value.type === 'wechat' ? '微信' : '银行卡'}(${newAccount.value.number.slice(-4)})`

  withdrawAccounts.value.push({
    id: `acc${Date.now()}`,
    label: accountLabel,
    value: `${newAccount.value.type}_${newAccount.value.number.slice(-4)}`
  })

  ElMessage.success('账户添加成功')
  showAddAccountDialog.value = false

  // 重置表单
  newAccount.value = {
    type: 'alipay',
    name: '',
    number: '',
    bank: ''
  }
}

// 查看全部提现记录
const viewAllWithdraws = () => {
  ElMessage.info('查看全部提现记录功能开发中...')
}

// 查看全部收入明细
const viewAllIncome = () => {
  ElMessage.info('查看全部收入明细功能开发中...')
}

// 关闭提现弹窗
const handleWithdrawClose = () => {
  withdrawForm.value = { amount: '', account: '' }
  showWithdrawDialog.value = false
}

onMounted(() => {
  loadWalletInfo()
  loadIncomeStats()
  loadWithdrawHistory()
  loadIncomeHistory()
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

/* 房子图标 */
.css-icon.house::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 10px;
  border: 2px solid currentColor;
  border-top: none;
}

.css-icon.house::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 0;
  border-left: 10px solid transparent;
  border-right: 10px solid transparent;
  border-bottom: 8px solid currentColor;
}

/* 数据分析图标 */
.css-icon.data-analysis::before {
  content: '';
  position: absolute;
  bottom: 0;
  left: 2px;
  width: 3px;
  height: 6px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 4px 0 0 currentColor, 8px 0 0 currentColor, 12px 0 0 currentColor;
}

.css-icon.data-analysis::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 2px;
  width: 3px;
  height: 10px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 8px 0 0 currentColor;
}

/* 列表图标 */
.css-icon.list::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 0 4px 0 currentColor, 0 8px 0 currentColor;
}

.css-icon.list::after {
  content: '';
  position: absolute;
  top: 0;
  right: 2px;
  width: 10px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
  box-shadow: 0 4px 0 currentColor, 0 8px 0 currentColor;
}

/* 用户图标 */
.css-icon.user::before {
  content: '';
  position: absolute;
  top: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 6px;
  background: currentColor;
  border-radius: 50%;
}

.css-icon.user::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 10px;
  height: 8px;
  background: currentColor;
  border-radius: 5px 5px 0 0;
}

.rider-wallet {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 60px;
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

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
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

.withdraw-btn {
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 20px;
  color: #333;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.withdraw-btn:hover {
  background: white;
}

/* 钱包概览 */
.wallet-overview {
  padding: 15px;
}

.balance-card {
  background: linear-gradient(135deg, #FFD700 0%, #FFA500 100%);
  border-radius: 16px;
  padding: 25px;
  color: white;
  box-shadow: 0 4px 20px rgba(255, 215, 0, 0.3);
}

.balance-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.balance-label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 5px;
}

.balance-value {
  font-size: 32px;
  font-weight: bold;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.balance-icon {
  font-size: 40px;
  opacity: 0.8;
}

.balance-details {
  margin-bottom: 20px;
  padding: 15px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.detail-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.detail-label {
  font-size: 14px;
  opacity: 0.9;
}

.detail-value {
  font-size: 14px;
  font-weight: 500;
}

.detail-value.available {
  color: #67C23A;
}

.balance-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.stat-item {
  text-align: center;
  padding: 10px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
}

.stat-value {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  opacity: 0.9;
}

/* 收入统计 */
.income-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.section-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.time-tabs {
  display: flex;
  background: #f5f5f5;
  border-radius: 20px;
  padding: 3px;
}

.time-tab {
  padding: 6px 12px;
  font-size: 12px;
  color: #666;
  cursor: pointer;
  border-radius: 17px;
  transition: all 0.3s ease;
}

.time-tab.active {
  background: #FFD700;
  color: white;
}

.income-chart {
  height: 150px;
  background: #f8f9fa;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-bottom: 15px;
}

.chart-placeholder {
  text-align: center;
  color: #999;
}

.chart-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.chart-text {
  font-size: 14px;
  margin-bottom: 4px;
}

.chart-period {
  font-size: 12px;
  opacity: 0.8;
}

.income-summary {
  border-top: 1px solid #f0f0f0;
  padding-top: 15px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.summary-item:last-child {
  margin-bottom: 0;
}

.summary-label {
  font-size: 14px;
  color: #666;
}

.summary-value {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.summary-value.bonus {
  color: #E6A23C;
}

.summary-value.total {
  color: #67C23A;
  font-weight: bold;
}

/* 提现记录 */
.withdraw-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.empty-withdraw {
  padding: 40px 0;
}

.withdraw-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.withdraw-item:last-child {
  border-bottom: none;
}

.withdraw-info {
  text-align: left;
}

.withdraw-amount {
  font-size: 16px;
  font-weight: bold;
  color: #F56C6C;
  margin-bottom: 4px;
}

.withdraw-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
  background: #f0f0f0;
  color: #666;
}

.withdraw-status.success {
  background: #f0f9ff;
  color: #67C23A;
}

.withdraw-status.processing {
  background: #fff7e6;
  color: #E6A23C;
}

.withdraw-status.failed {
  background: #fef0f0;
  color: #F56C6C;
}

.withdraw-details {
  text-align: right;
}

.withdraw-time {
  font-size: 12px;
  color: #999;
  margin-bottom: 2px;
}

.withdraw-account {
  font-size: 14px;
  color: #666;
}

/* 收入明细 */
.income-details-section {
  margin: 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.empty-income {
  padding: 40px 0;
}

.income-item {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.income-item:last-child {
  border-bottom: none;
}

.income-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  margin-right: 12px;
  flex-shrink: 0;
}

.income-icon.order {
  background: #e8f5e8;
}

.income-icon.bonus {
  background: #fff7e6;
}

.income-icon.reward {
  background: #f0f9ff;
}

.income-icon.penalty {
  background: #fef0f0;
}

.income-info {
  flex: 1;
}

.income-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 2px;
}

.income-time {
  font-size: 12px;
  color: #999;
  margin-bottom: 2px;
}

.income-remark {
  font-size: 12px;
  color: #666;
}

.income-amount {
  text-align: right;
}

.amount {
  font-size: 16px;
  font-weight: bold;
}

.amount.order {
  color: #67C23A;
}

.amount.bonus {
  color: #E6A23C;
}

.amount.reward {
  color: #409EFF;
}

.amount.penalty {
  color: #F56C6C;
}

.load-more {
  text-align: center;
  padding: 20px 0;
}

/* 底部导航 */
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  border-top: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-around;
  padding: 5px 0;
  z-index: 100;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 5px 15px;
  cursor: pointer;
  color: #999;
  transition: all 0.3s ease;
}

.nav-item.active {
  color: #FFD700;
}

.nav-item .css-icon {
  font-size: 20px;
}

.nav-item span {
  font-size: 12px;
}

/* 提现弹窗 */
.withdraw-dialog-content {
  padding: 10px 0;
}

.withdraw-form {
  padding: 0 10px;
}

.form-item {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  font-weight: 500;
}

.amount-input-group {
  position: relative;
  display: flex;
  align-items: center;
}

.currency {
  position: absolute;
  left: 15px;
  font-size: 16px;
  color: #666;
  font-weight: 500;
}

.amount-input {
  flex: 1;
  height: 44px;
  padding: 0 15px 0 35px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  font-size: 16px;
  outline: none;
  transition: all 0.3s ease;
}

.amount-input:focus {
  border-color: #FFD700;
}

.amount-tips {
  font-size: 12px;
  color: #999;
  margin-top: 5px;
}

.account-actions {
  margin-top: 8px;
  text-align: right;
}

.withdraw-notice {
  margin-top: 20px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.withdraw-notice h4 {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: #333;
}

.withdraw-notice ul {
  margin: 0;
  padding-left: 20px;
}

.withdraw-notice li {
  font-size: 12px;
  color: #666;
  margin-bottom: 5px;
}

/* 添加账户弹窗 */
.add-account-form {
  padding: 10px;
}

.form-input {
  width: 100%;
  height: 40px;
  padding: 0 15px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: all 0.3s ease;
}

.form-input:focus {
  border-color: #FFD700;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .wallet-overview {
    padding: 10px;
  }

  .balance-card {
    padding: 20px;
  }

  .balance-value {
    font-size: 28px;
  }

  .balance-stats {
    gap: 15px;
  }

  .income-section,
  .withdraw-section,
  .income-details-section {
    margin: 10px;
    padding: 15px;
  }
}
</style>