package resilience

import (
	"errors"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"github.com/d1zero/itmo-ddia-course/lab3/pkg/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestHedging(t *testing.T) {
	type hedgingArgs struct {
		maxLatency   time.Duration
		hedgingDelay time.Duration
	}

	tests := map[string]struct {
		hedgingArgs   hedgingArgs
		mocks         []SubRequester
		wantMockCalls []int64
		wantResult    string
		wantErr       error
	}{
		"first slow, than fast": {
			hedgingArgs: hedgingArgs{
				maxLatency:   500 * time.Millisecond,
				hedgingDelay: 100 * time.Millisecond,
			},
			mocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 300 * time.Millisecond, WantRes: "slow"},
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantRes: "fast"},
			},
			wantResult:    "fast",
			wantMockCalls: []int64{1, 1},
		},
		"first fast, than not used": {
			hedgingArgs: hedgingArgs{
				maxLatency:   500 * time.Millisecond,
				hedgingDelay: 100 * time.Millisecond,
			},
			mocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantRes: "fast"},
				&mocks.SubRequesterMock{Delay: 300 * time.Millisecond, WantRes: "slow"},
			},
			wantResult:    "fast",
			wantMockCalls: []int64{1, 0},
		},
		"all errors": {
			hedgingArgs: hedgingArgs{
				maxLatency:   500 * time.Millisecond,
				hedgingDelay: 100 * time.Millisecond,
			},
			mocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrServerError},
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrServerError},
			},
			wantErr: models.ErrServerError,
		},
		"maxLatency timeout exceeded": {
			hedgingArgs: hedgingArgs{
				maxLatency:   200 * time.Millisecond,
				hedgingDelay: 100 * time.Millisecond,
			},
			mocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 600 * time.Millisecond},
				&mocks.SubRequesterMock{Delay: 600 * time.Millisecond},
			},
			wantErr: errors.New("maxLatency timeout exceeded"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			strategy := NewHedging(tt.hedgingArgs.maxLatency, tt.hedgingArgs.hedgingDelay)

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
