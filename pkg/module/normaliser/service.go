package normaliser

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/pemistahl/lingua-go"
	"review-curator/pkg/module/scraper"
)

type NormaliserService struct {
	rawRepo interface {
		GetRawReviewByID(ctx context.Context, id string) (*scraper.RawReview, error)
	}
	normRepo Repository
	detector lingua.LanguageDetector
}

func NewNormaliserService(
	rawRepo interface {
		GetRawReviewByID(ctx context.Context, id string) (*scraper.RawReview, error)
	},
	normRepo Repository,
) *NormaliserService {
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(lingua.Indonesian, lingua.English).
		Build()
	return &NormaliserService{rawRepo: rawRepo, normRepo: normRepo, detector: detector}
}

func (s *NormaliserService) Process(ctx context.Context, rawReviewID string) error {
	raw, err := s.rawRepo.GetRawReviewByID(ctx, rawReviewID)
	if err != nil {
		return fmt.Errorf("normaliser: get raw review: %w", err)
	}

	fields, err := s.extractFields(raw)
	if err != nil {
		return fmt.Errorf("normaliser: extract fields: %w", err)
	}

	lang := s.detectLanguage(fields["review_text"])
	sentiment := s.scoreSentiment(fields["review_text"])
	dedupeHash := s.computeHash(string(raw.Platform), fields["product_id"],
		fields["author_id"], fields["reviewed_at"], fields["review_text"])

	reviewedAt, _ := time.Parse(time.RFC3339, fields["reviewed_at"])
	if reviewedAt.IsZero() {
		reviewedAt = raw.CrawledAt
	}

	rating := parseRating(fields["rating"])

	nr := NormalisedReview{
		ID:             uuid.New().String(),
		RawReviewID:    raw.ID,
		Platform:       string(raw.Platform),
		ProductID:      fields["product_id"],
		AuthorID:       fields["author_id"],
		AuthorName:     fields["author_name"],
		Rating:         rating,
		ReviewText:     fields["review_text"],
		Language:       lang,
		SentimentScore: sentiment,
		ReviewedAt:     reviewedAt,
		NormalisedAt:   time.Now(),
		DedupeHash:     dedupeHash,
	}

	return s.normRepo.UpsertNormalisedReview(ctx, nr)
}

