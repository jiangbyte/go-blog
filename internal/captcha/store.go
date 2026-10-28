package captcha

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store 对齐 mypage CaptchaStore：挑战答案 + 一次性 captcha_token。
type Store struct {
	rdb           *redis.Client
	challengeTTL  time.Duration
	tokenTTL      time.Duration
	tolerancePx   float64
	maxFailures   int
}

func NewStore(rdb *redis.Client, challengeTTL, tokenTTL, tolerancePx, maxFailures int) *Store {
	if maxFailures < 1 {
		maxFailures = 5
	}
	return &Store{
		rdb:          rdb,
		challengeTTL: time.Duration(challengeTTL) * time.Second,
		tokenTTL:     time.Duration(tokenTTL) * time.Second,
		tolerancePx:  float64(tolerancePx),
		maxFailures:  maxFailures,
	}
}

func (s *Store) challengeKey(id string) string { return "captcha:challenge:" + id }
func (s *Store) failKey(id string) string      { return "captcha:fail:" + id }
func (s *Store) tokenKey(token string) string  { return "captcha:token:" + token }

func (s *Store) SaveChallenge(ctx context.Context, answerX float64, y int) (string, error) {
	id, err := tokenURLSafe(16)
	if err != nil {
		return "", err
	}
	payload, _ := json.Marshal(map[string]any{"x": answerX, "y": y})
	if err := s.rdb.Set(ctx, s.challengeKey(id), payload, s.challengeTTL).Err(); err != nil {
		return "", err
	}
	return id, nil
}

// VerifySlide 校验滑动位移并签发一次性 captcha_token（对齐 mypage）。
func (s *Store) VerifySlide(ctx context.Context, captchaID string, x float64) (string, error) {
	// 1. 读取挑战答案；缺失/损坏视为已失效
	cid := strings.TrimSpace(captchaID)
	if cid == "" {
		return "", fmt.Errorf("验证码已失效，请刷新后重试")
	}
	raw, err := s.rdb.Get(ctx, s.challengeKey(cid)).Bytes()
	if err == redis.Nil {
		return "", fmt.Errorf("验证码已失效，请刷新后重试")
	}
	if err != nil {
		return "", err
	}
	var data struct {
		X float64 `json:"x"`
		Y int     `json:"y"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", fmt.Errorf("验证码已失效，请刷新后重试")
	}

	// 2. 位移超容差则累计失败次数，达上限作废挑战
	if math.Abs(x-data.X) > s.tolerancePx {
		fails, err := s.rdb.Incr(ctx, s.failKey(cid)).Result()
		if err != nil {
			return "", err
		}
		_ = s.rdb.Expire(ctx, s.failKey(cid), s.challengeTTL).Err()
		if int(fails) >= s.maxFailures {
			_ = s.rdb.Del(ctx, s.challengeKey(cid), s.failKey(cid)).Err()
		}
		return "", fmt.Errorf("验证码校验失败")
	}

	// 3. 校验通过：删除挑战，签发短时一次性 token
	_ = s.rdb.Del(ctx, s.challengeKey(cid), s.failKey(cid)).Err()
	token, err := tokenURLSafe(24)
	if err != nil {
		return "", err
	}
	if err := s.rdb.Set(ctx, s.tokenKey(token), "1", s.tokenTTL).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// ConsumeToken 业务入口一次性消费 captcha_token。
func (s *Store) ConsumeToken(ctx context.Context, captchaToken string) error {
	token := strings.TrimSpace(captchaToken)
	if token == "" {
		return fmt.Errorf("请先完成滑块验证")
	}
	n, err := s.rdb.Del(ctx, s.tokenKey(token)).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("验证码已失效，请重新验证")
	}
	return nil
}

func tokenURLSafe(nbytes int) (string, error) {
	buf := make([]byte, nbytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
