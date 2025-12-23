package mocks

import (
	"context"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"time"
)

type SubRequesterMock struct {
	Delay      time.Duration
	WantRes    string
	WantErr    error
	TotalCalls int64
}

func (m *SubRequesterMock) Request(ctx context.Context) (string, error) {
	m.TotalCalls++
	select {
	case <-time.After(m.Delay):
		if m.WantErr != nil {
			return "", m.WantErr
		}
		return m.WantRes, nil
	case <-ctx.Done():
		return "", models.ErrTimeout
	}
}
