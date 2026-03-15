package browser_test

import (
	"context"
	"testing"

	"github.com/go-rod/rod/lib/proto"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"review-curator/pkg/platform/browser"
)

func TestSessionStore_SaveLoad(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer func() { _ = rdb.Close() }()

	store := browser.NewSessionStore(rdb)
	ctx := context.Background()

	cookies := []*proto.NetworkCookieParam{
		{Name: "session_id", Value: "abc123", Domain: ".shopee.co.id"},
	}

	err := store.Save(ctx, "shopee", "http://proxy:8080", cookies)
	require.NoError(t, err)

	loaded, err := store.Load(ctx, "shopee", "http://proxy:8080")
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.Equal(t, "abc123", loaded[0].Value)
}

func TestSessionStore_MissReturnsNil(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer func() { _ = rdb.Close() }()

	store := browser.NewSessionStore(rdb)
	loaded, err := store.Load(context.Background(), "shopee", "http://nonexistent:9999")
	require.NoError(t, err)
	assert.Nil(t, loaded)
}
