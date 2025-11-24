<template>
  <div class="security-settings">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">账户安全</h1>
      <div class="placeholder"></div>
    </div>

    <!-- 安全状态卡片 -->
    <div class="security-card">
      <div class="security-score">
        <div class="score-circle">
          <div class="score-value">{{ securityScore }}</div>
          <div class="score-text">安全分</div>
        </div>
        <div class="security-tip">{{ securityTip }}</div>
      </div>
    </div>

    <!-- 安全设置列表 -->
    <div class="security-list">
      <!-- 手机号绑定 -->
      <div class="security-item" @click="managePhone">
        <div class="item-icon">
          <i class="css-icon phone"></i>
        </div>
        <div class="item-content">
          <div class="item-title">手机号绑定</div>
          <div class="item-desc">用于登录和找回密码</div>
          <div class="item-status" :class="{ 'verified': phoneVerified }">
            {{ phoneVerified ? '已绑定' : '未绑定' }}
          </div>
        </div>
        <div class="item-arrow">
          <i class="css-icon arrow"></i>
        </div>
      </div>

      <!-- 密码修改 -->
      <div class="security-item" @click="changePassword">
        <div class="item-icon">
          <i class="css-icon lock"></i>
        </div>
        <div class="item-content">
          <div class="item-title">登录密码</div>
          <div class="item-desc">定期修改密码保护账户安全</div>
          <div class="item-status">已设置</div>
        </div>
        <div class="item-arrow">
          <i class="css-icon arrow"></i>
        </div>
      </div>

      <!-- 支付密码 -->
      <div class="security-item" @click="managePayPassword">
        <div class="item-icon">
          <i class="css-icon pay"></i>
        </div>
        <div class="item-content">
          <div class="item-title">支付密码</div>
          <div class="item-desc">用于提现和支付验证</div>
          <div class="item-status" :class="{ 'verified': payPasswordSet }">
            {{ payPasswordSet ? '已设置' : '未设置' }}
          </div>
        </div>
        <div class="item-arrow">
          <i class="css-icon arrow"></i>
        </div>
      </div>

      <!-- 实名认证 -->
      <div class="security-item" @click="manageVerification">
        <div class="item-icon">
          <i class="css-icon id-card"></i>
        </div>
        <div class="item-content">
          <div class="item-title">实名认证</div>
          <div class="item-desc">保障资金安全，提升服务信任度</div>
          <div class="item-status" :class="{ 'verified': verificationStatus }">
            {{ getVerificationText() }}
          </div>
        </div>
        <div class="item-arrow">
          <i class="css-icon arrow"></i>
        </div>
      </div>

      <!-- 登录设备管理 -->
      <div class="security-item" @click="manageDevices">
        <div class="item-icon">
          <i class="css-icon device"></i>
        </div>
        <div class="item-content">
          <div class="item-title">登录设备管理</div>
          <div class="item-desc">管理已登录的设备</div>
          <div class="item-status">{{ deviceCount }}台设备</div>
        </div>
        <div class="item-arrow">
          <i class="css-icon arrow"></i>
        </div>
      </div>

      <!-- 异常登录提醒 -->
      <div class="security-item">
        <div class="item-icon">
          <i class="css-icon alert"></i>
        </div>
        <div class="item-content">
          <div class="item-title">异常登录提醒</div>
          <div class="item-desc">新设备登录时通知</div>
          <div class="item-status">
            <el-switch
              v-model="alertSettings.loginAlert"
              @change="updateAlertSettings('loginAlert', $event)"
              size="small"
            />
          </div>
        </div>
      </div>

      <!-- 密码保护提醒 -->
      <div class="security-item">
        <div class="item-icon">
          <i class="css-icon shield"></i>
        </div>
        <div class="item-content">
          <div class="item-title">密码保护提醒</div>
          <div class="item-desc">定期提醒修改密码</div>
          <div class="item-status">
            <el-switch
              v-model="alertSettings.passwordAlert"
              @change="updateAlertSettings('passwordAlert', $event)"
              size="small"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 账户冻结/注销 -->
    <div class="danger-zone">
      <h3 class="danger-title">危险操作</h3>
      <div class="danger-buttons">
        <el-button @click="freezeAccount" class="freeze-btn">冻结账户</el-button>
        <el-button type="danger" @click="deleteAccount" class="delete-btn">注销账户</el-button>
      </div>
    </div>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="passwordDialog.visible"
      title="修改密码"
      width="90%"
      :before-close="closePasswordDialog"
    >
      <div class="password-form">
        <div class="form-item">
          <div class="form-label">当前密码</div>
          <el-input
            v-model="passwordForm.currentPassword"
            type="password"
            placeholder="请输入当前密码"
            show-password
          />
        </div>
        <div class="form-item">
          <div class="form-label">新密码</div>
          <el-input
            v-model="passwordForm.newPassword"
            type="password"
            placeholder="请输入新密码（8-20位）"
            show-password
          />
        </div>
        <div class="form-item">
          <div class="form-label">确认新密码</div>
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          />
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="passwordDialog.visible = false">取消</el-button>
          <el-button type="primary" @click="submitPasswordChange">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 安全设置状态
