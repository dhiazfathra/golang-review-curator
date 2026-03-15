package selector

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/jmoiron/sqlx"
)

type selectorRow struct {
	Platform string `db:"platform"`
	Field    string `db:"field"`
	Rules    []byte `db:"rules"`
}

func (r selectorRow) toConfig() SelectorConfig {
	var rules []SelectorRule
	_ = json.Unmarshal(r.Rules, &rules)
	return SelectorConfig{Platform: r.Platform, Field: r.Field, Rules: rules}
}

// SelectorStore manages CSS/XPath selectors for platform scraping.
type SelectorStore struct {
	db      *sqlx.DB
	current unsafe.Pointer
}

// NewSelectorStore creates a new selector store.
func NewSelectorStore(db *sqlx.DB) (*SelectorStore, error) {
	s := &SelectorStore{db: db}
	m, err := s.loadFromDB(context.Background())
	if err != nil {
		return nil, fmt.Errorf("selector store init: %w", err)
	}
	atomic.StorePointer(&s.current, unsafe.Pointer(&m))
	return s, nil
}

// Get retrieves a selector config for a specific platform and field.
func (s *SelectorStore) Get(platform, field string) SelectorConfig {
	m := *(*map[string]SelectorConfig)(atomic.LoadPointer(&s.current))
	return m[platform+":"+field]
}

// All returns all active selector configs.
func (s *SelectorStore) All() []SelectorConfig {
	m := *(*map[string]SelectorConfig)(atomic.LoadPointer(&s.current))
	out := make([]SelectorConfig, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// StartHotReload starts a background goroutine that reloads selectors from the database every minute.
func (s *SelectorStore) StartHotReload(ctx context.Context) {
	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				fresh, err := s.loadFromDB(ctx)
				if err == nil {
					atomic.StorePointer(&s.current, unsafe.Pointer(&fresh))
				}
			}
		}
	}()
}

func (s *SelectorStore) loadFromDB(ctx context.Context) (map[string]SelectorConfig, error) {
	var rows []selectorRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT platform, field, rules FROM selector_configs WHERE active = true`)
	if err != nil {
		return nil, fmt.Errorf("selector store load: %w", err)
	}
	out := make(map[string]SelectorConfig, len(rows))
	for _, r := range rows {
		cfg := r.toConfig()
		out[cfg.Platform+":"+cfg.Field] = cfg
	}
	return out, nil
}
