import request from '@/api/merchant/request'

// Demo fallback data
const DEMO_PROFILE = {
  username: '游客',
  vipLevel: '普通用户',
  points: 120,
  orderCount: 3,
  couponCount: 2
}

export async function getProfile() {
  try {
    const res = await request.get('/user/profile')
    // Normalize: return inner data object if backend uses utils.Success wrapper
    return (res.data && res.data.data) ? res.data.data : (res.data || DEMO_PROFILE)
  } catch (e) {
    return DEMO_PROFILE
  }
}

export async function updateProfile(payload: { nickname?: string; phone?: string; avatar_url?: string }) {
  try {
    const res = await request.post('/user/profile/update', payload)
    return res.data
  } catch (e) {
    throw e
  }
}

export async function changePassword(payload: { username?: string; oldpassword: string; newpassword: string }) {
  try {
    const res = await request.post('/change_password', payload)
    return res.data
  } catch (e) {
    throw e
  }
}

export async function getWallet() {
  try {
    const res = await request.get('/user/wallet')
    return res.data || { balance: 0 }
  } catch (e) {
    return { balance: 0 }
  }
}

export async function getCoupons() {
  try {
    const res = await request.get('/user/coupons')
    return res.data || []
  } catch (e) {
    return []
  }
}

export async function logout() {
  try {
    // Try server logout if available
    await request.post('/user/logout')
    return { ok: true }
  } catch (e) {
    // Backend may not exist yet — still consider logout successful on frontend
    return { ok: true }
  }
}

export default { getProfile, getWallet, getCoupons, logout }
