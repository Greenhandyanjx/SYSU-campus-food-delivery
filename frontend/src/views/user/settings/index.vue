<template>
  <div class="my-bg">
    <div class="my-page">
      <div class="profile-card modern settings-card">
        <div class="profile-cover"></div>
        <div class="profile-content">
          <div class="left">
                    <div class="avatar-box">
                      <el-avatar :size="112" class="avatar-preview">
                        <img
                          :src="avatar || defaultAvatar"
                          alt="avatar"
                          style="width: 100%; height: 100%; object-fit: cover;"
                        />
                      </el-avatar>
                      <div style="margin-top:8px;display:flex;gap:8px;align-items:center;">
                        <el-button size="small" @click="() => (editingAvatar = !editingAvatar)">{{ editingAvatar ? '取消' : '修改头像' }}</el-button>
                        <el-button size="small" @click="openEditAvatar" v-if="!editingAvatar">上传图片</el-button>
                      </div>
                    </div>
                    <div v-if="editingAvatar" class="avatar-uploader-inline">
                      <ImgUpLoad v-model="editAvatar" :type="'.jpg,.jpeg,.png'" :size="4" />
                      <div style="margin-top:8px;display:flex;gap:8px;">
                        <el-button size="small" @click="saveAvatar">保存</el-button>
                        <el-button size="small" @click="() => { editingAvatar = false; editAvatar = avatar }">取消</el-button>
                      </div>
                    </div>
                  </div>
          <div class="info-col">
            <div class="field-row">
              <div class="field-label">昵称</div>
              <div class="field-value">
                <template v-if="!editingNickname">
                  {{ nickname || '未设置' }}
                </template>
                <template v-else>
                  <el-input v-model="editNickname" placeholder="输入新的昵称" style="width:320px" />
                </template>
              </div>
              <div class="field-action">
                <template v-if="!editingNickname">
                  <el-button type="text" @click.prevent="() => (editingNickname = true, editNickname = nickname)">修改</el-button>
                </template>
                <template v-else>
                  <el-button type="text" @click.prevent="() => (editingNickname = false)">取消</el-button>
                  <el-button type="primary" @click.prevent="saveNickname">保存</el-button>
                </template>
              </div>
            </div>

            <div class="field-row">
              <div class="field-label">手机号</div>
              <div class="field-value">
                <template v-if="!editingPhone">
                  {{ maskedPhone }}
                </template>
                <template v-else>
                  <div style="display:flex;gap:8px;align-items:center">
                    <el-input v-model="editPhone" placeholder="请输入手机号" style="width:220px" />
                    <el-button class="send-code" :disabled="sending" @click.prevent="sendCode">{{ sending ? (countdown>0 ? countdown + 's' : '发送中') : '发送验证码' }}</el-button>
                    <el-input v-model="codeInput" placeholder="验证码" style="width:140px" />
                  </div>
                </template>
              </div>
              <div class="field-action">
                <template v-if="!editingPhone">
                  <el-button type="text" @click.prevent="() => { editingPhone = true; editPhone = phone }">修改</el-button>
                </template>
                <template v-else>
                  <el-button type="text" @click.prevent="() => { editingPhone = false; codeInput = ''; sentCode=''; sentPhone=''; }">取消</el-button>
                  <el-button type="primary" @click.prevent="savePhone">确定</el-button>
                </template>
              </div>
            </div>
            <div class="field-row">
              <div class="field-label">密码</div>
              <div class="field-value">
                <template v-if="!editingPassword">******</template>
                <template v-else>
                  <div style="display:flex;flex-direction:column;gap:8px;max-width:420px">
                    <el-input v-model="pwdForm.oldpassword" placeholder="旧密码" type="password" />
                    <el-input v-model="pwdForm.newpassword" placeholder="新密码" type="password" />
                    <el-input v-model="pwdForm.confirm" placeholder="确认新密码" type="password" />
                  </div>
                </template>
              </div>
              <div class="field-action">
                <template v-if="!editingPassword">
                  <el-button type="text" @click.prevent="editingPassword = true">修改密码</el-button>
                </template>
                <template v-else>
                  <el-button type="text" @click.prevent="() => { editingPassword = false; pwdForm = { oldpassword:'', newpassword:'', confirm:'' } }">取消</el-button>
                  <el-button type="primary" @click.prevent="savePassword">保存</el-button>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 编辑昵称/手机号使用内联展开，无独立对话框 -->

    <!-- 编辑头像对话框（输入 URL 作为快速方式） -->
    <el-dialog title="修改头像" :visible.sync="dlgAvatar">
      <div style="display:flex;gap:8px;align-items:center;">
        <el-avatar :size="96" :src="editAvatar || avatar || defaultAvatar" />
        <el-input v-model="editAvatar" placeholder="输入图片 URL 或上传后粘贴" />
      </div>
      <p style="margin-top:8px;color:#888">建议使用已上传到服务器或 CDN 的图片 URL；也可以先使用“上传图片”页面上传后把链接粘贴到这里。</p>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dlgAvatar = false">取消</el-button>
        <el-button type="primary" @click="saveAvatar">保存</el-button>
      </span>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog title="修改密码" :visible.sync="dlgPwd">
      <el-form :model="pwdForm" label-width="100px">
        <el-form-item label="旧密码"><el-input v-model="pwdForm.oldpassword" type="password" /></el-form-item>
        <el-form-item label="新密码"><el-input v-model="pwdForm.newpassword" type="password" /></el-form-item>
        <el-form-item label="确认密码"><el-input v-model="pwdForm.confirm" type="password" /></el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dlgPwd = false">取消</el-button>
        <el-button type="primary" @click="savePassword">保存</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import ImgUpLoad from '@/components/ImgUpLoad/index.vue'
