<template>
  <div  class="addDish">
    <div class="leftCont">
      <div v-show="seachKey.trim() == ''"
           class="tabBut">
        <span v-for="(item, index) in dishType"
              :key="index"
              :class="{ act: index == keyInd }"
              @click="checkTypeHandle(index, item.id)">{{ item.name }}</span>
      </div>
      <div class="tabList">
        <div class="table"
             :class="{ borderNone: !dishList.length }">
          <div v-if="dishList.length == 0"
               style="padding-left: 10px">
            <Empty />
          </div>
          <el-checkbox-group v-else
                             v-model="checkedList"
                             @change="checkedListHandle">
            <div v-for="(item, index) in dishList"
                 :key="item.name + item.id"
                 class="items">
              <el-checkbox :key="index"
                           :label="item.name"
                           :disabled="item.status === 0">
                <div class="item">
                  <span style="flex: 3; text-align: left">{{
                    item.dishName
                  }}</span>
                  <span :style="{ color: item.status === 0 ? '#999' : '#67c23a' }">
                    {{ item.status === 0 ? '停售' : '在售' }}
                  </span>
                  <span>￥{{ Number(item.price).toFixed(2) }}</span>
                </div>
              </el-checkbox>
            </div>
          </el-checkbox-group>
        </div>
      </div>
    </div>
    <div class="ritCont">
        <div class="tit">
        已选菜品({{ checkedListAll.length }})
      </div>
      <div class="items">
        <div v-for="(item, ind) in checkedListAll"
             :key="ind"
             class="item">
          <span>{{ item.dishName || item.name }}</span>
          <span class="price">￥ {{ Number(Number(item.price).toFixed(2)) }} </span>
      <span class="del"
        @click="delCheck(item.name)">
      <img src="/src/assets/icons/btn_clean@2x.png"
         alt="">
      </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, toRef } from 'vue'
import { ElMessage } from 'element-plus'
import { getCategoryList, queryDishList } from '@/api/merchant/dish'
import Empty from '@/components/Empty/index.vue'

const props = defineProps<{
  checkList?: any[]
  seachKey?: string
}>()
const emit = defineEmits<{
  (e: 'check-list', payload: any[]): void
}>()

const dishType = ref<any[]>([])
const dishList = ref<any[]>([])
const allDishList = ref<any[]>([])
const dishListCache = ref<any[]>([])
const keyInd = ref(0)
const searchValue = ref('')
const checkedList = ref<any[]>([])
const checkedListAll = ref<any[]>([])
const ids = ref(new Set())

const seachKey = toRef(props, 'seachKey')

function init() {
  getDishType()
  checkedList.value = (props.checkList || []).map((it: any) => it.name)
  // copy and reverse to keep original behavior without mutating prop
  checkedListAll.value = [...(props.checkList || [])].reverse()
}

function getDishType() {
  getCategoryList({ type: 1 }).then((res: any) => {
    if (res && res.data && Number(res.data.code) === 1) {
      dishType.value = res.data.data || []
      if (dishType.value.length) {
        getDishList(dishType.value[0].id)
      }
    } else {
      ElMessage.error(res?.data?.msg || '获取菜品分类失败')
    }
  })
}

function getDishList(id: number) {
  queryDishList({ categoryId: id }).then((res: any) => {
    if (res && res.data && Number(res.data.code) === 1) {
      if (!res.data.data || res.data.data.length === 0) {
        dishList.value = []
        return
      }
      const newArr = res.data.data.map((n: any) => ({
        ...n,
        dishId: n.id,
        copies: 1,
        dishName: n.name
      }))
      dishList.value = newArr
      if (!ids.value.has(id)) {
        allDishList.value = [...allDishList.value, ...newArr]
      }
      ids.value.add(id)
    } else {
      ElMessage.error(res?.data?.msg || '获取菜品失败')
    }
  })
}

function getDishForName(name: any) {
  queryDishList({ name }).then((res: any) => {
    if (res && res.data && Number(res.data.code) === 1) {
      const newArr = (res.data.data || []).map((n: any) => ({
        ...n,
        dishId: n.id,
        dishName: n.name
      }))
      dishList.value = newArr
    } else {
      ElMessage.error(res?.data?.msg || '搜索菜品失败')
    }
  })
}

function checkTypeHandle(ind: number, id: any) {
  keyInd.value = ind
  getDishList(id)
}