func (s *NormaliserService) extractFields(raw *scraper.RawReview) (map[string]string, error) {
	var wrapper map[string]any
	if err := json.Unmarshal(raw.Payload, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}
	source, _ := wrapper["source"].(string)
	switch source {
	case "xhr", "graphql":
		return extractFromXHR(raw.Platform, wrapper)
	case "dom":
		fields, _ := wrapper["fields"].(map[string]any)
		out := make(map[string]string, len(fields))
		for k, v := range fields {
			out[k], _ = v.(string)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unknown source: %s", source)
	}
}

func extractFromXHR(platform scraper.Platform, wrapper map[string]any) (map[string]string, error) {
	bodyStr, _ := wrapper["body"].(string)
	switch platform {
	case scraper.PlatformShopee:
		return extractShopeeXHR(bodyStr)
	case scraper.PlatformTokopedia:
		return extractTokopediaGraphQL(bodyStr)
	case scraper.PlatformBlibli:
		return extractBlibliXHR(bodyStr)
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}
}

func extractShopeeXHR(body string) (map[string]string, error) {
	var doc struct {
		Data struct {
			Ratings []struct {
				AuthorUsername string `json:"author_username"`
				Comment        string `json:"comment"`
				RatingStar     int    `json:"rating_star"`
				Ctime          int64  `json:"ctime"`
				ItemID         int64  `json:"item_id"`
			} `json:"ratings"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &doc); err != nil || len(doc.Data.Ratings) == 0 {
		return map[string]string{}, nil
	}
	r := doc.Data.Ratings[0]
	reviewedAt := time.Now().Format(time.RFC3339)
	if r.Ctime > 0 {
		reviewedAt = time.Unix(r.Ctime, 0).Format(time.RFC3339)
	}
	productID := ""
	if r.ItemID > 0 {
		productID = fmt.Sprintf("%d", r.ItemID)
	}
	return map[string]string{
		"author_id":   r.AuthorUsername,
		"author_name": r.AuthorUsername,
		"review_text": r.Comment,
		"rating":      fmt.Sprintf("%d", r.RatingStar),
		"reviewed_at": reviewedAt,
		"product_id":  productID,
	}, nil
}

func extractTokopediaGraphQL(body string) (map[string]string, error) {
	var doc struct {
		Data struct {
			ProductRevGetProductReviewList struct {
				Data struct {
					List []struct {
						UserID     string `json:"userId"`
						UserName   string `json:"userName"`
						Rating     int    `json:"rating"`
						ReviewText string `json:"reviewText"`
						ReviewTime string `json:"reviewTime"`
					} `json:"list"`
				} `json:"data"`
			} `json:"productRevGetProductReviewList"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		return map[string]string{}, nil
	}
	list := doc.Data.ProductRevGetProductReviewList.Data.List
	if len(list) == 0 {
		return map[string]string{}, nil
	}
	r := list[0]
	return map[string]string{
		"author_id":   r.UserID,
		"author_name": r.UserName,
		"review_text": r.ReviewText,
		"rating":      fmt.Sprintf("%d", r.Rating),
		"reviewed_at": r.ReviewTime,
		"product_id":  "",
	}, nil
}

func extractBlibliXHR(body string) (map[string]string, error) {
	var doc struct {
		Data struct {
			Reviews []struct {
				AuthorID   string  `json:"authorId"`
				AuthorName string  `json:"authorName"`
				Rating     float64 `json:"rating"`
				ReviewText string  `json:"reviewText"`
				ReviewDate string  `json:"reviewDate"`
			} `json:"reviews"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &doc); err != nil || len(doc.Data.Reviews) == 0 {
		return map[string]string{}, nil
	}
	r := doc.Data.Reviews[0]
	return map[string]string{
		"author_id":   r.AuthorID,
		"author_name": r.AuthorName,
		"review_text": r.ReviewText,
		"rating":      fmt.Sprintf("%g", r.Rating),
		"reviewed_at": r.ReviewDate,
		"product_id":  "",
	}, nil
}

func (s *NormaliserService) detectLanguage(text string) string {
	lang, ok := s.detector.DetectLanguageOf(text)
	if !ok {
		return "unknown"
	}
	switch lang {
	case lingua.Indonesian:
		return "id"
	case lingua.English:
		return "en"
	default:
		return "unknown"
	}
}

func (s *NormaliserService) scoreSentiment(text string) float64 {
	positive := []string{"bagus", "mantap", "recommended", "puas", "good", "great", "excellent", "best"}
	negative := []string{"jelek", "buruk", "kecewa", "rusak", "bad", "poor", "terrible", "broken"}
	lower := strings.ToLower(text)
	pos, neg := 0, 0
	for _, w := range positive {
		if strings.Contains(lower, w) {
			pos++
		}
	}
	for _, w := range negative {
		if strings.Contains(lower, w) {
			neg++
		}
	}
	total := pos + neg
	if total == 0 {
		return 0
	}
	return float64(pos-neg) / float64(total)
}

func (s *NormaliserService) computeHash(platform, productID, authorID, reviewedAt, reviewText string) string {
	text := reviewText
	if utf8.RuneCountInString(text) > 100 {
		runes := []rune(text)
		text = string(runes[:100])
	}
	sum := sha256.Sum256([]byte(platform + productID + authorID + reviewedAt + text))
	return fmt.Sprintf("%x", sum)
}

func parseRating(s string) int {
	var r int
	if _, err := fmt.Sscanf(s, "%d", &r); err != nil {
		return 1
	}
	if r < 1 {
		r = 1
	}
	if r > 5 {
		r = 5
	}
	return r
}
