// src/api/merchant/order.ts
import request from './request'

/**
 * 订单相关接口说明（商家端）
 * 联动说明：订单状态变更会影响统计（merchant/index.ts）、库存和用户端的订单列表，
 * 后端在变更订单状态时应同步更新这些模块或提供事件/消息通知。
 */

// 查询不同状态订单（分页）
export function getOrderListBy(params: any) {
  /**
   * GET /merchant/orders/status
   * 功能：按订单状态查询（分页）
   * 参数示例：{ status: 'pending|delivering|completed', page:1, size:20 }
   * 返回示例：{ code:1, data: { items:[{ orderId, status, total, createdAt }], total } }
   */
  return request({
    url: '/merchant/orders/status',
    method: 'get',
    params,
  })
}

// 获取订单分页详情（组件中作为 getOrderDetailPage 使用）
export function getOrderDetailPage(params: any) {
  /**
   * GET /merchant/orders/page
   * 功能：分页获取订单详情（可用于后台订单管理视图）
   * 参数示例：{ page:1, size:20, begin?, end? }
   */
  return request({
    url: '/merchant/orders/page',
    method: 'get',
    params,
  })
}

// 合并相同分页请求（仅合并并发中的请求，不做响应缓存，保证状态变更后刷新到最新数据）
const _inFlightOrderPage: Record<string, Promise<any>> = {}
export function getOrderDetailPageCoalesced(params: any) {
  const key = JSON.stringify({ url: '/merchant/orders/page', params: params || {} })
  if (_inFlightOrderPage[key]) {
    return _inFlightOrderPage[key]
  }
  const p = getOrderDetailPage(params)
  _inFlightOrderPage[key] = p
  p.then((r: any) => {
    delete _inFlightOrderPage[key]
    return r
  }).catch((err: any) => {
    delete _inFlightOrderPage[key]
    throw err
  })
  return p
}

// 根据订单 id 查询订单详情
export function queryOrderDetailById(params: { orderId: string | number }) {
  /**
   * GET /merchant/order/detail?orderId=xxx
   * 功能：查询单个订单的完整详情
   * 返回示例：{ code:1, data: { orderId, items:[{name,qty,price}], status, address, delivery } }
   */
  return request({
    url: '/merchant/order/detail',
    method: 'get',
    params,
  })
}

// 合并相同的订单详情请求（仅合并并发中的请求，不缓存响应）
const _inFlightOrderDetails: Record<string, Promise<any>> = {}
export function queryOrderDetailByIdCoalesced(params: { orderId: string | number }) {
  const id = String(params.orderId || '')
  const key = JSON.stringify({ url: '/merchant/order/detail', id })
  if (_inFlightOrderDetails[key]) {
    return _inFlightOrderDetails[key]
  }
  const p = queryOrderDetailById(params)
  _inFlightOrderDetails[key] = p
  p.then((r: any) => {
    delete _inFlightOrderDetails[key]
    return r
  }).catch((err: any) => {
    delete _inFlightOrderDetails[key]
    throw err
  })
  return p
}

// 商家接单
export function orderAccept(data: any) {
  /**
   * POST /merchant/order/accept
   * 请求体示例：{ orderId: 'o1', acceptBy: 'merchantId' }
   * 返回示例：{ code:1, data: { success:true } }
   * 说明：接单后应改变订单状态为 'accepted' 并触发配送流程。
   */
  return request({
    url: '/merchant/order/accept',
    method: 'post',
    data,
  })
}

// 商家拒单
export function orderReject(data: any) {
  /**
   * POST /merchant/order/reject
   * 请求体示例：{ orderId: 'o1', reason: '缺货' }
   * 返回示例：{ code:1, data: { success:true } }
   * 说明：拒单后应回滚库存并通知用户。
   */
  return request({
    url: '/merchant/order/reject',
    method: 'post',
    data,
  })
}

// 商家取消订单（后台操作）
export function orderCancel(data: any) {
  /**
   * POST /merchant/order/cancel
   * 请求体示例：{ orderId: 'o1', reason: '商家取消' }
   * 返回示例：{ code:1, data: { success:true } }
   */
  return request({
    url: '/merchant/order/cancel',
    method: 'post',
    data,
  })
}

// 派送（商家标记已派送）
export function deliveryOrder(data: any) {
  /**
   * POST /merchant/order/delivery
   * 请求体示例：{ orderId: 'o1', courierId: 'r1' }
   * 返回示例：{ code:1, data: { success:true } }
   * 说明：标记为已派送后配送状态更新为 'delivering' 或 'shipped'。
   */
  return request({
    url: '/merchant/order/delivery',
    method: 'post',
    data,
  })
}

// 完成订单
export function completeOrder(data: any) {
  /**
   * POST /merchant/order/complete
   * 请求体示例：{ orderId: 'o1' }
   * 返回示例：{ code:1, data: { success:true } }
   * 说明：标记订单完成并触发结算/评价流程。
   */
  return request({
    url: '/merchant/order/complete',
    method: 'post',
    data,
  })
}

