package queue

import (
	"github.com/hibiken/asynq"
)

func NewServer(redisURL string, crawlConc, normConc int) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisURL},
		asynq.Config{
			Queues: map[string]int{
				"crawl":     crawlConc,
				"normalise": normConc,
			},
		},
	)
}

func RegisterHandlers(mux *asynq.ServeMux, handlers map[string]asynq.Handler) {
	for pattern, h := range handlers {
		mux.Handle(pattern, h)
	}
}
