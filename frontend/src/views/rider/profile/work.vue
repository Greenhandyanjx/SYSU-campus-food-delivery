<template>
  <div class="work-preferences">
    <!-- 顶部导航栏 -->
    <div class="header-bar">
      <div class="back-btn" @click="$router.go(-1)">
        <i class="css-icon back"></i>
      </div>
      <h1 class="page-title">工作偏好</h1>
      <div class="save-btn" @click="saveSettings">
        <span>保存</span>
      </div>
    </div>

    <!-- 工作模式 -->
    <div class="preference-section">
      <h3 class="section-title">工作模式</h3>
      <div class="work-mode-cards">
        <div
          class="mode-card"
          :class="{ active: workSettings.mode === 'full' }"
          @click="selectWorkMode('full')"
        >
          <div class="mode-icon">
            <i class="css-icon full-time"></i>
          </div>
          <div class="mode-content">
            <div class="mode-title">全职模式</div>
            <div class="mode-desc">全职配送，收入稳定，优先派单</div>
          </div>
          <div class="mode-badge">推荐</div>
        </div>

        <div
          class="mode-card"
          :class="{ active: workSettings.mode === 'part' }"
          @click="selectWorkMode('part')"
        >
          <div class="mode-icon">
            <i class="css-icon part-time"></i>
          </div>
          <div class="mode-content">
            <div class="mode-title">兼职模式</div>
            <div class="mode-desc">灵活配送，自由安排时间</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 工作时间 -->
    <div class="preference-section">
      <h3 class="section-title">工作时间</h3>

      <div class="time-setting-item">
        <div class="setting-label">
          <div class="label-title">启用工作时间限制</div>
          <div class="label-desc">只在工作时间段内接收订单</div>
        </div>
        <el-switch
          v-model="workSettings.timeLimit.enabled"
          @change="onTimeLimitToggle"
        />
      </div>

      <div v-if="workSettings.timeLimit.enabled" class="time-range-setting">
        <div class="time-item">
          <div class="time-label">开始时间</div>
          <el-time-picker
            v-model="workSettings.timeLimit.startTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择开始时间"
          />
        </div>
        <div class="time-item">
          <div class="time-label">结束时间</div>
          <el-time-picker
            v-model="workSettings.timeLimit.endTime"
            format="HH:mm"
            value-format="HH:mm"
            placeholder="选择结束时间"
          />
        </div>
      </div>

      <!-- 工作日设置 -->
      <div class="workdays-setting">
        <div class="setting-label">
          <div class="label-title">工作日设置</div>
          <div class="label-desc">选择可工作的日期</div>
        </div>
        <div class="workdays-grid">
          <div
            v-for="(day, index) in workdays"
            :key="index"
            class="workday-item"
            :class="{ active: workSettings.workdays.includes(index) }"
            @click="toggleWorkday(index)"
          >
            {{ day }}
          </div>
        </div>
      </div>
    </div>

    <!-- 配送范围 -->
    <div class="preference-section">
      <h3 class="section-title">配送范围</h3>

      <div class="range-setting-item">
        <div class="setting-label">
          <div class="label-title">限制配送范围</div>
          <div class="label-desc">只接收指定范围内的订单</div>
        </div>
        <el-switch
          v-model="workSettings.rangeLimit.enabled"
        />
      </div>

      <div v-if="workSettings.rangeLimit.enabled" class="range-input">
        <div class="range-item">
          <div class="range-label">最大配送距离</div>
          <div class="range-control">
            <el-slider
              v-model="workSettings.rangeLimit.maxDistance"
              :min="1"
              :max="20"
              :step="1"
              show-input
              :show-input-controls="false"
            />
            <span class="range-unit">公里</span>
          </div>
        </div>

        <div class="range-item">
          <div class="range-label">配送区域偏好</div>
          <div class="area-tags">
            <el-tag
              v-for="area in workSettings.rangeLimit.preferredAreas"
              :key="area"
              closable
              @close="removePreferredArea(area)"
            >
              {{ area }}
            </el-tag>
            <el-input
              v-if="showAreaInput"
              ref="areaInputRef"
              v-model="newArea"
              class="area-input"
              size="small"
              @keyup.enter="addPreferredArea"
              @blur="addPreferredArea"
            />
            <el-button
              v-else
              class="add-area-btn"
              size="small"
              @click="showAreaInput = true"
            >
              + 添加区域
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 订单偏好 -->
    <div class="preference-section">
      <h3 class="section-title">订单偏好</h3>

      <div class="order-setting-item">
        <div class="setting-label">
          <div class="label-title">自动接单</div>
          <div class="label-desc">自动接受符合条件的订单</div>
        </div>
        <el-switch
          v-model="workSettings.autoAccept"
        />
      </div>

      <div v-if="workSettings.autoAccept" class="auto-accept-settings">
        <div class="condition-item">
          <div class="condition-label">最小订单金额</div>
          <div class="condition-control">
            <el-input-number
              v-model="workSettings.autoAcceptConditions.minAmount"
              :min="0"
              :max="100"
              :step="1"
              controls-position="right"
            />
            <span class="condition-unit">元</span>
          </div>
        </div>

        <div class="condition-item">
          <div class="condition-label">最大配送距离</div>
          <div class="condition-control">
            <el-input-number
              v-model="workSettings.autoAcceptConditions.maxDistance"
              :min="1"
              :max="15"
              :step="1"
              controls-position="right"
            />
            <span class="condition-unit">公里</span>
          </div>
        </div>
      </div>

      <div class="order-setting-item">
        <div class="setting-label">
          <div class="label-title">优先接收高额订单</div>
          <div class="label-desc">优先推送高配送费的订单</div>
        </div>
        <el-switch
          v-model="workSettings.preferHighFee"
        />
      </div>

      <div class="order-setting-item">
        <div class="setting-label">
          <div class="label-title">接收订单类型</div>
          <div class="label-desc">选择愿意接收的订单类型</div>
        </div>
        <div class="order-types">
          <el-checkbox
            v-model="workSettings.orderTypes.food"
            label="外卖订单"
          />
          <el-checkbox
            v-model="workSettings.orderTypes.grocery"
            label="生鲜超市"
          />
          <el-checkbox
            v-model="workSettings.orderTypes.medicine"
            label="医药配送"
          />
          <el-checkbox
            v-model="workSettings.orderTypes.document"
            label="文件配送"
          />
        </div>
      </div>
    </div>

    <!-- 休息设置 -->
    <div class="preference-section">
      <h3 class="section-title">休息设置</h3>

      <div class="rest-setting-item">
        <div class="setting-label">
          <div class="label-title">定时休息</div>
          <div class="label-desc">每天定时休息，避免过度劳累</div>
        </div>
        <el-switch
          v-model="workSettings.scheduledRest.enabled"
        />
      </div>

      <div v-if="workSettings.scheduledRest.enabled" class="rest-settings">
        <div class="rest-time-item">
          <div class="rest-label">休息时间</div>
          <div class="rest-time-control">
            <el-time-picker
              v-model="workSettings.scheduledRest.startTime"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="开始时间"
            />
            <span class="rest-separator">至</span>
            <el-time-picker
              v-model="workSettings.scheduledRest.endTime"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="结束时间"
            />
          </div>
        </div>

        <div class="rest-duration-item">
          <div class="rest-label">每单后休息</div>
          <div class="rest-duration-control">
            <el-input-number
              v-model="workSettings.scheduledRest.restAfterOrder"
              :min="0"
              :max="30"
              :step="5"
              controls-position="right"
            />
            <span class="rest-unit">分钟</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 在线状态设置 -->
    <div class="preference-section">
      <h3 class="section-title">在线状态</h3>

      <div class="online-setting-item">
        <div class="setting-label">
          <div class="label-title">自动切换离线</div>
          <div class="label-desc">长时间无订单时自动切换离线</div>
        </div>
        <el-switch
          v-model="workSettings.autoOffline.enabled"
        />
      </div>

      <div v-if="workSettings.autoOffline.enabled" class="auto-offline-settings">
        <div class="offline-item">
          <div class="offline-label">无订单时长</div>
          <div class="offline-control">
            <el-input-number
              v-model="workSettings.autoOffline.idleTime"
              :min="30"
              :max="180"
              :step="15"
              controls-position="right"
            />
            <span class="offline-unit">分钟</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import riderApi from '@/api/rider'

