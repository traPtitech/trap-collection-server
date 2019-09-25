package model

//AdminList 管理者一覧の構造体
type AdminList struct {
	Admin string `json:"id,omitempty" db:"user_traqid"`
}
