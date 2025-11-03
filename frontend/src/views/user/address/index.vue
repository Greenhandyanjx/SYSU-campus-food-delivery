<template>
  <div class="address-page">
    <div class="address-header">
      <h2>地址管理</h2>
      <el-button type="primary" @click="openAdd">新增地址</el-button>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="我的收货地址" name="mine">
        <div class="addr-list">
          <el-card v-for="(a, i) in myAddresses" :key="i" class="addr-card">
            <div class="addr-row">
              <div>
                <div class="addr-name">{{ a.name }} <span class="addr-tag">{{ a.tag }}</span></div>
                <div class="addr-detail">{{ a.detail }}</div>
              </div>
              <div class="addr-actions">
                <el-button type="text" size="small" @click="editAddress(i)">编辑</el-button>
                <el-button type="text" size="small" @click="removeAddress(i)">删除</el-button>
              </div>
            </div>
          </el-card>
          <div v-if="myAddresses.length === 0" class="empty">你还没有收货地址，点击“新增地址”添加。</div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="附近地址" name="nearby">
        <div class="nearby-list">
          <el-card v-for="(a, i) in nearbyAddresses" :key="i" class="addr-card">
            <div class="addr-row">
              <div>
                <div class="addr-name">{{ a.name }}</div>
                <div class="addr-detail">{{ a.detail }}</div>
              </div>
              <div class="addr-actions">
                <el-button type="primary" size="small" @click="useNearby(a)">选择</el-button>
              </div>
            </div>
          </el-card>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog :model-value="showDialog" title="新增地址" width="520px">
      <el-form :model="form">
        <el-form-item label="收货人" label-width="80px">
          <el-input v-model="form.name" placeholder="姓名" />
        </el-form-item>
        <el-form-item label="电话" label-width="80px">
          <el-input v-model="form.phone" placeholder="手机号码" />
        </el-form-item>
        <el-form-item label="地址" label-width="80px">
          <el-input v-model="form.detail" placeholder="省/市/区 详细地址" />
        </el-form-item>
        <el-form-item label="标签" label-width="80px">
          <el-input v-model="form.tag" placeholder="例如：家、公司" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="closeDialog">取消</el-button>
        <el-button type="primary" @click="saveAddress">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const activeTab = ref('mine')
const showDialog = ref(false)

const myAddresses = ref<any[]>([
  { name: '张三', phone: '13800000000', detail: '广东省 广州市 中山大学', tag: '家' },
])

const nearbyAddresses = ref<any[]>([
  { name: '教学楼南门', detail: '中山大学南门旁' },
  { name: '学生食堂', detail: '第一食堂附近' },
])

const form = ref({ name: '', phone: '', detail: '', tag: '' })

function openAdd() {
  form.value = { name: '', phone: '', detail: '', tag: '' }
  showDialog.value = true
}

function closeDialog() { showDialog.value = false }

function saveAddress() {
  if (!form.value.detail || !form.value.name) {
    // 简单校验
    return
  }
  myAddresses.value.push({ ...form.value })
  showDialog.value = false
}

function editAddress(i: number) {
  const a = myAddresses.value[i]
  form.value = { ...a }
  showDialog.value = true
  // 删除原条目，保存时会追加或你也可以替换
  myAddresses.value.splice(i, 1)
}

function removeAddress(i: number) { myAddresses.value.splice(i, 1) }

function useNearby(a: any) {
  // 将附近地址作为收货地址示例加入我的地址并切到我的地址页
  myAddresses.value.push({ name: a.name, phone: '', detail: a.detail, tag: '附近' })
  activeTab.value = 'mine'
}

// 如果从 navbar 跳过来时需要特殊处理，可在 mounted 中读取 route.query

</script>

<style scoped>
.address-page { padding: 18px }
.address-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:12px }
.addr-list { display:flex; flex-direction:column; gap:10px }
.addr-card { padding: 12px }
.addr-row { display:flex; justify-content:space-between; align-items:center }
.addr-name { font-weight:600 }
.addr-tag { margin-left:8px; padding:2px 8px; background:#eaf6ff; color:#154b75; border-radius:6px; font-size:12px }
.addr-detail { color:#666; margin-top:6px }
.empty { color:#999; padding:18px; background:#fff; border-radius:8px }

.nearby-list { display:flex; flex-direction:column; gap:8px }

</style>
