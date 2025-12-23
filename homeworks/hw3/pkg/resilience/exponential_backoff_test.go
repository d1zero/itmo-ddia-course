package resilience

import (
	"errors"
	"fmt"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"github.com/d1zero/itmo-ddia-course/lab3/pkg/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	type exponentialBackoffArgs struct {
		maxQueries              int
		maxClientErrors         int
		maxServerErrors         int
		maxTimeouts             int
		maxLatency              time.Duration
		maxSingleRequestLatency time.Duration
		initialBackoff          time.Duration
		maxBackoff              time.Duration
	}

	tests := map[string]struct {
		exponentialBackoffArgs exponentialBackoffArgs
		mockDelay              time.Duration
		mockBehaviours         []func(*mocks.SubRequesterMock)
		wantResult             string
		wantErr                error
		extraDurationCheck     func(time.Duration) error
	}{
		"success": {
			exponentialBackoffArgs: exponentialBackoffArgs{
				maxQueries:              3,
				maxClientErrors:         2,
				maxServerErrors:         2,
				maxTimeouts:             2,
				maxLatency:              1000 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
				initialBackoff:          100 * time.Millisecond,
				maxBackoff:              400 * time.Millisecond,
			},
			mockDelay: 50 * time.Millisecond,
			mockBehaviours: []func(*mocks.SubRequesterMock){
				func(m *mocks.SubRequesterMock) {
					m.WantErr = models.ErrServerError
				},
				func(m *mocks.SubRequesterMock) {
					m.WantErr = models.ErrServerError
				},
				func(m *mocks.SubRequesterMock) {
					m.WantErr = nil
					m.WantRes = "success"
				},
			},
			wantResult: "success",
			wantErr:    nil,
			extraDurationCheck: func(duration time.Duration) error {
				// Примерно backoff 100 + 200
				if duration < 300*time.Millisecond {
					return fmt.Errorf("недостаточная задержка: %v", duration)
				}
				return nil
			},
		},
		"client error - no backoff": {
			exponentialBackoffArgs: exponentialBackoffArgs{
				maxQueries:              3,
				maxClientErrors:         2,
				maxServerErrors:         2,
				maxTimeouts:             2,
				maxLatency:              1000 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
				initialBackoff:          100 * time.Millisecond,
				maxBackoff:              400 * time.Millisecond,
			},
			mockDelay: 50 * time.Millisecond,
			mockBehaviours: []func(*mocks.SubRequesterMock){
				func(m *mocks.SubRequesterMock) {
					m.WantErr = models.ErrClientError
				},
				func(m *mocks.SubRequesterMock) {
					m.WantErr = nil
					m.WantRes = "success"
				},
			},
			wantResult: "success",
			wantErr:    nil,
			extraDurationCheck: func(duration time.Duration) error {
				// Без backoff, только delay(210 проходит локально)
				if duration > 210*time.Millisecond {
					return fmt.Errorf("лишняя задержка: %v", duration)
				}
				return nil
			},
		},
		"max latency reached": {
			exponentialBackoffArgs: exponentialBackoffArgs{
				maxQueries:              3,
				maxClientErrors:         2,
				maxServerErrors:         2,
				maxTimeouts:             2,
				maxLatency:              100 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
				initialBackoff:          100 * time.Millisecond,
				maxBackoff:              400 * time.Millisecond,
			},
			mockDelay: 50 * time.Millisecond,
			mockBehaviours: []func(*mocks.SubRequesterMock){
				func(m *mocks.SubRequesterMock) {
					m.WantErr = models.ErrServerError
				},
			},
			wantResult: "",
			wantErr:    errors.New("backoff timeout exceeded"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			strategy := NewExponentialBackoff(tt.exponentialBackoffArgs.maxQueries, tt.exponentialBackoffArgs.maxClientErrors,
				tt.exponentialBackoffArgs.maxServerErrors, tt.exponentialBackoffArgs.maxTimeouts,
				tt.exponentialBackoffArgs.maxLatency, tt.exponentialBackoffArgs.maxSingleRequestLatency,
				tt.exponentialBackoffArgs.initialBackoff, tt.exponentialBackoffArgs.maxBackoff,
			)

			var (
				result string
				err    error
			)

			mock := mocks.SubRequesterMock{
				WantRes: tt.wantResult,
				WantErr: tt.wantErr,
				Delay:   tt.mockDelay,
			}
			start := time.Now()

			for _, mockBehaviour := range tt.mockBehaviours {
				mockBehaviour(&mock)
				result, err = strategy.Execute(&mock)
			}

			duration := time.Since(start)

			require.Equal(t, tt.wantResult, result)
			require.Equal(t, tt.wantErr, err)
			if tt.extraDurationCheck != nil {
				require.NoError(t, tt.extraDurationCheck(duration))
			}
		})
	}
}
