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

const props = defineProps<{ data: any }>()

watch(
  () => props.data,
  async () => {
    console.log("props.data 实际收到：", props.data)
    await nextTick()
    initChart()
  },
  { immediate: true }
)

function initChart() {
  const chartDom = document.getElementById('main') as HTMLElement
  if (!chartDom) return

  const myChart = echarts.init(chartDom)

  // === 关键修复：把字符串分割成数组 ===
const dateList = (props.data?.dateList || '')
  .split(',')
  .filter(Boolean)

const turnoverList = (props.data?.turnoverList || '')

  console.log(dateList)
  console.log(turnoverList)
  const option = {
    tooltip: { trigger: 'axis' },
    grid: { 
      top: '10%', 
      left: '15', 
      right: '50',    // 原来写 '50' 太大了，会把图挤变形，建议改小
      bottom: '15%', 
      containLabel: true 
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      axisLabel: { 
        color: '#666', 
        fontSize: 12,
        interval: 0,  // 强制显示所有日期（数据少时很有用）
      },
      axisLine: { lineStyle: { color: '#E5E4E4' } },
      data: dateList,
    },
    yAxis: {
      type: 'value',
      min: 0,
      axisLabel: { color: '#666', fontSize: 12 },
      splitLine: { lineStyle: { color: '#f0f0f0' } },
    },
    series: [
      {
        name: '营业额',
        type: 'line',
        smooth: false,
        showSymbol: true,           // 建议开启点，便于看数据
        symbolSize: 8,
        itemStyle: {
          color: '#F29C1B',
          borderWidth: 3,
          borderColor: '#fff',
        },
        lineStyle: { color: '#FFD000', width: 3 },
        areaStyle: {                        // 加个面积图更好看（美团/饿了么风格）
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(255, 208, 0, 0.3)' },
            { offset: 1, color: 'rgba(255, 208, 0, 0)' },
          ]),
        },
        data: turnoverList,
      },
    ],
  }

  myChart.setOption(option)
}
</script>
