package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

type GameGenre struct {
	db *DB
}

func NewGameGenre(db *DB) *GameGenre {
	return &GameGenre{
		db: db,
	}
}

var _ repository.GameGenre = &GameGenre{}

func (gameGenre *GameGenre) GetGenresByGameID(ctx context.Context, gameID values.GameID) ([]*domain.GameGenre, error) {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var genres []*migrate.GameGenreTable
	err = db.
		Model(&migrate.GameGenreTable{}).
		Joins("JOIN game_genre_relations ON game_genres.id = game_genre_relations.genre_id").
		Where("game_genre_relations.game_id = ?", uuid.UUID(gameID)).
		Order("`created_at` DESC").
		Find(&genres).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*domain.GameGenre{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get GameGenre: %w", err)
	}

	result := make([]*domain.GameGenre, 0, len(genres))

	for _, genre := range genres {
		result = append(result, domain.NewGameGenre(values.GameGenreID(genre.ID), values.GameGenreName(genre.Name), genre.CreatedAt))
	}
	return result, nil
}

func (gameGenre *GameGenre) RemoveGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := db.
		Select("Games").
		Delete(&migrate.GameGenreTable{ID: uuid.UUID(gameGenreID)})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to remove game genre: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}

func (gameGenre *GameGenre) GetGameGenresWithNames(ctx context.Context, gameGenreNames []values.GameGenreName) ([]*domain.GameGenre, error) {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var genres []migrate.GameGenreTable
	result := db.
		Where("name IN ?", gameGenreNames).
		Find(&genres)
	if result.RowsAffected == 0 {
		return nil, repository.ErrRecordNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get game genres with names: %w", err)
	}

	resultGenres := make([]*domain.GameGenre, 0, len(genres))
	for _, genre := range genres {
		resultGenres = append(resultGenres, domain.NewGameGenre(values.GameGenreID(genre.ID), values.GameGenreName(genre.Name), genre.CreatedAt))
	}

	return resultGenres, nil
}

