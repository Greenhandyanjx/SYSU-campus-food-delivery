import noImg from '@/assets/noImg.png'

export default [
  {
    name: '黄焖鸡米饭',
    desc: '经典家常菜，味道鲜香',
    logo: noImg,
    tags: ['家常菜', '下饭王'],
    rating: 4.8,
    sales: 320,
    minOrder: 15,
    deliveryFee: 2,
    dishes: [{ name: '黄焖鸡套餐' }, { name: '香菇滑鸡饭' }, { name: '青椒土豆丝' }]
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
    dishes: [{ name: '乌龙奶茶' }, { name: '杨枝甘露' }, { name: '芝士奶盖' }]
  }
  // 以下为新增测试数据，包含字段变体（有的没有 desc 或 tags）
  , {
    title: '麻辣香锅小站',
    summary: '香辣过瘾，配菜丰富',
    logo: noImg,
    categories: ['家常快炒', '麻辣'],
    rating: 4.6,
    sales: 210,
    dishes: [{ name: '麻辣香锅' }, { name: '香辣鸡排' }]
  },
  {
    shopName: '煲仔饭皇',
    brief: '广式风味，热气腾腾',
    logo: noImg,
    tags: ['煲仔饭'],
    rating: 4.7,
    sales: 180,
    dishes: [{ title: '咸鱼鸡粒煲' }, { title: '腊味煲仔饭' }]
  },
  {
    name: '清新沙拉屋',
    // 没有 desc 或 tags，测试回退到 dishes
    logo: noImg,
    rating: 4.5,
    sales: 95,
    dishes: [{ name: '凯撒沙拉' }, { name: '水果沙拉' }]
  },
  {
    name: '老北京炸酱面',
    desc: '经典北方面食',
    logo: noImg,
    tags: ['面食'],
    rating: 4.4,
    sales: 260,
    dishes: [{ name: '炸酱面' }, { name: '卤煮' }]
  }
]
