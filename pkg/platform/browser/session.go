package browser

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/redis/go-redis/v9"
)

const sessionTTL = 4 * time.Hour

type SessionStore struct {
	rdb *redis.Client
}

func NewSessionStore(rdb *redis.Client) *SessionStore {
	return &SessionStore{rdb: rdb}
}

func (s *SessionStore) Save(ctx context.Context, platform, proxyURL string, cookies []*proto.NetworkCookieParam) error {
	if len(cookies) == 0 {
		return nil
	}
	data, err := json.Marshal(cookies)
	if err != nil {
		return fmt.Errorf("session store: marshal: %w", err)
	}
	key := sessionKey(platform, proxyURL)
	return s.rdb.Set(ctx, key, data, sessionTTL).Err()
}

func (s *SessionStore) Load(ctx context.Context, platform, proxyURL string) ([]*proto.NetworkCookieParam, error) {
	key := sessionKey(platform, proxyURL)
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session store: get: %w", err)
	}
	var cookies []*proto.NetworkCookieParam
	if err := json.Unmarshal(data, &cookies); err != nil {
		return nil, fmt.Errorf("session store: unmarshal: %w", err)
	}
	return cookies, nil
}

func sessionKey(platform, proxyURL string) string {
	h := md5.Sum([]byte(proxyURL))
	return fmt.Sprintf("session:%s:%x", platform, h)
}