import * as myApi from '@/api/user/my'
import { ElMessage } from 'element-plus'
import defaultAvatar from '@/assets/user.png'
import defaultAvatar2x from '@/assets/user@2x.png'

const nickname = ref('')
const phone = ref('')
const avatar = ref('')
// const defaultAvatar = '@/assets/user.png'

const dlgAvatar = ref(false)
const dlgPwd = ref(false)

const editNickname = ref('')
const editPhone = ref('')
const editAvatar = ref('')
const editingAvatar = ref(false)
const editingPassword = ref(false)
// inline edit state for phone and nickname
const editingPhone = ref(false)
const editingNickname = ref(false)
const codeInput = ref('')

const sending = ref(false)
const countdown = ref(0)
let _codeTimer: any = null
const sentCode = ref('')
const sentPhone = ref('')

const phoneValid = (p: string) => {
  const s = (p || '').toString().trim()
  return /^1[3-9]\d{9}$/.test(s)
}

const pwdForm = ref({ oldpassword: '', newpassword: '', confirm: '' })

onMounted(async () => {
  try {
    const p: any = await myApi.getProfile()
    nickname.value = p.username || p.nickname || ''
    phone.value = p.phone || ''
    avatar.value = p.avatar_url || p.avatar || ''
  } catch (e) {
    // ignore
  }
})

const maskedPhone = computed(() => {
  if (!phone.value) return '未绑定'
  const s = String(phone.value)
  if (s.length >= 7) return s.slice(0, 3) + '****' + s.slice(s.length-4)
  return s.replace(/.(?=.{2})/g, '*')
})

function openEditNickname() { editNickname.value = nickname.value; editingNickname.value = true }
function openEditPhone() { editPhone.value = phone.value; editingPhone.value = true }
function openEditAvatar() { editAvatar.value = avatar.value; dlgAvatar.value = true }
function openChangePassword() { pwdForm.value = { oldpassword: '', newpassword: '', confirm: '' }; dlgPwd.value = true }

