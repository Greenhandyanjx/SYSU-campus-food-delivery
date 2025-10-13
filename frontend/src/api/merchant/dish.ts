import request from '@/api/merchant/request'

export function getDishPage(params: any) {
  return request({
    url: '/merchant/dishes/page',
    method: 'get',
    params,
  })
}

export function editDish(data: any) {
  return request({
    url: '/merchant/dish/edit',
    method: 'post',
    data,
  })
}

export function deleteDish(idOrList: any) {
  return request({
    url: '/merchant/dish/delete',
    method: 'post',
    data: { id: idOrList },
  })
}

export function dishStatusByStatus(data: any) {
  return request({
    url: '/merchant/dish/status',
    method: 'post',
    data,
  })
}

export function getCategoryList(params: any) {
  return request({
    url: '/merchant/dish/categories',
    method: 'get',
    params,
  })
}

export function queryDishList(params: any) {
  // 支持按 categoryId 或 name 搜索
  return request({
    url: '/merchant/dish/list',
    method: 'get',
    params,
  })
}

export function queryDishById(id: any) {
  return request({
    url: '/merchant/dish/query',
    method: 'get',
    params: { id },
  })
}

export function addDish(data: any) {
  return request({
    url: '/merchant/dish/add',
    method: 'post',
    data,
  })
}

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

