import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import { enhancedAPI, connectionHealth, networkUtils } from '@/utils/enhanced-api'

/**
 * Vue组合式函数：增强的API状态管理
 * - 提供统一的错误处理和用户反馈
 * - 自动重试和连接状态监控
 * - 加载状态和错误状态管理
 * - 支持演示数据回退
 */

export function useEnhancedApi(options = {}) {
  const {
    showErrorNotification = true,
    showSuccessMessage = false,
    autoRetry = true,
    cacheEnabled = true
  } = options

  // 响应式状态
  const loading = ref(false)
  const error = ref(null)
  const data = ref(null)
  const connectionStatus = reactive({
    isOnline: networkUtils.isOnline(),
    isHealthy: connectionHealth.isHealthy(),
    networkInfo: null,
    lastCheck: Date.now()
  })

  // 错误统计
  const errorStats = reactive({
    totalErrors: 0,
    networkErrors: 0,
    timeoutErrors: 0,
    serverErrors: 0,
    lastError: null
  })

  // 网络状态监听
  let onlineListener = null
  let offlineListener = null

  // 更新连接状态
  const updateConnectionStatus = () => {
    connectionStatus.isOnline = networkUtils.isOnline()
    connectionStatus.isHealthy = connectionHealth.isHealthy()
    connectionStatus.networkInfo = networkUtils.getConnectionInfo()
    connectionStatus.lastCheck = Date.now()
  }

  // 错误处理函数
  const handleError = (error, context = '') => {
    console.error('API Error handled by composable:', error)

    // 更新错误状态
    error.value = error
    errorStats.totalErrors++
    errorStats.lastError = {
      message: error.userMessage || error.message,
      type: error.type,
      time: Date.now(),
      context
    }

    // 错误分类统计
    switch (error.type) {
      case 'network':
        errorStats.networkErrors++
        break
      case 'timeout':
        errorStats.timeoutErrors++
        break
      case 'server':
        errorStats.serverErrors++
        break
    }

    // 显示用户友好的错误提示
    if (showErrorNotification) {
      const message = error.userMessage || '请求失败，请稍后重试'

      // 根据错误类型选择不同的展示方式
      if (error.isOffline) {
        ElNotification({
          title: '网络连接异常',
          message: '设备已离线，请检查网络设置',
          type: 'warning',
          duration: 5000,
          showClose: true
        })
      } else if (error.shouldRetry && autoRetry) {
        ElMessage({
          message: `${message}，正在自动重试...`,
          type: 'warning',
          duration: 3000
        })
      } else {
        ElMessage({
          message,
          type: 'error',
          duration: 5000
        })
      }
    }

    // 如果是演示数据，显示提示
    if (error.fallback || error.canUseDemo) {
      ElNotification({
        title: '使用演示数据',
        message: '由于网络问题，当前显示演示数据',
        type: 'info',
        duration: 4000,
        showClose: true
      })
    }
  }

  // 成功处理函数
  const handleSuccess = (result, context = '') => {
    data.value = result.data
    error.value = null
    loading.value = false

    if (showSuccessMessage && !result.cached && !result.fallback) {
      ElMessage({
        message: `${context}加载成功`,
        type: 'success',
        duration: 2000
      })
    }

    // 显示缓存状态提示
    if (result.cached) {
      console.log(`Using cached data for ${context}`)
    }

    return result
  }

  // 增强的请求执行函数
  const execute = async (apiCall, context = 'API请求') => {
    loading.value = true
    error.value = null

    try {
      const result = await enhancedAPI.enhancedRequest(apiCall, {
        context,
        useCache: cacheEnabled,
        customErrorHandler: (error) => {
          // 可以在这里自定义错误处理逻辑
          if (error.type === 'auth') {
            // 认证错误，可能需要跳转到登录页
            ElMessageBox.confirm(
              '登录状态已过期，请重新登录',
              '认证失败',
              {
                confirmButtonText: '重新登录',
                cancelButtonText: '取消',
                type: 'warning'
              }
            ).then(() => {
              // 这里可以添加跳转到登录页的逻辑
              console.log('Redirect to login page')
            }).catch(() => {
              // 用户取消
            })
            return { success: false, error, data: null }
          }
        }
      })

      return handleSuccess(result, context)

    } catch (err) {
      handleError(err, context)
      return { success: false, error: err, data: null }
    }
  }

  // 批量请求
  const executeBatch = async (apiCalls, context = '批量请求', options = {}) => {
    loading.value = true
    error.value = null

    try {
      const results = await enhancedAPI.batchRequests(apiCalls, {
        failFast: options.failFast || false
      })

      loading.value = false
      return results

    } catch (err) {
      handleError(err, context)
      return []
    }
  }

  // 重试函数
  const retry = async (apiCall, context = '重试请求') => {
    ElMessage({
      message: '正在重试请求...',
      type: 'info',
      duration: 1500
    })

    return execute(apiCall, context)
  }

  // 重置状态
  const reset = () => {
    loading.value = false
    error.value = null
    data.value = null
    Object.assign(errorStats, {
      totalErrors: 0,
      networkErrors: 0,
      timeoutErrors: 0,
      serverErrors: 0,
      lastError: null
    })
  }

  // 获取连接状态摘要
  const getConnectionSummary = computed(() => {
    return {
      ...connectionStatus,
      ...errorStats,
      connectionHealth: connectionHealth.getConnectionSummary()
    }
  })

  // 生命周期
  onMounted(() => {
    updateConnectionStatus()

    // 添加网络状态监听
    onlineListener = () => {
      connectionStatus.isOnline = true
      ElMessage({
        message: '网络连接已恢复',
        type: 'success',
        duration: 3000
      })
    }

    offlineListener = () => {
      connectionStatus.isOnline = false
      ElMessage({
        message: '网络连接已断开',
        type: 'warning',
        duration: 3000
      })
    }

    window.addEventListener('online', onlineListener)
    window.addEventListener('offline', offlineListener)

    // 定期更新连接状态
    const interval = setInterval(updateConnectionStatus, 30000)

    // 清理函数
    onUnmounted(() => {
      window.removeEventListener('online', onlineListener)
      window.removeEventListener('offline', offlineListener)
      clearInterval(interval)
    })
  })

  return {
    // 状态
    loading,
    error,
    data,
    connectionStatus,
    errorStats,

    // 计算属性
    getConnectionSummary,

    // 方法
    execute,
    executeBatch,
    retry,
    reset,
    updateConnectionStatus
  }
}

