import request from '@/api/merchant/request'
import storeApi from './store'

/**
 * 用户端订单接口说明（前端调用）
 * 以下接口用于用户查看订单、取消、支付、再次购买等场景。
 * 后端实现时请保证订单状态与库存、支付模块的联动一致。
 */

/**
 * getOrderList(params)
 * 功能：获取用户的订单列表（可分页/按状态筛选）
 * 请求：GET /user/order/list
 * 参数示例：{ page:1, size:20, status?: 'pending|paid|delivering|completed|refund' }
 * 返回示例：{ code:1, data: { items:[{ id, orderNo, status, total, createdAt }], total } }
 */
export function getOrderList(params?: any) {
  return request({ url: '/user/order/list', method: 'get', params })
}

/**
 * getOrderDetail(id)
 * 功能：获取单个订单详情
 * 请求：GET /user/order/{id}
 * 返回示例：{ code:1, data: { id, orderNo, items:[{name,qty,price}], status, address, payInfo } }
 */
export function getOrderDetail(id: string) {
  return request({ url: `/user/order/${id}`, method: 'get' })
}

/**
 * cancelOrder(id)
 * 功能：用户取消订单（在可取消状态内）
 * 请求：POST /user/order/cancel
 * 请求体示例：{ id: 'o1', reason?: '用户取消' }
 * 返回示例：{ code:1, data:{ success:true } }
 * 说明：后端需检查订单状态并处理退款/回滚库存等。
 */
export function cancelOrder(id: string) {
  return request({ url: `/user/order/cancel`, method: 'post', data: { id } })
}

/**
 * payOrder(id, payload)
 * 功能：发起支付（或记录支付回调）
 * 请求：POST /user/order/pay
 * 请求体示例：{ id:'o1', payChannel:'alipay|wechat', amount:100 }
 * 返回示例：{ code:1, data:{ success:true, paymentId:'p1', redirectUrl?:'' } }
 * 说明：后端需与支付网关联动并返回支付结果/跳转地址。
 */
export function payOrder(id: string, payload?: any) {
  return request({ url: `/user/order/pay`, method: 'post', data: { id, ...payload } })
}

/**
 * reorder(order)
 * 功能：再次购买 - 将历史订单内菜品加入购物车
 * 前端实现：调用 storeApi.addToCart 逐条加入（若后端支持可提供批量接口）
 * 请求（后端可实现批量接口）：POST /user/cart/addBatch { storeId, items:[{dishId, qty}] }
 * 返回示例：{ code:1, data:{ success:true, added: { ... } } }
 */
export async function reorder(order: any) {
  if (!order || !order.items) return Promise.reject(new Error('invalid order'))
  // 根据后端需求，这里逐条加入购物车；建议后端实现批量接口以提升性能
  for (const it of order.items) {
    await storeApi.addToCart({ storeId: order.storeId, dishId: it.id || null, name: it.name, qty: it.count || 1 })
  }
  return Promise.resolve({ success: true })
}

/**
 * confirmOrder(id)
 * 功能：用户确认收货
 * 请求：POST /user/order/{id}/confirm
 * 返回示例：{ code:1, data:{ success:true } }
 */
export function confirmOrder(id: string) {
  return request({ url: `/user/order/${id}/confirm`, method: 'post' })
}

/**
 * contactRider(id)
 * 功能：联系骑手
 * 请求：GET /user/order/{id}/rider
 * 返回示例：{ code:1, data:{ name, phone, avatar } }
 */
export function contactRider(id: string) {
  return request({ url: `/user/order/${id}/rider`, method: 'get' })
}

/**
 * refundDetail(id)
 * 功能：获取退款详情
 * 请求：GET /user/order/{id}/refund
 * 返回示例：{ code:1, data:{ status, reason, amount, createdAt } }
 */
export function refundDetail(id: string) {
  return request({ url: `/user/order/${id}/refund`, method: 'get' })
}

/**
 * reviewOrder(id, data)
 * 功能：评价订单
 * 请求：POST /user/order/{id}/review
 * 请求体：{ rating: number, content: string, images?: string[] }
 * 返回：{ code:1, data:{ success:true } }
 */
export function reviewOrder(id: string, data: any) {
  return request({ url: `/user/order/${id}/review`, method: 'post', data })
}

export default {
  getOrderList,
  getOrderDetail,
  cancelOrder,
  payOrder,
  reorder,
  confirmOrder,
  contactRider,
  refundDetail,
  reviewOrder
}
