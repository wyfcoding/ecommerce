-- E-commerce 数据仓库建表脚本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS ecommerce_dw
COMMENT 'E-commerce Data Warehouse'
LOCATION '/user/hive/warehouse/ecommerce_dw.db';

USE ecommerce_dw;

-- ==================== ODS 层（操作数据存储）====================

-- 订单事实表
CREATE EXTERNAL TABLE IF NOT EXISTS ods_order_fact (
    order_id BIGINT COMMENT '订单ID',
    order_sn STRING COMMENT '订单编号',
    user_id BIGINT COMMENT '用户ID',
    total_amount DECIMAL(10,2) COMMENT '订单总金额',
    payment_amount DECIMAL(10,2) COMMENT '实付金额',
    shipping_fee DECIMAL(10,2) COMMENT '运费',
    discount_amount DECIMAL(10,2) COMMENT '优惠金额',
    status STRING COMMENT '订单状态',
    payment_method STRING COMMENT '支付方式',
    created_at TIMESTAMP COMMENT '创建时间',
    paid_at TIMESTAMP COMMENT '支付时间',
    shipped_at TIMESTAMP COMMENT '发货时间',
    completed_at TIMESTAMP COMMENT '完成时间'
)
COMMENT '订单事实表'
PARTITIONED BY (dt STRING COMMENT '日期分区 YYYY-MM-DD')
STORED AS PARQUET
LOCATION '/user/hive/warehouse/ecommerce_dw.db/ods_order_fact';

-- 订单明细表
CREATE EXTERNAL TABLE IF NOT EXISTS ods_order_item (
    order_item_id BIGINT COMMENT '订单项ID',
    order_id BIGINT COMMENT '订单ID',
    product_id BIGINT COMMENT '商品ID',
    sku_id BIGINT COMMENT 'SKU ID',
    product_name STRING COMMENT '商品名称',
    sku_name STRING COMMENT 'SKU名称',
    price DECIMAL(10,2) COMMENT '单价',
    quantity INT COMMENT '数量',
    total_price DECIMAL(10,2) COMMENT '总价'
)
COMMENT '订单明细表'
PARTITIONED BY (dt STRING)
STORED AS PARQUET
LOCATION '/user/hive/warehouse/ecommerce_dw.db/ods_order_item';

-- 用户维度表
CREATE EXTERNAL TABLE IF NOT EXISTS ods_user (
    user_id BIGINT COMMENT '用户ID',
    username STRING COMMENT '用户名',
    name STRING COMMENT '姓名',
    gender STRING COMMENT '性别',
    age INT COMMENT '年龄',
    city STRING COMMENT '城市',
    province STRING COMMENT '省份',
    created_at TIMESTAMP COMMENT '注册时间'
)
COMMENT '用户维度表'
PARTITIONED BY (dt STRING)
STORED AS PARQUET
LOCATION '/user/hive/warehouse/ecommerce_dw.db/ods_user';

-- 商品维度表
CREATE EXTERNAL TABLE IF NOT EXISTS ods_product (
    product_id BIGINT COMMENT '商品ID',
    product_name STRING COMMENT '商品名称',
    category_id BIGINT COMMENT '分类ID',
    brand_id BIGINT COMMENT '品牌ID',
    price DECIMAL(10,2) COMMENT '价格',
    status STRING COMMENT '状态'
)
COMMENT '商品维度表'
PARTITIONED BY (dt STRING)
STORED AS PARQUET
LOCATION '/user/hive/warehouse/ecommerce_dw.db/ods_product';

-- ==================== DWD 层（明细数据层）====================

