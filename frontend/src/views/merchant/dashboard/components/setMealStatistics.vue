<template>
  <div class="container">
    <h2 class="homeTitle">
      套餐总览
      <span>
        <router-link to="/merchant/meal">套餐管理</router-link>
      </span>
    </h2>
    <div class="orderviewBox">
      <ul>
        <li>
          <span class="status">
            <i class="iconfont icon-open"></i>已启售
          </span>
          <span class="num">{{ setMealData?.sold ??0}}</span>
        </li>
        <li>
          <span class="status"><i class="iconfont icon-stop"></i>已停售</span>
          <span class="num">{{ setMealData?.discontinued??0 }}</span>
        </li>
        <li class="add">
          <router-link to="setmeal/add">
            <i class="iconfont icon-plus"></i>
            <p>新增套餐</p>
          </router-link>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  setMealData?: {
    sold?: number
    discontinued?: number
    [key: string]: any
  }
}>()

const setMealData = computed(() => {
  return props.setMealData || { sold: 0, discontinued: 0 }
})
</script>
<style scoped lang="scss">
.container {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
}

.homeTitle {
  font-size: 18px;
  font-weight: 600;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.homeTitle span a {
  color: #409eff;
  text-decoration: none;
}

.orderviewBox ul {
  display: flex;
  justify-content: space-between;
  align-items: stretch;
  gap: 20px;
  list-style: none;
  padding: 0;
  margin: 0;
}

.orderviewBox li {
  flex: 1;
  text-align: center;
  background: #f5f7fa;
  border-radius: 10px;
  padding: 20px 0;
  font-size: 16px;
  color: #333;
  transition: all 0.3s ease;
}

.orderviewBox li:hover {
  background-color: #e8f3ff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.status {
  display: block;
  color: #666;
  font-size: 14px;
  margin-bottom: 8px;
}

.num {
  font-size: 28px;
  font-weight: 700;
  color: #409eff;
}

.add {
  background-color: #409eff;
  color: #fff;
  cursor: pointer;
}

.add a {
  color: #fff;
  text-decoration: none;
}

.add:hover {
  background-color: #66b1ff;
}

.iconfont {
  margin-right: 6px;
}
</style>
