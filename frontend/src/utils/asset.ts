export function resolveAssetUrl(url) {
  if (!url) return ''
  const u = String(url)
  if (u.startsWith('http://') || u.startsWith('https://') || u.startsWith('data:')) return u
  if (u.startsWith('/')) {
    // Prefix with configured API base if available (for uploaded images served by backend)
    const base = import.meta.env.VITE_API_BASE_URL || ''
    if (base) {
      return base.replace(/\/$/, '') + u
    }
    return u
  }
  // If it's a bare filename (no leading slash), assume it's stored under /images on the backend
  const base = import.meta.env.VITE_API_BASE_URL || ''
  if (base) {
    return base.replace(/\/$/, '') + '/images/' + u
  }
  return '/images/' + u
}

export function safeImage(url, fallback) {
  const r = resolveAssetUrl(url)
  if (!r) return fallback
  return r
}