async function saveNickname() {
  try {
    await myApi.updateProfile({ nickname: editNickname.value })
    nickname.value = editNickname.value
    editingNickname.value = false
    ElMessage.success('昵称已更新')
  } catch (e) { ElMessage.error('更新失败') }
}

async function savePhone() {
  try {
    // verify demo code
    if (!sentCode.value || sentPhone.value !== editPhone.value || String(codeInput.value || '') !== String(sentCode.value)) {
      ElMessage.error('验证码错误或未发送')
      return
    }
    await myApi.updateProfile({ phone: editPhone.value })
    phone.value = editPhone.value
    editingPhone.value = false
    // clear code
    codeInput.value = ''
    sentCode.value = ''
    sentPhone.value = ''
    ElMessage.success('手机号已更新')
  } catch (e) { ElMessage.error('更新失败') }
}

function sendCode() {
  if (!phoneValid(editPhone.value)) { ElMessage.warning('请输入有效的手机号'); return }
  if (sending.value) return
  sending.value = true
  const code = Math.floor(100000 + Math.random() * 900000).toString()
  sentCode.value = code
  sentPhone.value = editPhone.value
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

// cleanup timer when component unmounts
import { onBeforeUnmount } from 'vue'
onBeforeUnmount(() => { if (_codeTimer) clearInterval(_codeTimer) })

async function saveAvatar() {
  try {
    // if upload component provided a full URL, prefer that; otherwise use existing
    const url = editAvatar.value || avatar.value || ''
    await myApi.updateProfile({ avatar_url: url })
    avatar.value = url
    editingAvatar.value = false
    dlgAvatar.value = false
    ElMessage.success('头像已更新')
  } catch (e) { ElMessage.error('更新失败') }
}

async function savePassword() {
  if (pwdForm.value.newpassword !== pwdForm.value.confirm) { ElMessage.error('两次新密码不一致'); return }
  if (!pwdForm.value.oldpassword) { ElMessage.error('请输入旧密码'); return }
  try {
    await myApi.changePassword({ oldpassword: pwdForm.value.oldpassword, newpassword: pwdForm.value.newpassword })
    editingPassword.value = false
    pwdForm.value = { oldpassword: '', newpassword: '', confirm: '' }
    ElMessage.success('密码已修改')
  } catch (e) { ElMessage.error('旧密码错误') }
}
</script>

<style scoped>
.my-bg { width: 100%; min-height: 100vh; background: url('/src/assets/login/img_denglu_bj.jpg') center/cover no-repeat; background-attachment: fixed; display: flex; justify-content: center; align-items: flex-start; padding: 60px 0; }
.my-page { width: 60%; background: rgba(255, 248, 225, 0.96); border-radius: 16px; box-shadow: 0 8px 24px rgba(255, 193, 7, 0.35); padding: 28px; backdrop-filter: blur(6px); transition: 0.3s; position: relative; z-index: 2; }
.settings-card { padding: 0; }
.profile-content { display: flex; gap: 18px; padding: 18px; align-items: center; }
.left { display:flex; flex-direction:column; align-items:center; gap:8px; padding: 18px }
.info-col { flex: 1; padding: 12px 18px }
.profile-content { flex-wrap: wrap; }
.avatar-preview img { width: 100%; height: 100%; object-fit: cover; display: block }
.avatar-uploader-inline { margin-top: 12px; }
.avatar-box { display:flex; flex-direction:column; align-items:center }
.field-row { display:flex; align-items:center; gap:12px; padding: 8px 0; }
.field-label { width: 80px; color:#7a5a14; font-weight:600 }
.field-value { flex:1 }
.field-action { width: 80px; text-align:right }

@media(max-width:900px){ .my-page { width: 92%; padding: 12px } }
@media(max-width:600px){
  .profile-content { flex-direction: column; align-items: flex-start }
  .left { width: 100%; }
  .info-col { width: 100%; padding: 12px 0 }
}
</style>
