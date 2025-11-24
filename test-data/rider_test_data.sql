-- 骑手端测试数据
-- 包含骑手信息、订单数据、钱包信息、收入记录等

-- 清理现有数据（谨慎使用）
-- DELETE FROM rider_income_records;
-- DELETE FROM rider_wallets;
-- DELETE FROM rider_profiles;
-- DELETE FROM riders;
-- DELETE FROM orders WHERE rider_id IS NOT NULL;

-- 1. 骑手基本信息表 (riders)
INSERT INTO riders (id, base_id, realname, idnumber, phone) VALUES
(1, 1001, '李明', '440106199001011234', '13800138001'),
(2, 1002, '王强', '440106199002022345', '13800138002'),
(3, 1003, '张伟', '440106199003033456', '13800138003'),
(4, 1004, '刘洋', '440106199004044567', '13800138004'),
(5, 1005, '陈静', '440106199005055678', '13800138005');

-- 2. 骑手详细档案表 (rider_profiles)
INSERT INTO rider_profiles (user_id, rider_id, name, avatar, phone, rating, completed_orders, is_online, latitude, longitude, address, online_hours, created_at, updated_at) VALUES
(1001, 1, '李骑手', 'https://example.com/avatar1.jpg', '13800138001', 4.8, 1250, true, 22.3689, 113.5432, '中山大学珠海校区榕园宿舍', 156.5, NOW(), NOW()),
(1002, 2, '王骑手', 'https://example.com/avatar2.jpg', '13800138002', 4.6, 980, false, 22.3701, 113.5456, '中山大学珠海校区荔园宿舍', 142.3, NOW(), NOW()),
(1003, 3, '张骑手', 'https://example.com/avatar3.jpg', '13800138003', 4.9, 1560, true, 22.3655, 113.5410, '中山大学珠海校区榕园食堂', 189.7, NOW(), NOW()),
(1004, 4, '刘骑手', 'https://example.com/avatar4.jpg', '13800138004', 4.7, 750, true, 22.3722, 113.5488, '中山大学珠海校区翰林宿舍', 98.2, NOW(), NOW()),
(1005, 5, '陈骑手', 'https://example.com/avatar5.jpg', '13800138005', 4.5, 620, false, 22.3678, 113.5399, '中山大学珠海校区岁月湖', 76.8, NOW(), NOW());

-- 3. 骑手钱包表 (rider_wallets)
INSERT INTO rider_wallets (rider_id, balance, frozen_amount, total_income) VALUES
(1, 2580.50, 120.00, 15680.50),
(2, 1825.30, 80.50, 12250.30),
(3, 3456.80, 200.00, 18920.80),
(4, 1567.90, 0.00, 9870.90),
(5, 890.45, 45.00, 6540.45);

-- 4. 订单表 (orders) - 包含不同状态的订单供测试
INSERT INTO orders (consigneeid, phone, consignee, address, pickuppoint, dropofpoint, expectedtime, status, totalprice, merchantid, notes, numberoftableware, rider_id, pickup_code, accepted_at, pickup_at, deliver_at, finish_at, created_at, updated_at) VALUES

-- 待接单订单 (status = 1)
-- 注意：这些订单没有rider_id，表示新订单等待骑手接单
(1, '13666666666', '张同学', '中山大学珠海校区榕园宿舍A栋', NOW() + INTERVAL 10 MINUTE, NOW() + INTERVAL 40 MINUTE, NOW() + INTERVAL 50 MINUTE, 1, 35.50, 1, '不要辣', 1, NULL, NULL, NULL, NULL, NULL, NULL, NOW(), NOW()),
(2, '13777777777', '王同学', '中山大学珠海校区荔园宿舍B栋', NOW() + INTERVAL 5 MINUTE, NOW() + INTERVAL 35 MINUTE, NOW() + INTERVAL 45 MINUTE, 1, 42.80, 2, '多加香菜', 0, NULL, NULL, NULL, NULL, NULL, NULL, NOW(), NOW()),
(3, '13888888888', '李同学', '中山大学珠海校区翰林宿舍C栋', NOW() + INTERVAL 15 MINUTE, NOW() + INTERVAL 45 MINUTE, NOW() + INTERVAL 55 MINUTE, 1, 28.90, 3, '麻烦快一点', 2, NULL, NULL, NULL, NULL, NULL, NULL, NOW(), NOW()),

