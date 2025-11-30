<template>
  <div :key="vueRest"
       class="addBrand-container">
    <div :key="restKey"
         class="container">
    <el-form ref="ruleFormRef"
               :model="ruleForm"
               :rules="rules"
               :inline="true"
               label-width="180px"
               class="demo-ruleForm">
        <div>
          <el-form-item label="菜品名称:"
                        prop="name">
            <el-input v-model="ruleForm.name"
                      placeholder="请填写菜品名称"
                      maxlength="20" />
          </el-form-item>
          <el-form-item label="菜品分类:" prop="categoryIds">
            <!-- 支持多分类选择 -->
            <el-select v-model="ruleForm.categoryIds"
                       placeholder="请选择菜品分类"
                       multiple
                       clearable>
              <el-option v-for="item in dishList"
                         :key="item.id"
                         :label="item.name"
                         :value="item.id" />
            </el-select>
          </el-form-item>
        </div>
        <div>
          <el-form-item label="菜品价格:"
                        prop="price">
            <el-input v-model="ruleForm.price"
                      placeholder="请设置菜品价格" />
          </el-form-item>
        </div>
        <!-- 标签（可选）：多标签、可输入新标签 -->
        <el-form-item label="标签:" prop="tags">
          <el-select v-model="ruleForm.tags" multiple filterable allow-create placeholder="输入或选择标签">
            <el-option v-for="t in dishTagList" :key="t" :label="t" :value="t" />
          </el-select>
        </el-form-item>
        <el-form-item label="口味做法配置:">
          <el-form-item>
            <div class="flavorBox">
              <span v-if="dishFlavors.length == 0"
                    class="addBut"
                    @click="addFlavore">
                + 添加口味</span>
              <div v-if="dishFlavors.length != 0"
                   class="flavor">
                <div class="title">
                  <span>口味名（3个字内）</span>
                  <!-- <span class="des-box">口味标签（输入标签回车添加）</span> -->
                </div>
                <div class="cont">
                  <div v-for="(item, index) in dishFlavors"
                       :key="index"
                       class="items">
                    <div class="itTit">
                      <!-- :dish-flavors-data="filterDishFlavorsData()" -->
<SelectInput
  :dishFlavorsData="leftDishFlavors"
  :index="index"
  @select="selectHandle"
/>

                    </div>
                    <div class="labItems"
                         style="display: flex">
                      <span v-for="(it, ind) in item.value"
                            :key="ind">{{ it }}
                        <i @click="delFlavorLabel(index, ind)">X</i></span>
                      <div class="inputBox"
                           :style="inputStyle" />
                    </div>
                    <span class="delFlavor delBut non"
                          @click="delFlavor(item.name)">删除</span>
                  </div>
                </div>
                <div v-if="
                       !!leftDishFlavors.length &&
                         dishFlavors.length < dishFlavorsData.length
                     "
                     class="addBut"
                     @click="addFlavore">
                  添加口味
                </div>
              </div>
            </div>
          </el-form-item>
        </el-form-item>
        <div>
          <el-form-item label="菜品图片:"
                        prop="image">
            <image-upload :prop-image-url="imageUrl"
                          :headers="uploadHeaders"
                          @imageChange="imageChange">
              图片大小不超过2M<br>仅能上传 PNG JPEG JPG类型图片<br>建议上传200*200或300*300尺寸的图片
            </image-upload>
          </el-form-item>
        </div>
        <div class="address">
          <el-form-item label="菜品描述:"
                        prop="region">
            <el-input v-model="ruleForm.description"
                      type="textarea"
                      :rows="3"
                      maxlength="200"
                      placeholder="菜品描述，最长200字" />
          </el-form-item>
        </div>
        <div class="subBox address">
          <el-button @click="() => $router.back()">
            取消
          </el-button>
          <el-button type="primary"
                     :class="{ continue: actionType === 'add' }"
                     @click="submitForm('ruleForm')">
            保存
          </el-button>
          <el-button v-if="actionType == 'add'"
                     type="primary"
                     @click="submitForm('ruleForm', 'goAnd')">
            保存并继续添加
          </el-button>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import SelectInput from './components/SelectInput.vue'