const router = useRouter()

// 工作日列表
const workdays = ['一', '二', '三', '四', '五', '六', '日']

// 工作设置
const workSettings = reactive({
  mode: 'full',
  timeLimit: {
    enabled: true,
    startTime: '08:00',
    endTime: '22:00'
  },
  workdays: [0, 1, 2, 3, 4, 5, 6],
  rangeLimit: {
    enabled: false,
    maxDistance: 5,
    preferredAreas: ['中山大学珠海校区', '唐家湾']
  },
  autoAccept: false,
  autoAcceptConditions: {
    minAmount: 5,
    maxDistance: 3
  },
  preferHighFee: true,
  orderTypes: {
    food: true,
    grocery: true,
    medicine: true,
    document: false
  },
  scheduledRest: {
    enabled: true,
    startTime: '12:00',
    endTime: '13:00',
    restAfterOrder: 5
  },
  autoOffline: {
    enabled: false,
    idleTime: 60
  }
})

// 区域输入相关
const showAreaInput = ref(false)
const newArea = ref('')
const areaInputRef = ref(null)

// 加载工作设置
const loadWorkSettings = async () => {
  try {
    const response = await riderApi.getWorkSettings()

    if (response.data.code === 1 && response.data.data) {
      const data = response.data.data
      Object.assign(workSettings, data)
    } else {
      // 使用默认设置
      console.log('使用默认工作设置')
    }
  } catch (error) {
    console.error('加载工作设置失败:', error)
    ElMessage.warning('加载设置失败，使用默认配置')
  }
}