-- 已接单待取货 (status = 2)
(4, '13999999999', '陈同学', '中山大学珠海校区榕园宿舍D栋', NOW() - INTERVAL 5 MINUTE, NOW() + INTERVAL 25 MINUTE, NOW() + INTERVAL 35 MINUTE, 2, 51.20, 1, '加急', 1, 1, 'A123', NOW() - INTERVAL 3 MINUTE, NULL, NULL, NULL, NOW() - INTERVAL 10 MINUTE, NOW()),
(5, '13111111111', '赵同学', '中山大学珠海校区荔园宿舍E栋', NOW() - INTERVAL 8 MINUTE, NOW() + INTERVAL 22 MINUTE, NOW() + INTERVAL 32 MINUTE, 2, 38.70, 2, '少油', 0, 3, 'B456', NOW() - INTERVAL 6 MINUTE, NULL, NULL, NULL, NOW() - INTERVAL 15 MINUTE, NOW()),

-- 取货中 (status = 3)
(6, '13222222222', '孙同学', '中山大学珠海校区翰林宿舍F栋', NOW() - INTERVAL 15 MINUTE, NOW() + INTERVAL 15 MINUTE, NOW() + INTERVAL 25 MINUTE, 3, 45.60, 3, '正常配送', 1, 1, 'C789', NOW() - INTERVAL 12 MINUTE, NOW() - INTERVAL 2 MINUTE, NULL, NULL, NOW() - INTERVAL 20 MINUTE, NOW()),

-- 配送中 (status = 4)
(7, '13333333333', '周同学', '中山大学珠海校区榕园宿舍G栋', NOW() - INTERVAL 25 MINUTE, NOW() - INTERVAL 5 MINUTE, NOW() + INTERVAL 5 MINUTE, 4, 33.80, 1, '小心轻放', 2, 2, 'D012', NOW() - INTERVAL 22 MINUTE, NOW() - INTERVAL 12 MINUTE, NOW() - INTERVAL 2 MINUTE, NULL, NOW() - INTERVAL 30 MINUTE, NOW()),
(8, '13444444444', '吴同学', '中山大学珠海校区荔园宿舍H栋', NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 10 MINUTE, NOW(), 4, 29.90, 2, '放门口即可', 0, 3, 'E345', NOW() - INTERVAL 27 MINUTE, NOW() - INTERVAL 17 MINUTE, NOW() - INTERVAL 7 MINUTE, NULL, NOW() - INTERVAL 35 MINUTE, NOW()),

-- 已完成 (status = 5)
(9, '13555555555', '郑同学', '中山大学珠海校区翰林宿舍I栋', NOW() - INTERVAL 60 MINUTE, NOW() - INTERVAL 40 MINUTE, NOW() - INTERVAL 30 MINUTE, 5, 58.40, 3, '配送及时', 1, 1, 'F678', NOW() - INTERVAL 57 MINUTE, NOW() - INTERVAL 47 MINUTE, NOW() - INTERVAL 37 MINUTE, NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 70 MINUTE, NOW() - INTERVAL 30 MINUTE),
(10, '13666666667', '黄同学', '中山大学珠海校区榕园宿舍J栋', NOW() - INTERVAL 120 MINUTE, NOW() - INTERVAL 100 MINUTE, NOW() - INTERVAL 90 MINUTE, 5, 41.20, 1, '感谢', 1, 2, 'G901', NOW() - INTERVAL 117 MINUTE, NOW() - INTERVAL 107 MINUTE, NOW() - INTERVAL 97 MINUTE, NOW() - INTERVAL 90 MINUTE, NOW() - INTERVAL 130 MINUTE, NOW() - INTERVAL 90 MINUTE);

-- 5. 收入记录表 (rider_income_records)
INSERT INTO rider_income_records (rider_id, order_id, amount, type, remark, created_at) VALUES
-- 李骑手的收入记录
(1, 9, 6.50, 'order', '订单配送费', NOW() - INTERVAL 30 MINUTE),
(1, 10, 7.00, 'order', '订单配送费', NOW() - INTERVAL 90 MINUTE),
(1, NULL, 2.00, 'bonus', '准时配送奖励', NOW() - INTERVAL 45 MINUTE),
(1, NULL, 1.50, 'bonus', '好评奖励', NOW() - INTERVAL 120 MINUTE),

