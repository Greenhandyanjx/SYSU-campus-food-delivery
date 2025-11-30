import request from '@/api/merchant/request'
import { number } from 'echarts'

/**
 * cart.ts API 说明（后端实现参考）
 *
 * 数据结构（前端与后端约定示例）:
 * {
 *   shops: [
 *     {
 *       storeId: string,
 *       name: string,
 *       selected: boolean, // 客户端选择状态（可选，后端可忽略）
 *       items: [
 *         {
 *           dishId: string,
 *           name: string,
 *           price: number,
 *           qty: number,         // 当前购物车中的数量
 *           originalQty:number,  // 初始加入购物车的数量（用于管理模式约束）
 *           selected: boolean,
 *           category?: string
 *         }
 *       ]
 *     }
 *   ]
 * }
 *
 * 后端接口约定（建议实现）：
 * 1) GET /user/cart
 *    请求: 无
 *    返回: { code: 1, data: { shops: [...] } }
 *    说明: 返回当前用户购物车明细。每个 items 中应包含 qty（当前数量）以及 originalQty（可选，若后端不提供，前端应以 qty 填充 originalQty）。
 *
 * 2) POST /user/cart/update
 *    请求: { storeId, dishId, qty }
 *    返回: { code:1, data: { success: true, updatedItem: { storeId,dishId,qty } } }
 *    说明: 更新购物车内指定菜品数量（qty 不得超过用户已下单或库存限制，返回更新后的数量）。
 *
 * 3) POST /user/cart/selectItem
 *    请求: { storeId, dishId, selected }
 *    返回: { code:1, data: { success:true } }
 *    说明: 标记单个菜品的预结算选择状态（有助于服务端计算优惠等）。
 *
 * 4) POST /user/cart/selectShop
 *    请求: { storeId, selected }
 *    返回: { code:1, data: { success:true } }
 *    说明: 标记店铺下所有菜品为选中/未选中。
 *
 * 5) POST /user/cart/selectAll
 *    请求: { selected }
 *    返回: { code:1, data: { success:true } }
 *    说明: 全局全选/全不选。
 *
 * 6) POST /user/cart/checkout
 *    请求: 无或 { shops:[{storeId, items:[{dishId, qty}]}] }
 *    返回: { code:1, data: { success:true, orderId } }
 *    说明: 下单/结算接口，返回订单 id。前端可传被选中的商品信息以供结算。
 *
 * 7) POST /user/cart/deleteSelected
 *    请求: 无或 { shops:[{storeId, items:[{dishId}] }] }
 *    返回: { code:1, data: { success:true, removed: {...} } }
 *    说明: 删除已选商品（管理模式删除）。
 *
 * 前端容错策略: 若后端接口调用失败（网络或尚未实现），前端使用 ensureDemo() 提供的 demo 数据并在内存中完成本地模拟。
 */

// Demo 数据与回退：若后端不可达，使用全局 demo 数据
function ensureDemo() {
  // 结构: window.__DEMO_CART__ = { shops: [ { storeId, name, selected, items: [ { dishId, name, price, qty, selected, category } ] } ] }
  // @ts-ignore
  if (!window.__DEMO_CART__) {
    // @ts-ignore
    window.__DEMO_CART__ = {
      shops: [
        {
          storeId: 's1',
          name: '黄焖鸡米饭',
          selected: false,
          items: [
            { dishId: 'd1', name: '黄焖鸡套餐', price: 18, qty: 1, selected: false, category: '招牌套餐' },
            { dishId: 'd2', name: '香菇滑鸡饭', price: 16, qty: 2, selected: false, category: '家常快炒' }
          ]
        },
        {
          storeId: 's2',
          name: '茶百道',
          selected: false,
          items: [
            { dishId: 'd3', name: '乌龙奶茶', price: 12, qty: 1, selected: false, category: '奶茶咖啡' },
            { dishId: 'd4', name: '百香果绿茶', price: 11, qty: 3, selected: false, category: '奶茶咖啡' }
          ]
        },
        {
          storeId: 's3',
          name: '茶百道',
          selected: false,
          items: [
            { dishId: 'd3', name: '乌龙奶茶', price: 12, qty: 1, selected: false, category: '奶茶咖啡' },
            { dishId: 'd4', name: '百香果绿茶', price: 11, qty: 3, selected: false, category: '奶茶咖啡' }
          ]
        },
        {
          storeId: 's4',
          name: '茶百道',
          selected: false,
          items: [
            { dishId: 'd3', name: '乌龙奶茶', price: 12, qty: 1, selected: false, category: '奶茶咖啡' },
            { dishId: 'd4', name: '百香果绿茶', price: 11, qty: 3, selected: false, category: '奶茶咖啡' }
          ]
        }
      ]
    }
  }
  // @ts-ignore
  return window.__DEMO_CART__
}

export async function getCart() {
  try {
    const res = await request.get('/user/cart')
    return res.data
  } catch (e) {
    // 返回 demo
    return ensureDemo()
  }
}

/**
 * getCart()
 * 请求: GET /user/cart
 * 返回示例: { code:1, data: { shops: [ { storeId, name, items:[{ dishId, name, price, qty, originalQty, selected, category }] } ] } }
 * 说明: 返回购物车完整结构，前端会使用 items.qty 与 items.originalQty（若后端未提供 originalQty，则前端以 qty 填充 originalQty）。
 */

