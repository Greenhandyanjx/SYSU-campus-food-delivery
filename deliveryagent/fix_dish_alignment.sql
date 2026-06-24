-- ========================================
-- 菜品-店铺对齐修正脚本
-- dishes.merchant_id = merchants.base_id
-- 店铺 6-13 (base_id=68-75) 的菜品数据错位
-- ========================================

-- 1. 删除麦当劳（福乐购百货店, base=4）的测试垃圾数据
DELETE FROM dishes WHERE merchant_id = 4 AND id IN (
  3,4,5,6,8,9,10,11,12,15,16,18,19,20,32,33,35,36,37,38,39,71,73,74
);

-- 2. 删除竹林上禾便当 (base=10) 的单个测试菜品: 爆虫 (id=21)
DELETE FROM dishes WHERE id = 21;

-- 3. 删除一点点（中大店, base=71）的测试垃圾数据
DELETE FROM dishes WHERE merchant_id = 71 AND id IN (22,27,28);
-- 保留 base=71 的有意义数据: 猪肝粥(23), 蛋挞(24), 酱烤一口臭豆腐(63), 新疆碳烤牛肉(64), 碳烤鱿鱼(65), 扇贝串(66), 碳烤鸡心(67)

-- 4. 将港都热炒（base=42）的菜品从 merchant_id=42 对齐（已经正确，无需操作）
-- 5. 将豪奶(id=29)保留在港都热炒

-- ========================================
-- 店铺 6-13 的菜品旋转修正
-- 当前数据: base=68有奶茶, 69有牛肉面, 70有麻辣烫, 71有烧烤/垃圾
--           72有鸡排饭, 73有家常菜, 74有汉堡+煲仔饭, 75有咖啡+沙县
-- 修正后:   68←家常菜+鸡排饭(学一食堂), 69←汉堡(金拱门)
--           70←咖啡(瑞幸), 71←奶茶(一点点)
--           72←牛肉面(兰州拉面), 73←麻辣烫(杨国福)
--           74←煲仔饭(啫啫煲仔饭), 75←沙县小吃
-- ========================================

-- 6. 奶茶饮品: base=68 → base=71 (一点点)
UPDATE dishes SET merchant_id = 71 WHERE merchant_id = 68;

-- 7. 兰州牛肉面等: base=69 → base=72 (兰州拉面)
UPDATE dishes SET merchant_id = 72 WHERE merchant_id = 69;

-- 8. 经典麻辣烫等: base=70 → base=73 (杨国福麻辣烫)
UPDATE dishes SET merchant_id = 73 WHERE merchant_id = 70;

-- 9. 烧烤类: 从 base=71 移到 base=9 (夯肉先生炭烤店)
--    (酱烤一口臭豆腐, 新疆碳烤牛肉, 碳烤鱿鱼, 扇贝串, 碳烤鸡心)
UPDATE dishes SET merchant_id = 9 WHERE id IN (63,64,65,66,67);

-- 10. 保留一点点的剩余菜品: 猪肝粥(23), 蛋挞(24)
--     这些保留在 base=71，是一点点的现有商品

-- 11. 鸡排饭等: base=72 → base=68 (中山大学学一食堂)
UPDATE dishes SET merchant_id = 68 WHERE merchant_id = 72;

-- 12. 家常菜: base=73 → base=68 (中山大学学一食堂)
UPDATE dishes SET merchant_id = 68 WHERE merchant_id = 73;

-- 13. 麦当劳汉堡类: 从 base=74 移到 base=4 (麦当劳福乐购百货店)
--     (巨无霸汉堡, 麦辣鸡腿堡, 麦香鱼汉堡, 薯条, 可乐, 麦乐鸡, 新地, 苹果派)
UPDATE dishes SET merchant_id = 4 WHERE merchant_id = 74 AND id IN (85,86,87,88,89,90,91,92);

-- 14. 煲仔饭: 保留在 base=74 (啫啫煲仔饭)
--     (腊味煲仔饭, 排骨煲仔饭, 滑鸡煲仔饭, 窝蛋牛肉饭, 煲仔饭加窝蛋)
--     这些已经在 base=74，无需操作

-- 15. 咖啡类: 从 base=75 移到 base=70 (瑞幸咖啡)
--     (厚乳拿铁, 标准美式, 陨石拿铁, 抹茶拿铁, 杨梅瑞纳冰)
UPDATE dishes SET merchant_id = 70 WHERE merchant_id = 75 AND id IN (94,95,96,97,98);

-- 16. 沙县小吃: 保留在 base=75 (蒸饺, 拌面, 扁肉, 鸭腿饭, 炖罐, 卤蛋, 花生酱拌面)
--     这些已经在 base=75，无需操作

-- 17. 处理孤儿菜品: id=1 (潮汕牛肉丸, merchant_id=0)
--     分配到 夯肉先生炭烤店 (base=9) 的烧烤类
UPDATE dishes SET merchant_id = 9 WHERE id = 1;

-- 18. 删除金拱门（中大店）当前剩余的凉皮和肉夹馍（如果还保留了不匹配的）
--     实际上金拱门的菜品已经在第7步移走了，现在金拱门没有菜品
--     我们需要给金拱门分配一些商品
--     从 base=74 分一些汉堡给 金拱门(base=69):
--     (巨无霸汉堡, 麦辣鸡腿堡)
INSERT INTO dishes (merchant_id, dish_name, price, image_path, category, status)
VALUES
(69, '巨无霸汉堡', '25.00', 'http://localhost:3000/images/meals/dish_mcd_big_mac.png', 3, 1),
(69, '麦辣鸡腿堡', '22.00', 'http://localhost:3000/images/meals/dish_mcd_spicy_chicken.png', 3, 1),
(69, '麦香鱼汉堡', '20.00', 'http://localhost:3000/images/meals/dish_mcd_fish.png', 3, 1),
(69, '薯条（大）', '12.00', 'http://localhost:3000/images/meals/dish_mcd_fries.png', 5, 1),
(69, '可乐（大）', '8.00', 'http://localhost:3000/images/meals/dish_mcd_cola.png', 4, 1),
(69, '麦乐鸡（6块）', '14.00', 'http://localhost:3000/images/meals/dish_mcd_nuggets.png', 5, 1),
(69, '苹果派', '8.00', 'http://localhost:3000/images/meals/dish_mcd_pie.png', 5, 1);
