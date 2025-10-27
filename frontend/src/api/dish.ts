import request from '@/api/merchant/request'

/**
 * 菜品 API（前端使用）说明
 * 下面每个函数前均以中文说明功能、请求路径、参数示例及返回示例，
 * 方便后端开发按注释实现接口并与前端对接。
 */

/**
 * getDishPage(params)
 * 功能：分页查询菜品（后台管理页）
 * 请求：GET /merchant/dishes/page
 * 参数示例：{ page:1, size:20, name?:string, categoryId?:string }
 * 返回示例：{ code:1, data: { items:[{ id,name,price,status,categoryId }], total:100 } }
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
 * 功能：编辑菜品信息（含上下架、价格、描述、图片、分类）
 * 请求：POST /merchant/dish/edit
 * 请求体示例：{ id:'d1', name:'宫保鸡丁', price:28, categoryId:'c1', images:[], status:'on' }
 * 返回示例：{ code:1, data:{ success:true, dishId:'d1' } }
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
 * 返回示例：{ code:1, data:{ success:true, removed:['d1'] } }
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
 * 功能：变更菜品状态（上架/下架/推荐等）
 * 请求：POST /merchant/dish/status
 * 请求体示例：{ id:'d1', status:'off' }
 * 返回示例：{ code:1, data:{ success:true } }
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
 * 功能：获取菜品分类（用于编辑/筛选）
 * 请求：GET /merchant/dish/categories
 * 返回示例：{ code:1, data: [ { id:'c1', name:'主食' } ] }
 */
export function getCategoryList(params: any) {
  return request({
    url: '/merchant/dish/categories',
    method: 'get',
    params,
  })
}

/**
 * queryDishById(id)
 * 功能：查询单个菜品详情，用于编辑页展示
 * 请求：GET /merchant/dish/query?id=xxx
 * 返回示例：{ code:1, data: { id,name,price,images,description,categoryId,stock } }
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
 * 请求体示例：{ name, price, categoryId, images:[], description, stock }
 * 返回示例：{ code:1, data:{ success:true, id:'newId' } }
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
 * 功能：通用文件下载接口（导出菜品/报表等）
 * 请求：GET /merchant/common/download
 * 参数示例：{ type:'dishes', format:'csv' }
 * 返回示例：文件流（responseType: blob）
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