-- 王骑手的收入记录
(2, 7, 5.80, 'order', '订单配送费', NOW() - INTERVAL 5 MINUTE),
(2, NULL, 2.50, 'bonus', '雨天配送奖励', NOW() - INTERVAL 30 MINUTE),
(2, NULL, -1.00, 'adjustment', '订单调整扣款', NOW() - INTERVAL 60 MINUTE),

-- 张骑手的收入记录
(3, 8, 5.20, 'order', '订单配送费', NOW() - INTERVAL 7 MINUTE),
(3, 6, 6.80, 'order', '订单配送费', NOW() - INTERVAL 15 MINUTE),
(3, NULL, 3.00, 'bonus', '高峰时段奖励', NOW() - INTERVAL 45 MINUTE),
(3, NULL, 1.00, 'bonus', '新人奖励', NOW() - INTERVAL 180 MINUTE),

-- 刘骑手的收入记录
(4, 5, 5.50, 'order', '订单配送费', NOW() - INTERVAL 2 MINUTE),
(4, NULL, 2.00, 'bonus', '夜间配送奖励', NOW() - INTERVAL 20 MINUTE),

-- 陈骑手的收入记录
(5, 4, 7.20, 'order', '订单配送费', NOW() - INTERVAL 5 MINUTE),
(5, NULL, 1.50, 'bonus', '优质服务奖励', NOW() - INTERVAL 40 MINUTE);

-- 6. 额外创建一些历史订单数据用于分页测试
INSERT INTO orders (consigneeid, phone, consignee, address, pickuppoint, dropofpoint, expectedtime, status, totalprice, merchantid, notes, numberoftableware, rider_id, pickup_code, accepted_at, pickup_at, deliver_at, finish_at, created_at, updated_at) VALUES

-- 更多历史已完成订单（用于分页测试）
(7, '13777777778', '钱同学', '中山大学珠海校区榕园宿舍K栋', NOW() - INTERVAL 180 MINUTE, NOW() - INTERVAL 160 MINUTE, NOW() - INTERVAL 150 MINUTE, 5, 32.50, 2, '谢谢', 1, 1, 'H234', NOW() - INTERVAL 177 MINUTE, NOW() - INTERVAL 167 MINUTE, NOW() - INTERVAL 157 MINUTE, NOW() - INTERVAL 150 MINUTE, NOW() - INTERVAL 190 MINUTE, NOW() - INTERVAL 150 MINUTE),
(8, '13888888889', '冯同学', '中山大学珠海校区荔园宿舍L栋', NOW() - INTERVAL 240 MINUTE, NOW() - INTERVAL 220 MINUTE, NOW() - INTERVAL 210 MINUTE, 5, 48.90, 3, '很满意', 2, 2, 'I567', NOW() - INTERVAL 237 MINUTE, NOW() - INTERVAL 227 MINUTE, NOW() - INTERVAL 217 MINUTE, NOW() - INTERVAL 210 MINUTE, NOW() - INTERVAL 250 MINUTE, NOW() - INTERVAL 210 MINUTE),
(9, '13999999998', '卫同学', '中山大学珠海校区翰林宿舍M栋', NOW() - INTERVAL 300 MINUTE, NOW() - INTERVAL 280 MINUTE, NOW() - INTERVAL 270 MINUTE, 5, 26.80, 1, '配送很快', 0, 3, 'J890', NOW() - INTERVAL 297 MINUTE, NOW() - INTERVAL 287 MINUTE, NOW() - INTERVAL 277 MINUTE, NOW() - INTERVAL 270 MINUTE, NOW() - INTERVAL 310 MINUTE, NOW() - INTERVAL 270 MINUTE),
(10, '13111111112', '蒋同学', '中山大学珠海校区榕园宿舍N栋', NOW() - INTERVAL 360 MINUTE, NOW() - INTERVAL 340 MINUTE, NOW() - INTERVAL 330 MINUTE, 5, 55.30, 2, '辛苦了', 1, 1, 'K123', NOW() - INTERVAL 357 MINUTE, NOW() - INTERVAL 347 MINUTE, NOW() - INTERVAL 337 MINUTE, NOW() - INTERVAL 330 MINUTE, NOW() - INTERVAL 370 MINUTE, NOW() - INTERVAL 330 MINUTE);