<template>
  <div class="dashboard-container home">
    <!-- 标题 -->
    <h2>数据统计</h2>
    <TitleIndex @sendTitleInd="getTitleNum" :flag="flag" :tateData="tateData" />
    <!-- end -->
    <div class="homeMain">
      <!-- 营业额统计 -->
      <TurnoverStatistics :turnoverdata="turnoverData" />
      <!-- end -->
      <!-- 用户统计 -->
      <UserStatistics :userdata="userData" />
      <!-- end -->
    </div>
    <div class="homeMain homecon">
      <!-- 订单统计 -->
      <OrderStatistics :orderdata="orderData" :overviewData="overviewData" />
      <!-- end -->
      <!-- 销量排名TOP10 -->
      <Top :top10data="top10Data" />
      <!-- end -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick, onMounted } from 'vue'
import {
  get1stAndToday,
  past7Day,
  past30Day,
  pastWeek,
  pastMonth,
} from '@/utils/formValidate'
import {
  getDataOverView, // 数据概览
  getTurnoverStatistics,
  getUserStatistics,
  getOrderStatistics,
  getTop,
} from '@/api/merchant/index'

import TitleIndex from './components/titleIndex.vue'
import TurnoverStatistics from './components/turnoverStatistics.vue'
import UserStatistics from './components/userStatistics.vue'
import OrderStatistics from './components/orderStatistics.vue'
import Top from './components/top10.vue'

const overviewData = ref<any>({})
const flag = ref(2)
const tateData = ref<any[]>([])
const turnoverData = ref<any>({})
const userData = ref<any>({})
const orderData = reactive<any>({ data: {} })
const top10Data = ref<any>({})

function init(begin: any, end: any) {
  nextTick(() => {
    getTurnoverStatisticsData(begin, end)
    getUserStatisticsData(begin, end)
    getOrderStatisticsData(begin, end)
    getTopData(begin, end)
  })
}

async function getTurnoverStatisticsData(begin: any, end: any) {
  const data = await getTurnoverStatistics({ begin: begin, end: end })
  const d = data.data.data
  turnoverData.value = {
    dateList: d.dateList.split(','),
    turnoverList: d.turnoverList.split(','),
  }
}

async function getUserStatisticsData(begin: any, end: any) {
  const data = await getUserStatistics({ begin: begin, end: end })
  const d = data.data.data
  userData.value = {
    dateList: d.dateList.split(','),
    totalUserList: d.totalUserList.split(','),
    newUserList: d.newUserList.split(',')
  }
}

async function getOrderStatisticsData(begin: any, end: any) {
  const data = await getOrderStatistics({ begin: begin, end: end })
  const d = data.data.data
  orderData.data = {
    dateList: d.dateList.split(','),
    orderCountList: d.orderCountList.split(','),
    validOrderCountList: d.validOrderCountList.split(','),
  }
  orderData.totalOrderCount = d.totalOrderCount
  orderData.validOrderCount = d.validOrderCount
  orderData.orderCompletionRate = d.orderCompletionRate
}

async function getTopData(begin: any, end: any) {
  const data = await getTop({ begin: begin, end: end })
  const d = data.data.data
  top10Data.value = {
    nameList: d.nameList.split(',').reverse(),
    numberList: d.numberList.split(',').reverse(),
  }
}

function getTitleNum(data: number) {
  switch (data) {
    case 1:
      tateData.value = get1stAndToday()
      break
    case 2:
      tateData.value = past7Day()
      break
    case 3:
      tateData.value = past30Day()
      break
    case 4:
      tateData.value = pastWeek()
      break
    case 5:
      tateData.value = pastMonth()
      break
  }
  init(tateData.value[0], tateData.value[1])
}

onMounted(() => {
  getTitleNum(2)
})
</script>
