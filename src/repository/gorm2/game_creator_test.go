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

				actualJobsMap := make(map[values.GameCreatorJobID]*domain.GameCreatorJob, len(creatorWithJobs.GetJobs()))
				for _, job := range creatorWithJobs.GetJobs() {
					actualJobsMap[job.GetID()] = job
				}
				assert.Len(t, actualJobsMap, len(testCase.expected[i].GetJobs()))
				for _, expectedJob := range testCase.expected[i].GetJobs() {
					actualJob, ok := actualJobsMap[expectedJob.GetID()]
					assert.Truef(t, ok, "expected job ID %v not found in actual jobs", expectedJob.GetID())
					assert.Equal(t, expectedJob.GetID(), actualJob.GetID())
					assert.Equal(t, expectedJob.GetDisplayName(), actualJob.GetDisplayName())
					assert.WithinDuration(t, expectedJob.GetCreatedAt(), actualJob.GetCreatedAt(), time.Second)
				}

				actualCustomJobsMap := make(map[values.GameCreatorJobID]*domain.GameCreatorCustomJob, len(creatorWithJobs.GetCustomJobs()))
				for _, job := range creatorWithJobs.GetCustomJobs() {
					actualCustomJobsMap[job.GetID()] = job
				}
				assert.Len(t, actualCustomJobsMap, len(testCase.expected[i].GetCustomJobs()))
				for _, expectedJob := range testCase.expected[i].GetCustomJobs() {
					actualJob, ok := actualCustomJobsMap[expectedJob.GetID()]
					assert.Truef(t, ok, "expected custom job ID %v not found in actual custom jobs", expectedJob.GetID())
					assert.Equal(t, expectedJob.GetID(), actualJob.GetID())
					assert.Equal(t, expectedJob.GetDisplayName(), actualJob.GetDisplayName())
					assert.Equal(t, expectedJob.GetGameID(), actualJob.GetGameID())
					assert.WithinDuration(t, expectedJob.GetCreatedAt(), actualJob.GetCreatedAt(), time.Second)
				}
			}
		})
	}
}

func TestGetGameCreatorPresetJobs(t *testing.T) {
	now := time.Now()
	job1 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "プログラマー",
		CreatedAt:   now,
	}
	job2 := &schema.GameCreatorJobTable{
		ID:          uuid.New(),
		DisplayName: "デザイナー",
		CreatedAt:   now,
	}

	testCases := map[string]struct {
		presetJobs []*schema.GameCreatorJobTable
		expected   []*domain.GameCreatorJob
		err        error
	}{
		"プリセットジョブが存在する場合は取得できる": {
			presetJobs: []*schema.GameCreatorJobTable{job1, job2},
			expected: []*domain.GameCreatorJob{
				domain.NewGameCreatorJob(
					values.GameCreatorJobID(job1.ID),
					values.GameCreatorJobDisplayName(job1.DisplayName),
					job1.CreatedAt,
				),
				domain.NewGameCreatorJob(
					values.GameCreatorJobID(job2.ID),
					values.GameCreatorJobDisplayName(job2.DisplayName),
					job2.CreatedAt,
				),
			},
		},
		"プリセットジョブが存在しない場合は空配列を返す": {
			presetJobs: []*schema.GameCreatorJobTable{},
			expected:   []*domain.GameCreatorJob{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			if len(testCase.presetJobs) > 0 {
				err = db.Create(testCase.presetJobs).Error
				require.NoError(t, err)
			}

			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(testCase.presetJobs).Error
				require.NoError(t, err)
			})

			repo := NewGameCreator(testDB)

			jobs, err := repo.GetGameCreatorPresetJobs(t.Context())

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, jobs, len(testCase.expected))

			expectedMap := make(map[values.GameCreatorJobID]*domain.GameCreatorJob, len(testCase.expected))
			for _, job := range testCase.expected {
				expectedMap[job.GetID()] = job
			}

			for _, job := range jobs {
				expected, ok := expectedMap[job.GetID()]
				if !assert.Truef(t, ok, "unexpected job ID: %v", job.GetID()) {
					continue
				}
				assert.Equal(t, expected.GetDisplayName(), job.GetDisplayName())
				assert.WithinDuration(t, expected.GetCreatedAt(), job.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestGetGameCreatorCustomJobsByGameID(t *testing.T) {
	db, err := testDB.getDB(t.Context())
	require.NoError(t, err)

	now := time.Now()

	gameID1 := values.NewGameID()
	game1 := &schema.GameTable2{
		ID:               uuid.UUID(gameID1),
		Name:             "test game 1",
		VisibilityTypeID: 1,
	}
	gameID2 := values.NewGameID()
	game2 := &schema.GameTable2{
		ID:               uuid.UUID(gameID2),
		Name:             "test game 2",
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

	customJob1 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "ボーカル",
		GameID:      uuid.UUID(gameID1),
		CreatedAt:   now,
	}
	customJob2 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "作曲者",
		GameID:      uuid.UUID(gameID2),
		CreatedAt:   now,
	}
	customJob3 := &schema.GameCreatorCustomJobTable{
		ID:          uuid.New(),
		DisplayName: "イラストレーター",
		GameID:      uuid.UUID(gameID2),
		CreatedAt:   now,
	}

	testCases := map[string]struct {
		gameID     values.GameID
		customJobs []*schema.GameCreatorCustomJobTable
		expected   []*domain.GameCreatorCustomJob
		err        error
	}{
		"対象ゲームのカスタムジョブを1件取得できる": {
			gameID:     gameID1,
			customJobs: []*schema.GameCreatorCustomJobTable{customJob1, customJob2, customJob3},
			expected: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					values.GameCreatorJobID(customJob1.ID),
					values.GameCreatorJobDisplayName(customJob1.DisplayName),
					values.GameID(customJob1.GameID),
					customJob1.CreatedAt,
				),
			},
		},
		"対象ゲームのカスタムジョブを複数件取得できる": {
			gameID:     gameID2,
			customJobs: []*schema.GameCreatorCustomJobTable{customJob1, customJob2, customJob3},
			expected: []*domain.GameCreatorCustomJob{
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
		},
		"対象ゲームにカスタムジョブが存在しない場合は空配列": {
			gameID:     gameID1,
			customJobs: []*schema.GameCreatorCustomJobTable{},
			expected:   []*domain.GameCreatorCustomJob{},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			db, err := testDB.getDB(t.Context())
			require.NoError(t, err)

			if len(testCase.customJobs) > 0 {
				err = db.Create(testCase.customJobs).Error
				require.NoError(t, err)
			}

			t.Cleanup(func() {
				db, err := testDB.getDB(context.Background())
				require.NoError(t, err)
				err = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(testCase.customJobs).Error
				require.NoError(t, err)
			})

			repo := NewGameCreator(testDB)

			customJobs, err := repo.GetGameCreatorCustomJobsByGameID(t.Context(), testCase.gameID)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, customJobs, len(testCase.expected))

			expectedMap := make(map[values.GameCreatorJobID]*domain.GameCreatorCustomJob, len(testCase.expected))
			for _, job := range testCase.expected {
				expectedMap[job.GetID()] = job
			}

			for _, job := range customJobs {
				expected, ok := expectedMap[job.GetID()]
				if !assert.Truef(t, ok, "unexpected custom job ID: %v", job.GetID()) {
					continue
				}
				assert.Equal(t, expected.GetDisplayName(), job.GetDisplayName())
				assert.Equal(t, expected.GetGameID(), job.GetGameID())
				assert.WithinDuration(t, expected.GetCreatedAt(), job.GetCreatedAt(), time.Second)
			}
		})
	}
}
