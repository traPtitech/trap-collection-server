/*
 * traPCollection API
 *
 * traPCollectionのAPI
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// NewGame - 新しいゲームの名前
type NewGame struct {

	// 修正後のゲームの名前
	Name string `json:"name"`

	// 修正後のゲームの説明文
	Description string `json:"description"`
}
