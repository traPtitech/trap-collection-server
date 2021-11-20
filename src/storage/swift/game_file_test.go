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

func TestFileKey(t *testing.T) {
	t.Parallel()

	// clientは使わないのでnilでOK
	gameFileStorage := NewGameFile(nil)

	loopNum := 100

	for i := 0; i < loopNum; i++ {
		fileID := values.NewGameFileID()

		file := domain.NewGameFile(
			fileID,
			values.GameFileType(rand.Intn(3)),
			"path/to/file",
			[]byte("hash"),
		)

		key := gameFileStorage.fileKey(file)

		assert.Equal(t, fmt.Sprintf("files/%s", uuid.UUID(file.GetID()).String()), key)
	}
}
