func (g *GamePlayLogV2) GetGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) (*domain.GamePlayLog, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, err
	}

	var gamePlayLog schema.GamePlayLogTable //migrateではなくschemaに定義されている構造体を使う
	err = db.
		Where("id = ?", uuid.UUID(playLogID)). //playLogIDに合致したレコードを取得
		First(&gamePlayLog).Error              //1件を取得
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRecordNotFound
		}
		return nil, err
	}

	var endTime *time.Time // endTimeはNULL許容なのでポインタで扱う
	if gamePlayLog.EndTime.Valid {
		endTime = &gamePlayLog.EndTime.Time
	}

	return domain.NewGamePlayLog(
		values.GamePlayLogID(gamePlayLog.ID),
		values.LauncherVersionID(gamePlayLog.EditionID),
		values.GameID(gamePlayLog.GameID),
		values.GameVersionID(gamePlayLog.GameVersionID),
		gamePlayLog.StartTime,
		endTime,
		gamePlayLog.CreatedAt,
		gamePlayLog.UpdatedAt,
	), nil
}