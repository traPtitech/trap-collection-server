package gorm2

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
)

type Game struct {
	db *DB
}

func NewGame(db *DB) *Game {
	return &Game{
		db: db,
	}
}

func (g *Game) GetGame(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.Game, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock type: %w", err)
	}

	var game GameTable
	err = db.
		Where("id = ?", uuid.UUID(gameID)).
		Take(&game).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return domain.NewGame(
		values.NewGameIDFromUUID(game.ID),
		values.NewGameName(game.Name),
		values.NewGameDescription(game.Description),
		game.CreatedAt,
	), nil
}

func (g *Game) GetGamesByIDs(ctx context.Context, gameIDs []values.GameID, lockType repository.LockType) ([]*domain.Game, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock type: %w", err)
	}

	uuidGameIDs := make([]uuid.UUID, 0, len(gameIDs))
	for _, gameID := range gameIDs {
		uuidGameIDs = append(uuidGameIDs, uuid.UUID(gameID))
	}

	var games []GameTable
	err = db.
		Where("id IN ?", uuidGameIDs).
		Find(&games).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}

	gamesDomain := make([]*domain.Game, 0, len(games))
	for _, game := range games {
		gamesDomain = append(gamesDomain, domain.NewGame(
			values.NewGameIDFromUUID(game.ID),
			values.NewGameName(game.Name),
			values.NewGameDescription(game.Description),
			game.CreatedAt,
		))
	}

	return gamesDomain, nil
}

func (g *Game) GetGamesByLauncherVersion(ctx context.Context, launcherVersionID values.LauncherVersionID) ([]*domain.Game, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var launcherVersion LauncherVersionTable
	err = db.
		Where("id = ?", uuid.UUID(launcherVersionID)).
		Preload("Games", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at desc")
		}).
		Find(&launcherVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	gamesDomain := make([]*domain.Game, 0, len(launcherVersion.Games))
	for _, game := range launcherVersion.Games {
		gamesDomain = append(gamesDomain, domain.NewGame(
			values.NewGameIDFromUUID(game.ID),
			values.NewGameName(game.Name),
			values.NewGameDescription(game.Description),
			game.CreatedAt,
		))
	}

	return gamesDomain, nil
}

