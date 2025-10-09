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
