package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// LauncherVersion ランチャーのバージョンの構造体
type LauncherVersion struct {
	ID                   string                  `json:"id" gorm:"type:varchar(36);PRIMARY_KEY;"`
	Name                 string                `json:"name,omitempty" gorm:"type:varchar(32);NOT NULL;"`
	AnkeToURL            string                `json:"anke_to,omitempty" gorm:"column:anke_to_url;type:text;default:NULL;"`
	GameVersionRelations []GameVersionRelation `json:"games" gorm:"foreignkey:LauncherVersionID;"`
	CreatedAt            time.Time             `json:"created_at,omitempty" gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
	DeletedAt            time.Time             `json:"deleted_at,omitempty" gorm:"type:datetime;default:NULL;"`
}

// LauncherVersionMeta launcher_versionテーブルのリポジトリ
type LauncherVersionMeta interface {
	GetLauncherVersions() ([]*openapi.Version, error)
	GetLauncherVersionDetailsByID(id uint) (versionDetails *openapi.VersionDetails, err error)
	InsertLauncherVersion(name string, ankeToURL string) (*openapi.VersionMeta, error)
}

// GetLauncherVersions ランチャーのバージョン一覧取得
func (*DB) GetLauncherVersions() ([]*openapi.Version, error) {
	var launcherVersions []LauncherVersion
	err := db.Find(&launcherVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher versions: %w", err)
	}

	apiLauncherVersions := make([]*openapi.Version, 0, len(launcherVersions))
	for _, launcherVersion := range launcherVersions {
		apiLauncherVersion := openapi.Version{
			Id:        launcherVersion.ID,
			Name:      launcherVersion.Name,
			AnkeTo:    launcherVersion.AnkeToURL,
			CreatedAt: launcherVersion.CreatedAt,
		}
		apiLauncherVersions = append(apiLauncherVersions, &apiLauncherVersion)
	}

	return apiLauncherVersions, nil
}

// GetLauncherVersionDetailsByID ランチャーのバージョンをIDから取得
func (*DB) GetLauncherVersionDetailsByID(id string) (versionDetails *openapi.VersionDetails, err error) {
	versionDetails = &openapi.VersionDetails{
		Games: []openapi.GameMeta{},
	}

	rows, err := db.Table("launcher_versions").
		Joins("LEFT OUTER JOIN game_version_relations ON launcher_versions.id = game_version_relations.launcher_version_id").
		Joins("LEFT OUTER JOIN games ON game_version_relations.game_id <=> games.id").
		Where("launcher_versions.id = ?", id).
		Select("launcher_versions.id,launcher_versions.name,launcher_versions.anke_to,launcher_versions.created_at,games.id, games.name").
		Rows()
	if err != nil {
		return &openapi.VersionDetails{}, fmt.Errorf("Failed In Getting Launcher Versions:%w", err)
	}
	for rows.Next() {
		var gameID sql.NullString
		var gameName sql.NullString
		err = rows.Scan(&versionDetails.Id, &versionDetails.Name, &versionDetails.AnkeTo, &versionDetails.CreatedAt, &gameID, &gameName)
		if err != nil {
			return &openapi.VersionDetails{}, fmt.Errorf("Failed In Scaning Launcher Version:%w", err)
		}
		if gameID.Valid {
			versionDetails.Games = append(versionDetails.Games, openapi.GameMeta{
				Id:   gameID.String,
				Name: gameName.String,
			})
		}
	}

	return
}

// InsertLauncherVersion ランチャーのバージョンの追加
func (*DB) InsertLauncherVersion(name string, ankeToURL string) (*openapi.VersionMeta, error) {
	var apiVersion openapi.VersionMeta
	err := db.Transaction(func(tx *gorm.DB) error {
		launcherVersion := LauncherVersion{
			Name:      name,
			AnkeToURL: ankeToURL,
		}

		err := tx.Create(&launcherVersion).Error
		if err != nil {
			return fmt.Errorf("failed to insert a lancher version record: %w", err)
		}

		err = tx.Last(&launcherVersion).Error
		if err != nil {
			return fmt.Errorf("failed to get the last launcher version record: %w", err)
		}
		apiVersion = openapi.VersionMeta{
			Id:        int32(launcherVersion.ID),
			Name:      launcherVersion.Name,
			AnkeTo:    launcherVersion.AnkeToURL,
			CreatedAt: launcherVersion.CreatedAt,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return &apiVersion, nil
}
