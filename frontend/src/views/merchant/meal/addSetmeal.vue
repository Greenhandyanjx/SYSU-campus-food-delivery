<template>
  <div class="addBrand-container">
    <div class="container">
  <el-form ref="ruleFormRef"
               :model="ruleForm"
               :rules="rules"
               :inline="true"
               label-width="180px"
               class="demo-ruleForm">
        <div>
          <el-form-item label="套餐名称:"
                        prop="name">
            <el-input v-model="ruleForm.name"
                      placeholder="请填写套餐名称"
                      maxlength="14" />
          </el-form-item>
          <el-form-item label="套餐分类:" prop="categoryId">
            <el-select style="min-width: 200px;" v-model="ruleForm.categoryId" placeholder="请选择套餐分类">
              <el-option
                v-for="(item, index) in setMealList"
                :key="index"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
        </div>
        <div>
          <el-form-item label="套餐价格:"
                        prop="price">
            <el-input v-model="ruleForm.price"
                      placeholder="请设置套餐价格" />
          </el-form-item>
        </div>
        <div>
          <el-form-item label="套餐菜品:" required>
            <div class="addDish">
              <el-button type="primary" @click="openAddDish">+ 添加菜品</el-button>
            
              <el-table
                v-if="dishTable.length > 0"
                :data="dishTable"
                style="width: 100%; margin-top: 15px;"
              >
                <el-table-column prop="name" label="名称" width="180" align="center" />
                <el-table-column prop="price" label="原价" width="180" align="center">
                  <template #default="{ row }">
                    {{ Number(row.price).toFixed(2) }}
                  </template>
                </el-table-column>
                <el-table-column label="份数" align="center">
                  <template #default="{ row }">
                    <el-input-number v-model.number="row.copies" size="small" :min="1" :max="99" controls-position="right"/>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="120" align="center">
                  <template #default="{ $index }">
                    <el-button type="danger" size="small" @click="delDishHandle($index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </el-form-item>
        </div>
        <div>
          <el-form-item label="套餐图片:"
                        required
                        prop="image">
            <image-upload :prop-image-url="imageUrl"
                          @imageChange="imageChange">
              图片大小不超过2M<br>仅能上传 PNG JPEG JPG类型图片<br>建议上传200*200或300*300尺寸的图片
            </image-upload>
          </el-form-item>
        </div>
        <div class="address">
          <el-form-item label="套餐描述:">
            <el-input v-model="ruleForm.description"
                      type="textarea"
                      :rows="3"
                      maxlength="200"
                      placeholder="套餐描述，最长200字" />
          </el-form-item>
        </div>
        <div class="subBox address">
          <el-form-item>
            <el-button @click="() => $router.back()">
              取消
            </el-button>
            <el-button type="primary"
                       :class="{ continue: actionType === 'add' }"
                       @click="submitForm('ruleForm', false)">
              保存
            </el-button>
            <el-button v-if="actionType == 'add'"
                       type="primary"
                       @click="submitForm('ruleForm', true)">
              保存并继续添加
            </el-button>
          </el-form-item>
        </div>
      </el-form>
    </div>
  <el-dialog
         title="添加菜品"
         class="addDishList"
         v-model="dialogVisible"
         width="60%"
         :before-close="handleClose">
    <AddDish
         ref="adddish"
         :check-list="checkList"
         :seach-key="seachKey"
         @check-list="getCheckList" />
      <template #footer class="dialog-footer">
        <el-button @click="handleClose">取 消</el-button>
        <el-button type="primary"
                   @click="addTableList">添 加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import HeadLable from '@/components/HeadLable/index.vue'
import ImageUpload from '@/components/ImgUpLoad/index.vue'
import AddDish from './components/AddDish.vue'
import { querySetmealById, addSetmeal, editSetmeal } from '@/api/merchant/setMeal'
import { getCategoryList } from '@/api/dish'
import { CATEGORIES } from '@/constants/categories'

const route = useRoute()
const router = useRouter()

const setMealList = ref<any[]>(CATEGORIES.filter(c => c.id !== 0))
const seachKey = ref('')
const dishList = ref<any[]>([])
const imageUrl = ref('')
const actionType = ref('')
const dishTable = ref<any[]>([])
const dialogVisible = ref(false)
const checkList = ref<any[]>([])
const ruleFormRef = ref<any>(null)

const ruleForm = reactive<any>({
  name: '',
  categoryId: '',
  price: '',
  code: '',
  image: '/images/fresh_salad_set.jpg',
  description: '',
  dishList: [],
  status: true,
})

