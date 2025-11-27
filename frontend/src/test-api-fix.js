// 测试脚本：验证修复效果
import riderApi from '@/api/rider'
import { connectionHealth, networkUtils } from '@/utils/request'

async function testOrderHistoryAPI() {
  console.log('=== 开始测试历史订单API修复 ===')

  // 1. 检查连接状态
  console.log('1. 检查连接状态:')
  console.log('在线状态:', networkUtils.isOnline())
  console.log('连接健康:', connectionHealth.isHealthy())
  console.log('网络信息:', networkUtils.getConnectionInfo())
  console.log('推荐超时时间:', networkUtils.getRecommendedTimeout() + 'ms')

  // 2. 测试原始API（可能会超时）
  console.log('\n2. 测试原始API:')
  try {
    const originalResult = await riderApi.getOrderHistory({
      page: 1,
      pageSize: 5
    })
    console.log('原始API成功:', originalResult)
  } catch (error) {
    console.log('原始API失败:', error.message)
  }

  // 3. 测试增强API（应该自动回退到演示数据）
  console.log('\n3. 测试增强API:')
  try {
    const enhancedResult = await riderApi.getOrderHistoryEnhanced({
      page: 1,
      pageSize: 5
    })
    console.log('增强API结果:', {
      success: enhancedResult.success,
      hasData: !!enhancedResult.data,
      isFallback: enhancedResult.fallback,
      hasError: !!enhancedResult.error
    })

    if (enhancedResult.fallback) {
      console.log('✅ 成功回退到演示数据')
    }
  } catch (error) {
    console.log('增强API失败:', error.message)
  }

  // 4. 测试演示数据直接调用
  console.log('\n4. 测试演示数据:')
  try {
    const demoResult = await riderApi.getOrderHistoryWithDemo()
    console.log('演示数据数量:', demoResult.data.items.length)
    console.log('演示数据格式:', Object.keys(demoResult.data))
  } catch (error) {
    console.log('演示数据调用失败:', error.message)
  }

  console.log('\n=== 测试完成 ===')
}

// 测试工作统计API
async function testWorkStatsAPI() {
  console.log('\n=== 测试工作统计API ===')

  try {
    const result = await riderApi.getWorkStatsEnhanced()
    console.log('工作统计结果:', {
      success: result.success,
      hasData: !!result.data,
      isFallback: result.fallback,
      hasError: !!result.error
    })

    if (result.success && result.data) {
      console.log('统计数据:', {
        完成订单: result.data.data?.completedOrders,
        取消订单: result.data.data?.cancelledOrders,
        总收入: result.data.data?.totalIncome,
        效率: result.data.data?.efficiency
      })
    }
  } catch (error) {
    console.log('工作统计API失败:', error.message)
  }
}

// 导出测试函数
export { testOrderHistoryAPI, testWorkStatsAPI }

// 如果直接运行此文件
if (typeof window !== 'undefined') {
  window.testOrderHistoryAPI = testOrderHistoryAPI
  window.testWorkStatsAPI = testWorkStatsAPI
  console.log('测试函数已添加到 window 对象，可以在控制台中使用:')
  console.log('testOrderHistoryAPI() - 测试历史订单API')
  console.log('testWorkStatsAPI() - 测试工作统计API')
}