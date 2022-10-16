package gorm2

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

func TestCreateGameVersionV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	gameVersionRepository := NewGameVersionV2(testDB)

	type test struct {
		description        string
		gameID             values.GameID
		imageID            values.GameImageID
		videoID            values.GameVideoID
		optionURL          types.Option[values.GameURLLink]
		fileIDs            []values.GameFileID
		version            *domain.GameVersion
		existGame          bool
		existImage         bool
		existVideo         bool
		files              []migrate.GameFileTable2
		beforeGameVersions []migrate.GameVersionTable2
		expectGameVersions []migrate.GameVersionTable2
		isErr              bool
		err                error
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

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()
	gameVersionID8 := values.NewGameVersionID()
	gameVersionID9 := values.NewGameVersionID()
	gameVersionID10 := values.NewGameVersionID()
	gameVersionID11 := values.NewGameVersionID()
	gameVersionID12 := values.NewGameVersionID()
	gameVersionID13 := values.NewGameVersionID()
	gameVersionID14 := values.NewGameVersionID()
	gameVersionID15 := values.NewGameVersionID()
	gameVersionID16 := values.NewGameVersionID()
	gameVersionID17 := values.NewGameVersionID()

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

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()
	fileID4 := values.NewGameFileID()
	fileID5 := values.NewGameFileID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	var imageType migrate.GameImageTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameImageTypeJpeg).
		Select("id").
		Take(&imageType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	var videoType migrate.GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameVideoTypeMp4).
		Select("id").
		Take(&videoType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	var fileType migrate.GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameFileTypeJar).
		Select("id").
		Take(&fileType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			imageID:     imageID1,
			videoID:     videoID1,
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID1,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:          true,
			existImage:         true,
			existVideo:         true,
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID1),
					GameID:      uuid.UUID(gameID1),
					GameImageID: uuid.UUID(imageID1),
					GameVideoID: uuid.UUID(videoID1),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "既にバージョンが存在してもエラーなし",
			gameID:      gameID2,
			imageID:     imageID2,
			videoID:     videoID2,
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID2,
				values.NewGameVersionName("v1.1.0"),
				values.NewGameVersionDescription("アップデート"),
				now,
			),
			existGame:  true,
			existImage: true,
			existVideo: true,
			beforeGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID3),
					GameID:      uuid.UUID(gameID2),
					GameImageID: uuid.UUID(imageID3),
					GameVideoID: uuid.UUID(videoID3),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now.Add(-time.Hour),
					GameImage: migrate.GameImageTable2{
						ID:          uuid.UUID(imageID3),
						GameID:      uuid.UUID(gameID2),
						ImageTypeID: imageType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
					GameVideo: migrate.GameVideoTable2{
						ID:          uuid.UUID(videoID3),
						GameID:      uuid.UUID(gameID2),
						VideoTypeID: videoType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
				},
			},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID2),
					GameID:      uuid.UUID(gameID2),
					GameImageID: uuid.UUID(imageID2),
					GameVideoID: uuid.UUID(videoID2),
					Name:        "v1.1.0",
					Description: "アップデート",
					URL:         "https://example.com",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameVersionID3),
					GameID:      uuid.UUID(gameID2),
					GameImageID: uuid.UUID(imageID3),
					GameVideoID: uuid.UUID(videoID3),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
		},
		{
			// 実際には発生しないが、念のため確認
			description: "既にIDが同じバージョンが存在するのでエラー",
			gameID:      gameID3,
			imageID:     imageID4,
			videoID:     videoID4,
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID4,
				values.NewGameVersionName("v1.1.0"),
				values.NewGameVersionDescription("アップデート"),
				now,
			),
			existGame:  true,
			existImage: true,
			existVideo: true,
			beforeGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID4),
					GameID:      uuid.UUID(gameID3),
					GameImageID: uuid.UUID(imageID5),
					GameVideoID: uuid.UUID(videoID5),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now.Add(-time.Hour),
					GameImage: migrate.GameImageTable2{
						ID:          uuid.UUID(imageID5),
						GameID:      uuid.UUID(gameID3),
						ImageTypeID: imageType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
					GameVideo: migrate.GameVideoTable2{
						ID:          uuid.UUID(videoID5),
						GameID:      uuid.UUID(gameID3),
						VideoTypeID: videoType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
				},
			},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID4),
					GameID:      uuid.UUID(gameID3),
					GameImageID: uuid.UUID(imageID5),
					GameVideoID: uuid.UUID(videoID5),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
			isErr: true,
		},
		{
			description: "同名のバージョンが存在してもエラーなし",
			gameID:      gameID4,
			imageID:     imageID6,
			videoID:     videoID6,
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID5,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("アップデート"),
				now,
			),
			existGame:  true,
			existImage: true,
			existVideo: true,
			beforeGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID6),
					GameID:      uuid.UUID(gameID4),
					GameImageID: uuid.UUID(imageID7),
					GameVideoID: uuid.UUID(videoID7),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now.Add(-time.Hour),
					GameImage: migrate.GameImageTable2{
						ID:          uuid.UUID(imageID7),
						GameID:      uuid.UUID(gameID4),
						ImageTypeID: imageType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
					GameVideo: migrate.GameVideoTable2{
						ID:          uuid.UUID(videoID7),
						GameID:      uuid.UUID(gameID4),
						VideoTypeID: videoType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
				},
			},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID5),
					GameID:      uuid.UUID(gameID4),
					GameImageID: uuid.UUID(imageID6),
					GameVideoID: uuid.UUID(videoID6),
					Name:        "v1.0.0",
					Description: "アップデート",
					URL:         "https://example.com",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameVersionID6),
					GameID:      uuid.UUID(gameID4),
					GameImageID: uuid.UUID(imageID7),
					GameVideoID: uuid.UUID(videoID7),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now.Add(-time.Hour),
				},
			},
		},
		{
			description: "バージョン名が32文字でもエラーなし",
			gameID:      gameID5,
			imageID:     imageID8,
			videoID:     videoID8,
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID7,
				values.NewGameVersionName("v1.0.123456789012345678901234567"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:          true,
			existImage:         true,
			existVideo:         true,
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID7),
					GameID:      uuid.UUID(gameID5),
					GameImageID: uuid.UUID(imageID8),
					GameVideoID: uuid.UUID(videoID8),
					Name:        "v1.0.123456789012345678901234567",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "バージョン名が33文字なのでエラー",
			gameID:      gameID6,
			imageID:     values.NewGameImageID(),
			videoID:     values.NewGameVideoID(),
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID8,
				values.NewGameVersionName("v1.0.1234567890123456789012345678"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:          true,
			existImage:         true,
			existVideo:         true,
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{},
			isErr:              true,
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが存在しないのでエラー",
			gameID:      gameID7,
			imageID:     values.NewGameImageID(),
			videoID:     values.NewGameVideoID(),
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID9,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existImage:         true,
			existVideo:         true,
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{},
			isErr:              true,
		},
		{
			// 実際には発生しないが、念のため確認
			description: "バージョン名が空文字でもエラーなし",
			gameID:      gameID8,
			imageID:     imageID9,
			videoID:     videoID9,
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID10,
				values.NewGameVersionName(""),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:          true,
			existImage:         true,
			existVideo:         true,
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID10),
					GameID:      uuid.UUID(gameID8),
					GameImageID: uuid.UUID(imageID9),
					GameVideoID: uuid.UUID(videoID9),
					Name:        "",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "説明が空文字でもエラーなし",
			gameID:      gameID9,
			imageID:     imageID10,
			videoID:     videoID10,
			fileIDs:     []values.GameFileID{},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID11,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription(""),
				now,
			),
			existGame:          true,
			existImage:         true,
			existVideo:         true,
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID11),
					GameID:      uuid.UUID(gameID9),
					GameImageID: uuid.UUID(imageID10),
					GameVideoID: uuid.UUID(videoID10),
					Name:        "v1.0.0",
					Description: "",
					URL:         "https://example.com",
					CreatedAt:   now,
				},
			},
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ファイルidのスライスがnilでもエラーなし",
			gameID:      gameID10,
			imageID:     imageID11,
			videoID:     videoID11,
			fileIDs:     nil,
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID12,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:          true,
			existImage:         true,
			existVideo:         true,
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID12),
					GameID:      uuid.UUID(gameID10),
					GameImageID: uuid.UUID(imageID11),
					GameVideoID: uuid.UUID(videoID11),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "ファイルでもエラーなし",
			gameID:      gameID11,
			imageID:     imageID12,
			videoID:     videoID12,
			fileIDs:     []values.GameFileID{fileID1},
			version: domain.NewGameVersion(
				gameVersionID13,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:  true,
			existImage: true,
			existVideo: true,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID1),
					GameID:     uuid.UUID(gameID11),
					FileTypeID: fileType.ID,
					Hash:       "hash",
					EntryPoint: "/path/to/game.jar",
					CreatedAt:  now,
				},
			},
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID13),
					GameID:      uuid.UUID(gameID11),
					GameImageID: uuid.UUID(imageID12),
					GameVideoID: uuid.UUID(videoID12),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(fileID1),
							GameID:     uuid.UUID(gameID11),
							FileTypeID: fileType.ID,
							Hash:       "hash",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
		},
		{
			description: "ファイルが複数でもエラーなし",
			gameID:      gameID12,
			imageID:     imageID13,
			videoID:     videoID13,
			fileIDs:     []values.GameFileID{fileID2, fileID3},
			version: domain.NewGameVersion(
				gameVersionID14,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:  true,
			existImage: true,
			existVideo: true,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID2),
					GameID:     uuid.UUID(gameID12),
					FileTypeID: fileType.ID,
					Hash:       "hash",
					EntryPoint: "/path/to/game.jar",
					CreatedAt:  now,
				},
				{
					ID:         uuid.UUID(fileID3),
					GameID:     uuid.UUID(gameID12),
					FileTypeID: fileType.ID,
					Hash:       "hash",
					EntryPoint: "/path/to/game.jar",
					CreatedAt:  now.Add(-time.Hour),
				},
			},
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID14),
					GameID:      uuid.UUID(gameID12),
					GameImageID: uuid.UUID(imageID13),
					GameVideoID: uuid.UUID(videoID13),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(fileID2),
							GameID:     uuid.UUID(gameID12),
							FileTypeID: fileType.ID,
							Hash:       "hash",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
						{
							ID:         uuid.UUID(fileID3),
							GameID:     uuid.UUID(gameID12),
							FileTypeID: fileType.ID,
							Hash:       "hash",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now.Add(-time.Hour),
						},
					},
				},
			},
		},
		{
			description: "ファイルとurlが両方あってもエラーなし",
			gameID:      gameID13,
			imageID:     imageID14,
			videoID:     videoID14,
			fileIDs:     []values.GameFileID{fileID4},
			optionURL:   types.NewOption(values.NewGameURLLink(urlLink)),
			version: domain.NewGameVersion(
				gameVersionID15,
				values.NewGameVersionName("v1.0.0"),
				values.NewGameVersionDescription("リリース"),
				now,
			),
			existGame:  true,
			existImage: true,
			existVideo: true,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID4),
					GameID:     uuid.UUID(gameID13),
					FileTypeID: fileType.ID,
					Hash:       "hash",
					EntryPoint: "/path/to/game.jar",
					CreatedAt:  now,
				},
			},
			beforeGameVersions: []migrate.GameVersionTable2{},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID15),
					GameID:      uuid.UUID(gameID13),
					GameImageID: uuid.UUID(imageID14),
					GameVideoID: uuid.UUID(videoID14),
					Name:        "v1.0.0",
					Description: "リリース",
					URL:         "https://example.com",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(fileID4),
							GameID:     uuid.UUID(gameID13),
							FileTypeID: fileType.ID,
							Hash:       "hash",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
		},
		{
			description: "同じファイルを複数バージョンで使っていてもエラーなし",
			gameID:      gameID14,
			imageID:     imageID15,
			videoID:     videoID15,
			fileIDs:     []values.GameFileID{fileID5},
			version: domain.NewGameVersion(
				gameVersionID16,
				values.NewGameVersionName("v1.1.0"),
				values.NewGameVersionDescription("アップデート"),
				now,
			),
			existGame:  true,
			existImage: true,
			existVideo: true,
			files: []migrate.GameFileTable2{
				{
					ID:         uuid.UUID(fileID5),
					GameID:     uuid.UUID(gameID14),
					FileTypeID: fileType.ID,
					Hash:       "hash",
					EntryPoint: "/path/to/game.jar",
					CreatedAt:  now,
				},
			},
			beforeGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID17),
					GameID:      uuid.UUID(gameID14),
					GameImageID: uuid.UUID(imageID15),
					GameVideoID: uuid.UUID(videoID15),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
					GameImage: migrate.GameImageTable2{
						ID:          uuid.UUID(imageID16),
						GameID:      uuid.UUID(gameID14),
						ImageTypeID: imageType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
					GameVideo: migrate.GameVideoTable2{
						ID:          uuid.UUID(videoID16),
						GameID:      uuid.UUID(gameID14),
						VideoTypeID: videoType.ID,
						CreatedAt:   now.Add(-time.Hour),
					},
					GameFiles: []migrate.GameFileTable2{
						{
							ID: uuid.UUID(fileID5),
						},
					},
				},
			},
			expectGameVersions: []migrate.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID16),
					GameID:      uuid.UUID(gameID14),
					GameImageID: uuid.UUID(imageID15),
					GameVideoID: uuid.UUID(videoID15),
					Name:        "v1.1.0",
					Description: "アップデート",
					CreatedAt:   now,
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(fileID5),
							GameID:     uuid.UUID(gameID14),
							FileTypeID: fileType.ID,
							Hash:       "hash",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
				{
					ID:          uuid.UUID(gameVersionID17),
					GameID:      uuid.UUID(gameID14),
					GameImageID: uuid.UUID(imageID16),
					GameVideoID: uuid.UUID(videoID16),
					Name:        "v1.0.0",
					Description: "リリース",
					CreatedAt:   now.Add(-time.Hour),
					GameFiles: []migrate.GameFileTable2{
						{
							ID:         uuid.UUID(fileID5),
							GameID:     uuid.UUID(gameID14),
							FileTypeID: fileType.ID,
							Hash:       "hash",
							EntryPoint: "/path/to/game.jar",
							CreatedAt:  now,
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.existGame {
				err := db.
					Session(&gorm.Session{}).
					Create(&migrate.GameTable2{
						ID:          uuid.UUID(testCase.gameID),
						Name:        "test",
						Description: "test",
						CreatedAt:   time.Now(),
					}).Error
				if err != nil {
					t.Fatalf("failed to create game table: %+v\n", err)
				}

				if len(testCase.files) != 0 {
					err := db.
						Session(&gorm.Session{}).
						Create(&testCase.files).Error
					if err != nil {
						t.Fatal("failed to create game files:", err)
					}
				}

				if len(testCase.beforeGameVersions) != 0 {
					err = db.
						Session(&gorm.Session{}).
						Create(&testCase.beforeGameVersions).Error
					if err != nil {
						t.Fatalf("failed to create game version table: %+v\n", err)
					}
				}

				if testCase.existImage {
					err := db.
						Session(&gorm.Session{}).
						Create(&migrate.GameImageTable2{
							ID:          uuid.UUID(testCase.imageID),
							GameID:      uuid.UUID(testCase.gameID),
							ImageTypeID: imageType.ID,
							CreatedAt:   time.Now(),
						}).Error
					if err != nil {
						t.Fatalf("failed to create game image table: %+v\n", err)
					}
				}

				if testCase.existVideo {
					err := db.
						Session(&gorm.Session{}).
						Create(&migrate.GameVideoTable2{
							ID:          uuid.UUID(testCase.videoID),
							GameID:      uuid.UUID(testCase.gameID),
							VideoTypeID: videoType.ID,
							CreatedAt:   time.Now(),
						}).Error
					if err != nil {
						t.Fatalf("failed to create game video table: %+v\n", err)
					}
				}
			}

			err := gameVersionRepository.CreateGameVersion(
				ctx,
				testCase.gameID,
				testCase.imageID,
				testCase.videoID,
				testCase.optionURL,
				testCase.fileIDs,
				testCase.version,
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

			var gameVersions []migrate.GameVersionTable2
			err = db.
				Where("game_id = ?", uuid.UUID(testCase.gameID)).
				Preload("GameFiles").
				Find(&gameVersions).Error
			if err != nil {
				t.Fatalf("failed to get game versions: %+v\n", err)
			}

			assert.Len(t, gameVersions, len(testCase.expectGameVersions))

			versionMap := make(map[uuid.UUID]migrate.GameVersionTable2, len(gameVersions))
			for _, version := range gameVersions {
				versionMap[version.ID] = version
			}

			for _, expectVersion := range testCase.expectGameVersions {
				actualVersion, ok := versionMap[expectVersion.ID]
				if !ok {
					t.Errorf("failed to find version: %s", expectVersion.Name)
					continue
				}

				assert.Equal(t, expectVersion.ID, actualVersion.ID)
				assert.Equal(t, expectVersion.GameID, actualVersion.GameID)
				assert.Equal(t, expectVersion.Name, actualVersion.Name)
				assert.Equal(t, expectVersion.Description, actualVersion.Description)
				assert.Equal(t, expectVersion.URL, actualVersion.URL)
				assert.WithinDuration(t, expectVersion.CreatedAt, actualVersion.CreatedAt, 2*time.Second)

				assert.Len(t, actualVersion.GameFiles, len(expectVersion.GameFiles))

				fileMap := make(map[uuid.UUID]migrate.GameFileTable2, len(actualVersion.GameFiles))
				for _, file := range actualVersion.GameFiles {
					fileMap[file.ID] = file
				}

				for _, expectFile := range expectVersion.GameFiles {
					actualFile, ok := fileMap[expectFile.ID]
					if !ok {
						t.Errorf("failed to find file: %s", expectFile.EntryPoint)
						continue
					}

					assert.Equal(t, expectFile.ID, actualFile.ID)
					assert.Equal(t, expectFile.GameID, actualFile.GameID)
					assert.Equal(t, expectFile.FileTypeID, actualFile.FileTypeID)
					assert.Equal(t, expectFile.Hash, actualFile.Hash)
					assert.Equal(t, expectFile.EntryPoint, actualFile.EntryPoint)
					assert.WithinDuration(t, expectFile.CreatedAt, actualFile.CreatedAt, 2*time.Second)
				}
			}
		})
	}
}

