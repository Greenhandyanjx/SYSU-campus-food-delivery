
import request from './request'

/**
 * 获取分类列表（按类型）
 * 功能：查询商家维度的菜品分类，用于商家后台和前端展示分类筛选。
 * 请求：GET /merchant/category/list
 * 请求参数示例：{ type?: string, page?: number, size?: number }
 * 返回示例：{ code: 1, data: [ { id: 'c1', name: '主食', type: 'dish', sort: 1 }, ... ] }
 * 说明：
 *  - 如果后端分页，请返回 data:{ items: [...], total: N } 或与前端约定的分页格式；
 *  - 前端调用此接口以展示分类下拉/侧栏，或在新增/编辑菜品时选择分类。
 */
export function getCategoryByType(params: any) {
  return request({
    url: '/merchant/category/list',
    method: 'get',
    params,
  })
}


