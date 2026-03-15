package normaliser

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type NormaliserHandler struct {
	service *NormaliserService
}

func NewNormaliserHandler(service *NormaliserService) *NormaliserHandler {
	return &NormaliserHandler{service: service}
}

func (h *NormaliserHandler) Handle(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		RawReviewID string `json:"raw_review_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("normaliser handler: unmarshal: %w", err)
	}
	return h.service.Process(ctx, payload.RawReviewID)
}
