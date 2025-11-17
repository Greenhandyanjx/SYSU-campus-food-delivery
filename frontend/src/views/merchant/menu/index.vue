
<template>
  <div class="dashboard-container">
  <div class="container main-container">
      <div class="tableBar">
        <label style="margin-right: 10px ;width: 100px;">菜品名称：</label>
  <el-input v-model="input"
      placeholder="请填写菜品名称"
      style="width: 30%"
      clearable
      @clear="init"
      @keyup.enter="initFun" />

        <label style="margin-right: 10px; margin-left: 20px">菜品分类：</label>
        <el-select v-model="categoryId"
                   style="width: 30%"
                   placeholder="请选择"
                   clearable
                   @clear="init">
        <el-option v-for="item in dishCategoryOptions"
                     :key="item.value"
                     :label="item.label"
                     :value="item.value" />
        </el-select>

        <label style="margin-right: 10px; margin-left: 20px">售卖状态：</label>
        <el-select v-model="dishStatus"
                   style="width: 20%"
                   placeholder="请选择"
                   clearable
                   @clear="init">
          <el-option v-for="item in saleStatus"
                     :key="item.value"
                     :label="item.label"
                     :value="item.value" />
        </el-select>
        <el-button class="normal-btn continue"
                   @click="init(true)">
          查询
        </el-button>

      </div>
      <div style="float: right;">
        <!-- <span class="delBut non"
              @click="deleteHandle('批量', null)">批量删除</span>
        <span class="blueBug non" @click="statusHandle('1')">批量启售</span>
        <span
          style="border: none"
          class="delBut non"
          @click="statusHandle('0')"
          >批量停售</span
        > -->
        <el-button type="primary"
                   style="margin-left: 15px;color: white;"
                   @click="deleteHandle('批量', null)">
          批量删除
        </el-button>
        <el-button type="primary"
                   style="margin-left: 15px;color: white;"
                   @click="statusHandle('1')">
          批量启售
        </el-button>
        <el-button type="primary"
                   style="margin-left: 15px;color: white;"
                   @click="statusHandle('0')">
         批量停售
        </el-button>
        <el-button type="primary"
                   style="margin-left: 15px;color: white;"
                   @click="addDishtype('add')">
          + 新建菜品
        </el-button>
      </div>
  <el-table v-if="tableData.length"
    :data="tableData"
    :key="$route.fullPath"
    stripe
    class="tableBox"
    @selection-change="handleSelectionChange">
        <el-table-column type="selection"
                         width="25" />
        <el-table-column prop="name"
                         label="菜品名称" />
        <el-table-column prop="imageUrl"
                         label="图片">
          <template #default="{ row }">
            <el-image style="width: 80px; height: 40px; border: none; cursor: pointer"
                      :src="row.imageUrl">
              <template #error>
                <div class="image-slot">
                  <img src="/src/assets/noImg.png"
                       style="width: auto; height: 40px; border: none">
                </div>
              </template>
            </el-image>
          </template>
        </el-table-column>
        <el-table-column prop="categoryName"
                         label="菜品分类" />
        <el-table-column label="售价">
          <template #default="scope">
            <span style="margin-right: 10px">￥{{ Number(scope.row.price||0 ).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="售卖状态">
          <template #default="scope">
            <div class="tableColumn-status"
                 :class="{ 'stop-use': String(scope.row.status) === '0' }">
              {{ String(scope.row.status) === '0' ? '停售' : '启售' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="updateTime"
                         label="最后操作时间" />
        <el-table-column label="操作"
                         width="250"
                         align="center">
          <template #default="scope">
            <el-button type="text"
                       size="small"
                       class="blueBug"
                       @click="addDishtype(scope.row.id)">
              修改
            </el-button>
            <el-button type="text"
                       size="small"
                       class="delBut"
                       @click="deleteHandle('单删', scope.row.id)">
              删除
            </el-button>
            <el-button type="text"
                       size="small"
                       class="non"
                       :class="{
                         blueBug: scope.row.status == '0',
                         delBut: scope.row.status != '0'
                       }"
                       @click="statusHandle(scope.row)">
              {{ scope.row.status == '0' ? '启售' : '停售' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <Empty v-else
             :is-search="isSearch"
              />
      <el-pagination v-if="counts > 10"
                     class="pageList"
                     :page-sizes="[10, 20, 30, 40]"
                     :page-size="pageSize"
                     layout="total, sizes, prev, pager, next, jumper"
                     :total="counts"
                     @size-change="handleSizeChange"
                     @current-change="handleCurrentChange" />
    </div>
  </div>
</template>


<script setup lang="ts">
import { ref, onMounted ,onBeforeUnmount} from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import HeadLable from '@/components/HeadLable/index.vue'
import InputAutoComplete from '@/components/InputAutoComplete/index.vue'
import Empty from '@/components/Empty/index.vue'
import {
  getDishPage,
  editDish,
  deleteDish,
  dishStatusByStatus,
  dishCategoryList as fetchDishCategoryList
} from '@/api/merchant/dish'

const router = useRouter()
const input = ref('')
const counts = ref(0)
const page = ref(1)
const pageSize = ref(10)
const checkList = ref<string[]>([])
const tableData = ref<any[]>([])
const dishState = ref<any>('')
const dishCategoryOptions = ref<any[]>([])
const categoryId = ref('')
const dishStatus = ref('')
const imageUrl = ref('')
const isSearch = ref(false)
const isUnmounted = ref(false)
const saleStatus = ref([
  { value: 0, label: '停售' },
  { value: 1, label: '启售' }
])

async function init(isSearchFlag?: boolean) {
  isSearch.value = !!isSearchFlag
  try {
    const res = await getDishPage({
      page: page.value,
      pageSize: pageSize.value,
      name: input.value || undefined,
      categoryId: categoryId.value || undefined,
      imageUrl: imageUrl.value || undefined,
      status: dishStatus.value
    })

    // 组件已卸载则直接返回（防止卸载后更新导致 DOM 访问错误）
    if (isUnmounted.value) return

    // 兼容后端返回 "1" 或 1
    if (res && res.data && Number(res.data.code) === 1) {
      const d = res.data.data || {}
      const rawList = d.records || d.items || d.list || []

      // 为避免单条数据字段缺失导致渲染抛错，统一填充默认 demo 值
      function createDefaultRow(r: any) {
        const nowStr = new Date().toLocaleString()
        return {
          id: r?.id ?? `demo-${Math.random().toString(36).slice(2, 8)}`, // <- 使用反引号或字符串包裹
          name: r?.name || '示例菜名',
          imageUrl: r?.imageUrl || '/src/assets/noImg.png',
          categoryName: r?.categoryName || '默认分类',
          price: typeof r?.price === 'number' ? r.price : (r?.price ? Number(r.price) : 0),
          status: r?.status ?? 1,
          updateTime: r?.updateTime || nowStr,
          // 保留原有字段，防止丢失其他自定义属性
          ...r
        }
      }

      tableData.value = rawList.map((it: any) => createDefaultRow(it))
      counts.value = Number(d.total || d.totalCount || rawList.length || 0)
    } else {
      // 后端返回码不为 1，按需处理（比如显示提示，或清空列表）
      tableData.value = []
      counts.value = 0
      // 可视情况：ElMessage.warning(res?.data?.msg || '请求异常')
    }
  } catch (err: any) {
    if (!isUnmounted.value) {
      ElMessage.error('请求出错了：' + (err?.message || err))
    }
  }
}

function initFun() {
  page.value = 1
  init()
}

function initProp(val: any) {
  input.value = val
  initFun()
}

function addDishtype(st: any) {
  if (st === 'add') router.push({ path: '/merchant/menu/add' })
  else router.push({ path: '/merchant/menu/add', query: { id: st } })
}

function deleteHandle(type: string, id: any) {
  if (type === '批量' && id === null) {
    if (checkList.value.length === 0) {
      return ElMessage.error('请选择删除对象')
    }
  }
  ElMessageBox.confirm('确认删除该菜品, 是否继续?', '确定删除', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      try {
        const res = await deleteDish(type === '批量' ? checkList.value.join(',') : id)
        if (res && res.data && Number(res.data.code) === 1) {
          ElMessage.success('删除成功！')
          init()
        } else {
          ElMessage.error(res.data.msg)
        }
      } catch (err: any) {
        ElMessage.error('请求出错了：' + err.message)
      }
    })
    .catch(() => {})
}

async function getDishCategoryList() {
  try {
    const res = await fetchDishCategoryList({ type: 1 })
    if (res && res.data && Number(res.data.code) === 1) {
      dishCategoryOptions.value = (res.data?.data || []).map((item: any) => ({ value: item.id, label: item.name }))
    }
  } catch (err) {
    // ignore
  }
}

function statusHandle(row: any) {
  const params: any = {}
  if (typeof row === 'string') {
    if (checkList.value.length === 0) {
      ElMessage.error('批量操作，请先勾选操作菜品！')
      return false
    }
    params.id = checkList.value.join(',')
    params.status = row
  } else {
    params.id = row.id
    params.status = row.status ? '0' : '1'
  }
  dishState.value = params
  ElMessageBox.confirm('确认更改该菜品状态?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      try {
        const res = await dishStatusByStatus(dishState.value)
        if (res && res.data && Number(res.data.code) === 1) {
          ElMessage.success('菜品状态已经更改成功！')
          init()
        } else {
          ElMessage.error(res.data.msg)
        }
      } catch (err: any) {
        ElMessage.error('请求出错了：' + err.message)
      }
    })
    .catch(() => {})
}

function handleSelectionChange(val: any) {
  checkList.value = val.map((n: any) => n.id)
}

function handleSizeChange(val: any) {
  pageSize.value = val
  init()
}

function handleCurrentChange(val: any) {
  page.value = val
  init()
}

onMounted(() => {
  init()
  getDishCategoryList()
})
onBeforeUnmount(() => {
  isUnmounted.value = true
})

</script>
<style lang="scss">
.el-table-column--selection .cell {
  padding-left: 10px;
}
</style>
<style lang="scss" scoped>
.dashboard {
  &-container {
    margin: 30px;
    .container, .main-container {
      background: #fff;
      position: relative;
      z-index: 1;
      max-width: 1200px;
      width: 100%;
      margin: 0 auto;
      padding: 26px 32px;
      border-radius: 10px;
      box-shadow: 0 8px 30px rgba(20,24,31,0.06);

      // 查询按钮样式
      .normal-btn {
        background: #333333;
        color: white;
        margin-left: 20px;
        padding: 8px 16px;
        border-radius: 6px;
      }

      .tableBar {
        margin-bottom: 18px;
        display: flex;
        align-items: center;
        justify-content: space-between;

        .left-controls {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .tableLab {
          display: flex;
          align-items: center;
          gap: 12px;
          span {
            cursor: pointer;
            display: inline-block;
            font-size: 14px;
            padding: 6px 12px;
            color: #666;
            border-radius: 6px;
            transition: background 0.2s;
          }
        }
      }

      /* 修复 input placeholder 显示被裁切的问题，以及统一输入高度 */
      .el-input,
      .el-select {
        margin-right: 8px;
      }

      .el-input .el-input__inner,
      .el-select .el-input__inner {
        height: 38px;
        line-height: 38px;
        padding: 0 12px;
        box-sizing: border-box;
        border-radius: 6px;
      }

      .tableBox {
        width: 100%;
        border: 1px solid $gray-5;
        border-bottom: 0;
        border-radius: 8px;
        overflow: hidden;
      }

      .pageList {
        text-align: center;
        margin-top: 30px;
      }
    }
  }
}
</style>
