package gorm2

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestGetGameCreatorsByGameID(t *testing.T) {
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)

	now := time.Now()

	gameID1 := values.NewGameID()
	game1 := &schema.GameTable2{
		ID:               uuid.UUID(gameID1),
		Name:             "Test Game 1",
		VisibilityTypeID: 1,
	}
	gameID2 := values.NewGameID()
	game2 := &schema.GameTable2{
		ID:               uuid.UUID(gameID2),
		Name:             "Test Game 2",
		VisibilityTypeID: 1,
	}
	err = db.Create([]*schema.GameTable2{game1, game2}).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error
		require.NoError(t, err)
	})

	job1 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "job1",
	}
	job2 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "job2",
	}
	err = db.Create([]*schema.GameCreatorJobTable{job1, job2}).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameCreatorJobTable{}).Error
		require.NoError(t, err)
	})

	customJob1 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "customJob1",
		GameID:      uuid.UUID(gameID1),
	}
	customJob2 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "customJob2",
		GameID:      uuid.UUID(gameID2),
	}
	customJob3 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "customJob3",
		GameID:      uuid.UUID(gameID2),
	}
	err = db.Create([]*schema.GameCreatorCustomJobTable{customJob1, customJob2, customJob3}).Error
	require.NoError(t, err)
	t.Cleanup(func() {
		db, err := testDB.getDB(context.Background())
		require.NoError(t, err)
		err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameCreatorCustomJobTable{}).Error
		require.NoError(t, err)
	})

	creator1 := &schema.GameCreatorTable{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		UserName: "creator1",
		GameID:   uuid.UUID(gameID1),
		CreatorJobs: []schema.GameCreatorJobTable{
			*job1,
		},
		CustomCreatorJobs: []schema.GameCreatorCustomJobTable{
			*customJob1,
		},
		CreatedAt: now,
	}
	creator2 := &schema.GameCreatorTable{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		UserName: "creator2",
		GameID:   uuid.UUID(gameID2),
		CreatorJobs: []schema.GameCreatorJobTable{
			*job2,
		},
		CustomCreatorJobs: []schema.GameCreatorCustomJobTable{
			*customJob2,
			*customJob3,
		},
		CreatedAt: now,
	}
	creator3 := &schema.GameCreatorTable{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		UserName:  "creator3",
		GameID:    uuid.UUID(gameID2),
		CreatedAt: now.Add(1 * time.Hour),
	}

	testCases := map[string]struct {
		gameID      values.GameID
		allCreators []*schema.GameCreatorTable
		expected    []*domain.GameCreatorWithJobs
		err         error
	}{
		"一人を正しく取得できる": {
			gameID:      gameID1,
			allCreators: []*schema.GameCreatorTable{creator1, creator2},
			expected: []*domain.GameCreatorWithJobs{
				domain.NewGameCreatorWithJobs(
					domain.NewGameCreator(
						values.GameCreatorID(creator1.ID),
						values.TraPMemberID(creator1.UserID),
						values.TraPMemberName(creator1.UserName),
						creator1.CreatedAt),
					[]*domain.GameCreatorJob{
						domain.NewGameCreatorJob(
							values.GameCreatorJobID(job1.ID),
							values.GameCreatorJobDisplayName(job1.DisplayName),
							job1.CreatedAt,
						),
					},
					[]*domain.GameCreatorCustomJob{
						domain.NewGameCreatorCustomJob(
							values.GameCreatorJobID(customJob1.ID),
							values.GameCreatorJobDisplayName(customJob1.DisplayName),
							values.GameID(customJob1.GameID),
							customJob1.CreatedAt,
						),
					},
				),
			},
		},
		"複数人を正しく取得できる": {
			gameID:      gameID2,
			allCreators: []*schema.GameCreatorTable{creator1, creator2, creator3},
			expected: []*domain.GameCreatorWithJobs{
				domain.NewGameCreatorWithJobs(
					domain.NewGameCreator(
						values.GameCreatorID(creator2.ID),
						values.TraPMemberID(creator2.UserID),
						values.TraPMemberName(creator2.UserName),
						creator2.CreatedAt),
					[]*domain.GameCreatorJob{
						domain.NewGameCreatorJob(
							values.GameCreatorJobID(job2.ID),
							values.GameCreatorJobDisplayName(job2.DisplayName),
							job2.CreatedAt,
						),
					},
					[]*domain.GameCreatorCustomJob{
						domain.NewGameCreatorCustomJob(
							values.GameCreatorJobID(customJob2.ID),
							values.GameCreatorJobDisplayName(customJob2.DisplayName),
							values.GameID(customJob2.GameID),
							customJob2.CreatedAt,
						),
						domain.NewGameCreatorCustomJob(
							values.GameCreatorJobID(customJob3.ID),
							values.GameCreatorJobDisplayName(customJob3.DisplayName),
							values.GameID(customJob3.GameID),
							customJob3.CreatedAt,
						),
					},
				),
				domain.NewGameCreatorWithJobs(
					domain.NewGameCreator(
						values.GameCreatorID(creator3.ID),
						values.TraPMemberID(creator3.UserID),
						values.TraPMemberName(creator3.UserName),
						creator3.CreatedAt),
					[]*domain.GameCreatorJob{},
					[]*domain.GameCreatorCustomJob{},
				),
			},
		},
		"該当するゲームクリエイターがいない場合は空配列を返す": {
			gameID:      values.NewGameID(),
			allCreators: []*schema.GameCreatorTable{creator1, creator2},
			expected:    []*domain.GameCreatorWithJobs{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			err = db.Create(testCase.allCreators).Error
			require.NoError(t, err)
			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Select(clause.Associations).Delete(testCase.allCreators).Error
				require.NoError(t, err)
			})

			repo := NewGameCreator(testDB)

			creatorsWithJobs, err := repo.GetGameCreatorsByGameID(t.Context(), testCase.gameID)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)

			assert.Len(t, creatorsWithJobs, len(testCase.expected))

			for i, creatorWithJobs := range creatorsWithJobs {
				creator := creatorWithJobs.GetGameCreator()
				expectedCreator := testCase.expected[i].GetGameCreator()
				assert.Equal(t, expectedCreator.GetID(), creator.GetID())
				assert.Equal(t, expectedCreator.GetUserID(), creator.GetUserID())
				assert.Equal(t, expectedCreator.GetUserName(), creator.GetUserName())
				assert.WithinDuration(t, expectedCreator.GetCreatedAt(), creator.GetCreatedAt(), time.Second)

				actualJobs := creatorWithJobs.GetJobs()
				assert.Len(t, actualJobs, len(testCase.expected[i].GetJobs()))
				for j, expectedJob := range testCase.expected[i].GetJobs() {
					actualJob := actualJobs[j]
					assert.Equal(t, expectedJob.GetID(), actualJob.GetID())
					assert.Equal(t, expectedJob.GetDisplayName(), actualJob.GetDisplayName())
					assert.WithinDuration(t, expectedJob.GetCreatedAt(), actualJob.GetCreatedAt(), time.Second)
				}

				actualCustomJobs := creatorWithJobs.GetCustomJobs()
				assert.Len(t, actualCustomJobs, len(testCase.expected[i].GetCustomJobs()))
				for j, expectedJob := range testCase.expected[i].GetCustomJobs() {
					actualJob := actualCustomJobs[j]
					assert.Equal(t, expectedJob.GetID(), actualJob.GetID())
					assert.Equal(t, expectedJob.GetDisplayName(), actualJob.GetDisplayName())
					assert.Equal(t, expectedJob.GetGameID(), actualJob.GetGameID())
					assert.WithinDuration(t, expectedJob.GetCreatedAt(), actualJob.GetCreatedAt(), time.Second)
				}
			}
		})
	}
}
