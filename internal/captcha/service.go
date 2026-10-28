package captcha

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/jiangbyte/go-blog/internal/config"
)

// CreateResult 对齐 mypage CaptchaCreateResult / 前端字段。
type CreateResult struct {
	CaptchaID  string `json:"captcha_id"`
	BGURL      string `json:"bg_url"`
	PuzzleURL  string `json:"puzzle_url"`
	Y          int    `json:"y"`
	BGWidth    int    `json:"bg_width"`
	BGHeight   int    `json:"bg_height"`
	PuzzleSize int    `json:"puzzle_size"`
}

// Service 对齐 mypage CaptchaApplicationService（创建拼图 + 校验滑动 + 消费 token）。
type Service struct {
	cfg    config.CaptchaConfig
	images *ImageSource
	store  *Store
}

func NewService(rdb *redis.Client, cfg config.CaptchaConfig) *Service {
	return &Service{
		cfg:    cfg,
		images: NewImageSource(cfg, nil),
		store:  NewStore(rdb, cfg.TTLSeconds, cfg.TokenTTLSeconds, cfg.TolerancePx, cfg.MaxFailures),
	}
}

func (s *Service) Create(ctx context.Context) (*CreateResult, error) {
	// 1. 按配置拉取远端/本地底图（remote 失败才落 solid 兜底）
	raw := s.images.FetchBytes()
	// 2. 切拼图；解码失败（极少）再用兜底底图重试一次
	puzzle, err := CreatePuzzleFromBytes(raw, s.cfg.BGWidth, s.cfg.BGHeight, s.cfg.PuzzleSize)
	if err != nil {
		puzzle, err = CreatePuzzleFromBytes(
			SolidFallbackImage(s.cfg.BGWidth, s.cfg.BGHeight),
			s.cfg.BGWidth, s.cfg.BGHeight, s.cfg.PuzzleSize,
		)
		if err != nil {
			return nil, err
		}
	}
	id, err := s.store.SaveChallenge(ctx, float64(puzzle.X), puzzle.Y)
	if err != nil {
		return nil, err
	}
	return &CreateResult{
		CaptchaID:  id,
		BGURL:      puzzle.BGDataURL,
		PuzzleURL:  puzzle.PuzzleDataURL,
		Y:          puzzle.Y,
		BGWidth:    puzzle.BGWidth,
		BGHeight:   puzzle.BGHeight,
		PuzzleSize: puzzle.PuzzleSize,
	}, nil
}

func (s *Service) Verify(ctx context.Context, captchaID string, x float64) (string, error) {
	return s.store.VerifySlide(ctx, captchaID, x)
}

func (s *Service) ConsumeToken(ctx context.Context, captchaToken string) error {
	return s.store.ConsumeToken(ctx, captchaToken)
}
