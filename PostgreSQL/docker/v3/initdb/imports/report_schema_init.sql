-- ====================================================================
-- 文件: report_schema_init.sql
-- 说明: report_db 报表/数仓系统表结构初始化
-- ====================================================================

-- 创建月度活跃用户统计表
CREATE TABLE IF NOT EXISTS public.rpt_monthly_active_users (
    report_month DATE PRIMARY KEY,        -- 统计月份，如 '2026-07-01'
    total_active_count INTEGER NOT NULL DEFAULT 0,
    new_register_count INTEGER NOT NULL DEFAULT 0,
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建销售数据汇总表
CREATE TABLE IF NOT EXISTS public.rpt_sales_summary (
    id SERIAL PRIMARY KEY,
    category_name VARCHAR(100) NOT NULL,
    total_revenue DECIMAL(15, 2) DEFAULT 0.00,
    statistic_date DATE NOT NULL
);

-- 为统计日期创建索引，加速报表查询
CREATE INDEX idx_rpt_sales_date ON public.rpt_sales_summary(statistic_date);