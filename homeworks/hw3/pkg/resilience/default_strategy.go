package resilience

import (
	"context"
	"errors"
	"time"
)

// DefaultStrategy Базовая стратегия по умолчанию.
type DefaultStrategy struct {
	maxLatency time.Duration
}

func NewDefaultStrategy(maxLatency time.Duration) *DefaultStrategy {
	return &DefaultStrategy{
		maxLatency: maxLatency,
	}
}

func (s *DefaultStrategy) Execute(subReq SubRequester) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.maxLatency)
	defer cancel()
	return subReq.Request(ctx)
}

// DefaultMultiStrategy Базовая стратегия по умолчанию для мультиклиента.
type DefaultMultiStrategy struct {
	maxLatency time.Duration
}

func NewDefaultMultiStrategy(maxLatency time.Duration) *DefaultMultiStrategy {
	return &DefaultMultiStrategy{
		maxLatency: maxLatency,
	}
}

func (s *DefaultMultiStrategy) Execute(subReqs []SubRequester) (string, error) {
	if len(subReqs) == 0 {
		return "", errors.New("no sub requests provided")
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.maxLatency)
	defer cancel()
	return subReqs[0].Request(ctx)
}
