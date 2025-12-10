<template>
  <div>
    <button class="chat-launcher" @click.stop="openChat">
      <img src="/JDlogo.png" alt="嘉递" class="chat-launcher-logo" />
      <span class="chat-launcher-text">{{ triggerText }}</span>
    </button>

    <teleport to="body">
      <div v-if="showChat" class="chat-overlay" @click.self="closeChat">
        <div class="chat-modal" role="dialog" aria-modal="true">
          <!-- <div class="chat-header">
            <div class="title">与 {{ merchantNameLocal }} 聊天</div>
            <button class="close" @click="closeChat">✕</button>
          </div> -->
          <div class="chat-body">
            <ChatWindow
              :merchantId="merchantIdLocal"
              :merchantName="merchantNameLocal"
              :merchantAvatar="merchantAvatarLocal"
              :userBaseId="userBaseId"
              :userAvatar="userAvatar"
              :token="token"
              @close="closeChat"
            />
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ChatWindow from '@/components/Chat/ChatWindow.vue'
import { getMerchantDetail, getBaseUserDetail } from '@/api/chat'
import { ElMessage } from 'element-plus'

const props = defineProps({
  merchantId: { type: [String, Number], required: false },
  merchantName: { type: String, default: '' },
  merchantAvatar: { type: String, default: '' },
  triggerText: { type: String, default: '联系商家' },
})

const showChat = ref(false)
const merchantIdLocal = ref(props.merchantId)
const merchantNameLocal = ref(props.merchantName)
const merchantAvatarLocal = ref(props.merchantAvatar)

const userBaseId = ref(null)
const userAvatar = ref('')
const token = (typeof window !== 'undefined' && localStorage.getItem('token')) || ''

async function openChat() {
  if (!merchantIdLocal.value) {
    ElMessage.error('无法定位商家 ID，无法发起聊天')
    return
  }

  try {
    // 尝试拉取商家详情（补充名称或头像）
    if (!merchantNameLocal.value || !merchantAvatarLocal.value) {
      const r = await getMerchantDetail(merchantIdLocal.value)
      const data = r && r.data && (r.data.data || r.data)
      if (data) {
        merchantNameLocal.value = merchantNameLocal.value || data.shop_name || data.shopName || data.name || merchantNameLocal.value
        merchantAvatarLocal.value = merchantAvatarLocal.value || data.logo || data.logoUrl || merchantAvatarLocal.value
      }
    }

    // 拉取当前登录用户基础信息
    const u = await getBaseUserDetail()
    const ud = u && u.data && (u.data.data || u.data)
    if (ud) {
      userBaseId.value = ud.id
      userAvatar.value = ud.avatar || ud.avatarUrl || ''
    }

    showChat.value = true
  } catch (e: any) {
    ElMessage.error('打开聊天失败：' + (e?.message || e))
  }
}

function closeChat() {
  showChat.value = false
}
</script>

<style scoped>
.chat-launcher {
  padding: 6px 10px;
  border-radius: 18px;
  background: linear-gradient(90deg,#ffb300,#ff8c00);
  color: white;
  border: none;
  cursor: pointer;
  font-size: 10px;
}
.chat-launcher-logo {
  width: 18px;
  height: 18px;
  vertical-align: middle;
  margin-right: 6px;
  border-radius: 3px;
}
.chat-launcher-text { vertical-align: middle }
.chat-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2147483647; /* very high to ensure top-most */
  -webkit-font-smoothing: antialiased;
}
.chat-modal { width: 400px; max-width: 92%; background: #fff; border-radius: 10px; overflow: hidden; display:flex; flex-direction:column; z-index:2147483648; position:relative }
.chat-header { display:flex; justify-content:space-between; align-items:center; padding:10px 14px; border-bottom:1px solid #eee; }
.chat-header .title { font-weight:700 }
.chat-header .close { border: none; background: transparent; font-size:10px; cursor:pointer }
.chat-body { height: 700px; }
</style>
