/*
 * traPCollection API
 *
 * traPCollectionのAPI
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"time"
)

// GameVersion - ゲームのバージョン
type GameVersion struct {

	// ID
	Id string `json:"id"`

	// 名前
	Name string `json:"name"`

	// バージョンの説明
	Description string `json:"description"`

	// 登録時刻
	CreatedAt time.Time `json:"createdAt"`
}