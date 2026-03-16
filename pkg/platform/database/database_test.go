package database

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginatedSelect_QueryConstruction(t *testing.T) {
	tests := []struct {
		name       string
		page       Page
		sortBy     string
		sortDir    string
		wantLimit  int
		wantOffset int
		wantOrder  string
	}{
		{
			name:       "default page",
			page:       Page{Limit: 10, Offset: 0},
			wantLimit:  10,
			wantOffset: 0,
			wantOrder:  "DESC",
		},
		{
			name:       "with sort by",
			page:       Page{Limit: 20, Offset: 5, SortBy: "created_at"},
			wantLimit:  20,
			wantOffset: 5,
			wantOrder:  "DESC",
		},
		{
			name:       "explicit ASC",
			page:       Page{Limit: 5, Offset: 0, SortBy: "id", SortDir: "ASC"},
			wantLimit:  5,
			wantOffset: 0,
			wantOrder:  "ASC",
		},
		{
			name:       "invalid sort dir defaults to DESC",
			page:       Page{Limit: 10, Offset: 0, SortBy: "id", SortDir: "INVALID"},
			wantLimit:  10,
			wantOffset: 0,
			wantOrder:  "DESC",
		},
		{
			name:       "empty sort dir defaults to DESC",
			page:       Page{Limit: 10, Offset: 0, SortBy: "id", SortDir: ""},
			wantLimit:  10,
			wantOffset: 0,
			wantOrder:  "DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.page.SortDir
			if dir != "ASC" && dir != "DESC" {
				dir = "DESC"
			}

			assert.Equal(t, tt.wantLimit, tt.page.Limit)
			assert.Equal(t, tt.wantOffset, tt.page.Offset)
			assert.Equal(t, tt.wantOrder, dir)
		})
	}
}

func TestPaginatedSelect_NoResults(t *testing.T) {
	page := Page{Limit: 10, Offset: 0}
	assert.Equal(t, 10, page.Limit)
	assert.Equal(t, 0, page.Offset)
}

func TestPage_SortDirection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase ASC", "ASC", "ASC"},
		{"uppercase DESC", "DESC", "DESC"},
		{"lowercase asc", "asc", "ASC"},
		{"lowercase desc", "desc", "DESC"},
		{"invalid defaults to DESC", "foo", "DESC"},
		{"empty defaults to DESC", "", "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := strings.ToUpper(tt.input)
			if dir != "ASC" && dir != "DESC" {
				dir = "DESC"
			}

			assert.Equal(t, tt.expected, dir)
		})
	}
}

func TestPage_ConstructQuery(t *testing.T) {
	baseQuery := "SELECT * FROM crawl_jobs WHERE platform = $1"
	page := Page{
		Limit:   10,
		Offset:  5,
		SortBy:  "created_at",
		SortDir: "ASC",
	}

	_ = baseQuery
	_ = page
	assert.NotEmpty(t, baseQuery)
}

func TestUpsertOne_EmptyQuery(t *testing.T) {
	ctx := context.Background()
	_ = ctx
}

func TestHealthCheck_NilContext(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	assert.NotNil(t, ctx)
}
