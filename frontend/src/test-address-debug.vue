<template>
  <div class="address-debug">
    <h2>地址解析调试工具</h2>

    <div class="test-section">
      <h3>测试地址</h3>
      <el-input v-model="testAddress" placeholder="输入测试地址" />
      <el-button @click="testGeocoding" type="primary" style="margin-top: 10px;">
        测试地址解析
      </el-button>
    </div>

    <div class="test-section">
      <h3>测试订单数据</h3>
      <div v-if="testOrder">
        <p><strong>餐厅:</strong> {{ testOrder.restaurant }}</p>
        <p><strong>商家地址:</strong> {{ testOrder.pickupAddress }}</p>
        <p><strong>客户:</strong> {{ testOrder.customer }}</p>
        <p><strong>配送地址:</strong> {{ testOrder.deliveryAddress }}</p>
      </div>

      <el-button @click="testOrderData" type="success" style="margin-top: 10px;">
        模拟订单数据测试
      </el-button>
    </div>

    <div class="test-section">
      <h3>地图弹窗测试</h3>
      <el-button @click="testMerchantMap" type="info">测试商家位置</el-button>
      <el-button @click="testUserMap" type="warning" style="margin-left: 10px;">测试用户位置</el-button>
    </div>

    <!-- 地图弹窗 -->
    <AmapModal
      v-model="showMapModal"
      :merchant-data="merchantData"
      :user-data="userData"
      :default-location="[113.299, 23.099]"
      :initial-location-type="initialType"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import AmapModal from './components/AmapModal.vue'
import amapLoader from './utils/amap'

const testAddress = ref('中山大学珠海校区榕园4号')
const showMapModal = ref(false)
const merchantData = ref<any>(null)
const userData = ref<any>(null)
const initialType = ref<'merchant' | 'user'>('merchant')

const testOrder = ref<any>({
  restaurant: '测试餐厅',
  pickupAddress: '广东省珠海市香洲区中山大学珠海校区榕园',
  customer: '张三',
  deliveryAddress: '广东省珠海市香洲区中山大学珠海校区荔园'
})

const testGeocoding = async () => {
  if (!testAddress.value) {
    ElMessage.warning('请输入测试地址')
    return
  }

  try {
    console.log('开始测试地址解析:', testAddress.value)
    const AMap = await amapLoader.load({
      plugins: ['AMap.Geocoder']
    })

    const geocoder = new AMap.Geocoder({
      city: '珠海'
    })

    geocoder.getLocation(testAddress.value, (status: string, result: any) => {
      console.log('地址解析结果:', { status, result })

      if (status === 'complete' && result.geocodes && result.geocodes.length > 0) {
        const location = result.geocodes[0].location
        ElMessage.success(`解析成功: ${location.lng}, ${location.lat}`)
        console.log('解析成功:', {
          地址: testAddress.value,
          经度: location.lng,
          纬度: location.lat,
          级别: result.geocodes[0].level,
          匹配度: result.geocodes[0].confidence
        })
      } else {
        ElMessage.error('解析失败: ' + (result?.info || '未知错误'))
        console.error('解析失败:', { status, result })
      }
    })
  } catch (error) {
    console.error('测试失败:', error)
    ElMessage.error('测试失败: ' + error)
  }
}

const testOrderData = () => {
  console.log('测试订单数据:', testOrder.value)
  ElMessage.info('订单数据已输出到控制台，请检查地址字段')
}

const testMerchantMap = () => {
  merchantData.value = {
    title: testOrder.value.restaurant,
    address: testOrder.value.pickupAddress,
    type: 'merchant'
  }
  userData.value = null
  initialType.value = 'merchant'
  showMapModal.value = true

  console.log('测试商家地图:', merchantData.value)
}

const testUserMap = () => {
  merchantData.value = null
  userData.value = {
    title: testOrder.value.customer,
    address: testOrder.value.deliveryAddress,
    type: 'user'
  }
  initialType.value = 'user'
  showMapModal.value = true

  console.log('测试用户地图:', userData.value)
}
</script>

<style scoped>
.address-debug {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.test-section {
  margin-bottom: 30px;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
}

h3 {
  margin-top: 0;
  color: #333;
}

p {
  margin: 8px 0;
  line-height: 1.5;
}
</style>