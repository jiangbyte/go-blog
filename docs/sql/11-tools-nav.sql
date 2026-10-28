-- Author: go-blog — 站点顶栏注入「工具」导航（已存在则跳过）
-- 注意：JSON 对象勿经用户变量再插入，否则会被当成 STRING。
SET NAMES utf8mb4;

UPDATE blog_site
SET config_json = JSON_SET(
  config_json,
  '$.nav',
  JSON_ARRAY_INSERT(
    JSON_EXTRACT(config_json, '$.nav'),
    CONCAT(
      '$[',
      IFNULL(
        (
          SELECT idx
          FROM JSON_TABLE(
            JSON_EXTRACT(config_json, '$.nav'),
            '$[*]' COLUMNS (
              idx FOR ORDINALITY,
              name VARCHAR(64) CHARACTER SET utf8mb4 PATH '$.name'
            )
          ) t
          WHERE t.name = '项目'
          LIMIT 1
        ),
        2
      ),
      ']'
    ),
    JSON_OBJECT(
      'name', '工具',
      'children', JSON_ARRAY(
        JSON_OBJECT('name', 'JSON 格式化', 'href', '/tools/json', 'icon', 'code'),
        JSON_OBJECT('name', '子网划分', 'href', '/tools/subnet', 'icon', 'apps')
      )
    )
  )
)
WHERE id = 1
  AND JSON_EXTRACT(config_json, '$.nav') IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM JSON_TABLE(
      JSON_EXTRACT(config_json, '$.nav'),
      '$[*]' COLUMNS (name VARCHAR(64) CHARACTER SET utf8mb4 PATH '$.name')
    ) t
    WHERE t.name = '工具'
  );
