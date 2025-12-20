//后端字段映射建议：
// 对于 getStoreByName 返回：建议后端返回 
// { code: 1, data: { id, name, desc, ... } } 
// 或直接 { id,name,... }。
// 前端目前会尝试从 res.data.data、res.data 或直接 res 取值。
// 对于 getDishesByStore 返回：
// 建议返回数组形式的菜品对象，包含 id、name、price、image。前端把这些直接渲染，count 字段本地维护（或从 getCart 合并）。
import request from '@/api/merchant/request'

// 用户侧门店与购物车相关接口
export function getStoreByName(name: string) {
  return request({
    url: '/store/query',
    method: 'get',
    params: { name },
  })
}

export function getStoreById(id: string | number) {
  return request({
    url: '/store/query',
    method: 'get',
    // 兼容后端：同时传递 id 与 base_id（有些后端以 base_id 查找商家）
    params: { id, base_id: id },
  })
}

export function getDishesByStore(storeId: string | number) {
  return request({
    url: '/store/dishes',
    method: 'get',
    params: { storeId },
  })
}

// 将菜品加入用户购物车（后端应接收 storeId 或 storeName + dishId/name + qty）
export function addToCart(data: any) {
  return request({
    url: '/user/cart/add',
    method: 'post',
    data,
  })
}

export function removeFromCart(data: any) {
  return request({
    url: '/user/cart/remove',
    method: 'post',
    data,
  })
}

export function getCart(params: any) {
  return request({
    url: '/user/cart',
    method: 'get',
    params,
  })
}

export function getDeliveryConfig(base_id: string | number) {
  return request({
    url: '/merchant/delivery_config',
    method: 'get',
    params: { base_id },
  })
}

export function getMealPublicById(id: string | number) {
  return request({
    url: '/store/meal/query',
    method: 'get',
    params: { id },
  })
}

export default {
  getStoreByName,
  getStoreById,
  getDishesByStore,
  addToCart,
  removeFromCart,
  getCart,
}