function checkedListHandle(value: string[]) {
  // reverse to mimic original behavior
  checkedListAll.value.reverse()
  const list = allDishList.value.filter((item: any) => value.includes(item.name))

  const dishListCat = [...checkedListAll.value, ...list]
  let arrData: any[] = []
  checkedListAll.value = dishListCat.filter((item: any) => {
    let selected
    if (arrData.length === 0) {
      arrData.push(item.name)
      selected = item
    } else {
      const st = arrData.some(it => item.name === it)
      if (!st) {
        arrData.push(item.name)
        selected = item
      }
    }
    return selected
  })
  // 如果是减菜
  if (value.length < arrData.length) {
    checkedListAll.value = checkedListAll.value.filter((item: any) => value.some(it => it === item.name))
  }
  emit('check-list', checkedListAll.value)
  checkedListAll.value.reverse()
}

function open(done: any) {
  dishListCache.value = JSON.parse(JSON.stringify(props.checkList || []))
}

function close(done: any) {
  // inform parent to reset checkList to cached value
  emit('check-list', dishListCache.value)
}

function delCheck(name: any) {
  const index = checkedList.value.findIndex(it => it === name)
  const indexAll = checkedListAll.value.findIndex((it: any) => it.name === name)

  if (index > -1) checkedList.value.splice(index, 1)
  if (indexAll > -1) checkedListAll.value.splice(indexAll, 1)
  emit('check-list', checkedListAll.value)
}

watch(seachKey, (value) => {
  if (value && value.trim()) {
    getDishForName(value)
  }
})

onMounted(() => {
  init()
})
</script>
<style lang="scss" scoped>
.addDish {
  display: flex;
  gap: 20px;
  // padding: 20px;
  background-color: #fff;
  border-radius: 8px;
  min-height: 460px;

  /* 左侧分类 */
  .leftCont {
    flex: 3;
    display: flex;
    border-right: 1px solid #ebeef5;
    overflow: hidden;

    .tabBut {
      width: 120px;
      display: flex;
      flex-direction: column;
      align-items: stretch;
      background-color: #f8f9fa;
      border-right: 1px solid #ebeef5;
      padding: 10px 0;
      span {
        padding: 10px 0;
        text-align: center;
        font-weight: 500;
        font-size: 14px;
        color: #606266;
        cursor: pointer;
        transition: all 0.2s ease;
        border-left: 3px solid transparent;

        &:hover {
          background-color: #f0f8ff;
          color: #409eff;
        }
      }
      .act {
        background-color: #ecf5ff;
        color: #409eff;
        border-left-color: #409eff;
      }
    }

    .tabList {
      flex: 1;
      padding: 15px 20px;
      height: 420px;
      overflow-y: auto;
      scrollbar-width: thin;

      .table {
        display: flex;
        flex-direction: column;
        gap: 10px;

        .items {
          background: #fff;
          border: 1px solid #ebeef5;
          border-radius: 8px;
          padding: 10px 12px;
          transition: 0.2s;
          cursor: pointer;

          &:hover {
            background-color: #f9fbff;
            box-shadow: 0 1px 4px rgba(64, 158, 255, 0.15);
          }

          .el-checkbox {
            width: 100%;
          }

          .item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 14px;
            color: #303133;
            span:first-child {
              flex: 2;
              text-align: left;
              font-weight: 500;
            }
            span:nth-child(2) {
              flex: 1;
              color: #67c23a;
            }
            span:nth-child(3) {
              flex: 1;
              text-align: right;
              font-weight: 600;
              color: #f56c6c;
            }
          }
        }

        /* 状态颜色 */
        .items .item span:nth-child(2)[style*='#999'] {
          color: #999 !important;
        }
      }
    }
  }

  /* 右侧已选菜品 */
  .ritCont {
    flex: 2;
    display: flex;
    flex-direction: column;
    padding: 10px 20px;

    .tit {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      border-bottom: 1px solid #ebeef5;
      padding-bottom: 10px;
      margin-bottom: 10px;
    }

    .items {
      flex: 1;
      overflow-y: auto;
      padding-right: 5px;
    }

    .item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: #fafafa;
      border-radius: 8px;
      padding: 10px 12px;
      margin-bottom: 10px;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
      transition: 0.2s ease;
      font-size: 14px;

      &:hover {
        background-color: #f0f9ff;
      }

      span:first-child {
        flex: 2;
        color: #333;
        font-weight: 500;
      }

      .price {
        flex: 1;
        color: #67c23a;
        text-align: right;
        font-weight: 600;
      }

      .del {
        margin-left: 10px;
        cursor: pointer;
        transition: all 0.2s ease;
        border-radius: 50%;
        padding: 4px;

        img {
          width: 18px;
          height: 18px;
        }

        &:hover {
          background-color: #ffeaea;
          transform: scale(1.1);
        }
      }
    }
  }

  /* 滚动条优化 */
  ::-webkit-scrollbar {
    width: 6px;
  }
  ::-webkit-scrollbar-thumb {
    background-color: #dcdfe6;
    border-radius: 4px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background-color: #bcbec4;
  }
}
</style>
