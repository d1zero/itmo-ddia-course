package resilience

import (
	"errors"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"github.com/d1zero/itmo-ddia-course/lab3/pkg/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRoundRobin(t *testing.T) {
	baseMockDelay := 50 * time.Millisecond

	type roundRobinArgs struct {
		maxQueries      int
		maxClientErrors int
		maxServerErrors int
		maxTimeouts     int

		maxLatency              time.Duration
		maxSingleRequestLatency time.Duration
	}

	tests := map[string]struct {
		roundRobinArgs roundRobinArgs
		mocks          []SubRequester
		wantMockCalls  []int64
		wantResult     string
		wantErr        error
	}{
		"success": {
			roundRobinArgs: roundRobinArgs{
				maxQueries:              3,
				maxClientErrors:         2,
				maxServerErrors:         2,
				maxTimeouts:             2,
				maxLatency:              500 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
			},
			mocks: []SubRequester{
				&mocks.SubRequesterMock{WantErr: models.ErrServerError, Delay: baseMockDelay},
				&mocks.SubRequesterMock{WantErr: models.ErrServerError, Delay: baseMockDelay},
				&mocks.SubRequesterMock{WantRes: "success", Delay: baseMockDelay},
			},
			wantResult:    "success",
			wantMockCalls: []int64{1, 1, 1},
		},
		"retry max queries exceeded": {
			roundRobinArgs: roundRobinArgs{
				maxQueries:              2,
				maxClientErrors:         2,
				maxServerErrors:         2,
				maxTimeouts:             2,
				maxLatency:              500 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
			},
			mocks: []SubRequester{
				&mocks.SubRequesterMock{WantErr: models.ErrServerError, Delay: baseMockDelay},
				&mocks.SubRequesterMock{WantErr: models.ErrServerError, Delay: baseMockDelay},
			},
			wantErr: errors.New("max retries limit reached"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			strategy := NewRoundRobin(
				tt.roundRobinArgs.maxQueries, tt.roundRobinArgs.maxClientErrors,
				tt.roundRobinArgs.maxServerErrors, tt.roundRobinArgs.maxTimeouts,
				tt.roundRobinArgs.maxLatency, tt.roundRobinArgs.maxSingleRequestLatency,
			)

			var (
				result string
				err    error
			)

			result, err = strategy.Execute(tt.mocks)

			require.Equal(t, tt.wantResult, result)
			require.Equal(t, tt.wantErr, err)
			if tt.wantMockCalls != nil {
				for idx, mock := range tt.mocks {
					subMock := mock.(*mocks.SubRequesterMock)
					require.Equalf(t, tt.wantMockCalls[idx], subMock.TotalCalls, "mock %d calls are not equal", idx)
				}
			}
		})
	}
}
