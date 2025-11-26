import request from '@/api/merchant/request'

export async function listAddresses() {
  try {
    const res = await request.get('/user/addresses')
    return res.data
  } catch (e) {
    return { code: 0, msg: 'failed', data: null }
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
