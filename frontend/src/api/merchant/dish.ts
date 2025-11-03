import request from '@/api/merchant/request'

/**
 * 菜品相关接口说明（用于后端对照实现）
 * 说明：以下每个函数在注释中给出功能、请求方法与路径、参数示例、返回示例以及注意事项。
 * 联动提示：菜品分类接口（getCategoryList / dishCategoryList）与 getDishPage/queryDishList 在前端会联合使用，
 * 当新增/编辑菜品后请后端在成功后刷新缓存/返回最新分类列表以保证前端展示一致。
 */

/**
 * getDishPage(params)
 * 功能：分页获取菜品列表（后台管理分页）
 * 请求：GET /merchant/dishes/page
 * 参数示例：{ page: 1, size: 20, name?: string, categoryId?: string }
 * 返回示例：{ code:1, data: { items: [ { id, name, price, status, categoryId, stock } ], total: 123 } }
 * 备注：分页字段名与格式请与前端约定一致（items+total）。
 */
export function getDishPage(params: any) {
  return request({
    url: '/merchant/dishes/page',
    method: 'get',
    params,
  })
}

/**
 * editDish(data)
 * 功能：编辑菜品信息（包含下架/上架、价格、描述、图片、分类等）
 * 请求：POST /merchant/dish/edit
 * 请求体示例：{ id: 'd1', name: '宫保鸡丁', price: 28.0, categoryId: 'c1', images: ['url1'], status: 'on' }
 * 返回示例：{ code:1, data: { success:true, dishId: 'd1' } }
 * 备注：前端在成功后会刷新菜品列表与可能的分类缓存。
 */
export function editDish(data: any) {
  return request({
    url: '/merchant/dish/edit',
    method: 'post',
    data,
  })
}

/**
 * deleteDish(idOrList)
 * 功能：删除单个或批量删除菜品
 * 请求：POST /merchant/dish/delete
 * 请求体示例：{ id: 'd1' } 或 { id: ['d1','d2'] }
 * 返回示例：{ code:1, data: { success:true, removed: ['d1'] } }
 * 备注：删除应处理与订单/库存/统计的级联影响，或做软删除。
 */
export function deleteDish(idOrList: any) {
  return request({
    url: '/merchant/dish/delete',
    method: 'post',
    data: { id: idOrList },
  })
}

/**
 * dishStatusByStatus(data)
 * 功能：修改菜品状态（如上架、下架、推荐等）
 * 请求：POST /merchant/dish/status
 * 请求体示例：{ id: 'd1', status: 'off' }
 * 返回示例：{ code:1, data: { success:true } }
 */
export function dishStatusByStatus(data: any) {
  return request({
    url: '/merchant/dish/status',
    method: 'post',
    data,
  })
}

/**
 * getCategoryList(params)
 * 功能：获取菜品所属分类列表（用于菜品编辑/筛选）
 * 请求：GET /merchant/dish/categories
 * 参数示例：{ merchantId?: string }
 * 返回示例：{ code:1, data: [ { id:'c1', name:'主食'} ] }
 * 菜品分类包括：后面需要根据这些分类进行菜品的筛选和展示
 *
  { label: '招牌套餐'
  { label: '现煮粉面'
  { label: '汉堡炸鸡'
  { label: '奶茶咖啡'
  { label: '日式便当'
  { label: '烧烤烤肉'
  { label: '水果拼盘'
  { label: '精致甜品'
  { label: '家常快炒'
  { label: '粥粉面饭'
 */
export function getCategoryList(params: any) {
  return request({
    url: '/merchant/dish/categories',
    method: 'get',
    params,
  })
}

/**
 * queryDishList(params)
 * 功能：基于条件查询菜品（不一定分页，常用于选择/下拉列表）
 * 请求：GET /merchant/dish/list
 * 参数示例：{ categoryId?: string, name?: string }
 * 返回示例：{ code:1, data: [ { id, name, price } ] }
 */
export function queryDishList(params: any) {
  // 支持按 categoryId 或 name 搜索
  return request({
    url: '/merchant/dish/list',
    method: 'get',
    params,
  })
}

/**
 * queryDishById(id)
 * 功能：查询单个菜品详情（详细信息用于编辑页）
 * 请求：GET /merchant/dish/query?id=xxx
 * 返回示例：{ code:1, data: { id, name, price, images:[], description, categoryId, stock } }
 */
export function queryDishById(id: any) {
  return request({
    url: '/merchant/dish/query',
    method: 'get',
    params: { id },
  })
}

/**
 * addDish(data)
 * 功能：新增菜品
 * 请求：POST /merchant/dish/add
 * 请求体示例：{ name, price, categoryId, images:[...], description, stock }
 * 返回示例：{ code:1, data: { success:true, id: 'newDishId' } }
 */
export function addDish(data: any) {
  return request({
    url: '/merchant/dish/add',
    method: 'post',
    data,
  })
}

/**
 * commonDownload(params)
 * 功能：通用下载接口（导出/下载文件流）
 * 请求：GET /merchant/common/download
 * 参数示例：{ type: 'dishes', format: 'csv' }
 * 返回示例：文件流（responseType: 'blob'）
 */
export function commonDownload(params: any) {
  return request({
    url: '/merchant/common/download',
    method: 'get',
    params,
  })
}

export function dishCategoryList(params: any) {
  // alias for getCategoryList
  return getCategoryList(params)
}

