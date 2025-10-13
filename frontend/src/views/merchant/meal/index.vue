<template>
  <div class="dashboard-container">
    <div class="container">
      <div class="tableBar">
        <label style="margin-right: 5px">
          套餐名称: 
        </label>
        <el-input v-model="name" placeholder="请输入套餐名称" style="width: 15%" />

        <label style="margin-right: 5px">
          套餐分类: 
        </label>
        <el-select v-model="categoryId" placeholder="请选择">
          <el-option
            v-for="item in options"
            :key="item.id"
            :label="item.name"
            :value="item.id">
          </el-option>
        </el-select>

        <label style="margin-right: 5px">
          售卖状态: 
        </label>
        <el-select v-model="status" placeholder="请选择">
          <el-option
            v-for="item in statusArr"
            :key="item.value"
            :label="item.label"
            :value="item.value">
          </el-option>
        </el-select>

        <el-button type="primary" style="margin-left: 25px" @click="pageQuery()">查询</el-button>
        <div style="float: right">
         <el-button type="danger" @click="handleDelete('B')">批量删除</el-button>
         <el-button type="info" @click="() => router.push('/merchant/meal/add')">+ 新建套餐</el-button>
        </div>
      </div>
      <el-table :data="records" stripe class="tableBox" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="25" />
        <el-table-column prop="name" label="套餐名称" />
        <el-table-column label="图片">
            <template #default="{ row }">
              <el-image style="width: 80px; height: 40px; border: none" :src="row.image"></el-image>
            </template>
        </el-table-column>
        <el-table-column prop="categoryName" label="套餐分类" />
        <el-table-column prop="price" label="套餐价"/>
        <el-table-column label="售卖状态">
            <template #default="{ row }">
              <div class="tableColumn-status" :class="{ 'stop-use': row.status === 0 }">
                {{ row.status === 0 ? '停售' : '启售' }}
              </div>
            </template>
        </el-table-column>
        <el-table-column prop="updateTime" label="最后操作时间" />
        <el-table-column label="操作" align="center" width="250px">
            <template #default="{ row }">
              <el-button type="text" size="small"> 修改 </el-button>
              <el-button type="text" size="small" @click="handleStartOrStop(row)">
                {{ row.status == '1' ? '停售' : '启售' }}
              </el-button>
              <el-button type="text" size="small" @click="handleDelete('S', row.id)"> 删除 </el-button>
            </template>
        </el-table-column>
      </el-table>
      <el-pagination class="pageList"
                     :page-sizes="[10, 20, 30, 40]"
                     :page-size="pageSize"
                     layout="total, sizes, prev, pager, next, jumper"
                     :total="total"
                     @size-change="handleSizeChange"
                     @current-change="handleCurrentChange" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { getCategoryByType } from '@/api/merchant/category'
import { getSetmealPage, enableOrDisableSetmeal, deleteSetmeal } from '@/api/merchant/setMeal'

const router = useRouter()

const name = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const records = ref<any[]>([])
const options = ref<any[]>([])
const categoryId = ref<any>('')
const statusArr = [
  { value: '0', label: '停售' },
  { value: '1', label: '启售' }
]
const status = ref<any>('')
const multipleSelection = ref<any[]>([])

function pageQuery() {
  const params = {
    page: page.value,
    pageSize: pageSize.value,
    name: name.value,
    status: status.value,
    categoryId: categoryId.value
  }
  getSetmealPage(params).then((res: any) => {
    if (res.data.code === 1) {
      total.value = res.data.data.total
      records.value = res.data.data.records
    }
  })
}

function handleSizeChange(pSize: number) {
  pageSize.value = pSize
  pageQuery()
}

function handleCurrentChange(p: number) {
  page.value = p
  pageQuery()
}

function handleStartOrStop(row: any) {
  const p = {
    id: row.id,
    status: !row.status ? 1 : 0
  }
  ElMessageBox.confirm('确认调整当前套餐的售卖状态, 是否继续?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    enableOrDisableSetmeal(p).then((res: any) => {
      if (res.data.code === 1) {
        ElMessage.success('套餐售卖状态修改成功！')
        pageQuery()
      }
    })
  }).catch(() => {})
}

function handleDelete(type: string, id?: string) {
  ElMessageBox.confirm('确认删除当前指定的套餐, 是否继续?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    let param = ''
    if (type === 'B') {
      const arr: any[] = []
      multipleSelection.value.forEach(element => {
        arr.push(element.id)
      })
      param = arr.join(',')
    } else {
      param = (id as string) || ''
    }
    deleteSetmeal(param).then((res: any) => {
      if (res.data.code === 1) {
        ElMessage.success('删除成功！')
        pageQuery()
      } else {
        ElMessage.error(res.data.msg)
      }
    })
  }).catch(() => {})
}

function handleSelectionChange(val: any) {
  multipleSelection.value = val
}

onMounted(() => {
  getCategoryByType({ type: 2 }).then((res: any) => {
    if (res.data.code === 1) {
      options.value = res.data.data
    }
  })
  pageQuery()
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

    .container {
      background: #fff;
      position: relative;
      z-index: 1;
      padding: 30px 28px;
      border-radius: 4px;

      .tableBar {
        margin-bottom: 20px;
        .tableLab {
          float: right;
          span {
            cursor: pointer;
            display: inline-block;
            font-size: 14px;
            padding: 0 20px;
            color: $gray-2;
          }
        }
      }

      .tableBox {
        width: 100%;
        border: 1px solid $gray-5;
        border-bottom: 0;
      }

      .pageList {
        text-align: center;
        margin-top: 30px;
      }
      //查询黑色按钮样式
      .normal-btn {
        background: #333333;
        color: white;
        margin-left: 20px;
      }
    }
  }
}
</style>