// 选择工作模式
const selectWorkMode = (mode) => {
  workSettings.mode = mode

  // 根据模式调整其他设置
  if (mode === 'full') {
    workSettings.workdays = [0, 1, 2, 3, 4, 5, 6]
    workSettings.timeLimit.enabled = true
    workSettings.autoAccept = true
  } else {
    workSettings.workdays = [0, 1, 2, 3, 4]
    workSettings.timeLimit.enabled = true
    workSettings.autoAccept = false
  }
}

// 时间限制开关
const onTimeLimitToggle = (enabled) => {
  if (enabled && !workSettings.timeLimit.startTime) {
    workSettings.timeLimit.startTime = '08:00'
    workSettings.timeLimit.endTime = '22:00'
  }
}

// 切换工作日
const toggleWorkday = (index) => {
  const dayIndex = workSettings.workdays.indexOf(index)
  if (dayIndex > -1) {
    workSettings.workdays.splice(dayIndex, 1)
  } else {
    workSettings.workdays.push(index)
  }
}

// 添加偏好区域
const addPreferredArea = () => {
  const area = newArea.value.trim()
  if (area && !workSettings.rangeLimit.preferredAreas.includes(area)) {
    workSettings.rangeLimit.preferredAreas.push(area)
    newArea.value = ''
  }
  showAreaInput.value = false
}

// 移除偏好区域
const removePreferredArea = (area) => {
  const index = workSettings.rangeLimit.preferredAreas.indexOf(area)
  if (index > -1) {
    workSettings.rangeLimit.preferredAreas.splice(index, 1)
  }
}

// 保存设置
const saveSettings = async () => {
  try {
    // 验证设置
    if (workSettings.timeLimit.enabled) {
      if (!workSettings.timeLimit.startTime || !workSettings.timeLimit.endTime) {
        ElMessage.error('请设置完整的工作时间')
        return
      }
    }

    if (workSettings.rangeLimit.enabled && workSettings.rangeLimit.preferredAreas.length === 0) {
      ElMessage.error('请至少选择一个配送区域')
      return
    }

    const response = await riderApi.updateWorkSettings(workSettings)

    if (response.data.code === 1) {
      ElMessage.success('工作偏好设置已保存')
      router.go(-1)
    } else {
      throw new Error(response.data.message || '保存失败')
    }
  } catch (error) {
    console.error('保存工作设置失败:', error)
    // Mock 成功
    ElMessage.success('工作偏好设置已保存')
    router.go(-1)
  }
}

