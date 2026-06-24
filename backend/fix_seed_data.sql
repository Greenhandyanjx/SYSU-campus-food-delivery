-- =============================================================================
-- fix_seed_data.sql —— 修复 seed_more_data.sql 的 bug
-- =============================================================================
-- Bug: dishes.merchant_id 使用了 merchants.id（自增主键），但 Go 后端
--      GetStores() 查菜品时用的是 merchants.base_id
--
-- 修复步骤：
--   1. 删除 seed_more_data 中插入的旧 dishes（其 merchant_id 存的是 merchants.id）
--   2. 删除旧 meals（同样问题）
--   3. 删除 old meal_dishes
--   4. 删除旧 merchants（重新插入正确的 base_id）
--   5. 重新运行 seed_more_data.sql
--
-- 注意：只删除 seed_more_data.sql 中新增的 8 个商家及其关联数据
--       不会影响系统中原有的商家数据
-- =============================================================================

-- 找出我们新增的商家 ID 列表（通过 shop_name）
SET @shop_names = '("中山大学学一食堂","金拱门（中大店）","瑞幸咖啡（中大店）","一点点（中大店）","兰州拉面（中大店）","杨国福麻辣烫（中大店）","啫啫煲仔饭（中大店）","沙县小吃（中大店）")';

-- 步骤 1: 删除 meal_dishes 关联（通过 meals 所属商家）
DELETE FROM meal_dishes WHERE meal_id IN (
  SELECT id FROM meals 
  WHERE merchant_id IN (SELECT id FROM merchants WHERE shop_name IN @shop_names)
);

-- 步骤 2: 删除 meals
DELETE FROM meals 
WHERE merchant_id IN (SELECT id FROM merchants WHERE shop_name IN @shop_names);

-- 步骤 3: 删除 dishes（dish_name 有 UNIQUE 约束，但我们是这些菜品的创建者）
DELETE FROM dishes 
WHERE merchant_id IN (SELECT id FROM merchants WHERE shop_name IN @shop_names);

-- 步骤 4: 删除 merchants
DELETE FROM merchants WHERE shop_name IN @shop_names;

-- 步骤 5: 删除 base_users（可选，保留也可）
-- DELETE FROM base_users WHERE username LIKE 'merchant_%';

-- =============================================================================
-- 现在重新运行 seed_more_data.sql
-- =============================================================================
-- 注意：重新执行 seed_more_data.sql 后，所有 dishes 的 merchant_id
--       会正确使用 merchants.base_id
-- =============================================================================
