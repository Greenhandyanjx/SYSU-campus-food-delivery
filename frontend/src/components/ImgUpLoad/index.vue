<!--  -->
<template>
  <div class="upload-item">
  <!-- <el-upload ref="uploadfiles"
         :accept="type"
         :class="{ borderNone: imageUrl }"
         class="avatar-uploader"
         :show-file-list="false"
         :http-request="httpRequest"
         :on-success="handleAvatarSuccess"
         :on-remove="handleRemove"
         :on-error="handleError"
         :before-upload="beforeAvatarUpload"
         :headers="headers"> -->
      <el-upload
      ref="uploadfiles"
      :accept="type"
      class="avatar-uploader"
      :show-file-list="false"
      :http-request="httpRequest"
      :before-upload="beforeAvatarUpload"
      :headers="headers"
    >
      <img v-if="imageUrl"
           :src="imageUrl"
           class="avatar">

      <i v-else
         class="el-icon-plus avatar-uploader-icon" />
      <span v-if="imageUrl"
            class="el-upload-list__item-actions">
        <span class="el-upload-span"
              @click.stop="oploadImgDel">
          删除图片
        </span>
        <span class="el-upload-span"> 重新上传 </span>
      </span>
    </el-upload>
    <p class="upload-tips">
      <slot />
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { getToken } from '@/utils/cookies'
import request from '@/api/merchant/request'

const props = withDefaults(
  defineProps<{
    type?: string
    size?: number
    propImageUrl?: string
  }>(),
  {
    type: '.jpg,.jpeg,.png',
    size: 2,
    propImageUrl: ''
  }
)

const emit = defineEmits<{
  (e: 'imageChange', value: string): void
}>()

const uploadfiles = ref<any>(null)
const imageUrl = ref<string>(props.propImageUrl || '')

const headers = computed(() => {
  console.log('上传token：', getToken())  // ✅ 在 return 之前打印
  return {
    Authorization: `Bearer ${getToken()}`
  }
})



// watch(
//   () => props.propImageUrl,
//   (val) => {
//     imageUrl.value = val || ''
//   }
// )

function handleRemove() {
  // placeholder for remove hook if needed
}

function handleError(err: any, file: any, fileList: any) {
  console.error('handleError', err, file, fileList)
  ElMessage.error('图片上传失败')
}

// function handleAvatarSuccess(response: any, file: any, fileList: any) {
//   console.log('✅ 上传返回：', response)
//   let url = ''

//   if (typeof response === 'string') {
//     url = response
//   } else if (response?.data && typeof response.data === 'string') {
//     url = response.data
//   } else if (response?.data?.url) {
//     url = response.data.url
//   } else if (response?.url) {
//     url = response.url
//   } else if (response?.path) {
//     url = response.path
//   }

//   // ✅ 自动补全 baseURL（替换为你的后端地址）
//   if (url && !/^https?:\/\//.test(url)) {
//     url = import.meta.env.VITE_API_BASE_URL + url
//   }

//   imageUrl.value = url
//   emit('imageChange', imageUrl.value)
//   console.log('✅ 最终图片地址：', imageUrl.value)
// }
function handleAvatarSuccess(response: any) {
  console.log('✅ 上传返回：', response)
  let url = ''
  const body = response?.data ?? response

  if (typeof body?.url === 'string') url = body.url
  else if (typeof body?.data === 'string') url = body.data
  else if (typeof body?.data?.url === 'string') url = body.data.url

  console.log('✅ 最终图片地址：', url)
  imageUrl.value = url
  emit('imageChange', url)
}

function oploadImgDel() {
  imageUrl.value = ''
  emit('imageChange', imageUrl.value)
}

function beforeAvatarUpload(file: File) {
  const maxM = props.size || 2
  const isLt = file.size / 1024 / 1024 < maxM
  if (!isLt) {
    ElMessage.error(`上传文件大小不能超过${maxM}M!`)
    return false
  }
  return true
}

// custom http request handler so we use the project's axios instance (request)
async function httpRequest({ file, onProgress, onError }: any) {
  try {
    const form = new FormData()
    form.append('file', file)
    const resp = await request.post('/common/upload', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e: any) => {
        const loaded = e.loaded || (e.detail && e.detail.loaded) || 0
        const total = e.total || (e.detail && e.detail.total) || 0
        if (total && typeof onProgress === 'function') {
          onProgress({ percent: Math.round((loaded / total) * 100) })
        }
      }
    })

    // ✅ 不再调用 onSuccess，而是自己手动调用 handleAvatarSuccess 一次
    handleAvatarSuccess(resp)

  } catch (err: any) {
    if (typeof onError === 'function') onError(err)
    ElMessage.error('图片上传失败：' + (err.message || ''))
  }
}

</script>
<style lang='scss'>
.borderNone {
  .el-upload {
    border: 1px solid #d9d9d9 !important;
  }
}
</style>
<style scoped lang="scss">
.avatar-uploader .el-icon-plus:after {
  position: absolute;
  display: inline-block;
  content: ' ' !important;
  left: calc(50% - 20px);
  top: calc(50% - 40px);
  width: 40px;
  height: 40px;
  background: url('./../../assets/icons/icon_upload@2x.png') center center
    no-repeat;
  background-size: 20px;
}

.el-upload-list__item-actions:hover .upload-icon {
  display: inline-block;
}
.el-icon-zoom-in:before {
  content: '\E626';
}
.el-icon-delete:before {
  content: '\E612';
}
.el-upload-list__item-actions:hover {
  opacity: 1;
}
.upload-item {
  display: flex;
  align-items: center;
  .el-form-item__content {
    width: 500px !important;
  }
}
.upload-tips {
  font-size: 12px;
  color: #666666;
  display: inline-block;
  line-height: 17px;
  margin-left: 36px;
}
.el-upload-list__item-actions {
  position: absolute;
  width: 100%;
  height: 100%;
  left: 0;
  top: 0;
  cursor: default;
  text-align: center;
  color: #fff;
  opacity: 0;
  font-size: 20px;
  background-color: rgba(0, 0, 0, 0.5);
  transition: opacity 0.3s;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-direction: column;
}
.avatar-uploader .el-upload {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}
.avatar-uploader {
  display: inline-block;
}

.avatar-uploader .el-upload:hover {
  border-color: #ffc200;
}
.el-upload-span {
  width: 100px;
  height: 30px;
  border: 1px solid #ffffff;
  border-radius: 4px;
  font-size: 14px;
  text-align: center;
  line-height: 30px;
}

.el-upload-span:first-child {
  margin-bottom: 20px;
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 200px;
  height: 160px;
  line-height: 160px;
  text-align: center;
}

.avatar {
  width: 200px;
  height: 160px;
  display: block;
}
</style>
