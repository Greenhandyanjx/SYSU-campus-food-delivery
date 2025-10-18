<template>
  <div class="login">
    <div class="login-box">
      <div class="login-form">
        <el-form ref="registerRef" :model="form" :rules="rules" label-width="0">
          <div class="login-form-title">
            <span class="title-label">注册新账号</span>
          </div>

          <el-form-item prop="username">
            <el-input v-model="form.username" placeholder="用户名" />
          </el-form-item>

          <el-form-item  prop="role">
            <el-radio-group v-model="form.role">
              <el-radio-button label="user">用户</el-radio-button>
              <el-radio-button label="rider">骑手</el-radio-button>
              <el-radio-button label="merchant">商家</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item prop="password">
            <el-input v-model="form.password" type="password" placeholder="密码 (至少6位)" />
          </el-form-item>

          <el-form-item prop="confirm">
            <el-input v-model="form.confirm" type="password" placeholder="确认密码" @keyup.enter="handleRegister" />
          </el-form-item>

          <!-- code is optional and defaulted to '0' so no input shown -->

          <el-form-item>
            <el-button type="primary" style="width:100%" @click="handleRegister">注册</el-button>
          </el-form-item>

          <el-form-item>
            <el-button type="text" style="width:100%" @click="() => router.push('/login')">返回登录</el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { registerApi } from '@/api/auth'

const router = useRouter()
const registerRef = ref()
const form = ref({ username: '', password: '', confirm: '', role: 'user', code: '0' })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirm: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: (rule: any, value: string, callback: Function) => {
      if (value !== form.value.password) {
        callback(new Error('两次密码输入不一致'))
      } else {
        callback()
      }
    }, trigger: 'blur' }
  ]
}

async function handleRegister() {
  if (!registerRef.value) return
  ;(registerRef.value as any).validate(async (valid: boolean) => {
    if (!valid) return
    try {
  const res = await registerApi({ username: form.value.username, password: form.value.password, role: form.value.role, code: form.value.code })
      if (res && res.data && String(res.data.code) === '1') {
        ElMessage.success('注册成功，正在返回登录页')
        router.push('/login')
      } else {
        ElMessage.error(res.data.msg || '注册失败')
      }
    } catch (err: any) {
      ElMessage.error(err?.message || '注册请求失败')
    }
  })
}
</script>

<style scoped lang="scss">
.login {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  background-color: #333;
}
.login-box {
  width: 100%;
  // max-width: 520px;
  margin: 0 auto;
  border-radius: 12px;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 16px;
}
.login-form {
  background: #ffffff;
  width: 30%;
  padding: 20px;
  border-radius: 8px;
}
.login-form .el-form-item {
  margin-bottom: 12px;
}
.login-form-title {
  height: 36px;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 20px;
  .title-label {
    font-weight: 500;
    font-size: 20px;
    color: #333333;
  }
}
</style>
