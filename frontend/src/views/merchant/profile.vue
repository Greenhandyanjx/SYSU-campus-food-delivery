<template>
  <div class="merchant-profile-page">
    <div class="profile-wrap">
      <h2 class="page-title">商家信息</h2>

      <div class="cards-grid">
        <!-- Logo 卡片 -->
        <div class="card-block">
          <div class="card-header">店铺 Logo</div>
          <div class="card-body logo-body">
            <img :src="form.logo || defaultLogo" class="logo-preview" @error="onLogoError" />
            <div class="logo-actions">
              <el-button size="small" @click="editingLogo = !editingLogo">{{ editingLogo ? '取消' : '修改' }}</el-button>
              <el-button size="small" v-if="editingLogo" @click="saveField('logo')" type="primary">保存</el-button>
            </div>
          </div>
          <div v-if="editingLogo" class="card-footer">
            <ImgUpLoad v-model="form.logo" :size="4" />
          </div>
        </div>

        <!-- 店铺名称 卡片 -->
        <div class="card-block">
          <div class="card-header">店铺名称</div>
          <div class="card-body">
            <div v-if="!editing.shop_name">{{ form.shop_name || '未设置' }}</div>
            <div v-else>
              <el-input v-model="editValues.shop_name" placeholder="店铺名称" />
            </div>
          </div>
          <div class="card-footer actions">
            <template v-if="!editing.shop_name">
              <el-button type="text" @click="startEdit('shop_name')">修改</el-button>
            </template>
            <template v-else>
              <el-button type="text" @click="cancelEdit('shop_name')">取消</el-button>
              <el-button type="primary" @click="saveField('shop_name')">保存</el-button>
            </template>
          </div>
        </div>

        <!-- 联系人 卡片 -->
        <div class="card-block">
          <div class="card-header">联系人</div>
          <div class="card-body">
            <div v-if="!editing.owner">{{ form.owner || '未设置' }}</div>
            <div v-else>
              <el-input v-model="editValues.owner" placeholder="联系人姓名" />
            </div>
          </div>
          <div class="card-footer actions">
            <template v-if="!editing.owner">
              <el-button type="text" @click="startEdit('owner')">修改</el-button>
            </template>
            <template v-else>
              <el-button type="text" @click="cancelEdit('owner')">取消</el-button>
              <el-button type="primary" @click="saveField('owner')">保存</el-button>
            </template>
          </div>
        </div>

        <!-- 手机号 卡片（带验证码流程） -->
        <div class="card-block">
          <div class="card-header">联系电话</div>
          <div class="card-body">
            <div v-if="!editing.phone">{{ form.phone || '未设置' }}</div>
            <div v-else class="phone-edit-row">
              <el-input v-model="editValues.phone" placeholder="请输入手机号" style="width:220px" />
              <el-button class="send-code" :disabled="sending" @click.prevent="sendCode">{{ sending ? (countdown>0 ? countdown + 's' : '发送中') : '发送验证码' }}</el-button>
              <el-input v-model="codeInput" placeholder="验证码" style="width:140px" />
            </div>
          </div>
          <div class="card-footer actions">
            <template v-if="!editing.phone">
              <el-button type="text" @click="startEdit('phone')">修改</el-button>
            </template>
            <template v-else>
              <el-button type="text" @click="cancelEdit('phone')">取消</el-button>
              <el-button type="primary" @click="saveField('phone')">保存</el-button>
            </template>
          </div>
        </div>

        <!-- 地址 卡片 -->
        <div class="card-block big">
          <div class="card-header">店铺地址</div>
          <div class="card-body">
            <div v-if="!editing.shop_location" class="shop-location-display" :title="form.shop_location">{{ form.shop_location || '未设置' }}</div>
            <div v-else>
              <div style="cursor: pointer;width: auto;display: flex;" @click="openAddressPicker" :title="editValues.shop_location || form.shop_location">
                <el-input class="address-input" v-model="editValues.shop_location" placeholder="店铺地址" readonly />
              </div>
            </div>
          </div>
          <div class="card-footer actions">
            <template v-if="!editing.shop_location">
              <el-button type="text" @click="startEdit('shop_location')">修改</el-button>
            </template>
            <template v-else>
              <el-button type="text" @click="cancelEdit('shop_location')">取消</el-button>
              <el-button type="primary" @click="saveField('shop_location')">保存</el-button>
            </template>
          </div>
        </div>
        <AddressPicker v-model="pickerValue" :visible="pickerVisible" @update:visible="pickerVisible = $event" @update:modelValue="onPickerConfirm" @close="onPickerClose" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import AddressPicker from '@/components/AddressPicker.vue'
import ImgUpLoad from '@/components/ImgUpLoad/index.vue'
import { getMerchantProfile, updateMerchantProfile } from '@/api/merchant/profile'
import merchantSvg from '@/assets/merchant.svg'
import noImg from '@/assets/noImg.png'

const defaultLogo = merchantSvg

