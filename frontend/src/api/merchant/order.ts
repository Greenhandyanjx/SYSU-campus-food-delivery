// src/api/merchant/order.ts
import request from './request'

// 查询不同状态订单（分页）
export function getOrderListBy(params: any) {
  return request({
    url: '/merchant/orders/status',
    method: 'get',
    params,
  })
}

// 获取订单分页详情（组件中作为 getOrderDetailPage 使用）
export function getOrderDetailPage(params: any) {
  return request({
    url: '/merchant/orders/page',
    method: 'get',
    params,
  })
}

// 根据订单 id 查询订单详情
export function queryOrderDetailById(params: { orderId: string | number }) {
  return request({
    url: '/merchant/order/detail',
    method: 'get',
    params,
  })
}

// 商家接单
export function orderAccept(data: any) {
  return request({
    url: '/merchant/order/accept',
    method: 'post',
    data,
  })
}

// 商家拒单
export function orderReject(data: any) {
  return request({
    url: '/merchant/order/reject',
    method: 'post',
    data,
  })
}

// 商家取消订单（后台操作）
export function orderCancel(data: any) {
  return request({
    url: '/merchant/order/cancel',
    method: 'post',
    data,
  })
}

// 派送（商家标记已派送）
export function deliveryOrder(data: any) {
  return request({
    url: '/merchant/order/delivery',
    method: 'post',
    data,
  })
}

// 完成订单
export function completeOrder(data: any) {
  return request({
    url: '/merchant/order/complete',
    method: 'post',
    data,
  })
}

