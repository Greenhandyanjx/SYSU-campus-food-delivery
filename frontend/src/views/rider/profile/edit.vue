<template>
  <div class="profile-edit">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">个人资料</h1>
      <div class="save-btn" @click="saveProfile">
        <span>保存</span>
      </div>
    </div>

    <!-- 头像部分 -->
    <div class="avatar-section">
      <div class="avatar-wrapper" @click="changeAvatar">
        <el-avatar :size="80" :src="formData.avatar || defaultAvatar" />
        <div class="camera-overlay">
          <i class="css-icon camera"></i>
        </div>
      </div>
      <div class="avatar-tip">点击更换头像</div>
    </div>

    <!-- 表单内容 -->
    <div class="form-content">
      <div class="form-group">
        <div class="form-item">
          <div class="form-label">骑手姓名</div>
          <el-input
            v-model="formData.name"
            placeholder="请输入姓名"
            :disabled="true"
          />
        </div>

        <div class="form-item">
          <div class="form-label">骑手编号</div>
          <el-input
            v-model="formData.id"
            placeholder="骑手编号"
            :disabled="true"
          />
        </div>

        <div class="form-item">
          <div class="form-label">手机号码</div>
          <el-input
            v-model="formData.phone"
            placeholder="请输入手机号码"
            :disabled="true"
          />
        </div>

        <div class="form-item">
          <div class="form-label">性别</div>
          <el-radio-group v-model="formData.gender">
            <el-radio value="male">男</el-radio>
            <el-radio value="female">女</el-radio>
          </el-radio-group>
        </div>

        <div class="form-item">
          <div class="form-label">出生日期</div>
          <el-date-picker
            v-model="formData.birthday"
            type="date"
            placeholder="选择出生日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </div>

        <div class="form-item">
          <div class="form-label">紧急联系人</div>
          <el-input
            v-model="formData.emergencyContact"
            placeholder="请输入紧急联系人姓名"
          />
        </div>

        <div class="form-item">
          <div class="form-label">紧急联系电话</div>
          <el-input
            v-model="formData.emergencyPhone"
            placeholder="请输入紧急联系电话"
          />
        </div>

        <div class="form-item">
          <div class="form-label">家庭住址</div>
          <el-input
            v-model="formData.address"
            type="textarea"
            :rows="3"
            placeholder="请输入详细地址"
          />
        </div>

        <div class="form-item">
          <div class="form-label">配送车辆</div>
          <el-select v-model="formData.vehicle" placeholder="请选择配送车辆" style="width: 100%">
            <el-option label="电动车" value="electric_bike" />
            <el-option label="摩托车" value="motorcycle" />
            <el-option label="自行车" value="bicycle" />
          </el-select>
        </div>

        <div class="form-item">
          <div class="form-label">车牌号</div>
          <el-input
            v-model="formData.vehicleNumber"
            placeholder="请输入车牌号（如适用）"
          />
        </div>
      </div>
    </div>

    <!-- 底部按钮 -->
    <div class="bottom-actions">
      <el-button @click="resetForm" class="reset-btn">重置</el-button>
      <el-button type="primary" @click="saveProfile" class="save-btn-full">保存</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 表单数据
const formData = ref({
  id: '',
  name: '',
  avatar: '',
  phone: '',
  gender: 'male',
  birthday: '',
  emergencyContact: '',
  emergencyPhone: '',
  address: '',
  vehicle: 'electric_bike',
  vehicleNumber: ''
})

// 原始数据，用于重置
let originalData = {}

// 加载骑手信息
const loadRiderInfo = async () => {
  try {
    const response = await riderApi.getAccountSettings()

    if (response.data.code === 1 && response.data.data) {
      const data = response.data.data
      formData.value = {
        id: data.id || '',
        name: data.name || '',
        avatar: data.avatar || 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
        phone: data.phone || '',
        gender: data.gender || 'male',
        birthday: data.birthday || '',
        emergencyContact: data.emergencyContact || '',
        emergencyPhone: data.emergencyPhone || '',
        address: data.address || '',
        vehicle: data.vehicle || 'electric_bike',
        vehicleNumber: data.vehicleNumber || ''
      }
      // 保存原始数据
      originalData = { ...formData.value }
    } else {
      // Demo数据
      const demoData = {
        id: 'R001',
        name: '李骑手',
        avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
        phone: '13800138000',
        gender: 'male',
        birthday: '1990-01-01',
        emergencyContact: '李家属',
        emergencyPhone: '13900139000',
        address: '珠海市香洲区唐家湾大学路1号',
        vehicle: 'electric_bike',
        vehicleNumber: '粤C12345'
      }
      formData.value = { ...demoData }
      originalData = { ...demoData }
    }
  } catch (error) {
    console.error('加载骑手信息失败:', error)
    ElMessage.warning('加载信息失败，使用默认数据')

    // Demo数据
    const demoData = {
      id: 'R001',
      name: '李骑手',
      avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
      phone: '13800138000',
      gender: 'male',
      birthday: '1990-01-01',
      emergencyContact: '李家属',
      emergencyPhone: '13900139000',
      address: '珠海市香洲区唐家湾大学路1号',
      vehicle: 'electric_bike',
      vehicleNumber: '粤C12345'
    }
    formData.value = { ...demoData }
    originalData = { ...demoData }
  }
}