func (g *Game) GetGameInfosByLauncherVersion(ctx context.Context, launcherVersionID values.LauncherVersionID, fileTypes []values.GameFileType) ([]*repository.GameInfo, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	strFileTypes := make([]string, 0, len(fileTypes))
	for _, fileType := range fileTypes {
		switch fileType {
		case values.GameFileTypeJar:
			strFileTypes = append(strFileTypes, gameFileTypeJar)
		case values.GameFileTypeWindows:
			strFileTypes = append(strFileTypes, gameFileTypeWindows)
		case values.GameFileTypeMac:
			strFileTypes = append(strFileTypes, gameFileTypeMac)
		default:
			return nil, fmt.Errorf("invalid file type: %d", fileType)
		}
	}

	var launcherVersion LauncherVersionTable
	err = db.
		Where("id = ?", uuid.UUID(launcherVersionID)).
		Preload("Games", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at desc")
		}).
		Preload("Games.GameImages", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("GameImageType").
				Joins("JOIN (" +
					"SELECT game_id, MAX(created_at) AS created_at FROM game_images GROUP BY game_id" +
					") as max_images ON game_images.game_id = max_images.game_id AND game_images.created_at = max_images.created_at")
		}).
		Preload("Games.GameVideos", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("GameVideoType").
				Joins("JOIN (" +
					"SELECT game_id, MAX(created_at) AS created_at FROM game_videos GROUP BY game_id" +
					") as max_videos ON game_videos.game_id = max_videos.game_id AND game_videos.created_at = max_videos.created_at")
		}).
		Preload("Games.GameVersions", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("GameURL").
				Joins("JOIN (" +
					"SELECT game_id, MAX(created_at) AS created_at FROM game_versions GROUP BY game_id" +
					") as max_versions ON game_versions.game_id = max_versions.game_id AND game_versions.created_at = max_versions.created_at")
		}).
		Preload("Games.GameVersions.GameFiles", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("GameFileType").
				Order("created_at desc").
				Where("GameFileType.name IN ?", strFileTypes)
		}).
		Take(&launcherVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	games := make([]*repository.GameInfo, 0, len(launcherVersion.Games))
	for _, game := range launcherVersion.Games {
		// バージョンが存在しないゲームは除外
		if game.GameVersions == nil || len(game.GameVersions) == 0 {
			continue
		}

		var gameImage *domain.GameImage
		if len(game.GameImages) != 0 {
			var imageType values.GameImageType
			switch game.GameImages[0].GameImageType.Name {
			case gameImageTypeJpeg:
				imageType = values.GameImageTypeJpeg
			case gameImageTypePng:
				imageType = values.GameImageTypePng
			case gameImageTypeGif:
				imageType = values.GameImageTypeGif
			default:
				// ゲームの画像の種類の1つが誤っているだけでランチャーを動かなくしたくないのでエラーにしない
				log.Printf("error: invalid game image type: %s", game.GameImages[0].GameImageType.Name)
				continue
			}

			gameImage = domain.NewGameImage(
				values.GameImageIDFromUUID(game.GameImages[0].ID),
				imageType,
				game.GameImages[0].CreatedAt,
			)
		} else {
			continue
		}

		var gameVideo *domain.GameVideo
		if len(game.GameVideos) != 0 {
			var videoType values.GameVideoType
			switch game.GameVideos[0].GameVideoType.Name {
			case gameVideoTypeMp4:
				videoType = values.GameVideoTypeMp4
			default:
				// ゲームの動画の種類の1つが誤っているだけでランチャーを動かなくしたくないのでエラーにしない
				log.Printf("error: invalid game video type: %s", game.GameVideos[0].GameVideoType.Name)
				goto VideoEnd
			}

			gameVideo = domain.NewGameVideo(
				values.NewGameVideoIDFromUUID(game.GameVideos[0].ID),
				videoType,
				game.GameVideos[0].CreatedAt,
			)
		}
	VideoEnd:

		var gameURL *domain.GameURL
		if game.GameVersions[0].GameURL.ID != [16]byte{} {
			link, err := url.Parse(game.GameVersions[0].GameURL.URL)
			if err != nil {
				// ゲームのURLの1つが不正なだけでランチャーを動かなくしたくはないので、returnはしない
				log.Printf("error: failed to parse game url(%s): %v", game.GameVersions[0].GameURL.URL, err)
				goto URLEnd
			}

			gameURL = domain.NewGameURL(
				values.NewGameURLIDFromUUID(game.GameVersions[0].GameURL.ID),
				values.NewGameURLLink(link),
				game.GameVersions[0].GameURL.CreatedAt,
			)
		}
	URLEnd:

		var gameFiles []*domain.GameFile
		if game.GameVersions[0].GameFiles != nil {
			for _, gameFile := range game.GameVersions[0].GameFiles {
				var fileType values.GameFileType
				switch gameFile.GameFileType.Name {
				case gameFileTypeJar:
					fileType = values.GameFileTypeJar
				case gameFileTypeWindows:
					fileType = values.GameFileTypeWindows
				case gameFileTypeMac:
					fileType = values.GameFileTypeMac
				default:
					// ゲームのファイルの種類の1つが誤っているだけでランチャーを動かなくしたくないのでエラーにしない
					log.Printf("error: invalid game file type: %s", gameFile.GameFileType.Name)
					continue
				}

				bytesHash, err := hex.DecodeString(gameFile.Hash)
				if err != nil {
					// ゲームのファイルのハッシュ値の1つが不正なだけでランチャーを動かなくしたくはないので、returnはしない
					log.Printf("error: failed to parse game file hash(%s): %v", gameFile.Hash, err)
					continue
				}

				gameFiles = append(gameFiles, domain.NewGameFile(
					values.NewGameFileIDFromUUID(gameFile.ID),
					fileType,
					values.NewGameFileEntryPoint(gameFile.EntryPoint),
					values.NewGameFileHashFromBytes(bytesHash),
					gameFile.CreatedAt,
				))
			}
		}

		games = append(games, &repository.GameInfo{
			Game: domain.NewGame(
				values.NewGameIDFromUUID(game.ID),
				values.NewGameName(game.Name),
				values.NewGameDescription(game.Description),
				game.CreatedAt,
			),
			LatestVersion: domain.NewGameVersion(
				values.NewGameVersionIDFromUUID(game.GameVersions[0].ID),
				values.NewGameVersionName(game.GameVersions[0].Name),
				values.NewGameVersionDescription(game.GameVersions[0].Description),
				game.GameVersions[0].CreatedAt,
			),
			LatestURL:   gameURL,
			LatestFiles: gameFiles,
			LatestImage: gameImage,
			LatestVideo: gameVideo,
		})
	}

	return games, nil
}
