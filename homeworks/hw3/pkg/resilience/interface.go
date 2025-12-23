package resilience

import "context"

// SubRequester Интерфейс для отправки подзапросов к серверу.
type SubRequester interface {
	Request(ctx context.Context) (string, error)
}

// ResilienceStrategy Интерфейс для стратегии устойчивости для одного клиента.
type ResilienceStrategy interface {
	Execute(subReq SubRequester) (string, error)
}

// MultiResilienceStrategy Интерфейс для стратегии устойчивости для мультиклиента.
type MultiResilienceStrategy interface {
	Execute(subReqs []SubRequester) (string, error)
}
