/*
 * traPCollection API
 *
 * traPCollectionのAPI
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// GameFile - ゲームのファイルの情報
type GameFile struct {

	// アセットのID
	Id string `json:"id"`

	// ゲームの種類（jar,windows,mac）
	Type string `json:"type"`

	// ゲームの起動時に実行するファイル
	EntryPoint string `json:"entryPoint"`
}
