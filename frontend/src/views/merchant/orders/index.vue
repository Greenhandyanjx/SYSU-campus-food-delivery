<template>
  <div class="dashboard-container">
    <TabChange
      :order-statics="orderStatics"
      :default-activity="defaultActivity"
      @tabChange="change"
    />
  <div class="container main-container" :class="{ hContainer: tableData.length }" >
      <!-- 搜索项 -->
      <div class="tableBar">
        <label style="margin-right: 10px">订单号：</label>
        <el-input
          v-model="input"
          placeholder="请填写订单号"
          style="width: 15%"
          clearable
          @clear="init(orderStatus)"
          @keyup.enter="initFun(orderStatus)"
        />
        <label style="margin-left: 20px">手机号：</label>
        <el-input
          v-model="phone"
          placeholder="请填写手机号"
          style="width: 15%"
          clearable
          @clear="init(orderStatus)"
          @keyup.enter="initFun(orderStatus)"
        />
        <label style="margin-left: 20px">下单时间：</label>
        <el-date-picker
          v-model="valueTime"
          clearable
          value-format="yyyy-MM-dd HH:mm:ss"
          range-separator="至"
          :default-time="['00:00:00', '23:59:59']"
          type="daterange"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          style="width: 25%; margin-left: 10px"
          @clear="init(orderStatus)"
        />
        <el-button class="normal-btn continue" @click="init(orderStatus, true)">
          查询
        </el-button>
      </div>
      <el-table
        v-if="tableData.length"
        :data="tableData"
        stripe
        class="tableBox"
      >
        <el-table-column key="number" prop="number" label="订单号" />
        <el-table-column
          v-if="[2, 3, 4].includes(orderStatus)"
          key="orderDishes"
          prop="orderDishes"
          label="订单菜品"
        />
        <el-table-column
          v-if="[0].includes(orderStatus)"
          key="status"
          prop="订单状态"
          label="订单状态"
        >
          <template #default="{ row }">
            <span>{{ getOrderType(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          v-if="[0, 5, 6].includes(orderStatus)"
          key="consignee"
          prop="consignee"
          label="用户名"
          show-overflow-tooltip
        />
        <el-table-column
          v-if="[0, 5, 6].includes(orderStatus)"
          key="phone"
          prop="phone"
          label="手机号"
        />
        <el-table-column
          v-if="[0, 2, 3, 4, 5, 6].includes(orderStatus)"
          key="address"
          prop="address"
          label="地址"
          :class-name="orderStatus === 6 ? 'address' : ''"
        />
        <el-table-column
          v-if="[0, 6].includes(orderStatus)"
          key="orderTime"
          prop="orderTime"
          label="下单时间"
          class-name="orderTime"
          min-width="110"
        />
        <el-table-column
          v-if="[6].includes(orderStatus)"
          key="cancelTime"
          prop="cancelTime"
          class-name="cancelTime"
          label="取消时间"
          min-width="110"
        />
        <el-table-column
          v-if="[6].includes(orderStatus)"
          key="cancelReason"
          prop="cancelReason"
          label="取消原因"
          class-name="cancelReason"
          :min-width="[6].includes(orderStatus) ? 80 : 'auto'"
        />
        <el-table-column
          v-if="[5].includes(orderStatus)"
          key="deliveryTime"
          prop="deliveryTime"
          label="送达时间"
        />
        <el-table-column
          v-if="[2, 3, 4].includes(orderStatus)"
          key="estimatedDeliveryTime"
          prop="estimatedDeliveryTime"
          label="预计送达时间"
          min-width="110"
        />
        <el-table-column
          v-if="[0, 2, 5].includes(orderStatus)"
          key="amount"
          prop="amount"
          label="实收金额"
          align="center"
        >
          <template #default="{ row }">
            <span>￥{{ (typeof row.amount === 'number' ? row.amount : Number(row.amount || 0)).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          v-if="[2, 3, 4, 5].includes(orderStatus)"
          key="remark"
          prop="remark"
          label="备注"
          align="center"
        />
        <el-table-column
          v-if="[2, 3, 4].includes(orderStatus)"
          key="tablewareNumber"
          prop="tablewareNumber"
          label="餐具数量"
          align="center"
          min-width="80"
        />
        <el-table-column
          prop="btn"
          label="操作"
          align="center"
          :class-name="orderStatus === 0 ? 'operate' : 'otherOperate'"
          :min-width="
            [2, 3, 4].includes(orderStatus)
              ? 130
              : [0].includes(orderStatus)
              ? 140
              : 'auto'
          "
        >
          <template #default="{ row }">
            <!-- <el-divider direction="vertical" /> -->
            <div class="before">
              <el-button
                v-if="row.status === 2"
                type="text"
                class="blueBug"
                @click="orderAcceptHandler(row, true)"
              >
                接单
              </el-button>
              <el-button
                v-if="row.status === 3"
                type="text"
                class="blueBug"
                @click="cancelOrDeliveryOrComplete(3, row.id)"
              >
                派送
              </el-button>
              <el-button
                v-if="row.status === 4"
                type="text"
                class="blueBug"
                @click="cancelOrDeliveryOrComplete(4, row.id)"
              >
                完成
              </el-button>
            </div>
            <div class="middle">
              <el-button
                v-if="row.status === 2"
                type="text"
                class="delBut"
                @click="orderRejectHandler(row, true)"
              >
                拒单
              </el-button>
              <el-button
                v-if="[1, 3, 4, 5].includes(row.status)"
                type="text"
                class="delBut"
                @click="cancelOrderHandler(row)"
              >
                取消
              </el-button>
            </div>
            <div class="after">
              <el-button
                type="text"
                class="blueBug non"
                @click="goDetail(row.id, row.status, row)"
              >
                查看
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <Empty v-else :is-search="isSearch" />
      <el-pagination
        v-if="counts > 10"
        class="pageList"
        :page-sizes="[10, 20, 30, 40]"
        :page-size="pageSize"
        layout="total, sizes, prev, pager, next, jumper"
        :total="counts"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 查看弹框部分 -->
    <el-dialog
      title="订单信息"
      :visible.sync="dialogVisible"
      width="53%"
      :before-close="handleClose"
      class="order-dialog"
    >
      <el-scrollbar style="height: 100%">
        <div class="order-top">
          <div>
            <div style="display: inline-block">
              <label style="font-size: 16px">订单号：</label>
              <div class="order-num">
                {{ diaForm.number }}
              </div>
            </div>
            <div
              style="display: inline-block"
              class="order-status"
              :class="{ status3: [3, 4].includes(dialogOrderStatus) }"
            >
              {{
                orderList.filter((item) => item.value === dialogOrderStatus)[0]
                  .label
              }}
            </div>
          </div>
          <p><label>下单时间：</label>{{ diaForm.orderTime }}</p>
        </div>

        <div class="order-middle">
          <div class="user-info">
            <div class="user-info-box">
              <div class="user-name">
                <label>用户名：</label>
                <span>{{ diaForm.consignee }}</span>
              </div>
              <div class="user-phone">
                <label>手机号：</label>
                <span>{{ diaForm.phone }}</span>
              </div>
              <div
                v-if="[2, 3, 4, 5].includes(dialogOrderStatus)"
                class="user-getTime"
              >
                <label>{{
                  dialogOrderStatus === 5 ? '送达时间：' : '预计送达时间：'
                }}</label>
                <span>{{
                  dialogOrderStatus === 5
                    ? diaForm.deliveryTime
                    : diaForm.estimatedDeliveryTime
                }}</span>
              </div>
              <div class="user-address">
                <label>地址：</label>
                <span>{{ diaForm.address }}</span>
              </div>
            </div>
            <div
              class="user-remark"
              :class="{ orderCancel: dialogOrderStatus === 6 }"
            >
              <div>{{ dialogOrderStatus === 6 ? '取消原因' : '备注' }}</div>
              <span>{{
                dialogOrderStatus === 6
                  ? diaForm.cancelReason || diaForm.rejectionReason
                  : diaForm.remark
              }}</span>
            </div>
          </div>

          <div class="dish-info">
            <div class="dish-label">菜品</div>
            <div class="dish-list">
              <div
                v-for="(item, index) in diaForm.orderDetailList"
                :key="index"
                class="dish-item"
              >
                <div class="dish-item-box">
                  <span class="dish-name">{{ item.name }}</span>
                  <span class="dish-num">x{{ item.number }}</span>
                </div>
                <span class="dish-price"
                  >￥{{ item.amount ? item.amount.toFixed(2) : '' }}</span
                >
              </div>
            </div>
            <div class="dish-all-amount">
              <label>菜品小计</label>
              <span
                >￥{{
                  ((Number(diaForm.amount || 0) - 6 - Number(diaForm.packAmount || 0)).toFixed(2))
                }}</span
              >
            </div>
          </div>
        </div>

        <div class="order-bottom">
          <div class="amount-info">
            <div class="amount-label">费用</div>
            <div class="amount-list">
              <div class="dish-amount">
                <span class="amount-name">菜品小计：</span>
                <span class="amount-price"
                  >￥{{
                    (Number(Number(diaForm.amount || 0) - 6 - Number(diaForm.packAmount || 0)).toFixed(2))
                  }}</span
                >
              </div>
              <div class="send-amount">
                <span class="amount-name">派送费：</span>
                <span class="amount-price">￥{{ 6 }}</span>
              </div>
              <div class="package-amount">
                <span class="amount-name">打包费：</span>
                <span class="amount-price"
                  >￥{{
                    diaForm.packAmount
                      ? Number(diaForm.packAmount).toFixed(2)
                      : ''
                  }}</span
                >
              </div>
              <div class="all-amount">
                <span class="amount-name">合计：</span>
                <span class="amount-price"
                  >￥{{
                    diaForm.amount
                      ? Number(diaForm.amount).toFixed(2)
                      : ''
                  }}</span
                >
              </div>
              <div class="pay-type">
                <span class="pay-name">支付渠道：</span>
                <span class="pay-value">{{
                  diaForm.payMethod === 1 ? '微信支付' : '支付宝支付'
                }}</span>
              </div>
              <div class="pay-time">
                <span class="pay-name">支付时间：</span>
                <span class="pay-value">{{ diaForm.checkoutTime }}</span>
              </div>
            </div>
          </div>
        </div>
      </el-scrollbar>
      <span v-if="dialogOrderStatus !== 6" slot="footer" class="dialog-footer">
        <el-checkbox
          v-if="dialogOrderStatus === 2 && orderStatus === 2"
          v-model="isAutoNext"
          >处理完自动跳转下一条</el-checkbox
        >
        <el-button
          v-if="dialogOrderStatus === 2"
          @click="orderRejectHandler(row, false)"
          >拒 单</el-button
        >
        <el-button
          v-if="dialogOrderStatus === 2"
          type="primary"
          @click="orderAcceptHandler(row, false)"
          >接 单</el-button
        >

        <el-button
          v-if="[1, 3, 4, 5].includes(dialogOrderStatus)"
          @click="dialogVisible = false"
          >返 回</el-button
        >
        <el-button
          v-if="dialogOrderStatus === 3"
          type="primary"
          @click="cancelOrDeliveryOrComplete(3, row.id)"
          >派 送</el-button
        >
        <el-button
          v-if="dialogOrderStatus === 4"
          type="primary"
          @click="cancelOrDeliveryOrComplete(4, row.id)"
          >完 成</el-button
        >
        <el-button
          v-if="[1].includes(dialogOrderStatus)"
          type="primary"
          @click="cancelOrderHandler(row)"
          >取消订单</el-button
        >
      </span>
    </el-dialog>
    <!-- 拒单，取消弹窗 -->
    <el-dialog
      :title="cancelDialogTitle + '原因'"
      :visible.sync="cancelDialogVisible"
      width="42%"
      :before-close="() => ((cancelDialogVisible = false), (cancelReason = ''))"
      class="cancelDialog"
    >
      <el-form label-width="90px">
        <el-form-item :label="cancelDialogTitle + '原因：'">
          <el-select
            v-model="cancelReason"
            :placeholder="'请选择' + cancelDialogTitle + '原因'"
          >
            <el-option
              v-for="(item, index) in cancelDialogTitle === '取消'
                ? cancelrReasonList
                : cancelOrderReasonList"
              :key="index"
              :label="item.label"
              :value="item.label"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="cancelReason === '自定义原因'" label="原因：">
          <el-input
            v-model.trim="remark"
            type="textarea"
            :placeholder="'请填写您' + cancelDialogTitle + '的原因（限20字内）'"
            maxlength="20"
          />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click=";(cancelDialogVisible = false), (cancelReason = '')"
          >取 消</el-button
        >
        <el-button type="primary" @click="confirmCancel">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import HeadLable from '@/components/HeadLable/index.vue'
import InputAutoComplete from '@/components/InputAutoComplete/index.vue'
import TabChange from './tabChange.vue'
import Empty from '@/components/Empty/index.vue'
import {
  getOrderDetailPage,
  queryOrderDetailById,
  completeOrder,
  deliveryOrder,
  orderCancel,
  orderReject,
  orderAccept,
  getOrderListBy,
} from '@/api/merchant/order'

const router = useRouter()
const route = useRoute()

const defaultActivity = ref<number>(0)
const orderStatics = reactive<any>({})
const row = ref<any>({})
const isAutoNext = ref(true)
const isTableOperateBtn = ref(true)
const currentPageIndex = ref(0)
const orderId = ref('')
const input = ref('')
const phone = ref('')
const valueTime = ref<any[]>([])
const dialogVisible = ref(false)
const cancelDialogVisible = ref(false)
const cancelDialogTitle = ref('')
const cancelReason = ref('')
const remark = ref('')
const counts = ref(0)
const page = ref(1)
const pageSize = ref(10)
const tableData = ref<any[]>([])
const diaForm = ref<any>({})
const isSearch = ref(false)
const orderStatus = ref(0)
const dialogOrderStatus = ref(0)

const cancelOrderReasonList = [
  { value: 1, label: '订单量较多，暂时无法接单' },
  { value: 2, label: '菜品已销售完，暂时无法接单' },
  { value: 3, label: '餐厅已打烊，暂时无法接单' },
  { value: 0, label: '自定义原因' },
]

const cancelrReasonList = [
  { value: 1, label: '订单量较多，暂时无法接单' },
  { value: 2, label: '菜品已销售完，暂时无法接单' },
  { value: 3, label: '骑手不足无法配送' },
  { value: 4, label: '客户电话取消' },
  { value: 0, label: '自定义原因' },
]

const orderList = [
  { label: '全部订单', value: 0 },
  { label: '待付款', value: 1 },
  { label: '待接单', value: 2 },
  { label: '待派送', value: 3 },
  { label: '派送中', value: 4 },
  { label: '已完成', value: 5 },
  { label: '已取消', value: 6 },
]

onMounted(() => {
  const status = Number(route.query.status) || 0
  defaultActivity.value = status
  init(status)

  if (route.query.orderId && route.query.orderId !== 'undefined') {
    goDetail(route.query.orderId as string, 2)
  }
})

function initFun(st: number) {
  page.value = 1
  init(st)
}

function change(activeIndex: number) {
  if (activeIndex === orderStatus.value) return
  init(activeIndex)
  input.value = ''
  phone.value = ''
  valueTime.value = []
  dialogOrderStatus.value = 0
  router.push('/merchant/orders')
}

async function getOrderListBy3Status() {
  try {
    const res = await getOrderListBy({})
    if (res.data.code === 1) {
      Object.assign(orderStatics, res.data.data)
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

function getOrderType(row: any) {
  switch (row.status) {
    case 1:
      return '待付款'
    case 2:
      return '待接单'
    case 3:
      return '待派送'
    case 4:
      return '派送中'
    case 5:
      return '已完成'
    case 6:
      return '已取消'
    default:
      return '退款'
  }
}

async function init(activeIndex = 0, isSearchFlag?: boolean) {
  isSearch.value = !!isSearchFlag
  const params: any = {
    page: page.value,
    pageSize: pageSize.value,
    number: input.value || undefined,
    phone: phone.value || undefined,
    beginTime: valueTime.value && valueTime.value.length > 0 ? valueTime.value[0] : undefined,
    endTime: valueTime.value && valueTime.value.length > 0 ? valueTime.value[1] : undefined,
    status: activeIndex || undefined,
  }
  try {
    const res = await getOrderDetailPage({ ...params })
    if (res.data.code === 1) {
      const data = res.data.data || {}
      const raw = data.records || []
      // 格式化时间字段，防止前端出现 NaN 或 undefined
      tableData.value = raw.map((it: any) => {
        const safeFormat = (v: any) => {
          if (!v && v !== 0) return ''
          try {
            const d = new Date(v)
            if (isNaN(d.getTime())) return String(v)
            return d.toLocaleString()
          } catch (e) {
            return String(v)
          }
        }
        return {
          ...it,
          orderTime: safeFormat(it.orderTime),
          cancelTime: safeFormat(it.cancelTime),
          deliveryTime: safeFormat(it.deliveryTime),
          estimatedDeliveryTime: safeFormat(it.estimatedDeliveryTime),
          checkoutTime: safeFormat(it.checkoutTime),
          // 保证 amount 是数字或空字符串，避免模板中 toFixed 报错
          amount: typeof it.amount === 'number' ? it.amount : it.amount ? Number(it.amount) : 0,
        }
      })
      orderStatus.value = activeIndex
      counts.value = Number(data.total || data.totalCount || tableData.value.length || 0)
      getOrderListBy3Status()
      if (
        dialogOrderStatus.value === 2 &&
        orderStatus.value === 2 &&
        isAutoNext.value &&
        !isTableOperateBtn.value &&
        tableData.value.length > 1
      ) {
        const r = tableData.value[0]
        goDetail(r.id, r.status, r)
      }
    } else {
      ElMessage.error(res?.data?.msg || '获取订单列表失败')
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

async function goDetail(id: any, status: number, r?: any) {
  diaForm.value = {}
  dialogVisible.value = true
  dialogOrderStatus.value = status
  orderId.value = id
  try {
    const { data } = await queryOrderDetailById({ orderId: id })
    diaForm.value = data.data
    row.value = r || { id: route.query.orderId, status }
  if (route.query.orderId) router.push('/merchant/orders')
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

function orderRejectHandler(r: any, setTableFlag = true) {
  cancelDialogVisible.value = true
  orderId.value = r.id
  dialogOrderStatus.value = r.status
  cancelDialogTitle.value = '拒绝'
  dialogVisible.value = false
  cancelReason.value = ''
  isTableOperateBtn.value = setTableFlag
}

async function orderAcceptHandler(r: any, setTableFlag = true) {
  orderId.value = r.id
  dialogOrderStatus.value = r.status
  isTableOperateBtn.value = setTableFlag
  try {
    const res = await orderAccept({ id: orderId.value })
    if (res.data.code === 1) {
      ElMessage.success('操作成功')
      orderId.value = ''
      dialogVisible.value = false
      init(orderStatus.value)
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

function cancelOrderHandler(r: any) {
  cancelDialogVisible.value = true
  orderId.value = r.id
  dialogOrderStatus.value = r.status
  cancelDialogTitle.value = '取消'
  dialogVisible.value = false
  cancelReason.value = ''
}

async function confirmCancel() {
  if (!cancelReason.value) return ElMessage.error(`请选择${cancelDialogTitle.value}原因`)
  if (cancelReason.value === '自定义原因' && !remark.value) return ElMessage.error(`请输入${cancelDialogTitle.value}原因`)
  const fn = cancelDialogTitle.value === '取消' ? orderCancel : orderReject
  const payload: any = {
    id: orderId.value,
  }
  payload[cancelDialogTitle.value === '取消' ? 'cancelReason' : 'rejectionReason'] =
    cancelReason.value === '自定义原因' ? remark.value : cancelReason.value
  try {
    const res = await fn(payload)
    if (res.data.code === 1) {
      ElMessage.success('操作成功')
      cancelDialogVisible.value = false
      orderId.value = ''
      init(orderStatus.value)
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

async function cancelOrDeliveryOrComplete(status: number, id: string) {
  const params = { status, id }
  try {
    const res = await (status === 3 ? deliveryOrder(params) : completeOrder(params))
    if (res.data.code === 1) {
      ElMessage.success('操作成功')
      orderId.value = ''
      dialogVisible.value = false
      init(orderStatus.value)
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

function handleClose() {
  dialogVisible.value = false
}

function handleSizeChange(val: any) {
  pageSize.value = val
  init(orderStatus.value)
}

function handleCurrentChange(val: any) {
  page.value = val
  init(orderStatus.value)
}
</script>

<style lang="scss" scoped>
.dashboard {
  &-container {
    margin: 30px;
    // height: 100%;
    min-height: 700px;
    .container, .main-container {
      background: #fff;
      position: relative;
      z-index: 1;
      max-width: 1200px;
      width: 100%;
      margin: 0 auto;
      padding: 28px 32px;
      border-radius: 10px;
      box-shadow: 0 8px 30px rgba(20,24,31,0.06);

      .tableBar {
        margin-bottom: 18px;
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: 12px;
      }

      .tableLab {
        span {
          cursor: pointer;
          display: inline-block;
          font-size: 14px;
          padding: 0 20px;
          color: white;
          border-right: solid 1px $gray-4;
        }
      }

      /* Ensure inputs have consistent height so placeholder text isn't clipped */
      .tableBar ::v-deep .el-input__inner,
      .tableBar ::v-deep .el-select .el-input__inner,
      .tableBar ::v-deep .el-date-editor--daterange .el-input__inner {
        height: 40px;
        line-height: 40px;
        padding: 8px 12px;
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
      .normal-btn {
        background: #333333;
        color: white;
        margin-left: 20px;
        padding: 8px 14px;
        border-radius: 6px;
      }
    }
    .hContainer {
      height: auto !important;
    }
  }
}

.search-btn {
  margin-left: 20px;
}

.info-box {
  margin: -15px -44px 20px;
  p {
    display: flex;
    height: 20px;
    line-height: 20px;
    font-size: 14px;
    font-weight: 400;
    color: #666666;
    text-align: left;
    margin-bottom: 14px;
    &:last-child {
      margin-bottom: 0;
    }
    label {
      width: 100px;
      display: inline-block;
      color: #666;
    }
    .des {
      flex: 1;
      color: #333333;
    }
  }
}

.order-top {
  // height: 80px;
  border-bottom: 1px solid #e7e6e6;
  padding-bottom: 26px;
  padding-left: 22px;
  padding-right: 22px;
  // margin: 0 30px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  .order-status {
    width: 57.25px;
    height: 27px;
    background: #333333;
    border-radius: 13.5px;
    color: white;
    margin-left: 19px;
    text-align: center;
    line-height: 27px;
  }
  .status3 {
    background: #f56c6c;
  }
  p {
    color: #333;
    label {
      color: #666;
    }
  }
  .order-num {
    font-size: 16px;
    color: #2a2929;
    font-weight: bold;
    display: inline-block;
  }
}

.order-middle {
  .user-info {
    min-height: 140px;
    background: #fbfbfa;
    margin-top: 23px;

    padding: 20px 43px;
    color: #333;
    .user-info-box {
      min-height: 55px;
      display: flex;
      flex-wrap: wrap;
      .user-name {
        flex: 67%;
      }
      .user-phone {
        flex: 33%;
      }
      .user-getTime {
        margin-top: 14px;
        flex: 80%;
        label {
          margin-right: 3px;
        }
      }
      label {
        margin-right: 17px;
        color: #666;
      }

      .user-address {
        margin-top: 14px;
        flex: 80%;
        label {
          margin-right: 30px;
        }
      }
    }
    .user-remark {
      min-height: 43px;
      line-height: 43px;
      background: #fffbf0;
      border: 1px solid #fbe396;
      border-radius: 4px;
      margin-top: 10px;
      padding: 6px;
      display: flex;
      align-items: center;
      div {
        display: inline-block;
        min-width: 53px;
        height: 32px;
        background: #fbe396;
        border-radius: 4px;
        text-align: center;
        line-height: 32px;
        color: #333;
        margin-right: 30px;
        // padding: 12px 6px;
      }
      span {
        color: #f2a402;
        line-height: 1.15;
      }
    }
    .orderCancel {
      background: #ffffff;
      border: 1px solid #b6b6b6;

      div {
        padding: 0 10px;
        background-color: #e5e4e4;
      }
      span {
        color: #f56c6c;
      }
    }
  }
  .dish-info {
    // min-height: 180px;
    display: flex;
    flex-wrap: wrap;
    padding: 20px 40px;
    border-bottom: 1px solid #e7e6e6;
    .dish-label {
      color: #666;
    }
    .dish-list {
      flex: 80%;
      display: flex;
      flex-wrap: wrap;
      .dish-item {
        flex: 50%;
        margin-bottom: 14px;
        color: #333;
  /* .dish-num rule removed (was empty) */
        .dish-item-box {
          display: inline-block;
          width: 120px;
        }
      }
    }
    .dish-label {
      margin-right: 65px;
    }
    .dish-all-amount {
      flex: 1;
      padding-left: 92px;
      margin-top: 10px;
      label {
        color: #333333;
        font-weight: bold;
        margin-right: 5px;
      }
      span {
        color: #f56c6c;
      }
    }
  }
}
.order-bottom {
  .amount-info {
    // min-height: 180px;
    display: flex;
    flex-wrap: wrap;
    padding: 20px 40px;
    padding-bottom: 0px;
    .amount-label {
      color: #666;
      margin-right: 65px;
    }
    .amount-list {
      flex: 80%;
      display: flex;
      flex-wrap: wrap;
      color: #333;
      // height: 65px;
      .dish-amount,
      .package-amount,
      .pay-type {
        display: inline-block;
        width: 300px;
        margin-bottom: 14px;
        flex: 50%;
      }
      .send-amount,
      .all-amount,
      .pay-time {
        display: inline-block;
        flex: 50%;
        padding-left: 10%;
      }
      .package-amount {
        .amount-name {
          margin-right: 14px;
        }
      }
      .all-amount {
        .amount-name {
          margin-right: 24px;
        }
        .amount-price {
          color: #f56c6c;
        }
      }
      .send-amount {
        .amount-name {
          margin-right: 10px;
        }
      }
    }
  }
}
</style>

<style lang="scss">
.dashboard-container {
  .cancelReason {
    padding-left: 40px;
  }
  .cancelTime {
    padding-left: 50px;
  }
  .orderTime {
    padding-left: 50px;
  }
  td.operate .cell {
    .before,
    .middle,
    .after {
      height: 39px;
      width: 48px;
    }
  }
  td.operate .cell,
  td.otherOperate .cell {
    display: flex;
    flex-wrap: nowrap;
    justify-content: center;
  }
  .order-dialog {
    .el-dialog {
      max-height: 764px !important;
      display: flex;
      flex-direction: column;
      margin: 0 !important;
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      max-height: calc(100% - 30px);
      max-width: calc(100% - 30px);
    }
    .el-dialog__body {
      height: 520px !important;
    }
  }
}
.el-dialog__body {
  padding-top: 34px;
  padding-left: 30px;
  padding-right: 30px;
}
.cancelDialog {
  .el-dialog__body {
    padding-left: 64px;
  }
  .el-select,
  .el-textarea {
    width: 293px;
  }
  .el-textarea textarea {
    height: 114px;
  }
}
.el-dialog__footer {
  .el-checkbox {
    float: left;
    margin-left: 40px;
  }
  .el-checkbox__label {
    color: #333333 !important;
  }
}
.empty-box {
  display: flex;
  align-items: center;
  justify-content: center;
  img {
    margin-top: 0 !important;
  }
}
</style>
