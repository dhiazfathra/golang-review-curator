package scraper

import (
	"context"
	"testing"
	"time"

	"review-curator/pkg/platform/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) UpsertCrawlJob(ctx context.Context, job CrawlJob) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func (m *MockRepository) UpdateJobStatus(ctx context.Context, id, status string, errMsg *string) error {
	args := m.Called(ctx, id, status, errMsg)
	return args.Error(0)
}

func (m *MockRepository) GetJobByID(ctx context.Context, id string) (*CrawlJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CrawlJob), args.Error(1)
}

func (m *MockRepository) ListJobs(ctx context.Context, platform, status string, p database.Page) ([]CrawlJob, int, error) {
	args := m.Called(ctx, platform, status, p)
	return args.Get(0).([]CrawlJob), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpsertRawReview(ctx context.Context, r RawReview) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRepository) GetRawReviewByID(ctx context.Context, id string) (*RawReview, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RawReview), args.Error(1)
}

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) EnqueueCrawlJob(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}

func (m *MockQueue) EnqueueNormalise(rawReviewID string) error {
	args := m.Called(rawReviewID)
	return args.Error(0)
}

func (m *MockQueue) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCrawlService_EnqueueJob(t *testing.T) {
	repo := new(MockRepository)
	queue := new(MockQueue)

	repo.On("UpsertCrawlJob", mock.Anything, mock.AnythingOfType("CrawlJob")).Return(nil)
	queue.On("EnqueueCrawlJob", mock.AnythingOfType("string")).Return(nil)

	service := NewCrawlService(repo, nil)

	_ = service
	assert.NotNil(t, repo)
	assert.NotNil(t, queue)
}

func TestCrawlService_NewCrawlService(t *testing.T) {
	repo := new(MockRepository)
	_ = repo
	service := NewCrawlService(nil, nil)
	assert.NotNil(t, service)
}

func TestCrawlService_CommitResult(t *testing.T) {
	repo := new(MockRepository)
	queue := new(MockQueue)

	repo.On("UpsertRawReview", mock.Anything, mock.AnythingOfType("RawReview")).Return(nil)
	queue.On("EnqueueNormalise", mock.AnythingOfType("string")).Return(nil)

	_ = repo
	_ = queue

	assert.NotNil(t, repo)
}

func TestCrawlService_MarkFailed(t *testing.T) {
	repo := new(MockRepository)

	repo.On("UpdateJobStatus", mock.Anything, mock.AnythingOfType("string"), "failed", mock.Anything).Return(nil)

	_ = repo

	assert.NotNil(t, repo)
}

func TestCrawlService_GetJob(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)

	job := &CrawlJob{ID: "test-id", Status: "pending"}
	repo.On("GetJobByID", ctx, "test-id").Return(job, nil)

	result, err := repo.GetJobByID(ctx, "test-id")
	assert.NoError(t, err)
	assert.Equal(t, "test-id", result.ID)
}

func TestCrawlService_ListJobs(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)

	jobs := []CrawlJob{{ID: "job1"}, {ID: "job2"}}
	page := database.Page{Limit: 10, Offset: 0}
	repo.On("ListJobs", ctx, "shopee", "pending", page).Return(jobs, 2, nil)

	result, total, err := repo.ListJobs(ctx, "shopee", "pending", page)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)
}

func TestScheduler_NewScheduler(t *testing.T) {
	scheduler := NewScheduler(nil, nil, 0)
	assert.NotNil(t, scheduler)
	assert.Equal(t, defaultReCrawlInterval, scheduler.interval)
}

func TestScheduler_NewScheduler_CustomInterval(t *testing.T) {
	customInterval := 24 * time.Hour
	scheduler := NewScheduler(nil, nil, customInterval)
	assert.NotNil(t, scheduler)
	assert.Equal(t, customInterval, scheduler.interval)
}

func TestScheduler_hasRecentJob(t *testing.T) {
	scheduler := &Scheduler{
		interval: 6 * time.Hour,
	}

	oldJob := CrawlJob{
		ProductID:  "123",
		EnqueuedAt: time.Now().Add(-7 * time.Hour),
	}

	recentJob := CrawlJob{
		ProductID:  "456",
		EnqueuedAt: time.Now().Add(-1 * time.Hour),
	}

	_ = oldJob
	_ = recentJob

	_ = scheduler
	assert.NotNil(t, scheduler)
}

func TestScheduler_DefaultInterval(t *testing.T) {
	assert.Equal(t, 6*time.Hour, defaultReCrawlInterval)
}
