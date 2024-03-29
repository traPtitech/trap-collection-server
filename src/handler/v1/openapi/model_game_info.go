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

// GameInfo - ゲーム名とID
type GameInfo struct {

	// 追加されたゲームのUUID
	Id string `json:"id"`

	// 追加されたゲームの名前
	Name string `json:"name"`

	// 追加されたゲームの説明
	Description string `json:"description"`

	// ゲームの追加された時刻
	CreatedAt time.Time `json:"createdAt"`
}
