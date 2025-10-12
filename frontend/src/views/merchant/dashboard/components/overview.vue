<template>
  <div class="container">
    <h2 class="homeTitle">
      今日数据<i>{{ days[1] }}</i
      ><span><router-link to="statistics">详细数据</router-link></span>
    </h2>
    <div class="overviewBox">
      <ul>
        <li>
          <p class="tit">营业额</p>
          <p class="num">¥ {{ overviewData.turnover }}</p>
        </li>
        <li>
          <p class="tit">有效订单</p>
          <p class="num">{{ overviewData.validOrderCount }}</p>
        </li>
        <li>
          <p class="tit">订单完成率</p>
          <p class="num">
            {{ (overviewData.orderCompletionRate * 100).toFixed(0) }}%
          </p>
        </li>
        <li>
          <p class="tit">平均客单价</p>
          <p class="num">¥ {{ overviewData.unitPrice }}</p>
        </li>

        <li>
          <p class="tit">新增用户</p>
          <p class="num">{{ overviewData.newUsers }}</p>
        </li>
      </ul>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { getday } from '@/utils/formValidate'

const props = defineProps<{ overviewData?: any }>()

const days = getday()

const overviewData = computed(() => {
  return (
    props.overviewData || {
      turnover: 0,
      validOrderCount: 0,
      orderCompletionRate: 0,
      unitPrice: 0,
      newUsers: 0,
    }
  )
})
</script>
