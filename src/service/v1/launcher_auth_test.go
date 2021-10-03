package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestCreateLauncherUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockLauncherUserRepository := mock.NewMockLauncherUser(ctrl)
	mockLauncherSessionRepository := mock.NewMockLauncherSession(ctrl)

	launcherAuthService := NewLauncherAuth(
		mockDB,
		mockLauncherVersionRepository,
		mockLauncherUserRepository,
		mockLauncherSessionRepository,
	)

	type args struct {
		userNum int
	}
	type test struct {
		description string
		args
		GetLauncherVersionErr error
		launcherUsers         []*domain.LauncherUser
		CreateLauncherUserErr error
		isErr                 bool
		err                   error
	}

	productKey, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	testCases := []test{
		{
			description: "ランチャーバージョンが存在し、ユーザー作成も成功するのでエラーなし",
			args: args{
				userNum: 1,
			},
			GetLauncherVersionErr: nil,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey,
				),
			},
			CreateLauncherUserErr: nil,
		},
		{
			description: "ランチャーバージョンが存在しないのでエラー",
			args: args{
				userNum: 1,
			},
			GetLauncherVersionErr: repository.ErrRecordNotFound,
			isErr:                 true,
			err:                   service.ErrInvalidLauncherVersion,
		},
		{
			description: "ランチャーバージョンのチェックに失敗するのでエラー",
			args: args{
				userNum: 1,
			},
			GetLauncherVersionErr: errors.New("failed to get launcher version"),
			isErr:                 true,
		},
		{
			description: "ユーザー作成に失敗するのでエラー",
			args: args{
				userNum: 1,
			},
			GetLauncherVersionErr: nil,
			launcherUsers:         nil,
			CreateLauncherUserErr: errors.New("failed to create launcher user"),
			isErr:                 true,
		},
		{
			description: "作成するユーザー数が0でもエラーなし",
			args: args{
				userNum: 0,
			},
			GetLauncherVersionErr: nil,
			launcherUsers:         []*domain.LauncherUser{},
			CreateLauncherUserErr: nil,
		},
		{
			description: "作成するユーザー数が複数でもエラーなし",
			args: args{
				userNum: 2,
			},
			GetLauncherVersionErr: nil,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey,
				),
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey,
				),
			},
			CreateLauncherUserErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			launcherVersionID := values.NewLauncherVersionID()

			mockLauncherVersionRepository.
				EXPECT().
				GetLauncherVersion(ctx, launcherVersionID).
				Return(&domain.LauncherVersion{}, testCase.GetLauncherVersionErr)
			if testCase.GetLauncherVersionErr == nil {
				mockLauncherUserRepository.
					EXPECT().
					CreateLauncherUsers(ctx, launcherVersionID, gomock.Any()).
					Return(testCase.launcherUsers, testCase.CreateLauncherUserErr)
			}

			launcherUsers, err := launcherAuthService.CreateLauncherUser(ctx, launcherVersionID, testCase.args.userNum)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.ElementsMatch(t, testCase.launcherUsers, launcherUsers)
		})
	}
}