// // SaveGameGenres
// // ゲームジャンルを作成する。
// // 名前が重複するゲームジャンルが1つでも存在するとき、ErrDuplicatedUniqueKeyを返す。
func (gameGenre *GameGenre) SaveGameGenres(ctx context.Context, gameGenres []*domain.GameGenre) error {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	genres := make([]migrate.GameGenreTable, 0, len(gameGenres))
	for i := range gameGenres {
		genres = append(genres, migrate.GameGenreTable{
			ID:        uuid.UUID(gameGenres[i].GetID()),
			Name:      string(gameGenres[i].GetName()),
			CreatedAt: gameGenres[i].GetCreatedAt(),
		})
	}

	err = db.Create(&genres).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == 1062 {
			return repository.ErrDuplicatedUniqueKey
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// RegisterGenresToGame
// ゲームにゲームジャンルを登録する。
func (gameGenre *GameGenre) RegisterGenresToGame(ctx context.Context, gameID values.GameID, gameGenreIDs []values.GameGenreID) error {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var game migrate.GameTable2

	if err = db.First(&game, uuid.UUID(gameID)).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return repository.ErrRecordNotFound
	} else if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	gameGenreUUIDs := make([]uuid.UUID, 0, len(gameGenreIDs))
	for _, genre := range gameGenreIDs {
		gameGenreUUIDs = append(gameGenreUUIDs, uuid.UUID(genre))
	}

	var gameGenres []migrate.GameGenreTable
	err = db.Where("`id` IN ?", gameGenreUUIDs).Find(&gameGenres).Error
	if err != nil {
		return fmt.Errorf("failed to get game genres: %w", err)
	}

	if len(gameGenres) != len(gameGenreIDs) {
		return repository.ErrIncludeInvalidArgs
	}

	err = db.Model(&game).Association("GameGenres").Replace(gameGenres)
	if err != nil {
		return fmt.Errorf("failed to register genres to game: %w", err)
	}

	return nil
}

type gameGenreInfo struct {
	migrate.GameGenreTable
	Num int
}

func (gameGenre *GameGenre) GetGameGenres(ctx context.Context, visibilities []values.GameVisibility) ([]*repository.GameGenreInfo, error) {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	visibilityTypeNames := make([]string, 0, len(visibilities))
	for _, visibility := range visibilities {
		var visibilityTypeName string
		switch visibility {
		case values.GameVisibilityTypePublic:
			visibilityTypeName = migrate.GameVisibilityTypePublic
		case values.GameVisibilityTypeLimited:
			visibilityTypeName = migrate.GameVisibilityTypeLimited
		case values.GameVisibilityTypePrivate:
			visibilityTypeName = migrate.GameVisibilityTypePrivate
		default:
			return nil, fmt.Errorf("invalid game visibility: %v", visibility)
		}
		visibilityTypeNames = append(visibilityTypeNames, visibilityTypeName)
	}

	// ジャンルごとのゲーム数を数え、ジャンルと一緒に返す。
	query := "SELECT COUNT(ggr.game_id) AS Num, game_genres.* FROM game_genres " +
		"LEFT JOIN game_genre_relations AS ggr ON game_genres.id = ggr.genre_id " +
		"LEFT JOIN games AS g ON g.id = ggr.game_id " +
		"LEFT JOIN game_visibility_types AS gvt ON g.visibility_type_id = gvt.id " +
		"WHERE gvt.name IN (?) " +
		"GROUP BY ggr.genre_id " +
		"ORDER BY game_genres.created_at DESC"

	var gameGenreInfos []gameGenreInfo
	err = db.Raw(query, visibilityTypeNames).Scan(&gameGenreInfos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game genres: %w", err)
	}

	result := make([]*repository.GameGenreInfo, 0, len(gameGenreInfos))

	for i := range gameGenreInfos {
		result = append(result, &repository.GameGenreInfo{
			GameGenre: *domain.NewGameGenre(
				values.GameGenreID(gameGenreInfos[i].ID),
				values.GameGenreName(gameGenreInfos[i].Name),
				gameGenreInfos[i].CreatedAt,
			),
			Num: gameGenreInfos[i].Num,
		})
	}

	return result, nil
}

func (gameGenre *GameGenre) UpdateGameGenre(ctx context.Context, genre *domain.GameGenre) error {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	newGenre := migrate.GameGenreTable{
		Name: string(genre.GetName()),
	}

	result := db.Model(&migrate.GameGenreTable{
		ID: uuid.UUID(genre.GetID()),
	}).Updates(&newGenre)
	var mysqlErr *mysql.MySQLError
	if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
		return repository.ErrDuplicatedUniqueKey
	}
	if result.Error != nil {
		return fmt.Errorf("failed to update game genre: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}

func (gameGenre *GameGenre) GetGameGenre(ctx context.Context, gameGenreID values.GameGenreID) (*domain.GameGenre, error) {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var genre migrate.GameGenreTable
	if err := db.First(&genre, uuid.UUID(gameGenreID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to get game genre: %w", err)
	}

	return domain.NewGameGenre(
		values.GameGenreID(genre.ID),
		values.GameGenreName(genre.Name),
		genre.CreatedAt), nil
}

func (gameGenre *GameGenre) GetGamesByGenreID(ctx context.Context, gameGenreID values.GameGenreID) ([]*domain.Game, error) {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var games []migrate.GameTable2
	err = db.
		Model(&migrate.GameTable2{}).
		Preload("GameVisibilityType").
		Joins("JOIN game_genre_relations ON games.id = game_genre_relations.game_id").
		Where("genre_id = ?", uuid.UUID(gameGenreID)).
		Find(&games).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get games by genre id: %w", err)
	}

	result := make([]*domain.Game, 0, len(games))
	for _, game := range games {
		var visibility values.GameVisibility
		switch game.GameVisibilityType.Name {
		case migrate.GameVisibilityTypePublic:
			visibility = values.GameVisibilityTypePublic
		case migrate.GameVisibilityTypeLimited:
			visibility = values.GameVisibilityTypeLimited
		case migrate.GameVisibilityTypePrivate:
			visibility = values.GameVisibilityTypePrivate
		default:
			return nil, fmt.Errorf("invalid game visibility: %v", game.GameVisibilityType.Name)
		}

		result = append(result, domain.NewGame(
			values.GameID(game.ID),
			values.GameName(game.Name),
			values.GameDescription(game.Description),
			visibility,
			game.CreatedAt,
		))
	}

	return result, nil
}
