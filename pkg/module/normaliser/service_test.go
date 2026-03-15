package normaliser

import (
	"encoding/json"
	"testing"

	"github.com/pemistahl/lingua-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"review-curator/pkg/module/scraper"
)

func TestExtractShopeeXHR(t *testing.T) {
	svc := &NormaliserService{
		detector: lingua.NewLanguageDetectorBuilder().
			FromLanguages(lingua.Indonesian, lingua.English).
			Build(),
	}
	payload := map[string]any{
		"source": "xhr",
		"body":   `{"data":{"ratings":[{"rating_star":5,"author_username":"shopuser123","comment":"Produk sesuai deskripsi","ctime":1704067200,"item_id":123456}]}}`,
	}
	raw := &scraper.RawReview{
		Platform: scraper.PlatformShopee,
		Payload:  mustMarshal(payload),
	}

	fields, err := svc.extractFields(raw)
	require.NoError(t, err)
	assert.Equal(t, "shopuser123", fields["author_id"])
	assert.Equal(t, "shopuser123", fields["author_name"])
	assert.Equal(t, "Produk sesuai deskripsi", fields["review_text"])
	assert.Equal(t, "5", fields["rating"])
	assert.Equal(t, "123456", fields["product_id"])
}

func TestExtractTokopediaGraphQL(t *testing.T) {
	svc := &NormaliserService{
		detector: lingua.NewLanguageDetectorBuilder().
			FromLanguages(lingua.Indonesian, lingua.English).
			Build(),
	}
	payload := map[string]any{
		"source": "graphql",
		"body":   `{"data":{"productRevGetProductReviewList":{"data":{"list":[{"userId":"12345678","userName":"tokped_user","rating":5,"reviewText":"Bagus sekali","reviewTime":"2024-01-15T10:30:00Z"}]}}}}`,
	}
	raw := &scraper.RawReview{
		Platform: scraper.PlatformTokopedia,
		Payload:  mustMarshal(payload),
	}

	fields, err := svc.extractFields(raw)
	require.NoError(t, err)
	assert.Equal(t, "12345678", fields["author_id"])
	assert.Equal(t, "tokped_user", fields["author_name"])
	assert.Equal(t, "Bagus sekali", fields["review_text"])
	assert.Equal(t, "5", fields["rating"])
}

func TestExtractBlibliXHR(t *testing.T) {
	svc := &NormaliserService{
		detector: lingua.NewLanguageDetectorBuilder().
			FromLanguages(lingua.Indonesian, lingua.English).
			Build(),
	}
	payload := map[string]any{
		"source": "xhr",
		"body":   `{"data":{"reviews":[{"authorId":"9876543210","authorName":"blibliuser","rating":5,"reviewText":"Produk berkualitas","reviewDate":"2024-01-01T00:00:00Z"}]}}`,
	}
	raw := &scraper.RawReview{
		Platform: scraper.PlatformBlibli,
		Payload:  mustMarshal(payload),
	}

	fields, err := svc.extractFields(raw)
	require.NoError(t, err)
	assert.Equal(t, "9876543210", fields["author_id"])
	assert.Equal(t, "blibliuser", fields["author_name"])
	assert.Equal(t, "Produk berkualitas", fields["review_text"])
	assert.Equal(t, "5", fields["rating"])
}

func TestExtractFieldsUnknownSource(t *testing.T) {
	svc := &NormaliserService{
		detector: lingua.NewLanguageDetectorBuilder().
			FromLanguages(lingua.Indonesian, lingua.English).
			Build(),
	}
	payload := map[string]any{
		"source": "unknown",
		"body":   `{}`,
	}
	raw := &scraper.RawReview{
		Platform: scraper.PlatformShopee,
		Payload:  mustMarshal(payload),
	}

	_, err := svc.extractFields(raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown source")
}

func TestParseRating(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Valid 5", "5", 5},
		{"Valid 1", "1", 1},
		{"Valid 3", "3", 3},
		{"Over 5", "10", 5},
		{"Under 1", "-2", 1},
		{"Invalid", "abc", 1},
		{"Empty", "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRating(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComputeHash(t *testing.T) {
	svc := &NormaliserService{}

	hash1 := svc.computeHash("shopee", "123", "user1", "2024-01-01", "great product")
	hash2 := svc.computeHash("shopee", "123", "user1", "2024-01-01", "great product")
	hash3 := svc.computeHash("shopee", "123", "user1", "2024-01-01", "different product")

	assert.Equal(t, hash1, hash2, "same inputs should produce same hash")
	assert.NotEqual(t, hash1, hash3, "different inputs should produce different hash")
}

func TestScoreSentiment(t *testing.T) {
	svc := &NormaliserService{}

	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"Positive", "produk ini bagus dan recommended", 1.0},
		{"Negative", "produk ini jelek dan buruk", -1.0},
		{"Neutral", "produk ini biasa saja", 0.0},
		{"Mixed", "produk ini bagus tapi packaging buruk", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.scoreSentiment(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	svc := &NormaliserService{
		detector: lingua.NewLanguageDetectorBuilder().
			FromLanguages(lingua.Indonesian, lingua.English).
			Build(),
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Indonesian", "Produk ini sangat bagus", "id"},
		{"English", "This product is great", "en"},
		{"Empty", "", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.detectLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func mustMarshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
