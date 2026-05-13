-- =============================================================================
-- seed_more_data.sql — 插入更多真实感的中大校园外卖数据
-- =============================================================================
-- 目标数据库：MySQL (mysql2.sqlpub.com:3307 / sysu_campus_food)
--
-- 本脚本会插入：
--   1. Merchants（商家）  — 8 家真实的中山大学校园店铺
--   2. BaseUsers          — 每个商家对应的后台账号（role='merchant'）
--   3. Dishes（菜品）     — 每家 8~15 道菜，共约 80+ 道
--   4. Meals（套餐）      — 部分商家配 2~4 个套餐
--   5. MealDishes（套餐-菜品关联）
--
-- 注意事项：
--   - dishes.dish_name 有 UNIQUE 约束，故使用 INSERT IGNORE
--   - 已有 category 数据（ID 1-15），直接引用
--   - 所有 SQL 语句均为标准 MySQL 语法
-- =============================================================================

-- =============================================================================
-- 1. 插入商家对应的 BaseUser（后台登录账号）
-- =============================================================================
-- 先检查是否已存在，避免重复
INSERT IGNORE INTO base_users (username, password, role, created_at, updated_at)
VALUES
  ('merchant_xueyi',   'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW()),
  ('merchant_mcd',     'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW()),
  ('merchant_luckin',  'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW()),
  ('merchant_yidiandian', 'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW()),
  ('merchant_lamian',  'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW()),
  ('merchant_yangguofu', 'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW()),
  ('merchant_baozifan', 'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW()),
  ('merchant_shaxian', 'e10adc3949ba59abbe56e057f20f883e', 'merchant', NOW(), NOW())
ON DUPLICATE KEY UPDATE username = VALUES(username);

-- =============================================================================
-- 2. 插入商家（Merchant）
-- =============================================================================
-- 先查询刚刚插入的 base_user ID
-- 注意：这里假设 base_user 的 ID 从某个值开始，我们直接用 SELECT 方式确定

-- 检查商家表中是否已有数据，避免重复
-- 注意：merchant 表名是 "merchants"

INSERT IGNORE INTO merchants (base_id, shop_name, shop_location, owner, phone, logo, license, status, menu_count, top_category1, top_category2, avg_score, score_count)
VALUES
  -- (1) 中山大学学一食堂
  ((SELECT id FROM base_users WHERE username = 'merchant_xueyi'),
   '中山大学学一食堂', '中山大学南校园学一食堂一楼',
   '张建国', '13800138001', '', '', 'open', 30, 1, 5, 4.5, 888),
  -- (2) 金拱门（中大店）
  ((SELECT id FROM base_users WHERE username = 'merchant_mcd'),
   '金拱门（中大店）', '中山大学南校园东门商业街A101',
   '李伟', '13800138002', '', '', 'open', 25, 2, 3, 4.3, 1256),
  -- (3) 瑞幸咖啡（中大店）
  ((SELECT id FROM base_users WHERE username = 'merchant_luckin'),
   '瑞幸咖啡（中大店）', '中山大学南校园图书馆负一层',
   '王芳', '13800138003', '', '', 'open', 20, 4, 13, 4.6, 2100),
  -- (4) 一点点（中大店）
  ((SELECT id FROM base_users WHERE username = 'merchant_yidiandian'),
   '一点点（中大店）', '中山大学南校园西区商业街B102',
   '陈晓', '13800138004', '', '', 'open', 18, 4, 13, 4.4, 1567),
  -- (5) 兰州拉面（中大店）
  ((SELECT id FROM base_users WHERE username = 'merchant_lamian'),
   '兰州拉面（中大店）', '中山大学南校园东门商业街B201',
   '马强', '13800138005', '', '', 'open', 15, 5, 1, 4.2, 890),
  -- (6) 杨国福麻辣烫（中大店）
  ((SELECT id FROM base_users WHERE username = 'merchant_yangguofu'),
   '杨国福麻辣烫（中大店）', '中山大学南校园西区商业街A105',
   '刘洋', '13800138006', '', '', 'open', 22, 6, 12, 4.5, 1345),
  -- (7) 煲仔饭（中大店）
  ((SELECT id FROM base_users WHERE username = 'merchant_baozifan'),
   '啫啫煲仔饭（中大店）', '中山大学南校园北门外食街3号',
   '黄明', '13800138007', '', '', 'open', 12, 5, 1, 4.1, 678),
  -- (8) 沙县小吃（中大店）
  ((SELECT id FROM base_users WHERE username = 'merchant_shaxian'),
   '沙县小吃（中大店）', '中山大学南校园东门商业街C103',
   '林峰', '13800138008', '', '', 'open', 16, 5, 1, 4.0, 567);

-- =============================================================================
-- 3. 插入菜品（Dish）
-- =============================================================================
-- 使用 INSERT IGNORE 避免重复（dish_name 有 UNIQUE 约束）
--
-- category 字段取值说明（已有分类 ID 1-15）：
--   1 = 主食  2 = 快餐  3 = 汉堡  4 = 饮品  5 = 小吃
--   6 = 麻辣烫  7 = 川菜  8 = 粤菜  9 = 日料  10 = 西餐
--   11 = 甜品  12 = 汤类  13 = 咖啡  14 = 水果  15 = 其他

-- ===== 商家 1：中山大学学一食堂 (merchant_id 需要动态获取，此处用子查询) =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('红烧肉套餐',   '18.00', '经典红烧肉 + 米饭 + 时蔬，肥而不腻，入口即化',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["招牌","套餐"]', NOW(), NOW()),
  ('番茄炒蛋饭',   '12.00', '家常番茄炒蛋盖浇饭，酸甜可口',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["经典","素食可选"]', NOW(), NOW()),
  ('鱼香肉丝饭',   '14.00', '经典川菜鱼香肉丝盖饭，微辣开胃',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["川味"]', NOW(), NOW()),
  ('糖醋里脊饭',   '15.00', '酥脆糖醋里脊配米饭，酸酸甜甜',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["酸甜","招牌"]', NOW(), NOW()),
  ('麻婆豆腐饭',   '11.00', '正宗麻婆豆腐盖饭，麻辣鲜香',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["麻辣","川味"]', NOW(), NOW()),
  ('清蒸鲈鱼套餐', '22.00', '新鲜清蒸鲈鱼 + 米饭 + 汤，鲜美嫩滑',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["招牌","套餐"]', NOW(), NOW()),
  ('宫保鸡丁饭',   '14.00', '经典宫保鸡丁盖饭，花生酥脆，鸡肉嫩滑',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["川味","经典"]', NOW(), NOW()),
  ('紫菜蛋花汤',   '4.00',  '清淡紫菜蛋花汤，餐前暖胃',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 12, 1, '["汤类"]', NOW(), NOW()),
  ('凉拌黄瓜',     '5.00',  '清爽凉拌黄瓜，夏天必点',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 5, 1, '["凉菜","素食"]', NOW(), NOW()),
  ('蒸水蛋',       '6.00',  '嫩滑蒸水蛋，适合清淡口味',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 5, 1, '["清淡","素食"]', NOW(), NOW()),
  ('青椒土豆丝',   '8.00',  '经典家常青椒土豆丝',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 5, 1, '["家常","素食"]', NOW(), NOW()),
  ('茶叶蛋',       '2.50',  '卤制茶叶蛋，入味鲜香',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 5, 1, '["早餐"]', NOW(), NOW()),
  ('煎蛋',         '2.00',  '现煎荷包蛋',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 5, 1, '["早餐"]', NOW(), NOW()),
  ('米饭',         '1.00',  '白米饭一份',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, '["主食"]', NOW(), NOW()),
  ('玉米排骨汤',   '8.00',  '甜玉米炖排骨汤，营养丰富',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 12, 1, '["汤类","营养"]', NOW(), NOW());

-- ===== 商家 2：金拱门（中大店） =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('巨无霸汉堡',       '25.00', '双层牛肉饼 + 特制酱料，经典招牌',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 3, 1, '["招牌","汉堡"]', NOW(), NOW()),
  ('麦辣鸡腿堡',       '20.00', '香辣鸡腿肉 + 新鲜生菜，辣味十足',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 3, 1, '["辣味","汉堡"]', NOW(), NOW()),
  ('板烧鸡腿堡',       '22.00', '铁板烧鸡腿肉，鲜嫩多汁',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 3, 1, '["汉堡","招牌"]', NOW(), NOW()),
  ('麦香鱼汉堡',       '18.00', '深海鳕鱼排搭配塔塔酱，鲜美清淡',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 3, 1, '["海鲜","汉堡"]', NOW(), NOW()),
  ('薯条（大）',       '12.00', '金黄酥脆大薯条',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 5, 1, '["小吃"]', NOW(), NOW()),
  ('麦乐鸡（6块）',    '10.00', '6块金黄麦乐鸡，配甜酸酱',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 5, 1, '["小吃","分享"]', NOW(), NOW()),
  ('可口可乐（大）',   '8.00',  '冰镇可口可乐大杯',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 4, 1, '["饮品"]', NOW(), NOW()),
  ('苹果派',           '7.00',  '酥脆苹果派，内馅香甜',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 11, 1, '["甜品"]', NOW(), NOW()),
  ('香芋派',           '7.00',  '香芋馅派，甜而不腻',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 11, 1, '["甜品"]', NOW(), NOW()),
  ('新地（草莓）',     '10.00', '草莓酱冰淇淋新地',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 11, 1, '["甜品","冰淇淋"]', NOW(), NOW()),
  ('玉米杯',           '8.00',  '甜玉米粒，健康小食',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 5, 1, '["健康"]', NOW(), NOW()),
  ('鸡翅（4块）',      '12.00', '4块香脆鸡翅，奥尔良风味',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 5, 1, '["小吃","鸡肉"]', NOW(), NOW());

-- ===== 商家 3：瑞幸咖啡（中大店） =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('生椰拿铁',         '19.90', '椰浆 + 浓缩咖啡，丝滑浓郁，经典爆款',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","爆款"]', NOW(), NOW()),
  ('厚乳拿铁',         '18.00', '厚牛乳 + 咖啡，醇厚顺滑',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","经典"]', NOW(), NOW()),
  ('陨石拿铁',         '19.00', '黑糖啵啵 + 拿铁，嚼得到的快乐',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","网红"]', NOW(), NOW()),
  ('美式咖啡',         '14.00', '经典美式，清爽提神',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","经典"]', NOW(), NOW()),
  ('澳瑞白',           '17.00', '澳白咖啡，奶泡细腻',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","精品"]', NOW(), NOW()),
  ('抹茶拿铁',         '16.00', '日式抹茶 + 牛乳，清新茶香',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["茶饮","咖啡"]', NOW(), NOW()),
  ('橙C美式',          '18.00', '鲜橙汁 + 美式咖啡，果味咖啡',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","水果"]', NOW(), NOW()),
  ('椰云拿铁',         '20.00', '椰云奶盖 + 咖啡，云朵般轻盈',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","限定"]', NOW(), NOW()),
  ('茉莉花香拿铁',     '17.00', '茉莉花茶 + 拿铁，花香四溢',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","花香"]', NOW(), NOW()),
  ('海盐芝士厚乳拿铁', '21.00', '海盐芝士奶盖 + 拿铁，咸甜交织',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 13, 1, '["咖啡","芝士"]', NOW(), NOW()),
  ('轻乳茶（茉莉）',   '14.00', '茉莉轻乳茶，清新不腻',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 4, 1, '["茶饮"]', NOW(), NOW()),
  ('小火山面包',       '8.00',  '酥皮面包，配咖啡绝佳',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 11, 1, '["甜品","面包"]', NOW(), NOW()),
  ('半熟芝士蛋糕',     '12.00', '轻芝士蛋糕，入口即化',
   (SELECT id FROM merchants WHERE shop_name = '瑞幸咖啡（中大店）'), '', 11, 1, '["甜品","芝士"]', NOW(), NOW());

-- ===== 商家 4：一点点（中大店） =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('四季春茶',         '8.00',  '清爽四季春茶，茶香悠长',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["茶饮","经典"]', NOW(), NOW()),
  ('波霸奶茶',         '12.00', 'Q弹波霸 + 经典奶茶，人气爆款',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["奶茶","爆款"]', NOW(), NOW()),
  ('珍珠奶茶',         '11.00', '经典珍珠奶茶，丝滑醇厚',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["奶茶","经典"]', NOW(), NOW()),
  ('四季春茶加珍波椰',  '13.00', '四季春茶 + 珍珠 + 波霸 + 椰果，满满好料',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["茶饮","加料"]', NOW(), NOW()),
  ('阿华田拿铁',       '15.00', '阿华田 + 拿铁，童年味道',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["奶茶","特色"]', NOW(), NOW()),
  ('冰淇淋红茶',       '13.00', '香草冰淇淋 + 红茶，冰爽一夏',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["限定","夏日"]', NOW(), NOW()),
  ('柠檬养乐多',       '12.00', '鲜榨柠檬 + 养乐多，酸甜开胃',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["饮品","水果"]', NOW(), NOW()),
  ('蜂蜜绿茶',         '9.00',  '蜂蜜 + 绿茶，清甜解暑',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["茶饮","清淡"]', NOW(), NOW()),
  ('椰果奶茶',         '12.00', '椰果 + 奶茶，嚼得到的快乐',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["奶茶","加料"]', NOW(), NOW()),
  ('芒果青',           '14.00', '芒果 + 青茶，热带果茶',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["水果茶"]', NOW(), NOW()),
  ('百香绿',           '13.00', '百香果 + 绿茶，酸甜清爽',
   (SELECT id FROM merchants WHERE shop_name = '一点点（中大店）'), '', 4, 1, '["水果茶"]', NOW(), NOW());

-- ===== 商家 5：兰州拉面（中大店） =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('兰州牛肉拉面',     '16.00', '正宗兰州牛肉拉面，汤清肉烂，宽细可选',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["面食","招牌"]', NOW(), NOW()),
  ('牛肉刀削面',       '17.00', '手工刀削面，筋道有嚼劲',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["面食","招牌"]', NOW(), NOW()),
  ('西红柿鸡蛋面',     '12.00', '番茄鸡蛋汤面，酸甜开胃',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["面食","清淡"]', NOW(), NOW()),
  ('炸酱面',           '14.00', '北京炸酱面，酱香浓郁',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["面食","经典"]', NOW(), NOW()),
  ('牛肉炒拉条',       '18.00', '拉条子面 + 牛肉翻炒，西北风味',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["面食","炒面"]', NOW(), NOW()),
  ('兰州炒饭',         '14.00', '兰州风味炒饭，配鸡蛋牛肉',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 1, 1, '["炒饭"]', NOW(), NOW()),
  ('凉皮',             '10.00', '陕西凉皮，酸辣爽口',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["凉菜","小吃"]', NOW(), NOW()),
  ('肉夹馍',           '8.00',  '白吉馍夹腊汁肉，西北经典',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["小吃","招牌"]', NOW(), NOW()),
  ('羊肉串（5串）',    '15.00', '炭烤羊肉串，孜然飘香',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["烧烤"]', NOW(), NOW()),
  ('鸡蛋牛肉炒面',     '16.00', '鸡蛋 + 牛肉 + 炒面，营养丰富',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["面食","炒面"]', NOW(), NOW()),
  ('牛肉水饺（12只）', '15.00', '手工牛肉水饺，皮薄馅大',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["饺子"]', NOW(), NOW()),
  ('小菜拼盘',         '6.00',  '凉拌黄瓜 + 海带丝 + 花生米',
   (SELECT id FROM merchants WHERE shop_name = '兰州拉面（中大店）'), '', 5, 1, '["凉菜"]', NOW(), NOW());

-- ===== 商家 6：杨国福麻辣烫（中大店） =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('经典麻辣烫（大）', '22.00', '自选菜品麻辣烫，骨汤底料，鲜辣过瘾',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["麻辣","招牌"]', NOW(), NOW()),
  ('经典麻辣烫（小）', '16.00', '自选菜品麻辣烫小份，一人食刚好',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["麻辣","小份"]', NOW(), NOW()),
  ('番茄麻辣烫',       '20.00', '番茄汤底麻辣烫，酸甜微辣',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["番茄","不辣"]', NOW(), NOW()),
  ('麻辣拌',           '18.00', '干拌麻辣烫，酱汁浓郁，无需汤底',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["干拌","招牌"]', NOW(), NOW()),
  ('肥牛卷（加料）',   '8.00',  '肥牛卷一份，麻辣烫加料首选',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["加料","肉"]', NOW(), NOW()),
  ('午餐肉（加料）',   '5.00',  '午餐肉一份',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["加料","肉"]', NOW(), NOW()),
  ('方便面（加料）',   '3.00',  '方便面饼一份，加在麻辣烫里超满足',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["加料","主食"]', NOW(), NOW()),
  ('金针菇（加料）',   '3.00',  '金针菇一份',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["加料","蔬菜"]', NOW(), NOW()),
  ('鹌鹑蛋（加料）',   '4.00',  '鹌鹑蛋一份',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 6, 1, '["加料","蛋"]', NOW(), NOW()),
  ('冰粉',             '6.00',  '红糖冰粉，解辣神器',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 11, 1, '["甜品","解辣"]', NOW(), NOW()),
  ('酸梅汤',           '5.00',  '冰镇酸梅汤，清爽解腻',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 4, 1, '["饮品","解腻"]', NOW(), NOW()),
  ('米饭',             '2.00',  '白米饭一份',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 1, 1, '["主食"]', NOW(), NOW());

-- ===== 商家 7：啫啫煲仔饭（中大店） =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('腊味煲仔饭',       '22.00', '广式腊肠 + 腊肉煲仔饭，锅巴金黄香脆',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 1, 1, '["煲仔饭","招牌"]', NOW(), NOW()),
  ('滑鸡煲仔饭',       '20.00', '鲜嫩滑鸡煲仔饭，姜葱提味',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 1, 1, '["煲仔饭","鸡肉"]', NOW(), NOW()),
  ('排骨煲仔饭',       '22.00', '豉汁排骨煲仔饭，排骨香嫩入味',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 1, 1, '["煲仔饭","排骨"]', NOW(), NOW()),
  ('牛肉煲仔饭',       '23.00', '嫩滑牛肉煲仔饭，窝蛋流心',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 1, 1, '["煲仔饭","牛肉"]', NOW(), NOW()),
  ('啫啫鸡煲',         '28.00', '啫啫鸡煲，砂锅高温快啫，酱香浓郁',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 8, 1, '["粤菜","啫啫"]', NOW(), NOW()),
  ('啫啫通菜',         '16.00', '虾酱啫通菜，经典粤式素菜',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 8, 1, '["粤菜","啫啫","素食"]', NOW(), NOW()),
  ('蚝油生菜',         '10.00', '白灼生菜淋蚝油，清淡鲜甜',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 5, 1, '["素食","清淡"]', NOW(), NOW()),
  ('例汤',             '5.00',  '每日例汤，随饭赠送价',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 12, 1, '["汤类"]', NOW(), NOW()),
  ('叉烧煲仔饭',       '21.00', '蜜汁叉烧煲仔饭，甜香可口',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 1, 1, '["煲仔饭","叉烧"]', NOW(), NOW()),
  ('白切鸡煲仔饭',     '22.00', '白切鸡煲仔饭，保留鸡肉原味',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 1, 1, '["煲仔饭","鸡肉","粤菜"]', NOW(), NOW()),
  ('咸蛋芥菜汤',       '6.00',  '咸蛋芥菜汤，下火解腻',
   (SELECT id FROM merchants WHERE shop_name = '啫啫煲仔饭（中大店）'), '', 12, 1, '["汤类"]', NOW(), NOW());

-- ===== 商家 8：沙县小吃（中大店） =====
INSERT IGNORE INTO dishes (dish_name, price, description, merchant_id, image_path, category, status, tags, created_at, updated_at)
VALUES
  ('蒸饺（10只）',     '8.00',  '沙县招牌蒸饺，皮薄馅鲜，蘸醋更佳',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["招牌","小吃"]', NOW(), NOW()),
  ('拌面',             '6.00',  '花生酱拌面，沙县经典，便宜又好吃',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["面食","招牌"]', NOW(), NOW()),
  ('扁肉（馄饨）',     '7.00',  '沙县扁肉（小馄饨），皮滑肉嫩',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["汤","招牌"]', NOW(), NOW()),
  ('炖罐（排骨）',     '10.00', '排骨炖罐，慢火炖制，营养滋补',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 12, 1, '["炖罐","营养"]', NOW(), NOW()),
  ('炖罐（乌鸡）',     '12.00', '乌鸡炖罐，滋补养颜',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 12, 1, '["炖罐","滋补"]', NOW(), NOW()),
  ('鸭腿饭',           '14.00', '卤鸭腿 + 米饭 + 配菜，实惠饱腹',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 1, 1, '["套餐","饭"]', NOW(), NOW()),
  ('鸡腿饭',           '13.00', '卤鸡腿饭，简单美味',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 1, 1, '["套餐","饭"]', NOW(), NOW()),
  ('卤蛋',             '2.00',  '卤制入味小卤蛋',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["小吃","卤味"]', NOW(), NOW()),
  ('豆干',             '2.00',  '五香豆干，实惠小吃',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["小吃","素食"]', NOW(), NOW()),
  ('茶叶蛋',           '2.00',  '卤制茶叶蛋',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["小吃","早餐"]', NOW(), NOW()),
  ('炒米粉',           '10.00', '福建炒米粉，配鸡蛋蔬菜',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["主食","炒粉"]', NOW(), NOW()),
  ('猪心炖罐',         '11.00', '猪心炖罐，安神补气',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 12, 1, '["炖罐","滋补"]', NOW(), NOW()),
  ('花生酱拌面加蛋', '8.00', '拌面加煎蛋，更饱腹',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 5, 1, '["面食"]', NOW(), NOW());

-- =============================================================================
-- 4. 插入套餐（Meal）
-- =============================================================================
-- 为部分商家添加套餐，关联到 meal 和 meal_dish 表

-- ===== 金拱门套餐 =====
INSERT IGNORE INTO meals (mealname, price, description, merchant_id, image_path, status, category, created_at, updated_at)
VALUES
  ('超值午餐A', '29.00', '巨无霸汉堡 + 薯条（中）+ 可乐（中）',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 1, 2, NOW(), NOW()),
  ('超值午餐B', '27.00', '麦辣鸡腿堡 + 薯条（中）+ 可乐（中）',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 1, 2, NOW(), NOW()),
  ('快乐儿童餐', '22.00', '麦香鱼汉堡 + 小薯条 + 小可乐 + 玩具',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 1, 2, NOW(), NOW()),
  ('双人分享餐', '49.00', '2个巨无霸 + 大薯条 + 2杯中可乐 + 4块麦乐鸡',
   (SELECT id FROM merchants WHERE shop_name = '金拱门（中大店）'), '', 1, 2, NOW(), NOW());

-- ===== 学一食堂套餐 =====
INSERT IGNORE INTO meals (mealname, price, description, merchant_id, image_path, status, category, created_at, updated_at)
VALUES
  ('超值午餐', '20.00', '任选主菜（红烧肉/鱼香肉丝/宫保鸡丁）+ 时蔬 + 米饭 + 汤',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, NOW(), NOW()),
  ('营养套餐', '18.00', '番茄炒蛋 + 青菜 + 蒸水蛋 + 米饭',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, NOW(), NOW()),
  ('丰盛晚餐', '25.00', '清蒸鲈鱼 + 凉拌黄瓜 + 米饭 + 玉米排骨汤',
   (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂'), '', 1, 1, NOW(), NOW());

-- ===== 沙县小吃套餐 =====
INSERT IGNORE INTO meals (mealname, price, description, merchant_id, image_path, status, category, created_at, updated_at)
VALUES
  ('沙县经典套餐', '15.00', '蒸饺 + 拌面 + 炖罐（排骨）',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 1, 5, NOW(), NOW()),
  ('沙县超值套餐', '18.00', '鸭腿饭 + 蒸饺 + 卤蛋',
   (SELECT id FROM merchants WHERE shop_name = '沙县小吃（中大店）'), '', 1, 1, NOW(), NOW());

-- ===== 杨国福麻辣烫套餐 =====
INSERT IGNORE INTO meals (mealname, price, description, merchant_id, image_path, status, category, created_at, updated_at)
VALUES
  ('单人麻辣烫套餐', '25.00', '经典麻辣烫 + 肥牛卷 + 冰粉 + 米饭',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 1, 6, NOW(), NOW()),
  ('双人麻辣烫套餐', '42.00', '两份经典麻辣烫 + 肥牛卷 + 金针菇 + 酸梅汤2杯',
   (SELECT id FROM merchants WHERE shop_name = '杨国福麻辣烫（中大店）'), '', 1, 6, NOW(), NOW());

-- =============================================================================
-- 5. 插入套餐-菜品关联（MealDish）
-- =============================================================================
-- 只有设置了套餐的商家才需要关联

-- 金拱门 超值午餐A: 巨无霸汉堡 + 薯条 + 可乐
INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐A' AND d.dish_name = '巨无霸汉堡';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐A' AND d.dish_name = '薯条（大）';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐A' AND d.dish_name = '可口可乐（大）';

-- 金拱门 超值午餐B: 麦辣鸡腿堡 + 薯条 + 可乐
INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐B' AND d.dish_name = '麦辣鸡腿堡';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐B' AND d.dish_name = '薯条（大）';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐B' AND d.dish_name = '可口可乐（大）';

-- 金拱门 快乐儿童餐: 麦香鱼汉堡 + 薯条 + 可乐
INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '快乐儿童餐' AND d.dish_name = '麦香鱼汉堡';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '快乐儿童餐' AND d.dish_name = '薯条（大）';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '快乐儿童餐' AND d.dish_name = '可口可乐（大）';

-- 学一食堂 超值午餐: 红烧肉套餐（主菜）+ 米饭 + 紫菜蛋花汤
INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐' AND m.merchant_id = (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂')
  AND d.dish_name = '红烧肉套餐';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐' AND m.merchant_id = (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂')
  AND d.dish_name = '米饭';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '超值午餐' AND m.merchant_id = (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂')
  AND d.dish_name = '紫菜蛋花汤';

-- 学一食堂 营养套餐: 番茄炒蛋饭 + 蒸水蛋 + 凉拌黄瓜
INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '营养套餐' AND m.merchant_id = (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂')
  AND d.dish_name = '番茄炒蛋饭';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '营养套餐' AND m.merchant_id = (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂')
  AND d.dish_name = '蒸水蛋';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '营养套餐' AND m.merchant_id = (SELECT id FROM merchants WHERE shop_name = '中山大学学一食堂')
  AND d.dish_name = '凉拌黄瓜';

-- 沙县经典套餐: 蒸饺 + 拌面 + 炖罐（排骨）
INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '沙县经典套餐' AND d.dish_name = '蒸饺（10只）';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '沙县经典套餐' AND d.dish_name = '拌面';

INSERT IGNORE INTO meal_dishes (meal_id, dish_id, num)
SELECT m.id, d.id, 1
FROM meals m, dishes d
WHERE m.mealname = '沙县经典套餐' AND d.dish_name = '炖罐（排骨）';

-- =============================================================================
-- 完成！可以运行以下命令执行本脚本：
-- mysql -h mysql2.sqlpub.com -P 3307 -u yjxnbhh -p sysu_campus_food < seed_more_data.sql
-- =============================================================================
