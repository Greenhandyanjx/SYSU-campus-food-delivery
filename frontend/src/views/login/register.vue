<template>
  <div class="register-page">
    <div class="register-container">
      <div class="register-form" style="opacity: 0.9;">
  <el-form ref="registerRef" :model="form" :rules="rules" label-width="0" @keydown.enter.prevent="handleRegister">

          <!-- 标题 -->
          <div class="form-title">
            <span>注册新账号</span>
          </div>

          <!-- 基础信息 -->
            <el-form-item prop="username">
              <div class="form-item" :class="{ 'has-value': form.username }">
                <el-input v-model="form.username" prefix-icon="User" />
                <label>用户名</label>
              </div>
            </el-form-item>

          <el-form-item prop="role">
            <el-radio-group v-model="form.role" size="large">
              <el-radio-button label="user">用户</el-radio-button>
              <el-radio-button label="rider">骑手</el-radio-button>
              <el-radio-button label="merchant">商家</el-radio-button>
            </el-radio-group>
          </el-form-item>

          <el-form-item prop="password">
            <div class="form-item" :class="{ 'has-value': form.password }">
              <el-input v-model="form.password" type="password" prefix-icon="Lock" />
              <label>密码（至少6位）</label>
            </div>
          </el-form-item>

          <el-form-item prop="confirm" style="margin-bottom: 10px;">
            <div class="form-item" :class="{ 'has-value': form.confirm }">
              <el-input v-model="form.confirm" type="password" prefix-icon="Check" />
              <label>确认密码</label>
            </div>
          </el-form-item>

          <!-- ===== 动态表单：用户端 ===== -->
          <template v-if="form.role === 'user'">
            <el-divider>用户信息</el-divider>

            <el-form-item prop="nickname">
              <div class="form-item" :class="{ 'has-value': form.nickname }">
                <el-input v-model="form.nickname" prefix-icon="User" />
                <label>昵称</label>
              </div>
            </el-form-item>

            <el-form-item prop="phone">
              <div class="form-item" :class="{ 'has-value': form.phone }">
                <el-input v-model="form.phone" prefix-icon="Phone" />
                <label>手机号</label>
              </div>
            </el-form-item>
            <el-form-item prop="code">
              <div class="form-item" :class="{ 'has-value': form.code }">
                <el-input v-model="form.code" prefix-icon="Message" />
                <label>短信验证码</label>
                <el-button class="send-code" type="text" @click.prevent="sendCode" :disabled="sending || !phoneValid">{{ sending ? (countdown>0 ? countdown + 's' : '发送中') : '发送验证码' }}</el-button>
              </div>
            </el-form-item>

            <el-form-item prop="address">
              <div class="form-item" :class="{ 'has-value': form.address }">
                <el-input v-model="form.address" prefix-icon="Location" />
                <label>收货地址（宿舍楼/门牌号）</label>
              </div>
            </el-form-item>
          </template>

          <!-- ===== 动态表单：骑手端 ===== -->
          <template v-if="form.role === 'rider'">
            <el-divider>骑手认证</el-divider>

            <el-form-item prop="realname">
              <div class="form-item" :class="{ 'has-value': form.realname }">
                <el-input v-model="form.realname" prefix-icon="User" />
                <label>真实姓名</label>
              </div>
            </el-form-item>

            <el-form-item prop="idNumber">
              <div class="form-item" :class="{ 'has-value': form.idNumber }">
                <el-input v-model="form.idNumber" prefix-icon="Document" />
                <label>身份证号</label>
              </div>
            </el-form-item>

            <el-form-item prop="phone">
              <div class="form-item" :class="{ 'has-value': form.phone }">
                <el-input v-model="form.phone" prefix-icon="Phone" />
                <label>联系电话</label>
              </div>
            </el-form-item>

            <el-form-item prop="code">
              <div class="form-item" :class="{ 'has-value': form.code }">
                <el-input v-model="form.code" prefix-icon="Message" />
                <label>短信验证码</label>
                <el-button class="send-code" type="text" @click.prevent="sendCode" :disabled="sending || !phoneValid">{{ sending ? (countdown>0 ? countdown + 's' : '发送中') : '发送验证码' }}</el-button>
              </div>
            </el-form-item>

            <!-- <el-form-item prop="vehicle">
              <el-select v-model="form.vehicle" placeholder="交通工具">
                <el-option label="步行" value="walk" />
                <el-option label="电动车" value="ebike" />
                <el-option label="自行车" value="bike" />
              </el-select>
            </el-form-item> -->
