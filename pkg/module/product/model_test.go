package product

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProduct_Fields(t *testing.T) {
	now := time.Now()
	product := Product{
		ID:         "prod-123",
		Name:       "Test Product",
		Platform:   "shopee",
		ProductURL: "https://shopee.co.id/product/123",
		ProductID:  "123",
		Active:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	assert.Equal(t, "prod-123", product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "shopee", product.Platform)
	assert.Equal(t, "https://shopee.co.id/product/123", product.ProductURL)
	assert.Equal(t, "123", product.ProductID)
	assert.True(t, product.Active)
	assert.Equal(t, now, product.CreatedAt)
	assert.Equal(t, now, product.UpdatedAt)
}

func TestProduct_DefaultValues(t *testing.T) {
	product := Product{}

	assert.Empty(t, product.ID)
	assert.Empty(t, product.Name)
	assert.Empty(t, product.Platform)
	assert.Empty(t, product.ProductURL)
	assert.Empty(t, product.ProductID)
	assert.False(t, product.Active)
}

func TestProduct_SetActive(t *testing.T) {
	product := Product{Active: false}
	assert.False(t, product.Active)

	product.Active = true
	assert.True(t, product.Active)
}

func TestProduct_Inactive(t *testing.T) {
	product := Product{
		ID:         "prod-456",
		Name:       "Inactive Product",
		Platform:   "tokopedia",
		ProductURL: "https://tokopedia.com/product/456",
		ProductID:  "456",
		Active:     false,
	}

	assert.Equal(t, "prod-456", product.ID)
	assert.False(t, product.Active)
}

func TestProduct_Timestamps(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	now := time.Now()

	product := Product{
		CreatedAt: past,
		UpdatedAt: now,
	}

	assert.True(t, product.CreatedAt.Before(product.UpdatedAt))
}

func TestNewRepository(t *testing.T) {
	repo := NewRepository(nil)
	assert.NotNil(t, repo)
}

func TestProduct_String(t *testing.T) {
	product := Product{
		ID:         "test-id",
		Name:       "Test Name",
		Platform:   "blibli",
		ProductURL: "https://blibli.com/product/789",
		ProductID:  "789",
	}

	assert.Contains(t, product.ProductURL, "blibli")
	assert.Equal(t, "test-id", product.ID)
}
