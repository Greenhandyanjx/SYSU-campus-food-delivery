import request from '@/api/merchant/request'

export const getChatHistory = (merchantId: number | string, userBaseId?: number | string) => {
  const params: any = { merchantId }
  if (userBaseId) params.userBaseId = userBaseId
  return request({
    url: '/chat/history',
    method: 'GET',
    params
  })
}

// 返回 WS 地址（在前端可用来 new WebSocket(url)）
export const getWsUrl = () => {
  // 前端直接构造 ws 地址，后端需要支持 token 验证
  const host = window.location.hostname || 'localhost'
  const port = 3000
  return `ws://${host}:${port}/api/chat/ws`
}

// 获取商家详情（用于显示名称/头像、确认 merchantId）
export const getMerchantDetail = (merchantId: number | string) => {
  return request({ url: '/merchant/detail', method: 'get', params: { id: merchantId } }).catch(err => {
    // if not found or other error, return a normalized empty response to avoid unhandled rejections
    return { data: { code: 0, data: null, msg: err?.response?.data || err?.message } }
  })
}

// 获取 base_user 详情；若不传 id，后端使用 Authorization token 推断当前用户
export const getBaseUserDetail = (id?: number | string) => {
  const params: any = {}
  if (id) params.id = id
  return request({ url: '/baseuser/detail', method: 'get', params }).catch(err => {
    return { data: { code: 0, data: null, msg: err?.response?.data || err?.message } }
  })
}