import ImageUpload from '@/components/ImgUpLoad/index.vue'
import {
  queryDishById,
  addDish,
  editDish,
  getCategoryList,
  commonDownload
} from '@/api/merchant/dish'
import { CATEGORIES } from '@/constants/categories'
// token handled by request interceptor; getToken import removed
const uploadHeaders = {
  Authorization: `Bearer ${localStorage.getItem('token')}`
}
const route = useRoute()
const router = useRouter()
const mak = ref(false)
const restKey = ref(0)
const textarea = ref('')
const value = ref('')
const imageUrl = ref('')
const actionType = ref('')
const dishList = ref<any[]>(CATEGORIES.filter(c => c.id !== 0))
const dishFlavorsData = ref<any[]>([])
const dishFlavors = ref<any[]>([])
// 本地标签池（示例），可改为从后端获取
const dishTagList = ref<any[]>(['推荐', '新品', '促销'])
// const leftDishFlavors = ref<any[]>([])
const leftDishFlavors = ref<any[]>([
  { name: '甜味', value: ['无糖', '少糖', '半糖', '多糖', '全糖'] },
  { name: '温度', value: ['热饮', '常温', '去冰', '少冰', '多冰'] },
  { name: '忌口', value: ['不要葱', '不要蒜', '不要香菜', '不要辣'] },
  { name: '辣度', value: ['不辣', '微辣', '中辣', '重辣'] }
])
const vueRest = ref('1')
const index = ref(0)
const inputStyle = { flex: 1 }
// headers removed
const ruleFormRef = ref<any>(null)

// removed shared selectedFlavors binding; each SelectInput manages its own selections

// function selectHandle(type, index, val) {
//   console.log('选中口味:', type, val)
// }

const ruleForm = reactive<any>({
  name: '',
  id: '',
  price: '',
  code: '',
  image: '',
  description: '',
  dishFlavors: [],
  status: true,
  // 新数据结构：支持多分类与标签
  categoryIds: [] as number[],
  tags: [] as string[]
})

