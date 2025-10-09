<template>
  <div class="login">
    <div class="login-box">
      <!-- <img src="@/assets/login/login-l.png" alt="" /> -->
      <div class="login-form">
        <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules">
          <!-- <div class="login-form-title">
            <img
              src="@/assets/login/icon_logo.png"
              style="width: 149px; height: 38px"
              alt=""
            />
          </div> -->
          <div class="login-form-title">
            <span class="title-label">校园外卖管理系统</span>
          </div>
          <el-radio-group v-model="loginForm.role"  class="role-selector">
            <el-radio-button label="user">          
              <el-icon><User /></el-icon>
              <span>我是用户</span>
            </el-radio-button>
            <el-radio-button label="rider">
              <el-icon><Bicycle /></el-icon>
              <span>我是骑手</span>
            </el-radio-button>
            <el-radio-button label="merchant">          
              <el-icon><Shop /></el-icon>
              <span>我是商家</span>
            </el-radio-button>
          </el-radio-group>
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="账号"
              prefix-icon="User"
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="密码"
              prefix-icon="Lock"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              :loading="loading"
              class="login-btn"
              size="large"
              type="primary"
              style="width: 100%"
              @click.prevent="handleLogin"
            >
              <span v-if="!loading">登录</span>
              <span v-else>登录中...</span>
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, FormInstance, FormRules } from 'element-plus'
import { User, Bicycle, Shop } from '@element-plus/icons-vue'
// import { loginApi } from '@/api/user' ////等我们封装loginapi

const router = useRouter()
const loginFormRef = ref<FormInstance>()
const loginForm = ref({
  username: 'aaaaaa',
  password: '121341',
  role:'user' // 默认
})

const loginRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码必须在6位以上', trigger: 'blur' }
  ]
}

const loading = ref(false)

// 登录逻辑
const handleLogin = async () => {
  if (!loginFormRef.value) return

  await loginFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      loading.value = true

      try {
        //等我们封装loginapi
        // const res = await loginApi(loginForm.value)
        const res = { code: '1', msg: 'success' }

        if (String(res.code) === '1') {
          ElMessage.success('登录成功')
          switch (loginForm.value.role) {
            case 'user':
              router.push('/user/home')
              break
            case 'rider':
              router.push('/rider/dashboard')
              break
            case 'merchant':
              router.push('/merchant/dashboard')
              break
          }
        } else {
          ElMessage.error(res.msg || '用户名或密码错误')
        }
      } catch (error) {
        ElMessage.error('登录失败，请检查网络')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>
<style lang="scss">
//合适大小的登录窗口

.login {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  // background: #476dbe;
  background-color: #333;
}
.login-box {
  width: 100%;
  max-width: 1920px;
  margin: 0 auto;
  aspect-ratio: 1920 / 991;
  border-radius: 12px;
  display: flex;
  justify-content: center;
  align-items: center;

  background-image: url('@/assets/login/img_denglu_bj.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
  img{
    width: 60%;
    height: auto;
  }
}

// .login-box {
//   position: relative;
//   width: 100%;
//   aspect-ratio: 1920 / 991; /* 保持16:9比例自动缩放 */
//   border-radius: 8px;
//   display: flex;
//   overflow: hidden;
//   opacity: 0.95;
//   background-image: url('@/assets/login/img_denglu_bj.jpg');
//   background-size: cover;      /* 关键：等比例缩放填充 */
//   background-position: center; /* 居中显示 */
//   background-repeat: no-repeat;

//   img {
//     width: 60%;
//     height: auto;
//   }
// }

.title {
  margin: 0px auto 10px auto;
  text-align: left;
  color: #707070;
}

.login-form {
  background: #ffffff;
  opacity: 0.9;
  width: 40%;
  border-radius: 8px 8px 8px 8px;
  display: flex;
  justify-content: center;
  align-items: center;
  .el-form {
    width: 214px;
    height: 307px;
  }
  .el-form-item {   
    margin-bottom: 15px;
    opacity: 0.9;
  }
  .el-form-item.is-error .el-input__inner {
    border: 0 !important;
    border-bottom: 1px solid #fd7065 !important;
    background: #fff !important;
  }
.role-selector {
  display: flex;
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 0px;
}

.role-selector .el-radio-button {
  flex: 1;
  margin: 0;
}

.role-selector .el-radio-button__inner {
  border: none;
  border-right: 1px solid #dcdfe6;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12px 0;
  font-size: 14px;
  font-weight: 500;
  color: #606266;
  transition: all 0.3s;
  background-color: #f9f9f9;
}

.role-selector .el-radio-button:last-child .el-radio-button__inner {
  border-right: none;
}

.role-selector .el-radio-button__original-radio:checked + .el-radio-button__inner {
  background-color: #409eff;
  color: #fff;
}

.role-selector .el-radio-button__inner:hover {
  background-color: #ecf5ff;
  color: #409eff;
}

.role-selector .el-icon {
  font-size: 20px;
  margin-bottom: 4px;
}
  .input-icon {
    height: 32px;
    width: 18px;
    margin-left: -2px;
  }
  .el-input__inner {
    border: 0;
    border-bottom: 1px solid #e9e9e8;
    border-radius: 0;
    font-size: 12px;
    font-weight: 400;
    color: #333333;
    height: 32px;
    line-height: 32px;
  }
  .el-input__prefix {
    left: 0;
  }
  .el-input--prefix .el-input__inner {
    padding-left: 26px;
  }
  .el-input__inner::placeholder {
    color: #aeb5c4;
  }
  .el-form-item--medium .el-form-item__content {
    line-height: 32px;
  }
  .el-input--medium .el-input__icon {
    line-height: 32px;
  }
  .el-input__wrapper {
  height: 40px;
 }
}

.login-btn {
  border-radius: 17px;
  padding: 11px 20px !important;
  margin-top: 10px;
  font-weight: 500;
  font-size: 12px;
  border: 0;
  font-weight: 500;
  color: #333333;
  // background: #09a57a;
  background-color: #ffc200;
  &:hover,
  &:focus {
    // background: #09a57a;
    background-color: #ffc200;
    color: #ffffff;
  }
}
.login-form-title {
  height: 36px;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 40px;
  .title-label {
    font-weight: 500;
    font-size: 20px;
    color: #333333;
    margin-left: 10px;
  }
}

</style>