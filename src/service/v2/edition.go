package v2

import "github.com/traPtitech/trap-collection-server/src/repository"

type Edition struct {
	db                    repository.DB
	editionRepository     repository.Edition
	gameRepository        repository.GameV2
	gameVersionRepository repository.GameVersionV2
}

func NewEdition(
	db repository.DB,
	editionRepository repository.Edition,
	gameRepository repository.GameV2,
	gameVersionRepository repository.GameVersionV2,
) *Edition {
	return &Edition{
		db:                    db,
		editionRepository:     editionRepository,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
	}
}
