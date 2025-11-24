import request from '../merchant/request'

/**
 * 骑手端接口说明
 * 以下接口用于骑手接单、取货、配送、收入统计等场景
 * 后端实现时请保证订单状态与用户端、商家端的联动一致
 */

// ==================== 骑手信息相关 ====================

/**
 * getRiderInfo()
 * 功能：获取骑手个人信息
 * 请求：GET /rider/info
 * 返回示例：{ code:1, data:{ id, name, avatar, phone, rating, completedOrders, isOnline } }
 */
export function getRiderInfo() {
  return request({ url: '/rider/info', method: 'get' })
}

/**
 * updateOnlineStatus(isOnline)
 * 功能：更新骑手在线状态
 * 请求：POST /rider/status
 * 请求体：{ isOnline: boolean }
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function updateOnlineStatus(isOnline: boolean) {
  return request({ url: '/rider/status', method: 'post', data: { isOnline } })
}

// ==================== 订单相关 ====================

/**
 * getNewOrders()
 * 功能：获取新订单列表（待接单）
 * 请求：GET /rider/orders/new
 * 返回示例：{ code:1, data:[{ id, restaurant, pickupAddress, customer, deliveryAddress, distance, estimatedFee, estimatedTime, createdAt }] }
 */
export function getNewOrders() {
  return request({ url: '/rider/orders/new', method: 'get' })
}

/**
 * acceptOrder(orderId)
 * 功能：骑手接单
 * 请求：POST /rider/orders/{orderId}/accept
 * 返回示例：{ code:1, data:{ success:true, pickupCode:'A123' } }
 */
export function acceptOrder(orderId: string) {
  return request({ url: `/rider/orders/${orderId}/accept`, method: 'post' })
}

/**
 * getPickupOrders()
 * 功能：获取待取货订单列表
 * 请求：GET /rider/orders/pickup
 * 返回示例：{ code:1, data:[{ id, restaurant, pickupAddress, pickupCode, shopPhone, remainingTime }] }
 */
export function getPickupOrders() {
  return request({ url: '/rider/orders/pickup', method: 'get' })
}

/**
 * confirmPickup(orderId)
 * 功能：确认取货
 * 请求：POST /rider/orders/{orderId}/pickup
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function confirmPickup(orderId: string) {
  return request({ url: `/rider/orders/${orderId}/pickup`, method: 'post' })
}

/**
 * getDeliveringOrders()
 * 功能：获取配送中订单列表
 * 请求：GET /rider/orders/delivering
 * 返回示例：{ code:1, data:[{ id, customer, customerPhone, customerAvatar, deliveryAddress, remainingTime }] }
 */
export function getDeliveringOrders() {
  return request({ url: '/rider/orders/delivering', method: 'get' })
}

/**
 * completeDelivery(orderId)
 * 功能：完成配送
 * 请求：POST /rider/orders/{orderId}/complete
 * 返回示例：{ code:1, data:{ success:true, actualFee:6.5 } }
 */
export function completeDelivery(orderId: string) {
  return request({ url: `/rider/orders/${orderId}/complete`, method: 'post' })
}

/**
 * getOrderDetail(orderId)
 * 功能：获取订单详情
 * 请求：GET /rider/orders/{orderId}
 * 返回示例：{ code:1, data:{ id, items[], customerInfo, shopInfo, total, status, timeline[] } }
 */
export function getOrderDetail(orderId: string) {
  return request({ url: `/rider/orders/${orderId}`, method: 'get' })
}

// ==================== 收入统计相关 ====================

/**
 * getIncomeStats()
 * 功能：获取收入统计
 * 请求：GET /rider/income/stats
 * 查询参数：{ period?: 'today|week|month' }
 * 返回示例：{ code:1, data:{ dailyIncome:185.5, weeklyIncome:1280, monthlyIncome:5200, completedOrders:68 } }
 */
export function getIncomeStats(params?: { period?: string }) {
  return request({ url: '/rider/income/stats', method: 'get', params })
}

/**
 * getIncomeHistory()
 * 功能：获取收入明细
 * 请求：GET /rider/income/history
 * 查询参数：{ page:1, size:20, startDate?, endDate? }
 * 返回示例：{ code:1, data:{ items:[{ id, orderId, amount, type, time, remark }], total } }
 */
export function getIncomeHistory(params?: { page?: number; size?: number; startDate?: string; endDate?: string }) {
  return request({ url: '/rider/income/history', method: 'get', params })
}

/**
 * getWeeklyStats()
 * 功能：获取本周统计数据
 * 请求：GET /rider/stats/weekly
 * 返回示例：{ code:1, data:{ weekIncome:1280, weekOrders:68, onlineHours:45, avgRating:4.8 } }
 */
export function getWeeklyStats() {
  return request({ url: '/rider/stats/weekly', method: 'get' })
}

// ==================== 历史订单相关 ====================

/**
 * getOrderHistory()
 * 功能：获取历史订单列表
 * 请求：GET /rider/orders/history
 * 查询参数：{ page:1, size:20, status?, date? }
 * 返回示例：{ code:1, data:{ items:[{ id, restaurant, customer, fee, status, completedAt }], total } }
 */
