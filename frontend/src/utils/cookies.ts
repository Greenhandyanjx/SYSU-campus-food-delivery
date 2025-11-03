// 简单的 token 获取工具，优先从 localStorage 读取
export function getToken(): string {
  try {
    if (typeof window !== 'undefined' && window.localStorage) {
      return localStorage.getItem('token') || ''
    }
  } catch (e) {
    // ignore
  }
  return ''
}
