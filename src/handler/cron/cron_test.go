package cron

import (
	"testing"

	"github.com/stretchr/testify/assert"
	mockService "github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestDeleteLongLogs(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		deleteLongLogsErr error
	}{
		"正常に終了": {
			deleteLongLogsErr: nil,
		},
		"サービスエラー発生": {
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
				Return(tc.deleteLongLogsErr)

			cronHandler := NewCron(mockPlayLogService)

			cronHandler.deleteLongLogs()
		})
	}
}
