<template>
  <div class="container">
    <h2 class="homeTitle">
      <img src="/JDlogo.png" class="jd-logo" alt="嘉递外卖" />
      订单管理<i>{{ days[1] }}</i>
      ><span><router-link to="/merchant/orders">订单明细</router-link></span>
    </h2>
    <div class="orderviewBox">
      <ul>
        <li>
          <span class="status"
            ><i class="iconfont icon-waiting"></i>待接单</span
          >
          <span class="num tip"
            ><router-link to="/order?status=2">{{ Number.isFinite(Number(orderviewData.waitingOrders)) ? orderviewData.waitingOrders : 0 }}</router-link></span>
          >
        </li>
        <li>
          <span class="status"
            ><i class="iconfont icon-staySway"></i>待派送</span
          >
          <span class="num tip"
            ><router-link to="/order?status=3">{{ Number.isFinite(Number(orderviewData.deliveredOrders)) ? orderviewData.deliveredOrders : 0 }}</router-link></span>
          >
        </li>
        <li>
          <span class="status"
            ><i class="iconfont icon-complete"></i>已完成</span
          >
          <span class="num"
            ><router-link to="/order?status=5">{{ Number.isFinite(Number(orderviewData.completedOrders)) ? orderviewData.completedOrders : 0 }}</router-link></span>
          >
        </li>
        <li>
          <span class="status"><i class="iconfont icon-cancel"></i>已取消</span>
          <span class="num"
            ><router-link to="/order?status=6">{{ Number.isFinite(Number(orderviewData.cancelledOrders)) ? orderviewData.cancelledOrders : 0 }}</router-link></span>
          >
        </li>
        <li>
          <span class="status"><i class="iconfont icon-all"></i>全部订单</span>
          <span class="num"
            ><router-link to="/order">{{ Number.isFinite(Number(orderviewData.allOrders)) ? orderviewData.allOrders : 0 }}</router-link></span>
          >
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { getday } from '@/utils/formValidate'

const props = defineProps<{ orderviewData?: any }>()

const days = getday()

const orderviewData = computed(() => {
  return (
    props.orderviewData || {
      waitingOrders: 0,
      deliveredOrders: 0,
      completedOrders: 0,
      cancelledOrders: 0,
      allOrders: 0,
    }
  )
})
</script>

<style scoped>
.jd-logo { width:36px; height:36px; object-fit:contain; vertical-align:middle; margin-right:8px; }
</style>
