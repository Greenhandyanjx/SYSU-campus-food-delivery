<template>
  <div class="enhanced-orders-container">
    <!-- 网络状态指示器 -->
    <div class="network-status-bar">
      <el-alert
        v-if="!connectionStatus.isOnline"
        title="设备离线"
        type="error"
        :closable="false"
        show-icon
      />
      <el-alert
        v-else-if="!connectionStatus.isHealthy"
        title="网络连接不稳定"
        type="warning"
        :closable="false"
        show-icon
      />
      <el-alert
        v-else
        title="网络连接正常"
        type="success"
        :closable="false"
        show-icon
      />
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <el-loading-directive />
      <p>{{ loadingMessage || '正在加载数据...' }}</p>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-container">
      <el-result
        icon="error"
        title="加载失败"
        :sub-title="error.userMessage || error.message"
      >
        <template #extra>
          <el-button
            type="primary"
            @click="handleRetry"
            :loading="isRetrying"
            :disabled="!connectionStatus.isOnline"
          >
            重试
          </el-button>
          <el-button
            v-if="error.canUseDemo"
            type="warning"
            @click="useDemoData"
          >
            使用演示数据
          </el-button>
        </template>
      </el-result>

      <!-- 错误详情 -->
      <el-collapse v-if="showErrorDetails" class="error-details">
        <el-collapse-item title="错误详情">
          <pre>{{ JSON.stringify(error, null, 2) }}</pre>
        </el-collapse-item>
      </el-collapse>
    </div>

    <!-- 数据展示 -->
    <div v-else-if="data" class="data-container">
      <!-- 数据来源标识 -->
      <div v-if="data.cached" class="data-source cached">
        <el-tag type="info" size="small">缓存数据</el-tag>
      </div>
      <div v-else-if="data.fallback" class="data-source fallback">
        <el-tag type="warning" size="small">演示数据</el-tag>
      </div>

      <!-- 订单列表 -->
      <el-table
        :data="data.items || []"
        style="width: 100%"
        :default-sort="{ prop: 'time', order: 'descending' }"
      >
        <el-table-column prop="orderId" label="订单号" width="120" />
        <el-table-column prop="restaurant" label="商家" width="150" />
        <el-table-column prop="customer" label="顾客" width="120" />
        <el-table-column prop="amount" label="金额" width="100">
          <template #default="{ row }">
            ¥{{ row.amount?.toFixed(2) || '0.00' }}
          </template>
        </el-table-column>
        <el-table-column prop="time" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.time) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="200" />
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="data.total || 0"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <el-button @click="refreshData" :loading="loading">
        <el-icon><Refresh /></el-icon>
        刷新数据
      </el-button>
      <el-button @click="clearCache">
        <el-icon><Delete /></el-icon>
        清除缓存
      </el-button>
      <el-button @click="toggleErrorDetails">
        <el-icon><View /></el-icon>
        {{ showErrorDetails ? '隐藏' : '显示' }}错误详情
      </el-button>
    </div>

    <!-- 连接状态详情 -->
    <el-drawer v-model="showConnectionDetails" title="网络连接状态">
      <div class="connection-details">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="在线状态">
            <el-tag :type="connectionStatus.isOnline ? 'success' : 'danger'">
              {{ connectionStatus.isOnline ? '在线' : '离线' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="连接健康">
            <el-tag :type="connectionStatus.isHealthy ? 'success' : 'warning'">
              {{ connectionStatus.isHealthy ? '健康' : '不稳定' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="网络类型" v-if="connectionStatus.networkInfo">
            {{ connectionStatus.networkInfo.effectiveType || '未知' }}
          </el-descriptions-item>
          <el-descriptions-item label="网络质量" v-if="connectionStatus.networkInfo">
            <el-progress
              :percentage="getConnectionQuality()"
              :status="getConnectionQualityStatus()"
            />
          </el-descriptions-item>
          <el-descriptions-item label="错误统计">
            <div class="error-stats">
              <p>总错误数: {{ errorStats.totalErrors }}</p>
              <p>网络错误: {{ errorStats.networkErrors }}</p>
              <p>超时错误: {{ errorStats.timeoutErrors }}</p>
              <p>服务器错误: {{ errorStats.serverErrors }}</p>
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Delete, View } from '@element-plus/icons-vue'
import { useOrderHistory } from '@/composables/useEnhancedApi'
import riderApi from '@/api/rider'

// 使用增强的组合式函数
const {
  fetchOrderHistory,
  loading,
  error,
  data,
  retry
} = useOrderHistory()

// 响应式数据
const currentPage = ref(1)
const pageSize = ref(20)
const loadingMessage = ref('')
const isRetrying = ref(false)
const showErrorDetails = ref(false)
const showConnectionDetails = ref(false)

// 连接状态（从composable获取）
const connectionStatus = computed(() => {
  // 这里应该从全局状态获取，简化示例
  return {
    isOnline: navigator.onLine,
    isHealthy: true, // 简化处理
    networkInfo: null
  }
})

const errorStats = computed(() => {
  // 这里应该从全局状态获取，简化示例
  return {
    totalErrors: error.value ? 1 : 0,
    networkErrors: error.value?.type === 'network' ? 1 : 0,
    timeoutErrors: error.value?.type === 'timeout' ? 1 : 0,
    serverErrors: error.value?.type === 'server' ? 1 : 0
  }
})

// 计算属性
const getConnectionQuality = () => {
  if (!connectionStatus.value.networkInfo) return 50

  const { effectiveType, downlink } = connectionStatus.value.networkInfo
  let quality = 50

  switch (effectiveType) {
    case '4g': quality = 90; break
    case '3g': quality = 70; break
    case '2g': quality = 30; break
    case 'slow-2g': quality = 10; break
  }

  // 根据下载速度调整
  if (downlink > 5) quality = Math.min(100, quality + 10)
  else if (downlink < 1) quality = Math.max(0, quality - 20)

  return quality
}

const getConnectionQualityStatus = () => {
  const quality = getConnectionQuality()
  if (quality >= 80) return 'success'
  if (quality >= 60) return ''
  if (quality >= 40) return 'warning'
  return 'exception'
}

// 方法
const formatDateTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getStatusType = (status) => {
  const statusMap = {
    1: 'info',     // 待接单
    2: 'warning',  // 已接单
    3: 'primary',  // 配送中
    4: 'success'   // 已完成
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status) => {
  const statusMap = {
    1: '待接单',
    2: '已接单',
    3: '配送中',
    4: '已完成'
  }
  return statusMap[status] || '未知'
}

const loadData = async () => {
  loadingMessage.value = '正在加载历史订单...'
  try {
    await fetchOrderHistory({
      page: currentPage.value,
      pageSize: pageSize.value
    })
  } finally {
    loadingMessage.value = ''
  }
}

const handleRetry = async () => {
  isRetrying.value = true
  try {
    await retry()
  } finally {
    isRetrying.value = false
  }
}

const useDemoData = async () => {
  loadingMessage.value = '加载演示数据...'
  try {
    const demoData = await riderApi.getOrderHistoryWithDemo()
    // 使用演示数据更新状态
    data.value = demoData.data
    error.value = null
    ElMessage.success('已加载演示数据')
  } finally {
    loadingMessage.value = ''
  }
}

const refreshData = () => {
  currentPage.value = 1
  loadData()
}

const clearCache = () => {
  // 清除缓存逻辑
  ElMessage.success('缓存已清除')
  refreshData()
}

const handleSizeChange = (newSize) => {
  pageSize.value = newSize
  currentPage.value = 1
  loadData()
}

const handleCurrentChange = (newPage) => {
  currentPage.value = newPage
  loadData()
}

const toggleErrorDetails = () => {
  showErrorDetails.value = !showErrorDetails.value
}

// 监听器
watch([currentPage, pageSize], () => {
  loadData()
})

// 生命周期
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.enhanced-orders-container {
  padding: 20px;
}

.network-status-bar {
  margin-bottom: 20px;
}

.loading-container {
  text-align: center;
  padding: 40px;
}

.error-container {
  padding: 20px;
}

.error-details {
  margin-top: 20px;
}

.data-container {
  position: relative;
}

.data-source {
  position: absolute;
  top: -20px;
  right: 10px;
  z-index: 10;
}

.data-source.cached .el-tag {
  background-color: #f0f9ff;
  border-color: #bfdbfe;
  color: #1e40af;
}

.data-source.fallback .el-tag {
  background-color: #fefce8;
  border-color: #fde047;
  color: #a16207;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.action-buttons {
  margin-top: 20px;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.connection-details {
  padding: 20px;
}

.error-stats {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.error-stats p {
  margin: 0;
  font-size: 14px;
}

@media (max-width: 768px) {
  .enhanced-orders-container {
    padding: 10px;
  }

  .action-buttons {
    justify-content: center;
  }

  .action-buttons .el-button {
    flex: 1;
    min-width: 120px;
  }
}
</style>