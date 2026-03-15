package captcha_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"review-curator/pkg/platform/captcha"
)

type mockResolver struct {
	token string
	err   error
}

func (m *mockResolver) SolveImage(_ context.Context, _ string) (string, error) {
	return m.token, m.err
}
func (m *mockResolver) SolveRecaptchaV2(_ context.Context, _, _ string) (string, error) {
	return m.token, m.err
}
func (m *mockResolver) SolveRecaptchaV3(_ context.Context, _, _, _ string) (string, error) {
	return m.token, m.err
}

func TestDispatcher_PrimarySucceeds(t *testing.T) {
	d := captcha.NewDispatcher(
		&mockResolver{token: "tok1"},
		&mockResolver{token: "tok2"},
	)
	tok, err := d.SolveRecaptchaV2(context.Background(), "key", "url")
	require.NoError(t, err)
	assert.Equal(t, "tok1", tok)
}

func TestDispatcher_FallbackOnPrimaryError(t *testing.T) {
	d := captcha.NewDispatcher(
		&mockResolver{err: errors.New("primary down")},
		&mockResolver{token: "tok2"},
	)
	tok, err := d.SolveRecaptchaV2(context.Background(), "key", "url")
	require.NoError(t, err)
	assert.Equal(t, "tok2", tok)
}

func TestDispatcher_BothFail(t *testing.T) {
	d := captcha.NewDispatcher(
		&mockResolver{err: errors.New("primary down")},
		&mockResolver{err: errors.New("secondary down")},
	)
	_, err := d.SolveRecaptchaV2(context.Background(), "key", "url")
	require.Error(t, err)
}
