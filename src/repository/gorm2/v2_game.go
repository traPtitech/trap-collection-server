package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

type GameV2 struct {
	db *DB
}

func NewGameV2(db *DB) *GameV2 {
	return &GameV2{
		db: db,
	}
}

func (g *GameV2) SaveGame(ctx context.Context, game *domain.Game) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var visibilityTypeName string
	switch game.GetVisibility() {
	case values.GameVisibilityTypePublic:
		visibilityTypeName = migrate.GameVisibilityTypePublic
	case values.GameVisibilityTypeLimited:
		visibilityTypeName = migrate.GameVisibilityTypeLimited
	case values.GameVisibilityTypePrivate:
		visibilityTypeName = migrate.GameVisibilityTypePrivate
	default:
		return fmt.Errorf("invalid visibility type: %d", game.GetVisibility())
	}

	var visibilityType schema.GameVisibilityTypeTable
	err = db.
		Where("name = ?", visibilityTypeName).
		Select("id").Take(&visibilityType).Error
	if err != nil {
		return fmt.Errorf("failed to get visibility type: %w", err)
	}
	visibilityTypeID := visibilityType.ID

	gameTable := schema.GameTable2{
		ID:               uuid.UUID(game.GetID()),
		Name:             string(game.GetName()),
		Description:      string(game.GetDescription()),
		CreatedAt:        game.GetCreatedAt(),
		VisibilityTypeID: visibilityTypeID,
	}

	err = db.Create(&gameTable).Error
	if err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	return nil
}

func (g *GameV2) UpdateGame(ctx context.Context, game *domain.Game) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	visibility, err := g.getVisibility(ctx, game.GetVisibility())
	if err != nil {
		return fmt.Errorf("failed to get visibility: %w", err)
	}

	gameTable := schema.GameTable2{
		Name:             string(game.GetName()),
		Description:      string(game.GetDescription()),
		VisibilityTypeID: visibility.ID,
	}

	result := db.
		Where("id = ?", uuid.UUID(game.GetID())).
		Updates(gameTable)
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}

func (g *GameV2) RemoveGame(ctx context.Context, gameID values.GameID) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := db.
		Where("id = ?", uuid.UUID(gameID)).
		Delete(&schema.GameTable2{})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to remove game: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}

func (g *GameV2) GetGame(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.Game, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock type: %w", err)
	}

	var game schema.GameTable2
	err = db.
		Joins("GameVisibilityType").
		Where("games.id = ?", uuid.UUID(gameID)).
		Take(&game).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	var visibility values.GameVisibility
	switch game.GameVisibilityType.Name {
	case migrate.GameVisibilityTypePublic:
		visibility = values.GameVisibilityTypePublic
	case migrate.GameVisibilityTypeLimited:
		visibility = values.GameVisibilityTypeLimited
	case migrate.GameVisibilityTypePrivate:
		visibility = values.GameVisibilityTypePrivate
	}

	return domain.NewGame(
		values.NewGameIDFromUUID(game.ID),
		values.NewGameName(game.Name),
		values.NewGameDescription(game.Description),
		visibility,
		game.CreatedAt,
	), nil
}

