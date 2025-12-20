let _timer: any = null
let _pending: any = null
const _DEBOUNCE_MS = 250

export function emitOrderChanged(detail?: any) {
  try {
    _pending = { ...(typeof detail === 'object' ? detail : { orderId: detail }) }
    if (_timer) return
    _timer = setTimeout(() => {
      try {
        window.dispatchEvent(new CustomEvent('order:changed', { detail: _pending }))
      } catch (e) {
        // best-effort
        console.warn('emitOrderChanged dispatch failed', e)
      }
      _pending = null
      clearTimeout(_timer)
      _timer = null
    }, _DEBOUNCE_MS)
  } catch (e) {
    console.warn('emitOrderChanged failed', e)
  }
}

export function clearPendingEmit() {
  if (_timer) {
    clearTimeout(_timer)
    _timer = null
    _pending = null
  }
}

export default { emitOrderChanged, clearPendingEmit }
