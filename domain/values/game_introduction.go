package values

import "github.com/google/uuid"

type (
	GameIntroductionID string
	GameImageExtension string
	GameVideoExtension string
)

const (
	GameImageExtensionJpg GameImageExtension = "jpg"
	GameImageExtensionPng GameImageExtension = "png"
	GameImageExtensionGif GameImageExtension = "gif"
)

const (
	GameVideoExtensionMp4 GameVideoExtension = "mp4"
)

func NewGameIntoductionID() GameIntroductionID {
	return GameIntroductionID(uuid.New().String())
}

func NewGameIntoductionIDFromString(id string) (GameIntroductionID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidFormat
	}

	return GameIntroductionID(id), nil
}
