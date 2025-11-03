// src/api/index.ts
import request from './request'

/**
 * merchant/index.ts - 统计/概览类接口说明（后端实现参考）
 * 联动说明：这些统计接口通常依赖于订单、菜品、商家等数据表，
 * 后端在更新订单状态或菜品/销量变动时应保证统计计算数据的及时性或提供缓存刷新接口。
 */

// 营业数据
export function getBusinessData() {
  /**
   * GET /merchant/businessData
   * 功能：返回店铺的经营概览（营业额、订单数、客单价等）
   * 返回示例：{ code:1, data: { revenue: 12345.6, orders: 123, avgTicket: 100 } }
   */
  return request({
    url: '/merchant/businessData',
    method: 'get',
  })
}

// 今日订单
export function getOrderData() {
  /**
   * GET /merchant/orderData
   * 功能：当日订单统计（待处理/配送中/已完成等）
   * 返回示例：{ code:1, data: { pending: 3, delivering: 2, completed: 20 } }
   */
  return request({
    url: '/merchant/orderData',
    method: 'get',
  })
}

// 菜品总览
export function getOverviewDishes() {
  /**
   * GET /merchant/overviewDishes
   * 功能：菜品一览（例如售卖量、上架数、停售数）
   * 返回示例：{ code:1, data: { sold: 123, discontinued: 5 } }
   */
  return request({
    url: '/merchant/overviewDishes',
    method: 'get',
  })
}

// 套餐总览
export function getSetMealStatistics() {
  /**
   * GET /merchant/setMealStatistics
   * 功能：套餐销售/库存统计
   */
  return request({
    url: '/merchant/setMealStatistics',
    method: 'get',
  })
}

// -----------------------------
// 统计页面（merchant/statistics）相关接口
// -----------------------------

/**
 * 获取数据概览（营业额、有效订单、用户等概览数据）
 * params 示例: { begin: '2025-10-01', end: '2025-10-31' }
 */
export function getDataOverView(params?: any) {
  return request({
    url: '/merchant/statistics/overview',
    method: 'get',
    params,
  })
}

/** 营业额统计 */
export function getTurnoverStatistics(params?: any) {
  return request({
    url: '/merchant/statistics/turnover',
    method: 'get',
    params,
  })
}

/** 用户统计 */
export function getUserStatistics(params?: any) {
  return request({
    url: '/merchant/statistics/user',
    method: 'get',
    params,
  })
}

/** 订单统计 */
export function getOrderStatistics(params?: any) {
  return request({
    url: '/merchant/statistics/order',
    method: 'get',
    params,
  })
}

/** 销量排名 Top */
export function getTop(params?: any) {
  return request({
    url: '/merchant/statistics/top',
    method: 'get',
    params,
  })
}

/** 导出运营数据（返回文件流） */
export function exportInfor(params?: any) {
  return request({
    url: '/merchant/statistics/export',
    method: 'get',
    params,
    responseType: 'blob',
  })
}
