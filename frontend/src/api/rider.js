import request from '@/utils/request'
import { enhancedAPI } from '@/utils/enhanced-api'

// 骑手API接口
const riderApi = {
  // 基础信息相关
  // GET /rider/info
  getRiderInfo() {
    return request({
      url: '/rider/info',
      method: 'get'
    })
  },

  // POST /rider/status
  updateRiderStatus(data) {
    return request({
      url: '/rider/status',
      method: 'post',
      data
    })
  },

  // POST /rider/location
  updateRiderLocation(data) {
    return request({
      url: '/rider/location',
      method: 'post',
      data
    })
  },

  // 订单相关
  // GET /rider/orders/new
  getNewOrders() {
    return request({
      url: '/rider/orders/new',
      method: 'get'
    })
  },

  // 增强版本的新订单API（带演示数据回退）
  getNewOrdersEnhanced() {
    return enhancedAPI.wrapApiMethod(() => request({
      url: '/rider/orders/new',
      method: 'get'
    }), {
      context: '获取新订单',
      fallbackData: this.getNewOrdersWithDemo().data
    })();
  },

  // POST /rider/orders/:orderId/accept
  acceptOrder(orderId) {
    return request({
      url: `/rider/orders/${orderId}/accept`,
      method: 'post'
    })
  },

  // POST /rider/orders/:orderId/accept_safe
  acceptOrderSafe(orderId) {
    return request({
      url: `/rider/orders/${orderId}/accept_safe`,
      method: 'post'
    })
  },

  // POST /rider/orders/:orderId/pickup
  pickupOrder(orderId) {
    return request({
      url: `/rider/orders/${orderId}/pickup`,
      method: 'post'
    })
  },

  // GET /rider/orders/pickup
  getPickupOrders() {
    return request({
      url: '/rider/orders/pickup',
      method: 'get'
    })
  },

  // GET /rider/orders/delivering
  getDeliveringOrders() {
    return request({
      url: '/rider/orders/delivering',
      method: 'get'
    })
  },

  // POST /rider/orders/:orderId/accept (新增)
  acceptOrder(orderId) {
    return request({
      url: `/rider/orders/${orderId}/accept`,
      method: 'post'
    })
  },

  // POST /rider/orders/:orderId/accept_safe (新增)
  acceptOrderSafe(orderId) {
    return request({
      url: `/rider/orders/${orderId}/accept_safe`,
      method: 'post'
    })
  },

  // POST /rider/orders/:orderId/complete
  completeOrder(orderId) {
    return request({
      url: `/rider/orders/${orderId}/complete`,
      method: 'post'
    })
  },

  // GET /rider/orders/:orderId
  getOrderDetail(orderId) {
    return request({
      url: `/rider/orders/${orderId}`,
      method: 'get'
    })
  },

  // GET /rider/orders/history
  getOrderHistory(params) {
    return request({
      url: '/rider/orders/history',
      method: 'get',
      params
    })
  },

  // 增强版本的历史订单API（带演示数据回退）
  getOrderHistoryEnhanced(params) {
    return request({
      url: '/rider/orders/history',
      method: 'get',
      params
    }, {
      context: '获取历史订单',
      useCache: true,
      fallbackData: this.getOrderHistoryWithDemo().data
    });
  },

  // 配送状态扩展
  // PUT /rider/orders/:orderId/start
  startDelivery(orderId) {
    return request({
      url: `/rider/orders/${orderId}/start`,
      method: 'put'
    })
  },

  // PUT /rider/orders/:orderId/arrive-pickup
  arrivePickup(orderId, data) {
    return request({
      url: `/rider/orders/${orderId}/arrive-pickup`,
      method: 'put',
      data
    })
  },

  // PUT /rider/orders/:orderId/status
  updateDeliveryStatus(orderId, data) {
    return request({
      url: `/rider/orders/${orderId}/status`,
      method: 'put',
      data
    })
  },

  // POST /rider/orders/:orderId/issue
  reportIssue(orderId, data) {
    return request({
      url: `/rider/orders/${orderId}/issue`,
      method: 'post',
      data
    })
  },

  // 工作台数据
  // GET /rider/dashboard
  getDashboard() {
    return request({
      url: '/rider/dashboard',
      method: 'get'
    })
  },

  // 增强版本的Dashboard API
  getDashboardEnhanced() {
    return enhancedAPI.wrapApiMethod(() => request({
      url: '/rider/dashboard',
      method: 'get'
    }), {
      context: '获取工作台数据',
      useCache: true
    })();
  },

  // 收入相关
  // GET /rider/income/today
  getTodayIncome() {
    return request({
      url: '/rider/income/today',
      method: 'get'
    })
  },

  // GET /rider/income/summary
  getIncomeSummary() {
    return request({
      url: '/rider/income/summary',
      method: 'get'
    })
  },

  // GET /rider/income/month
  getMonthIncome() {
    return request({
      url: '/rider/income/month',
      method: 'get'
    })
  },

  // GET /rider/income/stats
  getIncomeStats() {
    return request({
      url: '/rider/income/stats',
      method: 'get'
    })
  },

  // 增强版本的收入统计API
  getIncomeStatsEnhanced() {
    return enhancedAPI.wrapApiMethod(() => request({
      url: '/rider/income/stats',
      method: 'get'
    }), {
      context: '获取收入统计',
      fallbackData: this.getIncomeStatsWithDemo().data
    })();
  },

  // GET /rider/income/details
  getIncomeDetails(params) {
    return request({
      url: '/rider/income/details',
      method: 'get',
      params
    })
  },

  // GET /rider/income/history
  getIncomeHistory(params) {
    return request({
      url: '/rider/income/history',
      method: 'get',
      params
    })
  },

  // 钱包相关
  // GET /rider/wallet
  getWalletInfo() {
    return request({
      url: '/rider/wallet',
      method: 'get'
    })
  },

  // POST /rider/wallet/withdraw
  withdraw(data) {
    return request({
      url: '/rider/wallet/withdraw',
      method: 'post',
      data
    })
  },

  // GET /rider/wallet/withdraw/history
  getWithdrawHistory() {
    return request({
      url: '/rider/wallet/withdraw/history',
      method: 'get'
    })
  },

  // GET /rider/delivery/records
  getDeliveryRecords(params) {
    return request({
      url: '/rider/delivery/records',
      method: 'get',
      params
    })
  },

  // GET /rider/delivery/route/:orderId
  getDeliveryRoute(orderId) {
    return request({
      url: `/rider/delivery/route/${orderId}`,
      method: 'get'
    })
  },

  // 统计相关
  // GET /rider/stats/work
  getWorkStats(params) {
    return request({
      url: '/rider/stats/work',
      method: 'get',
      params
    })
  },

  // 增强版本的工作统计API
  getWorkStatsEnhanced(params) {
    return enhancedAPI.wrapApiMethod(() => request({
      url: '/rider/stats/work',
      method: 'get',
      params
    }), {
      context: '获取工作统计',
      useCache: true,
      fallbackData: {
        code: 1,
        data: {
          completedOrders: 125,
          cancelledOrders: 8,
          totalIncome: 3456.78,
          efficiency: 95.5,
          avgDeliveryTime: 18.5,
          totalDistance: 1250.5,
          workDays: 15,
          dailyIncome: 230.45
        }
      }
    })();
  },

  // GET /rider/stats/monthly
  getMonthlyStats() {
    return request({
      url: '/rider/stats/monthly',
      method: 'get'
    })
  },

  // GET /rider/stats/weekly
  getWeeklyStats() {
    return request({
      url: '/rider/stats/weekly',
      method: 'get'
    })
  },

  // GET /rider/ranking/:type
  getRanking(type) {
    return request({
      url: `/rider/ranking/${type}`,
      method: 'get'
    })
  },

  // 评价相关
  // GET /rider/reviews
  getReviews(params) {
    return request({
      url: '/rider/reviews',
      method: 'get',
      params
    })
  },

  // 通知相关
  // GET /rider/notifications
  getNotifications(params) {
    return request({
      url: '/rider/notifications',
      method: 'get',
      params
    })
  },

  // PUT /rider/notifications/:id/read
  markNotificationRead(id) {
    return request({
      url: `/rider/notifications/${id}/read`,
      method: 'put'
    })
  },

  // GET /rider/messages/system
  getSystemMessages(params) {
    return request({
      url: '/rider/messages/system',
      method: 'get',
      params
    })
  },

  // GET /rider/heatmap
  getHeatmapData(params) {
    return request({
      url: '/rider/heatmap',
      method: 'get',
      params
    })
  },

  // 设置相关
  // GET /rider/settings/work
  getWorkSettings() {
    return request({
      url: '/rider/settings/work',
      method: 'get'
    })
  },

  // PUT /rider/settings/work
  updateWorkSettings(data) {
    return request({
      url: '/rider/settings/work',
      method: 'put',
      data
    })
  },

  // GET /rider/settings/account
  getAccountSettings() {
    return request({
      url: '/rider/settings/account',
      method: 'get'
    })
  },

  // PUT /rider/settings/account
  updateAccountSettings(data) {
    return request({
      url: '/rider/settings/account',
      method: 'put',
      data
    })
  },

  // GET /rider/settings/notification
  getNotificationSettings() {
    return request({
      url: '/rider/settings/notification',
      method: 'get'
    })
  },

  // PUT /rider/settings/notification
  updateNotificationSettings(data) {
    return request({
      url: '/rider/settings/notification',
      method: 'put',
      data
    })
  },

  // 认证相关
  // GET /rider/verification
  getVerification() {
    return request({
      url: '/rider/verification',
      method: 'get'
    })
  },

  // POST /rider/verification
  submitVerification(data) {
    return request({
      url: '/rider/verification',
      method: 'post',
      data
    })
  },

  // Demo数据接口（用于测试）
  getRiderInfoWithDemo() {
    return Promise.resolve({
      code: 1,
      data: {
        id: 1,
        name: '李骑手',
        avatar: 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png',
        phone: '13888888888',
        rating: 4.8,
        completedOrders: 1250,
        isOnline: true
      }
    })
  },

  getIncomeStatsWithDemo() {
    return Promise.resolve({
      code: 1,
      data: {
        dailyIncome: 185.5,
        weeklyIncome: 1280.0,
        monthlyIncome: 5200.0,
        completedOrders: 1250,
        estimatedIncome: 45.0
      }
    })
  },

  getNewOrdersWithDemo() {
    return Promise.resolve({
      code: 1,
      data: [
        {
          id: 1,
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
          id: 2,
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
    })
  },

  getOrderHistoryWithDemo() {
    return Promise.resolve({
      code: 1,
      data: {
        items: [
          {
            id: 1,
            orderId: 1,
            amount: 12.50,
            type: 'delivery',
            time: '2024-11-28T10:30:00Z',
            remark: '麦当劳订单配送完成',
            restaurant: '麦当劳',
            customer: '张同学',
            status: 4
          },
          {
            id: 2,
            orderId: 2,
            amount: 8.00,
            type: 'delivery',
            time: '2024-11-28T14:15:00Z',
            remark: '肯德基订单配送完成',
            restaurant: '肯德基',
            customer: '李老师',
            status: 4
          },
          {
            id: 3,
            orderId: 3,
            amount: 15.00,
            type: 'delivery',
            time: '2024-11-28T18:45:00Z',
            remark: '星巴克订单配送完成',
            restaurant: '星巴克',
            customer: '王同学',
            status: 4
          },
          {
            id: 4,
            orderId: 4,
            amount: 10.00,
            type: 'delivery',
            time: '2024-11-28T11:20:00Z',
            remark: '必胜客订单配送完成',
            restaurant: '必胜客',
            customer: '赵同学',
            status: 4
          },
          {
            id: 5,
            orderId: 5,
            amount: 7.50,
            type: 'delivery',
            time: '2024-11-28T12:10:00Z',
            remark: '汉堡王订单配送完成',
            restaurant: '汉堡王',
            customer: '钱同学',
            status: 4
          }
        ],
        total: 5,
        currentPage: 1,
        pageSize: 20
      }
    })
  }
}

export default riderApi