const rules = reactive({
  name: [
    {
      required: true,
      validator: (rule: any, value: string, callback: Function) => {
        if (!value) {
          callback(new Error('请输入菜品名称'))
        } else {
          const reg = /^([A-Za-z0-9\u4e00-\u9fa5]){2,20}$/
          if (!reg.test(value)) {
            callback(new Error('菜品名称输入不符，请输入2-20个字符'))
          } else {
            callback()
          }
        }
      },
      trigger: 'blur'
    }
  ],
  categoryId: [{ required: true, message: '请选择菜品分类', trigger: 'change' }],
  image: { required: true, message: '菜品图片不能为空' },
  price: [
    {
      required: true,
      validator: (rules: any, value: string, callback: Function) => {
        const reg = /^([1-9]\d{0,5}|0)(\.\d{1,2})?$/
        if (!reg.test(value) || Number(value) <= 0) {
          callback(new Error('菜品价格格式有误，请输入大于零且最多保留两位小数的金额'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  code: [{ required: true, message: '请填写商品码', trigger: 'blur' }]
})

onMounted(() => {
  getDishList()
  getFlavorListHand()
  actionType.value = route.query.id ? 'edit' : 'add'
  if (route.query.id) init()
})

watch(dishFlavors, () => {
  getLeftDishFlavors()
})

function getLeftDishFlavors() {
  const arr: any[] = []
  dishFlavorsData.value.forEach(item => {
    if (dishFlavors.value.findIndex(item1 => item.name === item1.name) === -1) {
      arr.push(item)
    }
  })
  leftDishFlavors.value = arr
}

function selectHandle(type: string, idx: number, val: string) {
  // type: flavor group name, idx: index of the current flavor item, val: option toggled
  if (typeof idx !== 'number' || idx < 0) return
  const arr = [...dishFlavors.value]
  if (!arr[idx]) arr[idx] = { name: '', value: [] }
  // set the flavor group name
  arr[idx].name = type
  if (!Array.isArray(arr[idx].value)) arr[idx].value = []
  const pos = arr[idx].value.indexOf(val)
  if (pos === -1) arr[idx].value.push(val)
  else arr[idx].value.splice(pos, 1)
  dishFlavors.value = arr
}

async function init() {
  try {
    const res = await queryDishById(route.query.id)
    if (res && res.data && Number(res.data.code) === 1) {
        const data = res.data.data
        // 兼容后端老字段或新字段（categoryId 或 categories）
        Object.assign(ruleForm, data)
        ruleForm.price = String(data.price)
        ruleForm.status = data.status == '1'
        // categories 可能为数组或单值 id
        if (Array.isArray(data.categories)) ruleForm.categoryIds = data.categories
        else if (data.categoryId) ruleForm.categoryIds = [data.categoryId]
        // tags 可能为数组或逗号分隔字符串
        if (Array.isArray(data.tags)) ruleForm.tags = data.tags
        else if (typeof data.tags === 'string' && data.tags) ruleForm.tags = data.tags.split(',')
        dishFlavors.value = (data.flavors || []).map((obj: any) => ({ ...obj, value: JSON.parse(obj.value) }))
        getLeftDishFlavors()
        imageUrl.value = data.image
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

function addFlavore() {
  dishFlavors.value.push({ name: '', value: [] })
}

function delFlavor(name: string) {
  const ind = dishFlavors.value.findIndex(item => item.name === name)
  if (ind > -1) dishFlavors.value.splice(ind, 1)
}

function delFlavorLabel(index: number, ind: number) {
  dishFlavors.value[index].value.splice(ind, 1)
}

function flavorPosition(idx: number) {
  index.value = idx
}

function keyDownHandle(val: any, e?: Event) {
  if (e) {
    e.cancelBubble = true
    e.preventDefault()
    e.stopPropagation()
  }
  if (val.target.innerText.trim() !== '') {
    dishFlavors.value[index.value].flavorData.push(val.target.innerText)
    val.target.innerText = ''
  }
}

// categories are now static constants (CATEGORIES); keep helper for backward compatibility
async function getDishList() {
  // try backend first, fallback to local constants
  try {
    const res = await getCategoryList({ type: 1 })
    if (res && res.data && Number(res.data.code) === 1 && Array.isArray(res.data.data) && res.data.data.length > 0) {
      dishList.value = res.data.data
      return
    }
  } catch (err) {
    // ignore and fallback
  }
  dishList.value = CATEGORIES.filter(c => c.id !== 0)
}

function getFlavorListHand() {
  dishFlavorsData.value = [
    { name: '甜味', value: ['无糖', '少糖', '半糖', '多糖', '全糖'] },
    { name: '温度', value: ['热饮', '常温', '去冰', '少冰', '多冰'] },
    { name: '忌口', value: ['不要葱', '不要蒜', '不要香菜', '不要辣'] },
    { name: '辣度', value: ['不辣', '微辣', '中辣', '重辣'] }
  ]
  // console.log('口味数据:', dishFlavorsData.value)
}

function imageChange(value: any) {
  ruleForm.image = value
}

async function submitForm(formRefName: string, st?: any) {
  const formRef = ruleFormRef.value
  if (formRef) {
    formRef.validate(async (valid: any) => {
        if (valid) {
        if (!ruleForm.image) return ElMessage.error('菜品图片不能为空')
        // 构造符合新后端的数据结构：dish + categories + tags
        const params: any = {
          ...ruleForm,
          status: actionType.value === 'add' ? 0 : ruleForm.status ? 1 : 0,
          // categories 提交为数组（后端会把它映射到中间表）
          categories: ruleForm.categoryIds || [],
          // 保持向后兼容：提交第一个 categoryId 到旧接口字段
          categoryId: (ruleForm.categoryIds && ruleForm.categoryIds[0]) || undefined,
          // tags 可以作为字符串数组提交
          tags: ruleForm.tags || [],
          flavors: dishFlavors.value.map((obj: any) => ({ ...obj, value: JSON.stringify(obj.value) }))
        }
        delete params.dishFlavors
        try {
          if (actionType.value === 'add') {
            delete params.id
            const res = await addDish(params)
            if (res.data.code === 1) {
              ElMessage.success('菜品添加成功！')
              if (!st) router.push({ path: '/menu' })
              else {
                dishFlavors.value = []
                imageUrl.value = ''
                Object.assign(ruleForm, { name: '', id: '', price: '', code: '', image: '', description: '', dishFlavors: [], status: true, categoryId: '' })
                restKey.value++
              }
            } else {
              ElMessage.error(res.data.desc || res.data.msg)
            }
          } else {
            delete params.createTime
            delete params.updateTime
            const res = await editDish(params)
            if (res && res.data && Number(res.data.code) === 1) {
              router.push({ path: '/menu' })
              ElMessage.success('菜品修改成功！')
            } else {
              ElMessage.error(res.data.desc || res.data.msg)
            }
          }
        } catch (err: any) {
          ElMessage.error('请求出错了：' + err.message)
        }
      } else {
        return false
      }
    })
  }
}
</script>
<style lang="scss" scoped>
.addBrand-container {
  .el-form--inline .el-form-item__content {
    width: 293px;
  }

  .el-input {
    width: 350px;
  }

  .address {
    .el-form-item__content {
      width: 777px !important;
    }
  }
}
</style>
<style lang="scss" scoped>
.addBrand {
  &-container {
    margin: 30px;

    .container {
      position: relative;
      z-index: 1;
      background: #fff;
      padding: 30px;
      border-radius: 4px;
      min-height: 500px;

      .subBox {
        padding-top: 30px;
        text-align: center;
        border-top: solid 1px $gray-5;
      }
      .upload-item {
        .el-form-item__error {
          top: 90%;
        }
      }
    }
  }
}

.flavorBox {
  width: 777px;

  .addBut {
    background: $blue;
    display: inline-block;
    padding: 0px 20px;
    border-radius: 3px;
    line-height: 40px;
    cursor: pointer;
    border-radius: 4px;
    color:white;
    font-weight: 500;
  }

  .flavor {
    border: solid 1px #dfe2e8;
    border-radius: 3px;
    padding: 15px;
    background: #fafafb;

    .title {
      color: #606168;
      .des-box {
        padding-left: 44px;
      }
    }

    .cont {
      .items {
        display: flex;
        margin: 10px 0;

        .itTit {
          width: 150px;
          margin-right: 15px;

          input {
            width: 100%;
            // line-height: 40px;
            // border-radius: 3px;
            // padding: 0 10px;
          }
        }

        .labItems {
          flex: 1;
          display: flex;
          flex-wrap: wrap;
          border-radius: 3px;
          min-height: 39px;
          border: solid 1px #d8dde3;
          background: #fff;
          padding: 0 5px;

          span {
            display: inline-block;
            color: #ffc200;
            margin: 5px;
            line-height: 26px;
            padding: 0 10px;
            background: #fffbf0;
            border: 1px solid #fbe396;
            border-radius: 4px;
            font-size: 12px;

            i {
              cursor: pointer;
              font-style: normal;
            }
          }

          .inputBox {
            display: inline-block;
            width: 100%;
            height: 36px;
            line-height: 36px;
            overflow: hidden;
          }
        }

        .delFlavor {
          display: inline-block;
          padding: 0 10px;
          color: #f19c59;
          cursor: pointer;
        }
      }
    }
  }
}
</style>
