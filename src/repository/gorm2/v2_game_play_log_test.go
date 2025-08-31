package gorm2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
)

func TestCreateGamePlayLog(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description string
		playLog     *domain.GamePlayLog
		isErr       bool
		err         error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description: "TODO: add test case",
			playLog:     nil,
			isErr:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := gamePlayLogRepository.CreateGamePlayLog(ctx, testCase.playLog)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGamePlayLog(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description  string
		playLogID    values.GamePlayLogID
		expectedLog  *domain.GamePlayLog
		isErr        bool
		err          error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description:  "TODO: add test case",
			playLogID:    values.NewGamePlayLogID(),
			expectedLog:  nil,
			isErr:        true,
			err:          repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			log, err := gamePlayLogRepository.GetGamePlayLog(ctx, testCase.playLogID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedLog, log)
			}
		})
	}
}

func TestUpdateGamePlayLogEndTime(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description string
		playLogID   values.GamePlayLogID
		endTime     time.Time
		isErr       bool
		err         error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description: "TODO: add test case",
			playLogID:   values.NewGamePlayLogID(),
			endTime:     time.Now(),
			isErr:       true,
			err:         repository.ErrNoRecordUpdated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := gamePlayLogRepository.UpdateGamePlayLogEndTime(ctx, testCase.playLogID, testCase.endTime)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGamePlayStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description     string
		gameID          values.GameID
		gameVersionID   *values.GameVersionID
		start           time.Time
		end             time.Time
		expectedStats   *domain.GamePlayStats
		isErr           bool
		err             error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description:     "TODO: add test case",
			gameID:          values.NewGameID(),
			gameVersionID:   nil,
			start:           time.Now().Add(-24 * time.Hour),
			end:             time.Now(),
			expectedStats:   nil,
			isErr:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			stats, err := gamePlayLogRepository.GetGamePlayStats(ctx, testCase.gameID, testCase.gameVersionID, testCase.start, testCase.end)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedStats, stats)
			}
		})
	}
}

func TestGetEditionPlayStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description   string
		editionID     values.LauncherVersionID
		start         time.Time
		end           time.Time
		expectedStats *domain.EditionPlayStats
		isErr         bool
		err           error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description:   "TODO: add test case",
			editionID:     values.NewLauncherVersionID(),
			start:         time.Now().Add(-24 * time.Hour),
			end:           time.Now(),
			expectedStats: nil,
			isErr:         false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			stats, err := gamePlayLogRepository.GetEditionPlayStats(ctx, testCase.editionID, testCase.start, testCase.end)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedStats, stats)
			}
		})
	}
}
