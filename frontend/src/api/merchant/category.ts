
import request from './request'

export function getCategoryByType(params: any) {
  return request({
    url: '/merchant/category/list',
    method: 'get',
    params,
  })
}


