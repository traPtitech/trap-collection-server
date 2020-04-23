package model

// GameVersionRelation ランチャーのバージョンに入るゲームの構造体
type GameVersionRelation struct {
	LauncherVersionID uint `gorm:"type:int(11);NOT NULL;PRIMARY_KEY;AUTO_INCREMENT;"`
	LauncherVersion   LauncherVersion
	GameID            string `gorm:"type:varchar(36);NOT NULL;PRIMARY_KEY;"`
	Game              Game
}
