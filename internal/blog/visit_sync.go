package blog

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/batch"
)

const (
	statsSyncInterval   = 60 * time.Second
	statsHydrateDays    = 90
	statsDBBatchSize    = batch.DefaultSize
	statsRedisPipeChunk = 200
)

// StartStatsSync 启动时从 MySQL 回填空 Redis，并周期性异步落库。
func (s *Service) StartStatsSync(ctx context.Context) {
	if s.rdb == nil || s.db == nil {
		return
	}
	if err := s.db.WithContext(ctx).AutoMigrate(
		&model.BlogStatsTotal{},
		&model.BlogStatsDaily{},
		&model.BlogEntityViews{},
	); err != nil {
		log.Printf("blog stats migrate: %v", err)
		return
	}
	_ = s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.BlogStatsTotal{ID: 1}).Error

	if err := s.hydrateStatsFromDB(ctx); err != nil {
		log.Printf("blog stats hydrate: %v", err)
	}
	go s.statsSyncLoop(ctx)
}

func (s *Service) statsSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(statsSyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			flushCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			if err := s.flushStatsToDB(flushCtx); err != nil {
				log.Printf("blog stats final flush: %v", err)
			}
			cancel()
			return
		case <-ticker.C:
			if err := s.flushStatsToDB(ctx); err != nil {
				log.Printf("blog stats flush: %v", err)
			}
		}
	}
}

func (s *Service) hydrateStatsFromDB(ctx context.Context) error {
	// Redis 已有累计 PV 则视为热数据在，跳过回填，避免覆盖
	if n, err := redisInt(ctx, s.rdb, redisKeyPVTotal); err == nil && n > 0 {
		return nil
	}

	var total model.BlogStatsTotal
	if err := s.db.WithContext(ctx).First(&total, 1).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	pipe := s.rdb.Pipeline()
	pending := 0
	flushPipe := func() error {
		if pending == 0 {
			return nil
		}
		_, err := pipe.Exec(ctx)
		pipe = s.rdb.Pipeline()
		pending = 0
		if err == redis.Nil {
			return nil
		}
		return err
	}
	add := func() error {
		pending++
		if pending >= statsRedisPipeChunk {
			return flushPipe()
		}
		return nil
	}

	if total.SitePV > 0 {
		pipe.Set(ctx, redisKeyPVTotal, total.SitePV, 0)
		if err := add(); err != nil {
			return err
		}
	}
	if total.SiteUV > 0 {
		pipe.Set(ctx, redisKeyUVBaseline, total.SiteUV, 0)
		pipe.Set(ctx, redisKeyUVTotal, total.SiteUV, 0)
		pipe.Del(ctx, redisKeyUVHLL)
		pending += 3
		if pending >= statsRedisPipeChunk {
			if err := flushPipe(); err != nil {
				return err
			}
		}
	}
	if total.ArticlePV > 0 {
		pipe.Set(ctx, redisKeyArticlePVTotal, total.ArticlePV, 0)
		if err := add(); err != nil {
			return err
		}
	}

	today := todayCST()
	start := today.AddDate(0, 0, -(statsHydrateDays - 1))
	var days []model.BlogStatsDaily
	if err := s.db.WithContext(ctx).Where("day >= ? AND day <= ?", start, today).Find(&days).Error; err != nil {
		return err
	}
	for _, d := range days {
		day := d.Day.In(cstZone)
		if d.SitePV > 0 {
			pipe.Set(ctx, dayKey(day, "pv"), d.SitePV, redisDayKeyTTL)
			if err := add(); err != nil {
				return err
			}
		}
		if d.SiteUV > 0 {
			pipe.Set(ctx, dayKey(day, "uv_count"), d.SiteUV, redisDayKeyTTL)
			if err := add(); err != nil {
				return err
			}
		}
		if d.ArticlePV > 0 {
			pipe.Set(ctx, dayKey(day, "article_pv"), d.ArticlePV, redisDayKeyTTL)
			if err := add(); err != nil {
				return err
			}
		}
	}

	// 分批读实体阅读，避免一次拉全表后再巨型 pipeline
	const pageSize = statsDBBatchSize
	for offset := 0; ; offset += pageSize {
		var entities []model.BlogEntityViews
		if err := s.db.WithContext(ctx).
			Order("kind, uid").
			Offset(offset).Limit(pageSize).
			Find(&entities).Error; err != nil {
			return err
		}
		if len(entities) == 0 {
			break
		}
		for _, e := range entities {
			if e.ViewCount <= 0 || e.UID == "" {
				continue
			}
			z := redis.Z{Score: float64(e.ViewCount), Member: e.UID}
			switch e.Kind {
			case "post":
				pipe.ZAdd(ctx, redisKeyPostViewsZ, z)
			case "project":
				pipe.ZAdd(ctx, redisKeyProjectViewsZ, z)
			default:
				continue
			}
			if err := add(); err != nil {
				return err
			}
		}
		if len(entities) < pageSize {
			break
		}
	}
	return flushPipe()
}