const phoneVerified = ref(true)
const payPasswordSet = ref(true)
const verificationStatus = ref('verified') // verified, pending, rejected
const deviceCount = ref(2)

// 提醒设置
const alertSettings = ref({
  loginAlert: true,
  passwordAlert: false
})

// 密码修改对话框
const passwordDialog = ref({
  visible: false
})

const passwordForm = ref({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 计算安全分数
const securityScore = computed(() => {
  let score = 0
  if (phoneVerified.value) score += 25
  if (payPasswordSet.value) score += 25
  if (verificationStatus.value === 'verified') score += 30
  if (alertSettings.value.loginAlert) score += 10
  if (alertSettings.value.passwordAlert) score += 10
  return score
})

// 安全提示
const securityTip = computed(() => {
  const score = securityScore.value
  if (score >= 90) return '账户安全等级：高'
  if (score >= 60) return '账户安全等级：中'
  return '账户安全等级：低，请尽快完善安全设置'
})

// 加载安全设置
const loadSecuritySettings = async () => {
  try {
    const response = await riderApi.getAccountSettings()

    if (response.data.code === 1 && response.data.data) {
      const data = response.data.data
      phoneVerified.value = data.phoneVerified !== false
      payPasswordSet.value = data.payPasswordSet !== false
      verificationStatus.value = data.verificationStatus || 'pending'
      deviceCount.value = data.deviceCount || 2
    } else {
      // Demo数据
      phoneVerified.value = true
      payPasswordSet.value = true
      verificationStatus.value = 'verified'
      deviceCount.value = 2
    }
  } catch (error) {
    console.error('加载安全设置失败:', error)
    // Demo数据
    phoneVerified.value = true
    payPasswordSet.value = true
    verificationStatus.value = 'verified'
    deviceCount.value = 2
  }
}

// 获取认证状态文本
const getVerificationText = () => {
  switch (verificationStatus.value) {
    case 'verified':
      return '已认证'
    case 'pending':
      return '审核中'
    case 'rejected':
      return '未通过'
    default:
      return '未认证'
  }
}

// 管理手机号
const managePhone = () => {
  if (phoneVerified.value) {
    ElMessage.info('手机号已绑定，如需修改请联系客服')
  } else {
    router.push('/rider/profile/bind-phone')
  }
}

// 修改密码
const changePassword = () => {
  passwordDialog.value.visible = true
}

// 关闭密码对话框
const closePasswordDialog = () => {
  passwordDialog.value.visible = false
  passwordForm.value = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}

// 提交密码修改
const submitPasswordChange = async () => {
  try {
    if (!passwordForm.value.currentPassword) {
      ElMessage.error('请输入当前密码')
      return
    }

    if (!passwordForm.value.newPassword) {
      ElMessage.error('请输入新密码')
      return
    }

    if (passwordForm.value.newPassword.length < 8 || passwordForm.value.newPassword.length > 20) {
      ElMessage.error('密码长度应为8-20位')
      return
    }

    if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
      ElMessage.error('两次输入的密码不一致')
      return
    }

    // 模拟API调用
    ElMessage.success('密码修改成功，请重新登录')
    closePasswordDialog()

    // 跳转到登录页面
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  } catch (error) {
    console.error('密码修改失败:', error)
    ElMessage.error('密码修改失败，请重试')
  }
}

// 管理支付密码
const managePayPassword = () => {
  if (payPasswordSet.value) {
    ElMessage.info('支付密码已设置')
  } else {
    router.push('/rider/profile/set-pay-password')
  }
}

// 管理实名认证
const manageVerification = () => {
  router.push('/rider/profile/verification')
}

// 管理登录设备
const manageDevices = () => {
  router.push('/rider/profile/devices')
}

// 更新提醒设置
const updateAlertSettings = async (key, value) => {
  try {
    // 模拟API调用
    ElMessage.success(`${key === 'loginAlert' ? '登录提醒' : '密码提醒'}设置已更新`)
  } catch (error) {
    console.error('更新提醒设置失败:', error)
    ElMessage.error('设置更新失败')
    // 恢复原值
    alertSettings.value[key] = !value
  }
}

// 冻结账户
const freezeAccount = () => {
  ElMessageBox.prompt(
    '请输入冻结原因（可选）',
    '确认冻结账户',
    {
      confirmButtonText: '确定冻结',
      cancelButtonText: '取消',
      inputType: 'textarea',
      inputPlaceholder: '请输入冻结原因'
    }
  ).then(({ value }) => {
    // 模拟API调用
    ElMessage.success('账户已冻结，如需解冻请联系客服')
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  }).catch(() => {
    // 用户取消
  })
}

// 注销账户
const deleteAccount = () => {
  ElMessageBox.prompt(
    '注销账户将清除所有数据且无法恢复，请输入"确认注销"以继续',
    '危险操作：注销账户',
    {
      confirmButtonText: '确认注销',
      cancelButtonText: '取消',
      inputType: 'text',
      inputPlaceholder: '请输入"确认注销"',
      inputValidator: (value) => {
        return value === '确认注销' || '请输入正确的确认文字'
      }
    }
  ).then(() => {
    // 模拟API调用
    ElMessage.success('账户注销申请已提交，将在7天后生效')
    setTimeout(() => {
      router.push('/login')
    }, 1500)
  }).catch(() => {
    // 用户取消
  })
}

onMounted(() => {
  loadSecuritySettings()
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

/* 箭头图标 */
.css-icon.arrow::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(45deg);
  width: 8px;
  height: 8px;
  border-right: 2px solid currentColor;
  border-top: 2px solid currentColor;
}

/* 手机图标 */
.css-icon.phone::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 20px;
  border: 2px solid currentColor;
  border-radius: 4px;
}

.css-icon.phone::after {
  content: '';
  position: absolute;
  bottom: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 4px;
  height: 4px;
  background: currentColor;
  border-radius: 50%;
}

/* 锁图标 */
.css-icon.lock::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 4px;
}