export async function updateQty(params: { storeId: string; dishId: string; qty: number }) {
  try {
    const res = await request.post('/user/cart/update', params)
    return res.data
  } catch (e) {
    const demo = ensureDemo()
    const shop = demo.shops.find((s: any) => s.storeId === params.storeId)
    if (shop) {
      const it = shop.items.find((x: any) => x.dishId === params.dishId)
      if (it) it.qty = params.qty
    }
    return demo
  }
}

/**
 * updateQty(params)
 * 请求: POST /user/cart/update
 * 请求体示例: { storeId: 's1', dishId: 'd1', qty: 2 }
 * 返回示例: { code:1, data: { success:true, updated: { storeId, dishId, qty } } }
 * 说明: 后端需要校验 qty 是否有效（如库存、最大可下单数等）。前端在界面上会限制 qty 不会超过 originalQty（原始加入数量）。
 */

export async function toggleItemSelection(params: { storeId: string; dishId: string | number; selected: boolean }) {
  try {
    const res = await request.post('/user/cart/selectItem', params)
    return res.data
  } catch (e) {
    const demo = ensureDemo()
    const shop = demo.shops.find((s: any) => s.storeId === params.storeId)
    if (shop) {
      const it = shop.items.find((x: any) => x.dishId === params.dishId)
      if (it) it.selected = !!params.selected
      // 更新店铺选中状态
      shop.selected = shop.items.every((x: any) => !!x.selected)
    }
    return demo
  }
}

/**
 * toggleItemSelection(params)
 * 请求: POST /user/cart/selectItem
 * 请求体示例: { storeId: 's1', dishId: 'd1', selected: true }
 * 返回示例: { code:1, data: { success:true } }
 * 说明: 标记单个菜品的结算选择状态，后端可据此计算应付金额/促销等。
 */

export async function toggleShopSelection(params: { storeId: string; selected: boolean }) {
  try {
    const res = await request.post('/user/cart/selectShop', params)
    return res.data
  } catch (e) {
    const demo = ensureDemo()
    const shop = demo.shops.find((s: any) => s.storeId === params.storeId)
    if (shop) {
      shop.selected = !!params.selected
      shop.items.forEach((it: any) => (it.selected = !!params.selected))
    }
    return demo
  }
}

/**
 * toggleShopSelection(params)
 * 请求: POST /user/cart/selectShop
 * 请求体示例: { storeId:'s1', selected: true }
 * 返回示例: { code:1, data: { success:true } }
 * 说明: 标记某店铺下所有菜品为选中/未选，返回成功状态。
 */

export async function selectAll(selected: boolean) {
  try {
    const res = await request.post('/user/cart/selectAll', { selected })
    return res.data
  } catch (e) {
    const demo = ensureDemo()
    demo.shops.forEach((s: any) => {
      s.selected = !!selected
      s.items.forEach((it: any) => (it.selected = !!selected))
    })
    return demo
  }
}

/**
 * selectAll(selected)
 * 请求: POST /user/cart/selectAll
 * 请求体示例: { selected: true }
 * 返回示例: { code:1, data: { success:true } }
 * 说明: 全局全选/全不选。
 */

export async function checkout(payload?: any) {
  try {
    // 调用后端的 createPayOrder 接口，后端会返回 code_url 与 orderId
    const res = await request.post('/order/createPayOrder', payload)
    return res.data
  } catch (e) {
    // 模拟结算：返回已结算订单 id
    return { success: true, orderId: 'demo-order-' + Date.now(), code_url: '' }
  }
}

export async function createPending(payload?: any) {
  try {
    const res = await request.post('/order/createPending', payload)
    return res.data
  } catch (e) {
    return { code: 0, message: 'failed to create pending', error: e }
  }
}

/**
 * checkout()
 * 请求: POST /user/cart/checkout
 * 请求体: (可选) { shops: [ { storeId, items:[{ dishId, qty }] } ] } 或者不带体
 * 返回示例: { code:1, data: { success:true, orderId: 'order-12345' } }
 * 说明: 执行结算，下单后返回订单 id。后端需执行库存/优惠/支付前置校验。
 */

export async function deleteSelected() {
  try {
    const res = await request.post('/user/cart/deleteSelected')
    return res.data
  } catch (e) {
    const demo = ensureDemo()
    demo.shops.forEach((s: any) => {
      s.items = s.items.filter((it: any) => !it.selected)
    })
    // 删除空店铺
    demo.shops = demo.shops.filter((s: any) => s.items.length > 0)
    return demo
  }
}

/**
 * deleteSelected()
 * 请求: POST /user/cart/deleteSelected
 * 请求体: (可选) { shops: [ { storeId, items:[{ dishId }] } ] }
 * 返回示例: { code:1, data: { success:true, removed: { ... } } }
 * 说明: 删除已选商品（管理模式下），返回删除结果。前端可使用此接口在后端同步删除操作。
 */

export default {
  getCart,
  updateQty,
  toggleItemSelection,
  toggleShopSelection,
  selectAll,
  checkout,
  createPending,
  deleteSelected
}
