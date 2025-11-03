<template>
  <div class="container">
    <h2 class="homeTitle">订单统计</h2>
    <div class="charBox">
      <div class="orderProportion">
        <div>
          <p>订单完成率</p>
          <p>{{ (orderdata.orderCompletionRate * 100).toFixed(1) }}%</p>
        </div>
        <div class="symbol">=</div>
        <div>
          <p>有效订单</p>
          <p>{{ orderdata.validOrderCount }}</p>
        </div>
        <div class="symbol">/</div>
        <div>
          <p>订单总数</p>
          <p>{{ orderdata.totalOrderCount }}</p>
        </div>
      </div>
      <div id="ordermain" style="width: 100%; height: 300px"></div>
      <ul class="orderListLine">
        <li class="one"><span></span>订单总数（个）</li>
        <li class="three"><span></span>有效订单（个）</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, watch, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{ orderdata: any; overviewData?: any }>()

watch(
  () => props.orderdata,
  async () => {
    await nextTick()
    initChart()
  },
  { immediate: true }
)

function initChart() {
  const chartDom = document.getElementById('ordermain') as any
  if (!chartDom) return
  const myChart = echarts.init(chartDom)
  const option: any = {
    tooltip: { trigger: 'axis', backgroundColor: '#fff', borderRadius: 2, textStyle: { color: '#333', fontSize: 12, fontWeight: 300 } },
    grid: { top: '5%', left: '20', right: '50', bottom: '12%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, axisLabel: { textStyle: { color: '#666', fontSize: '12px' } }, axisLine: { lineStyle: { color: '#E5E4E4', width: 1 } }, data: props.orderdata?.data?.dateList || [] },
    yAxis: [{ type: 'value', min: 0, interval: 50, axisLabel: { textStyle: { color: '#666', fontSize: '12px' } } }],
    series: [
      { name: '订单总数', type: 'line', smooth: false, showSymbol: false, symbolSize: 10, itemStyle: { normal: { color: '#FFD000', lineStyle: { color: '#FFD000' } }, emphasis: { color: '#fff', borderWidth: 5, borderColor: '#FFC100' } }, data: props.orderdata?.data?.orderCountList || [] },
      { name: '有效订单', type: 'line', smooth: false, showSymbol: false, symbolSize: 10, itemStyle: { normal: { color: '#FD7F7F', lineStyle: { color: '#FD7F7F' } }, emphasis: { color: '#fff', borderWidth: 5, borderColor: '#FD7F7F' } }, data: props.orderdata?.data?.validOrderCountList || [] },
    ],
  }
  option && myChart.setOption(option)
}
</script>
