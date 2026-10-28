package blog

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/jiangbyte/go-blog/internal/model"
)

var (
	cstZone     = time.FixedZone("CST", 8*3600)
	visitorIDRe = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	postUIDRe   = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
)

const (
	redisKeyPVTotal        = "blog:stats:pv:total"
	redisKeyArticlePVTotal = "blog:stats:article_pv:total"
	// 累计 UV = baseline（落库回填）+ HLL（重启后新增访客），避免 Redis 清空后重复计
	redisKeyUVBaseline    = "blog:stats:uv:baseline"
	redisKeyUVHLL         = "blog:stats:uv:hll"
	redisKeyUVTotal       = "blog:stats:uv:total" // 同步时折叠后的整数快照
	redisKeyPostViewsZ    = "blog:stats:post:views"
	redisKeyProjectViewsZ = "blog:stats:project:views"
	redisDayKeyTTL        = 90 * 24 * time.Hour
	// 同一访客短时重复打开/刷新同一页不重复计 PV / 阅读
	redisHitDedupeTTL = 30 * time.Minute
)

func dayUVSeenKey(day time.Time, visitorID string) string {
	return dayKey(day, "uv:seen:"+visitorID)
}

func (s *Service) lifetimeUV(ctx context.Context) (int64, error) {
	base, err := redisInt(ctx, s.rdb, redisKeyUVBaseline)
	if err != nil {
		return 0, err
	}
	hll, err := s.rdb.PFCount(ctx, redisKeyUVHLL).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	sum := base + hll
	if sum > 0 {
		return sum, nil
	}
	// 冷启动兜底：baseline/HLL 皆空时用上次落库快照
	return redisInt(ctx, s.rdb, redisKeyUVTotal)
}

type VisitInput struct {
	VisitorID  string `json:"visitor_id"`
	Path       string `json:"path"`
	PostUID    string `json:"post_uid"`
	ProjectUID string `json:"project_uid"`
}

type VisitResult struct {
	OK        bool  `json:"ok"`
	Counted   bool  `json:"counted"`
	ViewCount int64 `json:"view_count,omitempty"`
}

func todayCST() time.Time {
	now := time.Now().In(cstZone)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, cstZone)
}

func dayKey(day time.Time, suffix string) string {
	return "blog:stats:day:" + day.Format("2006-01-02") + ":" + suffix
}

func normalizeVisitorID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if !visitorIDRe.MatchString(id) {
		return "", fmt.Errorf("invalid visitor_id")
	}
	return strings.ToLower(id), nil
}

func normalizePath(raw string) string {
	p := strings.TrimSpace(raw)
	if p == "" {
		return "/"
	}
	if i := strings.IndexAny(p, "?#"); i >= 0 {
		p = p[:i]
	}
	if p == "" {
		p = "/"
	}
	if len(p) > 512 {
		p = p[:512]
	}
	return p
}

func hitDedupeKey(visitorID, path, postUID, projectUID string) string {
	scope := path
	if postUID != "" {
		scope = "post:" + postUID
	} else if projectUID != "" {
		scope = "project:" + projectUID
	}
	return "blog:stats:hit:" + visitorID + ":" + scope
}

func normalizeEntityUID(raw string) string {
	return normalizePostUID(raw)
}

func (s *Service) entityViewCount(ctx context.Context, zkey, uid string) int64 {
	if uid == "" {
		return 0
	}
	if s.rdb != nil {
		score, err := s.rdb.ZScore(ctx, zkey, uid).Result()
		if err == nil {
			return int64(score)
		}
	}
	// Redis 未命中时回落 MySQL 持久化值
	kind := "post"
	if zkey == redisKeyProjectViewsZ {
		kind = "project"
	}
	var row model.BlogEntityViews
	if err := s.db.WithContext(ctx).
		Select("view_count").
		Where("kind = ? AND uid = ?", kind, uid).
		First(&row).Error; err != nil {
		return 0
	}
	return row.ViewCount
}

func normalizePostUID(raw string) string {
	uid := strings.TrimSpace(raw)
	if uid == "" || !postUIDRe.MatchString(uid) {
		return ""
	}
	return strings.ToLower(uid)
}

func (s *Service) requireRedis() error {
	if s.rdb == nil {
		return fmt.Errorf("redis unavailable")
	}
	return nil
}