.css-icon.lock::after {
  content: '';
  position: absolute;
  top: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 8px;
  height: 6px;
  border: 2px solid currentColor;
  border-radius: 4px 4px 0 0;
}

/* 支付图标 */
.css-icon.pay::before {
  content: '¥';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 16px;
  font-weight: bold;
}

/* 身份证图标 */
.css-icon.id-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 10px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.id-card::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 12px;
  height: 6px;
  background: currentColor;
  border-radius: 1px;
}

/* 设备图标 */
.css-icon.device::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 18px;
  height: 12px;
  border: 2px solid currentColor;
  border-radius: 4px;
}

.css-icon.device::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 50%;
  transform: translateX(-50%);
  width: 6px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
}

/* 警告图标 */
.css-icon.alert::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-radius: 50%;
}

.css-icon.alert::after {
  content: '!';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: bold;
}

/* 盾牌图标 */
.css-icon.shield::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 20px;
  border: 2px solid currentColor;
  border-radius: 8px 8px 0 0;
}

.css-icon.shield::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: bold;
}

.security-settings {
  background: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 20px;
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

/* 安全状态卡片 */
.security-card {
  background: white;
  margin: 15px;
  padding: 30px 20px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.security-score {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.score-circle {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: linear-gradient(135deg, #FFD700, #FFA500);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  margin-bottom: 15px;
  box-shadow: 0 4px 15px rgba(255, 215, 0, 0.3);
}

.score-value {
  font-size: 32px;
  font-weight: bold;
  color: white;
  line-height: 1;
}

.score-text {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.9);
  margin-top: 4px;
}

.security-tip {
  font-size: 14px;
  color: #666;
}

/* 安全设置列表 */
.security-list {
  margin: 0 15px;
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.security-item {
  display: flex;
  align-items: center;
  padding: 20px 15px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.3s ease;
}

.security-item:last-child {
  border-bottom: none;
}

.security-item:hover {
  background: #f8f9fa;
}

.item-icon {
  width: 40px;
  height: 40px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.item-icon .css-icon {
  font-size: 20px;
  color: #FFD700;
}

.item-content {
  flex: 1;
}

.item-title {
  font-size: 16px;
  color: #333;
  margin-bottom: 4px;
  font-weight: 500;
}

.item-desc {
  font-size: 12px;
  color: #999;
  margin-bottom: 6px;
}

.item-status {
  font-size: 14px;
  color: #666;
}

.item-status.verified {
  color: #67C23A;
}

.item-arrow {
  width: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.item-arrow .css-icon {
  font-size: 16px;
  color: #ccc;
}

/* 危险操作区域 */
.danger-zone {
  margin: 20px 15px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.danger-title {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #F56C6C;
  font-weight: 500;
}

.danger-buttons {
  display: flex;
  gap: 15px;
}

.freeze-btn, .delete-btn {
  flex: 1;
  height: 40px;
  border-radius: 20px;
}

.freeze-btn {
  background: #f8f9fa;
  color: #666;
  border: 1px solid #ddd;
}

/* 密码修改表单 */
.password-form {
  padding: 0 20px;
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

/* 自定义输入框样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 12px 15px;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .security-card {
    margin: 10px;
    padding: 20px 15px;
  }

  .security-list {
    margin: 0 10px;
  }

  .danger-zone {
    margin: 15px 10px;
    padding: 15px;
  }

  .danger-buttons {
    flex-direction: column;
  }

  .security-item {
    padding: 15px 10px;
  }
}
</style>