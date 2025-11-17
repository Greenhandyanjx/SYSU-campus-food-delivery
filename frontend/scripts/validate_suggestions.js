import sampleStores from '../src/data/sampleStores.js'

function getSuggestions(q) {
  const s = (q || '').toString().trim().toLowerCase()
  if (!s) return sampleStores.slice(0, 5)
  const nameMatches = sampleStores.filter(x => (x.name || '').toLowerCase().includes(s))
  const dishMatches = sampleStores.filter(x => (x.dishes || []).some(d => (d.name || '').toLowerCase().includes(s)))
  const combined = [...nameMatches, ...dishMatches].filter((v, i, a) => a.indexOf(v) === i)
  return combined.length ? combined : sampleStores.slice(0, 5)
}

// tests
const cases = [
  { q: '', expectMin: 1, desc: '空字符串应返回推荐（至少1条）' },
  { q: '茶', expectContains: '茶百道', desc: '输入茶 应匹配 茶百道' },
  { q: '鸡', expectContains: '黄焖鸡米饭', desc: '输入鸡 应匹配 黄焖鸡米饭' },
  { q: '乌龙', expectContains: '茶百道', desc: '输入菜名 乌龙 -> 茶百道 via dish match' },
  { q: '不存在的词', expectMin: 1, desc: '不存在词 应返回推荐（非空）' }
]

let failures = 0
for (const c of cases) {
  const res = getSuggestions(c.q)
  const names = res.map(r => r.name)
  let ok = true
  if (c.expectContains) { ok = names.includes(c.expectContains) }
  if (c.expectMin) { ok = res.length >= c.expectMin }
  if (!ok) {
    console.error('FAILED:', c.desc, 'query=', c.q, 'got=', names)
    failures++
  } else {
    console.log('OK:', c.desc, '->', names)
  }
}

if (failures > 0) {
  console.error(failures, 'tests failed')
  process.exit(2)
} else {
  console.log('All suggestion tests passed')
}
