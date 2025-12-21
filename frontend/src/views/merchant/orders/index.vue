<template>
  <div class="dashboard-container">
    <div class="orders-header">
      <img src="/JDlogo.png" class="jd-logo" alt="嘉递外卖" />
      <TabChange
      :order-statics="orderStatics"
      :default-activity="defaultActivity"
      @tabChange="change"
    />
    </div>
  <div class="container main-container" :class="{ hContainer: tableData.length }" >
      <!-- 搜索项 -->
      <div class="tableBar">
        <label style="margin-right: 10px">订单号：</label>
        <el-input
          v-model="input"
          placeholder="请填写订单号"
          style="width: 15%"
          clearable
          @clear="onClear"
          @keyup.enter="initFun(orderStatus.value)"
        />
        <label style="margin-left: 20px">手机号：</label>
        <el-input
          v-model="phone"
          placeholder="请填写手机号"
          style="width: 15%"
          clearable
          @clear="onClear"
          @keyup.enter="initFun(orderStatus.value)"
        />
        <label style="margin-left: 20px">下单时间：</label>
<el-date-picker
  v-model="valueTime"
  type="daterange"
  unlink-panels
  clearable
  range-separator="至"
  start-placeholder="开始日期"
  end-placeholder="结束日期"
  class="date-range"
/>
<el-button 
  class="normal-btn continue" 
  @click="init(orderStatus.value, true)"
  :loading="loading"
>
  查询
