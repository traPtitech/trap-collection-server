package cron

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	mockService "github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestDeleteLongLogs(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		deletedIDs        []values.GamePlayLogID
		deleteLongLogsErr error
	}{
		"正常に削除": {
			deletedIDs: []values.GamePlayLogID{
				values.NewGamePlayLogID(),
				values.NewGamePlayLogID(),
			},
			deleteLongLogsErr: nil,
		},
		"削除対象ログなし": {
			deletedIDs:        []values.GamePlayLogID{},
			deleteLongLogsErr: nil,
		},
		"サービスエラー発生": {
			deletedIDs:        nil,
			deleteLongLogsErr: assert.AnError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockPlayLogService := mockService.NewMockGamePlayLogV2(ctrl)

			mockPlayLogService.
				EXPECT().
				DeleteLongLogs(gomock.Any()).
				Return(tc.deletedIDs, tc.deleteLongLogsErr)

			cronHandler := NewCron(mockPlayLogService)

			cronHandler.deleteLongLogs()
		})
	}
}
