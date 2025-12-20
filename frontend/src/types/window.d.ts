declare global {
  interface Window {
    __merchant_open_order_lock?: string | null
    __orders_refreshHandler?: any
  }
}

export { }
