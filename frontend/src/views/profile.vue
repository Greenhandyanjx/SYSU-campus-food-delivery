<template>
  <div class="profile-page">
    <div class="card">
      <h2>个人信息</h2>
      <p><strong>用户名：</strong> {{ username }}</p>

      <el-divider />

      <h3>修改密码</h3>
      <el-form :model="form" ref="formRef" label-width="100px">
        <el-form-item label="原密码" prop="oldPassword">
          <el-input v-model="form.oldPassword" type="password" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="form.newPassword" type="password" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="changePassword">保存</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref(localStorage.getItem('username') || '')

const formRef = ref()
const form = ref({ oldPassword: '', newPassword: '' })

const changePassword = async () => {
  // 简单前端校验，后端接口需实现
  if (!form.value.oldPassword || !form.value.newPassword) {
    ElMessage.error('请填写原密码和新密码')
    return
  }
  // TODO: 调用修改密码 API
  ElMessage.success('密码修改请求已发送（演示）')
  // 清空表单
  form.value.oldPassword = ''
  form.value.newPassword = ''
}
</script>

<style scoped>
.profile-page {
  padding: 20px;
}
.card {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
}
</style>