const rules = {
  name: {
    required: true,
    validator: (rule: any, value: string, callback: Function) => {
      if (!value) {
        callback(new Error('请输入套餐名称'))
      } else {
        const reg = /^([A-Za-z0-9\u4e00-\u9fa5]){2,20}$/
        if (!reg.test(value)) {
          callback(new Error('套餐名称输入不符，请输入2-20个字符'))
        } else {
          callback()
        }
      }
    },
    trigger: 'blur'
  },
  categoryId: { required: true, message: '请选择套餐分类', trigger: 'change' },
  image: { required: false, message: '套餐图片不能为空' },
  price: {
    required: true,
    validator: (rules: any, value: string, callback: Function) => {
      const reg = /^([1-9]\d{0,5}|0)(\.\d{1,2})?$/
      if (!reg.test(value) || Number(value) <= 0) {
        callback(new Error('套餐价格格式有误，请输入大于零且最多保留两位小数的金额'))
      } else {
        callback()
      }
    },
    trigger: 'blur'
  },
  code: { required: true, message: '请输入商品码', trigger: 'blur' }
}

// ✅ 修复版 init()
function init() {
  if (route.query.id) {
    querySetmealById(route.query.id).then((res: any) => {
      if (res && res.data && Number(res.data.code) === 1) {
        // 同步基础表单
        Object.assign(ruleForm, res.data.data)
        ruleForm.status = res.data.data.status == '1'
        ruleForm.price = res.data.data.price
        imageUrl.value = res.data.data.image

        // 初始化菜品列表
        const dishes = res.data.data.setmealDishes || []

        // ✅ 确保 copies 是数字且响应式
        checkList.value = dishes.map((dish: any) => ({
          ...dish,
          copies: Number(dish.copies) || 1
        }))

        // dishTable 同步（保持顺序一致）
        dishTable.value = [...checkList.value]

        // 分类回显
        ruleForm.categoryId = res.data.data.categoryId
      } else {
        ElMessage.error(res.data.msg || '加载套餐信息失败')
      }
    }).catch((err: any) => {
      ElMessage.error('加载出错：' + err.message)
    })
  }
}


function getDishTypeList() {
  // Prefer backend categories, fallback to local constants
  getCategoryList({ type: 2 }).then((res: any) => {
    if (res && res.data && Number(res.data.code) === 1 && Array.isArray(res.data.data) && res.data.data.length > 0) {
      setMealList.value = res.data.data
    } else {
      setMealList.value = CATEGORIES.filter(c => c.id !== 0)
    }
  }).catch(() => {
    setMealList.value = CATEGORIES.filter(c => c.id !== 0)
  })
}

function delDishHandle(index: number) {
  dishTable.value.splice(index, 1)
  checkList.value = dishTable.value
}

function getCheckList(value: any) {
  checkList.value = [...value].reverse()
}

function openAddDish(st: string) {
  dialogVisible.value = true
}

function handleClose() {
  dialogVisible.value = false
  checkList.value = JSON.parse(JSON.stringify(dishTable.value))
}

// ✅ 修复版 addTableList()
function addTableList() {
  // 使用 map 保持响应性，避免 JSON 深拷贝
  dishTable.value = checkList.value.map((item: any) => ({
    ...item,
    // 确保 copies 为 number 类型
    copies: Number(item.copies) || 1
  }))
  dialogVisible.value = false
}


function submitForm(formName: string, st: any) {
  ;(ruleFormRef.value as any)?.validate((valid: any) => {
    if (valid) {
      if (dishTable.value.length === 0) {
        return ElMessage.error('套餐下菜品不能为空')
      }
      if (!ruleForm.image) return ElMessage.error('套餐图片不能为空')
      const prams: any = { ...ruleForm }
      prams.setmealDishes = dishTable.value.map((obj: any) => ({
        copies: obj.copies,
        dishId: obj.dishId,
        name: obj.name,
        price: obj.price
      }))
      prams.status = actionType.value === 'add' ? 0 : ruleForm.status ? 1 : 0
      prams.categoryId = ruleForm.categoryId
      if (actionType.value === 'add') {
        delete prams.id
        addSetmeal(prams)
          .then((res: any) => {
            if (res && res.data && Number(res.data.code) === 1) {
              ElMessage.success('套餐添加成功！')
              if (!st) {
                router.push({ path: '/setmeal' })
              } else {
                ;(ruleFormRef.value as any).resetFields()
                dishList.value = []
                dishTable.value = []
                Object.assign(ruleForm, {
                  name: '',
                  categoryId: '',
                  price: '',
                  code: '',
                  image: '/images/fresh_salad_set.jpg',
                  description: '',
                  dishList: [],
                  status: true,
                  id: '',
                })
                imageUrl.value = '/images/fresh_salad_set.jpg'
              }
            } else {
              ElMessage.error(res.data.msg)
            }
          })
          .catch((err: any) => ElMessage.error('请求出错了：' + err.message))
      } else {
        delete prams.updateTime
        editSetmeal(prams)
          .then((res: any) => {
            if (res.data.code === 1) {
              ElMessage.success('套餐修改成功！')
              router.push({ path: '/setmeal' })
            }
          })
          .catch((err: any) => ElMessage.error('请求出错了：' + err.message))
      }
    } else {
      return false
    }
  })
}