-- 订单宽表
CREATE TABLE IF NOT EXISTS dwd_order_wide (
    order_id BIGINT,
    order_sn STRING,
    user_id BIGINT,
    user_name STRING,
    user_city STRING,
    user_province STRING,
    product_id BIGINT,
    product_name STRING,
    category_id BIGINT,
    category_name STRING,
    brand_id BIGINT,
    brand_name STRING,
    sku_id BIGINT,
    price DECIMAL(10,2),
    quantity INT,
    total_price DECIMAL(10,2),
    order_total_amount DECIMAL(10,2),
    payment_amount DECIMAL(10,2),
    discount_amount DECIMAL(10,2),
    shipping_fee DECIMAL(10,2),
    payment_method STRING,
    order_status STRING,
    created_at TIMESTAMP,
    paid_at TIMESTAMP,
    shipped_at TIMESTAMP,
    completed_at TIMESTAMP
)
COMMENT '订单宽表'
PARTITIONED BY (dt STRING)
STORED AS ORC
TBLPROPERTIES ('orc.compress'='SNAPPY');

-- ==================== DWS 层（汇总数据层）====================

-- 用户购买汇总表（天粒度）
CREATE TABLE IF NOT EXISTS dws_user_purchase_day (
    user_id BIGINT,
    order_count INT COMMENT '订单数',
    order_amount DECIMAL(10,2) COMMENT '订单金额',
    payment_count INT COMMENT '支付订单数',
    payment_amount DECIMAL(10,2) COMMENT '支付金额',
    refund_count INT COMMENT '退款订单数',
    refund_amount DECIMAL(10,2) COMMENT '退款金额',
    avg_order_value DECIMAL(10,2) COMMENT '平均订单价值'
)
COMMENT '用户购买汇总表（天）'
PARTITIONED BY (dt STRING)
STORED AS ORC;

-- 商品销售汇总表（天粒度）
CREATE TABLE IF NOT EXISTS dws_product_sales_day (
    product_id BIGINT,
    product_name STRING,
    category_id BIGINT,
    category_name STRING,
    brand_id BIGINT,
    brand_name STRING,
    sales_count INT COMMENT '销售件数',
    sales_amount DECIMAL(10,2) COMMENT '销售金额',
    order_count INT COMMENT '订单数',
    buyer_count INT COMMENT '购买人数',
    refund_count INT COMMENT '退款件数',
    refund_amount DECIMAL(10,2) COMMENT '退款金额'
)
COMMENT '商品销售汇总表（天）'
PARTITIONED BY (dt STRING)
STORED AS ORC;

-- 分类销售汇总表（天粒度）
CREATE TABLE IF NOT EXISTS dws_category_sales_day (
    category_id BIGINT,
    category_name STRING,
    sales_count INT,
    sales_amount DECIMAL(10,2),
    order_count INT,
    buyer_count INT
)
COMMENT '分类销售汇总表（天）'
PARTITIONED BY (dt STRING)
STORED AS ORC;

-- ==================== ADS 层（应用数据层）====================

-- 每日销售报表
CREATE TABLE IF NOT EXISTS ads_daily_sales_report (
    report_date DATE,
    total_orders INT COMMENT '总订单数',
    total_amount DECIMAL(10,2) COMMENT '总金额',
    total_users INT COMMENT '总用户数',
    new_users INT COMMENT '新用户数',
    active_users INT COMMENT '活跃用户数',
    avg_order_value DECIMAL(10,2) COMMENT '平均订单价值',
    conversion_rate DECIMAL(5,2) COMMENT '转化率',
    gmv DECIMAL(10,2) COMMENT 'GMV'
)
COMMENT '每日销售报表'
STORED AS ORC;

-- 商品销售排行榜
CREATE TABLE IF NOT EXISTS ads_product_sales_topn (
    rank INT COMMENT '排名',
    product_id BIGINT,
    product_name STRING,
    category_name STRING,
    brand_name STRING,
    sales_count INT,
    sales_amount DECIMAL(10,2),
    report_date DATE
)
COMMENT '商品销售排行榜'
STORED AS ORC;

-- 用户价值分层
CREATE TABLE IF NOT EXISTS ads_user_value_segment (
    user_id BIGINT,
    user_name STRING,
    total_orders INT,
    total_amount DECIMAL(10,2),
    avg_order_value DECIMAL(10,2),
    last_order_date DATE,
    user_segment STRING COMMENT '用户分层: VIP, 高价值, 中价值, 低价值, 流失',
    report_date DATE
)
COMMENT '用户价值分层'
STORED AS ORC;
