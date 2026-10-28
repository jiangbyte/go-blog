package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	UserType  string `json:"user_type"`
	Avatar    string `json:"avatar"`
	Gender    string `json:"gender"`
	Signature string `json:"signature"`
}

type SessionStore struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewSessionStore(rdb *redis.Client, ttlSeconds int) *SessionStore {
	return &SessionStore{rdb: rdb, ttl: time.Duration(ttlSeconds) * time.Second}
}

func (s *SessionStore) key(token string) string {
	return "jb:sess:" + token
}

func (s *SessionStore) Create(ctx context.Context, user SessionUser) (string, error) {
	token := uuid.NewString()
	raw, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	if err := s.rdb.Set(ctx, s.key(token), raw, s.ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *SessionStore) Get(ctx context.Context, token string) (*SessionUser, error) {
	raw, err := s.rdb.Get(ctx, s.key(token)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var u SessionUser
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, err
	}
	_ = s.rdb.Expire(ctx, s.key(token), s.ttl)
	return &u, nil
}

func (s *SessionStore) Delete(ctx context.Context, token string) error {
	return s.rdb.Del(ctx, s.key(token)).Err()
}

// Update refreshes the session payload and TTL.
func (s *SessionStore) Update(ctx context.Context, token string, user SessionUser) error {
	raw, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.key(token), raw, s.ttl).Err()
}
