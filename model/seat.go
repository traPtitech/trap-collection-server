package model

//PostSeat Postされる席の構造体
type PostSeat struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
}
