import request from './request'

/**
 * 套餐（SetMeal）相关接口说明
 * 这些接口用于商家管理套餐（分页/查询/新增/编辑/上下架/删除）
 */
export function getSetmealPage(params: any) {
  /**
   * GET /merchant/meal/page
   * 功能：分页获取套餐信息
   * 参数示例：{ page:1, size:20, name?:string }
   * 返回示例：{ code:1, data:{ items:[{ id,name,price,status}], total } }
   */
  return request({
    url: '/merchant/meal/page',
    method: 'get',
    params,
  })
}

export function enableOrDisableSetmeal(data: any) {
  /**
   * POST /merchant/meal/status
   * 功能：启用/禁用套餐
   * 请求体示例：{ id:'m1', status:'on' }
   * 返回示例：{ code:1, data:{ success:true } }
   */
  return request({
    url: '/merchant/meal/status',
    method: 'post',
    data,
  })
}

export function deleteSetmeal(idOrList: any) {
  /**
   * POST /merchant/meal/delete
   * 功能：删除套餐（单个或批量）
   * 请求体示例：{ id:'m1' } 或 { id:['m1','m2'] }
   */
  return request({
    url: '/merchant/meal/delete',
    method: 'post',
    data: { id: idOrList },
  })
}

export function querySetmealById(id: any) {
  /**
   * GET /merchant/meal/query?id=xxx
   * 功能：查询单个套餐详情（用于编辑）
   */
  return request({
    url: '/merchant/meal/query',
    method: 'get',
    params: { id },
  })
}

export function addSetmeal(data: any) {
  /**
   * POST /merchant/meal/add
   * 功能：新增套餐
   * 请求体示例：{ name, price, items:[{dishId, qty}], images:[], description }
   */
  return request({
    url: '/merchant/meal/add',
    method: 'post',
    data,
  })
}

export function editSetmeal(data: any) {
  /**
   * POST /merchant/meal/edit
   * 功能：编辑套餐
   */
  return request({
    url: '/merchant/meal/edit',
    method: 'post',
    data,
  })
}

