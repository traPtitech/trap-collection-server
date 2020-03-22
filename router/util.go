package router

import (
	"errors"
	"os"
)

var clientID string

// InitRouter router内で使う環境変数の初期化
func InitRouter() error {
	clientID = os.Getenv("CLIENT_ID")
	if len(clientID) == 0 {
		return errors.New("ENV CLIENT_ID IS NULL")
	}
	return nil
}
