package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type GameFeedback interface {
	GetFeedbackConfig(ctx context.Context, gameID values.GameID) (bool, error)
}