const visitRateLimitPerMin = 120

// AllowVisitRate 按 IP 限制 visit 上报频率，减轻刷量与 Redis 压力。
func (s *Service) AllowVisitRate(ctx context.Context, ip string) error {
	if err := s.requireRedis(); err != nil {
		return err
	}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}
	key := "blog:rl:visit:" + ip + ":" + time.Now().UTC().Format("200601021504")
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("rate limit unavailable")
	}
	if n == 1 {
		_ = s.rdb.Expire(ctx, key, 2*time.Minute).Err()
	}
	if n > visitRateLimitPerMin {
		return fmt.Errorf("too many requests")
	}
	return nil
}

// TrackVisit 高频写入全部走 Redis：PV 计数、HLL UV、文章/项目阅读 ZSET。
// 同一访客 30 分钟内重复访问同一路径/内容只记一次（避免刷新刷量）。
func (s *Service) TrackVisit(ctx context.Context, in VisitInput) (*VisitResult, error) {
	// 1. 校验 Redis 与访客/路径/实体 uid
	if err := s.requireRedis(); err != nil {
		return nil, err
	}
	vid, err := normalizeVisitorID(in.VisitorID)
	if err != nil {
		return nil, err
	}
	path := normalizePath(in.Path)
	postUID := normalizeEntityUID(in.PostUID)
	projectUID := normalizeEntityUID(in.ProjectUID)
	// 文章优先；避免同时带两个 uid
	if postUID != "" {
		projectUID = ""
	}
	day := todayCST()

	// 2. 累计 UV（HLL）与当日 UV（SetNX 去重后 INCR）
	_ = s.rdb.PFAdd(ctx, redisKeyUVHLL, vid).Err()
	if ok, err := s.rdb.SetNX(ctx, dayUVSeenKey(day, vid), "1", redisDayKeyTTL).Result(); err != nil {
		return nil, err
	} else if ok {
		_ = s.rdb.Incr(ctx, dayKey(day, "uv_count")).Err()
		_ = s.rdb.Expire(ctx, dayKey(day, "uv_count"), redisDayKeyTTL).Err()
	}

	// 3. 短窗去重：未计数则直接返回当前阅读量
	ok, err := s.rdb.SetNX(ctx, hitDedupeKey(vid, path, postUID, projectUID), "1", redisHitDedupeTTL).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		out := &VisitResult{OK: true, Counted: false}
		if postUID != "" {
			out.ViewCount = s.entityViewCount(ctx, redisKeyPostViewsZ, postUID)
		} else if projectUID != "" {
			out.ViewCount = s.entityViewCount(ctx, redisKeyProjectViewsZ, projectUID)
		}
		return out, nil
	}

	// 4. Pipeline 累加站点 PV；实体存在时再累加文章/项目阅读
	pipe := s.rdb.Pipeline()
	pipe.Incr(ctx, redisKeyPVTotal)
	pipe.Incr(ctx, dayKey(day, "pv"))
	pipe.Expire(ctx, dayKey(day, "pv"), redisDayKeyTTL)

	var viewCount int64
	if postUID != "" {
		var n int64
		if err := s.db.WithContext(ctx).Model(&model.BlogPost{}).Where("uid = ?", postUID).Count(&n).Error; err != nil {
			return nil, err
		}
		if n > 0 {
			pipe.Incr(ctx, redisKeyArticlePVTotal)
			pipe.Incr(ctx, dayKey(day, "article_pv"))
			pipe.Expire(ctx, dayKey(day, "article_pv"), redisDayKeyTTL)
			pipe.ZIncrBy(ctx, redisKeyPostViewsZ, 1, postUID)
		}
	} else if projectUID != "" {
		var n int64
		if err := s.db.WithContext(ctx).Model(&model.BlogProject{}).Where("uid = ?", projectUID).Count(&n).Error; err != nil {
			return nil, err
		}
		if n > 0 {
			pipe.ZIncrBy(ctx, redisKeyProjectViewsZ, 1, projectUID)
		}
	}

	if _, err = pipe.Exec(ctx); err != nil {
		return nil, err
	}
	if postUID != "" {
		viewCount = s.entityViewCount(ctx, redisKeyPostViewsZ, postUID)
	} else if projectUID != "" {
		viewCount = s.entityViewCount(ctx, redisKeyProjectViewsZ, projectUID)
	}
	return &VisitResult{OK: true, Counted: true, ViewCount: viewCount}, nil
}

