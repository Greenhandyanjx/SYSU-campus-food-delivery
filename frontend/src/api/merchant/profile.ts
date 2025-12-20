import request from '@/api/merchant/request'

export const getMerchantProfile = async () => {
  return request.get('/merchant/profile')
}

export const updateMerchantProfile = async (payload: any) => {
  return request.post('/merchant/profile/update', payload)
}

export default { getMerchantProfile, updateMerchantProfile }
