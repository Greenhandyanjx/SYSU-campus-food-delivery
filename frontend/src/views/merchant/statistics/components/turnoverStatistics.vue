<template>
  <div class="container">
    <h2 class="homeTitle">营业额统计</h2>
    <div class="charBox">
      <div id="main" style="width: 100%; height: 320px"></div>
      <ul class="orderListLine turnover">
        <li>营业额(元)</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, watch, nextTick } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{ turnoverdata: any }>()

watch(
  () => props.turnoverdata,
  async () => {
    await nextTick()
    initChart()
  },
  { immediate: true }
)

function initChart() {
  const chartDom = document.getElementById('main') as any
  if (!chartDom) return
  const myChart = echarts.init(chartDom)
  const option: any = {
    tooltip: { trigger: 'axis' },
    grid: { top: '5%', left: '10', right: '50', bottom: '12%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      axisLabel: { textStyle: { color: '#666', fontSize: '12px' } },
      axisLine: { lineStyle: { color: '#E5E4E4', width: 1 } },
      data: props.turnoverdata?.dateList || [],
    },
    yAxis: [{ type: 'value', min: 0, axisLabel: { textStyle: { color: '#666', fontSize: '12px' } } }],
    series: [{ name: '营业额', type: 'line', smooth: false, showSymbol: false, symbolSize: 10, itemStyle: { normal: { color: '#F29C1B', lineStyle: { color: '#FFD000' } }, emphasis: { color: '#fff', borderWidth: 5, borderColor: '#FFC100' } }, data: props.turnoverdata?.turnoverList || [] }],
  }
  option && myChart.setOption(option)
}
</script>
