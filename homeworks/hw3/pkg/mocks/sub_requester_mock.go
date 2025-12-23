package mocks

import (
	"context"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"time"
)

type SubRequesterMock struct {
	delay   time.Duration
	wantRes string
	wantErr error
}

func NewSubRequesterMock(delay time.Duration, wantRes string, wantErr error) *SubRequesterMock {
	return &SubRequesterMock{
		delay:   delay,
		wantRes: wantRes,
		wantErr: wantErr,
	}
}

func (m *SubRequesterMock) Request(ctx context.Context) (string, error) {
	select {
	case <-time.After(m.delay):
		if m.wantErr != nil {
			return "", m.wantErr
		}
		return m.wantRes, nil
	case <-ctx.Done():
		return "", models.ErrTimeout
	}
}