</el-button>
      </div>
      <el-table
        v-if="tableData.length"
        :data="visibleTableData"
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
        >
          <template #default="{ row }">
            <span>{{ formatDateToCN(row.orderTime) }}</span>
          </template>
        </el-table-column>
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
          key="expected_time"
          prop="expected_time"
          label="预计送达时间"
          min-width="110"
        >
          <template #default="{ row }">
            <span>{{formatDateToCN(row.expected_time) }}</span>
          </template>
        </el-table-column>
        <el-table-column
          v-if="[0, 2, 5].includes(orderStatus)"
          key="amount"
          prop="totalprice"
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
          prop="quantity"
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
              ? 180
              : [0].includes(orderStatus)
              ? 200
              : 240
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
                @click="goDetail(row.id || row.orderId || row.orderid, row.status, row, $event)"
              >
                查看
              </el-button>
              <el-button
                type="text"
                class="blueBug non"
                @click="openChatForOrder(row)"
              >
                聊天
              </el-button>
              
            </div>
          </template>
        </el-table-column>
      </el-table>
      <Empty v-else :is-search="isSearch" />
      <el-pagination
        v-if="counts > pageSize"
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
    <!-- 原生模态替代 el-dialog -->
    <div v-if="dialogVisible" class="native-modal-overlay" @click.self="dialogVisible = false">
      <div class="native-modal" :key="modalKey" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div>
            <label style="font-size:16px">订单号：</label>
            <span class="order-num">{{ diaForm.number || diaForm.orderNo || diaForm.ID || diaForm.id }}</span>
          </div>
          <div class="modal-close" @click="dialogVisible = false">✕</div>
        </div>

        <div class="modal-body">
          <div v-if="modalLoading" class="modal-loading-overlay">加载中...</div>
          <div class="order-top">
            <p><label>下单时间：</label>{{ diaForm.orderTime || diaForm.createTime || diaForm.createdAt }}</p>
          </div>

          <div class="order-middle">
            <div class="user-info">
              <div class="user-info-box">
                <div class="user-name"><label>用户名：</label><span>{{ diaForm.consignee }}</span></div>
                <div class="user-phone"><label>手机号：</label><span>{{ diaForm.phone }}</span></div>
                <div v-if="[2,3,4,5].includes(dialogOrderStatus)" class="user-getTime">
                  <label>{{ dialogOrderStatus === 5 ? '送达时间：' : '预计送达时间：' }}</label>
                  <span>{{ dialogOrderStatus === 5 ? formatDateToCN(diaForm.deliveryTime || diaForm.delivery_time) : formatDateToCN(diaForm.estimatedDeliveryTime || diaForm.expected_time) }}</span>
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
                  <div class="dish-item-box">
                    <span class="dish-name">{{ item.name }}</span>
                    <span class="dish-num">x{{ item.number || item.qty || item.quantity }}</span>
                  </div>
                  <span class="dish-price">￥{{ item.amount ? Number(item.amount).toFixed(2) : (item.price ? Number(item.price).toFixed(2) : '') }}</span>
                </div>
              </div>
              <div class="dish-all-amount"><label>菜品小计</label><span>￥{{ ((Number(diaForm.amount || 0) - diaForm.deliveryFee - Number(diaForm.packAmount || 0)).toFixed(2)) }}</span></div>
            </div>
          </div>

          <div class="order-bottom">
            <div class="amount-info">
              <div class="amount-label">费用</div>
              <div class="amount-list">
                <div class="dish-amount"><span class="amount-name">菜品小计：</span><span class="amount-price">￥{{ (Number(Number(diaForm.amount || 0) - diaForm.deliveryFee - Number(diaForm.packAmount || 0)).toFixed(2)) }}</span></div>
                <div class="send-amount"><span class="amount-name">派送费：</span><span class="amount-price">￥{{ diaForm.deliveryFee ? Number(diaForm.deliveryFee).toFixed(2) : '' }}</span></div>
                <div class="package-amount"><span class="amount-name">打包费：</span><span class="amount-price">￥{{ diaForm.packAmount ? Number(diaForm.packAmount||0).toFixed(2) : '0.00' }}</span></div>
                <div class="all-amount"><span class="amount-name">合计：</span><span class="amount-price">￥{{ diaForm.amount ? Number(diaForm.amount).toFixed(2) : '' }}</span></div>
                <div class="pay-type"><span class="pay-name">支付渠道：</span><span class="pay-value">{{ diaForm.payMethod === 1 ? '微信支付' : '支付宝支付' }}</span></div>
                <div class="pay-time"><span class="pay-name">支付时间：</span><span class="pay-value">{{ diaForm.checkoutTime }}</span></div>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer" v-if="dialogOrderStatus !== 6">
          <label v-if="dialogOrderStatus === 2 && orderStatus === 2" class="auto-next"><input type="checkbox" v-model="isAutoNext" /> 处理完自动跳转下一条</label>
          <button v-if="dialogOrderStatus === 2" class="btn" @click="orderRejectHandler(row, false)">拒 单</button>
          <button v-if="dialogOrderStatus === 2" class="btn primary" @click="orderAcceptHandler(row, false)">接 单</button>
          <button v-if="[1,3,4,5].includes(dialogOrderStatus)" class="btn" @click="dialogVisible = false">返 回</button>
          <button v-if="dialogOrderStatus === 3" class="btn primary" @click="cancelOrDeliveryOrComplete(3, row.id)">派 送</button>
          <button v-if="dialogOrderStatus === 4" class="btn primary" @click="cancelOrDeliveryOrComplete(4, row.id)">完 成</button>
          <button v-if="[1].includes(dialogOrderStatus)" class="btn primary" @click="cancelOrderHandler(row)">取消订单</button>
        </div>
      </div>
    </div>
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
import { ref, reactive, onMounted, onBeforeUnmount, watch, computed, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getMerchantProfile } from '@/api/merchant/profile'
import HeadLable from '@/components/HeadLable/index.vue'
import InputAutoComplete from '@/components/InputAutoComplete/index.vue'
import TabChange from './tabChange.vue'
import Empty from '@/components/Empty/index.vue'
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
import request from '@/api/merchant/request'

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
// key 用于强制重渲染详情模态，避免请求返回但 DOM 未更新的情况
const modalKey = ref(0)

