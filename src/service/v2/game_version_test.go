package v2

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/types"
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
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameImageRepository := mockRepository.NewMockGameImageV2(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideoV2(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

	gameVersionService := NewGameVersion(
		mockDB,
		mockGameRepository,
		mockGameImageRepository,
		mockGameVideoRepository,
		mockGameFileRepository,
		mockGameVersionRepository,
	)

	type test struct {
		description              string
		gameID                   values.GameID
		versionName              values.GameVersionName
		versionDescription       values.GameVersionDescription
		imageID                  values.GameImageID
		videoID                  values.GameVideoID
		assets                   *service.Assets
		fileIDs                  []values.GameFileID
		executeGetGame           bool
		getGameErr               error
		executeGetGameImage      bool
		image                    *repository.GameImageInfo
		getGameImageErr          error
		executeGetGameVideo      bool
		video                    *repository.GameVideoInfo
		getGameVideoErr          error
		executeGetGameFile       bool
		files                    []*repository.GameFileInfo
		getGameFilesErr          error
		executeGetGameVersions   bool
		limit                    uint
		offset                   uint
		num                      uint
		versions                 []*repository.GameVersionInfo
		getGameVersionsErr       error
		executeCreateGameVersion bool
		createGameVersionErr     error
		isErr                    bool
		err                      error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()
	gameID5 := values.NewGameID()
	gameID6 := values.NewGameID()
	gameID7 := values.NewGameID()
	gameID8 := values.NewGameID()
	gameID9 := values.NewGameID()
	gameID10 := values.NewGameID()
	gameID11 := values.NewGameID()
	gameID12 := values.NewGameID()
	gameID13 := values.NewGameID()
	gameID14 := values.NewGameID()
	gameID15 := values.NewGameID()
	gameID16 := values.NewGameID()
	gameID17 := values.NewGameID()
	gameID18 := values.NewGameID()
	gameID19 := values.NewGameID()
	gameID20 := values.NewGameID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	imageID5 := values.NewGameImageID()
	imageID6 := values.NewGameImageID()
	imageID7 := values.NewGameImageID()
	imageID8 := values.NewGameImageID()
	imageID9 := values.NewGameImageID()
	imageID10 := values.NewGameImageID()
	imageID11 := values.NewGameImageID()
	imageID12 := values.NewGameImageID()
	imageID13 := values.NewGameImageID()
	imageID14 := values.NewGameImageID()
	imageID15 := values.NewGameImageID()
	imageID16 := values.NewGameImageID()
	imageID17 := values.NewGameImageID()
	imageID18 := values.NewGameImageID()
	imageID19 := values.NewGameImageID()
	imageID20 := values.NewGameImageID()
	imageID21 := values.NewGameImageID()

	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()
	videoID5 := values.NewGameVideoID()
	videoID6 := values.NewGameVideoID()
	videoID7 := values.NewGameVideoID()
	videoID8 := values.NewGameVideoID()
	videoID9 := values.NewGameVideoID()
	videoID10 := values.NewGameVideoID()
	videoID11 := values.NewGameVideoID()
	videoID12 := values.NewGameVideoID()
	videoID13 := values.NewGameVideoID()
	videoID14 := values.NewGameVideoID()
	videoID15 := values.NewGameVideoID()
	videoID16 := values.NewGameVideoID()
	videoID17 := values.NewGameVideoID()
	videoID18 := values.NewGameVideoID()
	videoID19 := values.NewGameVideoID()
	videoID20 := values.NewGameVideoID()
	videoID21 := values.NewGameVideoID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()
	fileID6 := values.NewGameFileID()
	fileID7 := values.NewGameFileID()
	fileID8 := values.NewGameFileID()
	fileID9 := values.NewGameFileID()
	fileID10 := values.NewGameFileID()

	now := time.Now()

	testCases := []test{
		{
			description:        "特に問題ないのでエラーなし",
			gameID:             gameID1,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID1,
			videoID:            videoID1,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID1,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID1,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID1,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID1,
			},
			executeCreateGameVersion: true,
			fileIDs:                  []values.GameFileID{},
		},
		{
			description:        "assetがwindowsでもエラーなし",
			gameID:             gameID2,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID2,
			videoID:            videoID2,
			assets: &service.Assets{
				Windows: types.NewOption(fileID1),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID2,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID2,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID2,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID2,
			},
			executeGetGameFile: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID1,
						values.GameFileTypeWindows,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: gameID2,
				},
			},
			executeCreateGameVersion: true,
			fileIDs:                  []values.GameFileID{fileID1},
		},
		{
			description:        "assetがmacでもエラーなし",
			gameID:             gameID3,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID3,
			videoID:            videoID3,
			assets: &service.Assets{
				Mac: types.NewOption(fileID2),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID3,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID3,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID3,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID3,
			},
			executeGetGameFile: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID2,
						values.GameFileTypeMac,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: gameID3,
				},
			},
			executeCreateGameVersion: true,
			fileIDs:                  []values.GameFileID{fileID2},
		},
		{
			description:        "assetがjarでもエラーなし",
			gameID:             gameID4,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID4,
			videoID:            videoID4,
			assets: &service.Assets{
				Jar: types.NewOption(fileID3),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID4,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID4,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID4,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID4,
			},
			executeGetGameFile: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID3,
						values.GameFileTypeJar,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: gameID4,
				},
			},
			executeCreateGameVersion: true,
			fileIDs:                  []values.GameFileID{fileID3},
		},
		{
			description:        "Assetが空なのでErrNoAsset",
			gameID:             gameID5,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID5,
			videoID:            videoID5,
			assets:             &service.Assets{},
			isErr:              true,
			err:                service.ErrNoAsset,
		},
		{
			description:        "ゲームが存在しないのでErrInvalidGameID",
			gameID:             gameID6,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID6,
			videoID:            videoID6,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame: true,
			getGameErr:     repository.ErrRecordNotFound,
			isErr:          true,
			err:            service.ErrInvalidGameID,
		},
		{
			description:        "GetGameがエラーなのでエラー",
			gameID:             gameID7,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID7,
			videoID:            videoID7,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame: true,
			getGameErr:     errors.New("error"),
			isErr:          true,
		},
		{
			description:        "画像が存在しないのでErrInvalidGameImageID",
			gameID:             gameID8,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID8,
			videoID:            videoID8,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			getGameImageErr:        repository.ErrRecordNotFound,
			isErr:                  true,
			err:                    service.ErrInvalidGameImageID,
		},
		{
			description:        "GetGameImageがエラーなのでエラー",
			gameID:             gameID9,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID9,
			videoID:            videoID9,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			getGameImageErr:        errors.New("error"),
			isErr:                  true,
		},
		{
			description:        "画像に紐づくゲームが違うのでErrInvalidGameImageID",
			gameID:             values.NewGameID(),
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID10,
			videoID:            videoID10,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID10,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: values.NewGameID(),
			},
			isErr: true,
			err:   service.ErrInvalidGameImageID,
		},
		{
			description:        "動画が存在しないのでErrInvalidGameVideoID",
			gameID:             gameID10,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID11,
			videoID:            videoID11,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID11,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID10,
			},
			executeGetGameVideo: true,
			getGameVideoErr:     repository.ErrRecordNotFound,
			isErr:               true,
			err:                 service.ErrInvalidGameVideoID,
		},
		{
			description:        "GetGameVideoがエラーなのでエラー",
			gameID:             gameID11,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID12,
			videoID:            videoID12,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID12,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID11,
			},
			executeGetGameVideo: true,
			getGameVideoErr:     errors.New("error"),
			isErr:               true,
		},
		{
			description:        "動画に紐づくゲームが違うのでErrInvalidGameVideoID",
			gameID:             gameID12,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID13,
			videoID:            videoID13,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID13,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID12,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID13,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: values.NewGameID(),
			},
			isErr: true,
			err:   service.ErrInvalidGameVideoID,
		},
		{
			description:        "GetGameFilesがエラーなのでエラー",
			gameID:             gameID13,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID14,
			videoID:            videoID14,
			assets: &service.Assets{
				Windows: types.NewOption(fileID4),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID14,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID13,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID14,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID13,
			},
			executeGetGameFile: true,
			fileIDs:            []values.GameFileID{fileID4},
			getGameFilesErr:    errors.New("error"),
			isErr:              true,
		},
		{
			description:        "ファイルが存在しないのでErrInvalidGameFileID",
			gameID:             gameID14,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID15,
			videoID:            videoID15,
			assets: &service.Assets{
				Windows: types.NewOption(fileID5),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID15,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID14,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID15,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID14,
			},
			executeGetGameFile: true,
			fileIDs:            []values.GameFileID{fileID5},
			files:              []*repository.GameFileInfo{},
			isErr:              true,
			err:                service.ErrInvalidGameFileID,
		},
		{
			description:        "ファイルの種類が違うのでErrInvalidGameFileType",
			gameID:             gameID15,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID16,
			videoID:            videoID16,
			assets: &service.Assets{
				Windows: types.NewOption(fileID6),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID16,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID15,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID16,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID15,
			},
			executeGetGameFile: true,
			fileIDs:            []values.GameFileID{fileID6},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID6,
						values.GameFileTypeMac,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: gameID15,
				},
			},
			isErr: true,
			err:   service.ErrInvalidGameFileType,
		},
		{
			description:        "CreateGameVersionがエラーなのでエラー",
			gameID:             gameID16,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID17,
			videoID:            videoID17,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID17,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID16,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID17,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID16,
			},
			executeCreateGameVersion: true,
			fileIDs:                  []values.GameFileID{},
			createGameVersionErr:     errors.New("error"),
			isErr:                    true,
		},
		{
			description:        "urlとfileが両方あっても問題なし",
			gameID:             gameID17,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID18,
			videoID:            videoID18,
			assets: &service.Assets{
				URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				Windows: types.NewOption(fileID7),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID18,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID17,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID18,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID17,
			},
			executeGetGameFile: true,
			fileIDs:            []values.GameFileID{fileID7},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID7,
						values.GameFileTypeWindows,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: gameID17,
				},
			},
			executeCreateGameVersion: true,
		},
		{
			description:        "ファイルが複数あっても問題なし",
			gameID:             gameID18,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID19,
			videoID:            videoID19,
			assets: &service.Assets{
				Windows: types.NewOption(fileID8),
				Mac:     types.NewOption(fileID9),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID19,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID18,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID19,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID18,
			},
			executeGetGameFile: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID8,
						values.GameFileTypeWindows,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: gameID18,
				},
				{
					GameFile: domain.NewGameFile(
						fileID9,
						values.GameFileTypeMac,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: gameID18,
				},
			},
			executeCreateGameVersion: true,
			fileIDs:                  []values.GameFileID{fileID8, fileID9},
		},
		{
			description:        "ファイルに紐づくゲームが違うのでエラー",
			gameID:             gameID19,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("おいす〜"),
			imageID:            imageID20,
			videoID:            videoID20,
			assets: &service.Assets{
				Windows: types.NewOption(fileID10),
			},
			executeGetGame:         true,
			executeGetGameImage:    true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID2,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID19,
			},
			executeGetGameVideo: true,
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID20,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID19,
			},
			executeGetGameFile: true,
			fileIDs:            []values.GameFileID{fileID10},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID10,
						values.GameFileTypeWindows,
						values.NewGameFileEntryPoint("/path/to/file"),
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			isErr: true,
			err:   service.ErrInvalidGameFileID,
		},
		{
			description:        "同名のバージョンが存在するのでエラー",
			gameID:             gameID20,
			versionName:        values.NewGameVersionName("v1.0.0"),
			versionDescription: values.NewGameVersionDescription("アップデート"),
			imageID:            imageID21,
			videoID:            videoID21,
			assets: &service.Assets{
				URL: types.NewOption(values.NewGameURLLink(urlLink)),
			},
			executeGetGame:         true,
			executeGetGameVersions: true,
			image: &repository.GameImageInfo{
				GameImage: domain.NewGameImage(
					imageID21,
					values.GameImageTypeJpeg,
					now,
				),
				GameID: gameID20,
			},
			video: &repository.GameVideoInfo{
				GameVideo: domain.NewGameVideo(
					videoID21,
					values.GameVideoTypeMp4,
					now,
				),
				GameID: gameID20,
			},
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						values.NewGameVersionID(),
						"v1.0.0",
						"version description",
						now,
					),
					ImageID: imageID21,
					VideoID: videoID21,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
			fileIDs:            []values.GameFileID{},
			getGameVersionsErr: nil,
			isErr:              true,
			err:                service.ErrDuplicateGameVersion,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeGetGame {
				mockGameRepository.
					EXPECT().
					GetGame(gomock.Any(), testCase.gameID, repository.LockTypeRecord).
					Return(nil, testCase.getGameErr)
			}

			if testCase.executeGetGameImage {
				mockGameImageRepository.
					EXPECT().
					GetGameImage(gomock.Any(), testCase.imageID, repository.LockTypeRecord).
					Return(testCase.image, testCase.getGameImageErr)
			}

			if testCase.executeGetGameVideo {
				mockGameVideoRepository.
					EXPECT().
					GetGameVideo(gomock.Any(), testCase.videoID, repository.LockTypeRecord).
					Return(testCase.video, testCase.getGameVideoErr)
			}

			if testCase.executeGetGameFile {
				mockGameFileRepository.
					EXPECT().
					GetGameFilesWithoutTypes(gomock.Any(), testCase.fileIDs, repository.LockTypeRecord).
					Return(testCase.files, testCase.getGameFilesErr)
			}

			if testCase.executeCreateGameVersion {
				mockGameVersionRepository.
					EXPECT().
					CreateGameVersion(gomock.Any(), testCase.gameID, testCase.imageID, testCase.videoID, testCase.assets.URL, testCase.fileIDs, gomock.Any()).
					Return(testCase.createGameVersionErr)
			}

			if testCase.executeGetGameVersions {
				mockGameVersionRepository.
					EXPECT().
					GetGameVersions(gomock.Any(), testCase.gameID, testCase.limit, testCase.offset, repository.LockTypeNone).
					Return(testCase.num, testCase.versions, testCase.getGameVersionsErr)
			}

			gameVersion, err := gameVersionService.CreateGameVersion(
				ctx,
				testCase.gameID,
				testCase.versionName,
				testCase.versionDescription,
				testCase.imageID,
				testCase.videoID,
				testCase.assets,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.versionName, gameVersion.GetName())
			assert.Equal(t, testCase.versionDescription, gameVersion.GetDescription())
			assert.WithinDuration(t, gameVersion.GetCreatedAt(), time.Now(), 2*time.Second)
			assert.Equal(t, testCase.imageID, gameVersion.ImageID)
			assert.Equal(t, testCase.videoID, gameVersion.VideoID)
			if assert.NotNil(t, gameVersion.Assets) {
				assert.Equal(t, testCase.assets, gameVersion.Assets)
			}
		})
	}
}

func TestGetGameVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameImageRepository := mockRepository.NewMockGameImageV2(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideoV2(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

	gameVersionService := NewGameVersion(
		mockDB,
		mockGameRepository,
		mockGameImageRepository,
		mockGameVideoRepository,
		mockGameFileRepository,
		mockGameVersionRepository,
	)

	type test struct {
		description            string
		gameID                 values.GameID
		params                 *service.GetGameVersionsParams
		limit                  uint
		offset                 uint
		num                    uint
		executeGetGameVersions bool
		versions               []*repository.GameVersionInfo
		getGameVersionsErr     error
		executeGetGameFiles    bool
		fileIDs                []values.GameFileID
		files                  []*repository.GameFileInfo
		getGameFilesErr        error
		gameVersionInfos       []*service.GameVersionInfo
		isErr                  bool
		err                    error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	versionID1 := values.NewGameVersionID()
	versionID2 := values.NewGameVersionID()
	versionID3 := values.NewGameVersionID()
	versionID4 := values.NewGameVersionID()
	versionID5 := values.NewGameVersionID()
	versionID6 := values.NewGameVersionID()
	versionID7 := values.NewGameVersionID()
	versionID8 := values.NewGameVersionID()
	versionID9 := values.NewGameVersionID()
	versionID10 := values.NewGameVersionID()
	versionID11 := values.NewGameVersionID()
	versionID12 := values.NewGameVersionID()
	versionID13 := values.NewGameVersionID()
	versionID14 := values.NewGameVersionID()
	versionID15 := values.NewGameVersionID()
	versionID16 := values.NewGameVersionID()
	versionID17 := values.NewGameVersionID()
	versionID18 := values.NewGameVersionID()
	versionID19 := values.NewGameVersionID()
	versionID20 := values.NewGameVersionID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	imageID5 := values.NewGameImageID()
	imageID6 := values.NewGameImageID()
	imageID7 := values.NewGameImageID()
	imageID8 := values.NewGameImageID()
	imageID9 := values.NewGameImageID()
	imageID10 := values.NewGameImageID()
	imageID11 := values.NewGameImageID()
	imageID12 := values.NewGameImageID()
	imageID13 := values.NewGameImageID()
	imageID14 := values.NewGameImageID()
	imageID15 := values.NewGameImageID()
	imageID16 := values.NewGameImageID()
	imageID17 := values.NewGameImageID()
	imageID18 := values.NewGameImageID()
	imageID19 := values.NewGameImageID()

	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()
	videoID5 := values.NewGameVideoID()
	videoID6 := values.NewGameVideoID()
	videoID7 := values.NewGameVideoID()
	videoID8 := values.NewGameVideoID()
	videoID9 := values.NewGameVideoID()
	videoID10 := values.NewGameVideoID()
	videoID11 := values.NewGameVideoID()
	videoID12 := values.NewGameVideoID()
	videoID13 := values.NewGameVideoID()
	videoID14 := values.NewGameVideoID()
	videoID15 := values.NewGameVideoID()
	videoID16 := values.NewGameVideoID()
	videoID17 := values.NewGameVideoID()
	videoID18 := values.NewGameVideoID()
	videoID19 := values.NewGameVideoID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()
	fileID6 := values.NewGameFileID()
	fileID7 := values.NewGameFileID()
	fileID8 := values.NewGameFileID()
	fileID9 := values.NewGameFileID()
	fileID10 := values.NewGameFileID()
	fileID11 := values.NewGameFileID()
	fileID12 := values.NewGameFileID()
	fileID13 := values.NewGameFileID()
	fileID14 := values.NewGameFileID()

	now := time.Now()

	testCases := []test{
		{
			description:            "特に問題ないのでエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID1,
						"version",
						"version description",
						now,
					),
					ImageID: imageID1,
					VideoID: videoID1,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID1,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID1,
					VideoID: videoID1,
				},
			},
		},
		{
			description:            "versionがなくてもエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    0,
			executeGetGameVersions: true,
			versions:               []*repository.GameVersionInfo{},
			gameVersionInfos:       []*service.GameVersionInfo{},
		},
		{
			description:            "versionが複数でもエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    2,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID2,
						"version",
						"version description",
						now,
					),
					ImageID: imageID2,
					VideoID: videoID2,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID3,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					ImageID: imageID3,
					VideoID: videoID3,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID2,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID2,
					VideoID: videoID2,
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID3,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID3,
					VideoID: videoID3,
				},
			},
		},
		{
			description:            "画像を共有していてもエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    2,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID4,
						"version",
						"version description",
						now,
					),
					ImageID: imageID4,
					VideoID: videoID4,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID5,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					ImageID: imageID4,
					VideoID: videoID5,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID4,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID4,
					VideoID: videoID4,
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID5,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID4,
					VideoID: videoID5,
				},
			},
		},
		{
			description:            "動画を共有していてもエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    2,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID6,
						"version",
						"version description",
						now,
					),
					ImageID: imageID5,
					VideoID: videoID6,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID7,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					ImageID: imageID6,
					VideoID: videoID6,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID6,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID5,
					VideoID: videoID6,
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID7,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID6,
					VideoID: videoID6,
				},
			},
		},
		{
			description:            "ファイルがwindowsでもエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID8,
						"version",
						"version description",
						now,
					),
					ImageID: imageID7,
					VideoID: videoID7,
					FileIDs: []values.GameFileID{fileID1},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID1},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID1,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID1,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						Windows: types.NewOption(fileID1),
					},
					ImageID: imageID7,
					VideoID: videoID7,
				},
			},
		},
		{
			description:            "ファイルがmacでもエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID9,
						"version",
						"version description",
						now,
					),
					ImageID: imageID8,
					VideoID: videoID8,
					FileIDs: []values.GameFileID{fileID2},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID2},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID2,
						values.GameFileTypeMac,
						"/path/to/game.app",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID9,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						Mac: types.NewOption(fileID2),
					},
					ImageID: imageID8,
					VideoID: videoID8,
				},
			},
		},
		{
			description:            "ファイルがjarでもエラーなし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID10,
						"version",
						"version description",
						now,
					),
					ImageID: imageID9,
					VideoID: videoID9,
					FileIDs: []values.GameFileID{fileID3},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID3},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID3,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID10,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						Jar: types.NewOption(fileID3),
					},
					ImageID: imageID9,
					VideoID: videoID9,
				},
			},
		},
		{
			description:            "windowsのファイルが重複していた場合、fileIDが先のものが優先される",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID11,
						"version",
						"version description",
						now,
					),
					ImageID: imageID10,
					VideoID: videoID10,
					FileIDs: []values.GameFileID{fileID4, fileID5},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID4, fileID5},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID4,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
				{
					GameFile: domain.NewGameFile(
						fileID5,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now.Add(-time.Hour),
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID11,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						Windows: types.NewOption(fileID4),
					},
					ImageID: imageID10,
					VideoID: videoID10,
				},
			},
		},
		{
			description:            "macのファイルが重複していた場合、fileIDが先のものが優先される",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID12,
						"version",
						"version description",
						now,
					),
					ImageID: imageID11,
					VideoID: videoID11,
					FileIDs: []values.GameFileID{fileID6, fileID7},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID6, fileID7},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID6,
						values.GameFileTypeMac,
						"/path/to/game.app",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
				{
					GameFile: domain.NewGameFile(
						fileID7,
						values.GameFileTypeMac,
						"/path/to/game.app",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now.Add(-time.Hour),
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID12,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						Mac: types.NewOption(fileID6),
					},
					ImageID: imageID11,
					VideoID: videoID11,
				},
			},
		},
		{
			description:            "jarのファイルが重複していた場合、fileIDが先のものが優先される",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID13,
						"version",
						"version description",
						now,
					),
					ImageID: imageID12,
					VideoID: videoID12,
					FileIDs: []values.GameFileID{fileID8, fileID9},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID8, fileID9},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID8,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
				{
					GameFile: domain.NewGameFile(
						fileID9,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now.Add(-time.Hour),
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID13,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						Jar: types.NewOption(fileID8),
					},
					ImageID: imageID12,
					VideoID: videoID12,
				},
			},
		},
		{
			description:            "ファイルが存在しない場合、エラーになる",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID14,
						"version",
						"version description",
						now,
					),
					ImageID: imageID13,
					VideoID: videoID13,
					FileIDs: []values.GameFileID{fileID10},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID10},
			files:               []*repository.GameFileInfo{},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID14,
						"version",
						"version description",
						now,
					),
					Assets:  &service.Assets{},
					ImageID: imageID13,
					VideoID: videoID13,
				},
			},
		},
		{
			description: "limitが0なのでErrInvalidLimit",
			gameID:      values.NewGameID(),
			params: &service.GetGameVersionsParams{
				Limit: 0,
			},
			isErr: true,
			err:   service.ErrInvalidLimit,
		},
		{
			description: "limitが存在してもエラーなし",
			gameID:      values.NewGameID(),
			params: &service.GetGameVersionsParams{
				Limit: 1,
			},
			limit:                  1,
			offset:                 0,
			num:                    2,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID15,
						"version",
						"version description",
						now,
					),
					ImageID: imageID14,
					VideoID: videoID14,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID15,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID14,
					VideoID: videoID14,
				},
			},
		},
		{
			description: "offsetが存在してもエラーなし",
			gameID:      values.NewGameID(),
			params: &service.GetGameVersionsParams{
				Limit:  1,
				Offset: 1,
			},
			limit:                  1,
			offset:                 1,
			num:                    2,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID16,
						"version",
						"version description",
						now,
					),
					ImageID: imageID15,
					VideoID: videoID15,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID16,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						URL: types.NewOption(values.NewGameURLLink(urlLink)),
					},
					ImageID: imageID15,
					VideoID: videoID15,
				},
			},
		},
		{
			description:            "GetGameVersionsがエラーなのでエラー",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			getGameVersionsErr:     errors.New("error"),
			isErr:                  true,
		},
		{
			description:            "GetGameFilesがエラーなのでエラー",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID17,
						"version",
						"version description",
						now,
					),
					ImageID: imageID16,
					VideoID: videoID16,
					FileIDs: []values.GameFileID{fileID11, fileID12},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID11, fileID12},
			getGameFilesErr:     errors.New("error"),
			isErr:               true,
		},
		{
			description:            "ファイルが重複していても問題なし",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    2,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID18,
						"version",
						"version description",
						now,
					),
					ImageID: imageID17,
					VideoID: videoID17,
					FileIDs: []values.GameFileID{fileID13},
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID19,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					ImageID: imageID18,
					VideoID: videoID18,
					FileIDs: []values.GameFileID{fileID13},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID13},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID13,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID18,
						"version",
						"version description",
						now,
					),
					Assets: &service.Assets{
						Windows: types.NewOption(fileID13),
					},
					ImageID: imageID17,
					VideoID: videoID17,
				},
				{
					GameVersion: domain.NewGameVersion(
						versionID19,
						"version",
						"version description",
						now.Add(-time.Hour),
					),
					Assets: &service.Assets{
						Windows: types.NewOption(fileID13),
					},
					ImageID: imageID18,
					VideoID: videoID18,
				},
			},
		},
		{
			description:            "ファイルがwindows,mac,jarのいずれでもないファイルは無視される",
			gameID:                 values.NewGameID(),
			limit:                  0,
			offset:                 0,
			num:                    1,
			executeGetGameVersions: true,
			versions: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID20,
						"version",
						"version description",
						now,
					),
					ImageID: imageID19,
					VideoID: videoID19,
					FileIDs: []values.GameFileID{fileID14},
				},
			},
			executeGetGameFiles: true,
			fileIDs:             []values.GameFileID{fileID14},
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID14,
						values.GameFileType(100),
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfos: []*service.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						versionID20,
						"version",
						"version description",
						now,
					),
					Assets:  &service.Assets{},
					ImageID: imageID19,
					VideoID: videoID19,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeGetGameVersions {
				mockGameVersionRepository.
					EXPECT().
					GetGameVersions(gomock.Any(), testCase.gameID, testCase.limit, testCase.offset, repository.LockTypeNone).
					Return(testCase.num, testCase.versions, testCase.getGameVersionsErr)
			}

			if testCase.executeGetGameFiles {
				mockGameFileRepository.
					EXPECT().
					GetGameFilesWithoutTypes(gomock.Any(), testCase.fileIDs, repository.LockTypeNone).
					Return(testCase.files, testCase.getGameFilesErr)
			}

			num, gameVersions, err := gameVersionService.GetGameVersions(
				ctx,
				testCase.gameID,
				testCase.params,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.num, num)
			for i, gameVersion := range gameVersions {
				assert.Equal(t, testCase.gameVersionInfos[i].GameVersion.GetName(), gameVersion.GameVersion.GetName())
				assert.Equal(t, testCase.gameVersionInfos[i].GameVersion.GetDescription(), gameVersion.GameVersion.GetDescription())
				assert.WithinDuration(t, testCase.gameVersionInfos[i].GameVersion.GetCreatedAt(), gameVersion.GameVersion.GetCreatedAt(), 2*time.Second)
				assert.Equal(t, testCase.gameVersionInfos[i].ImageID, gameVersion.ImageID)
				assert.Equal(t, testCase.gameVersionInfos[i].VideoID, gameVersion.VideoID)
				if assert.NotNil(t, gameVersion.Assets) {
					assert.Equal(t, testCase.gameVersionInfos[i].Assets, gameVersion.Assets)
				}
			}
		})
	}
}

func TestGetLatestGameVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockRepository.NewMockDB(ctrl)
	mockGameRepository := mockRepository.NewMockGameV2(ctrl)
	mockGameImageRepository := mockRepository.NewMockGameImageV2(ctrl)
	mockGameVideoRepository := mockRepository.NewMockGameVideoV2(ctrl)
	mockGameFileRepository := mockRepository.NewMockGameFileV2(ctrl)
	mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

	gameVersionService := NewGameVersion(
		mockDB,
		mockGameRepository,
		mockGameImageRepository,
		mockGameVideoRepository,
		mockGameFileRepository,
		mockGameVersionRepository,
	)

	type test struct {
		description                 string
		gameID                      values.GameID
		getGameErr                  error
		executeGetLatestGameVersion bool
		version                     *repository.GameVersionInfo
		getLatestGameVersionErr     error
		executeGetGameFiles         bool
		files                       []*repository.GameFileInfo
		getGameFilesErr             error
		gameVersionInfo             *service.GameVersionInfo
		isErr                       bool
		err                         error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	versionID1 := values.NewGameVersionID()
	versionID2 := values.NewGameVersionID()
	versionID3 := values.NewGameVersionID()
	versionID4 := values.NewGameVersionID()
	versionID5 := values.NewGameVersionID()
	versionID6 := values.NewGameVersionID()
	versionID7 := values.NewGameVersionID()
	versionID8 := values.NewGameVersionID()
	versionID9 := values.NewGameVersionID()
	versionID10 := values.NewGameVersionID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	imageID5 := values.NewGameImageID()
	imageID6 := values.NewGameImageID()
	imageID7 := values.NewGameImageID()
	imageID8 := values.NewGameImageID()
	imageID9 := values.NewGameImageID()
	imageID10 := values.NewGameImageID()

	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()
	videoID5 := values.NewGameVideoID()
	videoID6 := values.NewGameVideoID()
	videoID7 := values.NewGameVideoID()
	videoID8 := values.NewGameVideoID()
	videoID9 := values.NewGameVideoID()
	videoID10 := values.NewGameVideoID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()
	fileID6 := values.NewGameFileID()
	fileID7 := values.NewGameFileID()
	fileID8 := values.NewGameFileID()
	fileID9 := values.NewGameFileID()
	fileID10 := values.NewGameFileID()
	fileID11 := values.NewGameFileID()
	fileID12 := values.NewGameFileID()
	fileID13 := values.NewGameFileID()

	now := time.Now()

	testCases := []test{
		{
			description:                 "特に問題ないのでエラーなし",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID1,
					"version",
					"version description",
					now,
				),
				ImageID: imageID1,
				VideoID: videoID1,
				URL:     types.NewOption(values.NewGameURLLink(urlLink)),
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID1,
					"version",
					"version description",
					now,
				),
				Assets: &service.Assets{
					URL: types.NewOption(values.NewGameURLLink(urlLink)),
				},
				ImageID: imageID1,
				VideoID: videoID1,
			},
		},
		{
			description:                 "versionがないのでErrNoGameVersion",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			getLatestGameVersionErr:     repository.ErrRecordNotFound,
			isErr:                       true,
			err:                         service.ErrNoGameVersion,
		},
		{
			description:                 "ファイルがwindowsでもエラーなし",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID2,
					"version",
					"version description",
					now,
				),
				ImageID: imageID2,
				VideoID: videoID2,
				FileIDs: []values.GameFileID{fileID1},
			},
			executeGetGameFiles: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID1,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID2,
					"version",
					"version description",
					now,
				),
				Assets: &service.Assets{
					Windows: types.NewOption(fileID1),
				},
				ImageID: imageID2,
				VideoID: videoID2,
			},
		},
		{
			description:                 "ファイルがmacでもエラーなし",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID3,
					"version",
					"version description",
					now,
				),
				ImageID: imageID3,
				VideoID: videoID3,
				FileIDs: []values.GameFileID{fileID2},
			},
			executeGetGameFiles: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID2,
						values.GameFileTypeMac,
						"/path/to/game.app",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID3,
					"version",
					"version description",
					now,
				),
				Assets: &service.Assets{
					Mac: types.NewOption(fileID2),
				},
				ImageID: imageID3,
				VideoID: videoID3,
			},
		},
		{
			description:                 "ファイルがjarでもエラーなし",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID4,
					"version",
					"version description",
					now,
				),
				ImageID: imageID4,
				VideoID: videoID4,
				FileIDs: []values.GameFileID{fileID3},
			},
			executeGetGameFiles: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID3,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID4,
					"version",
					"version description",
					now,
				),
				Assets: &service.Assets{
					Jar: types.NewOption(fileID3),
				},
				ImageID: imageID4,
				VideoID: videoID4,
			},
		},
		{
			description:                 "windowsのファイルが重複していた場合、fileIDが先のものが優先される",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID5,
					"version",
					"version description",
					now,
				),
				ImageID: imageID5,
				VideoID: videoID5,
				FileIDs: []values.GameFileID{fileID4, fileID5},
			},
			executeGetGameFiles: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID4,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
				{
					GameFile: domain.NewGameFile(
						fileID5,
						values.GameFileTypeWindows,
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now.Add(-time.Hour),
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID5,
					"version",
					"version description",
					now,
				),
				Assets: &service.Assets{
					Windows: types.NewOption(fileID4),
				},
				ImageID: imageID5,
				VideoID: videoID5,
			},
		},
		{
			description:                 "macのファイルが重複していた場合、fileIDが先のものが優先される",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID6,
					"version",
					"version description",
					now,
				),
				ImageID: imageID6,
				VideoID: videoID6,
				FileIDs: []values.GameFileID{fileID6, fileID7},
			},
			executeGetGameFiles: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID6,
						values.GameFileTypeMac,
						"/path/to/game.app",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
				{
					GameFile: domain.NewGameFile(
						fileID7,
						values.GameFileTypeMac,
						"/path/to/game.app",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now.Add(-time.Hour),
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID6,
					"version",
					"version description",
					now,
				),
				Assets: &service.Assets{
					Mac: types.NewOption(fileID6),
				},
				ImageID: imageID6,
				VideoID: videoID6,
			},
		},
		{
			description:                 "jarのファイルが重複していた場合、fileIDが先のものが優先される",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID7,
					"version",
					"version description",
					now,
				),
				ImageID: imageID7,
				VideoID: videoID7,
				FileIDs: []values.GameFileID{fileID8, fileID9},
			},
			executeGetGameFiles: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID8,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
				{
					GameFile: domain.NewGameFile(
						fileID9,
						values.GameFileTypeJar,
						"/path/to/game.jar",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now.Add(-time.Hour),
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID7,
					"version",
					"version description",
					now,
				),
				Assets: &service.Assets{
					Jar: types.NewOption(fileID8),
				},
				ImageID: imageID7,
				VideoID: videoID7,
			},
		},
		{
			description:                 "ファイルが存在しない場合、無視",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID8,
					"version",
					"version description",
					now,
				),
				ImageID: imageID8,
				VideoID: videoID8,
				FileIDs: []values.GameFileID{fileID10},
			},
			executeGetGameFiles: true,
			files:               []*repository.GameFileInfo{},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID8,
					"version",
					"version description",
					now,
				),
				Assets:  &service.Assets{},
				ImageID: imageID8,
				VideoID: videoID8,
			},
		},
		{
			description:                 "GetLatestGameVersionがエラーなのでエラー",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			getLatestGameVersionErr:     errors.New("error"),
			isErr:                       true,
		},
		{
			description:                 "GetGameFilesがエラーなのでエラー",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID9,
					"version",
					"version description",
					now,
				),
				ImageID: imageID9,
				VideoID: videoID9,
				FileIDs: []values.GameFileID{fileID11, fileID12},
			},
			executeGetGameFiles: true,
			getGameFilesErr:     errors.New("error"),
			isErr:               true,
		},
		{
			description:                 "ファイルがwindows,mac,jarのいずれでもないファイルは無視される",
			gameID:                      values.NewGameID(),
			executeGetLatestGameVersion: true,
			version: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID10,
					"version",
					"version description",
					now,
				),
				ImageID: imageID10,
				VideoID: videoID10,
				FileIDs: []values.GameFileID{fileID13},
			},
			executeGetGameFiles: true,
			files: []*repository.GameFileInfo{
				{
					GameFile: domain.NewGameFile(
						fileID13,
						values.GameFileType(100),
						"/path/to/game.exe",
						values.NewGameFileHashFromBytes([]byte{0x09, 0x8f, 0x6b, 0xcd, 0x46, 0x21, 0xd3, 0x73, 0xca, 0xde, 0x4e, 0x83, 0x26, 0x27, 0xb4, 0xf6}),
						now,
					),
					GameID: values.NewGameID(),
				},
			},
			gameVersionInfo: &service.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					versionID10,
					"version",
					"version description",
					now,
				),
				Assets:  &service.Assets{},
				ImageID: imageID10,
				VideoID: videoID10,
			},
		},
		{
			description: "ゲームが存在しないのでErrInvalidGameID",
			gameID:      values.NewGameID(),
			getGameErr:  repository.ErrRecordNotFound,
			isErr:       true,
			err:         service.ErrInvalidGameID,
		},
		{
			description: "GetGameがエラーなのでエラー",
			gameID:      values.NewGameID(),
			getGameErr:  errors.New("error"),
			isErr:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockGameRepository.
				EXPECT().
				GetGame(gomock.Any(), testCase.gameID, repository.LockTypeNone).
				Return(nil, testCase.getGameErr)

			if testCase.executeGetLatestGameVersion {
				mockGameVersionRepository.
					EXPECT().
					GetLatestGameVersion(gomock.Any(), testCase.gameID, repository.LockTypeNone).
					Return(testCase.version, testCase.getLatestGameVersionErr)
			}

			if testCase.executeGetGameFiles {
				mockGameFileRepository.
					EXPECT().
					GetGameFilesWithoutTypes(gomock.Any(), testCase.version.FileIDs, repository.LockTypeNone).
					Return(testCase.files, testCase.getGameFilesErr)
			}

			gameVersion, err := gameVersionService.GetLatestGameVersion(
				ctx,
				testCase.gameID,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, testCase.gameVersionInfo.GetID(), gameVersion.GetID())
			assert.Equal(t, testCase.gameVersionInfo.GetName(), gameVersion.GetName())
			assert.Equal(t, testCase.gameVersionInfo.GetDescription(), gameVersion.GetDescription())
			assert.WithinDuration(t, testCase.gameVersionInfo.GetCreatedAt(), gameVersion.GetCreatedAt(), 2*time.Second)
			assert.Equal(t, testCase.gameVersionInfo.ImageID, gameVersion.ImageID)
			assert.Equal(t, testCase.gameVersionInfo.VideoID, gameVersion.VideoID)
			if assert.NotNil(t, gameVersion.Assets) {
				assert.Equal(t, testCase.gameVersionInfo.Assets, gameVersion.Assets)
			}
		})
	}
}
