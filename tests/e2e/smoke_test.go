package e2e_test

func (s *E2ESuite) TestSystemHealth() {
	err := s.DB.Ping()
	s.NoError(err)

	err = s.Redis.Ping(s.Ctx).Err()
	s.NoError(err)

	var selectorCount int
	err = s.DB.Get(&selectorCount, "SELECT COUNT(*) FROM selector_configs WHERE active = true")
	s.NoError(err)
	s.T().Logf("Active selectors: %d", selectorCount)
	s.True(selectorCount > 0, "Should have active selectors")

	queues, err := s.Redis.SMembers(s.Ctx, "asynq:queues").Result()
	s.NoError(err)
	s.T().Logf("Available queues: %v", queues)
}

func (s *E2ESuite) TestDatabaseTablesExist() {
	tables := []string{"products", "raw_reviews", "normalised_reviews", "crawl_jobs", "selector_configs"}

	for _, table := range tables {
		var count int
		err := s.DB.Get(&count, "SELECT COUNT(*) FROM "+table)
		s.NoError(err)
		s.T().Logf("Table %s exists with %d rows", table, count)
	}
}

func (s *E2ESuite) TestRedisConnection() {
	ctx := s.Ctx

	err := s.Redis.Ping(ctx).Err()
	s.NoError(err)

	err = s.Redis.Set(ctx, "test_key", "test_value", 0).Err()
	s.NoError(err)

	val, err := s.Redis.Get(ctx, "test_key").Result()
	s.NoError(err)
	s.Equal("test_value", val)

	_, _ = s.Redis.Del(ctx, "test_key").Result()
}