// 表格只展示当前页的数据，防止后端返回全部项导致表格显示超过 pageSize
const visibleTableData = computed(() => {
  if (!Array.isArray(tableData.value)) return []
  // If backend returns only the current page (common), avoid slicing
  if (tableData.value.length <= pageSize.value) return tableData.value
  // Otherwise assume tableData contains full dataset and slice for page
  const start = (page.value - 1) * pageSize.value
  const end = start + pageSize.value
  return tableData.value.slice(start, end)
})
const currentMerchantId = ref<any>(null)
const diaForm = ref<any>({})
const modalLoading = ref(false)
const isSearch = ref(false)
const orderStatus = ref(0)
const dialogOrderStatus = ref(0)
const loading = ref(false)
const onClear = () => {
  valueTime.value = undefined // ❌ 不再是 []
  console.log("valueTime = ", valueTime.value)
  init(orderStatus.value)
}

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

// 辅助：把 Date 或可解析的时间字符串格式化为 API 所需的 'yyyy-MM-dd HH:mm:ss'
function formatForApi(v: any) {
  if (!v) return ''
  const d = v instanceof Date ? v : new Date(v)
  if (isNaN(d.getTime())) return String(v)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(
    d.getHours()
  )}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 格式化为中文可读时间：YYYY年MM月DD日 HH:mm
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
watch(() => route.query.orderId, (val) => {
    try {
      if (val && String(val) !== 'undefined') goDetail(String(val), 2)
    } catch (e) { console.warn('route query orderId watch failed', e) }
  })
  ;(async () => {
    try {
      const r: any = await getMerchantProfile()
      if (r && r.data && Number(r.data.code) === 1 && r.data.data) {
        currentMerchantId.value = r.data.data.id || r.data.data.ID || r.data.data.merchant_id || r.data.data.merchantId || null
      }
    } catch (e) {
      // defensive: ignore, backend should already filter
    }
})()
onMounted(() => {
  const status = Number(route.query.status) || 0
  defaultActivity.value = status
  init(status)

  if (route.query.orderId && route.query.orderId !== 'undefined') {
    goDetail(route.query.orderId as string, 2)
  }
  // 监听 route.query.orderId 的变化（通过其他组件路由跳转携带 orderId）
 

  // 监听 order:changed，去抖刷新页面数据
  let __orders_refreshTimer: any = null
  const __orders_refreshHandler = (ev: any) => {
    try {
      clearTimeout(__orders_refreshTimer)
      __orders_refreshTimer = setTimeout(() => {
        init(orderStatus.value)
      }, 500)
    } catch (e) { console.warn('order:changed handler failed', e) }
  }
  window.addEventListener('order:changed', __orders_refreshHandler)
  // store handler reference globally so duplicate listeners are not added accidentally
  ;(window as any).__orders_refreshHandler = __orders_refreshHandler

})


  // 监听外部打开指定订单的事件（作为路由 query 打开失败的回退）
  const __merchant_open_handler = (ev: any) => {
    try {
      const id = ev && ev.detail && (ev.detail.orderId || ev.detail.orderId === 0 ? ev.detail.orderId : null)
      if (id) goDetail(String(id), 2)
    } catch (e) { console.warn('merchant open handler failed', e) }
  }
  try { window.addEventListener('merchant:open_order', __merchant_open_handler) } catch (e) {}
  ;(window as any).__merchant_open_handler = __merchant_open_handler