const form = reactive({ shop_name: '', phone: '', logo: '', shop_location: '', owner: '' })
const editing = reactive({ shop_name: false, owner: false, phone: false, shop_location: false })
const editingLogo = ref(false)
const editValues = reactive({ shop_name: '', owner: '', phone: '', shop_location: '' })
const pickerVisible = ref(false)
const pickerValue = ref({ formatted: '', detail: '', lng: 0, lat: 0 })

// phone verification state
const codeInput = ref('')
const sending = ref(false)
const countdown = ref(0)
let _codeTimer: any = null
const sentCode = ref('')
const sentPhone = ref('')

function onLogoError(e: any) {
  try {
    // avoid endless loop: only replace if current src is different from fallback
    const tgt = e && e.target
    if (!tgt) return
    if ((tgt.src || '').includes(defaultLogo)) {
      // already attempted fallback, use a safe generic image
      tgt.src = noImg
    } else {
      tgt.src = defaultLogo
    }
  } catch (err) {}
}

onMounted(async () => {
  try {
    const r: any = await getMerchantProfile()
    if (r && r.data && r.data.data) {
      const d = r.data.data
      form.shop_name = d.shop_name || d.shopName || ''
      form.phone = d.phone || ''
      form.logo = d.logo || d.logoUrl || ''
      form.shop_location = d.shop_location || d.shopLocation || ''
      form.owner = d.owner || ''
    }
  } catch (e) {}
})

onBeforeUnmount(() => { if (_codeTimer) clearInterval(_codeTimer) })

function startEdit(field: string) {
  editing[field] = true
  editValues[field] = (form as any)[field] || ''
}
function cancelEdit(field: string) {
  editing[field] = false
  editValues[field] = ''
  if (field === 'phone') { codeInput.value = ''; sentCode.value = ''; sentPhone.value = '' }
}

function openAddressPicker() {
  console.log('openAddressPicker called')
  pickerValue.value = { formatted: editValues.shop_location || form.shop_location || '', detail: '', lng: 0, lat: 0 }
  pickerVisible.value = true
}

function onPickerClose() { pickerVisible.value = false }

function onPickerConfirm(val: any) {
  if (val && val.formatted) editValues.shop_location = val.formatted
  pickerVisible.value = false
}

function phoneValid(p: string) {
  const s = (p || '').toString().trim()
  return /^1[3-9]\d{9}$/.test(s)
}

function sendCode() {
  if (!phoneValid(editValues.phone)) { ElMessage.warning('请输入有效手机号'); return }
  if (sending.value) return
  sending.value = true
  const code = Math.floor(100000 + Math.random() * 900000).toString()
  sentCode.value = code
  sentPhone.value = editValues.phone
  countdown.value = 60
  ElMessage.success('验证码已发送（演示）: ' + code)
  _codeTimer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) { clearInterval(_codeTimer); sending.value = false; countdown.value = 0 }
  }, 1000)
}

async function saveField(field: string) {
  const payload: any = {}
  if (field === 'logo') { payload.logo = form.logo || '' }
  else if (field === 'phone') {
    // verify code
    if (!sentCode.value || sentPhone.value !== editValues.phone || String(codeInput.value || '') !== String(sentCode.value)) {
      ElMessage.error('验证码错误或未发送')
      return
    }
    payload.phone = editValues.phone
  } else { payload[field] = editValues[field] }
  try {
    await updateMerchantProfile(payload)
    // update local
    if (field === 'logo') { editingLogo.value = false }
    else { (form as any)[field] = editValues[field]; editing[field] = false }
    if (field === 'phone') { codeInput.value = ''; sentCode.value = ''; sentPhone.value = '' }
    ElMessage.success('保存成功')
  } catch (e) { ElMessage.error('保存失败') }
}
</script>

<style scoped>
.merchant-profile-page { padding: 24px }
.page-title { font-size:20px; font-weight:700; margin-bottom:16px }
.profile-wrap { max-width:1000px; margin:0 auto }
.cards-grid { display:grid; grid-template-columns: repeat(2,1fr); gap:16px }
.card-block { background:#fff; border-radius:10px; box-shadow:0 6px 18px rgba(0,0,0,0.06); overflow:hidden; display:flex; flex-direction:column }
.card-block.big { grid-column: 1 / -1 }
.card-header { padding:12px 16px; font-weight:600; border-bottom:1px solid #f3f3f3; background:linear-gradient(90deg, rgba(255,249,236,0.6), rgba(255,249,236,0.2)) }
.card-body { padding:16px; display:flex; align-items:center; gap:12px }
.card-footer { padding:12px 16px; border-top:1px solid #f7f7f7; display:flex; justify-content:flex-end }
.actions { align-items:center }
.logo-body { justify-content:space-between }
.logo-preview { width:96px; height:96px; object-fit:cover; border-radius:8px; border:1px solid #eee }
.logo-actions { display:flex; gap:8px }
.phone-edit-row { display:flex; gap:8px; align-items:center }

@media (max-width:900px) {
  .cards-grid { grid-template-columns: 1fr }
}
</style>
<style scoped>
:deep(.address-input) .el-input__inner {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}
.shop-location-display {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
