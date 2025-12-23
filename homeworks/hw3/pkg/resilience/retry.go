package resilience

import (
	"context"
	"errors"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"time"
)

type Retry struct {
	maxQueries              int
	maxClientErrors         int
	maxServerErrors         int
	maxTimeouts             int
	maxLatency              time.Duration
	maxSingleRequestLatency time.Duration
}

func NewRetry(
	maxQueries, maxClientErrors, maxServerErrors, maxTimeouts int,
	maxLatency, maxSingleRequestLatency time.Duration,
) *Retry {
	return &Retry{
		maxQueries:              maxQueries,
		maxClientErrors:         maxClientErrors,
		maxServerErrors:         maxServerErrors,
		maxTimeouts:             maxTimeouts,
		maxLatency:              maxLatency,
		maxSingleRequestLatency: maxSingleRequestLatency,
	}
}

func (r *Retry) Execute(subReq SubRequester) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.maxLatency)
	defer cancel()

	attempts := 0
	clientErrors := r.maxClientErrors
	serverErrors := r.maxServerErrors
	timeouts := r.maxTimeouts

	for {
		select {
		case <-ctx.Done():
			return "", errors.New("context deadline exceeded")
		default:
			subCtx, subCancel := context.WithTimeout(ctx, r.maxSingleRequestLatency)
			payload, err := subReq.Request(subCtx)
			subCancel()

			attempts++
			if err == nil {
				return payload, nil
			}

			switch {
			case errors.Is(err, models.ErrClientError):
				clientErrors--
				if clientErrors < 0 {
					return "", errors.New("client errors limit reached")
				}
			case errors.Is(err, models.ErrServerError):
				serverErrors--
				if serverErrors < 0 {
					return "", errors.New("server errors limit reached")
				}
			case errors.Is(err, models.ErrTimeout):
				timeouts--
				if timeouts < 0 {
					return "", errors.New("request timeout exceeded")
				}
			default:
				return "", err
			}

			if attempts >= r.maxQueries {
				return "", errors.New("max retries limit reached")
			}
		}
	}
}