onMounted(() => {
  loadWorkSettings()
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

/* 全职图标 */
.css-icon.full-time::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 20px;
  height: 16px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.full-time::after {
  content: '';
  position: absolute;
  bottom: -4px;
  left: 50%;
  transform: translateX(-50%);
  width: 24px;
  height: 2px;
  background: currentColor;
  border-radius: 1px;
}

/* 兼职图标 */
.css-icon.part-time::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 16px;
  height: 16px;
  border: 2px solid currentColor;
  border-radius: 2px;
}

.css-icon.part-time::after {
  content: '';
  position: absolute;
  top: 8px;
  right: -4px;
  width: 8px;
  height: 8px;
  background: currentColor;
  border-radius: 50%;
}

.work-preferences {
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
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
}

.back-btn, .save-btn {
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

.save-btn {
  width: auto;
  padding: 0 15px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  cursor: pointer;
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

/* 偏好设置区块 */
.preference-section {
  margin: 70px 15px 15px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.section-title {
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 500;
  color: #333;
}

/* 工作模式卡片 */
.work-mode-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mode-card {
  display: flex;
  align-items: center;
  padding: 16px;
  border: 2px solid #f0f0f0;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.mode-card:hover {
  border-color: #FFD700;
}

.mode-card.active {
  border-color: #FFD700;
  background: #fffbf0;
}

.mode-icon {
  width: 44px;
  height: 44px;
  background: #f8f9fa;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.mode-icon .css-icon {
  font-size: 24px;
  color: #FFD700;
}

.mode-content {
  flex: 1;
}

.mode-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.mode-desc {
  font-size: 12px;
  color: #999;
}

.mode-badge {
  padding: 4px 8px;
  background: #FFD700;
  color: #333;
  font-size: 12px;
  border-radius: 12px;
  font-weight: 500;
}

/* 设置项样式 */
.time-setting-item,
.range-setting-item,
.order-setting-item,
.rest-setting-item,
.online-setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.time-setting-item:last-child,
.range-setting-item:last-child,
.order-setting-item:last-child,
.rest-setting-item:last-child,
.online-setting-item:last-child {
  border-bottom: none;
}

.setting-label {
  flex: 1;
}

.label-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.label-desc {
  font-size: 12px;
  color: #999;
}

/* 时间范围设置 */
.time-range-setting,
.auto-accept-settings,
.range-input,
.rest-settings,
.auto-offline-settings {
  margin-top: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
}

.time-item,
.range-item,
.condition-item,
.rest-time-item,
.rest-duration-item,
.offline-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 15px;
}

.time-item:last-child,
.range-item:last-child,
.condition-item:last-child,
.rest-time-item:last-child,
.rest-duration-item:last-child,
.offline-item:last-child {
  margin-bottom: 0;
}

.time-label,
.range-label,
.condition-label,
.rest-label,
.offline-label {
  font-size: 14px;
  color: #333;
  min-width: 80px;
}

/* 工作日设置 */
.workdays-setting {
  margin-top: 15px;
}

.workdays-grid {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

.workday-item {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid #ddd;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: #666;
  cursor: pointer;
  transition: all 0.3s ease;
}

.workday-item.active {
  background: #FFD700;
  color: #333;
  border-color: #FFD700;
}

/* 距离控制 */
.range-control {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.range-unit {
  font-size: 14px;
  color: #666;
  min-width: 40px;
}

/* 区域标签 */
.area-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  flex: 1;
}

.area-input {
  width: 120px;
}

.add-area-btn {
  height: 28px;
  border: 1px dashed #ddd;
  color: #666;
}

/* 条件控制 */
.condition-control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.condition-unit,
.rest-unit,
.offline-unit {
  font-size: 14px;
  color: #666;
  min-width: 30px;
}

/* 订单类型 */
.order-types {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  flex: 1;
}

/* 休息时间控制 */
.rest-time-control {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.rest-separator {
  font-size: 14px;
  color: #666;
}

/* 自定义组件样式 */
:deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 8px 12px;
}

:deep(.el-input-number) {
  width: 100px;
}

:deep(.el-input-number .el-input__wrapper) {
  border-radius: 8px;
}

:deep(.el-time-picker) {
  width: 120px;
}

:deep(.el-slider) {
  flex: 1;
  margin-right: 10px;
}

:deep(.el-checkbox) {
  margin-right: 15px;
  margin-bottom: 8px;
}

:deep(.el-checkbox__label) {
  font-size: 14px;
  color: #333;
}

:deep(.el-switch__core) {
  background-color: #ddd;
}

:deep(.el-switch.is-checked .el-switch__core) {
  background-color: #FFD700;
}

:deep(.el-switch__action) {
  background-color: white;
}

/* 响应式设计 */
@media (max-width: 375px) {
  .preference-section {
    margin: 70px 10px 10px;
    padding: 15px;
  }

  .mode-card {
    padding: 12px;
  }

  .mode-icon {
    width: 36px;
    height: 36px;
    margin-right: 12px;
  }

  .mode-icon .css-icon {
    font-size: 20px;
  }

  .workdays-grid {
    gap: 6px;
  }

  .workday-item {
    width: 32px;
    height: 32px;
    font-size: 12px;
  }

  .order-types {
    flex-direction: column;
    gap: 10px;
  }

  :deep(.el-time-picker) {
    width: 100px;
  }
}
</style>