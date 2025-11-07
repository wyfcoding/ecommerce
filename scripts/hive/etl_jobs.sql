-- Hive ETL 任务脚本

USE ecommerce_dw;

-- ==================== ODS 层数据导入 ====================

-- 从 MySQL 导入订单数据到 ODS 层
INSERT OVERWRITE TABLE ods_order_fact PARTITION(dt='${hiveconf:dt}')
SELECT 
    id as order_id,
    order_no as order_sn,
    user_id,
    total_amount,
    actual_amount as payment_amount,
    shipping_fee,
    discount_amount,
    status,
    payment_method,
    created_at,
    paid_at,
    shipped_at,
    completed_at
FROM mysql_orders
WHERE DATE(created_at) = '${hiveconf:dt}';

-- ==================== DWD 层数据加工 ====================

-- 构建订单宽表
INSERT OVERWRITE TABLE dwd_order_wide PARTITION(dt='${hiveconf:dt}')
SELECT 
    o.order_id,
    o.order_sn,
    o.user_id,
    u.name as user_name,
    u.city as user_city,
    u.province as user_province,
    oi.product_id,
    p.product_name,
    p.category_id,
    c.name as category_name,
    p.brand_id,
    b.name as brand_name,
    oi.sku_id,
    oi.price,
    oi.quantity,
    oi.total_price,
    o.total_amount as order_total_amount,
    o.payment_amount,
    o.discount_amount,
    o.shipping_fee,
    o.payment_method,
    o.status as order_status,
    o.created_at,
    o.paid_at,
    o.shipped_at,
    o.completed_at
FROM ods_order_fact o
LEFT JOIN ods_order_item oi ON o.order_id = oi.order_id AND oi.dt = '${hiveconf:dt}'
LEFT JOIN ods_user u ON o.user_id = u.user_id AND u.dt = '${hiveconf:dt}'
LEFT JOIN ods_product p ON oi.product_id = p.product_id AND p.dt = '${hiveconf:dt}'
LEFT JOIN ods_category c ON p.category_id = c.category_id
LEFT JOIN ods_brand b ON p.brand_id = b.brand_id
WHERE o.dt = '${hiveconf:dt}';

-- ==================== DWS 层数据聚合 ====================

-- 用户购买汇总（天）
INSERT OVERWRITE TABLE dws_user_purchase_day PARTITION(dt='${hiveconf:dt}')
SELECT 
    user_id,
    COUNT(DISTINCT order_id) as order_count,
    SUM(order_total_amount) as order_amount,
    COUNT(DISTINCT CASE WHEN order_status = 'PAID' THEN order_id END) as payment_count,
    SUM(CASE WHEN order_status = 'PAID' THEN payment_amount ELSE 0 END) as payment_amount,
    COUNT(DISTINCT CASE WHEN order_status = 'REFUNDED' THEN order_id END) as refund_count,
    SUM(CASE WHEN order_status = 'REFUNDED' THEN payment_amount ELSE 0 END) as refund_amount,
    AVG(CASE WHEN order_status = 'PAID' THEN payment_amount END) as avg_order_value
FROM dwd_order_wide
WHERE dt = '${hiveconf:dt}'
GROUP BY user_id;

-- 商品销售汇总（天）
INSERT OVERWRITE TABLE dws_product_sales_day PARTITION(dt='${hiveconf:dt}')
SELECT 
    product_id,
    MAX(product_name) as product_name,
    MAX(category_id) as category_id,
    MAX(category_name) as category_name,
    MAX(brand_id) as brand_id,
    MAX(brand_name) as brand_name,
    SUM(quantity) as sales_count,
    SUM(total_price) as sales_amount,
    COUNT(DISTINCT order_id) as order_count,
    COUNT(DISTINCT user_id) as buyer_count,
    SUM(CASE WHEN order_status = 'REFUNDED' THEN quantity ELSE 0 END) as refund_count,
    SUM(CASE WHEN order_status = 'REFUNDED' THEN total_price ELSE 0 END) as refund_amount
FROM dwd_order_wide
WHERE dt = '${hiveconf:dt}'
  AND order_status IN ('PAID', 'SHIPPED', 'COMPLETED', 'REFUNDED')
GROUP BY product_id;

-- 分类销售汇总（天）
INSERT OVERWRITE TABLE dws_category_sales_day PARTITION(dt='${hiveconf:dt}')
SELECT 
    category_id,
    MAX(category_name) as category_name,
    SUM(quantity) as sales_count,
    SUM(total_price) as sales_amount,
    COUNT(DISTINCT order_id) as order_count,
    COUNT(DISTINCT user_id) as buyer_count
FROM dwd_order_wide
WHERE dt = '${hiveconf:dt}'
  AND order_status IN ('PAID', 'SHIPPED', 'COMPLETED')
GROUP BY category_id;

-- ==================== ADS 层报表生成 ====================

-- 每日销售报表
INSERT OVERWRITE TABLE ads_daily_sales_report
SELECT 
    '${hiveconf:dt}' as report_date,
    COUNT(DISTINCT order_id) as total_orders,
    SUM(payment_amount) as total_amount,
    COUNT(DISTINCT user_id) as total_users,
    COUNT(DISTINCT CASE WHEN is_new_user = 1 THEN user_id END) as new_users,
    COUNT(DISTINCT user_id) as active_users,
    AVG(payment_amount) as avg_order_value,
    COUNT(DISTINCT CASE WHEN order_status = 'PAID' THEN order_id END) * 100.0 / 
        COUNT(DISTINCT order_id) as conversion_rate,
    SUM(payment_amount) as gmv
FROM dwd_order_wide
WHERE dt = '${hiveconf:dt}';

-- 商品销售排行榜（Top 100）
INSERT OVERWRITE TABLE ads_product_sales_topn
SELECT 
    ROW_NUMBER() OVER (ORDER BY sales_amount DESC) as rank,
    product_id,
    product_name,
    category_name,
    brand_name,
    sales_count,
    sales_amount,
    '${hiveconf:dt}' as report_date
FROM dws_product_sales_day
WHERE dt = '${hiveconf:dt}'
ORDER BY sales_amount DESC
LIMIT 100;

-- 用户价值分层
INSERT OVERWRITE TABLE ads_user_value_segment
SELECT 
    user_id,
    MAX(user_name) as user_name,
    SUM(order_count) as total_orders,
    SUM(payment_amount) as total_amount,
    AVG(avg_order_value) as avg_order_value,
    MAX(dt) as last_order_date,
    CASE 
        WHEN SUM(payment_amount) >= 10000 THEN 'VIP'
        WHEN SUM(payment_amount) >= 5000 THEN '高价值'
        WHEN SUM(payment_amount) >= 1000 THEN '中价值'
        WHEN SUM(payment_amount) >= 100 THEN '低价值'
        ELSE '流失'
    END as user_segment,
    '${hiveconf:dt}' as report_date
FROM dws_user_purchase_day
WHERE dt >= DATE_SUB('${hiveconf:dt}', 90)
GROUP BY user_id;
