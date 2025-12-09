import noImg from '@/assets/noImg.png'

export const sampleStores = [
  {
    name: '黄焖鸡米饭',
    desc: '经典家常菜，味道鲜香',
    logo: noImg,
    tags: ['家常菜', '下饭王'],
    rating: 4.8,
    sales: 320,
    minOrder: 15,
    deliveryFee: 2,
    dishes: [
      { name: '黄焖鸡套餐' },
      { name: '香菇滑鸡饭' },
      { name: '青椒土豆丝' }
    ]
  },
  {
    name: '茶百道',
    desc: '奶香浓郁，果茶清爽',
    logo: noImg,
    tags: ['奶茶', '饮品', '水果茶'],
    rating: 4.9,
    sales: 520,
    minOrder: 12,
    deliveryFee: 1,
    dishes: [
      { name: '乌龙奶茶' },
      { name: '杨枝甘露' },
      { name: '芝士奶盖' }
    ]
  }
]

export default sampleStores