func (s *Service) flushStatsToDB(ctx context.Context) error {
	sitePV, siteUV, articlePV, err := s.statsTotals(ctx)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"site_pv", "site_uv", "article_pv", "updated_at",
		}),
	}).Create(&model.BlogStatsTotal{
		ID: 1, SitePV: sitePV, SiteUV: siteUV, ArticlePV: articlePV,
	}).Error; err != nil {
		return err
	}
	// 仅快照，不折叠 HLL（折叠会导致老访客再次 PFADD 被重复计数）
	_ = s.rdb.Set(ctx, redisKeyUVTotal, siteUV, 0).Err()

	today := todayCST()
	dayList := make([]time.Time, statsHydrateDays)
	for i := 0; i < statsHydrateDays; i++ {
		dayList[i] = today.AddDate(0, 0, -i)
	}
	dayRows, err := s.loadDayCounters(ctx, dayList)
	if err != nil {
		return err
	}
	daily := make([]model.BlogStatsDaily, 0, len(dayRows))
	for _, r := range dayRows {
		if r.SitePV == 0 && r.SiteUV == 0 && r.ArticlePV == 0 {
			continue
		}
		daily = append(daily, model.BlogStatsDaily{
			Day: r.Day, SitePV: r.SitePV, SiteUV: r.SiteUV, ArticlePV: r.ArticlePV,
		})
	}
	if len(daily) > 0 {
		if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "day"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"site_pv", "site_uv", "article_pv", "updated_at",
			}),
		}).CreateInBatches(daily, statsDBBatchSize).Error; err != nil {
			return err
		}
	}

	if err := s.flushEntityViews(ctx, "post", redisKeyPostViewsZ); err != nil {
		return err
	}
	return s.flushEntityViews(ctx, "project", redisKeyProjectViewsZ)
}

func (s *Service) flushEntityViews(ctx context.Context, kind, zkey string) error {
	const zscanCount int64 = statsDBBatchSize
	var cursor uint64
	for {
		keys, next, err := s.rdb.ZScan(ctx, zkey, cursor, "*", zscanCount).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		rows := make([]model.BlogEntityViews, 0, len(keys)/2)
		for i := 0; i+1 < len(keys); i += 2 {
			uid := keys[i]
			score, err := strconv.ParseFloat(keys[i+1], 64)
			if err != nil || uid == "" {
				continue
			}
			count := int64(score)
			if count < 0 {
				continue
			}
			rows = append(rows, model.BlogEntityViews{
				Kind: kind, UID: uid, ViewCount: count,
			})
		}
		if len(rows) > 0 {
			if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "kind"}, {Name: "uid"}},
				DoUpdates: clause.AssignmentColumns([]string{"view_count", "updated_at"}),
			}).CreateInBatches(rows, statsDBBatchSize).Error; err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
