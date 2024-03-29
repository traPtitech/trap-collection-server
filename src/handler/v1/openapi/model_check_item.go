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

// CheckItem - ゲームの更新・破損のチェック用リスト
type CheckItem struct {
	Id string `json:"id"`

	Md5 string `json:"md5"`

	// ゲームの種類（url,jar,windows,mac）
	Type string `json:"type"`

	// 実行ファイルの相対パス
	EntryPoint string `json:"entryPoint,omitempty"`

	// ゲーム本体の更新日時
	BodyUpdatedAt time.Time `json:"bodyUpdatedAt"`

	// 画像の更新日時
	ImgUpdatedAt time.Time `json:"imgUpdatedAt"`

	// 動画の更新日時
	MovieUpdatedAt *time.Time `json:"movieUpdatedAt,omitempty"`
}
