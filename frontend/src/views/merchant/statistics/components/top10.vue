<template>
  <div class="container top10">
    <h2 class="homeTitle">销量排名TOP10</h2>
    <div class="charBox">
      <div id="top" style="width: 100%; height: 380px"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, watch, nextTick, onMounted } from 'vue'
import * as echarts from 'echarts'

type EChartsInstance = ReturnType<typeof echarts.init>

let myChart: EChartsInstance | null = null
const props = defineProps<{ top10data: any }>()

onMounted(() => {
  // 初次渲染时初始化
  initChart()
})

watch(
  () => props.top10data,
  async () => {
    await nextTick()
    initChart()
  },
  { immediate: true }
)

function initChart() {
  const chartDom = document.getElementById('top')
  if (!chartDom) return

  if (!myChart) {
    myChart = echarts.init(chartDom)
  }

  const option = {
    tooltip: { trigger: 'axis' },
    grid: { top: 10, left: 0, right: 0, bottom: 0, containLabel: true },
    xAxis: { show: false },
    yAxis: {
      type: 'category',
      data: props.top10data?.nameList || []
    },
    series: [
      {
        type: 'bar',
        data: props.top10data?.numberList || [],
        showBackground: true,
        backgroundStyle: { color: '#F3F4F7' },
        itemStyle: {
          color: new echarts.graphic.LinearGradient(1, 0, 0, 0, [
            { offset: 0, color: '#FFBD00' },
            { offset: 1, color: '#FFD000' }
          ])
        }
      }
    ]
  }

  myChart.setOption(option)
  myChart.resize()
}
</script>
