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
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestCreateGameVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	gameVersionService := NewGameVersion(mockDB, mockGameRepository, mockGameVersionRepository)

	type test struct {
		description              string
		gameID                   values.GameID
		versionName              values.GameVersionName
		versionDescription       values.GameVersionDescription
		GetGameErr               error
		executeCreateGameVersion bool
		CreateGameVersionErr     error
		isErr                    bool
		err                      error
	}

	testCases := []test{
		{
			description:              "特に問題ないのでエラーなし",
			gameID:                   values.NewGameID(),
			versionName:              values.NewGameVersionName("1.0.0"),
			versionDescription:       values.NewGameVersionDescription("おいす〜"),
			executeCreateGameVersion: true,
		},
		{
			description:        "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:             values.NewGameID(),
			versionName:        values.NewGameVersionName("1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			GetGameErr:         repository.ErrRecordNotFound,
			isErr:              true,
			err:                service.ErrInvalidGameID,
		},
		{
			description:        "GetGameがエラーなのでエラー",
			gameID:             values.NewGameID(),
			versionName:        values.NewGameVersionName("1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			GetGameErr:         errors.New("error"),
			isErr:              true,
		},
		{
			description:              "CreateGameVersionがエラーなのでエラー",
			gameID:                   values.NewGameID(),
			versionName:              values.NewGameVersionName("1.0.0"),
			versionDescription:       values.NewGameVersionDescription("おいす〜"),
			executeCreateGameVersion: true,
			CreateGameVersionErr:     errors.New("error"),
			isErr:                    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
				Return(nil, testCase.GetGameErr)

			if testCase.executeCreateGameVersion {
				mockGameVersionRepository.
					EXPECT().
					CreateGameVersion(gomock.Any(), testCase.gameID, gomock.Any()).
					Return(testCase.CreateGameVersionErr)
			}

			gameVersion, err := gameVersionService.CreateGameVersion(ctx, testCase.gameID, testCase.versionName, testCase.versionDescription)

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

			assert.Equal(t, testCase.versionName, gameVersion.GetName())
			assert.Equal(t, testCase.versionDescription, gameVersion.GetDescription())
			assert.WithinDuration(t, gameVersion.GetCreatedAt(), time.Now(), 2*time.Second)
		})
	}
}

func TestGetGameVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)

	gameVersionService := NewGameVersion(mockDB, mockGameRepository, mockGameVersionRepository)

	type test struct {
		description            string
		gameID                 values.GameID
		GetGameErr             error
		executeGetGameVersions bool
		GetGameVersionsErr     error
		gameVersions           []*domain.GameVersion
		isErr                  bool
		err                    error
	}

	testCases := []test{
		{
			description:            "特に問題ないのでエラーなし",
			gameID:                 values.NewGameID(),
			executeGetGameVersions: true,
			gameVersions: []*domain.GameVersion{
				domain.NewGameVersion(
					values.NewGameVersionID(),
					values.NewGameVersionName("1.0.0"),
					values.NewGameVersionDescription("おいす〜"),
					time.Now(),
				),
			},
		},
		{
			description: "GetGameがErrRecordNotFoundなのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:            "GetGameVersionsがエラーなのでエラー",
			gameID:                 values.NewGameID(),
			executeGetGameVersions: true,
			GetGameVersionsErr:     errors.New("error"),
			isErr:                  true,
		},
		{
			description:            "gameVersionsが複数でもエラーなし",
			gameID:                 values.NewGameID(),
			executeGetGameVersions: true,
			gameVersions: []*domain.GameVersion{
				domain.NewGameVersion(
					values.NewGameVersionID(),
					values.NewGameVersionName("1.0.1"),
					values.NewGameVersionDescription("おいす〜"),
					time.Now(),
				),
				domain.NewGameVersion(
					values.NewGameVersionID(),
					values.NewGameVersionName("1.0.0"),
					values.NewGameVersionDescription("おいす〜"),
					time.Now(),
				),
			},
		},
		{
			description:            "gameVersionsが空でもエラーなし",
			gameID:                 values.NewGameID(),
			executeGetGameVersions: true,
			gameVersions:           []*domain.GameVersion{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetGameVersions {
				mockGameVersionRepository.
					EXPECT().
					GetGameVersions(gomock.Any(), testCase.gameID).
					Return(testCase.gameVersions, testCase.GetGameVersionsErr)
			}

			gameVersions, err := gameVersionService.GetGameVersions(ctx, testCase.gameID)

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

			assert.Equal(t, len(testCase.gameVersions), len(gameVersions))
			for i, gameVersion := range gameVersions {
				assert.Equal(t, testCase.gameVersions[i].GetName(), gameVersion.GetName())
				assert.Equal(t, testCase.gameVersions[i].GetDescription(), gameVersion.GetDescription())
				assert.WithinDuration(t, gameVersion.GetCreatedAt(), time.Now(), 2*time.Second)
			}
		})
	}
}