func TestGetGameVersionsV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	gameVersionRepository := NewGameVersionV2(testDB)

	type test struct {
		description           string
		gameID                values.GameID
		limit                 uint
		offset                uint
		lockType              repository.LockType
		games                 []migrate.GameTable2
		expectNum             uint
		expectGameVersionInfo []*repository.GameVersionInfo
		isErr                 bool
		err                   error
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
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

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()
	gameVersionID8 := values.NewGameVersionID()
	gameVersionID9 := values.NewGameVersionID()
	gameVersionID10 := values.NewGameVersionID()
	gameVersionID11 := values.NewGameVersionID()
	gameVersionID12 := values.NewGameVersionID()
	gameVersionID13 := values.NewGameVersionID()
	gameVersionID14 := values.NewGameVersionID()
	gameVersionID15 := values.NewGameVersionID()

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

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()

	var imageType migrate.GameImageTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameImageTypeJpeg).
		Select("id").
		Take(&imageType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	var videoType migrate.GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameVideoTypeMp4).
		Select("id").
		Take(&videoType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	var fileType migrate.GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameFileTypeJar).
		Select("id").
		Take(&fileType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID1),
							GameID:      uuid.UUID(gameID1),
							GameImageID: uuid.UUID(imageID1),
							GameVideoID: uuid.UUID(videoID1),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID1),
								GameID:      uuid.UUID(gameID1),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID1),
								GameID:      uuid.UUID(gameID1),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectNum: 1,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						time.Now(),
					),
					ImageID: imageID1,
					VideoID: videoID1,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが存在しなくてもエラーなし",
			gameID:      gameID2,
			games:       []migrate.GameTable2{},
			expectNum:   0,
		},
		{
			description: "バージョンが複数あってもエラーなし",
			gameID:      gameID4,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID3),
							GameID:      uuid.UUID(gameID4),
							GameImageID: uuid.UUID(imageID2),
							GameVideoID: uuid.UUID(videoID2),
							Name:        "v1.1.0",
							Description: "アップデート",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID2),
								GameID:      uuid.UUID(gameID4),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID2),
								GameID:      uuid.UUID(gameID4),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
						{
							ID:          uuid.UUID(gameVersionID4),
							GameID:      uuid.UUID(gameID4),
							GameImageID: uuid.UUID(imageID3),
							GameVideoID: uuid.UUID(videoID3),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now().Add(-time.Hour),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID3),
								GameID:      uuid.UUID(gameID4),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID3),
								GameID:      uuid.UUID(gameID4),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectNum: 2,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID3,
						values.NewGameVersionName("v1.1.0"),
						values.NewGameVersionDescription("アップデート"),
						time.Now(),
					),
					ImageID: imageID2,
					VideoID: videoID2,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID4,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						time.Now().Add(-time.Hour),
					),
					ImageID: imageID3,
					VideoID: videoID3,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
		},
		{
			description: "バージョンが存在しなくてもエラーなし",
			gameID:      gameID5,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			expectNum:             0,
			expectGameVersionInfo: []*repository.GameVersionInfo{},
		},
		{
			description: "別のゲームのバージョンが混ざることはない",
			gameID:      gameID6,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID5),
							GameID:      uuid.UUID(gameID6),
							GameImageID: uuid.UUID(imageID4),
							GameVideoID: uuid.UUID(videoID4),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID4),
								GameID:      uuid.UUID(gameID6),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID4),
								GameID:      uuid.UUID(gameID6),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID6),
							GameID:      uuid.UUID(gameID7),
							GameImageID: uuid.UUID(imageID5),
							GameVideoID: uuid.UUID(videoID5),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID5),
								GameID:      uuid.UUID(gameID7),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID5),
								GameID:      uuid.UUID(gameID7),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectNum: 1,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID5,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						time.Now(),
					),
					ImageID: imageID4,
					VideoID: videoID4,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
		},
		{
			description: "ファイルが存在してもエラーなし",
			gameID:      gameID8,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID8),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID7),
							GameID:      uuid.UUID(gameID8),
							GameImageID: uuid.UUID(imageID6),
							GameVideoID: uuid.UUID(videoID6),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID6),
								GameID:      uuid.UUID(gameID8),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID6),
								GameID:      uuid.UUID(gameID8),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
							GameFiles: []migrate.GameFileTable2{
								{
									ID:         uuid.UUID(fileID1),
									GameID:     uuid.UUID(gameID8),
									FileTypeID: fileType.ID,
									Hash:       "hash",
									EntryPoint: "/path/to/game.exe",
									CreatedAt:  now,
								},
							},
						},
					},
				},
			},
			expectNum: 1,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID7,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						time.Now(),
					),
					ImageID: imageID6,
					VideoID: videoID6,
					FileIDs: []values.GameFileID{fileID1},
				},
			},
		},
		{
			description: "ファイルが複数でもエラーなし",
			gameID:      gameID9,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID9),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID8),
							GameID:      uuid.UUID(gameID9),
							GameImageID: uuid.UUID(imageID7),
							GameVideoID: uuid.UUID(videoID7),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID7),
								GameID:      uuid.UUID(gameID9),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID7),
								GameID:      uuid.UUID(gameID9),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
							GameFiles: []migrate.GameFileTable2{
								{
									ID:         uuid.UUID(fileID2),
									GameID:     uuid.UUID(gameID9),
									FileTypeID: fileType.ID,
									Hash:       "hash",
									EntryPoint: "/path/to/game.exe",
									CreatedAt:  now,
								},
								{
									ID:         uuid.UUID(fileID3),
									GameID:     uuid.UUID(gameID9),
									FileTypeID: fileType.ID,
									Hash:       "hash",
									EntryPoint: "/path/to/game.exe",
									CreatedAt:  now,
								},
							},
						},
					},
				},
			},
			expectNum: 1,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID8,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						time.Now(),
					),
					ImageID: imageID7,
					VideoID: videoID7,
					FileIDs: []values.GameFileID{fileID2, fileID3},
				},
			},
		},
		{
			description: "limitが存在してもエラーなし",
			gameID:      gameID10,
			limit:       1,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID10),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID9),
							GameID:      uuid.UUID(gameID10),
							GameImageID: uuid.UUID(imageID8),
							GameVideoID: uuid.UUID(videoID8),
							Name:        "v1.1.0",
							Description: "アップデート",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID8),
								GameID:      uuid.UUID(gameID10),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID8),
								GameID:      uuid.UUID(gameID10),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
						{
							ID:          uuid.UUID(gameVersionID10),
							GameID:      uuid.UUID(gameID10),
							GameImageID: uuid.UUID(imageID9),
							GameVideoID: uuid.UUID(videoID9),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now().Add(-time.Hour),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID9),
								GameID:      uuid.UUID(gameID10),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID9),
								GameID:      uuid.UUID(gameID10),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectNum: 2,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID9,
						values.NewGameVersionName("v1.1.0"),
						values.NewGameVersionDescription("アップデート"),
						time.Now(),
					),
					ImageID: imageID8,
					VideoID: videoID8,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
		},
		{
			description: "limit,offsetが存在してもエラーなし",
			gameID:      gameID11,
			limit:       1,
			offset:      1,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID11),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID11),
							GameID:      uuid.UUID(gameID11),
							GameImageID: uuid.UUID(imageID10),
							GameVideoID: uuid.UUID(videoID10),
							Name:        "v1.1.0",
							Description: "アップデート",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID10),
								GameID:      uuid.UUID(gameID11),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID10),
								GameID:      uuid.UUID(gameID11),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
						{
							ID:          uuid.UUID(gameVersionID12),
							GameID:      uuid.UUID(gameID11),
							GameImageID: uuid.UUID(imageID11),
							GameVideoID: uuid.UUID(videoID11),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now().Add(-time.Hour),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID11),
								GameID:      uuid.UUID(gameID11),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID11),
								GameID:      uuid.UUID(gameID11),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectNum: 2,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID12,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						time.Now().Add(-time.Hour),
					),
					ImageID: imageID11,
					VideoID: videoID11,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
		},
		{
			description: "limitなし、offsetありなのでエラー",
			gameID:      gameID12,
			offset:      1,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID12),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID13),
							GameID:      uuid.UUID(gameID12),
							GameImageID: uuid.UUID(imageID12),
							GameVideoID: uuid.UUID(videoID12),
							Name:        "v1.1.0",
							Description: "アップデート",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID12),
								GameID:      uuid.UUID(gameID12),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID12),
								GameID:      uuid.UUID(gameID12),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
						{
							ID:          uuid.UUID(gameVersionID14),
							GameID:      uuid.UUID(gameID12),
							GameImageID: uuid.UUID(imageID13),
							GameVideoID: uuid.UUID(videoID13),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now().Add(-time.Hour),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID13),
								GameID:      uuid.UUID(gameID12),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID13),
								GameID:      uuid.UUID(gameID12),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			isErr: true,
		},
		{
			description: "lockありでもエラーなし",
			gameID:      gameID13,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID13),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID15),
							GameID:      uuid.UUID(gameID13),
							GameImageID: uuid.UUID(imageID14),
							GameVideoID: uuid.UUID(videoID14),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID14),
								GameID:      uuid.UUID(gameID13),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID14),
								GameID:      uuid.UUID(gameID13),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectNum: 1,
			expectGameVersionInfo: []*repository.GameVersionInfo{
				{
					GameVersion: domain.NewGameVersion(
						gameVersionID15,
						values.NewGameVersionName("v1.0.0"),
						values.NewGameVersionDescription("リリース"),
						time.Now(),
					),
					ImageID: imageID14,
					VideoID: videoID14,
					URL:     types.NewOption(values.NewGameURLLink(urlLink)),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.games) != 0 {
				err := db.Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("failed to create games: %v", err)
				}
			}

			num, gameVersions, err := gameVersionRepository.GetGameVersions(
				ctx,
				testCase.gameID,
				testCase.limit,
				testCase.offset,
				testCase.lockType,
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

			assert.Equal(t, testCase.expectNum, num)
			assert.Len(t, gameVersions, len(testCase.expectGameVersionInfo))
			for i, expectVersion := range testCase.expectGameVersionInfo {
				actualVersion := gameVersions[i]

				assert.Equal(t, expectVersion.GameVersion.GetID(), actualVersion.GameVersion.GetID())
				assert.Equal(t, expectVersion.GameVersion.GetName(), actualVersion.GameVersion.GetName())
				assert.Equal(t, expectVersion.GameVersion.GetDescription(), actualVersion.GameVersion.GetDescription())
				assert.WithinDuration(t, expectVersion.GameVersion.GetCreatedAt(), actualVersion.GameVersion.GetCreatedAt(), 2*time.Second)
				assert.Equal(t, expectVersion.ImageID, actualVersion.ImageID)
				assert.Equal(t, expectVersion.VideoID, actualVersion.VideoID)
				assert.Equal(t, expectVersion.URL, actualVersion.URL)

				assert.Len(t, actualVersion.FileIDs, len(expectVersion.FileIDs))

				fileIDMap := make(map[values.GameFileID]struct{}, len(actualVersion.FileIDs))
				for _, fileID := range actualVersion.FileIDs {
					fileIDMap[fileID] = struct{}{}
				}

				for _, expectFileID := range expectVersion.FileIDs {
					_, ok := fileIDMap[expectFileID]
					assert.True(t, ok)
				}
			}
		})
	}
}

