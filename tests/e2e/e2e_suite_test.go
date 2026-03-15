package e2e_test

import (
	"context"
	"os"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type E2ESuite struct {
	suite.Suite
	DB      *sqlx.DB
	Redis   *redis.Client
	Asynq   *asynq.Client
	Cleanup func()
	Ctx     context.Context
}

func (s *E2ESuite) SetupSuite() {
	if testing.Short() {
		s.T().Skip("Skipping E2E tests in short mode")
	}

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

func (s *E2ESuite) SetupTest() {
	s.DB.MustExec("DELETE FROM normalised_reviews")
	s.DB.MustExec("DELETE FROM raw_reviews")
	s.DB.MustExec("DELETE FROM crawl_jobs")
	s.DB.MustExec("DELETE FROM products")

	s.Redis.FlushDB(s.Ctx)
}

func (s *E2ESuite) TeardownSuite() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}
