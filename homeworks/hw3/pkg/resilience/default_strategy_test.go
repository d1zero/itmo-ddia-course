package resilience

import (
	"github.com/d1zero/itmo-ddia-course/lab3/models"
	"github.com/d1zero/itmo-ddia-course/lab3/pkg/mocks"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestDefaultStrategy(t *testing.T) {
	maxLatency := 200 * time.Millisecond

	tests := map[string]struct {
		delay      time.Duration
		wantResult string
		wantErr    error
	}{
		"success": {
			delay:      maxLatency - maxLatency/2,
			wantResult: "success",
			wantErr:    nil,
		},
		"timeout": {
			delay:   maxLatency * maxLatency / 2,
			wantErr: models.ErrTimeout,
		},
		"client error": {
			delay:   maxLatency - maxLatency/2,
			wantErr: models.ErrClientError,
		},
		"server error": {
			delay:   maxLatency - maxLatency/2,
			wantErr: models.ErrServerError,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			subReqMock := &mocks.SubRequesterMock{Delay: tt.delay, WantRes: tt.wantResult, WantErr: tt.wantErr}

			strategy := NewDefaultStrategy(maxLatency)

			result, err := strategy.Execute(subReqMock)
			require.Equal(t, tt.wantResult, result)
			require.Equal(t, tt.wantErr, err)
		})
	}
}

func TestDefaultMultiStrategy(t *testing.T) {
	maxLatency := 200 * time.Millisecond
	tests := map[string]struct {
		subReqMocks []SubRequester
		wantResult  string
		wantErr     error
	}{
		"success": {
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: maxLatency - maxLatency/2, WantRes: "success"},
			},
			wantResult: "success",
			wantErr:    nil,
		},
		"timeout": {
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: maxLatency / 2, WantErr: models.ErrTimeout},
			},
			wantErr: models.ErrTimeout,
		},
		"client error": {
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: maxLatency - maxLatency/2, WantErr: models.ErrClientError},
			},
			wantErr: models.ErrClientError,
		},
		"server error": {
			subReqMocks: []SubRequester{
				&mocks.SubRequesterMock{Delay: maxLatency - maxLatency/2, WantErr: models.ErrServerError},
			},
			wantErr: models.ErrServerError,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			strategy := NewDefaultMultiStrategy(maxLatency)

			result, err := strategy.Execute(tt.subReqMocks)
			require.Equal(t, tt.wantResult, result)
			require.Equal(t, tt.wantErr, err)
		})
	}
}
