package v1

import (
	"context"
	"errors"
	"net/url"
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

func TestSaveGameURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)
	mockGameURLRepository := mockRepository.NewMockGameURL(ctrl)

	gameURLService := NewGameURL(mockDB, mockGameRepository, mockGameVersionRepository, mockGameURLRepository)

	type test struct {
		description                 string
		gameID                      values.GameID
		link                        values.GameURLLink
		GetGameErr                  error
		executeGetLatestGameVersion bool
		gameVersion                 *domain.GameVersion
		GetLatestGameVersionErr     error
		executeGetGameURL           bool
		GetGameURLErr               error
		executeSaveGameURL          bool
		SaveGameURLErr              error
		isErr                       bool
		err                         error
	}

	gameID := values.NewGameID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}
	link := values.NewGameURLLink(urlLink)

	testCases := []test{
		{
			description:                 "特に問題ないのでエラーなし",
			gameID:                      gameID,
			link:                        link,
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeGetGameURL:  true,
			GetGameURLErr:      repository.ErrRecordNotFound,
			executeSaveGameURL: true,
		},
		{
			description: "ゲームが存在しないのでエラー",
			gameID:      gameID,
			link:        link,
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      gameID,
			link:        link,
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                 "ゲームバージョンが存在しないのでエラー",
			gameID:                      gameID,
			link:                        link,
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     repository.ErrRecordNotFound,
			isErr:                       true,
			err:                         service.ErrNoGameVersion,
		},
		{
			description:                 "GetLatestGameVersionがエラーなのでエラー",
			gameID:                      gameID,
			link:                        link,
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     errors.New("error"),
			isErr:                       true,
		},
		{
			description:                 "既にURLが存在しているのでエラー",
			gameID:                      gameID,
			link:                        link,
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeGetGameURL: true,
			isErr:             true,
			err:               service.ErrGameURLAlreadyExists,
		},
		{
			description:                 "GetGameURLがエラーなのでエラー",
			gameID:                      gameID,
			link:                        link,
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeGetGameURL: true,
			GetGameURLErr:     errors.New("error"),
			isErr:             true,
		},
		{
			description:                 "SaveGameURLがエラーなのでエラー",
			gameID:                      gameID,
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeGetGameURL:  true,
			GetGameURLErr:      repository.ErrRecordNotFound,
			executeSaveGameURL: true,
			SaveGameURLErr:     errors.New("error"),
			isErr:              true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetLatestGameVersion {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersion(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
					Return(testCase.gameVersion, testCase.GetLatestGameVersionErr)
			}

			if testCase.executeGetGameURL {
				mockGameURLRepository.
					EXPECT().
					GetGameURL(gomock.Any(), testCase.gameVersion.GetID()).
					Return(nil, testCase.GetGameURLErr)
			}

			if testCase.executeSaveGameURL {
				mockGameURLRepository.
					EXPECT().
					SaveGameURL(gomock.Any(), testCase.gameVersion.GetID(), gomock.Any()).
					Return(testCase.SaveGameURLErr)
			}

			gameURL, err := gameURLService.SaveGameURL(ctx, testCase.gameID, testCase.link)

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

			assert.Equal(t, testCase.link, gameURL.GetLink())
		})
	}
}

func TestGetGameURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGame(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersion(ctrl)
	mockGameURLRepository := mockRepository.NewMockGameURL(ctrl)

	gameURLService := NewGameURL(mockDB, mockGameRepository, mockGameVersionRepository, mockGameURLRepository)

	type test struct {
		description                 string
		gameID                      values.GameID
		GetGameErr                  error
		executeGetLatestGameVersion bool
		gameVersion                 *domain.GameVersion
		GetLatestGameVersionErr     error
		executeGetGameURL           bool
		gameURL                     *domain.GameURL
		GetGameURLErr               error
		isErr                       bool
		err                         error
	}

	gameID := values.NewGameID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}
	link := values.NewGameURLLink(urlLink)

	testCases := []test{
		{
			description:                 "特に問題ないのでエラーなし",
			gameID:                      gameID,
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeGetGameURL: true,
			gameURL:           domain.NewGameURL(values.NewGameURLID(), link, time.Now()),
		},
		{
			description: "ゲームが存在しないのでエラー",
			gameID:      gameID,
			GetGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      gameID,
			GetGameErr:  errors.New("error"),
			isErr:       true,
		},
		{
			description:                 "ゲームバージョンが存在しないのでエラー",
			gameID:                      gameID,
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     repository.ErrRecordNotFound,
			isErr:                       true,
			err:                         service.ErrNoGameVersion,
		},
		{
			description:                 "GetLatestGameVersionがエラーなのでエラー",
			gameID:                      gameID,
			executeGetLatestGameVersion: true,
			GetLatestGameVersionErr:     errors.New("error"),
			isErr:                       true,
		},
		{
			description:                 "GetGameURLがErrRecordNotFoundなのでエラー",
			gameID:                      gameID,
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeGetGameURL: true,
			GetGameURLErr:     repository.ErrRecordNotFound,
			isErr:             true,
			err:               service.ErrNoGameURL,
		},
		{
			description:                 "GetGameURLがエラーなのでエラー",
			gameID:                      gameID,
			executeGetLatestGameVersion: true,
			gameVersion: domain.NewGameVersion(
				values.NewGameVersionID(),
				"v1.0.0",
				"リリース",
				time.Now(),
			),
			executeGetGameURL: true,
			GetGameURLErr:     errors.New("error"),
			isErr:             true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.GetGameErr)

			if testCase.executeGetLatestGameVersion {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersion(gomock.Any(), testCase.gameID, repository.LockTypeNone).
					Return(testCase.gameVersion, testCase.GetLatestGameVersionErr)
			}

			if testCase.executeGetGameURL {
				mockGameURLRepository.
					EXPECT().
					GetGameURL(gomock.Any(), testCase.gameVersion.GetID()).
					Return(testCase.gameURL, testCase.GetGameURLErr)
			}

			gameURL, err := gameURLService.GetGameURL(ctx, testCase.gameID)

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

			assert.Equal(t, testCase.gameURL.GetID(), gameURL.GetID())
			assert.Equal(t, testCase.gameURL.GetLink(), gameURL.GetLink())
		})
	}
}
