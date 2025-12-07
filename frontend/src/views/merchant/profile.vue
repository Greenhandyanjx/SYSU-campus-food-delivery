<template>
  <div class="merchant-profile-page">
    <div class="card">
        <h2>商家信息</h2>

        <div class="profile-grid">
          <div class="field-row">
            <div class="field-label">店铺 Logo</div>
            <div class="field-value">
              <img :src="form.logo || defaultLogo" class="logo-preview" @error="onLogoError" />
            </div>
            <div class="field-action">
              <el-button size="small" @click="editingLogo = !editingLogo">{{ editingLogo ? '取消' : '修改' }}</el-button>
            </div>
          </div>
          <div v-if="editingLogo" class="uploader-row">
            <ImgUpLoad v-model="form.logo" :size="4" />
            <el-button type="primary" @click="saveField('logo')">保存</el-button>
          </div>

          <div class="field-row">
            <div class="field-label">店铺名称</div>
            <div class="field-value">
              <template v-if="!editing.shop_name">{{ form.shop_name || '未设置' }}</template>
              <template v-else>
                <el-input v-model="editValues.shop_name" style="width:320px" />
              </template>
            </div>
            <div class="field-action">
              <template v-if="!editing.shop_name">
                <el-button type="text" @click="startEdit('shop_name')">修改</el-button>
              </template>
              <template v-else>
                <el-button type="text" @click="cancelEdit('shop_name')">取消</el-button>
                <el-button type="primary" @click="saveField('shop_name')">保存</el-button>
              </template>
            </div>
          </div>

          <div class="field-row">
            <div class="field-label">联系人</div>
            <div class="field-value">
              <template v-if="!editing.owner">{{ form.owner || '未设置' }}</template>
              <template v-else>
                <el-input v-model="editValues.owner" style="width:320px" />
              </template>
            </div>
            <div class="field-action">
              <template v-if="!editing.owner">
                <el-button type="text" @click="startEdit('owner')">修改</el-button>
              </template>
              <template v-else>
                <el-button type="text" @click="cancelEdit('owner')">取消</el-button>
                <el-button type="primary" @click="saveField('owner')">保存</el-button>
              </template>
            </div>
          </div>

          <div class="field-row">
            <div class="field-label">联系电话</div>
            <div class="field-value">
              <template v-if="!editing.phone">{{ form.phone || '未设置' }}</template>
              <template v-else>
                <el-input v-model="editValues.phone" style="width:220px" />
              </template>
            </div>
            <div class="field-action">
              <template v-if="!editing.phone">
                <el-button type="text" @click="startEdit('phone')">修改</el-button>
              </template>
              <template v-else>
                <el-button type="text" @click="cancelEdit('phone')">取消</el-button>
                <el-button type="primary" @click="saveField('phone')">保存</el-button>
              </template>
            </div>
          </div>

          <div class="field-row">
            <div class="field-label">店铺地址</div>
            <div class="field-value">
              <template v-if="!editing.shop_location">{{ form.shop_location || '未设置' }}</template>
              <template v-else>
                <el-input v-model="editValues.shop_location" style="width:320px" />
              </template>
            </div>
            <div class="field-action">
              <template v-if="!editing.shop_location">
                <el-button type="text" @click="startEdit('shop_location')">修改</el-button>
              </template>
              <template v-else>
                <el-button type="text" @click="cancelEdit('shop_location')">取消</el-button>
                <el-button type="primary" @click="saveField('shop_location')">保存</el-button>
              </template>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import ImgUpLoad from '@/components/ImgUpLoad/index.vue'
import { getMerchantProfile, updateMerchantProfile } from '@/api/merchant/profile'

const defaultLogo = '/src/assets/merchant.svg'

const form = reactive({ shop_name: '', phone: '', logo: '', shop_location: '', owner: '' })
const editing = reactive({ shop_name: false, owner: false, phone: false, shop_location: false })
const editingLogo = ref(false)
const editValues = reactive({ shop_name: '', owner: '', phone: '', shop_location: '' })

function onLogoError(e: any) {
  e.target.src = defaultLogo
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
  } catch (e) {
    // ignore
  }
})

function startEdit(field: string) {
  editing[field] = true
  editValues[field] = (form as any)[field] || ''
}
function cancelEdit(field: string) {
  editing[field] = false
  editValues[field] = ''
}

async function saveField(field: string) {
  const payload: any = {}
  if (field === 'logo') {
    payload.logo = form.logo || ''
  } else {
    payload[field] = editValues[field]
  }
  try {
    await updateMerchantProfile(payload)
    // update local view
    if (field === 'logo') {
      editingLogo.value = false
    } else {
      (form as any)[field] = editValues[field]
      editing[field] = false
    }
    ElMessage.success('保存成功')
  } catch (e) {
    ElMessage.error('保存失败')
  }
}
</script>

<style scoped>
.merchant-profile-page { padding: 20px }
.card { background: #fff; padding: 20px; border-radius: 8px }
</style>
