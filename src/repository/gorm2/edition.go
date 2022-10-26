package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

var _ repository.Edition = (*Edition)(nil)

type Edition struct {
	db *DB
}

func NewEdition(db *DB) *Edition {
	return &Edition{
		db: db,
	}
}

func (e *Edition) SaveEdition(ctx context.Context, edition *domain.LauncherVersion) error {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	questionnaireURL, err := edition.GetQuestionnaireURL()
	if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
		return fmt.Errorf("failed to get questionnaire url: %w", err)
	}

	var strQuestionnaireURL sql.NullString
	if errors.Is(err, domain.ErrNoQuestionnaire) {
		strQuestionnaireURL = sql.NullString{
			Valid: false,
		}
	} else {
		strQuestionnaireURL = sql.NullString{
			String: (*url.URL)(questionnaireURL).String(),
			Valid:  true,
		}
	}

	err = db.
		Create(&migrate.EditionTable2{
			ID:               uuid.UUID(edition.GetID()),
			Name:             string(edition.GetName()),
			QuestionnaireURL: strQuestionnaireURL,
			CreatedAt:        edition.GetCreatedAt(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to save edition: %w", err)
	}

	return nil
}

func (e *Edition) UpdateEdition(ctx context.Context, edition *domain.LauncherVersion) error {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	questionnaireURL, err := edition.GetQuestionnaireURL()
	if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
		return fmt.Errorf("failed to get questionnaire url: %w", err)
	}

	var strQuestionnaireURL sql.NullString
	if errors.Is(err, domain.ErrNoQuestionnaire) {
		strQuestionnaireURL = sql.NullString{
			Valid: false,
		}
	} else {
		strQuestionnaireURL = sql.NullString{
			String: (*url.URL)(questionnaireURL).String(),
			Valid:  true,
		}
	}

	result := db.
		Where("id = ?", uuid.UUID(edition.GetID())).
		Updates(migrate.EditionTable2{
			Name:             string(edition.GetName()),
			QuestionnaireURL: strQuestionnaireURL,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update edition: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}

func (e *Edition) DeleteEdition(ctx context.Context, editionID values.LauncherVersionID) error {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := db.
		Where("id = ?", uuid.UUID(editionID)).
		Delete(&migrate.EditionTable2{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete edition: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}

func (e *Edition) GetEditions(ctx context.Context, lockType repository.LockType) ([]*domain.LauncherVersion, error) {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var editions []*migrate.EditionTable2
	err = db.
		Find(&editions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get editions: %w", err)
	}

	var result []*domain.LauncherVersion
	for _, edition := range editions {
		var domainEdition *domain.LauncherVersion
		if edition.QuestionnaireURL.Valid {
			questionnaireURL, err := url.Parse(edition.QuestionnaireURL.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse questionnaire url: %w", err)
			}

			domainEdition = domain.NewLauncherVersionWithQuestionnaire(
				values.NewLauncherVersionIDFromUUID(edition.ID),
				values.NewLauncherVersionName(edition.Name),
				values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
				edition.CreatedAt,
			)
		} else {
			domainEdition = domain.NewLauncherVersionWithoutQuestionnaire(
				values.NewLauncherVersionIDFromUUID(edition.ID),
				values.NewLauncherVersionName(edition.Name),
				edition.CreatedAt,
			)
		}

		result = append(result, domainEdition)
	}

	return result, nil
}

func (e *Edition) GetEdition(ctx context.Context, editionID values.LauncherVersionID, lockType repository.LockType) (*domain.LauncherVersion, error) {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = e.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var edition migrate.EditionTable2
	err = db.
		Where("id = ?", uuid.UUID(editionID)).
		Take(&edition).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition: %w", err)
	}

	var domainEdition *domain.LauncherVersion
	if edition.QuestionnaireURL.Valid {
		questionnaireURL, err := url.Parse(edition.QuestionnaireURL.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse questionnaire url: %w", err)
		}

		domainEdition = domain.NewLauncherVersionWithQuestionnaire(
			values.NewLauncherVersionIDFromUUID(edition.ID),
			values.NewLauncherVersionName(edition.Name),
			values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
			edition.CreatedAt,
		)
	} else {
		domainEdition = domain.NewLauncherVersionWithoutQuestionnaire(
			values.NewLauncherVersionIDFromUUID(edition.ID),
			values.NewLauncherVersionName(edition.Name),
			edition.CreatedAt,
		)
	}

	return domainEdition, nil
}

func (e *Edition) UpdateEditionGameVersions(
	ctx context.Context,
	editionID values.LauncherVersionID,
	gameVersionIDs []values.GameVersionID,
) error {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	gameVersions := make([]migrate.GameVersionTable2, 0, len(gameVersionIDs))
	for _, gameVersionID := range gameVersionIDs {
		gameVersions = append(gameVersions, migrate.GameVersionTable2{
			ID: uuid.UUID(gameVersionID),
		})
	}

	err = db.
		Model(&migrate.EditionTable2{
			ID: uuid.UUID(editionID),
		}).
		Association("GameVersions").
		Replace(gameVersions)
	if err != nil {
		return fmt.Errorf("failed to update edition game versions: %w", err)
	}

	return nil
}

func (e *Edition) GetEditionGameVersions(ctx context.Context, editionID values.LauncherVersionID, lockType repository.LockType) ([]*repository.GameVersionInfoWithGameID, error) {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = e.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var gameVersions []*migrate.GameVersionTable2
	err = db.
		Session(&gorm.Session{}).
		Model(&migrate.EditionTable2{
			ID: uuid.UUID(editionID),
		}).
		Preload("GameFiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id")
		}).
		Joins("JOIN games ON games.id = v2_game_versions.game_id AND games.deleted_at IS NULL").
		Association("GameVersions").
		Find(&gameVersions)
	if err != nil {
		return nil, fmt.Errorf("failed to get edition game versions: %w", err)
	}

	var result []*repository.GameVersionInfoWithGameID
	for _, gameVersion := range gameVersions {
		var optionURL types.Option[values.GameURLLink]
		if len(gameVersion.URL) != 0 {
			url, err := url.Parse(gameVersion.URL)
			if err != nil {
				return nil, fmt.Errorf("failed to parse game version url: %w", err)
			}

			optionURL = types.NewOption(values.NewGameURLLink(url))
		}

		fileIDs := make([]values.GameFileID, 0, len(gameVersion.GameFiles))
		for _, file := range gameVersion.GameFiles {
			fileIDs = append(fileIDs, values.NewGameFileIDFromUUID(file.ID))
		}

		result = append(result, &repository.GameVersionInfoWithGameID{
			GameVersion: domain.NewGameVersion(
				values.NewGameVersionIDFromUUID(gameVersion.ID),
				values.NewGameVersionName(gameVersion.Name),
				values.NewGameVersionDescription(gameVersion.Description),
				gameVersion.CreatedAt,
			),
			GameID:  values.NewGameIDFromUUID(gameVersion.GameID),
			ImageID: values.GameImageIDFromUUID(gameVersion.GameImageID),
			VideoID: values.NewGameVideoIDFromUUID(gameVersion.GameVideoID),
			URL:     optionURL,
			FileIDs: fileIDs,
		})
	}

	return result, nil
}

func (e *Edition) GetEditionGameVersionByGameID(ctx context.Context, editionID values.LauncherVersionID, gameID values.GameID, lockType repository.LockType) (*domain.GameVersion, error) {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = e.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var gameVersion migrate.GameVersionTable2
	err = db.
		Joins("INNER JOIN edition_game_version_relations ON edition_game_version_relations.game_version_id = v2_game_versions.id").
		Where("v2_game_versions.game_id = ?", uuid.UUID(gameID)).
		Take(&gameVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition game version: %w", err)
	}

	return domain.NewGameVersion(
		values.NewGameVersionIDFromUUID(gameVersion.ID),
		values.NewGameVersionName(gameVersion.Name),
		values.NewGameVersionDescription(gameVersion.Description),
		gameVersion.CreatedAt,
	), nil
}

func (e *Edition) GetEditionGameVersionByImageID(ctx context.Context, editionID values.LauncherVersionID, imageID values.GameImageID, lockType repository.LockType) (*domain.GameVersion, error) {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = e.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var gameVersion migrate.GameVersionTable2
	err = db.
		Where("game_image_id = ?", uuid.UUID(imageID)).
		Take(&gameVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition game version: %w", err)
	}

	return domain.NewGameVersion(
		values.NewGameVersionIDFromUUID(gameVersion.ID),
		values.NewGameVersionName(gameVersion.Name),
		values.NewGameVersionDescription(gameVersion.Description),
		gameVersion.CreatedAt,
	), nil
}

func (e *Edition) GetEditionGameVersionByVideoID(ctx context.Context, editionID values.LauncherVersionID, videoID values.GameVideoID, lockType repository.LockType) (*domain.GameVersion, error) {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = e.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var gameVersion migrate.GameVersionTable2
	err = db.
		Where("game_video_id = ?", uuid.UUID(videoID)).
		Take(&gameVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition game version: %w", err)
	}

	return domain.NewGameVersion(
		values.NewGameVersionIDFromUUID(gameVersion.ID),
		values.NewGameVersionName(gameVersion.Name),
		values.NewGameVersionDescription(gameVersion.Description),
		gameVersion.CreatedAt,
	), nil
}

func (e *Edition) GetEditionGameVersionByFileID(ctx context.Context, editionID values.LauncherVersionID, fileID values.GameFileID, lockType repository.LockType) (*domain.GameVersion, error) {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = e.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var gameVersion migrate.GameVersionTable2
	err = db.
		Joins("INNER JOIN game_version_game_file_relations ON game_version_game_file_relations.game_version_id = v2_game_versions.id").
		Where("game_version_game_file_relations.game_file_id = ?", uuid.UUID(fileID)).
		Take(&gameVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition game version: %w", err)
	}

	return domain.NewGameVersion(
		values.NewGameVersionIDFromUUID(gameVersion.ID),
		values.NewGameVersionName(gameVersion.Name),
		values.NewGameVersionDescription(gameVersion.Description),
		gameVersion.CreatedAt,
	), nil
}
