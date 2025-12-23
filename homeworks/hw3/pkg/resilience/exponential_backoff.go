package resilience

import (
	"context"
	"errors"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"math"
	"time"
)

type ExponentialBackoff struct {
	maxQueries              int
	maxClientErrors         int
	maxServerErrors         int
	maxTimeouts             int
	maxLatency              time.Duration
	maxSingleRequestLatency time.Duration
	initialBackoff          time.Duration
	maxBackoff              time.Duration
}

func NewExponentialBackoff(
	maxQueries, maxClientErrors, maxServerErrors, maxTimeouts int,
	maxLatency, maxSingleRequestLatency, initialBackoff, maxBackoff time.Duration,
) *ExponentialBackoff {
	return &ExponentialBackoff{
		maxQueries:              maxQueries,
		maxClientErrors:         maxClientErrors,
		maxServerErrors:         maxServerErrors,
		maxTimeouts:             maxTimeouts,
		maxLatency:              maxLatency,
		maxSingleRequestLatency: maxSingleRequestLatency,
		initialBackoff:          initialBackoff,
		maxBackoff:              maxBackoff,
	}
}

func (b *ExponentialBackoff) Execute(subReq SubRequester) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.maxLatency)
	defer cancel()

	attempts := 0
	clientErrors := b.maxClientErrors
	serverErrors := b.maxServerErrors
	timeouts := b.maxTimeouts

	for {
		select {
		case <-ctx.Done():
		case <-ctx.Done():
			return "", errors.New("context deadline exceeded")
		default:
			subCtx, subCancel := context.WithTimeout(ctx, b.maxSingleRequestLatency)
			payload, err := subReq.Request(subCtx)
			subCancel()

			attempts++
			if err == nil {
				return payload, nil
			}

			var shouldBackoff bool
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
				shouldBackoff = true
			case errors.Is(err, models.ErrTimeout):
				timeouts--
				if timeouts < 0 {
					return "", errors.New("request timeout exceeded")
				}
				shouldBackoff = true
			default:
				return "", err
			}

			if attempts >= b.maxQueries {
				return "", errors.New("max retries limit reached")
			}

			if shouldBackoff {
				backoff := b.initialBackoff * time.Duration(math.Pow(2, float64(attempts-1)))
				if backoff > b.maxBackoff {
					backoff = b.maxBackoff
				}
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return "", errors.New("иbackoff timeout exceeded")
				}
			}
		}
	}
}
