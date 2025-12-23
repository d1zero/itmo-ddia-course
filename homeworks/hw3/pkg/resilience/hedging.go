package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Hedging struct {
	maxLatency   time.Duration // Только общий latency budget, нет бюджетов на ошибки
	hedgingDelay time.Duration // Задержка перед отправкой на остальные
}

func NewHedging(maxLatency, hedgingDelay time.Duration) *Hedging {
	return &Hedging{
		maxLatency:   maxLatency,
		hedgingDelay: hedgingDelay,
	}
}

func (h *Hedging) Execute(subReqs []SubRequester) (string, error) {
	if len(subReqs) == 0 {
		return "", errors.New("no sub requests provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.maxLatency)
	defer cancel()

	resultCh := make(chan struct {
		payload string
		err     error
	}, len(subReqs))

	firstSubReq := subReqs[0]
	go func() {
		payload, err := firstSubReq.Request(ctx)
		resultCh <- struct {
			payload string
			err     error
		}{payload, err}
	}()

	select {
	case res := <-resultCh:
		if res.err == nil {
			return res.payload, nil
		}
	case <-time.After(h.hedgingDelay):
	}

	var wg sync.WaitGroup
	for i := 1; i < len(subReqs); i++ {
		wg.Add(1)
		go func(subReq SubRequester) {
			defer wg.Done()
			payload, err := subReq.Request(ctx)
			resultCh <- struct {
				payload string
				err     error
			}{payload, err}
		}(subReqs[i])
	}

	successCount := 0
	var lastErr error
	for i := 0; i < len(subReqs)-1; i++ {
		select {
		case res := <-resultCh:
			if res.err == nil {
				return res.payload, nil
			}
			lastErr = res.err
			successCount++
		case <-ctx.Done():
			return "", errors.New("maxLatency timeout exceeded")
		}
	}

	return "", lastErr
}
