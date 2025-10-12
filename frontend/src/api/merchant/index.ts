// src/api/index.ts
import request from './request'

// 营业数据
export function getBusinessData() {
  return request({
    url: '/merchant/businessData',
    method: 'get',
  })
}

// 今日订单
export function getOrderData() {
  return request({
    url: '/merchant/orderData',
    method: 'get',
  })
}

// 菜品总览
export function getOverviewDishes() {
  return request({
    url: '/merchant/overviewDishes',
    method: 'get',
  })
}

// 套餐总览
export function getSetMealStatistics() {
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
