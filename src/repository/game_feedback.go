package repository

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFeedback interface {
	GetFeedbackConfig(ctx context.Context, gameID values.GameID, lockType LockType) (bool, error)
}