func redisInt(ctx context.Context, rdb *redis.Client, key string) (int64, error) {
	raw, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		// 历史日 UV 曾用 HLL（值为 HYLL…），丢弃坏 key
		_ = rdb.Del(ctx, key).Err()
		return 0, nil
	}
	return n, nil
}

func (s *Service) statsTotals(ctx context.Context) (sitePV, siteUV, articlePV int64, err error) {
	if err = s.requireRedis(); err != nil {
		return
	}
	sitePV, err = redisInt(ctx, s.rdb, redisKeyPVTotal)
	if err != nil {
		return
	}
	articlePV, err = redisInt(ctx, s.rdb, redisKeyArticlePVTotal)
	if err != nil {
		return
	}
	siteUV, err = s.lifetimeUV(ctx)
	return
}

type dayCountersRow struct {
	Day       time.Time
	SitePV    int64
	SiteUV    int64
	ArticlePV int64
}

func parseRedisStringInt(raw string, err error) (int64, error) {
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, nil
	}
	return n, nil
}

// loadDayCounters 一次 pipeline 拉取多日计数，避免按天 N 次往返。
func (s *Service) loadDayCounters(ctx context.Context, days []time.Time) ([]dayCountersRow, error) {
	if err := s.requireRedis(); err != nil {
		return nil, err
	}
	if len(days) == 0 {
		return nil, nil
	}
	type cmds struct {
		pv, uv, art *redis.StringCmd
	}
	pipe := s.rdb.Pipeline()
	cs := make([]cmds, len(days))
	for i, d := range days {
		cs[i].pv = pipe.Get(ctx, dayKey(d, "pv"))
		cs[i].uv = pipe.Get(ctx, dayKey(d, "uv_count"))
		cs[i].art = pipe.Get(ctx, dayKey(d, "article_pv"))
	}
	_, _ = pipe.Exec(ctx)
	out := make([]dayCountersRow, len(days))
	for i, d := range days {
		pv, err := parseRedisStringInt(cs[i].pv.Result())
		if err != nil {
			return nil, err
		}
		uv, err := parseRedisStringInt(cs[i].uv.Result())
		if err != nil {
			return nil, err
		}
		art, err := parseRedisStringInt(cs[i].art.Result())
		if err != nil {
			return nil, err
		}
		out[i] = dayCountersRow{Day: d, SitePV: pv, SiteUV: uv, ArticlePV: art}
	}
	return out, nil
}

func (s *Service) dayCounters(ctx context.Context, day time.Time) (pv, uv, articlePV int64, err error) {
	rows, err := s.loadDayCounters(ctx, []time.Time{day})
	if err != nil || len(rows) == 0 {
		return 0, 0, 0, err
	}
	return rows[0].SitePV, rows[0].SiteUV, rows[0].ArticlePV, nil
}