// 专门用于历史订单的组合函数
export function useOrderHistory() {
  const { execute, loading, error, data, retry } = useEnhancedApi({
    showErrorNotification: true,
    autoRetry: true,
    cacheEnabled: true
  })

  // 获取历史订单
  const fetchOrderHistory = async (params = {}) => {
    return execute({
      url: '/rider/orders/history',
      method: 'get',
      params
    }, '历史订单')
  }

  return {
    fetchOrderHistory,
    loading,
    error,
    data,
    retry
  }
}

// 专门用于新订单的组合函数
export function useNewOrders() {
  const { execute, loading, error, data, retry } = useEnhancedApi({
    showErrorNotification: true,
    autoRetry: true,
    cacheEnabled: false // 新订单不缓存
  })

  // 获取新订单
  const fetchNewOrders = async () => {
    return execute({
      url: '/rider/orders/new',
      method: 'get'
    }, '新订单')
  }

  return {
    fetchNewOrders,
    loading,
    error,
    data,
    retry
  }
}

// 专门用于工作台数据的组合函数
export function useDashboard() {
  const { execute, loading, error, data, retry } = useEnhancedApi({
    showErrorNotification: true,
    autoRetry: true,
    cacheEnabled: true
  })

  // 获取工作台数据
  const fetchDashboard = async () => {
    return execute({
      url: '/rider/dashboard',
      method: 'get'
    }, '工作台数据')
  }

  return {
    fetchDashboard,
    loading,
    error,
    data,
    retry
  }
}

// 专门用于收入统计的组合函数
export function useIncomeStats() {
  const { execute, loading, error, data, retry } = useEnhancedApi({
    showErrorNotification: true,
    autoRetry: true,
    cacheEnabled: true
  })

  // 获取收入统计
  const fetchIncomeStats = async () => {
    return execute({
      url: '/rider/income/stats',
      method: 'get'
    }, '收入统计')
  }

  return {
    fetchIncomeStats,
    loading,
    error,
    data,
    retry
  }
}