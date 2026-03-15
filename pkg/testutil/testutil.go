package testutil

import (
	"context"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	DB      *sqlx.DB
	Redis   *redis.Client
	Asynq   *asynq.Client
	Cleanup func()
	Ctx     context.Context
}

func NewE2ETestSuite() *E2ETestSuite {
	return &E2ETestSuite{}
}

func (s *E2ETestSuite) SetupSuite() {
	s.Ctx = context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://curator:curator@localhost:5432/review_curator?sslmode=disable"
	}
	db, err := sqlx.Connect("postgres", dbURL)
	s.Require().NoError(err)
	s.DB = db

	err = goose.Up(db.DB, "migrations")
	s.Require().NoError(err)

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	opts, err := redis.ParseURL(redisURL)
	s.Require().NoError(err)
	s.Redis = redis.NewClient(opts)
	s.Require().NoError(s.Redis.Ping(s.Ctx).Err())

	s.Asynq = asynq.NewClient(asynq.RedisClientOpt{Addr: getRedisAddr(redisURL)})

	s.Cleanup = func() {
		_ = db.Close()
		_ = s.Redis.Close()
		_ = s.Asynq.Close()
	}
}

func getRedisAddr(redisURL string) string {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return "localhost:6379"
	}
	return opts.Addr
}

func (s *E2ETestSuite) SetupTest() {
	s.DB.MustExec("DELETE FROM normalised_reviews")
	s.DB.MustExec("DELETE FROM raw_reviews")
	s.DB.MustExec("DELETE FROM crawl_jobs")
	s.DB.MustExec("DELETE FROM products")

	s.Redis.FlushDB(s.Ctx)
}

func (s *E2ETestSuite) TeardownSuite() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func SetupTestDB(t interface{ Helper() }) (func(), *sqlx.DB, error) {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://curator:curator@localhost:5432/review_curator?sslmode=disable"
	}
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return nil, nil, err
	}

	err = goose.Up(db.DB, "migrations")
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	cleanup := func() { _ = db.Close() }
	return cleanup, db, nil
}

func SeedSelectors(db *sqlx.DB) {
	selectors := []struct {
		platform, field, selectorType, value string
		active                               bool
	}{
		{"shopee", "review_text", "css", ".shopee-review-card__main-content", true},
		{"shopee", "rating", "css", ".shopee-review-card__rating", true},
		{"shopee", "author_name", "css", ".shopee-review-card__author", true},
		{"tokopedia", "review_text", "css", ".tokopedia-review-card__content", true},
		{"tokopedia", "rating", "css", ".tokopedia-review-card__rating", true},
		{"blibli", "review_text", "css", ".blibli-review-card__content", true},
		{"blibli", "rating", "css", ".blibli-review-card__rating", true},
	}

	for _, sel := range selectors {
		db.MustExec(`
			INSERT INTO selector_configs (id, platform, field, type, value, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (platform, field, type) DO NOTHING
		`, generateUUID(), sel.platform, sel.field, sel.selectorType, sel.value, sel.active, time.Now(), time.Now())
	}
}

func generateUUID() string {
	return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"[0:36]
}