// AdminStatsOverview 博客运营台：内容库存 + 流量读 Redis，不统计后台用户。
func (s *Service) AdminStatsOverview(ctx context.Context, days int) (map[string]any, error) {
	if days != 7 && days != 14 && days != 30 {
		days = 7
	}

	// 1. 内容库存：文章/项目按状态、taxonomy 数量（禁止用户数）
	var (
		postPublished, postDraft       int64
		projectPublished, projectDraft int64
		categoryCount, tagCount        int64
		projectTagCount                int64
	)
	_ = s.db.WithContext(ctx).Model(&model.BlogPost{}).Where("status = ?", StatusPublished).Count(&postPublished).Error
	_ = s.db.WithContext(ctx).Model(&model.BlogPost{}).Where("status = ?", StatusDraft).Count(&postDraft).Error
	_ = s.db.WithContext(ctx).Model(&model.BlogProject{}).Where("status = ?", StatusPublished).Count(&projectPublished).Error
	_ = s.db.WithContext(ctx).Model(&model.BlogProject{}).Where("status = ?", StatusDraft).Count(&projectDraft).Error
	_ = s.db.WithContext(ctx).Model(&model.BlogTaxonomy{}).Where("kind = ?", "category").Count(&categoryCount).Error
	_ = s.db.WithContext(ctx).Model(&model.BlogTaxonomy{}).Where("kind = ?", "tag").Count(&tagCount).Error
	_ = s.db.WithContext(ctx).Model(&model.BlogTaxonomy{}).Where("kind = ?", "project_tag").Count(&projectTagCount).Error

	sitePV, siteUV, articlePV, err := s.statsTotals(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 近 N 天流量序列 + 区间合计
	today := todayCST()
	dayList := make([]time.Time, 0, days)
	for i := days - 1; i >= 0; i-- {
		dayList = append(dayList, today.AddDate(0, 0, -i))
	}
	dayRows, err := s.loadDayCounters(ctx, dayList)
	if err != nil {
		return nil, err
	}
	traffic := make([]map[string]any, 0, len(dayRows))
	pvSeries := make([]map[string]any, 0, len(dayRows))
	var periodPV, periodUV, periodArticlePV int64
	for _, r := range dayRows {
		key := r.Day.Format("2006-01-02")
		traffic = append(traffic, map[string]any{
			"date": key, "site_pv": r.SitePV, "site_uv": r.SiteUV, "article_pv": r.ArticlePV,
		})
		pvSeries = append(pvSeries, map[string]any{"date": key, "count": r.ArticlePV})
		periodPV += r.SitePV
		periodUV += r.SiteUV
		periodArticlePV += r.ArticlePV
	}

	// 3. 阅读 Top（Redis ZSET）
	topOut := make([]map[string]any, 0, 5)
	if s.rdb != nil {
		pairs, err := s.rdb.ZRevRangeWithScores(ctx, redisKeyPostViewsZ, 0, 4).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		uids := make([]string, 0, len(pairs))
		scoreByUID := make(map[string]int64, len(pairs))
		for _, p := range pairs {
			uid := fmt.Sprint(p.Member)
			uids = append(uids, uid)
			scoreByUID[uid] = int64(p.Score)
		}
		if len(uids) > 0 {
			var posts []model.BlogPost
			if err := s.db.WithContext(ctx).Select("uid", "title").Where("uid IN ?", uids).Find(&posts).Error; err != nil {
				log.Printf("AdminStatsOverview load post titles: %v", err)
			}
			titleByUID := make(map[string]string, len(posts))
			for _, p := range posts {
				titleByUID[p.UID] = p.Title
			}
			for _, uid := range uids {
				title := titleByUID[uid]
				if title == "" {
					title = uid
				}
				topOut = append(topOut, map[string]any{
					"id": uid, "uid": uid, "title": title,
					"view_count": scoreByUID[uid], "slug": uid,
				})
			}
		}
	}

	// 4. 待处理草稿（最近 5 篇，供运营台快捷入口）
	draftOut := make([]map[string]any, 0, 5)
	var draftRows []model.BlogPost
	if err := s.db.WithContext(ctx).
		Select("uid", "title", "updated_at", "created_at").
		Where("status = ?", StatusDraft).
		Order("id DESC").
		Limit(5).
		Find(&draftRows).Error; err != nil {
		log.Printf("AdminStatsOverview load drafts: %v", err)
	} else {
		for _, p := range draftRows {
			updated := p.CreatedAt
			if p.ContentUpdatedAt != nil {
				updated = *p.ContentUpdatedAt
			}
			draftOut = append(draftOut, map[string]any{
				"uid":        p.UID,
				"title":      p.Title,
				"updated_at": updated.Format(time.RFC3339),
			})
		}
	}

	now := time.Now().In(cstZone)
	return map[string]any{
		"days": days,
		"counts": map[string]any{
			"post_published":    postPublished,
			"post_draft":        postDraft,
			"project_published": projectPublished,
			"project_draft":     projectDraft,
			"category":          categoryCount,
			"tag":               tagCount,
			"project_tag":       projectTagCount,
			// 兼容旧字段：article/project 为总量
			"article": postPublished + postDraft,
			"project": projectPublished + projectDraft,
		},
		"period": map[string]any{
			"site_pv": periodPV, "site_uv": periodUV, "article_pv": periodArticlePV,
		},
		"total": map[string]any{
			"site_pv": sitePV, "site_uv": siteUV, "article_pv": articlePV,
		},
		"traffic": traffic,
		"article": map[string]any{
			"pv_series": pvSeries,
			"top":       topOut,
		},
		"drafts":       draftOut,
		"generated_at": now.Format(time.RFC3339),
	}, nil
}