<!-- 
            <label>上传身份证照片</label>
            <el-form-item prop="idPhoto">
              <ImageUpload v-model="form.idPhoto" />
            </el-form-item> -->
          </template>

          <!-- ===== 动态表单：商家端 ===== -->
          <template v-if="form.role === 'merchant'">
            <el-divider>商家信息</el-divider>

            <el-form-item prop="shopName">
              <div class="form-item" :class="{ 'has-value': form.shopName }">
                <el-input v-model="form.shopName" prefix-icon="Shop" />
                <label>店铺名称</label>
              </div>
            </el-form-item>

            <el-form-item prop="shopLocation">
              <div class="form-item" :class="{ 'has-value': form.shopLocation }">
                <el-input v-model="form.shopLocation" prefix-icon="Location" />
                <label>店铺地址/位置</label>
              </div>
            </el-form-item>

            <!-- <el-form-item prop="owner">
              <el-input v-model="form.owner" placeholder="负责人姓名" prefix-icon="User" />
              <label>店铺地址/位置</label>
            </el-form-item> -->
            <el-form-item prop="owner">
              <div class="form-item" :class="{ 'has-value': form.owner }">
                <el-input v-model="form.owner" prefix-icon="User" />
                <label>负责人姓名</label>
              </div>
            </el-form-item>

            <!-- <el-form-item prop="phone">
              <el-input v-model="form.phone" placeholder="联系电话" prefix-icon="Phone" />
            </el-form-item> -->
            <el-form-item prop="phone">
              <div class="form-item" :class="{ 'has-value': form.phone }">
                <el-input v-model="form.phone" prefix-icon="Phone" />
                <label>联系电话</label>
              </div>
            </el-form-item>   

            <el-form-item prop="code" style="margin-bottom: 5px;">
              <div class="form-item" :class="{ 'has-value': form.code }">
                <el-input v-model="form.code" prefix-icon="Message" />
                <label>短信验证码</label>
                <el-button class="send-code" type="text" @click.prevent="sendCode" :disabled="sending || !phoneValid">{{ sending ? (countdown>0 ? countdown + 's' : '发送中') : '发送验证码' }}</el-button>
              </div>
            </el-form-item>

            <div style="display: flex; gap: 20px;">
              <el-form-item prop="license" >
                <label>营业执照照片</label>
                <ImageUpload v-model="form.license" />
              </el-form-item>
              
              <el-form-item prop="logo">
                <label>店铺 Logo</label>
                <ImageUpload v-model="form.logo" />
              </el-form-item>
            </div>

          </template>

          <!-- ===== 注册按钮 ===== -->
          <el-form-item style="margin-bottom: 5px;">
            <el-button type="primary" class="btn-register" @click="handleRegister">注册</el-button>
          </el-form-item>

          <el-form-item style="margin: 0;">
            <el-button type="text" style="width: 100%" @click="() => router.push('/login')">
              返回登录
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { User, Lock, Check, Phone, Location, Document, Shop } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { registerApi, registerUser, registerRider, registerMerchant } from '@/api/auth'
import ImageUpload from '@/components/ImgUpLoad/index.vue'

const router = useRouter()
const registerRef = ref()

// 表单数据
const form = ref({
  username: '',
  password: '',
  confirm: '',
  role: 'user',
  code: '0',
  nickname: '',
  phone: '',
  address: '',
  realname: '',
  idNumber: '',
  vehicle: '',
  idPhoto: '',
  shopName: '',
  owner: '',
  license: '',
  logo: ''
  , shopLocation: ''
})

// SMS / phone helpers
const sending = ref(false)
const countdown = ref(0)
let _codeTimer: any = null
const sentCode = ref('')
const sentPhone = ref('')

const phoneValid = computed(() => {
  const p = (form.value.phone || '').toString().trim()
  // 简单中国手机号校验：1[3-9] + 9位
  return /^1[3-9]\d{9}$/.test(p)
})

function sendCode() {
  if (!phoneValid.value) { ElMessage.warning('请输入有效的手机号'); return }
  if (sending.value) return
  sending.value = true
  // 生成 6 位随机验证码（演示用）
  const code = Math.floor(100000 + Math.random() * 900000).toString()
  sentCode.value = code
  sentPhone.value = form.value.phone
  countdown.value = 60
  ElMessage.success('验证码已发送（演示）: ' + code)
  _codeTimer = setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      clearInterval(_codeTimer)
      sending.value = false
      countdown.value = 0
    }
  }, 1000)
}

onBeforeUnmount(() => {
  if (_codeTimer) clearInterval(_codeTimer)
})

