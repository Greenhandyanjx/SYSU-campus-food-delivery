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
 * 功能：获取新订单列表（待接单，status=1）
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
 * acceptOrderSafe(orderId)
 * 功能：骑手安全接单（带事务锁，防止并发接单）
 * 请求：POST /rider/orders/{orderId}/accept_safe
 * 返回示例：{ code:1, data:{ success:true, pickupCode:'A123' } }
 */
export function acceptOrderSafe(orderId: string) {
  return request({ url: `/rider/orders/${orderId}/accept_safe`, method: 'post' })
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
 * 功能：获取收入统计（根据时间段查询）
 * 请求：GET /rider/income/stats
 * 查询参数：{ period?: 'today|week|month' }
 * 返回示例：{ code:1, data:{ dailyIncome:185.5, weeklyIncome:1280, monthlyIncome:5200, completedOrders:68 } }
 */
export function getIncomeStats(params?: { period?: string }) {
  return request({ url: '/rider/income/stats', method: 'get', params })
}

/**
 * getTodayIncome()
 * 功能：获取今日收入统计
 * 请求：GET /rider/income/today
 * 返回示例：{ code:1, data:{ todayIncome:185.5, todayOrders:8 } }
 */
export function getTodayIncome() {
  return request({ url: '/rider/income/today', method: 'get' })
}

/**
 * getIncomeSummary()
 * 功能：获取收入汇总统计
 * 请求：GET /rider/income/summary
 * 返回示例：{ code:1, data:{ totalIncome:12580.5, completedOrders:156 } }
 */
export function getIncomeSummary() {
  return request({ url: '/rider/income/summary', method: 'get' })
}

/**
 * getMonthIncome()
 * 功能：获取月度收入数据
 * 请求：GET /rider/income/month
 * 返回示例：{ code:1, data:[{ date: '2024-01-01', money: 185.5 }] }
 */
export function getMonthIncome() {
  return request({ url: '/rider/income/month', method: 'get' })
}

/**
 * getIncomeHistory()
 * 功能：获取收入明细（分页）
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

/**
 * getRiderDashboard()
 * 功能：获取骑手仪表板数据
 * 请求：GET /rider/dashboard
 * 返回示例：{ code:1, data:{ todayIncome:185.5, todayOrders:8, delivering:2, waitPickup:1 } }
 */
export function getRiderDashboard() {
  return request({ url: '/rider/dashboard', method: 'get' })
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

// ==================== 认证相关 ====================

/**
 * getVerification()
 * 功能：获取骑手认证信息
 * 请求：GET /rider/verification
 * 返回示例：{ code:1, data:{ id, realName, idCard, phone, status, submitTime } }
 */
export function getVerification() {
  return request({ url: '/rider/verification', method: 'get' })
}

/**
 * submitVerification(data)
 * 功能：提交实名认证
 * 请求：POST /rider/verification
 * 请求体：{ realName, idCard, idCardFront, idCardBack, healthCert }
 * 返回示例：{ code:1, data:{ success:true, verificationId:'v1' } }
 */
export function submitVerification(data: {
  realName: string
  idCard: string
  idCardFront: string
  idCardBack: string
  healthCert: string
}) {
  return request({ url: '/rider/verification', method: 'post', data })
}

// ==================== 工作统计相关 ====================

/**
 * getWorkStats()
 * 功能：获取工作数据统计
 * 请求：GET /rider/stats/work
 * 查询参数：{ period?: 'today|week|month' }
 * 返回示例：{ code:1, data:{ totalOrders:68, totalIncome:1280.5, completionRate:95.8 } }
 */
export function getWorkStats(params?: { period?: string }) {
  return request({ url: '/rider/stats/work', method: 'get', params })
}

/**
 * getMonthlyStats()
 * 功能：获取月统计数据
 * 请求：GET /rider/stats/monthly
 * 返回示例：{ code:1, data:{ monthOrders:280, monthIncome:5200, onlineDays:25 } }
 */
export function getMonthlyStats() {
  return request({ url: '/rider/stats/monthly', method: 'get' })
}

// ==================== 收入明细相关 ====================

/**
 * getIncomeDetails()
 * 功能：获取收入明细
 * 请求：GET /rider/income/details
 * 查询参数：{ page:1, size:20, type?, startDate?, endDate? }
 * 返回示例：{ code:1, data:{ items:[{ id, orderId, amount, type, time, remark }], total } }
 */
export function getIncomeDetails(params?: {
  page?: number
  size?: number
  type?: string
  startDate?: string
  endDate?: string
}) {
  return request({ url: '/rider/income/details', method: 'get', params })
}

// ==================== 配送记录相关 ====================

/**
 * getDeliveryRecords()
 * 功能：获取配送记录
 * 请求：GET /rider/delivery/records
 * 查询参数：{ page:1, size:20, status?, startDate?, endDate? }
 * 返回示例：{ code:1, data:{ items:[{ id, orderNo, distance, duration, completedAt }], total } }
 */
export function getDeliveryRecords(params?: {
  page?: number
  size?: number
  status?: string
  startDate?: string
  endDate?: string
}) {
  return request({ url: '/rider/delivery/records', method: 'get', params })
}

// ==================== 评价相关 ====================

/**
 * getReviews()
 * 功能：获取用户评价
 * 请求：GET /rider/reviews
 * 查询参数：{ page:1, size:20 }
 * 返回示例：{ code:1, data:{ items:[{ id, orderId, rating, comment, createdAt }], avgRating:4.8 } }
 */
export function getReviews(params?: { page?: number; size?: number }) {
  return request({ url: '/rider/reviews', method: 'get', params })
}

// ==================== 排行榜相关 ====================

/**
 * getRanking()
 * 功能：获取排行榜数据
 * 请求：GET /rider/ranking/{type}
 * 路径参数：type - 'income' | 'orders' | 'rating' | 'efficiency'
 * 返回示例：{ code:1, data:[{ rank, name, avatar, value, isSelf }] }
 */
export function getRanking(type: string) {
  return request({ url: `/rider/ranking/${type}`, method: 'get' })
}

// ==================== 通知相关 ====================

/**
 * getNotifications()
 * 功能：获取通知列表
 * 请求：GET /rider/notifications
 * 查询参数：{ page:1, size:20, read?: boolean }
 * 返回示例：{ code:1, data:{ items:[{ id, title, content, type, isRead, createdAt }], unreadCount } }
 */
export function getNotifications(params?: { page?: number; size?: number; read?: boolean }) {
  return request({ url: '/rider/notifications', method: 'get', params })
}

/**
 * markNotificationRead()
 * 功能：标记通知已读
 * 请求：PUT /rider/notifications/{id}/read
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function markNotificationRead(id: string) {
  return request({ url: `/rider/notifications/${id}/read`, method: 'put' })
}

// ==================== 系统消息相关 ====================

/**
 * getSystemMessages()
 * 功能：获取系统消息
 * 请求：GET /rider/messages/system
 * 查询参数：{ page:1, size:20 }
 * 返回示例：{ code:1, data:{ items:[{ id, title, content, type, publishedAt }] } }
 */
export function getSystemMessages(params?: { page?: number; size?: number }) {
  return request({ url: '/rider/messages/system', method: 'get', params })
}

// ==================== 热力图相关 ====================

/**
 * getHeatmapData()
 * 功能：获取配送热力图数据
 * 请求：GET /rider/heatmap
 * 查询参数：{ start?, end? }
 * 返回示例：{ code:1, data:[{ date, areas:[{ lat, lng, count }] }] }
 */
export function getHeatmapData(params?: { start?: string; end?: string }) {
  return request({ url: '/rider/heatmap', method: 'get', params })
}

// ==================== 配送状态更新相关 ====================

/**
 * startDelivery()
 * 功能：开始配送
 * 请求：PUT /rider/orders/{orderId}/start
 * 返回示例：{ code:1, data:{ success:true, startTime } }
 */
export function startDelivery(orderId: string) {
  return request({ url: `/rider/orders/${orderId}/start`, method: 'put' })
}

/**
 * arrivePickup()
 * 功能：到达取餐点
 * 请求：PUT /rider/orders/{orderId}/arrive-pickup
 * 请求体：{ latitude?, longitude?, code? }
 * 返回示例：{ code:1, data:{ success:true, arrivedAt } }
 */
export function arrivePickup(orderId: string, data?: { latitude?: number; longitude?: number; code?: string }) {
  return request({ url: `/rider/orders/${orderId}/arrive-pickup`, method: 'put', data })
}

/**
 * updateDeliveryStatus()
 * 功能：更新配送状态
 * 请求：PUT /rider/orders/{orderId}/status
 * 请求体：{ status, latitude?, longitude?, note? }
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function updateDeliveryStatus(orderId: string, data: {
  status: string
  latitude?: number
  longitude?: number
  note?: string
}) {
  return request({ url: `/rider/orders/${orderId}/status`, method: 'put', data })
}

// ==================== 异常报告相关 ====================

/**
 * reportIssue()
 * 功能：异常情况报告
 * 请求：POST /rider/orders/{orderId}/issue
 * 请求体：{ type, description, images?, timestamp? }
 * 返回示例：{ code:1, data:{ success:true, issueId } }
 */
export function reportIssue(orderId: string, data: {
  type: string
  description: string
  images?: string[]
  timestamp?: number
}) {
  return request({ url: `/rider/orders/${orderId}/issue`, method: 'post', data })
}

// ==================== 设置相关 ====================

/**
 * getWorkSettings()
 * 功能：获取工作设置
 * 请求：GET /rider/settings/work
 * 返回示例：{ code:1, data:{ autoAccept, deliveryRange, workTime, restTime, maxOrders } }
 */
export function getWorkSettings() {
  return request({ url: '/rider/settings/work', method: 'get' })
}

/**
 * updateWorkSettings()
 * 功能：更新工作设置
 * 请求：PUT /rider/settings/work
 * 请求体：{ autoAccept?, deliveryRange?, workTime?, restTime?, maxOrders? }
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function updateWorkSettings(data: {
  autoAccept?: boolean
  deliveryRange?: number
  workTime?: { start: string; end: string }
  restTime?: { enabled: boolean; start: string; end: string }
  maxOrders?: number
}) {
  return request({ url: '/rider/settings/work', method: 'put', data })
}

/**
 * getAccountSettings()
 * 功能：获取账户设置
 * 请求：GET /rider/settings/account
 * 返回示例：{ code:1, data:{ phone, email, wechat, alipay, bankCard } }
 */
export function getAccountSettings() {
  return request({ url: '/rider/settings/account', method: 'get' })
}

/**
 * updateAccountSettings()
 * 功能：更新账户设置
 * 请求：PUT /rider/settings/account
 * 请求体：{ phone?, email?, wechat?, alipay?, bankCard? }
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function updateAccountSettings(data: {
  phone?: string
  email?: string
  wechat?: string
  alipay?: string
  bankCard?: string
}) {
  return request({ url: '/rider/settings/account', method: 'put', data })
}

/**
 * getNotificationSettings()
 * 功能：获取通知设置
 * 请求：GET /rider/settings/notification
 * 返回示例：{ code:1, data:{ orderNotification, systemNotification, soundEnabled, vibrationEnabled } }
 */
export function getNotificationSettings() {
  return request({ url: '/rider/settings/notification', method: 'get' })
}

/**
 * updateNotificationSettings()
 * 功能：更新通知设置
 * 请求：PUT /rider/settings/notification
 * 请求体：{ orderNotification?, systemNotification?, soundEnabled?, vibrationEnabled? }
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function updateNotificationSettings(data: {
  orderNotification?: boolean
  systemNotification?: boolean
  soundEnabled?: boolean
  vibrationEnabled?: boolean
}) {
  return request({ url: '/rider/settings/notification', method: 'put', data })
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
  acceptOrderSafe,
  getPickupOrders,
  confirmPickup,
  getDeliveringOrders,
  completeDelivery,
  getOrderDetail,
  getIncomeStats,
  getTodayIncome,
  getIncomeSummary,
  getMonthIncome,
  getIncomeHistory,
  getWeeklyStats,
  getRiderDashboard,
  getOrderHistory,
  getWalletInfo,
  withdraw,
  getWithdrawHistory,
  updateLocation,
  getDeliveryRoute,

  // 认证相关
  getVerification,
  submitVerification,

  // 工作统计相关
  getWorkStats,
  getMonthlyStats,

  // 收入明细相关
  getIncomeDetails,

  // 配送记录相关
  getDeliveryRecords,

  // 评价相关
  getReviews,

  // 排行榜相关
  getRanking,

  // 通知相关
  getNotifications,
  markNotificationRead,

  // 系统消息相关
  getSystemMessages,

  // 热力图相关
  getHeatmapData,

  // 配送状态更新相关
  startDelivery,
  arrivePickup,
  updateDeliveryStatus,

  // 异常报告相关
  reportIssue,

  // 设置相关
  getWorkSettings,
  updateWorkSettings,
  getAccountSettings,
  updateAccountSettings,
  getNotificationSettings,
  updateNotificationSettings,

  // 带Demo数据回退的接口
  getRiderInfoWithDemo,
  getNewOrdersWithDemo,
  getIncomeStatsWithDemo
}