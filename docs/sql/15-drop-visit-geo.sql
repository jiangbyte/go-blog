USE go_blog;

-- 已移除访问地图功能：删除地区聚合表（可选手动执行）
DROP TABLE IF EXISTS blog_stats_geo_city;
DROP TABLE IF EXISTS blog_stats_geo_country;