func TestGetLatestGameVersionV2(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	gameVersionRepository := NewGameVersionV2(testDB)

	type test struct {
		description           string
		gameID                values.GameID
		lockType              repository.LockType
		games                 []migrate.GameTable2
		expectGameVersionInfo *repository.GameVersionInfo
		isErr                 bool
		err                   error
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

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()
	gameVersionID5 := values.NewGameVersionID()
	gameVersionID6 := values.NewGameVersionID()
	gameVersionID7 := values.NewGameVersionID()
	gameVersionID8 := values.NewGameVersionID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	imageID5 := values.NewGameImageID()
	imageID6 := values.NewGameImageID()
	imageID7 := values.NewGameImageID()
	imageID8 := values.NewGameImageID()

	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()
	videoID5 := values.NewGameVideoID()
	videoID6 := values.NewGameVideoID()
	videoID7 := values.NewGameVideoID()
	videoID8 := values.NewGameVideoID()

	fileID1 := values.NewGameFileID()
	fileID2 := values.NewGameFileID()
	fileID3 := values.NewGameFileID()

	var imageType migrate.GameImageTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameImageTypeJpeg).
		Select("id").
		Take(&imageType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	var videoType migrate.GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameVideoTypeMp4).
		Select("id").
		Take(&videoType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	var fileType migrate.GameFileTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", migrate.GameFileTypeJar).
		Select("id").
		Take(&fileType).Error
	if err != nil {
		t.Fatalf("failed to get role type table: %+v\n", err)
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      gameID1,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID1),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID1),
							GameID:      uuid.UUID(gameID1),
							GameImageID: uuid.UUID(imageID1),
							GameVideoID: uuid.UUID(videoID1),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID1),
								GameID:      uuid.UUID(gameID1),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID1),
								GameID:      uuid.UUID(gameID1),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectGameVersionInfo: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID1,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now(),
				),
				ImageID: imageID1,
				VideoID: videoID1,
				URL:     types.NewOption(values.NewGameURLLink(urlLink)),
			},
		},
		{
			// 実際には発生しないが、念のため確認
			description: "ゲームが存在しないのでRecordNotFound",
			gameID:      gameID2,
			games:       []migrate.GameTable2{},
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
		{
			description: "バージョンが複数あっても最新のものを取得",
			gameID:      gameID3,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID3),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID2),
							GameID:      uuid.UUID(gameID3),
							GameImageID: uuid.UUID(imageID2),
							GameVideoID: uuid.UUID(videoID2),
							Name:        "v1.1.0",
							Description: "アップデート",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID2),
								GameID:      uuid.UUID(gameID3),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID2),
								GameID:      uuid.UUID(gameID3),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
						{
							ID:          uuid.UUID(gameVersionID3),
							GameID:      uuid.UUID(gameID3),
							GameImageID: uuid.UUID(imageID3),
							GameVideoID: uuid.UUID(videoID3),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now().Add(-time.Hour),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID3),
								GameID:      uuid.UUID(gameID3),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID3),
								GameID:      uuid.UUID(gameID3),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectGameVersionInfo: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID2,
					values.NewGameVersionName("v1.1.0"),
					values.NewGameVersionDescription("アップデート"),
					time.Now(),
				),
				ImageID: imageID2,
				VideoID: videoID2,
				URL:     types.NewOption(values.NewGameURLLink(urlLink)),
			},
		},
		{
			description: "バージョンが存在しないのでRecordNotFound",
			gameID:      gameID4,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID4),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
				},
			},
			isErr: true,
			err:   repository.ErrRecordNotFound,
		},
		{
			description: "別のゲームのバージョンが混ざることはない",
			gameID:      gameID5,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID5),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID4),
							GameID:      uuid.UUID(gameID5),
							GameImageID: uuid.UUID(imageID4),
							GameVideoID: uuid.UUID(videoID4),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID4),
								GameID:      uuid.UUID(gameID5),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID4),
								GameID:      uuid.UUID(gameID5),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
				{
					ID:          uuid.UUID(gameID6),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID5),
							GameID:      uuid.UUID(gameID6),
							GameImageID: uuid.UUID(imageID5),
							GameVideoID: uuid.UUID(videoID5),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID5),
								GameID:      uuid.UUID(gameID6),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID5),
								GameID:      uuid.UUID(gameID6),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectGameVersionInfo: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID4,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now(),
				),
				ImageID: imageID4,
				VideoID: videoID4,
				URL:     types.NewOption(values.NewGameURLLink(urlLink)),
			},
		},
		{
			description: "ファイルが存在してもエラーなし",
			gameID:      gameID7,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID7),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID6),
							GameID:      uuid.UUID(gameID7),
							GameImageID: uuid.UUID(imageID6),
							GameVideoID: uuid.UUID(videoID6),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID6),
								GameID:      uuid.UUID(gameID7),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID6),
								GameID:      uuid.UUID(gameID7),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
							GameFiles: []migrate.GameFileTable2{
								{
									ID:         uuid.UUID(fileID1),
									GameID:     uuid.UUID(gameID7),
									FileTypeID: fileType.ID,
									Hash:       "hash",
									EntryPoint: "/path/to/game.exe",
									CreatedAt:  now,
								},
							},
						},
					},
				},
			},
			expectGameVersionInfo: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID6,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now(),
				),
				ImageID: imageID6,
				VideoID: videoID6,
				FileIDs: []values.GameFileID{fileID1},
			},
		},
		{
			description: "ファイルが複数でもエラーなし",
			gameID:      gameID8,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID8),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID7),
							GameID:      uuid.UUID(gameID8),
							GameImageID: uuid.UUID(imageID7),
							GameVideoID: uuid.UUID(videoID7),
							Name:        "v1.0.0",
							Description: "リリース",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID7),
								GameID:      uuid.UUID(gameID8),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID7),
								GameID:      uuid.UUID(gameID8),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
							GameFiles: []migrate.GameFileTable2{
								{
									ID:         uuid.UUID(fileID2),
									GameID:     uuid.UUID(gameID8),
									FileTypeID: fileType.ID,
									Hash:       "hash",
									EntryPoint: "/path/to/game.exe",
									CreatedAt:  now,
								},
								{
									ID:         uuid.UUID(fileID3),
									GameID:     uuid.UUID(gameID8),
									FileTypeID: fileType.ID,
									Hash:       "hash",
									EntryPoint: "/path/to/game.exe",
									CreatedAt:  now,
								},
							},
						},
					},
				},
			},
			expectGameVersionInfo: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID7,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now(),
				),
				ImageID: imageID7,
				VideoID: videoID7,
				FileIDs: []values.GameFileID{fileID2, fileID3},
			},
		},
		{
			description: "lockありでもエラーなし",
			gameID:      gameID9,
			games: []migrate.GameTable2{
				{
					ID:          uuid.UUID(gameID9),
					Name:        "test",
					Description: "test",
					CreatedAt:   time.Now(),
					GameVersions: []migrate.GameVersionTable2{
						{
							ID:          uuid.UUID(gameVersionID8),
							GameID:      uuid.UUID(gameID9),
							GameImageID: uuid.UUID(imageID8),
							GameVideoID: uuid.UUID(videoID8),
							Name:        "v1.0.0",
							Description: "リリース",
							URL:         "https://example.com",
							CreatedAt:   time.Now(),
							GameImage: migrate.GameImageTable2{
								ID:          uuid.UUID(imageID8),
								GameID:      uuid.UUID(gameID9),
								ImageTypeID: imageType.ID,
								CreatedAt:   now,
							},
							GameVideo: migrate.GameVideoTable2{
								ID:          uuid.UUID(videoID8),
								GameID:      uuid.UUID(gameID9),
								VideoTypeID: videoType.ID,
								CreatedAt:   now,
							},
						},
					},
				},
			},
			expectGameVersionInfo: &repository.GameVersionInfo{
				GameVersion: domain.NewGameVersion(
					gameVersionID8,
					values.NewGameVersionName("v1.0.0"),
					values.NewGameVersionDescription("リリース"),
					time.Now(),
				),
				ImageID: imageID8,
				VideoID: videoID8,
				URL:     types.NewOption(values.NewGameURLLink(urlLink)),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if len(testCase.games) != 0 {
				err := db.Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("failed to create games: %v", err)
				}
			}

			actualVersion, err := gameVersionRepository.GetLatestGameVersion(
				ctx,
				testCase.gameID,
				testCase.lockType,
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

			expectVersion := testCase.expectGameVersionInfo
			assert.Equal(t, expectVersion.GameVersion.GetID(), actualVersion.GameVersion.GetID())
			assert.Equal(t, expectVersion.GameVersion.GetName(), actualVersion.GameVersion.GetName())
			assert.Equal(t, expectVersion.GameVersion.GetDescription(), actualVersion.GameVersion.GetDescription())
			assert.WithinDuration(t, expectVersion.GameVersion.GetCreatedAt(), actualVersion.GameVersion.GetCreatedAt(), 2*time.Second)
			assert.Equal(t, expectVersion.ImageID, actualVersion.ImageID)
			assert.Equal(t, expectVersion.VideoID, actualVersion.VideoID)
			assert.Equal(t, expectVersion.URL, actualVersion.URL)

			assert.Len(t, actualVersion.FileIDs, len(expectVersion.FileIDs))

			fileIDMap := make(map[values.GameFileID]struct{}, len(actualVersion.FileIDs))
			for _, fileID := range actualVersion.FileIDs {
				fileIDMap[fileID] = struct{}{}
			}

			for _, expectFileID := range expectVersion.FileIDs {
				_, ok := fileIDMap[expectFileID]
				assert.True(t, ok)
			}
		})
	}
}
