package model

import "time"

// GameVersionRelation ランチャーのバージョンに入るゲームの構造体
type GameVersionRelation struct {
	LauncherVersionID uint `gorm:"type:int(11);NOT NULL;PRIMARY_KEY;AUTO_INCREMENT;"`
	LauncherVersion   LauncherVersion
	GameID            string `gorm:"type:varchar(36);NOT NULL;PRIMARY_KEY;"`
	Game              Game
}

// Game gameの構造体
type Game struct {
	ID          string    `gorm:"type:varchar(36);PRIMARY_KEY;"`
	Name        string    `gorm:"type:varchar(32);NOT NULL;"`
	Description string    `gorm:"type:text;"`
	CreatedAt   time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameVersion gameのversionの構造体
type GameVersion struct {
	ID          uint      `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameID      string    `gorm:"type:varchar(36);NOT NULL;"`
	Game        Game      `gorm:"FOREIGNKEY:GameID"`
	Name        string    `gorm:"type:varchar(36);NOT NULL;"`
	Description string    `gorm:"type:text;"`
	CreatedAt   time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameAsset gameのassetの構造体
type GameAsset struct {
	ID            uint `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameVersionID uint `gorm:"type:int(11);NOT NULL;"`
	GameVersion   GameVersion
	Type          uint8  `gorm:"type:tinyint;NOT NULL;"`
	Md5           string `gorm:"type:binary(16);"`
	URL           string `gorm:"type:text"`
}

// GameIntroduction gameのintroductionの構造体
type GameIntroduction struct {
	ID        uint   `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameID    string `gorm:"type:varchar(36);NOT NULL;"`
	Game      Game
	Role      uint8     `gorm:"type:tinyint;NOT NULL;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
}

// Maintainer gameのmaintainerの構造体
type Maintainer struct {
	ID        uint   `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameID    string `gorm:"type:varchar(36);NOT NULL;"`
	Game      Game
	UserID    string    `gorm:"type:varchar(32);NOT NULL;"`
	Role      uint8     `gorm:"type:tinyint;NOT NULL;DEFAULT:0;"`
	MimeType  string    `gorm:"type:text;NOT NULL;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt time.Time `gorm:"type:datetime;DEFAULT:NULL;"`
}

// LauncherVersion ランチャーのバージョンの構造体
type LauncherVersion struct {
	ID                   uint                  `json:"id" gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;default:0;"`
	Name                 string                `json:"name,omitempty" gorm:"type:varchar(32);NOT NULL;UNIQUE;"`
	GameVersionRelations []GameVersionRelation `json:"games" gorm:"foreignkey:LauncherVersionID;"`
	Questions            []Question            `json:"questions" gorm:"foreignkey:LauncherVersionID;"`
	CreatedAt            time.Time             `json:"created_at,omitempty" gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
	DeletedAt            time.Time             `json:"deleted_at,omitempty" gorm:"type:datetime;default:NULL;"`
}

// ProductKey アクセストークンの構造体
type ProductKey struct {
	ID                uint   `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;default:0;"`
	Key               string `gorm:"type:varchar(29);NOT NULL;PRIMARY_KEY;default:\"\";"`
	LauncherVersionID uint   `gorm:"type:int(11) unsigned;"`
	LauncherVersion   LauncherVersion
}

// Player プレイヤーの履歴の構造体
type Player struct {
	ID        uint      `gorm:"type:int(11) unsigned;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT;"`
	ProductKeyID    uint      `gorm:"type:int(11) unsigned;not null;"`
	StartedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
	EndedAt   time.Time `gorm:"type:datetime;default:null;"`
}

// Question 質問の構造体
type Question struct {
	ID                uint             `gorm:"type:int(11) unsigned;auto_increament;primary_key;"`
	LauncherVersionID uint             `gorm:"type:int(11) unsigned;not null;"`
	QuestionNum       uint             `gorm:"type:int(11) unsigned;not null;"`
	Type              uint8            `gorm:"type:tinyint unsigned;not null;"`
	Content           string           `gorm:"type:text;not null;"`
	Required          bool             `gorm:"type:boolean;not null;default:true;"`
	QuestionOptions   []QuestionOption `gorm:"foreign_key:QuestionID;"`
	CreatedAt         time.Time        `gorm:"type:datetime;not null;default:current_timestamp;"`
	DeletedAt         time.Time        `gorm:"type:datetime;default:null;"`
}

// QuestionOption 選択肢の構造体
type QuestionOption struct {
	ID         uint `gorm:"type:int(11) unsigned;not null;primary_key;auto_increament;"`
	QuestionID uint `gorm:"type:int (11) unsigned;not null"`
	Question   Question
	Label      string `gorm:"type:text;not null;"`
}

// Response 回答の構造体
type Response struct {
	ID                string `gorm:"type:varchar(36);not null;primary_key;"`
	PlayerID          uint   `gorm:"type:int(11);not null;"`
	Player            Player
	Remark            string    `gorm:"type:text;"`
	CreatedAt         time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
}

// TextAnswer テキスト形式の回答の構造体
type TextAnswer struct {
	ID         uint   `gorm:"type:int(11) unsigned;not null;primary_key;auto_increment;"`
	ResponseID string `gorm:"type:varchar(36);not null;"`
	Response   Response
	QuestionID uint `gorm:"type:int(11) unsigned;not null;"`
	Question   Question
	Content    string `gorm:"type:text;not null;"`
}

// OptionAnswer 選択肢式の回答の構造体
type OptionAnswer struct {
	ID             uint   `gorm:"type:int(11) unsigned;not null;primary_key;auto_increment;"`
	ResponseID     string `gorm:"type:varchar(36);not null;"`
	Response       Response
	QuestionID     uint `gorm:"type:int(11) unsigned;not null;"`
	Question       Question
	OptionID       uint           `gorm:"type:int(11) unsigned;not null;"`
	QuestionOption QuestionOption `gorm:"foreign_key:OptionID;"`
}

// GameRating ゲームの評価の構造体
type GameRating struct {
	ID            uint `gorm:"type:int(11) unsigned;not null;primary_key;auto_increment;"`
	ResponseID    uint `gorm:"type:varchar(36);not null;"`
	Response      Response
	GameVersionID uint `gorm:"type:int(11) unsigned;not null;"`
	GameVersion   GameVersion
	Star          uint8 `gorm:"type:tinyint unsigned;not null;"`
	PlayTime      uint  `gorm:"type:int(11) unsigned;not null;"`
}