import request from '@/api/merchant/request'

export const getChatHistory = (merchantId: number | string, userBaseId: number | string) => {
  return request({
    url: '/chat/history',
    method: 'GET',
    params: { merchantId, userBaseId }
  })
}

// 返回 WS 地址（在前端可用来 new WebSocket(url)）
export const getWsUrl = () => {
  // 前端直接构造 ws 地址，后端需要支持 token 验证
  const host = window.location.hostname || 'localhost'
  const port = 3000
  return `ws://${host}:${port}/api/chat/ws`
}
