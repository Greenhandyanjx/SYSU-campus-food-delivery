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
    modelValue?: string;   // ← 新增
    type?: string;
    size?: number;
    propImageUrl?: string; // ← 老的不删，但不推荐继续用
  }>(),
  {
    type: '.jpg,.jpeg,.png',
    size: 2,
    propImageUrl: '',
    modelValue: ''   // 默认值
  }
)


const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
  (e: 'imageChange', value: string): void;  // 保留你的旧事件（可选）
}>()


const uploadfiles = ref<any>(null)
const imageUrl = ref<string>(props.modelValue || props.propImageUrl || '')

watch(() => props.modelValue, val => {
  imageUrl.value = val || ''
})

const headers = computed(() => {
  const token = getToken()
  if (token) {
    return {
      Authorization: `Bearer ${token}`
    }
  }
  return {}   // 没登录 → 不带 token（注册页）
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
  emit('update:modelValue', url)  
  emit('imageChange', url)
}

function oploadImgDel() {
  imageUrl.value = ''
  emit('update:modelValue', '')
  emit('imageChange', '')
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
      headers: {
        ...headers.value,  // ⭐ 动态是否附带 token
        'Content-Type': 'multipart/form-data'
      },
      onUploadProgress: (e: any) => {
        const loaded = e.loaded || (e.detail && e.detail.loaded) || 0
        const total = e.total || (e.detail && e.detail.total) || 0
        if (total && typeof onProgress === 'function') {
          onProgress({ percent: Math.round((loaded / total) * 100) })
        }
      }
    })

    handleAvatarSuccess(resp)

  } catch (err: any) {
    if (typeof onError === 'function') onError(err)
    ElMessage.error('图片上传失败：' + (err.message || ''))
  }
}


</script>
<style scoped lang="scss">
.borderNone {
  .el-upload {
    border: 1px solid #d9d9d9 !important;
  }
}
</style>
<style scoped lang="scss">
.avatar-uploader .el-icon-plus {
  position: relative;
}

.avatar-uploader .el-icon-plus:after {
  position: absolute;
  content: '';
  left: calc(100% - 20px);
  top: calc(100% - 20px);
  width: 40px;
  height: 40px;
  background: url('./../../assets/icons/icon_upload@2x.png') center center no-repeat;
  background-size: 20px;
  transform: translate(-50%, -50%); /* ✅ 永远居中，无论容器大小 */
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
  position: relative; /* 限定绝对定位子元素在此范围内，避免覆盖父卡片 */
}
.upload-tips {
  font-size: 12px;
  color: #6c6b6b;
  display: inline-block;
  line-height: 17px;
  margin-left: 36px;
}
.el-upload-list__item-actions {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;

  /* 更浅、更柔和的透明黑背景 */
  background-color: rgba(0, 0, 0, 0.25);

  /* 字体居中 */
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;

  color: #fff;
  font-size: 13px;
  padding: 5px 8px;

  /* 嵌入图片底部的圆角 */
  border-radius: 0 0 6px 6px;

  /* 不再 hover 才出现：始终显示 */
  opacity: 1;
  transition: opacity 0.15s;

  /* 保证在图片之上但不挡太多 */
  z-index: 5;
}

.avatar-uploader .el-upload {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  display: inline-block; /* 避免占满父容器宽度 */
}
.avatar-uploader {
  display: inline-block;
}

.avatar-uploader .el-upload:hover {
  border-color: #ffc200;
}
.el-upload-span {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 13px;
  color: #ffffff;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.16);
}

.el-upload-span:first-child {
  margin-right: 6px;
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 160px;
  height: 120px;
  line-height: 120px;
  text-align: center;
}

.avatar {
  max-width: 160px;
  max-height: 120px;
  width: 100%;
  height: auto;
  display: block;
  object-fit: cover;
}
</style>