onBeforeUnmount(() => {
  try { window.removeEventListener('order:changed', (window as any).__orders_refreshHandler || __orders_refreshHandler) } catch (e) {}
  try { window.removeEventListener('merchant:open_order', (window as any).__merchant_open_handler || __merchant_open_handler) } catch (e) {}
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
    if (Number(res.data.code) === 1) {
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
//   if (!Array.isArray(valueTime.value)) {
//   valueTime.value = []
// }
  console.log('init 调用', { activeIndex, isSearchFlag })
  if (loading.value) return
  loading.value = true
  try
  {console.log("valueTime = ", valueTime.value)
  isSearch.value = !!isSearchFlag
  const params: any = {
    page: page.value,
    pageSize: pageSize.value,
    number: input.value || undefined,
    phone: phone.value || undefined,
// 在发送请求时手动格式化
beginTime: valueTime.value[0] ? formatForApi(valueTime.value[0]) : undefined,
endTime: valueTime.value[1] ? formatForApi(valueTime.value[1]) : undefined,

    status: activeIndex || undefined,
  }
  try {
    const res = await getOrderDetailPageCoalesced({ ...params })
    if (Number(res.data.code) === 1) {
      const data = res.data.data || {}
      const raw = data.items || []
      // 后端应只返回当前商家的订单；为保险起见，在前端二次筛选
      const filtered = raw.filter((it: any) => {
        if (!currentMerchantId.value) return true
        const mid = it.merchant_id ?? it.merchantId ?? it.merchantid ?? it.MerchantID ?? it.merchant
        return String(mid) === String(currentMerchantId.value)
      })
      // 格式化时间字段，防止前端出现 NaN 或 undefined
      tableData.value = filtered.map((it: any) => {
        const safeFormat = (v: any) => {
          if (v === null || v === undefined || v === '') return ''
          try {
            const d = new Date(v)
            if (!d || isNaN(d.getTime())) return String(v)
            return d.toLocaleString()
          } catch (e) {
            return String(v)
          }
        }

        // 兼容后端不同命名：id/orderid/OrderID，订单号使用 number 字段
        const id = it.id ?? it.ID ?? it.orderId ?? it.orderID ?? it.orderid ?? it.OrderID
        const number = it.number ?? it.orderNumber ?? it.orderNo ?? it.orderid ?? it.orderId ?? id

        // 金额兼容：amount / totalPrice / totalprice / total_price
        const rawAmount =
          it.amount ?? it.totalPrice ?? it.totalprice ?? it.total_price ?? it.totalPrice ?? 0

        // 订单明细兼容
        const orderDetailList = it.orderDetailList ?? it.orderDetails ?? it.details ?? it.items ?? it.order_items ?? []

        return {
          ...it,
          id,
          number,
          orderDetailList,

          orderTime: safeFormat(
            it.orderTime ??
            it.order_time ??
            it.createTime ??
            it.create_time ??
            it.createdAt ??
            it.CreatedAt ??
            it.created_at
          ),

          cancelTime: safeFormat(
            it.cancelTime ??
            it.cancel_time ??
            it.cancelAt ??
            it.CancelAt
          ),

          deliveryTime: safeFormat(
            it.deliveryTime ??
            it.delivery_time ??
            it.deliveredAt ??
            it.DeliveredAt
          ),

          expected_time: safeFormat(
            it.estimatedDeliveryTime ??
            it.estimated_delivery_time ??
            it.expected_time ??
            it.Expectedtime ??
            it.expectedTime ??
            it.ExpectedTime
          ),

          checkoutTime: safeFormat(
            it.checkoutTime ??
            it.checkout_time ??
            it.paidAt ??
            it.PaidAt ??
            it.paymentTime
          ),

          amount:
            typeof rawAmount === 'number'
              ? rawAmount
              : rawAmount
              ? Number(rawAmount)
              : 0,

          packAmount:
            it.packAmount ??
            it.pack_amount ??
            it.packageFee ??
            it.packFee ??
            0,
        }

      })
      console.log("valueTime = ", valueTime.value)
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
        goDetail(r.id || r.orderId || r.orderid, r.status, r)
      }
      console.log("后端返回原始数据:", raw)
      console.log("格式化后的数据:", tableData.value)

    } else {
      ElMessage.error(res?.data?.msg || '获取订单列表失败')
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }}finally {
    loading.value = false
  }
}

async function goDetail(id: any, status: number, r?: any) {
  if (!id) return
  try {
    // 如果已有锁但不是当前要打开的订单，则阻止重复打开；
    // 如果锁正是当前 id（由触发方设置），允许继续处理。
    if (window.__merchant_open_order_lock && window.__merchant_open_order_lock !== String(id)) {
      return
    }
  } catch (e) {}
  try { window.__merchant_open_order_lock = String(id) } catch (e) {}
  // 自动在 3 秒后解锁，避免死锁
  setTimeout(() => { try { if (window.__merchant_open_order_lock === String(id)) window.__merchant_open_order_lock = null } catch (e) {} }, 3000)

  // 显示加载状态，等数据准备好再显示弹窗，避免空白需要手动刷新
  modalLoading.value = true
  dialogOrderStatus.value = status
  orderId.value = id
  try {
    // add client-side timeout to avoid hanging UI
    const fetchPromise = queryOrderDetailByIdCoalesced({ orderId: id })
    const timeoutPromise = new Promise((_, rej) => setTimeout(() => rej(new Error('请求超时（客户端）')), 8000))
    // be defensive: some API wrappers return `response` while others return `response.data`
    const resp = await Promise.race([fetchPromise, timeoutPromise])
    // normalize response shape
    const respData = resp && (resp.data || resp)
    const raw = (respData && (respData.data || respData)) || {}

    const safeFormat = (v: any) => {
      if (!v) return ''
      let s: string
      if (typeof v === 'string') s = v
      else if (v instanceof Date) s = v.toISOString()
      else s = String(v)
      if (s === '0001-01-01T00:00:00Z' || s.startsWith('0001-01-01')) return ''
      if (s.includes(' ') && !s.includes('T')) s = s.replace(' ', 'T')
      const d = new Date(s)
      if (isNaN(d.getTime())) return ''
      const pad = (n: number) => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    }

    const idVal = raw.id ?? raw.ID ?? raw.orderId ?? raw.orderID ?? raw.orderid ?? raw.OrderID
    const numberVal = raw.number ?? raw.orderNumber ?? raw.orderNo ?? raw.orderid ?? raw.orderId ?? idVal
    const amountVal = raw.amount ?? raw.totalPrice ?? raw.totalprice ?? raw.total_price ?? 0
    const orderDetailListVal = raw.orderDetailList ?? raw.orderDetails ?? raw.details ?? raw.items ?? raw.order_items ?? []

    diaForm.value = {
      ...raw,
      id: idVal,
      number: numberVal,
      orderDetailList: orderDetailListVal,
      orderTime: safeFormat(raw.orderTime ?? raw.createTime ?? raw.createdAt ?? raw.created_at ?? raw.create_time),
      cancelTime: safeFormat(raw.cancelTime ?? raw.cancel_time ?? raw.cancelAt),
      deliveryTime: safeFormat(raw.deliveryTime ?? raw.delivery_time ?? raw.deliveredAt),
      estimatedDeliveryTime: safeFormat(raw.estimatedDeliveryTime ?? raw.expectedtime ?? raw.expectedTime),
      checkoutTime: safeFormat(raw.checkoutTime ?? raw.paidAt ?? raw.paymentTime),
      amount: typeof amountVal === 'number' ? amountVal : amountVal ? Number(amountVal) : 0,
      packAmount: raw.packAmount ?? raw.pack_amount ?? raw.packageFee ?? raw.packFee ?? '',
    }

    row.value = r || { id: route.query.orderId, status }
    // 清理 URL 中的 orderId 查询参数：使用 history.replaceState 避免通过 router 触发组件重渲染
    try {
      if (route.query.orderId) window.history.replaceState(undefined, '', '/merchant/orders')
    } catch (e) {}

    // 强制刷新 modal 渲染并在数据准备好后显示弹窗
    try { modalKey.value = Date.now(); await nextTick() } catch (e) {}
    dialogVisible.value = true
  } catch (err: any) {
    ElMessage.error('请求出错了：' + (err?.message || err))
    // 仍然尝试显示弹窗以便用户看到错误信息或手动重试
    dialogVisible.value = true
  } finally {
    modalLoading.value = false
    try { if (window.__merchant_open_order_lock === String(id)) window.__merchant_open_order_lock = null } catch (e) {}
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
    if (Number(res.data.code) === 1) {
      ElMessage.success('操作成功')
      const emittedId = orderId.value
      orderId.value = ''
      dialogVisible.value = false
      try { emitOrderChanged({ orderId: emittedId }) } catch (e) {}
    } else {
      ElMessage.error(res.data.msg)
    }
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
}

function cancelOrderHandler(r: any) {
  // 直接调用取消接口，将订单状态改为 6（商家取消）
  if (!r || !r.id) return ElMessage.error('订单 ID 缺失')
  orderId.value = r.id
  dialogOrderStatus.value = r.status
  isTableOperateBtn.value = true
  cancelDialogVisible.value = false
  try {
    ;(async () => {
      const payload: any = { id: orderId.value, cancelReason: '商家取消' }
      const res = await orderCancel(payload)
      if (Number(res.data.code) === 1) {
        ElMessage.success('取消订单成功')
        orderId.value = ''
        try { emitOrderChanged({ orderId: r.id }) } catch (e) {}
        // 刷新当前页列表
        init(orderStatus.value)
      } else {
        ElMessage.error(res.data.msg || '取消订单失败')
      }
    })()
  } catch (err: any) {
    ElMessage.error('请求出错了：' + err.message)
  }
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
    if (Number(res.data.code) === 1) {
      ElMessage.success('操作成功')
      cancelDialogVisible.value = false
      const emittedId = orderId.value
      orderId.value = ''
      try { emitOrderChanged({ orderId: emittedId }) } catch (e) {}
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
    if (Number(res.data.code) === 1) {
      ElMessage.success('操作成功')
      const emittedId = orderId.value
      orderId.value = ''
      dialogVisible.value = false
      try { emitOrderChanged({ orderId: emittedId }) } catch (e) {}
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

// 顶层：根据订单打开聊天窗口（触发全局事件，dashboard/全局 modal 会响应）
async function openChatForOrder(row: any) {
  // 首先尝试从订单信息中读取 consignee id
  const consigneeId = row.consigneeid || row.consigneeId || row.Consigneeid || row.ConsigneeID || row.consigneeID
  if (!consigneeId) {
    ElMessage.error('无法定位订单对应的 consignee id')
    return
  }
  try {
    const res = await request.get('/consignee/query', { params: { id: consigneeId } })
    const data = res && res.data && (res.data.data || res.data)
    const uid = data && (data.userid || data.Userid || data.userId || data.UserID)
    if (!uid) {
      ElMessage.error('未找到收货人对应的用户 ID')
      return
    }
    const mid = currentMerchantId.value ?? null
    window.dispatchEvent(new CustomEvent('chat:open', { detail: { merchantId: mid, userBaseId: uid } }))
  } catch (e: any) {
    console.warn('openChatForOrder failed', e)
    ElMessage.error('打开聊天失败：' + (e?.message || e))
  }
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
      .tableBar .el-input {
  height: 40px;
}

.tableBar .el-input__wrapper {
  height: 40px;
  border-radius: 6px;
  padding: 0 12px;
}

      /* Ensure inputs have consistent height so placeholder text isn't clipped */
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
.date-range {
  width: 320px;
}

</style>

<style scoped>
.orders-header { display:flex; align-items:center; gap:10px; }
.jd-logo { width:36px; height:36px; object-fit:contain; }
</style>

/* 原生模态样式 */
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

/* 弹窗加载遮罩 */
.modal-loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255,255,255,0.85);
  z-index: 30;
  font-size: 16px;
  color: #333;
}

</style>