func (g *GameV2) GetGames(
	ctx context.Context, limit int, offset int, sort repository.GamesSortType,
	visibilities []values.GameVisibility, userID *values.TraPMemberID, gameGenreIDs []values.GameGenreID, name string) ([]*domain.GameWithGenres, int, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get db: %w", err)
	}

	if limit < 0 {
		return nil, 0, repository.ErrNegativeLimit
	}
	if limit == 0 && offset != 0 {
		return nil, 0, errors.New("bad limit and offset")
	}

	var orderBy string
	switch sort {
	case repository.GamesSortTypeCreatedAt:
		orderBy = "games.created_at DESC"
	case repository.GamesSortTypeLatestVersion:
		orderBy = "games.latest_version_updated_at DESC"
	default:
		return nil, 0, fmt.Errorf("invalid sort type: %v", sort)
	}

	// visibilityの指定が無い時は全てのvisibilityを取得する
	if len(visibilities) == 0 {
		visibilities = []values.GameVisibility{
			values.GameVisibilityTypePublic,
			values.GameVisibilityTypeLimited,
			values.GameVisibilityTypePrivate,
		}
	}

	visibilityNames := make([]string, len(visibilities))
	for i := range visibilities {
		switch visibilities[i] {
		case values.GameVisibilityTypePublic:
			visibilityNames[i] = migrate.GameVisibilityTypePublic
		case values.GameVisibilityTypeLimited:
			visibilityNames[i] = migrate.GameVisibilityTypeLimited
		case values.GameVisibilityTypePrivate:
			visibilityNames[i] = migrate.GameVisibilityTypePrivate
		default:
			return nil, 0, fmt.Errorf("invalid game visibility args: %v", visibilities[i])
		}
	}

	var games []schema.GameTable2

	tx := db.
		Model(&schema.GameTable2{}).
		Preload("GameGenres").
		Preload("GameVisibilityType").
		Joins("JOIN game_visibility_types ON game_visibility_types.id = games.visibility_type_id").
		Where("game_visibility_types.name IN ?", visibilityNames)

	if userID != nil {
		tx = tx.
			Joins("JOIN game_management_roles ON game_management_roles.game_id = games.id").
			Where("game_management_roles.user_id = ?", uuid.UUID(*userID))
	}

	if len(gameGenreIDs) > 0 {
		gameGenreUUIDs := make([]uuid.UUID, 0, len(gameGenreIDs))
		for _, gameGenreID := range gameGenreIDs {
			gameGenreUUIDs = append(gameGenreUUIDs, uuid.UUID(gameGenreID))
		}

		// サブクエリをJOINする。
		// 指定されたゲームジャンル全てを持っている必要があるので、
		// INで少なくとも一つは指定されたジャンルを持つものに絞ったうえで、COUNTでジャンルの数を数えて、
		// その数が指定されたジャンルの数と一致するものを取得する。
		//
		// SELECT game_id FROM game_genre_relations WHERE genre_id IN (ジャンルのid全部)
		// GROUP BY game_id HAVING COUNT(DISTINCT genre_id) = ジャンルの数
		subQuery := db.
			Table("game_genre_relations").
			Where("genre_id IN ?", gameGenreUUIDs).
			Group("game_id").
			Having("COUNT(DISTINCT genre_id) = ?", len(gameGenreUUIDs)).
			Select("game_id")

		tx = tx.
			Joins("JOIN (?) AS sub ON sub.game_id = games.id", subQuery)
	}

	if name != "" {
		tx = tx.Where("games.name LIKE ?", "%"+name+"%")
	}

	txSelect := tx.
		Order(orderBy)

	if limit > 0 {
		txSelect = txSelect.Session(&gorm.Session{}).Limit(limit).Offset(offset)
	}
	err = txSelect.Find(&games).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games: %w", err)
	}

	gamesDomain := make([]*domain.GameWithGenres, 0, len(games))
	for i := range games {
		var visibility values.GameVisibility
		switch games[i].GameVisibilityType.Name {
		case migrate.GameVisibilityTypePublic:
			visibility = values.GameVisibilityTypePublic
		case migrate.GameVisibilityTypeLimited:
			visibility = values.GameVisibilityTypeLimited
		case migrate.GameVisibilityTypePrivate:
			visibility = values.GameVisibilityTypePrivate
		default:
			return nil, 0, fmt.Errorf("invalid game visibility: '%s'", games[i].GameVisibilityType.Name)
		}

		var gameGenresDomain []*domain.GameGenre
		for j := range games[i].GameGenres {
			gameGenresDomain = append(gameGenresDomain, domain.NewGameGenre(
				values.GameGenreIDFromUUID(games[i].GameGenres[j].ID),
				values.NewGameGenreName(games[i].GameGenres[j].Name),
				games[i].GameGenres[j].CreatedAt,
			))
		}

		gamesDomain = append(gamesDomain, domain.NewGameWithGenres(
			domain.NewGame(values.GameID(games[i].ID), values.GameName(games[i].Name), values.GameDescription(games[i].Description), visibility, games[i].CreatedAt),
			gameGenresDomain,
		))
	}

	var gamesNumber int64
	err = tx.Count(&gamesNumber).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games number: %w", err)
	}

	return gamesDomain, int(gamesNumber), nil
}

func (g *GameV2) GetGamesByIDs(ctx context.Context, gameIDs []values.GameID, lockType repository.LockType) ([]*domain.Game, error) {
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

	var games []schema.GameTable2
	err = db.
		Model(&schema.GameTable2{}).
		Preload("GameVisibilityType").
		Where("id IN ?", uuidGameIDs).
		Find(&games).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}

	gamesDomains := make([]*domain.Game, 0, len(games))
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
			return nil, fmt.Errorf("invalid game visibility: %s", game.GameVisibilityType.Name)
		}

		gamesDomains = append(gamesDomains, domain.NewGame(
			values.NewGameIDFromUUID(game.ID),
			values.NewGameName(game.Name),
			values.NewGameDescription(game.Description),
			visibility,
			game.CreatedAt,
		))
	}

	return gamesDomains, nil
}

func (g *GameV2) getVisibility(ctx context.Context, visibility values.GameVisibility) (*schema.GameVisibilityTypeTable, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	visibilityTypeName, err := convertVisibilityType(visibility)
	if err != nil {
		return nil, fmt.Errorf("failed to convert visibility type: %w", err)
	}

	var visibilityType schema.GameVisibilityTypeTable
	err = db.
		Where("name = ?", visibilityTypeName).
		Take(&visibilityType).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get visibility type: %w", err)
	}

	return &visibilityType, nil
}
