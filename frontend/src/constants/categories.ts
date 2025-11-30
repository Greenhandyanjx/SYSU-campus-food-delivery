export const CATEGORIES = [
  { id: 0, label: '全部', key: 'all', filter: null },
  { id: 1, label: '招牌套餐', key: 'setmeal', filter: 'setmeal' },
  { id: 2, label: '现煮粉面', key: 'noodle', filter: 'noodle' },
  { id: 3, label: '汉堡炸鸡', key: 'burger', filter: 'burger' },
  { id: 4, label: '奶茶咖啡', key: 'milktea', filter: 'milktea' },
  { id: 5, label: '日式便当', key: 'bento', filter: 'bento' },
  { id: 6, label: '烧烤烤肉', key: 'bbq', filter: 'bbq' },
  { id: 7, label: '水果拼盘', key: 'fruit', filter: 'fruit' },
  { id: 8, label: '精致甜品', key: 'dessert', filter: 'dessert' },
  { id: 9, label: '家常快炒', key: 'stirfry', filter: 'stirfry' },
  { id: 10, label: '粥粉面饭', key: 'rice', filter: 'rice' },
  { id: 11, label: '极速配送', key: 'fast_delivery', filter: 'fast' },
  { id: 12, label: '午餐推荐', key: 'lunch', filter: 'lunch' },
  { id: 13, label: '低价满减', key: 'low_price', filter: 'discount' },
  { id: 14, label: '沙拉轻食', key: 'salad', filter: 'salad' },
  { id: 15, label: '精致下午茶', key: 'afternoon', filter: 'afternoon' }
]

export function isValidCategoryId(id: number) {
  return id >= 1 && id <= 15
}
