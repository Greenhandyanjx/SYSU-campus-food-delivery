<template>
  <div class="review-page">
    <div class="card">
      <h3>订单评价 #{{ id }}</h3>
      <div class="row">
        <label>商家评分</label>
        <div class="stars">
          <span v-for="n in 5" :key="n" class="star" :class="{ active: n <= merchantScore }" @click="merchantScore = n">★</span>
        </div>
      </div>
      <div class="row">
        <label>骑手评分</label>
        <div class="stars">
          <span v-for="n in 5" :key="n" class="star" :class="{ active: n <= riderScore }" @click="riderScore = n">★</span>
        </div>
      </div>
      <div class="row">
        <label>备注（可选）</label>
        <el-input type="textarea" v-model="comment" placeholder="写下你的评价（非必填）" rows="3"></el-input>
      </div>
      <div class="actions">
        <el-button @click="onCancel">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">提交</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import orderApi from '@/api/user/order'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const id = route.params.id
const merchantScore = ref(5)
const riderScore = ref(5)
const comment = ref('')
const submitting = ref(false)

function onCancel() { router.back() }

async function onSubmit() {
  if (!id) return ElMessage.error('order id missing')
  // prepare payload
  const payload = {
    merchant_score: merchantScore.value,
    rider_score: riderScore.value,
    comment: comment.value
  }
  try {
    submitting.value = true
    const res = await orderApi.reviewOrder(String(id), payload)
    if (res && res.data && Number(res.data.code) === 1) {
      ElMessage.success('评价提交成功')
      try { window.dispatchEvent(new CustomEvent('order:changed', { detail: { orderId: id } })) } catch (e) {}
      router.push('/user/orderlist')
    } else {
      ElMessage.error(res?.data?.msg || '提交失败')
    }
  } catch (e) {
    ElMessage.error('提交失败')
  } finally {
    submitting.value = false
  }
}

onMounted(()=>{})
</script>

<style scoped>
.review-page { display:flex; justify-content:center; padding:40px 0 }
.card { width:720px; background:#fff; padding:20px; border-radius:8px; box-shadow:0 8px 24px rgba(0,0,0,0.08) }
.row { display:flex; align-items:center; gap:16px; margin:12px 0 }
.row label { width:100px; color:#666 }
.stars { font-size:28px; color:#ddd; cursor:pointer }
.star.active { color:#f5b301 }
.actions { display:flex; justify-content:flex-end; gap:12px; margin-top:18px }
</style>