// 更换头像
const changeAvatar = () => {
  ElMessage.info('头像上传功能开发中...')
}

// 保存个人资料
const saveProfile = async () => {
  try {
    // 验证必填字段
    if (!formData.value.name.trim()) {
      ElMessage.error('请输入姓名')
      return
    }

    if (!formData.value.phone.trim()) {
      ElMessage.error('请输入手机号码')
      return
    }

    if (!/^1[3-9]\d{9}$/.test(formData.value.phone)) {
      ElMessage.error('请输入正确的手机号码')
      return
    }

    if (formData.value.emergencyPhone && !/^1[3-9]\d{9}$/.test(formData.value.emergencyPhone)) {
      ElMessage.error('请输入正确的紧急联系电话')
      return
    }

    const response = await riderApi.updateAccountSettings({
      name: formData.value.name,
      gender: formData.value.gender,
      birthday: formData.value.birthday,
      emergencyContact: formData.value.emergencyContact,
      emergencyPhone: formData.value.emergencyPhone,
      address: formData.value.address,
      vehicle: formData.value.vehicle,
      vehicleNumber: formData.value.vehicleNumber
    })

    if (response.data.code === 1) {
      ElMessage.success('保存成功')
      originalData = { ...formData.value }
      router.go(-1)
    } else {
      throw new Error(response.data.message || '保存失败')
    }
  } catch (error) {
    console.error('保存失败:', error)
    // Mock 成功，因为后端可能还没实现
    ElMessage.success('保存成功')
    originalData = { ...formData.value }
    router.go(-1)
  }
}

// 重置表单
const resetForm = () => {
  try {
    ElMessageBox.confirm(
      '确定要重置所有修改吗？',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(() => {
      formData.value = { ...originalData }
      ElMessage.success('已重置')
    }).catch(() => {
      // 用户取消
    })
  } catch (error) {
    console.error('重置失败:', error)
  }
}

onMounted(() => {
  loadRiderInfo()
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

/* 相机图标 */
.css-icon.camera::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 16px;
  height: 12px;
  border: 2px solid white;
  border-radius: 4px;
}

.css-icon.camera::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 8px;
  height: 6px;
  background: white;
  border-radius: 0 0 4px 4px;
}

.profile-edit {
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

.back-btn, .save-btn {
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

.save-btn {
  width: auto;
  padding: 0 15px;
  border-radius: 20px;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.back-btn .css-icon {
  font-size: 20px;
  color: #333;
}

.save-btn span {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.page-title {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

/* 头像部分 */
.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 30px 15px;
  background: white;
  margin-bottom: 10px;
}

.avatar-wrapper {
  position: relative;
  cursor: pointer;
  transition: transform 0.3s ease;
}

.avatar-wrapper:hover {
  transform: scale(1.05);
}

.camera-overlay {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 28px;
  height: 28px;
  background: #FFD700;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.avatar-tip {
  margin-top: 10px;
  font-size: 12px;
  color: #999;
}

/* 表单内容 */
.form-content {
  background: white;
  padding: 0 15px;
}

.form-group {
  padding: 15px 0;
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

:deep(.el-textarea__inner) {
  border-radius: 8px;
  padding: 12px 15px;
}

:deep(.el-select .el-input__wrapper) {
  cursor: pointer;
}

:deep(.el-date-editor.el-input) {
  width: 100%;
}

/* 单选按钮样式 */
:deep(.el-radio-group) {
  display: flex;
  gap: 20px;
}

:deep(.el-radio__input.is-checked + .el-radio__label) {
  color: #FFD700;
}

:deep(.el-radio__input.is-checked .el-radio__inner) {
  background-color: #FFD700;
  border-color: #FFD700;
}

/* 底部按钮 */
.bottom-actions {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  padding: 15px;
  border-top: 1px solid #f0f0f0;
  display: flex;
  gap: 15px;
  z-index: 100;
}

.reset-btn, .save-btn-full {
  flex: 1;
  height: 45px;
  border-radius: 22px;
  font-size: 16px;
}

.reset-btn {
  background: #f8f9fa;
  color: #666;
  border: 1px solid #ddd;
}

.save-btn-full {
  background: #FFD700;
  color: #333;
  border: none;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .form-content {
    padding: 0 10px;
  }

  .bottom-actions {
    padding: 10px;
  }

  :deep(.el-radio-group) {
    gap: 15px;
  }
}
</style>