function imageChange(value: any) {
  ruleForm.image = value
}

onMounted(() => {
  getDishTypeList()
  actionType.value = route.query.id ? 'edit' : 'add'
  if (actionType.value === 'edit') {
    init()
  }
})
</script>
<style lang="scss" scoped>
.addBrand-container {
  margin: 24px;

  .container {
    background: #fff;
    padding: 32px;
    border-radius: 8px;
    min-height: 500px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);

    .page-title {
      margin-bottom: 24px;
      font-size: 20px;
      font-weight: 500;
      color: #1f2d3d;
    }

    /* ========== 表单基础样式 ========== */
    .el-form-item {
      margin-bottom: 24px;

      .el-form-item__label {
        font-size: 14px;
        color: #333;
      }

      .el-input {
        width: 350px;
        .el-input__inner {
          height: 40px;
          line-height: 40px;
          padding: 0 15px;
          font-size: 14px;
          &::placeholder {
            color: #909399;
            font-size: 13px;
          }
        }
      }

      textarea.el-textarea__inner {
        font-size: 14px;
        padding: 10px;
      }
    }

    .address .el-form-item__content {
      width: 780px !important;
    }

    /* ========== 图片上传区域 ========== */
    .avatar-uploader {
      .el-upload {
        border: 1px dashed #d9d9d9;
        border-radius: 6px;
        cursor: pointer;
        position: relative;
        overflow: hidden;
        transition: border-color 0.3s;
      }

      .el-upload:hover {
        border-color: #ffc200;
      }

      .avatar-uploader-icon {
        font-size: 28px;
        color: #8c939d;
        width: 200px;
        height: 160px;
        line-height: 160px;
        text-align: center;
      }

      .avatar {
        width: 200px;
        height: 160px;
        display: block;
        border-radius: 4px;
        object-fit: cover;
      }

      .el-icon-plus:after {
        content: '';
        position: absolute;
        left: calc(50% - 20px);
        top: calc(50% - 40px);
        width: 40px;
        height: 40px;
        background: url('./../../assets/icons/icon_upload@2x.png') center center no-repeat;
        background-size: 20px;
      }
    }

    /* ========== 添加菜品区域 ========== */
    .addDish {
      width: 100%;
      max-width: 900px;
      // margin: 0 auto;

      .addBut {
        background: #409eff;
        color: #fff;
        border-radius: 4px;
        padding: 0 20px;
        height: 38px;
        line-height: 38px;
        display: inline-flex;
        align-items: center;
        cursor: pointer;
        font-weight: 500;
        transition: 0.2s;

        &:hover {
          background: #66b1ff;
          transform: translateY(-1px);
        }

        i {
          margin-right: 6px;
        }
      }

      .content {
        background: #fafafb;
        padding: 20px;
        border: 1px solid #ebeef5;
        border-radius: 6px;
        margin-top: 16px;

        .el-table {
          width: 100%;
          border-radius: 4px;
          overflow: hidden;

          th {
            background-color: #f5f7fa;
            font-weight: 500;
            padding: 12px 0;
            text-align: center;
          }

          td {
            padding: 12px 0;
            text-align: center;
          }

          .el-button {
            color: #f56c6c;
            &:hover {
              color: #ff7875;
            }
          }
        }
      }

      .el-input-number__increase,
      .el-input-number__decrease {
        background: #fff8e1;
        border-color: #ffe082;
      }
    }

    /* ========== 底部按钮区 ========== */
    .subBox {
      padding-top: 32px;
      text-align: right;
      border-top: 1px solid #ebeef5;

      .el-button {
        padding: 10px 24px;
        font-size: 14px;
        & + .el-button {
          margin-left: 12px;
        }
      }

      .el-button.continue {
        background-color: #67c23a;
        border-color: #67c23a;
        &:hover {
          background-color: #85ce61;
        }
      }
    }
  }
}

/* ========== 弹窗样式 ========== */
.addDishList {
  .el-dialog__header {
    font-size: 16px;
    font-weight: 500;
  }

  .el-dialog__body {
    padding: 0;
    border-bottom: 1px solid #efefef;
  }

  .el-dialog__footer {
    padding-top: 20px;
  }

  .seachDish {
    position: absolute;
    top: 12px;
    right: 20px;

    .el-input__inner {
      height: 40px;
      line-height: 40px;
    }
  }
}
</style>

