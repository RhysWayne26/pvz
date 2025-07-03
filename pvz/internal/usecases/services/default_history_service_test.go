package services

import (
	"context"
	"errors"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/data/repositories/mocks"
	"pvz-cli/pkg/clock"
	"pvz-cli/tests/builders"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

// TestDefaultHistoryService_Record_Success verifies that the Record method of DefaultHistoryService saves an entry successfully.
func TestDefaultHistoryService_Record_Success(t *testing.T) {
	t.Parallel()
	clkForHistory := &clock.FakeClock{}
	mockRepo := mocks.NewHistoryRepositoryMock(t)
	svc := NewDefaultHistoryService(mockRepo)
	ctx := context.Background()
	entry := builders.NewHistoryEntryBuilder(clkForHistory).
		WithOrderID(52).
		WithEvent(models.EventIssued).
		WithOffset(time.Minute).
		Build()
	mockRepo.SaveMock.Expect(ctx, entry).Return(nil)
	require.NoError(t, svc.Record(ctx, entry))
}

// TestDefaultHistoryService_Record_SaveFails verifies behavior when saving a history record fails due to repository error.
func TestDefaultHistoryService_Record_SaveFails(t *testing.T) {
	t.Parallel()
	clkForHistory := &clock.FakeClock{}
	mockRepo := mocks.NewHistoryRepositoryMock(t)
	svc := NewDefaultHistoryService(mockRepo)
	ctx := context.Background()
	entry := builders.NewHistoryEntryBuilder(clkForHistory).
		WithOrderID(52).
		WithEvent(models.EventIssued).
		WithOffset(time.Minute).
		Build()
	mockRepo.SaveMock.Expect(ctx, entry).Return(errors.New("db unreachable"))
	err := svc.Record(ctx, entry)
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	require.Equal(t, apperrors.InternalError, ae.Code)
}

// TestDefaultHistoryService_List validates the behavior of the List method in DefaultHistoryService with various scenarios.
func TestDefaultHistoryService_List(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		filter       requests.OrderHistoryFilter
		mockReturn   []models.HistoryEntry
		mockCount    int
		mockErr      error
		wantEntries  []models.HistoryEntry
		wantErrCode  *apperrors.ErrorCode
		wantPlainErr bool
	}{
		{
			name:        "success with ID",
			filter:      requests.OrderHistoryFilter{OrderID: utils.Ptr(uint64(52))},
			mockReturn:  []models.HistoryEntry{entry(52, models.EventAccepted, 0)},
			mockCount:   1,
			wantEntries: []models.HistoryEntry{entry(52, models.EventAccepted, 0)},
		},
		{
			name:        "order not found",
			filter:      requests.OrderHistoryFilter{OrderID: utils.Ptr(uint64(52))},
			mockReturn:  nil,
			mockCount:   0,
			wantErrCode: utils.Ptr(apperrors.OrderNotFound),
		},
		{
			name:        "db error",
			filter:      requests.OrderHistoryFilter{OrderID: utils.Ptr(uint64(52))},
			mockReturn:  nil,
			mockCount:   0,
			mockErr:     errors.New("db fail"),
			wantErrCode: utils.Ptr(apperrors.InternalError),
		},
		{
			name:         "context cancelled",
			filter:       requests.OrderHistoryFilter{},
			wantPlainErr: true,
		},
		{
			name:        "no orderID returns empty",
			filter:      requests.OrderHistoryFilter{Page: 1, Limit: 10},
			mockReturn:  nil,
			mockCount:   0,
			wantEntries: []models.HistoryEntry(nil),
		},
		{
			name:        "with pagination",
			filter:      requests.OrderHistoryFilter{Page: 2, Limit: 5},
			mockReturn:  []models.HistoryEntry{entry(52, models.EventIssued, time.Minute)},
			mockCount:   1,
			wantEntries: []models.HistoryEntry{entry(52, models.EventIssued, time.Minute)},
		},
		{
			name:        "with both ID and pagination",
			filter:      requests.OrderHistoryFilter{OrderID: utils.Ptr(uint64(52)), Page: 2, Limit: 5},
			mockReturn:  []models.HistoryEntry{entry(52, models.EventReturnedByClient, 2*time.Minute)},
			mockCount:   1,
			wantEntries: []models.HistoryEntry{entry(52, models.EventReturnedByClient, 2*time.Minute)},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockRepo := mocks.NewHistoryRepositoryMock(t)
			svc := NewDefaultHistoryService(mockRepo)
			var ctx context.Context
			if tc.wantPlainErr {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
				defer cancel()
			} else {
				ctx = context.Background()
				mockRepo.ListMock.Expect(ctx, tc.filter).
					Return(tc.mockReturn, tc.mockCount, tc.mockErr)
			}

			got, err := svc.List(ctx, tc.filter)
			switch {
			case tc.wantPlainErr:
				require.Error(t, err)
				require.True(t, errors.Is(err, context.Canceled))
			case tc.wantErrCode != nil:
				var ae *apperrors.AppError
				require.ErrorAs(t, err, &ae)
				require.Equal(t, *tc.wantErrCode, ae.Code)
			default:
				require.NoError(t, err)
				require.Equal(t, tc.wantEntries, got)
			}
		})
	}
}

func entry(orderID uint64, event models.EventType, offset time.Duration) models.HistoryEntry {
	clkForHistory := &clock.FakeClock{}
	return builders.NewHistoryEntryBuilder(clkForHistory).
		WithOrderID(orderID).
		WithEvent(event).
		WithOffset(offset).
		Build()
}
