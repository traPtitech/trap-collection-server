package model

//PostSeat Postされる席の構造体
type PostSeat struct {
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Status string `json:"status,omitempty"`
}

//GetSeat Getされる席の構造体
type GetSeat struct {
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}
