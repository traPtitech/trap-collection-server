package v1

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestGetLauncherUsers(t *testing.T) {
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

	type test struct {
		description                            string
		GetLauncherVersionErr                  error
		launcherUsers                          []*domain.LauncherUser
		GetLauncherUsersByLauncherVersionIDErr error
		isErr                                  bool
		err                                    error
	}

	productKey, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	testCases := []test{
		{
			description:           "ランチャーバージョンが存在し、ユーザー作成も成功するのでエラーなし",
			GetLauncherVersionErr: nil,
			launcherUsers: []*domain.LauncherUser{
				domain.NewLauncherUser(
					values.NewLauncherUserID(),
					productKey,
				),
			},
			GetLauncherUsersByLauncherVersionIDErr: nil,
		},
		{
			description:           "ランチャーバージョンが存在しないのでエラー",
			GetLauncherVersionErr: repository.ErrRecordNotFound,
			isErr:                 true,
			err:                   service.ErrInvalidLauncherVersion,
		},
		{
			description:           "ランチャーバージョンのチェックに失敗するのでエラー",
			GetLauncherVersionErr: errors.New("failed to get launcher version"),
			isErr:                 true,
		},
		{
			description:                            "ユーザー取得に失敗するのでエラー",
			GetLauncherVersionErr:                  nil,
			launcherUsers:                          nil,
			GetLauncherUsersByLauncherVersionIDErr: errors.New("failed to get launcher users"),
			isErr:                                  true,
		},
		{
			description:                            "取得したユーザー数が0でもエラーなし",
			GetLauncherVersionErr:                  nil,
			launcherUsers:                          []*domain.LauncherUser{},
			GetLauncherUsersByLauncherVersionIDErr: nil,
		},
		{
			description:           "取得したユーザー数が複数でもエラーなし",
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
			GetLauncherUsersByLauncherVersionIDErr: nil,
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
				mockLauncherVersionRepository.
					EXPECT().
					GetLauncherUsersByLauncherVersionID(ctx, launcherVersionID).
					Return(testCase.launcherUsers, testCase.GetLauncherUsersByLauncherVersionIDErr)
			}

			launcherUsers, err := launcherAuthService.GetLauncherUsers(ctx, launcherVersionID)

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

func TestRevokeProductKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)

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

	type test struct {
		description        string
		DeleteLauncherUser error
		isErr              bool
		err                error
	}

	testCases := []test{
		{
			description: "削除に成功するのでエラーなし",
		},
		{
			description:        "ユーザーが存在しないのでエラー",
			DeleteLauncherUser: repository.ErrNoRecordDeleted,
			isErr:              true,
			err:                service.ErrInvalidLauncherUser,
		},
		{
			description:        "削除に失敗するのでエラー",
			DeleteLauncherUser: errors.New("failed to delete launcher user"),
			isErr:              true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			launcherUserID := values.NewLauncherUserID()

			mockLauncherUserRepository.
				EXPECT().
				DeleteLauncherUser(ctx, launcherUserID).
				Return(testCase.DeleteLauncherUser)

			err := launcherAuthService.RevokeProductKey(ctx, launcherUserID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoginLauncher(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)

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

	type test struct {
		description                    string
		GetLauncherUserByProductKeyErr error
		launcherSession                *domain.LauncherSession
		CreateLauncherSessionErr       error
		isErr                          bool
		err                            error
	}

	accessToken, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}
	testCases := []test{
		{
			description: "ログインに成功するのでエラーなし",
			launcherSession: domain.NewLauncherSession(
				values.NewLauncherSessionID(),
				accessToken,
				getExpiresAt(),
			),
		},
		{
			description:                    "ユーザーが存在しないのでエラー",
			GetLauncherUserByProductKeyErr: repository.ErrRecordNotFound,
			isErr:                          true,
			err:                            service.ErrInvalidLauncherUserProductKey,
		},
		{
			description:                    "ユーザー確認に失敗するのでエラー",
			GetLauncherUserByProductKeyErr: errors.New("failed to get launcher user by product key"),
			isErr:                          true,
		},
		{
			description:              "セッション作成に失敗するのでエラー",
			CreateLauncherSessionErr: errors.New("failed to create launcher session"),
			isErr:                    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			productKey, err := values.NewLauncherUserProductKey()
			if err != nil {
				t.Errorf("failed to create product key: %v", err)
			}

			mockLauncherUserRepository.
				EXPECT().
				GetLauncherUserByProductKey(ctx, productKey).
				Return(&domain.LauncherUser{}, testCase.GetLauncherUserByProductKeyErr)

			if testCase.GetLauncherUserByProductKeyErr == nil {
				mockLauncherSessionRepository.
					EXPECT().
					CreateLauncherSession(ctx, gomock.Any()).
					Return(testCase.launcherSession, testCase.CreateLauncherSessionErr)
			}

			session, err := launcherAuthService.LoginLauncher(ctx, productKey)

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

			assert.Equal(t, testCase.launcherSession, session)
		})
	}
}

func TestGetExpiresAt(t *testing.T) {
	t.Parallel()

	loopNum := 100
	for i := 0; i < loopNum; i++ {
		expiresAt := getExpiresAt()
		assert.InDelta(t, expiresIn*time.Second, time.Until(expiresAt), float64(time.Second))
	}
}

func TestLauncherAuth(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)

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

	type test struct {
		description                                         string
		launcherVersion                                     *domain.LauncherVersion
		launcherUser                                        *domain.LauncherUser
		launcherSession                                     *domain.LauncherSession
		GetLauncherVersionAndUserAndSessionByAccessTokenErr error
		isErr                                               bool
		err                                                 error
	}

	productKey, err := values.NewLauncherUserProductKey()
	if err != nil {
		t.Errorf("failed to create product key: %v", err)
	}

	accessToken, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		t.Errorf("failed to create access token: %v", err)
	}

	testCases := []test{
		{
			description: "認証に成功するのでエラーなし",
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				values.NewLauncherVersionID(),
				values.NewLauncherVersionName("2021.10.03"),
				time.Now(),
			),
			launcherUser: domain.NewLauncherUser(
				values.NewLauncherUserID(),
				productKey,
			),
			launcherSession: domain.NewLauncherSession(
				values.NewLauncherSessionID(),
				accessToken,
				getExpiresAt(),
			),
		},
		{
			description: "アクセストークンが存在しないのでエラー",
			GetLauncherVersionAndUserAndSessionByAccessTokenErr: repository.ErrRecordNotFound,
			isErr: true,
			err:   service.ErrInvalidLauncherSessionAccessToken,
		},
		{
			description: "アクセストークンが期限切れのためエラー",
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				values.NewLauncherVersionID(),
				values.NewLauncherVersionName("1.0.0"),
				time.Now(),
			),
			launcherUser: domain.NewLauncherUser(
				values.NewLauncherUserID(),
				productKey,
			),
			launcherSession: domain.NewLauncherSession(
				values.NewLauncherSessionID(),
				accessToken,
				time.Now().Add(-1*time.Hour),
			),
			isErr: true,
			err:   service.ErrLauncherSessionAccessTokenExpired,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				GetLauncherVersionAndUserAndSessionByAccessToken(ctx, accessToken).
				Return(testCase.launcherVersion, testCase.launcherUser, testCase.launcherSession, testCase.GetLauncherVersionAndUserAndSessionByAccessTokenErr)

			launcherUser, launcherVersion, err := launcherAuthService.LauncherAuth(ctx, accessToken)

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

			assert.Equal(t, testCase.launcherUser, launcherUser)
			assert.Equal(t, testCase.launcherVersion, launcherVersion)
		})
	}
}