// 基础表单校验规则（会在提交时做额外 role-specific 校验）
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ],
  confirm: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string, callback: Function) => {
        if (value !== form.value.password) {
          callback(new Error('两次密码输入不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 注册逻辑
async function handleRegister() {
  if (!registerRef.value) return
  ;(registerRef.value as any).validate(async (valid: boolean) => {
    if (!valid) return

    // 额外 role-specific 必填检查
    if (form.value.role === 'user') {
      if (!form.value.nickname) { ElMessage.warning('请填写昵称'); return }
      if (!form.value.phone) { ElMessage.warning('请填写手机号'); return }
      if (!form.value.address) { ElMessage.warning('请填写收货地址'); return }
    }
    if (form.value.role === 'rider') {
      if (!form.value.realname) { ElMessage.warning('请填写真实姓名'); return }
      if (!form.value.idNumber) { ElMessage.warning('请填写身份证号'); return }
      if (!form.value.phone) { ElMessage.warning('请填写联系电话'); return }
      // if (!form.value.vehicle) { ElMessage.warning('请选择交通工具'); return }
      // if (!form.value.idPhoto) { ElMessage.warning('请上传身份证照片'); return }
    }
    if (form.value.role === 'merchant') {
      if (!form.value.shopName) { ElMessage.warning('请填写店铺名称'); return }
      if (!form.value.owner) { ElMessage.warning('请填写负责人姓名'); return }
      if (!form.value.phone) { ElMessage.warning('请填写联系电话'); return }
      if (!form.value.license) { ElMessage.warning('请上传营业执照照片'); return }
      if (!form.value.logo) { ElMessage.warning('请上传店铺 Logo'); return }
    }

    // 验证手机号格式与短信验证码（如果手机号存在）
    if (form.value.phone) {
      if (!phoneValid.value) { ElMessage.warning('请输入有效的手机号'); return }
      if (!sentCode.value || sentPhone.value !== form.value.phone || String(form.value.code || '') !== String(sentCode.value)) {
        ElMessage.warning('验证码不正确或未发送，请检查手机号并重新获取')
        return
      }
    }

    try {
      let res: any = null
      // 将图片字段名与后端接口约定好：ImageUpload v-model 已经存储图片 URL（例如 form.idPhoto, form.license, form.logo）
      if (form.value.role === 'user') {
        res = await registerUser({
          username: form.value.username,
          password: form.value.password,
          nickname: form.value.nickname,
          phone: form.value.phone,
          address: form.value.address,
          code: form.value.code,
          role: "user",
        })
      } else if (form.value.role === 'rider') {
        res = await registerRider({
          username: form.value.username,
          password: form.value.password,
          realname: form.value.realname,
          idNumber: form.value.idNumber,
          phone: form.value.phone,
          vehicle: form.value.vehicle,
          idPhotoUrl: form.value.idPhoto,
          code: form.value.code,
          role: "rider",
        })
      } else if (form.value.role === 'merchant') {
        res = await registerMerchant({
          username: form.value.username,
          password: form.value.password,
          shopName: form.value.shopName,
          shopLocation: form.value.shopLocation,
          owner: form.value.owner,
          phone: form.value.phone,
          licenseUrl: form.value.license,
          logoUrl: form.value.logo,
          code: form.value.code,
          role: "merchant",
        })
      } else {
        // fallback to generic
        res = await registerApi(form.value)
      }

      if (res?.data?.code === 1 || String(res?.data?.code) === '1') {
        ElMessage.success('注册成功，正在跳转登录页')
        router.push('/login')
      } else {
        ElMessage.error(res?.data?.msg || '注册失败')
      }
    } catch (err: any) {
      ElMessage.error(err?.message || '注册请求失败')
    }
  })
}
</script>

<style scoped lang="scss">
.register-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  /* 使用与登录页相同的背景图片（覆盖整个页面），并保留淡色叠层以保证表单可读性 */
  background-image: url("@/assets/login/img_denglu_bj.jpg");
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}
.register-container {
  width: 60%;
  max-width: 500px;
  background: #ffffff;
  opacity: 0.9;
  border-radius: 12px;
  box-shadow: 0 8px 30px rgba(64,158,255,0.08);
  padding: 28px 32px;
}
.register-form {
  opacity: 0.9;
  .form-title {
    text-align: center;
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 10px;
    color: #333;
  }
  .icon {
    width: 18px;
    height: 18px;
    margin-right: 4px;
  }
  .el-form-item {
    margin-bottom: 18px;
  }
  /* 浮动标签输入样式，参考 Address 页面 */
  .form-item {
    position: relative;
    width: 100%;
  }
  .form-item :deep(.el-input__inner) {
    width: 100%;
    padding: 15px 14px 12px 20px; /* 留出前缀图标空间 */
    margin-bottom: 1px;
    // border: 1px solid #dcdfe6; 
    border-radius: 8px;
    background: #fff;
    transition: all 0.2s;
    height: 35px;
  }
  .form-item label {
    position: absolute;
    left: 20px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 14px;
    color: #909399;
    pointer-events: none;
    background: transparent;
    transition: 0.18s ease all;
    padding: 0 4px;
  }
  .form-item:focus-within label,
  .form-item.has-value label {
    top: -10px;
    font-size: 10px;
    color: #409eff;
    transform: translateY(0);
    // background: #fff;
    padding: 0 6px;
    z-index: 2;
  }
  /* 保证发送验证码按钮在右侧不会与输入文字冲突 */
  .send-code {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 13px;
    color: #409eff;
    z-index: 3;
  }
  .send-code {
    position: absolute;
    right: 8px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 13px;
    color: #409eff;
  }
  .btn-register {
    width: 100%;
    background-color: #409eff; /* Element blue */
    color: #fff;
    font-weight: 600;
    border: none;
  }
  .btn-register:hover {
    background-color: #3a8ee6;
  }
}
</style>
