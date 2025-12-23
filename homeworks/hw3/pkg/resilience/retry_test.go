package resilience

import (
	"errors"
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"github.com/d1zero/itmo-ddia-course/lab3/pkg/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	type retryArgs struct {
		maxQueries      int
		maxClientErrors int
		maxServerErrors int
		maxTimeouts     int

		maxLatency              time.Duration
		maxSingleRequestLatency time.Duration
	}

	tests := map[string]struct {
		retryArgs   retryArgs
		subReqMocks []SubRequester
		wantResult  string
		wantErr     error
	}{
		"success after retry": {
			retryArgs: retryArgs{
				maxQueries:              2,
				maxClientErrors:         1,
				maxServerErrors:         1,
				maxTimeouts:             1,
				maxLatency:              200 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
			},
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrServerError},
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantRes: "success"},
			},
			wantResult: "success",
			wantErr:    nil,
		},
		"client errors limit reached": {
			retryArgs: retryArgs{
				maxQueries:              3,
				maxClientErrors:         1,
				maxServerErrors:         1,
				maxTimeouts:             1,
				maxLatency:              300 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
			},
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrClientError},
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrClientError},
			},
			wantResult: "",
			wantErr:    errors.New("client errors limit reached"),
		},
		"server errors limit reached": {
			retryArgs: retryArgs{
				maxQueries:              3,
				maxClientErrors:         1,
				maxServerErrors:         1,
				maxTimeouts:             1,
				maxLatency:              300 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
			},
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrServerError},
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrServerError},
			},
			wantResult: "",
			wantErr:    errors.New("server errors limit reached"),
		},
		"timeout exceeded": {
			retryArgs: retryArgs{
				maxQueries:              3,
				maxClientErrors:         1,
				maxServerErrors:         1,
				maxTimeouts:             1,
				maxLatency:              300 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
			},
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrTimeout},
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrTimeout},
			},
			wantResult: "",
			wantErr:    errors.New("request timeout exceeded"),
		},
		"retries limit reached": {
			retryArgs: retryArgs{
				maxQueries:              1,
				maxClientErrors:         1,
				maxServerErrors:         1,
				maxTimeouts:             1,
				maxLatency:              300 * time.Millisecond,
				maxSingleRequestLatency: 100 * time.Millisecond,
			},
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: 50 * time.Millisecond, WantErr: models.ErrTimeout},
			},
			wantResult: "",
			wantErr:    errors.New("max retries limit reached"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			strategy := NewRetry(tt.retryArgs.maxQueries, tt.retryArgs.maxClientErrors, tt.retryArgs.maxServerErrors,
				tt.retryArgs.maxTimeouts, tt.retryArgs.maxLatency, tt.retryArgs.maxSingleRequestLatency)

			var (
				result string
				err    error
			)

			for _, mock := range tt.subReqMocks {
				result, err = strategy.Execute(mock)
			}

			require.Equal(t, tt.wantResult, result)
			require.Equal(t, tt.wantErr, err)
		})
	}
}
