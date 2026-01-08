<template>
  <div>
    <div class="container homecon">
      <h2 class="homeTitle homeTitleBtn">
        <img src="/JDlogo.png" class="jd-logo" alt="嘉递外卖" />
        订单信息
        <ul class="conTab">
          <li
            v-for="(item, index) in tabList"
            :key="index"
            :class="activeIndex === index ? 'active' : ''"
            @click="handleClass(index)"
          >
            <el-badge
              class="item"
              :class="item.num >= 10 ? 'badgeW' : ''"
              :value="(Number.isFinite(Number(item.num)) ? item.num : 0) > 99 ? '99+' : (Number.isFinite(Number(item.num)) ? item.num : 0)"
              :hidden="!([2, 3].includes(item.value) && (Number.isFinite(Number(item.num)) ? item.num : 0))"
              >{{ item.label }}</el-badge
            >
          </li>
        </ul>
      </h2>
      <div class="">
        <div v-if="orderData.length > 0">
          <el-table
            :data="orderData"
            stripe
            class="tableBox"
            style="width: 100%"
            @row-click="handleTable"
          >
            <el-table-column prop="orderid" label="订单号"> </el-table-column>
            <el-table-column label="订单菜品">
              <template v-slot="scope">
                <div class="ellipsisHidden">
                  <el-popover
                    placement="top-start"
                    title=""
                    width="200"
                    trigger="hover"
                    :content="scope.row.orderDishes"
                  >
                    <template v-slot:reference><span>{{ scope.row.orderDishes }}</span></template>
                  </el-popover>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              label="地址"
              :class-name="dialogOrderStatus === 2 ? 'address' : ''"
            >
              <template v-slot="scope">
                <div class="ellipsisHidden">
                  <el-popover
                    placement="top-start"
                    title=""
                    width="200"
                    trigger="hover"
                    :content="scope.row.address"
                  >
                    <template v-slot:reference><span>{{ scope.row.address }}</span></template>
                  </el-popover>
                </div>
              </template>
            </el-table-column>

            <el-table-column
              prop="expected_time"
              label="预计送达时间"
              sortable
              class-name="orderTime"
              min-width="130"
            >
              <template v-slot="{ row }">
                <span>{{ formatDateToCN(row.expected_time) }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="totalprice" label="实收金额"> </el-table-column>
            <el-table-column label="备注">
              <template v-slot="scope">
                <div class="ellipsisHidden">
                  <el-popover
                    placement="top-start"
                    title=""
                    width="200"
                    trigger="hover"
                    :content="scope.row.remark"
                  >
                    <template v-slot:reference><span>{{ scope.row.remark }}</span></template>
                  </el-popover>
                </div>
              </template>
            </el-table-column>
            <el-table-column
              prop="quantity"
              label="餐具数量"
              min-width="80"
              align="center"
              v-if="status === 3"
            >
            </el-table-column>
            <el-table-column
              label="操作"
              align="center"
              :class-name="dialogOrderStatus === 0 ? 'operate' : 'otherOperate'"
              :min-width="
                [2, 3].includes(dialogOrderStatus)
                  ? 130
                  : [0].includes(dialogOrderStatus)
                  ? 140
                  : 'auto'
              "
            >
              <template v-slot="{ row }">
                <!-- <el-divider direction="vertical" /> -->
                <div class="before">
                  <el-button
                    v-if="row.status === 2"
                    type="text"
                    class="blueBug"
                    @click="orderAcceptFn(row, $event), (isTableOperateBtn = true)"
                  >
                    接单
                  </el-button>
                  <el-button
                    v-if="row.status === 3"
                    type="text"
                    class="blueBug"
                    @click="cancelOrDeliveryOrComplete(3, row.id, $event)"
                  >
                    派送
                  </el-button>
                </div>
                <div class="middle">
                  <el-button
                    v-if="row.status === 2"
                    type="text"
                    class="delBut"
                    @click="orderRejectFn(row, $event), (isTableOperateBtn = true)"
                  >
                    拒单
                  </el-button>
                  <el-button
                    v-if="[1, 3, 4, 5].includes(row.status)"
                    type="text"
                    class="delBut"
                    @click="cancelOrderFn(row, $event)"
                  >
                    取消
                  </el-button>
                </div>
                <div class="after">
                  <el-button
                    type="text"
                    class="blueBug non"
                    @click="goDetail(row.id || row.orderId || row.orderid, row.status, row, $event)"
                  >
                    查看
                  </el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
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
    </div>
    <!-- 查看弹框（原生模态替代 el-dialog） -->
    <div v-if="dialogVisible" class="native-modal-overlay" @click.self="dialogVisible = false">
      <div class="native-modal" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div>
            <label style="font-size:16px">订单号：</label>
            <span class="order-num">{{ diaForm.number || diaForm.orderNo || diaForm.orderid || diaForm.id }}</span>
          </div>
          <div class="modal-close" @click="dialogVisible = false">✕</div>
        </div>

        <div class="modal-body">
          <div class="order-top">
            <p><label>下单时间：</label>{{ formatDateToCN(diaForm.orderTime) }}</p>
          </div>

          <div class="order-middle">
            <div class="user-info">
              <div class="user-info-box">
                <div class="user-name"><label>用户名：</label><span>{{ diaForm.consignee }}</span></div>
                <div class="user-phone"><label>手机号：</label><span>{{ diaForm.phone }}</span></div>
                <div v-if="[2,3,4,5].includes(dialogOrderStatus)" class="user-getTime">
                  <label>{{ dialogOrderStatus === 5 ? '送达时间：' : '预计送达时间：' }}</label>
                  <span>{{ dialogOrderStatus === 5 ? formatDateToCN(diaForm.deliveryTime) : formatDateToCN(diaForm.estimatedDeliveryTime) }}</span>
                </div>
                <div class="user-address"><label>地址：</label><span>{{ diaForm.address }}</span></div>
              </div>

              <div class="user-remark" :class="{ orderCancel: dialogOrderStatus === 6 }">
                <div>{{ dialogOrderStatus === 6 ? '取消原因' : '备注' }}</div>
                <span>{{ dialogOrderStatus === 6 ? (diaForm.cancelReason || diaForm.rejectionReason) : diaForm.remark }}</span>
              </div>
            </div>

            <div class="dish-info">
              <div class="dish-label">菜品</div>
              <div class="dish-list">
                <div v-for="(item, index) in diaForm.orderDetailList || []" :key="index" class="dish-item">
                  <span class="dish-name">{{ item.name }}</span>
                  <span class="dish-num">x{{ item.number || item.qty || item.quantity }}</span>
                  <span class="dish-price">￥{{item.price ? (Number(item.price).toFixed(2)) : '' }}</span>
                </div>
              </div>
              <div class="dish-all-amount"><label>菜品小计</label><span>￥{{ dishSubtotal }}</span></div>
            </div>
          </div>

          <div class="order-bottom">
            <div class="amount-info">
              <div class="amount-label">费用</div>
              <div class="amount-list">
                <div class="dish-amount"><span class="amount-name">菜品小计：</span><span class="amount-price">￥{{ dishSubtotalRaw }}</span></div>
                <div class="send-amount"><span class="amount-name">派送费：</span><span class="amount-price">￥{{ deliveryAmountDisplay }}</span></div>
                <div class="package-amount"><span class="amount-name">打包费：</span><span class="amount-price">￥{{ packAmountDisplay }}</span></div>
                <div class="all-amount"><span class="amount-name">合计：</span><span class="amount-price">￥{{ amountDisplay }}</span></div>
                <div class="pay-type"><span class="pay-name">支付渠道：</span><span class="pay-value">{{ diaForm.payMethod === 1 ? '微信支付' : '支付宝支付' }}</span></div>
                <div class="pay-time"><span class="pay-name">支付时间：</span><span class="pay-value">{{ formatDateToCN(diaForm.checkoutTime) }}</span></div>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer" v-if="dialogOrderStatus !== 6">
          <label v-if="dialogOrderStatus === 2 && status === 2" class="auto-next"><input type="checkbox" v-model="isAutoNext" /> 处理完自动跳转下一条</label>
          <button v-if="dialogOrderStatus === 2" class="btn" @click="orderRejectFn(row, $event), (isTableOperateBtn = false)">拒 单</button>
          <button v-if="dialogOrderStatus === 2" class="btn primary" @click="orderAcceptFn(row, $event), (isTableOperateBtn = false)">接 单</button>
          <button v-if="[1,3,4,5].includes(dialogOrderStatus)" class="btn" @click="dialogVisible = false">返 回</button>
          <button v-if="dialogOrderStatus === 3" class="btn primary" @click="cancelOrDeliveryOrComplete(3, row.id, $event)">派 送</button>
          <button v-if="dialogOrderStatus === 4" class="btn primary" @click="cancelOrDeliveryOrComplete(4, row.id, $event)">完 成</button>
          <button v-if="[1].includes(dialogOrderStatus)" class="btn primary" @click="cancelOrderFn(row, $event)">取消订单</button>
        </div>
      </div>
    </div>
    <!-- end -->
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
    <!-- end -->
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import Empty from '@/components/Empty/index.vue'
import { getMerchantProfile } from '@/api/merchant/profile'
import {
  getOrderDetailPage,
  queryOrderDetailById,
  queryOrderDetailByIdCoalesced,
  completeOrder,
  deliveryOrder,
  orderCancel,
  orderReject,
  orderAccept,
  getOrderListBy,
} from '@/api/merchant/order'
import { getOrderDetailPageCoalesced } from '@/api/merchant/order'
import { emitOrderChanged } from '@/utils/orderEvents'

const props = defineProps<{ orderStatics?: any }>()
const emit = defineEmits(['getOrderListBy3Status'])


const { proxy } = getCurrentInstance() as any

const orderId = ref('') //订单号
const dialogOrderStatus = ref(0) //弹窗所需订单状态，用于详情展示字段
const activeIndex = ref(0)

const dialogVisible = ref(false) //详情弹窗
const cancelDialogVisible = ref(false) //取消，拒单弹窗
const cancelDialogTitle = ref('') //取消，拒绝弹窗标题
const cancelReason = ref('')
const remark = ref('') //自定义原因
const diaForm = ref<any>({})
const row = ref<any>({})
const isAutoNext = ref(true)
const isSearch = ref(false)
const counts = ref(0)
const page = ref(1)
const pageSize = ref(10)
const status = ref(2)
const orderData = ref<any[]>([])
const currentMerchantId = ref<any>(null)
const isTableOperateBtn = ref(true)

const cancelOrderReasonList = ref([
  {
    value: 1,
    label: '订单量较多，暂时无法接单',
  },
  {
    value: 2,
    label: '菜品已销售完，暂时无法接单',
  },
  {
    value: 3,
    label: '餐厅已打烊，暂时无法接单',
  },
  {
    value: 0,
    label: '自定义原因',
  },
])

const cancelrReasonList = ref([
  {
    value: 1,
    label: '订单量较多，暂时无法接单',
  },
  {
    value: 2,
    label: '菜品已销售完，暂时无法接单',
  },
  {
    value: 3,
    label: '骑手不足无法配送',
  },
  {
    value: 4,
    label: '客户电话取消',
  },
  {
    value: 0,
    label: '自定义原因',
  },
])

const orderList = ref([
  {
    label: '全部订单',
    value: 0,
  },
  {
    label: '待付款',
    value: 1,
  },
  {
    label: '待接单',
    value: 2,
  },
  {
    label: '待派送',
    value: 3,
  },
  {
    label: '派送中',
    value: 4,
  },
  {
    label: '已完成',
    value: 5,
  },
  {
    label: '已取消',
    value: 6,
  },
])

const dishSubtotal = computed(() => {
  const a = diaForm.value?.amount
  const pack = Number(diaForm.value?.packAmount || 0)
  const delivery = Number(diaForm.value?.deliveryAmount || diaForm.value?.delivery_amount || diaForm.value?.deliveryFee || diaForm.value?.delivery_fee || 0)
  if (typeof a === 'number') {
    return (a - delivery - pack).toFixed(2)
  }
  return ''
})

const dishSubtotalRaw = computed(() => {
  const a = diaForm.value?.amount
  const pack = Number(diaForm.value?.packAmount || 0)
  const delivery = Number(diaForm.value?.deliveryAmount || diaForm.value?.delivery_amount || diaForm.value?.deliveryFee || diaForm.value?.delivery_fee || 0)
  if (typeof a === 'number') {
    return ((a - delivery - pack) * 100) / 100
  }
  return ''
})

const deliveryAmountDisplay = computed(() => {
  const v = diaForm.value?.deliveryAmount || diaForm.value?.delivery_amount || diaForm.value?.deliveryFee || diaForm.value?.delivery_fee || 0
  const n = Number(v || 0)
  return Number.isFinite(n) ? n.toFixed(2) : '0.00'
})

// 将时间格式化为：YYYY年MM月DD日 HH:mm
function formatDateToCN(s: any) {
  if (!s) return ''
  const dt = new Date(s)
  if (isNaN(dt.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  const yyyy = dt.getFullYear()
  const mm = pad(dt.getMonth() + 1)
  const dd = pad(dt.getDate())
  const HH = pad(dt.getHours())
  const MM = pad(dt.getMinutes())
  return `${yyyy}年${mm}月${dd}日 ${HH}:${MM}`
}

const packAmountDisplay = computed(() => {
  const p = diaForm.value?.packAmount
  return typeof p === 'number' ? ((p.toFixed(2) as unknown) as string) : ''
})

const amountDisplay = computed(() => {
  const a = diaForm.value?.amount
  return typeof a === 'number' ? ((a.toFixed(2) as unknown) as string) : ''
})

const tabList = computed(() => {
  return [
    {
      label: '待接单',
      value: 2,
      num: props.orderStatics?.toBeConfirmed,
    },
    {
      label: '待派送',
      value: 3,
      num: props.orderStatics?.confirmed,
    },
  ]
})


onMounted(() => {
  getOrderListData(status.value)
  // 监听来自其他组件（如聊天）的打开订单详情事件
  window.addEventListener('merchant:open_order_detail', (ev: any) => {
    try {
      const id = ev && ev.detail && (ev.detail.orderId || ev.detail.id)
      if (!id) return
      // 打开 modal 并加载详情
      goDetail(id, 2, {})
    } catch (e) { console.warn('open_order_detail handler failed', e) }
  })

  // 监听 order:changed 事件，去抖刷新列表，避免频繁刷新导致卡顿
  let refreshTimer: any = null
  const refreshHandler = (ev: any) => {
    try {
      clearTimeout(refreshTimer)
      refreshTimer = setTimeout(() => {
        getOrderListData(status.value)
      }, 500)
    } catch (e) { console.warn('order:changed handler failed', e) }
  }
  window.addEventListener('order:changed', refreshHandler)

  // 清理监听
  ;(window as any).____orderList_refreshHandler = refreshHandler
})

onBeforeUnmount(() => {
  const h = (window as any).____orderList_refreshHandler
  if (h) window.removeEventListener('order:changed', h)
})

// 获取当前商家 id（用于前端二次防御过滤）
;(async () => {
  try {
    const r: any = await getMerchantProfile()
    if (r && r.data && Number(r.data.code) === 1 && r.data.data) {
      currentMerchantId.value = r.data.data.id || r.data.data.ID || r.data.data.merchant_id || r.data.data.merchantId || null
    }
  } catch (e) {
    // 忽略：后端应已在服务端进行过滤
  }
})()

// 获取订单数据
async function getOrderListData(s: number) {
  const params = {
    page: page.value,
    pageSize: pageSize.value,
    status: s,
  }
  try {
    const data = await getOrderDetailPageCoalesced(params)
    // 保持原始响应结构的使用方式（并做一次字段规范化，确保每条都有 id/orderId/orderid）
    const rawItems = data?.data?.data?.items || []
    orderData.value = rawItems.map((it: any) => {
      const idVal = Number(it.id ?? it.ID ?? it.orderId ?? it.orderID ?? it.orderid ?? it.order_id ?? 0)
      const orderIdVal = it.orderId ?? it.orderid ?? it.orderID ?? it.id ?? idVal
      const deliveryAmount = it.deliveryAmount ?? it.delivery_amount ?? it.deliveryFee ?? it.delivery_fee ?? it.deliveryFee
      return Object.assign({}, it, {
        id: idVal,
        orderId: orderIdVal,
        orderid: orderIdVal,
        amount: (it.amount ?? it.totalPrice ?? it.total_price ?? 0),
        deliveryAmount: deliveryAmount,
      })
    })
    console.log("表格数据 =", orderData.value)
    counts.value = data?.data?.data?.total || 0
    emit('getOrderListBy3Status')

const recordsForAuto = data?.data?.data?.items || []
console.log('recordsForAuto', recordsForAuto)
    if (
      dialogOrderStatus.value === 2 &&
      status.value === 2 &&
      isAutoNext.value &&
      !isTableOperateBtn.value &&
      recordsForAuto.length >= 1
    ) {
      console.log('自动跳转下一条')
      const r = recordsForAuto[0]
      goDetail(r.orderid, r.status, r, r)
    }
  } catch (err: any) {
    proxy.$message && proxy.$message.error('请求出错了：' + err?.message)
  }
}

// 接单
function orderAcceptFn(rowItem: any, event?: Event) {
  event && event.stopPropagation()
  orderId.value = rowItem.id
  dialogOrderStatus.value = rowItem.status
  orderAccept({ id: orderId.value })
    .then((res: any) => {
        if (Number(res.data.code) === 1) {
        proxy.$message.success('操作成功')
        const emittedId = orderId.value
        orderId.value = ''
        dialogVisible.value = false
        // 广播变化，使用统一去抖 helper 刷新列表（避免重复请求）
        try { emitOrderChanged({ orderId: emittedId }) } catch (e) {}
      } else {
        proxy.$message.error(res.data.msg)
      }
    })
    .catch((err: any) => {
      proxy.$message.error('请求出错了：' + err.message)
    })
}

// 打开取消订单弹窗
function cancelOrderFn(rowItem: any, event?: Event) {
  event && event.stopPropagation()
  cancelDialogVisible.value = true
  orderId.value = rowItem.id
  dialogOrderStatus.value = rowItem.status
  cancelDialogTitle.value = '取消'
  dialogVisible.value = false
  cancelReason.value = ''
}

// 打开拒单弹窗
function orderRejectFn(rowItem: any, event?: Event) {
  event && event.stopPropagation()
  cancelDialogVisible.value = true
  orderId.value = rowItem.id
  dialogOrderStatus.value = rowItem.status
  cancelDialogTitle.value = '拒绝'
  dialogVisible.value = false
  cancelReason.value = ''
}

// 确认取消或拒绝订单并填写原因
function confirmCancel() {
  if (!cancelReason.value) {
    return proxy.$message.error(`请选择${cancelDialogTitle.value}原因`)
  } else if (cancelReason.value === '自定义原因' && !remark.value) {
    return proxy.$message.error(`请输入${cancelDialogTitle.value}原因`)
  }

  const fn = cancelDialogTitle.value === '取消' ? orderCancel : orderReject
  const payload: any = {
    id: orderId.value,
  }
  payload[
    cancelDialogTitle.value === '取消' ? 'cancelReason' : 'rejectionReason'
  ] = cancelReason.value === '自定义原因' ? remark.value : cancelReason.value

  fn(payload)
    .then((res: any) => {
        if (Number(res.data.code) === 1) {
        proxy.$message.success('操作成功')
        cancelDialogVisible.value = false
        const emittedId = orderId.value
        orderId.value = ''
        try { emitOrderChanged({ orderId: emittedId }) } catch (e) {}
      } else {
        proxy.$message.error(res.data.msg)
      }
    })
    .catch((err: any) => {
      proxy.$message.error('请求出错了：' + err.message)
    })
}

// 派送，完成
function cancelOrDeliveryOrComplete(s: number, id: string, event?: Event) {
  event && event.stopPropagation()
  const params = {
    status: s,
    id,
  }
  const fn = s === 3 ? deliveryOrder : completeOrder
  fn(params)
    .then((res: any) => {
        if (Number(res.data.code) === 1) {
        proxy.$message.success('操作成功')
        const emittedId = orderId.value
        orderId.value = ''
        dialogVisible.value = false
        try { emitOrderChanged({ orderId: emittedId }) } catch (e) {}
      } else {
        proxy.$message.error(res.data.msg)
      }
    })
    .catch((err: any) => {
      proxy.$message.error('请求出错了：' + err.message)
    })
}

// 查看详情
async function goDetail(id: any, s: number, r: any, event?: Event) {
  event && event.stopPropagation()
  diaForm.value = {}
  dialogVisible.value = true
  dialogOrderStatus.value = s
    console.log("请求 URL = /merchant/order/detail", "参数 =", { orderId: id })
  console.log("rowItem.id =", id)
  try {
    const { data } = await queryOrderDetailByIdCoalesced({ orderId: id })
    const raw = data.data || {}
    // Normalize returned payload to modal fields
    const idVal = raw.id ?? raw.ID ?? raw.orderId ?? raw.orderID ?? raw.orderid ?? raw.orderNo ?? raw.number
    const numberVal = raw.number ?? raw.orderNo ?? raw.orderNumber ?? raw.orderId ?? idVal
    const amountVal = Number(raw.amount ?? raw.totalPrice ?? raw.totalprice ?? raw.total_price ?? 0)
    const orderDetailListVal = raw.orderDetailList ?? raw.orderDetails ?? raw.items ?? raw.details ?? []
    const packAmount = Number(raw.packAmount ?? raw.pack_amount ?? raw.packamount ?? 0)
    const deliveryAmount = Number(raw.deliveryAmount ?? raw.delivery_amount ?? raw.deliveryFee ?? raw.delivery_fee ?? raw.delivery ?? 0)
    const checkoutTime = raw.checkoutTime ?? raw.payTime ?? raw.pay_time ?? raw.pay_at ?? raw.PayAt
    const expected = raw.expected_time ?? raw.expectedTime ?? raw.estimatedDeliveryTime ?? raw.expectedtime
    const deliveryTime = raw.deliveryTime ?? raw.deliverAt ?? raw.finishAt
    diaForm.value = Object.assign({}, raw, {
      id: idVal,
      orderid: numberVal,
      number: numberVal,
      amount: amountVal,
      orderDetailList: orderDetailListVal,
      packAmount: packAmount,
      deliveryAmount: deliveryAmount,
      checkoutTime: checkoutTime,
      expectedDeliveryTime: expected,
      deliveryTime: deliveryTime,
    })
    row.value = r || { id: idVal, status: s }
  } catch (err: any) {
    proxy.$message.error('请求出错了：' + err.message)
  }
}

// 关闭弹层
function handleClose() {
  dialogVisible.value = false
}

// tab切换
function handleClass(index: number) {
  activeIndex.value = index
  if (index === 0) {
    status.value = 2
    getOrderListData(2)
  } else {
    status.value = 3
    getOrderListData(3)
  }
}

// 触发table某一行
function handleTable(rowItem: any, column: any, event?: Event) {
  event && event.stopPropagation()
  
  goDetail(rowItem.orderid, rowItem.status, rowItem, event)
}

// 分页
function handleSizeChange(val: any) {
  pageSize.value = val
  getOrderListData(status.value)
}

function handleCurrentChange(val: any) {
  page.value = val
  getOrderListData(status.value)
}
</script>
<style  lang="scss" scoped >
.dashboard-container.home .homecon {
  margin-bottom: 0;
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

.jd-logo {
  width: 36px;
  height: 36px;
  object-fit: contain;
  vertical-align: middle;
  margin-right: 8px;
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
      height: 43px;
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
        .dish-num {
          margin-right: 51px;
        }
      }
      // .dish-item:nth-child(odd) {
      //   flex: 60%;
      // }
      // .dish-item:nth-child(even) {
      //   flex: 40%;
      // }
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
/* 原生模态样式（与上层样式分离，便于复用） */
<style scoped>
.native-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}
.native-modal {
  width: 860px;
  max-width: calc(100% - 40px);
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 8px 30px rgba(0,0,0,0.2);
  overflow: hidden;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 24px;
  border-bottom: 1px solid #eee;
}
.modal-header .order-num {
  font-weight: 700;
  margin-left: 12px;
  color: #222;
}
.modal-close {
  cursor: pointer;
  font-size: 16px;
  color: #999;
}
.modal-body {
  max-height: 520px;
  overflow: auto;
  padding: 18px 24px;
}
.modal-footer {
  padding: 12px 20px;
  border-top: 1px solid #f0f0f0;
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: flex-end;
}
.modal-footer .btn {
  padding: 8px 14px;
  border-radius: 6px;
  border: 1px solid #dcdcdc;
  background: #fff;
  cursor: pointer;
}
.modal-footer .btn.primary {
  background: #409EFF;
  color: #fff;
  border-color: #409EFF;
}
.modal-footer .auto-next {
  margin-right: auto;
  font-size: 13px;
  color: #333;
}
</style>
<style  lang="scss">
.dashboard-container {
  .cancelTime {
    padding-left: 30px;
  }
  .orderTime {
    padding-left: 30px;
  }
  td.operate .cell {
    display: flex;
    flex-wrap: nowrap;
    justify-content: center;
    align-items: center;
    gap: 8px;
  }
  td.otherOperate .cell {
    display: flex;
    flex-wrap: nowrap;
    justify-content: center;
    align-items: center;
    gap: 8px;
  }
  td.operate .cell .before,
  td.operate .cell .middle,
  td.operate .cell .after,
  td.otherOperate .cell .before,
  td.otherOperate .cell .middle,
  td.otherOperate .cell .after {
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
  }
  /* Ensure buttons don't wrap into multiple lines */
  td.operate .cell button,
  td.otherOperate .cell button {
    white-space: nowrap;
    padding: 0 6px;
  }
}
</style>