export function getOrderHistory(params?: { page?: number; size?: number; status?: string; date?: string }) {
  return request({ url: '/rider/orders/history', method: 'get', params })
}

// ==================== 钱包相关 ====================

/**
 * getWalletInfo()
 * 功能：获取钱包信息
 * 请求：GET /rider/wallet
 * 返回示例：{ code:1, data:{ balance:2580.5, frozenAmount:120, totalIncome:15680 } }
 */
export function getWalletInfo() {
  return request({ url: '/rider/wallet', method: 'get' })
}

/**
 * withdraw(amount)
 * 功能：提现申请
 * 请求：POST /rider/wallet/withdraw
 * 请求体：{ amount: number, account:string }
 * 返回示例：{ code:1, data:{ success:true, withdrawId:'w1' } }
 */
export function withdraw(data: { amount: number; account: string }) {
  return request({ url: '/rider/wallet/withdraw', method: 'post', data })
}

/**
 * getWithdrawHistory()
 * 功能：获取提现记录
 * 请求：GET /rider/wallet/withdraw/history
 * 返回示例：{ code:1, data:[{ id, amount, status, appliedAt, processedAt }] }
 */
export function getWithdrawHistory() {
  return request({ url: '/rider/wallet/withdraw/history', method: 'get' })
}

// ==================== 位置相关 ====================

/**
 * updateLocation(location)
 * 功能：更新骑手位置
 * 请求：POST /rider/location
 * 请求体：{ latitude:number, longitude:number, address:string }
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function updateLocation(location: { latitude: number; longitude: number; address: string }) {
  return request({ url: '/rider/location', method: 'post', data: location })
}

/**
 * getDeliveryRoute(orderId)
 * 功能：获取配送路线
 * 请求：GET /rider/orders/{orderId}/route
 * 返回示例：{ code:1, data:{ route:[{lat,lng}], distance:1200, estimatedTime:15 } }
 */
export function getDeliveryRoute(orderId: string) {
  return request({ url: `/rider/orders/${orderId}/route`, method: 'get' })
}

// ==================== Demo数据 ====================

// Demo数据供后端不可用时使用
const demoRiderInfo = {
  id: 'rider001',
  name: '李骑手',
  avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
  phone: '13800138000',
  rating: 4.8,
  completedOrders: 1250,
  isOnline: true
}

const demoNewOrders = [
  {
    id: 'order001',
    restaurant: '麦当劳',
    pickupAddress: '珠海市香洲区唐家湾大学路1号',
    customer: '张同学',
    deliveryAddress: '珠海市香洲区中山大学珠海校区榕园',
    distance: 1.2,
    estimatedFee: 6.5,
    estimatedTime: 20,
    createdAt: new Date().toISOString()
  },
  {
    id: 'order002',
    restaurant: '肯德基',
    pickupAddress: '珠海市香洲区唐家湾大学路101号',
    customer: '王老师',
    deliveryAddress: '珠海市香洲区中山大学珠海校区荔园',
    distance: 0.8,
    estimatedFee: 5.0,
    estimatedTime: 15,
    createdAt: new Date().toISOString()
  }
]

const demoIncomeStats = {
  dailyIncome: 185.5,
  weeklyIncome: 1280,
  monthlyIncome: 5200,
  completedOrders: 68,
  estimatedIncome: 45
}

// Demo数据回退函数
function ensureDemoRiderInfo() {
  return Promise.resolve({ data: { code: 1, data: demoRiderInfo } })
}

function ensureDemoNewOrders() {
  return Promise.resolve({ data: { code: 1, data: demoNewOrders } })
}

function ensureDemoIncomeStats() {
  return Promise.resolve({ data: { code: 1, data: demoIncomeStats } })
}

// 带Demo数据回退的接口包装
export async function getRiderInfoWithDemo() {
  try {
    const res = await getRiderInfo()
    return res.data
  } catch (e) {
    console.warn('骑手信息接口不可用，使用Demo数据')
    return ensureDemoRiderInfo().then(res => res.data)
  }
}

export async function getNewOrdersWithDemo() {
  try {
    const res = await getNewOrders()
    return res.data
  } catch (e) {
    console.warn('新订单接口不可用，使用Demo数据')
    return ensureDemoNewOrders().then(res => res.data)
  }
}

export async function getIncomeStatsWithDemo() {
  try {
    const res = await getIncomeStats()
    return res.data
  } catch (e) {
    console.warn('收入统计接口不可用，使用Demo数据')
    return ensureDemoIncomeStats().then(res => res.data)
  }
}

export default {
  // 基础接口
  getRiderInfo,
  updateOnlineStatus,
  getNewOrders,
  acceptOrder,
  getPickupOrders,
  confirmPickup,
  getDeliveringOrders,
  completeDelivery,
  getOrderDetail,
  getIncomeStats,
  getIncomeHistory,
  getWeeklyStats,
  getOrderHistory,
  getWalletInfo,
  withdraw,
  getWithdrawHistory,
  updateLocation,
  getDeliveryRoute,

  // 带Demo数据回退的接口
  getRiderInfoWithDemo,
  getNewOrdersWithDemo,
  getIncomeStatsWithDemo
}