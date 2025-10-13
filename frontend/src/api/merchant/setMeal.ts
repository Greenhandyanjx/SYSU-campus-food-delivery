import request from './request'

export function getSetmealPage(params: any) {
  return request({
    url: '/merchant/meal/page',
    method: 'get',
    params,
  })
}

export function enableOrDisableSetmeal(data: any) {
  return request({
    url: '/merchant/meal/status',
    method: 'post',
    data,
  })
}

export function deleteSetmeal(idOrList: any) {
  return request({
    url: '/merchant/meal/delete',
    method: 'post',
    data: { id: idOrList },
  })
}

export function querySetmealById(id: any) {
  return request({
    url: '/merchant/meal/query',
    method: 'get',
    params: { id },
  })
}

export function addSetmeal(data: any) {
  return request({
    url: '/merchant/meal/add',
    method: 'post',
    data,
  })
}

export function editSetmeal(data: any) {
  return request({
    url: '/merchant/meal/edit',
    method: 'post',
    data,
  })
}

