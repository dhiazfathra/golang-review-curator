package browser_test

import (
	"context"
	"testing"

	"review-curator/pkg/platform/browser"

	"github.com/stretchr/testify/require"
)

func TestPool_AcquireRelease(t *testing.T) {
	pool, err := browser.NewPool(1, "")
	require.NoError(t, err)
	defer pool.Close()

	page, release, err := pool.Acquire(context.Background())
	require.NoError(t, err)
	require.NotNil(t, page)
	require.NotNil(t, release)
	release()
}

func TestPool_RoundRobin(t *testing.T) {
	pool, err := browser.NewPool(2, "")
	require.NoError(t, err)
	defer pool.Close()

	for i := 0; i < 4; i++ {
		page, release, err := pool.Acquire(context.Background())
		require.NoError(t, err)
		require.NotNil(t, page)
		release()
	}
}
