<template>
  <div class="dashboard-container home">
    <!-- 营业数据 -->
    <Overview :overviewData="overviewData" />

    <!-- 订单管理 -->
    <Orderview :orderviewData="orderviewData" />

    <div class="homeMain">
      <!-- 菜品总览 -->
      <CuisineStatistics :dishesData="dishesData" />
      <!-- 套餐总览 -->
      <SetMealStatistics :setMealData="setMealData" />
    </div>

    <!-- 订单信息 -->
    <OrderList
      :order-statics="orderStatics"
      @getOrderListBy3Status="getOrderListBy3Status"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'

// 引入 API
import {
  getBusinessData,
  getOrderData,
  getOverviewDishes,
  getSetMealStatistics,
} from '@/api/merchant/index'
import { getOrderListBy } from '@/api/merchant/order'

// 引入组件
import Overview from './components/overview.vue'
import Orderview from './components/orderview.vue'
import CuisineStatistics from './components/cuisineStatistics.vue'
import SetMealStatistics from './components/setMealStatistics.vue'
import OrderList from './components/orderList.vue'

const overviewData = ref<any>({})
const orderviewData = ref<any>({})
const dishesData = ref<any>({})
const setMealData = ref<any>({})
const orderStatics = ref<any>({})

const page = ref(1)
const pageSize = ref(10)
const status = ref(2)


// 初始化
const init = async () => {
  await nextTick()
  await getBusinessDataFn()
  await getOrderStatisticsData()
  await getOverStatisticsData()
  await getSetMealStatisticsData()
  await getOrderListBy3Status()
}

// 获取营业数据
const getBusinessDataFn = async () => {
  try {
    const res = await getBusinessData()
    overviewData.value = res.data.data
  } catch (err: any) {
    ElMessage.error('获取营业数据失败：' + err.message)
  }
}

// 获取今日订单
const getOrderStatisticsData = async () => {
  try {
    const res = await getOrderData()
    orderviewData.value = res.data.data
  } catch (err: any) {
    ElMessage.error('获取今日订单失败：' + err.message)
  }
}

// 获取菜品总览数据
const getOverStatisticsData = async () => {
  try {
    const res = await getOverviewDishes()
    dishesData.value = res.data.data
  } catch (err: any) {
    ElMessage.error('获取菜品数据失败：' + err.message)
  }
}

// 获取套餐总览数据
const getSetMealStatisticsData = async () => {
  try {
    const res = await getSetMealStatistics()
    setMealData.value = res.data.data
  } catch (err: any) {
    ElMessage.error('获取套餐数据失败：' + err.message)
  }
}

// 获取待处理、待派送、派送中数量
const getOrderListBy3Status = async () => {
  try {
    const res = await getOrderListBy({})
    if (Number(res.data.code) === 1) {
      orderStatics.value = res.data.data
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err: any) {
    ElMessage.error('请求出错：' + err.message)
  }
}


onMounted(() => {
  init()
})
</script>

<style lang="scss" scoped>
.dashboard-container {
  padding: 20px;
}

.homeMain {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  margin-top: 20px;
}
</style>
