package swift

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

func TestImageKey(t *testing.T) {
	t.Parallel()

	// clientは使わないのでnilでOK
	gameImageStorage := NewGameImage(nil)

	loopNum := 100

	for i := 0; i < loopNum; i++ {
		imageID := values.NewGameImageID()

		image := domain.NewGameImage(
			imageID,
			values.GameImageType(rand.Intn(3)),
		)

		key := gameImageStorage.imageKey(image)

		assert.Equal(t, fmt.Sprintf("images/%s", uuid.UUID(image.GetID()).String()), key)
	}
}
