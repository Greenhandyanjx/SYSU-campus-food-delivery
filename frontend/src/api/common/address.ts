import request from '@/api/merchant/request'

export async function listAddresses() {
  try {
    const res = await request.get('/user/addresses')
    return res.data
  } catch (e: any) {
    console.error('[AddressAPI] listAddresses error:', e?.response?.status, e?.response?.data || e?.message || e)
    // 如果是 401（未登录/令牌过期），返回明确信息
    if (e?.response?.status === 401) {
      return { code: 401, msg: '登录已过期，请重新登录', data: null }
    }
    // 网络错误或超时
    if (e?.code === 'ECONNABORTED') {
      return { code: 0, msg: '请求超时，请检查网络', data: null }
    }
    return { code: 0, msg: e?.message || '加载地址失败', data: null }
  }
}

export async function addAddress(payload: any) {
  try {
    const res = await request.post('/user/address', payload)
    return res.data
  } catch (e) {
    return { code: 0, msg: 'failed' }
  }
}

export async function editAddress(id: number | string, payload: any) {
  try {
    const res = await request.put(`/user/address/${id}`, payload)
    return res.data
  } catch (e) {
    return { code: 0, msg: 'failed' }
  }
}

export async function setDefaultAddress(id: number | string) {
  try {
    const res = await request.post(`/user/address/${id}/default`)
    return res.data
  } catch (e) {
    return { code: 0, msg: 'failed' }
  }
}

export async function deleteAddress(id: number | string) {
  try {
    const res = await request.delete(`/user/address/${id}`)
    return res.data
  } catch (e) {
    return { code: 0, msg: 'failed' }
  }
}

export default {
  listAddresses,
  addAddress,
  editAddress,
  setDefaultAddress,
  deleteAddress,
}
